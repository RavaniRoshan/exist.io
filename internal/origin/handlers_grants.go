package origin

import (
	"net/http"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"exist.io/pkg/scopes"
)

// registerGrants wires the grant routes (Task 7). Grants use POST/DELETE on the
// collection with the principal carried in the request body (Origin does not
// put the principal in the path).
func (s *Server) registerGrants(mux *http.ServeMux) {
	s.register("OriginService_ListNamespaceGrants", scopes.Requirements{TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser}, Scopes: []string{scopes.NamespaceSettingsRead}})
	s.register("OriginService_UpsertNamespaceGrant", scopes.Requirements{TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser}, Scopes: []string{scopes.NamespaceSettingsWrite}})
	s.register("OriginService_DeleteNamespaceGrant", scopes.Requirements{TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser}, Scopes: []string{scopes.NamespaceSettingsWrite}})
	s.register("OriginService_ListRepositoryGrants", scopes.Requirements{TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser}, Scopes: []string{scopes.RepositorySettingsRead}})
	s.register("OriginService_UpsertRepositoryGrant", scopes.Requirements{TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser}, Scopes: []string{scopes.RepositorySettingsWrite}})
	s.register("OriginService_DeleteRepositoryGrant", scopes.Requirements{TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser}, Scopes: []string{scopes.RepositorySettingsWrite}})

	mux.Handle("GET /v1/origin/namespaces/{namespaceSlug}/grants", s.guard("OriginService_ListNamespaceGrants", s.listNamespaceGrants))
	mux.Handle("POST /v1/origin/namespaces/{namespaceSlug}/grants", s.guard("OriginService_UpsertNamespaceGrant", s.upsertNamespaceGrant))
	mux.Handle("DELETE /v1/origin/namespaces/{namespaceSlug}/grants", s.guard("OriginService_DeleteNamespaceGrant", s.deleteNamespaceGrant))
	mux.Handle("GET /v1/origin/repos/{ownerSlug}/{repoName}/grants", s.guard("OriginService_ListRepositoryGrants", s.listRepositoryGrants))
	mux.Handle("POST /v1/origin/repos/{ownerSlug}/{repoName}/grants", s.guard("OriginService_UpsertRepositoryGrant", s.upsertRepositoryGrant))
	mux.Handle("DELETE /v1/origin/repos/{ownerSlug}/{repoName}/grants", s.guard("OriginService_DeleteRepositoryGrant", s.deleteRepositoryGrant))
}

// grantRequest is the shared upsert/delete body. Exactly one principal is set.
type grantRequest struct {
	Permission string        `json:"permission"`
	User       *userActorDTO `json:"user"`
	Group      *groupDTO     `json:"group"`
	TeamGroup  *teamGroupDTO `json:"teamGroup"`
}

type userActorDTO struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Handle string `json:"handle"`
}

type groupDTO struct {
	ID string `json:"id"`
}

type teamGroupDTO struct {
	Kind string `json:"kind"` // members | admins
}

// grantDTO renders a grant on the wire. The permission keeps Origin's exact
// enum value (lowercase for repo grants, PERMISSION_* for namespace grants).
type grantDTO struct {
	Permission string        `json:"permission"`
	User       *userActorDTO `json:"user,omitempty"`
	Group      *groupDTO     `json:"group,omitempty"`
	TeamGroup  *teamGroupDTO `json:"teamGroup,omitempty"`
}

// resolvePrincipal turns a request principal into a store principal, resolving
// user_email→user_ and group_name→grp_ against exist.io's own tables (our
// replacement for Origin's Cursor admin bridges). teamGroup needs no resolution.
func (s *Server) resolvePrincipal(r *http.Request, namespaceID string, req grantRequest) (store.GrantPrincipal, error) {
	switch {
	case req.TeamGroup != nil:
		if req.TeamGroup.Kind != "members" && req.TeamGroup.Kind != "admins" {
			return store.GrantPrincipal{}, errBadPrincipal
		}
		return store.GrantPrincipal{Kind: "teamGroup", TeamGroup: req.TeamGroup.Kind}, nil
	case req.Group != nil:
		if req.Group.ID != "" {
			return store.GrantPrincipal{Kind: "group", GroupID: req.Group.ID}, nil
		}
		return store.GrantPrincipal{}, errBadPrincipal
	case req.User != nil:
		if req.User.ID != "" {
			return store.GrantPrincipal{Kind: "user", UserID: req.User.ID}, nil
		}
		if req.User.Email != "" {
			id, err := s.store.UpsertUserByEmail(r.Context(), req.User.Email, idgen.NewUser)
			if err != nil {
				return store.GrantPrincipal{}, err
			}
			return store.GrantPrincipal{Kind: "user", UserID: id}, nil
		}
		return store.GrantPrincipal{}, errBadPrincipal
	default:
		return store.GrantPrincipal{}, errBadPrincipal
	}
}

// errBadPrincipal marks a request that names no valid principal.
var errBadPrincipal = &badPrincipalError{}

type badPrincipalError struct{}

func (*badPrincipalError) Error() string {
	return "origin: exactly one principal (user, group, or teamGroup) is required"
}

func (s *Server) listNamespaceGrants(w http.ResponseWriter, r *http.Request) {
	ns, err := s.store.GetNamespaceBySlug(r.Context(), r.PathValue("namespaceSlug"))
	if err != nil {
		notFound(w, "namespace not found")
		return
	}
	grants, err := s.store.ListNamespaceGrants(r.Context(), ns.ID)
	if err != nil {
		internalError(w, "list grants: "+err.Error())
		return
	}
	out := make([]grantDTO, 0, len(grants))
	for _, g := range grants {
		out = append(out, s.grantDTO(g, false))
	}
	writeJSON(w, http.StatusOK, map[string]any{"grants": out, "nextPageToken": ""})
}

