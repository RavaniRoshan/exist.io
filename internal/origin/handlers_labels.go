package origin

import (
	"errors"
	"net/http"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"exist.io/pkg/scopes"
)

var (
	reqLabelsRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryLabelsRead},
	}
	reqLabelsWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryLabelsWrite},
	}
)

func (s *Server) registerLabels(mux *http.ServeMux) {
	s.register("OriginService_ListLabels", reqLabelsRead)
	s.register("OriginService_CreateLabel", reqLabelsWrite)
	s.register("OriginService_GetLabel", reqLabelsRead)
	s.register("OriginService_UpdateLabel", reqLabelsWrite)
	s.register("OriginService_DeleteLabel", reqLabelsWrite)

	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/labels",
		s.guard("OriginService_ListLabels", s.listLabels))
	mux.Handle("POST /v1/origin/repos/{ownerSlug}/{repoName}/labels",
		s.guard("OriginService_CreateLabel", s.createLabel))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/labels/{labelName}",
		s.guard("OriginService_GetLabel", s.getLabel))
	mux.Handle("PATCH /v1/origin/repos/{ownerSlug}/{repoName}/labels/{labelName}",
		s.guard("OriginService_UpdateLabel", s.updateLabel))
	mux.Handle("DELETE /v1/origin/repos/{ownerSlug}/{repoName}/labels/{labelName}",
		s.guard("OriginService_DeleteLabel", s.deleteLabel))
}

func (s *Server) listLabels(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	_, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	labels, err := s.store.ListLabels(r.Context(), ns.ID)
	if err != nil {
		internalError(w, "list labels: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, l := range labels {
		dtos = append(dtos, toLabelDTO(l))
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"labels":        dtos,
		"nextPageToken": "",
	})
}

type labelInputReq struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description,omitempty"`
}

func (s *Server) createLabel(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	_, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	var req labelInputReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Name == "" || req.Color == "" {
		badRequest(w, "name and color are required")
		return
	}

	lbl, err := s.store.GetOrCreateLabel(r.Context(), idgen.NewLabel(), ns.ID, req.Name, req.Color, req.Description)
	if err != nil {
		internalError(w, "create label: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toLabelDTO(lbl))
}

func (s *Server) getLabel(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	_, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	labelName := r.PathValue("labelName")
	lbl, err := s.store.GetLabel(r.Context(), ns.ID, labelName)
	if err != nil {
		notFound(w, "label not found")
		return
	}

	writeJSON(w, http.StatusOK, toLabelDTO(lbl))
}

func (s *Server) updateLabel(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	_, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	labelName := r.PathValue("labelName")
	var req labelInputReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	updated, err := s.store.UpdateLabel(r.Context(), ns.ID, labelName, req.Name, req.Color, req.Description)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFound(w, "label not found")
			return
		}
		internalError(w, "update label: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toLabelDTO(updated))
}

func (s *Server) deleteLabel(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	_, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	labelName := r.PathValue("labelName")
	if err := s.store.DeleteLabel(r.Context(), ns.ID, labelName); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFound(w, "label not found")
			return
		}
		internalError(w, "delete label: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toLabelDTO(l store.Label) map[string]any {
	dto := map[string]any{
		"id":    l.ID,
		"name":  l.Name,
		"color": l.Color,
	}
	if l.Description != "" {
		dto["description"] = l.Description
	}
	return dto
}
