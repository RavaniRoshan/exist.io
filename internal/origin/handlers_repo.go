package origin

import (
	"errors"
	"net/http"
	"strconv"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

// registerRepos wires the repository and namespace routes (Task 6).
func (s *Server) registerRepos(mux *http.ServeMux) {
	// Operations + scope requirements.
	s.register("OriginService_ListNamespaces", reqListNamespaces)
	s.register("OriginService_ListRepos", reqListRepos)
	s.register("OriginService_CreateRepo", reqCreateRepo)
	s.register("OriginService_GetRepo", reqGetRepo)
	s.register("OriginService_UpdateRepo", reqUpdateRepo)

	// Routes.
	mux.Handle("GET /v1/origin/namespaces", s.guard("OriginService_ListNamespaces", s.listNamespaces))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}", s.guard("OriginService_ListRepos", s.listRepos))
	mux.Handle("POST /v1/origin/repos/{ownerSlug}", s.guard("OriginService_CreateRepo", s.createRepo))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}", s.guard("OriginService_GetRepo", s.getRepo))
	mux.Handle("PATCH /v1/origin/repos/{ownerSlug}/{repoName}", s.guard("OriginService_UpdateRepo", s.updateRepo))
}

// createRepoRequest mirrors OriginService_CreateRepoRequest.
type createRepoRequest struct {
	OwnerSlug string    `json:"ownerSlug"`
	Repo      repoInDTO `json:"repo"`
}

type repoInDTO struct {
	Name          string `json:"name"`
	DefaultBranch string `json:"defaultBranch"`
	Visibility    string `json:"visibility"`
}

func (s *Server) listNamespaces(w http.ResponseWriter, r *http.Request) {
	// Origin lists the namespaces the caller can access. Without a principal→
	// namespace membership model wired yet, we list all namespaces (the store
	// is seeded by tests/admin). Grants (Task 7) refine per-principal access.
	namespaces, err := s.store.ListNamespaces(r.Context(), parsePageSize(r))
	if err != nil {
		internalError(w, "list namespaces: "+err.Error())
		return
	}
	out := make([]namespaceDTO, 0, len(namespaces))
	for _, ns := range namespaces {
		out = append(out, toNamespaceDTO(ns))
	}
	writeJSON(w, http.StatusOK, map[string]any{"namespaces": out, "nextPageToken": ""})
}

func (s *Server) listRepos(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("ownerSlug")
	pageSize := parsePageSize(r)
	repos, next, err := s.store.ListRepos(r.Context(), owner, pageSize, r.URL.Query().Get("pageToken"))
	if err != nil {
		internalError(w, "list repos: "+err.Error())
		return
	}
	ns, err := s.store.GetNamespaceBySlug(r.Context(), owner)
	if err != nil {
		// Unknown namespace → empty list (Origin returns 404 for unknown owner
		// on some ops; listing is tolerant).
		writeJSON(w, http.StatusOK, map[string]any{"repositories": []repoDTO{}, "nextPageToken": ""})
		return
	}
	dtos := make([]repoDTO, 0, len(repos))
	for _, repo := range repos {
		dtos = append(dtos, s.toRepoDTO(repo, ns))
	}
	writeJSON(w, http.StatusOK, map[string]any{"repositories": dtos, "nextPageToken": next})
}

func (s *Server) createRepo(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("ownerSlug")
	var req createRepoRequest
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Repo.Name == "" {
		badRequest(w, "repo.name is required")
		return
	}
	ns, err := s.store.GetNamespaceBySlug(r.Context(), owner)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	branch := req.Repo.DefaultBranch
	if branch == "" {
		branch = "main"
	}
	visibility := req.Repo.Visibility
	if visibility == "" {
		visibility = "private"
	}
	fullName := owner + "/" + req.Repo.Name

	repo := store.Repo{
		ID: idgen.NewRepo(), NamespaceID: ns.ID, Name: req.Repo.Name, FullName: fullName,
		DefaultBranch: branch, Visibility: visibility,
		CloneURL: s.cloneURL(owner, req.Repo.Name), WebURL: s.webURL(owner, req.Repo.Name),
		AllowMergeCommit: true, AllowSquashMerge: true,
	}
	if err := s.store.CreateRepo(r.Context(), repo); err != nil {
		if errors.Is(err, store.ErrConflict) {
			conflict(w, "repository already exists")
			return
		}
		internalError(w, "create repo: "+err.Error())
		return
	}
	// Initialize the bare git repo on disk.
	if _, err := s.git.Init(owner, req.Repo.Name, branch); err != nil {
		internalError(w, "init git repo: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.toRepoDTO(repo, ns))
}

func (s *Server) getRepo(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("ownerSlug")
	name := r.PathValue("repoName")
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}
	ns, err := s.store.GetNamespace(r.Context(), repo.NamespaceID)
	if err != nil {
		internalError(w, "load namespace: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.toRepoDTO(repo, ns))
}

func (s *Server) updateRepo(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("ownerSlug")
	name := r.PathValue("repoName")
	var in struct {
		DefaultBranch       *string `json:"defaultBranch"`
		Visibility          *string `json:"visibility"`
		AllowMergeCommit    *bool   `json:"allowMergeCommit"`
		AllowSquashMerge    *bool   `json:"allowSquashMerge"`
		DeleteBranchOnMerge *bool   `json:"deleteBranchOnMerge"`
	}
	if err := decodeJSON(r, &in); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	u := store.RepoUpdate{
		DefaultBranch: in.DefaultBranch, Visibility: in.Visibility,
		AllowMergeCommit: in.AllowMergeCommit, AllowSquashMerge: in.AllowSquashMerge,
		DeleteBranchOnMerge: in.DeleteBranchOnMerge,
	}
	if err := s.store.UpdateRepo(r.Context(), owner, name, u); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFound(w, "repository not found")
			return
		}
		internalError(w, "update repo: "+err.Error())
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		internalError(w, "reload repo: "+err.Error())
		return
	}
	ns, _ := s.store.GetNamespace(r.Context(), repo.NamespaceID)
	writeJSON(w, http.StatusOK, s.toRepoDTO(repo, ns))
}

// parsePageSize reads the pageSize query param, defaulting to 30 and clamping
// to 100 (Origin's documented behavior).
func parsePageSize(r *http.Request) int {
	v := r.URL.Query().Get("pageSize")
	if v == "" {
		return 30
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 30
	}
	if n > 100 {
		return 100
	}
	return n
}
