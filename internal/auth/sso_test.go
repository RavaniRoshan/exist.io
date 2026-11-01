package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"exist.io/internal/store"
)

func TestSSOEndpoints(t *testing.T) {
	dataDir := t.TempDir()
	db, err := store.Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Seed admin user and session
	adminPass, adminHash, _ := GenerateSessionToken()
	_ = adminPass
	adminUser := store.User{
		ID:           "user_admin",
		Email:        "admin@exist.io",
		Login:        "admin",
		Name:         "Admin",
		PasswordHash: adminHash,
		Role:         "admin",
		Status:       "active",
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	if err := db.CreateUserWithPassword(ctx, adminUser); err != nil {
		t.Fatalf("CreateUserWithPassword: %v", err)
	}

	adminSessionRaw, adminSessionHash, _ := GenerateSessionToken()
	if err := db.CreateSession(ctx, store.Session{
		ID:        "sess_admin",
		UserID:    adminUser.ID,
		TokenHash: adminSessionHash,
		ExpiresAt: time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	mux := http.NewServeMux()
	RegisterSSORoutes(mux, db, "https://git.enterprise.internal")

	// 2. GET SSO config unconfigured
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/sso/config", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get config status = %d, want 200", getRec.Code)
	}

	// 3. PUT SSO config with admin session cookie
	updateBody := `{
		"provider_type": "oidc",
		"issuer_url": "https://auth.enterprise.internal/oauth2/v1",
		"client_id": "client_enterprise_xyz",
		"client_secret": "secret_abc",
		"enabled": true,
		"enforce_sso": false
	}`
	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/auth/sso/config", strings.NewReader(updateBody))
	putReq.Header.Set("Authorization", "Bearer "+adminSessionRaw)
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	mux.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("update config status = %d, want 200: %s", putRec.Code, putRec.Body.String())
	}

	// 4. POST initiate login
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sso/login", strings.NewReader(`{}`))
	loginRec := httptest.NewRecorder()
	mux.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("initiate login status = %d, want 200", loginRec.Code)
	}
	if !strings.Contains(loginRec.Body.String(), "client_enterprise_xyz") {
		t.Fatalf("expected client_id in redirect url, got: %s", loginRec.Body.String())
	}

	// 5. Callback fails closed: no identity from client input, no session.
	callbackBody := `{
		"code": "auth_code_999",
		"email": "employee@enterprise.internal",
		"name": "Jane Employee"
	}`
	cbReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sso/callback", strings.NewReader(callbackBody))
	cbReq.Header.Set("Content-Type", "application/json")
	cbRec := httptest.NewRecorder()
	mux.ServeHTTP(cbRec, cbReq)
	if cbRec.Code != http.StatusNotImplemented {
		t.Fatalf("callback status = %d, want 501: %s", cbRec.Code, cbRec.Body.String())
	}
	for _, c := range cbRec.Result().Cookies() {
		if c.Name == SessionCookieName {
			t.Fatalf("callback set a session cookie for unverified input")
		}
	}
	if _, err := db.GetUserByEmail(ctx, "employee@enterprise.internal"); err == nil {
		t.Fatalf("callback auto-provisioned a user from client input")
	}

	// 6. An admin cannot be linked or taken over through the callback.
	adminBody := `{"code":"auth_code_1000","email":"admin@exist.io"}`
	adminReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sso/callback", strings.NewReader(adminBody))
	adminReq.Header.Set("Content-Type", "application/json")
	adminRec := httptest.NewRecorder()
	mux.ServeHTTP(adminRec, adminReq)
	if adminRec.Code != http.StatusNotImplemented {
		t.Fatalf("admin callback status = %d, want 501", adminRec.Code)
	}
}
