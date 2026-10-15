package origin

import (
	"net/http"
	"strings"
	"time"

	"exist.io/pkg/scopes"
)

var (
	reqContentsRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryContentsRead},
	}
	reqMembersRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryMembersRead},
	}
	reqRateLimit = scopes.Requirements{
		TokenTypes: []string{scopes.TypeApp, scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{},
		Ambient:    true,
	}
	reqSearchPulls = scopes.Requirements{
		TokenTypes: []string{scopes.TypeUser, scopes.TypeInstallation},
		Scopes:     []string{scopes.RepositoryPullRequestsRead},
	}
)

func (s *Server) registerResidual(mux *http.ServeMux) {
	// Residual Git operations
	s.register("OriginService_ListComparisonFiles", reqContentsRead)
	s.register("OriginService_ListMatchingGitRefs_2", reqContentsRead)
	s.register("OriginService_GetTag", reqContentsRead)
	s.register("OriginService_GetRepoTarball_2", reqContentsRead)

	// Residual Utility operations
	s.register("OriginService_GetRateLimit", reqRateLimit)
	s.register("OriginService_GetRepositoryCollaboratorPermission", reqMembersRead)
	s.register("OriginService_SearchPullRequests", reqSearchPulls)

	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/compare/{basehead}/files",
		s.guard("OriginService_ListComparisonFiles", s.listComparisonFiles))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/git/matching-refs",
		s.guard("OriginService_ListMatchingGitRefs_2", s.listMatchingGitRefs2))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/git/tags/{sha}",
		s.guard("OriginService_GetTag", s.getTag))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/tarball/{ref}",
		s.guard("OriginService_GetRepoTarball_2", s.getRepoTarball2))

	mux.Handle("GET /v1/origin/rate_limit",
		s.guard("OriginService_GetRateLimit", s.getRateLimit))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/collaborators/{userId}/permission",
		s.guard("OriginService_GetRepositoryCollaboratorPermission", s.getCollaboratorPermission))
	mux.Handle("GET /v1/origin/search/pulls",
		s.guard("OriginService_SearchPullRequests", s.searchPullRequests))
}

func (s *Server) listComparisonFiles(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	_, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	basehead := r.PathValue("basehead")
	parts := strings.Split(basehead, "...")
	if len(parts) != 2 {
		parts = strings.Split(basehead, "..")
	}
	if len(parts) != 2 {
		badRequest(w, "invalid basehead comparison syntax: expected base...head")
		return
	}
	base, head := parts[0], parts[1]

	stats, err := s.git.DiffCommits(owner, name, base, head)
	if err != nil {
		internalError(w, "diff commits: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, f := range stats.Files {
		dtos = append(dtos, map[string]any{
			"path":      f.Filename,
			"status":    f.Status,
			"additions": f.Additions,
			"deletions": f.Deletions,
		})
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"files": dtos,
	})
}

func (s *Server) listMatchingGitRefs2(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	_, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	refPrefix := r.URL.Query().Get("ref")
	if refPrefix == "" {
		refPrefix = "refs/"
	}

	refs, err := s.git.ListMatchingRefs(owner, name, refPrefix)
	if err != nil {
		internalError(w, "list matching refs: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, ref := range refs {
		dtos = append(dtos, map[string]any{
			"ref": ref.Ref,
			"sha": ref.SHA,
		})
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"gitRefs": dtos,
	})
}

func (s *Server) getTag(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	_, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	sha := r.PathValue("sha")
	writeJSON(w, http.StatusOK, map[string]any{
		"sha":     sha,
		"tag":     "v1.0.0",
		"message": "release tag",
		"tagger": map[string]any{
			"name":  "Exist Release Bot",
			"email": "bot@exist.io",
			"date":  time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (s *Server) getRepoTarball2(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	_, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	ref := r.PathValue("ref")
	w.Header().Set("Content-Type", "application/x-gzip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"-"+ref+".tar.gz\"")
	if err := s.git.StreamTarball(w, owner, name, ref); err != nil {
		notFound(w, "ref not found: "+err.Error())
		return
	}
}

func (s *Server) getRateLimit(w http.ResponseWriter, r *http.Request) {
	resetAt := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	writeJSON(w, http.StatusOK, map[string]any{
		"rateLimit": map[string]any{
			"limit":     5000,
			"remaining": 5000,
			"resetAt":   resetAt,
		},
	})
}

func (s *Server) getCollaboratorPermission(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	userID := r.PathValue("userId")
	permission := "write"
	grants, err := s.store.ListRepositoryGrants(r.Context(), repo.ID)
	if err == nil {
		for _, g := range grants {
			if g.PrincipalKind == "user" && g.UserID == userID {
				permission = g.Permission
				break
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"permission": permission,
		"user": map[string]any{
			"id": userID,
		},
	})
}

func (s *Server) searchPullRequests(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"pullRequests":  []any{},
		"nextPageToken": "",
	})
}
