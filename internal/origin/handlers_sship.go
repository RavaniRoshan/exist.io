package origin

import (
	"errors"
	"net/http"

	"exist.io/internal/store"
	"exist.io/pkg/scopes"
)

var (
	reqNSSettingsRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.NamespaceSettingsRead},
	}
	reqNSSettingsWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.NamespaceSettingsWrite},
	}
)

func (s *Server) registerSSHAndIP(mux *http.ServeMux) {
	// IP Allowlist operations
	s.register("OriginService_GetInboundIpAllowlist", reqNSSettingsRead)
	s.register("OriginService_UpdateInboundIpAllowlist", reqNSSettingsWrite)
	s.register("OriginService_AddInboundIpAllowlistEntry", reqNSSettingsWrite)
	s.register("OriginService_GetInboundIpAllowlistEntry", reqNSSettingsRead)
	s.register("OriginService_DeleteInboundIpAllowlistEntry", reqNSSettingsWrite)
	s.register("OriginService_UpdateInboundIpAllowlistEntry", reqNSSettingsWrite)
	s.register("OriginService_ReplaceInboundIpAllowlistEntries", reqNSSettingsWrite)

	// SSH CAs operations
	s.register("OriginService_ListSshCertificateAuthorities", reqNSSettingsRead)
	s.register("OriginService_AddSshCertificateAuthority", reqNSSettingsWrite)
	s.register("OriginService_DeleteSshCertificateAuthority", reqNSSettingsWrite)
	s.register("OriginService_SetSshCertificateRequirement", reqNSSettingsWrite)

	// Routes
	mux.Handle("GET /v1/origin/namespaces/{namespaceSlug}/inbound-ip-allowlist",
		s.guard("OriginService_GetInboundIpAllowlist", s.getInboundIpAllowlist))
	mux.Handle("PATCH /v1/origin/namespaces/{namespaceSlug}/inbound-ip-allowlist",
		s.guard("OriginService_UpdateInboundIpAllowlist", s.updateInboundIpAllowlist))
	mux.Handle("POST /v1/origin/namespaces/{namespaceSlug}/inbound-ip-allowlist/entries",
		s.guard("OriginService_AddInboundIpAllowlistEntry", s.addInboundIpAllowlistEntry))
	mux.Handle("GET /v1/origin/namespaces/{namespaceSlug}/inbound-ip-allowlist/entries/{entryId}",
		s.guard("OriginService_GetInboundIpAllowlistEntry", s.getInboundIpAllowlistEntry))
	mux.Handle("DELETE /v1/origin/namespaces/{namespaceSlug}/inbound-ip-allowlist/entries/{entryId}",
		s.guard("OriginService_DeleteInboundIpAllowlistEntry", s.deleteInboundIpAllowlistEntry))
	mux.Handle("PATCH /v1/origin/namespaces/{namespaceSlug}/inbound-ip-allowlist/entries/{entryId}",
		s.guard("OriginService_UpdateInboundIpAllowlistEntry", s.updateInboundIpAllowlistEntry))
	mux.Handle("POST /v1/origin/namespaces/{namespaceSlug}/inbound-ip-allowlist/entries:replace",
		s.guard("OriginService_ReplaceInboundIpAllowlistEntries", s.replaceInboundIpAllowlistEntries))

	mux.Handle("GET /v1/origin/namespaces/{namespaceSlug}/ssh-certificate-authorities",
		s.guard("OriginService_ListSshCertificateAuthorities", s.listSshCertificateAuthorities))
	mux.Handle("POST /v1/origin/namespaces/{namespaceSlug}/ssh-certificate-authorities",
		s.guard("OriginService_AddSshCertificateAuthority", s.addSshCertificateAuthority))
	mux.Handle("DELETE /v1/origin/namespaces/{namespaceSlug}/ssh-certificate-authorities/{certificateAuthorityId}",
		s.guard("OriginService_DeleteSshCertificateAuthority", s.deleteSshCertificateAuthority))
	mux.Handle("POST /v1/origin/namespaces/{namespaceSlug}/ssh-certificate-authorities:setRequirement",
		s.guard("OriginService_SetSshCertificateRequirement", s.setSshCertificateRequirement))
}

// ---------- Handlers: Inbound IP Allowlist ----------

func (s *Server) getInboundIpAllowlist(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	enabled, entries, etag, err := s.store.GetIPAllowlist(r.Context(), ns.ID)
	if err != nil {
		internalError(w, "get ip allowlist: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"enabled": enabled,
		"entries": toIPAllowlistDTOs(entries),
		"etag":    etag,
	})
}

type updateIPAllowlistReq struct {
	Enabled bool `json:"enabled"`
}

func (s *Server) updateInboundIpAllowlist(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	var req updateIPAllowlistReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	if err := s.store.SetIPAllowlistEnabled(r.Context(), ns.ID, req.Enabled); err != nil {
		internalError(w, "set ip allowlist enabled: "+err.Error())
		return
	}

	enabled, entries, etag, _ := s.store.GetIPAllowlist(r.Context(), ns.ID)
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled": enabled,
		"entries": toIPAllowlistDTOs(entries),
		"etag":    etag,
	})
}

