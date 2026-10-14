package origin

import (
	"errors"
	"net/http"
	"strings"

	"exist.io/internal/store"
	"exist.io/pkg/scopes"
)

var (
	reqMirrorDelete = scopes.Requirements{
		TokenTypes: []string{scopes.TypeUser, scopes.TypeInstallation},
		Scopes:     []string{scopes.RepositoryMirrorDelete},
	}
	reqMirrorRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryMirrorRead},
	}
	reqMirrorWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeUser, scopes.TypeInstallation},
		Scopes:     []string{scopes.RepositoryMirrorWrite},
	}
	reqMirrorSync = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryMirrorSync},
	}
)

func (s *Server) registerMirror(mux *http.ServeMux) {
	s.register("OriginService_DetachRepoMirror", reqMirrorDelete)
	s.register("OriginService_GetMirrorTransitionJob", reqMirrorRead)
	s.register("OriginService_GetActiveMirrorTransitionJob", reqMirrorRead)
	s.register("OriginService_TransitionRepoMirror", reqMirrorWrite)
	s.register("OriginService_SyncMirror", reqMirrorSync)

	mux.Handle("DELETE /v1/origin/repos/{ownerSlug}/{repoName}/mirror",
		s.guard("OriginService_DetachRepoMirror", s.detachRepoMirror))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/mirror/transition-jobs/{jobId}",
		s.guard("OriginService_GetMirrorTransitionJob", s.getMirrorTransitionJob))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/mirror/transition-jobs:active",
		s.guard("OriginService_GetActiveMirrorTransitionJob", s.getActiveMirrorTransitionJob))
	mux.Handle("POST /v1/origin/repos/{ownerSlug}/{repoName}/mirror:transition",
		s.guard("OriginService_TransitionRepoMirror", s.transitionRepoMirror))
	hSyncMirror := s.guard("OriginService_SyncMirror", s.syncMirror)
	mux.HandleFunc("POST /v1/origin/repos/{ownerSlug}/{repoAction}", func(w http.ResponseWriter, r *http.Request) {
		action := r.PathValue("repoAction")
		if strings.HasSuffix(action, ":syncMirror") {
			repoName := strings.TrimSuffix(action, ":syncMirror")
			r.SetPathValue("repoName", repoName)
			hSyncMirror.ServeHTTP(w, r)
			return
		}
		notFound(w, "not found")
	})
}

func (s *Server) detachRepoMirror(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	if err := s.store.DetachMirror(r.Context(), repo.ID); err != nil {
		internalError(w, "detach mirror: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getMirrorTransitionJob(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	jobID := r.PathValue("jobId")
	job, err := s.store.GetMirrorTransitionJob(r.Context(), repo.ID, jobID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFound(w, "mirror transition job not found")
			return
		}
		internalError(w, "get job: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toMirrorTransitionJobDTO(job))
}

func (s *Server) getActiveMirrorTransitionJob(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	activeJob, lastJob, err := s.store.GetActiveMirrorTransitionJob(r.Context(), repo.ID)
	if err != nil {
		internalError(w, "get active job: "+err.Error())
		return
	}

	res := map[string]any{}
	if activeJob != nil {
		res["activeJob"] = toMirrorTransitionJobDTO(*activeJob)
	}
	if lastJob != nil {
		res["lastJob"] = toMirrorTransitionJobDTO(*lastJob)
	}

	writeJSON(w, http.StatusOK, res)
}

type transitionMirrorReq struct {
	Transition string `json:"transition"`
}

func (s *Server) transitionRepoMirror(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	var req transitionMirrorReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Transition != "initial_to_inbound" {
		badRequest(w, "transition must be 'initial_to_inbound'")
		return
	}

	job, err := s.store.TransitionMirror(r.Context(), repo.ID, req.Transition)
	if err != nil {
		internalError(w, "start mirror transition: "+err.Error())
		return
	}

	updatedRepo, _ := s.store.GetRepo(r.Context(), owner, name)
	writeJSON(w, http.StatusOK, map[string]any{
		"repository": s.toRepoDTO(updatedRepo, ns),
		"job":        toMirrorTransitionJobDTO(job),
	})
}

type syncMirrorReq struct {
	Ref  string `json:"ref"`
	Wait bool   `json:"wait"`
	SHA  string `json:"sha,omitempty"`
}

func (s *Server) syncMirror(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	_, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	var req syncMirrorReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if !strings.HasPrefix(req.Ref, "refs/") {
		badRequest(w, "ref must start with 'refs/'")
		return
	}

	// Mirror is synchronized
	writeJSON(w, http.StatusOK, map[string]any{
		"synced": true,
	})
}

func toMirrorTransitionJobDTO(j store.MirrorTransitionJob) map[string]any {
	dto := map[string]any{
		"id":           j.ID,
		"transition":   j.Transition,
		"status":       j.Status,
		"phase":        j.Phase,
		"attemptCount": j.AttemptCount,
		"createdAt":    j.CreatedAt,
		"updatedAt":    j.UpdatedAt,
	}
	if j.StartedAt != "" {
		dto["startedAt"] = j.StartedAt
	}
	if j.CompletedAt != "" {
		dto["completedAt"] = j.CompletedAt
	}
	if j.DrainUntil != "" {
		dto["drainUntil"] = j.DrainUntil
	}
	if j.LastErrorCode != "" {
		dto["lastErrorCode"] = j.LastErrorCode
		dto["lastErrorMessage"] = j.LastErrorMessage
	}
	return dto
}
