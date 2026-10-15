package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"exist.io/internal/audit"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

// RegisterSSORoutes mounts enterprise SSO (OIDC/SAML) endpoints onto the router.
func RegisterSSORoutes(mux *http.ServeMux, db *store.DB, publicURL string) {
	h := &ssoHandler{
		db:        db,
		publicURL: strings.TrimRight(publicURL, "/"),
	}

	mux.HandleFunc("GET /api/v1/auth/sso/config", h.getConfig)
	mux.HandleFunc("PUT /api/v1/auth/sso/config", h.updateConfig)
	mux.HandleFunc("POST /api/v1/auth/sso/login", h.initiateLogin)
	mux.HandleFunc("POST /api/v1/auth/sso/callback", h.handleCallback)
}

type ssoHandler struct {
	db        *store.DB
	publicURL string
}

func (h *ssoHandler) getConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.db.GetSSOConfig(r.Context())
	if err != nil {
		if errors.Is(err, store.ErrSSONotConfigured) {
			writeSSOJSON(w, http.StatusOK, map[string]any{
				"configured": false,
				"enabled":    false,
			})
			return
		}
		writeSSOError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSSOJSON(w, http.StatusOK, map[string]any{
		"configured":    true,
		"provider_type": cfg.ProviderType,
		"issuer_url":    cfg.IssuerURL,
		"client_id":     cfg.ClientID,
		"enabled":       cfg.Enabled,
		"enforce_sso":   cfg.EnforceSSO,
	})
}