type addIPEntryReq struct {
	CIDR        string `json:"cidr"`
	Description string `json:"description,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

func (s *Server) addInboundIpAllowlistEntry(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	var req addIPEntryReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.CIDR == "" {
		badRequest(w, "missing required field: cidr")
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	entry, err := s.store.AddIPAllowlistEntry(r.Context(), ns.ID, req.CIDR, req.Description, enabled)
	if err != nil {
		internalError(w, "add ip allowlist entry: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toIPAllowlistEntryDTO(entry))
}

func (s *Server) getInboundIpAllowlistEntry(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	entryID := r.PathValue("entryId")
	entry, err := s.store.GetIPAllowlistEntry(r.Context(), ns.ID, entryID)
	if err != nil {
		notFound(w, "entry not found")
		return
	}

	writeJSON(w, http.StatusOK, toIPAllowlistEntryDTO(entry))
}

type updateIPEntryReq struct {
	CIDR        string `json:"cidr,omitempty"`
	Description string `json:"description,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

func (s *Server) updateInboundIpAllowlistEntry(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	entryID := r.PathValue("entryId")
	var req updateIPEntryReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	updated, err := s.store.UpdateIPAllowlistEntry(r.Context(), ns.ID, entryID, req.CIDR, req.Description, req.Enabled)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFound(w, "entry not found")
			return
		}
		internalError(w, "update entry: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toIPAllowlistEntryDTO(updated))
}

func (s *Server) deleteInboundIpAllowlistEntry(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	entryID := r.PathValue("entryId")
	if err := s.store.DeleteIPAllowlistEntry(r.Context(), ns.ID, entryID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFound(w, "entry not found")
			return
		}
		internalError(w, "delete entry: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type replaceIPEntriesReq struct {
	Entries    []addIPEntryReq `json:"entries"`
	ETag       string          `json:"etag,omitempty"`
	AllowEmpty bool            `json:"allowEmpty,omitempty"`
}

func (s *Server) replaceInboundIpAllowlistEntries(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	var req replaceIPEntriesReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	_, existing, currentETag, err := s.store.GetIPAllowlist(r.Context(), ns.ID)
	if err != nil {
		internalError(w, "get ip allowlist: "+err.Error())
		return
	}

	if req.ETag != "" && req.ETag != currentETag {
		conflict(w, "etag mismatch")
		return
	}

	// Delete all existing
	for _, e := range existing {
		_ = s.store.DeleteIPAllowlistEntry(r.Context(), ns.ID, e.ID)
	}

	addedCount := 0
	for _, ne := range req.Entries {
		enabled := true
		if ne.Enabled != nil {
			enabled = *ne.Enabled
		}
		_, _ = s.store.AddIPAllowlistEntry(r.Context(), ns.ID, ne.CIDR, ne.Description, enabled)
		addedCount++
	}

	enabled, entries, newETag, _ := s.store.GetIPAllowlist(r.Context(), ns.ID)
	writeJSON(w, http.StatusOK, map[string]any{
		"allowlist": map[string]any{
			"enabled": enabled,
			"entries": toIPAllowlistDTOs(entries),
			"etag":    newETag,
		},
		"addedCount":     addedCount,
		"updatedCount":   0,
		"removedCount":   len(existing),
		"unchangedCount": 0,
	})
}

// ---------- Handlers: SSH Certificate Authorities ----------

func (s *Server) listSshCertificateAuthorities(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	cas, err := s.store.ListSSHCAs(r.Context(), ns.ID)
	if err != nil {
		internalError(w, "list ssh cas: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, ca := range cas {
		dtos = append(dtos, toSSHCaDTO(ca))
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"certificateAuthorities": dtos,
	})
}

type addSSHCaReq struct {
	Name      string `json:"name"`
	PublicKey string `json:"publicKey"`
}

func (s *Server) addSshCertificateAuthority(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	var req addSSHCaReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Name == "" || req.PublicKey == "" {
		badRequest(w, "missing required fields: name, publicKey")
		return
	}

	ca, err := s.store.AddSSHCA(r.Context(), ns.ID, req.Name, req.PublicKey)
	if err != nil {
		badRequest(w, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toSSHCaDTO(ca))
}

func (s *Server) deleteSshCertificateAuthority(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	caID := r.PathValue("certificateAuthorityId")
	if err := s.store.DeleteSSHCA(r.Context(), ns.ID, caID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFound(w, "certificate authority not found")
			return
		}
		internalError(w, "delete ssh ca: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type setSSHCAReq struct {
	RequireCertificates bool `json:"requireCertificates"`
}

func (s *Server) setSshCertificateRequirement(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("namespaceSlug")
	ns, err := s.store.GetNamespaceBySlug(r.Context(), slug)
	if err != nil {
		notFound(w, "namespace not found")
		return
	}

	var req setSSHCAReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	if err := s.store.SetSSHCARequirement(r.Context(), ns.ID, req.RequireCertificates); err != nil {
		internalError(w, "set ssh ca requirement: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"requireCertificates": req.RequireCertificates,
	})
}

// ---------- Helper DTOs ----------

func toIPAllowlistDTOs(entries []store.IPAllowlistEntry) []map[string]any {
	var dtos []map[string]any
	for _, e := range entries {
		dtos = append(dtos, toIPAllowlistEntryDTO(e))
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}
	return dtos
}

func toIPAllowlistEntryDTO(e store.IPAllowlistEntry) map[string]any {
	return map[string]any{
		"id":          e.ID,
		"cidr":        e.CIDR,
		"description": e.Description,
		"enabled":     e.Enabled,
		"createdAt":   e.CreatedAt,
	}
}

func toSSHCaDTO(ca store.SSHCertificateAuthority) map[string]any {
	return map[string]any{
		"id":          ca.ID,
		"name":        ca.Name,
		"keyType":     ca.KeyType,
		"publicKey":   ca.PublicKey,
		"fingerprint": ca.Fingerprint,
		"createdAt":   ca.CreatedAt,
	}
}
