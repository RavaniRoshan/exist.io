package origin

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"exist.io/internal/audit"
	"exist.io/internal/gitstore"
	"exist.io/internal/metrics"
	"exist.io/internal/security"
	"exist.io/pkg/scopes"

	"github.com/go-git/go-git/v6/plumbing/object"
)

// registerRefs wires the git ref routes and createCommitFromFiles (Task 10).
func (s *Server) registerRefs(mux *http.ServeMux) {
	read := scopes.Requirements{TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser}, Scopes: []string{scopes.RepositoryContentsRead}}
	write := scopes.Requirements{TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser}, Scopes: []string{scopes.RepositoryContentsWrite}}
	s.register("OriginService_GetGitRef", read)
	s.register("OriginService_ListMatchingGitRefs", read)
	s.register("OriginService_CreateGitRef", write)
	s.register("OriginService_DeleteGitRef", write)
	s.register("OriginService_CreateCommitFromFiles", write)

	base := "/v1/origin/repos/{ownerSlug}/{repoName}"
	mux.Handle("GET "+base+"/git/ref/{ref...}", s.guard("OriginService_GetGitRef", s.getGitRef))
	mux.Handle("GET "+base+"/git/matching-refs/{ref...}", s.guard("OriginService_ListMatchingGitRefs", s.listMatchingGitRefs))
	mux.Handle("POST "+base+"/git/refs", s.guard("OriginService_CreateGitRef", s.createGitRef))
	mux.Handle("DELETE "+base+"/git/refs/{ref...}", s.guard("OriginService_DeleteGitRef", s.deleteGitRef))
	mux.Handle("POST "+base+"/git/commits:createFromFiles", s.guard("OriginService_CreateCommitFromFiles", s.createCommitFromFiles))
}

// gitRefDTO is the Origin GitRef shape.
type gitRefDTO struct {
	Ref    string `json:"ref"`
	Object struct {
		Type string `json:"type"`
		SHA  string `json:"sha"`
	} `json:"object"`
}

func fullRefRef(short string) string {
	if strings.HasPrefix(short, "refs/") {
		return short
	}
	return "refs/heads/" + short
}

func (s *Server) getGitRef(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	ref := r.PathValue("ref")
	info, err := s.git.GetRef(owner, name, ref)
	if err != nil {
		notFound(w, "ref not found")
		return
	}
	out := gitRefDTO{Ref: fullRefRef(info.Ref)}
	out.Object.Type = "commit"
	out.Object.SHA = info.SHA
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) listMatchingGitRefs(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	refs, err := s.git.ListMatchingRefs(owner, name, r.PathValue("ref"))
	if err != nil {
		internalError(w, "list refs: "+err.Error())
		return
	}
	out := make([]gitRefDTO, 0, len(refs))
	for _, info := range refs {
		g := gitRefDTO{Ref: fullRefRef(info.Ref)}
		g.Object.Type = "commit"
		g.Object.SHA = info.SHA
		out = append(out, g)
	}
	writeJSON(w, http.StatusOK, map[string]any{"refs": out})
}

