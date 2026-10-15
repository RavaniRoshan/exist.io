package origin

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"exist.io/internal/gitstore"
	"exist.io/internal/symbolsearch"
	"exist.io/pkg/scopes"
)

// registerContents wires the git-contents read routes (Task 9) and the ref
// read routes that live alongside them. Write routes (refs, createFromFiles)
// are wired in Task 10.
func (s *Server) registerContents(mux *http.ServeMux) {
	read := scopes.Requirements{TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser}, Scopes: []string{scopes.RepositoryContentsRead}}
	s.register("OriginService_GetContents", read)
	s.register("OriginService_BatchGetContents", read)
	s.register("OriginService_GetBlob", read)
	s.register("OriginService_GetTree", read)
	s.register("OriginService_GetGitCommit", read)
	s.register("OriginService_ListCommits", read)
	s.register("OriginService_GetCommit", read)
	s.register("OriginService_ListBranches", read)
	s.register("OriginService_GetRepoTarball", read)
	s.register("OriginService_GrepContents", read)
	s.register("OriginService_ListCommitFiles", read)
	s.register("OriginService_CompareCommits", read)
	s.register("OriginService_SearchSymbols", read)

	base := "/v1/origin/repos/{ownerSlug}/{repoName}"
	mux.Handle("GET "+base+"/contents", s.guard("OriginService_GetContents", s.getContents))
	mux.Handle("POST "+base+"/contents:batchGet", s.guard("OriginService_BatchGetContents", s.batchGetContents))
	mux.Handle("GET "+base+"/git/blobs/{sha}", s.guard("OriginService_GetBlob", s.getBlob))
	mux.Handle("GET "+base+"/git/trees/{sha}", s.guard("OriginService_GetTree", s.getTree))
	mux.Handle("GET "+base+"/git/commits/{sha}", s.guard("OriginService_GetGitCommit", s.getGitCommit))
	mux.Handle("GET "+base+"/commits", s.guard("OriginService_ListCommits", s.listCommits))
	mux.Handle("GET "+base+"/commits/{sha}", s.guard("OriginService_GetCommit", s.getCommit))
	mux.Handle("GET "+base+"/branches", s.guard("OriginService_ListBranches", s.listBranches))
	mux.Handle("GET "+base+"/tarball", s.guard("OriginService_GetRepoTarball", s.getTarball))
	mux.Handle("POST "+base+"/contents:grep", s.guard("OriginService_GrepContents", s.grepContents))
	mux.Handle("GET "+base+"/commits/{sha}/files", s.guard("OriginService_ListCommitFiles", s.listCommitFiles))
	mux.Handle("GET "+base+"/symbols", s.guard("OriginService_SearchSymbols", s.searchSymbols))
}

// contentDTO is the Origin Content shape (file or directory).
type contentDTO struct {
	Type     string       `json:"type"` // "file" | "dir"
	Encoding string       `json:"encoding,omitempty"`
	Size     string       `json:"size,omitempty"`
	Name     string       `json:"name,omitempty"`
	Path     string       `json:"path,omitempty"`
	SHA      string       `json:"sha,omitempty"`
	Content  string       `json:"content,omitempty"` // base64
	Entries  []contentDTO `json:"entries,omitempty"`
}

// openRepo resolves owner/name to a store.Repo, writing 404 on miss.
func (s *Server) openRepo(w http.ResponseWriter, r *http.Request) (owner, name string, ok bool) {
	owner, name = r.PathValue("ownerSlug"), r.PathValue("repoName")
	if _, err := s.store.GetRepo(r.Context(), owner, name); err != nil {
		notFound(w, "repository not found")
		return "", "", false
	}
	return owner, name, true
}

// resolveCommit resolves a ref (query "ref" or default) to a commit SHA.
func (s *Server) resolveCommit(w http.ResponseWriter, r *http.Request, owner, name, ref string) (sha string, ok bool) {
	h, err := s.git.ResolveRef(owner, name, ref)
	if err != nil {
		notFound(w, "ref not found")
		return "", false
	}
	return h.String(), true
}

