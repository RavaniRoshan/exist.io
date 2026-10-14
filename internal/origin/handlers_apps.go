package origin

import (
	"errors"
	"net/http"
	"time"

	"exist.io/internal/auth"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"exist.io/pkg/scopes"
)

var (
	reqAppMetadataRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeApp, scopes.TypeUser},
		Scopes:     []string{scopes.AppMetadataRead},
		Ambient:    true,
	}
	reqAppMetadataWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeApp, scopes.TypeUser},
		Scopes:     []string{scopes.AppMetadataWrite},
		Ambient:    true,
	}
	reqInstMetadataRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.InstallationMetadataRead},
		Ambient:    true,
	}
	reqNSAppsRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeUser, scopes.TypeInstallation},
		Scopes:     []string{scopes.NamespaceAppsRead},
	}
	reqNSAppsCreate = scopes.Requirements{
		TokenTypes: []string{scopes.TypeUser},
		Scopes:     []string{scopes.NamespaceAppsCreate},
	}
	reqAppSettingsWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeUser, scopes.TypeApp},
		Scopes:     []string{scopes.AppSettingsWrite},
	}
	reqNSInstallationsWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeUser},
		Scopes:     []string{scopes.NamespaceInstallationsWrite},
	}
)

func (s *Server) registerApps(mux *http.ServeMux) {
	// Operations
	s.register("OriginService_GetAuthenticatedApp", reqAppMetadataRead)
	s.register("OriginService_ListAppInstallations", reqAppMetadataRead)
	s.register("OriginService_GetAppInstallation", reqAppMetadataRead)
	s.register("OriginService_DeleteAppInstallation", reqAppMetadataWrite)
	s.register("OriginService_CreateInstallationAccessToken", reqAppMetadataRead)
	s.register("OriginService_CreateInstallationUserToken", reqAppMetadataRead)
	s.register("OriginService_ListAppInstallationRepositories", reqInstMetadataRead)
	s.register("OriginService_AddAppInstallationRepositories", reqNSInstallationsWrite)
	s.register("OriginService_ListNamespaceApps", reqNSAppsRead)
	s.register("OriginService_CreateApp", reqNSAppsCreate)
	s.register("OriginService_GetApp", reqNSAppsRead)
	s.register("OriginService_UpdateApp", reqAppSettingsWrite)
	s.register("OriginService_AddAppSigningKey", reqAppSettingsWrite)
	s.register("OriginService_RevokeAppSigningKey", reqAppSettingsWrite)

	// Routes
	mux.Handle("GET /v1/origin/app",
		s.guard("OriginService_GetAuthenticatedApp", s.getAuthenticatedApp))
	mux.Handle("GET /v1/origin/app/installations",
		s.guard("OriginService_ListAppInstallations", s.listAppInstallations))
	mux.Handle("GET /v1/origin/app/installations/{installationId}",
		s.guard("OriginService_GetAppInstallation", s.getAppInstallation))
	mux.Handle("DELETE /v1/origin/app/installations/{installationId}",
		s.guard("OriginService_DeleteAppInstallation", s.deleteAppInstallation))
	mux.Handle("POST /v1/origin/app/installations/{installationId}/access_tokens",
		s.guard("OriginService_CreateInstallationAccessToken", s.createInstallationAccessToken))
	mux.Handle("POST /v1/origin/app/installations/{installationId}/user_access_tokens",
		s.guard("OriginService_CreateInstallationUserToken", s.createInstallationUserToken))

	mux.Handle("GET /v1/origin/installation/repos",
		s.guard("OriginService_ListAppInstallationRepositories", s.listAppInstallationRepositories))
	mux.Handle("POST /v1/origin/namespaces/{namespaceSlug}/installations/{installationId}/repos",
		s.guard("OriginService_AddAppInstallationRepositories", s.addAppInstallationRepositories))

	mux.Handle("GET /v1/origin/namespaces/{namespaceSlug}/apps",
		s.guard("OriginService_ListNamespaceApps", s.listNamespaceApps))
	mux.Handle("POST /v1/origin/namespaces/{namespaceSlug}/apps",
		s.guard("OriginService_CreateApp", s.createApp))
	mux.Handle("GET /v1/origin/apps/{appId}",
		s.guard("OriginService_GetApp", s.getApp))
	mux.Handle("PATCH /v1/origin/apps/{appId}",
		s.guard("OriginService_UpdateApp", s.updateApp))
	mux.Handle("POST /v1/origin/apps/{appId}/signing_keys",
		s.guard("OriginService_AddAppSigningKey", s.addAppSigningKey))
	mux.Handle("DELETE /v1/origin/apps/{appId}/signing_keys/{kid}",
		s.guard("OriginService_RevokeAppSigningKey", s.revokeAppSigningKey))
}

