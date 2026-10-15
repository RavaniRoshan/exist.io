package origin

import (
	"errors"
	"fmt"
	"net/http"

	"exist.io/internal/audit"
	"exist.io/internal/store"
	"exist.io/pkg/scopes"
)

var (
	reqRulesetsRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryRulesetsRead},
	}
	reqRulesetsWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryRulesetsWrite},
	}
)

func (s *Server) registerRulesets(mux *http.ServeMux) {
	s.register("OriginService_ListRulesets", reqRulesetsRead)
	s.register("OriginService_CreateRuleset", reqRulesetsWrite)
	s.register("OriginService_GetRuleset", reqRulesetsRead)
	s.register("OriginService_UpdateRuleset", reqRulesetsWrite)
	s.register("OriginService_DeleteRuleset", reqRulesetsWrite)

	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/rulesets",
		s.guard("OriginService_ListRulesets", s.listRulesets))
	mux.Handle("POST /v1/origin/repos/{ownerSlug}/{repoName}/rulesets",
		s.guard("OriginService_CreateRuleset", s.createRuleset))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/rulesets/{rulesetId}",
		s.guard("OriginService_GetRuleset", s.getRuleset))
	mux.Handle("PUT /v1/origin/repos/{ownerSlug}/{repoName}/rulesets/{rulesetId}",
		s.guard("OriginService_UpdateRuleset", s.updateRuleset))
	mux.Handle("DELETE /v1/origin/repos/{ownerSlug}/{repoName}/rulesets/{rulesetId}",
		s.guard("OriginService_DeleteRuleset", s.deleteRuleset))
}

type rulesetInput struct {
	Name             string                     `json:"name"`
	Description      string                     `json:"description,omitempty"`
	Enforcement      string                     `json:"enforcement"` // active | evaluate | disabled
	Kind             string                     `json:"kind"`        // merge_branch | push_branch | push_tag | push_repository
	IncludedRefNames []string                   `json:"includedRefNames,omitempty"`
	ExcludedRefNames []string                   `json:"excludedRefNames,omitempty"`
	Rules            []store.RulesetRule        `json:"rules,omitempty"`
	BypassActors     []store.RulesetBypassActor `json:"bypassActors,omitempty"`
}

func (s *Server) listRulesets(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, ns, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	rulesets, err := s.store.ListRulesets(r.Context(), repo.ID)
	if err != nil {
		internalError(w, "list rulesets: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, rs := range rulesets {
		dtos = append(dtos, toRulesetDTO(rs))
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"rulesets": dtos,
		"repository": map[string]any{
			"id":   repo.ID,
			"name": repo.Name,
			"owner": map[string]any{
				"id":   ns.ID,
				"slug": ns.Slug,
				"type": ns.Type,
			},
		},
	})
}

func (s *Server) createRuleset(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	var req rulesetInput
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Name == "" || req.Enforcement == "" || req.Kind == "" {
		badRequest(w, "missing required fields: name, enforcement, kind")
		return
	}

	rsStore := store.Ruleset{
		RepoID:           repo.ID,
		Name:             req.Name,
		Description:      req.Description,
		Enforcement:      req.Enforcement,
		Kind:             req.Kind,
		IncludedRefNames: req.IncludedRefNames,
		ExcludedRefNames: req.ExcludedRefNames,
		Rules:            req.Rules,
		BypassActors:     req.BypassActors,
	}

	created, err := s.store.CreateRuleset(r.Context(), rsStore)
	if err != nil {
		internalError(w, "create ruleset: "+err.Error())
		return
	}

	_ = audit.Log(audit.Event{
		Actor:     s.resolveActor(r),
		ActorType: "user",
		Action:    "repo.ruleset_created",
		Resource:  fmt.Sprintf("%s/%s:%s", owner, name, created.Name),
		ClientIP:  r.RemoteAddr,
		Severity:  audit.SeverityWarning,
		Details:   fmt.Sprintf("Created ruleset %q with enforcement %q", created.Name, created.Enforcement),
	})

	writeJSON(w, http.StatusOK, toRulesetDTO(created))
}