func (s *Server) getContents(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	path := r.URL.Query().Get("path")
	sha, ok := s.resolveCommit(w, r, owner, name, r.URL.Query().Get("ref"))
	if !ok {
		return
	}
	data, blobSHA, size, isDir, err := s.git.ReadFileAtPath(owner, name, sha, path)
	if err != nil {
		notFound(w, "path not found")
		return
	}
	if isDir {
		_, entries, err := s.git.ReadTree(owner, name, sha, path, false)
		if err != nil {
			internalError(w, "read tree: "+err.Error())
			return
		}
		dto := contentDTO{Type: "dir", Path: path, SHA: blobSHA, Name: baseName(path)}
		for _, e := range entries {
			dto.Entries = append(dto.Entries, contentDTO{Type: e.Type, Path: e.Path, SHA: e.SHA, Name: baseName(e.Path)})
		}
		writeJSON(w, http.StatusOK, dto)
		return
	}
	writeJSON(w, http.StatusOK, contentDTO{
		Type: "file", Encoding: "base64", Size: strconv.FormatInt(size, 10),
		Name: baseName(path), Path: path, SHA: blobSHA,
		Content: base64.StdEncoding.EncodeToString(data),
	})
}

func (s *Server) batchGetContents(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	var req struct {
		Paths []string `json:"paths"`
		Ref   string   `json:"ref"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	sha, ok := s.resolveCommit(w, r, owner, name, req.Ref)
	if !ok {
		return
	}
	// Deduplicate paths, cap at 20 (Origin limit).
	seen := map[string]bool{}
	var paths []string
	for _, p := range req.Paths {
		if !seen[p] && len(paths) < 20 {
			seen[p] = true
			paths = append(paths, p)
		}
	}
	type result struct {
		Path    string      `json:"path"`
		Found   bool        `json:"found"`
		Content *contentDTO `json:"content,omitempty"`
	}
	var results []result
	for _, p := range paths {
		data, blobSHA, size, isDir, err := s.git.ReadFileAtPath(owner, name, sha, p)
		if err != nil || isDir {
			results = append(results, result{Path: p, Found: false})
			continue
		}
		results = append(results, result{Path: p, Found: true, Content: &contentDTO{
			Type: "file", Encoding: "base64", Size: strconv.FormatInt(size, 10),
			Name: baseName(p), Path: p, SHA: blobSHA,
			Content: base64.StdEncoding.EncodeToString(data),
		}})
	}
	writeJSON(w, http.StatusOK, map[string]any{"resolvedCommitSha": sha, "results": results})
}

func (s *Server) getBlob(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	data, size, err := s.git.ReadBlob(owner, name, r.PathValue("sha"))
	if err != nil {
		notFound(w, "blob not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sha": r.PathValue("sha"), "encoding": "base64",
		"size":    size,
		"content": base64.StdEncoding.EncodeToString(data),
	})
}

func (s *Server) getTree(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	sha := r.PathValue("sha")
	recursive := r.URL.Query().Get("recursive") == "true" || r.URL.Query().Get("recursive") == "1"
	treeSHA, entries, err := s.git.ReadTree(owner, name, sha, "", recursive)
	if err != nil {
		notFound(w, "tree not found")
		return
	}
	type entry struct {
		Mode string `json:"mode"`
		Path string `json:"path"`
		SHA  string `json:"sha"`
		Type string `json:"type"`
		Size int64  `json:"size,omitempty"`
	}
	out := make([]entry, 0, len(entries))
	for _, e := range entries {
		out = append(out, entry{Mode: e.Mode, Path: e.Path, SHA: e.SHA, Type: e.Type, Size: e.Size})
	}
	writeJSON(w, http.StatusOK, map[string]any{"sha": treeSHA, "tree": out, "truncated": false})
}

func (s *Server) getGitCommit(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	ci, err := s.git.ReadCommit(owner, name, r.PathValue("sha"))
	if err != nil {
		notFound(w, "commit not found")
		return
	}
	writeJSON(w, http.StatusOK, gitCommitDTO(ci))
}

func gitCommitDTO(ci gitstore.CommitInfo) map[string]any {
	return map[string]any{
		"sha":       ci.SHA,
		"author":    map[string]string{"name": ci.AuthorName, "email": ci.AuthorEmail, "date": ci.AuthorDate},
		"committer": map[string]string{"name": ci.CommitterName, "email": ci.CommitterEmail, "date": ci.CommitterDate},
		"message":   ci.Message,
		"tree":      ci.TreeSHA,
		"parents":   ci.Parents,
	}
}

func (s *Server) listCommits(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	commits, err := s.git.ListCommits(owner, name, r.URL.Query().Get("ref"), parsePageSize(r))
	if err != nil {
		notFound(w, "ref not found")
		return
	}
	out := make([]map[string]any, 0, len(commits))
	for _, c := range commits {
		out = append(out, map[string]any{"sha": c.SHA, "commit": map[string]any{
			"author":    map[string]string{"name": c.AuthorName, "email": c.AuthorEmail, "date": c.AuthorDate},
			"committer": map[string]string{"name": c.CommitterName, "email": c.CommitterEmail, "date": c.CommitterDate},
			"message":   c.Message, "tree": c.TreeSHA,
		}})
	}
	writeJSON(w, http.StatusOK, map[string]any{"commits": out, "nextPageToken": ""})
}

func (s *Server) getCommit(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	ci, err := s.git.ReadCommit(owner, name, r.PathValue("sha"))
	if err != nil {
		notFound(w, "commit not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sha": ci.SHA, "commit": map[string]any{
			"author":    map[string]string{"name": ci.AuthorName, "email": ci.AuthorEmail, "date": ci.AuthorDate},
			"committer": map[string]string{"name": ci.CommitterName, "email": ci.CommitterEmail, "date": ci.CommitterDate},
			"message":   ci.Message, "tree": ci.TreeSHA,
		},
		"parents": ci.Parents,
	})
}

func (s *Server) listBranches(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	branches, err := s.git.ListBranches(owner, name)
	if err != nil {
		internalError(w, "list branches: "+err.Error())
		return
	}
	type branch struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	out := make([]branch, 0, len(branches))
	for _, b := range branches {
		var br branch
		br.Name = b.Name
		br.Commit.SHA = b.SHA
		out = append(out, br)
	}
	writeJSON(w, http.StatusOK, map[string]any{"branches": out, "nextPageToken": ""})
}

func (s *Server) getTarball(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	ref := r.URL.Query().Get("ref")
	if ref == "" {
		ref = "HEAD"
	}
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+".tar.gz\"")
	w.WriteHeader(http.StatusOK)
	_ = s.git.StreamTarball(w, owner, name, ref)
}

func (s *Server) grepContents(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	var req struct {
		Query           string   `json:"query"`
		Ref             string   `json:"ref"`
		PathPrefixes    []string `json:"pathPrefixes"`
		CaseInsensitive bool     `json:"caseInsensitive"`
		Literal         bool     `json:"literal"`
		MaxResults      int      `json:"maxResults"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Query == "" {
		badRequest(w, "query is required")
		return
	}
	sha, ok := s.resolveCommit(w, r, owner, name, req.Ref)
	if !ok {
		return
	}
	matches, limitHit, err := s.git.Grep(owner, name, sha, req.Query, req.PathPrefixes, req.CaseInsensitive, req.Literal, req.MaxResults)
	if err != nil {
		internalError(w, "grep: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"matches": matches, "limitHit": limitHit})
}

func (s *Server) listCommitFiles(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	files, err := s.git.CommitFiles(owner, name, r.PathValue("sha"))
	if err != nil {
		notFound(w, "commit not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"files": files, "nextPageToken": ""})
}

func baseName(p string) string {
	if p == "" {
		return ""
	}
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func (s *Server) searchSymbols(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}

	ref := r.URL.Query().Get("ref")
	if ref == "" {
		repo, err := s.store.GetRepo(r.Context(), owner, name)
		if err == nil && repo.DefaultBranch != "" {
			ref = repo.DefaultBranch
		} else {
			ref = "HEAD"
		}
	}

	query := r.URL.Query().Get("q")
	kind := r.URL.Query().Get("kind")
	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	symbols, err := symbolsearch.Search(r.Context(), s.git, owner, name, ref, query, kind, limit)
	if err != nil {
		internalError(w, "symbol search failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"symbols": symbols,
		"ref":     ref,
		"count":   len(symbols),
	})
}