func (s *Server) createGitRef(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	var req struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Ref == "" || req.SHA == "" {
		badRequest(w, "ref and sha are required")
		return
	}
	info, err := s.git.CreateBranchRef(owner, name, req.Ref, req.SHA)
	if err != nil {
		if errors.Is(err, gitstore.ErrRefExists) {
			conflict(w, "ref already exists")
			return
		}
		if strings.Contains(err.Error(), "only branch references") || strings.Contains(err.Error(), "not found") {
			badRequest(w, err.Error())
			return
		}
		internalError(w, "create ref: "+err.Error())
		return
	}
	out := gitRefDTO{Ref: fullRefRef(info.Ref)}
	out.Object.Type = "commit"
	out.Object.SHA = info.SHA
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) deleteGitRef(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	if err := s.git.DeleteRef(owner, name, r.PathValue("ref")); err != nil {
		if errors.Is(err, gitstore.ErrRefNotFound) {
			notFound(w, "ref not found")
			return
		}
		internalError(w, "delete ref: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// createCommitFromFilesRequest is the Origin CreateCommitFromFilesRequest.
// Required: identifier, targetBranch, expectedHeadSha, message, author, files.
type createCommitFromFilesRequest struct {
	TargetBranch    string           `json:"targetBranch"`
	ExpectedHeadSHA string           `json:"expectedHeadSha"`
	Message         string           `json:"message"`
	Author          commitSignature  `json:"author"`
	Committer       *commitSignature `json:"committer"`
	Files           []fileChangeReq  `json:"files"`
}

type commitSignature struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type fileChangeReq struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	Delete   bool   `json:"delete"`
}

func (s *Server) createCommitFromFiles(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	var req createCommitFromFilesRequest
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.TargetBranch == "" || req.Message == "" || req.Author.Name == "" || req.Author.Email == "" {
		badRequest(w, "targetBranch, message, and author (name+email) are required")
		return
	}
	if len(req.Files) == 0 {
		badRequest(w, "at least one file change is required")
		return
	}
	// Paths must be unique within a request.
	seen := map[string]bool{}
	for _, f := range req.Files {
		if f.Path == "" {
			badRequest(w, "file path is required")
			return
		}
		if seen[f.Path] {
			badRequest(w, "duplicate file path: "+f.Path)
			return
		}
		seen[f.Path] = true
		if !f.Delete && f.Content == "" && f.Encoding != "base64" {
			// empty content with utf-8 is allowed (empty file), so no error here
		}
	}

	// expectedHeadSha precondition: the branch tip must match, else
	// FAILED_PRECONDITION (google.rpc code 9, mapped to HTTP 409).
	head, err := s.git.ResolveRef(owner, name, req.TargetBranch)
	if err != nil {
		notFound(w, "target branch not found")
		return
	}
	headSHA := head.String()
	if headSHA != req.ExpectedHeadSHA {
		writeError(w, http.StatusConflict, codeFailedPrecondition,
			"branch tip does not match expectedHeadSha")
		return
	}

	// Build the file changes.
	changes := make([]gitstore.FileChange, 0, len(req.Files))
	for _, f := range req.Files {
		content := f.Content
		if f.Encoding == "base64" {
			decoded, derr := base64.StdEncoding.DecodeString(f.Content)
			if derr != nil {
				badRequest(w, "invalid base64 content for "+f.Path)
				return
			}
			content = string(decoded)
		}

		if !f.Delete {
			violations := security.DefaultScanner.ScanFile(f.Path, []byte(content))
			if len(violations) > 0 {
				metrics.DefaultRegistry.IncSecretPushBlocks()
				_ = audit.Log(audit.Event{
					Actor:     req.Author.Email,
					ActorType: "user",
					Action:    "security.secret_commit_blocked",
					Resource:  owner + "/" + name + ":" + f.Path,
					Severity:  audit.SeverityCritical,
					Details:   fmt.Sprintf("Blocked commit containing sensitive credentials in %s", f.Path),
				})
				writeError(w, http.StatusUnprocessableEntity, codeFailedPrecondition,
					security.FormatPushRejection(violations))
				return
			}
		}

		changes = append(changes, gitstore.FileChange{Path: f.Path, Content: content, Delete: f.Delete})
	}

	sig := object.Signature{Name: req.Author.Name, Email: req.Author.Email}
	sha, err := s.git.WriteFiles(owner, name, req.TargetBranch, changes, sig, req.Message)
	if err != nil {
		internalError(w, "commit: "+err.Error())
		return
	}
	newCommit, err := s.git.ReadCommit(owner, name, sha.String())
	if err != nil {
		internalError(w, "read commit: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sha":             sha.String(),
		"treeSha":         newCommit.TreeSHA,
		"previousHeadSha": headSHA,
	})
}
