package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"exist.io/internal/store"
)

func TestWebAuthLifecycle(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// 1. Register first user (should automatically become "admin")
	regPayload := `{"email": "founder@exist.io", "login": "founder", "name": "Lead Architect", "password": "SuperSecretPassword123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(regPayload))
	rec := httptest.NewRecorder()
	RegisterHandler(db).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register status = %d; body=%s", rec.Code, rec.Body.String())
	}

	var regResp struct {
		User         UserProfileDTO `json:"user"`
		SessionToken string         `json:"sessionToken"`
		CSRFToken    string         `json:"csrfToken"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("unmarshal register: %v", err)
	}
	if regResp.User.Role != "admin" {
		t.Errorf("expected first user to be admin, got %s", regResp.User.Role)
	}
	if regResp.SessionToken == "" || regResp.CSRFToken == "" {
		t.Errorf("expected non-empty tokens in register response")
	}

	// 2. Check Session
	sessReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	sessReq.AddCookie(&http.Cookie{Name: SessionCookieName, Value: regResp.SessionToken})
	sessRec := httptest.NewRecorder()
	SessionHandler(db).ServeHTTP(sessRec, sessReq)

	if sessRec.Code != http.StatusOK {
		t.Fatalf("session check status = %d; body=%s", sessRec.Code, sessRec.Body.String())
	}

	// 3. Login with correct password
	loginPayload := `{"email": "founder@exist.io", "password": "SuperSecretPassword123!"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(loginPayload))
	loginRec := httptest.NewRecorder()
	LoginHandler(db).ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d; body=%s", loginRec.Code, loginRec.Body.String())
	}

	// 4. Login with bad password
	badLoginPayload := `{"email": "founder@exist.io", "password": "wrong-password"}`
	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(badLoginPayload))
	badRec := httptest.NewRecorder()
	LoginHandler(db).ServeHTTP(badRec, badReq)

	if badRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized on bad password, got %d", badRec.Code)
	}

	// 5. Admin List Users
	adminReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	adminRec := httptest.NewRecorder()
	AdminListUsersHandler(db).ServeHTTP(adminRec, adminReq)

	if adminRec.Code != http.StatusOK {
		t.Fatalf("admin list users status = %d", adminRec.Code)
	}
	var adminListResp struct {
		Users []UserProfileDTO `json:"users"`
	}
	_ = json.Unmarshal(adminRec.Body.Bytes(), &adminListResp)
	if len(adminListResp.Users) != 1 || adminListResp.Users[0].Login != "founder" {
		t.Errorf("expected 1 user named founder, got %+v", adminListResp.Users)
	}

	// 6. Logout
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq.AddCookie(&http.Cookie{Name: SessionCookieName, Value: regResp.SessionToken})
	logoutRec := httptest.NewRecorder()
	LogoutHandler(db).ServeHTTP(logoutRec, logoutReq)

	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", logoutRec.Code)
	}
}