func (s *Server) getRuleset(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	rulesetID := r.PathValue("rulesetId")
	rs, err := s.store.GetRuleset(r.Context(), rulesetID)
	if err != nil || rs.RepoID != repo.ID {
		notFound(w, "ruleset not found")
		return
	}

	writeJSON(w, http.StatusOK, toRulesetDTO(rs))
}

func (s *Server) updateRuleset(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	rulesetID := r.PathValue("rulesetId")
	existing, err := s.store.GetRuleset(r.Context(), rulesetID)
	if err != nil || existing.RepoID != repo.ID {
		notFound(w, "ruleset not found")
		return
	}

	var req rulesetInput
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Name == "" || req.Enforcement == "" || req.Kind == "" {
		badRequest(w, "missing required fields: name, enforcement, kind")
		return
	}

	rsStore := store.Ruleset{
		ID:               rulesetID,
		RepoID:           repo.ID,
		Name:             req.Name,
		Description:      req.Description,
		Enforcement:      req.Enforcement,
		Kind:             req.Kind,
		IncludedRefNames: req.IncludedRefNames,
		ExcludedRefNames: req.ExcludedRefNames,
		Rules:            req.Rules,
		BypassActors:     req.BypassActors,
	}

	updated, err := s.store.UpdateRuleset(r.Context(), rsStore)
	if err != nil {
		internalError(w, "update ruleset: "+err.Error())
		return
	}

	_ = audit.Log(audit.Event{
		Actor:     s.resolveActor(r),
		ActorType: "user",
		Action:    "repo.ruleset_updated",
		Resource:  fmt.Sprintf("%s/%s:%s", owner, name, updated.Name),
		ClientIP:  r.RemoteAddr,
		Severity:  audit.SeverityWarning,
		Details:   fmt.Sprintf("Updated ruleset %q (enforcement: %q)", updated.Name, updated.Enforcement),
	})

	writeJSON(w, http.StatusOK, toRulesetDTO(updated))
}

func (s *Server) deleteRuleset(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, _, err := s.loadRepoAndNS(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	rulesetID := r.PathValue("rulesetId")
	if err := s.store.DeleteRuleset(r.Context(), repo.ID, rulesetID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFound(w, "ruleset not found")
			return
		}
		internalError(w, "delete ruleset: "+err.Error())
		return
	}

	_ = audit.Log(audit.Event{
		Actor:     s.resolveActor(r),
		ActorType: "user",
		Action:    "repo.ruleset_deleted",
		Resource:  fmt.Sprintf("%s/%s:ruleset-%s", owner, name, rulesetID),
		ClientIP:  r.RemoteAddr,
		Severity:  audit.SeverityCritical,
		Details:   fmt.Sprintf("Deleted ruleset %s", rulesetID),
	})

	w.WriteHeader(http.StatusNoContent)
}

func toRulesetDTO(rs store.Ruleset) map[string]any {
	dto := map[string]any{
		"id":               rs.ID,
		"name":             rs.Name,
		"description":      rs.Description,
		"enforcement":      rs.Enforcement,
		"kind":             rs.Kind,
		"includedRefNames": rs.IncludedRefNames,
		"excludedRefNames": rs.ExcludedRefNames,
		"rules":            rs.Rules,
		"bypassActors":     rs.BypassActors,
	}
	if dto["includedRefNames"] == nil {
		dto["includedRefNames"] = []string{}
	}
	if dto["excludedRefNames"] == nil {
		dto["excludedRefNames"] = []string{}
	}
	if dto["rules"] == nil {
		dto["rules"] = []store.RulesetRule{}
	}
	if dto["bypassActors"] == nil {
		dto["bypassActors"] = []store.RulesetBypassActor{}
	}
	return dto
}