func (s *Server) getAuthenticatedApp(w http.ResponseWriter, r *http.Request) {
	appID := s.getActingAppID(r)
	app, err := s.store.GetApp(r.Context(), appID)
	if err != nil {
		notFound(w, "app not found")
		return
	}
	ns, _ := s.store.GetNamespace(r.Context(), app.NamespaceID)
	writeJSON(w, http.StatusOK, toAppDTO(app, ns.Slug))
}

func (s *Server) listAppInstallations(w http.ResponseWriter, r *http.Request) {
	appID := s.getActingAppID(r)
	insts, err := s.store.ListAppInstallations(r.Context(), appID)
	if err != nil {
		internalError(w, "list installations: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, inst := range insts {
		ns, _ := s.store.GetNamespace(r.Context(), inst.NamespaceID)
		dtos = append(dtos, toInstallationDTO(inst, ns))
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"installations": dtos,
		"nextPageToken": "",
	})
}

func (s *Server) getAppInstallation(w http.ResponseWriter, r *http.Request) {
	instID := r.PathValue("installationId")
	inst, err := s.store.GetInstallation(r.Context(), instID)
	if err != nil {
		notFound(w, "installation not found")
		return
	}
	ns, _ := s.store.GetNamespace(r.Context(), inst.NamespaceID)
	writeJSON(w, http.StatusOK, toInstallationDTO(inst, ns))
}

func (s *Server) deleteAppInstallation(w http.ResponseWriter, r *http.Request) {
	instID := r.PathValue("installationId")
	if err := s.store.DeleteInstallation(r.Context(), instID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFound(w, "installation not found")
			return
		}
		internalError(w, "delete installation: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createInstallationAccessToken(w http.ResponseWriter, r *http.Request) {
	instID := r.PathValue("installationId")
	inst, err := s.store.GetInstallation(r.Context(), instID)
	if err != nil {
		notFound(w, "installation not found")
		return
	}
	_ = inst

	token := idgen.NewInstallToken()
	expiresAt := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	writeJSON(w, http.StatusOK, map[string]any{
		"token":     token,
		"expiresAt": expiresAt,
	})
}

func (s *Server) createInstallationUserToken(w http.ResponseWriter, r *http.Request) {
	token := idgen.NewInstallToken()
	expiresAt := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	writeJSON(w, http.StatusOK, map[string]any{
		"token":     token,
		"expiresAt": expiresAt,
	})
}

func (s *Server) listAppInstallationRepositories(w http.ResponseWriter, r *http.Request) {
	instID := ""
	if tok, ok := auth.ContextToken(r.Context()); ok {
		instID = tok.Subject()
	}
	repos, err := s.store.ListInstallationRepos(r.Context(), instID)
	if err != nil {
		internalError(w, "list repos: "+err.Error())
		return
	}

	var dtos []repoDTO
	for _, repo := range repos {
		ns, _ := s.store.GetNamespace(r.Context(), repo.NamespaceID)
		dtos = append(dtos, s.toRepoDTO(repo, ns))
	}
	if dtos == nil {
		dtos = []repoDTO{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"repositories":  dtos,
		"nextPageToken": "",
	})
}

type addInstallationReposReq struct {
	RepoIDs []string `json:"repoIds"`
}

func (s *Server) addAppInstallationRepositories(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	instID := r.PathValue("installationId")
	inst, err := s.store.GetInstallation(r.Context(), instID)
	if err != nil || inst.NamespaceID != ns.ID {
		notFound(w, "installation not found")
		return
	}

	var req addInstallationReposReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	_ = s.store.AddInstallationRepos(r.Context(), instID, req.RepoIDs)
	writeJSON(w, http.StatusOK, toInstallationDTO(inst, ns))
}

func (s *Server) listNamespaceApps(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	apps, err := s.store.ListNamespaceApps(r.Context(), ns.ID)
	if err != nil {
		internalError(w, "list apps: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, app := range apps {
		dtos = append(dtos, toAppDTO(app, slug))
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"apps": dtos,
	})
}

type createAppReq struct {
	DisplayName   string   `json:"displayName"`
	Description   string   `json:"description,omitempty"`
	WebsiteURL    string   `json:"websiteUrl,omitempty"`
	WebhookURL    string   `json:"webhookUrl,omitempty"`
	Events        []string `json:"events,omitempty"`
	DefaultScopes []string `json:"defaultScopes,omitempty"`
}

func (s *Server) createApp(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	var req createAppReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.DisplayName == "" {
		badRequest(w, "displayName is required")
		return
	}

	appStore := store.App{
		NamespaceID:   ns.ID,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		WebsiteURL:    req.WebsiteURL,
		WebhookURL:    req.WebhookURL,
		Events:        req.Events,
		DefaultScopes: req.DefaultScopes,
	}

	created, err := s.store.CreateApp(r.Context(), appStore)
	if err != nil {
		internalError(w, "create app: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toAppDTO(created, slug))
}

func (s *Server) getApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appId")
	app, err := s.store.GetApp(r.Context(), appID)
	if err != nil {
		notFound(w, "app not found")
		return
	}
	ns, _ := s.store.GetNamespace(r.Context(), app.NamespaceID)
	writeJSON(w, http.StatusOK, toAppDTO(app, ns.Slug))
}

func (s *Server) updateApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appId")
	existing, err := s.store.GetApp(r.Context(), appID)
	if err != nil {
		notFound(w, "app not found")
		return
	}

	var req createAppReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	appStore := store.App{
		ID:            existing.ID,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		WebsiteURL:    req.WebsiteURL,
		WebhookURL:    req.WebhookURL,
		Events:        req.Events,
		DefaultScopes: req.DefaultScopes,
	}

	updated, err := s.store.UpdateApp(r.Context(), appStore)
	if err != nil {
		internalError(w, "update app: "+err.Error())
		return
	}
	ns, _ := s.store.GetNamespace(r.Context(), updated.NamespaceID)
	writeJSON(w, http.StatusOK, toAppDTO(updated, ns.Slug))
}

type addSigningKeyReq struct {
	KID       string `json:"kid"`
	PublicKey string `json:"publicKey"`
}

func (s *Server) addAppSigningKey(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appId")
	_, err := s.store.GetApp(r.Context(), appID)
	if err != nil {
		notFound(w, "app not found")
		return
	}

	var req addSigningKeyReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.KID == "" || req.PublicKey == "" {
		badRequest(w, "kid and publicKey are required")
		return
	}

	key, err := s.store.AddAppSigningKey(r.Context(), appID, req.KID, req.PublicKey)
	if err != nil {
		internalError(w, "add signing key: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"signingKey": map[string]any{
			"kid":       key.KID,
			"publicKey": key.PublicKey,
			"createdAt": key.CreatedAt,
		},
	})
}