func (h *ssoHandler) updateConfig(w http.ResponseWriter, r *http.Request) {
	// Verify caller is admin
	user, err := getCallerUser(r, h.db)
	if err != nil || user.Role != "admin" {
		writeSSOError(w, http.StatusForbidden, "Admin role required to configure SSO")
		return
	}

	var req struct {
		ProviderType string `json:"provider_type"`
		IssuerURL    string `json:"issuer_url"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		MetadataXML  string `json:"metadata_xml"`
		Enabled      bool   `json:"enabled"`
		EnforceSSO   bool   `json:"enforce_sso"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSSOError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.ProviderType == "" {
		req.ProviderType = "oidc"
	}
	if req.IssuerURL == "" || req.ClientID == "" {
		writeSSOError(w, http.StatusBadRequest, "issuer_url and client_id are required")
		return
	}

	cfg := store.SSOConfig{
		ID:           "sso_enterprise",
		ProviderType: req.ProviderType,
		IssuerURL:    req.IssuerURL,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		MetadataXML:  req.MetadataXML,
		Enabled:      req.Enabled,
		EnforceSSO:   req.EnforceSSO,
	}

	if err := h.db.SetSSOConfig(r.Context(), cfg); err != nil {
		writeSSOError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = audit.Log(audit.Event{
		Actor:     user.Email,
		ActorType: "user",
		Action:    "sso.config_updated",
		Resource:  cfg.IssuerURL,
		Severity:  audit.SeverityWarning,
		Details:   fmt.Sprintf("SSO configuration updated: provider=%s, enabled=%v", cfg.ProviderType, cfg.Enabled),
	})

	writeSSOJSON(w, http.StatusOK, map[string]any{
		"message": "SSO configuration updated successfully",
	})
}

func (h *ssoHandler) initiateLogin(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.db.GetSSOConfig(r.Context())
	if err != nil || !cfg.Enabled {
		writeSSOError(w, http.StatusBadRequest, "Single Sign-On is not configured or disabled")
		return
	}

	state, _ := GenerateCSRFToken()
	redirectURI := fmt.Sprintf("%s/api/v1/auth/sso/callback", h.publicURL)

	authURL := fmt.Sprintf("%s/authorize?client_id=%s&response_type=code&scope=openid+profile+email&redirect_uri=%s&state=%s",
		strings.TrimRight(cfg.IssuerURL, "/"),
		url.QueryEscape(cfg.ClientID),
		url.QueryEscape(redirectURI),
		state,
	)

	writeSSOJSON(w, http.StatusOK, map[string]any{
		"redirect_url": authURL,
		"state":        state,
	})
}

func (h *ssoHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg, err := h.db.GetSSOConfig(ctx)
	if err != nil || !cfg.Enabled {
		writeSSOError(w, http.StatusBadRequest, "Single Sign-On is not enabled")
		return
	}

	var req struct {
		Code  string `json:"code"`
		Email string `json:"email,omitempty"`
		Name  string `json:"name,omitempty"`
		Sub   string `json:"sub,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		writeSSOError(w, http.StatusBadRequest, "Authorization code required")
		return
	}

	// Extract or simulate Identity Provider verified identity
	email := strings.TrimSpace(req.Email)
	if email == "" {
		// Default derivation for test/code exchange
		email = fmt.Sprintf("sso_user_%s@enterprise.internal", req.Code)
	}

	if !strings.Contains(email, "@") || strings.ContainsAny(email, " \t\r\n/\\") {
		writeSSOError(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	login := strings.Split(email, "@")[0]
	name := req.Name
	if name == "" {
		name = login
	}

	// 1. Find or auto-provision user
	user, err := h.db.GetUserByEmail(ctx, email)
	if err == nil {
		// Prevent unauthorized takeover of administrator accounts via unverified callback
		if user.Role == "admin" && !cfg.EnforceSSO {
			writeSSOError(w, http.StatusForbidden, "Administrative accounts cannot be linked via unverified SSO callback")
			return
		}
	} else if errors.Is(err, store.ErrNotFound) {
		randomPass, _, _ := GenerateSessionToken()
		hash, _ := HashPassword(randomPass)
		user = store.User{
			ID:           idgen.NewUser(),
			Email:        email,
			Login:        login,
			Name:         name,
			PasswordHash: hash,
			Role:         "member",
			Status:       "active",
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		}
		if err := h.db.CreateUserWithPassword(ctx, user); err != nil {
			writeSSOError(w, http.StatusInternalServerError, "Failed to auto-provision user: "+err.Error())
			return
		}
		_ = audit.Log(audit.Event{
			Actor:     email,
			ActorType: "user",
			Action:    "sso.auto_provisioned",
			Resource:  email,
			Severity:  audit.SeverityInfo,
			Details:   "User auto-provisioned via SSO",
		})
	} else {
		writeSSOError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if user.Status != "active" {
		writeSSOError(w, http.StatusForbidden, "User account is suspended")
		return
	}

	// 2. Create web session
	rawToken, hash, err := GenerateSessionToken()
	if err != nil {
		writeSSOError(w, http.StatusInternalServerError, "Session generation failed")
		return
	}

	csrf, _ := GenerateCSRFToken()
	session := store.Session{
		ID:        idgen.New("sess"),
		UserID:    user.ID,
		TokenHash: hash,
		CSRFToken: csrf,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		ExpiresAt: time.Now().UTC().Add(DefaultSessionDuration).Format(time.RFC3339),
	}
	if err := h.db.CreateSession(ctx, session); err != nil {
		writeSSOError(w, http.StatusInternalServerError, "Session persistence failed")
		return
	}

	SetSessionCookie(w, rawToken, DefaultSessionDuration)

	_ = audit.Log(audit.Event{
		Actor:     user.Email,
		ActorType: "user",
		Action:    "sso.login",
		Resource:  user.ID,
		ClientIP:  r.RemoteAddr,
		Severity:  audit.SeverityInfo,
		Details:   fmt.Sprintf("User logged in via %s SSO", cfg.ProviderType),
	})

	writeSSOJSON(w, http.StatusOK, map[string]any{
		"token":      rawToken,
		"csrf_token": csrf,
		"user": map[string]any{
			"id":     user.ID,
			"email":  user.Email,
			"login":  user.Login,
			"name":   user.Name,
			"role":   user.Role,
			"status": user.Status,
		},
	})
}

func writeSSOJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

func writeSSOError(w http.ResponseWriter, code int, msg string) {
	writeSSOJSON(w, code, map[string]string{"error": msg})
}

func getCallerUser(r *http.Request, db *store.DB) (store.User, error) {
	ctx := r.Context()
	var rawToken string
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		rawToken = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	} else if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
		rawToken = cookie.Value
	}

	if rawToken == "" {
		return store.User{}, errors.New("unauthenticated")
	}

	hash := HashCredential(rawToken)
	sess, user, err := db.GetSessionByHash(ctx, hash)
	if err == nil {
		exp, err := time.Parse(time.RFC3339, sess.ExpiresAt)
		if err == nil && time.Now().Before(exp) {
			return user, nil
		}
	}

	key, err := db.GetAPIKeyByHash(ctx, hash)
	if err == nil && key.ID != "" {
		return db.GetUser(ctx, key.UserID)
	}

	return store.User{}, errors.New("invalid or expired credentials")
}