func (s *Server) upsertNamespaceGrant(w http.ResponseWriter, r *http.Request) {
	ns, err := s.store.GetNamespaceBySlug(r.Context(), r.PathValue("namespaceSlug"))
	if err != nil {
		notFound(w, "namespace not found")
		return
	}
	var req grantRequest
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if !validNamespacePermission(req.Permission) {
		badRequest(w, "invalid namespace permission: "+req.Permission)
		return
	}
	principal, err := s.resolvePrincipal(r, ns.ID, req)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	g, err := s.store.UpsertGrant(r.Context(), store.Grant{
		ID: idgenGrant(), Scope: "namespace", NamespaceID: ns.ID,
		PrincipalKind: principal.Kind, UserID: principal.UserID, GroupID: principal.GroupID, TeamGroup: principal.TeamGroup,
		Permission: req.Permission,
	})
	if err != nil {
		internalError(w, "upsert grant: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.grantDTO(g, false))
}

func (s *Server) deleteNamespaceGrant(w http.ResponseWriter, r *http.Request) {
	ns, err := s.store.GetNamespaceBySlug(r.Context(), r.PathValue("namespaceSlug"))
	if err != nil {
		notFound(w, "namespace not found")
		return
	}
	var req grantRequest
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	principal, err := s.resolvePrincipal(r, ns.ID, req)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	if err := s.store.DeleteGrant(r.Context(), "namespace", ns.ID, "", principal); err != nil {
		if err == store.ErrNotFound {
			notFound(w, "grant not found")
			return
		}
		internalError(w, "delete grant: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) listRepositoryGrants(w http.ResponseWriter, r *http.Request) {
	repo, err := s.store.GetRepo(r.Context(), r.PathValue("ownerSlug"), r.PathValue("repoName"))
	if err != nil {
		notFound(w, "repository not found")
		return
	}
	grants, err := s.store.ListRepositoryGrants(r.Context(), repo.ID)
	if err != nil {
		internalError(w, "list grants: "+err.Error())
		return
	}
	out := make([]grantDTO, 0, len(grants))
	for _, g := range grants {
		out = append(out, s.grantDTO(g, true))
	}
	writeJSON(w, http.StatusOK, map[string]any{"grants": out, "nextPageToken": "", "repository": repoRefDTO(repo)})
}

func (s *Server) upsertRepositoryGrant(w http.ResponseWriter, r *http.Request) {
	repo, err := s.store.GetRepo(r.Context(), r.PathValue("ownerSlug"), r.PathValue("repoName"))
	if err != nil {
		notFound(w, "repository not found")
		return
	}
	var req grantRequest
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if !validRepoPermission(req.Permission) {
		badRequest(w, "invalid repository permission: "+req.Permission)
		return
	}
	principal, err := s.resolvePrincipal(r, repo.NamespaceID, req)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	g, err := s.store.UpsertGrant(r.Context(), store.Grant{
		ID: idgenGrant(), Scope: "repository", NamespaceID: repo.NamespaceID, RepoID: repo.ID,
		PrincipalKind: principal.Kind, UserID: principal.UserID, GroupID: principal.GroupID, TeamGroup: principal.TeamGroup,
		Permission: req.Permission,
	})
	if err != nil {
		internalError(w, "upsert grant: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.grantDTO(g, true))
}

func (s *Server) deleteRepositoryGrant(w http.ResponseWriter, r *http.Request) {
	repo, err := s.store.GetRepo(r.Context(), r.PathValue("ownerSlug"), r.PathValue("repoName"))
	if err != nil {
		notFound(w, "repository not found")
		return
	}
	var req grantRequest
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	principal, err := s.resolvePrincipal(r, repo.NamespaceID, req)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	if err := s.store.DeleteGrant(r.Context(), "repository", repo.NamespaceID, repo.ID, principal); err != nil {
		if err == store.ErrNotFound {
			notFound(w, "grant not found")
			return
		}
		internalError(w, "delete grant: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

// grantDTO converts a store.Grant to the wire DTO. Permission is stored exactly
// as Origin's wire value, so it passes through unchanged.
func (s *Server) grantDTO(g store.Grant, repo bool) grantDTO {
	dto := grantDTO{Permission: g.Permission}
	switch g.PrincipalKind {
	case "user":
		dto.User = &userActorDTO{ID: g.UserID}
	case "group":
		dto.Group = &groupDTO{ID: g.GroupID}
	case "teamGroup":
		dto.TeamGroup = &teamGroupDTO{Kind: g.TeamGroup}
	}
	return dto
}

func repoRefDTO(r store.Repo) map[string]any {
	return map[string]any{"ownerSlug": "", "name": r.Name, "fullName": r.FullName}
}

func validNamespacePermission(p string) bool {
	switch p {
	case "PERMISSION_READ", "PERMISSION_CONTRIBUTOR", "PERMISSION_WRITE", "PERMISSION_ADMIN", "PERMISSION_CUSTOM":
		return true
	}
	return false
}

func validRepoPermission(p string) bool {
	switch p {
	case "read", "write", "admin", "custom":
		return true
	}
	return false
}

// idgenGrant returns a fresh internal grant id (grants are not a prefixed Origin
// entity exposed by id, so a ULID without a resource prefix suffices).
func idgenGrant() string { return idgen.New("grant") }