func (s *Server) revokeAppSigningKey(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appId")
	kid := r.PathValue("kid")

	if err := s.store.RevokeAppSigningKey(r.Context(), appID, kid); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFound(w, "key not found")
			return
		}
		internalError(w, "revoke key: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toAppDTO(a store.App, namespaceSlug string) map[string]any {
	dto := map[string]any{
		"id":                      a.ID,
		"displayName":             a.DisplayName,
		"namespaceSlug":           namespaceSlug,
		"createdAt":               a.CreatedAt,
		"updatedAt":               a.UpdatedAt,
		"installationRedirectUris": []string{},
	}
	if a.Description != "" {
		dto["description"] = a.Description
	}
	if a.WebsiteURL != "" {
		dto["websiteUrl"] = a.WebsiteURL
	}
	if a.WebhookURL != "" {
		dto["webhookUrl"] = a.WebhookURL
	}
	if len(a.Events) > 0 {
		dto["events"] = a.Events
	} else {
		dto["events"] = []string{}
	}
	if len(a.DefaultScopes) > 0 {
		dto["defaultScopes"] = a.DefaultScopes
	} else {
		dto["defaultScopes"] = []string{}
	}
	return dto
}

func toInstallationDTO(inst store.Installation, ns store.Namespace) map[string]any {
	dto := map[string]any{
		"id":    inst.ID,
		"appId": inst.AppID,
		"target": map[string]any{
			"id":   ns.ID,
			"slug": ns.Slug,
			"type": ns.Type,
		},
		"repoSelectionMode": "selected",
		"createdAt":         inst.CreatedAt,
		"updatedAt":         inst.UpdatedAt,
	}
	if len(inst.GrantedScopes) > 0 {
		dto["scopes"] = inst.GrantedScopes
	} else {
		dto["scopes"] = []string{}
	}
	return dto
}
