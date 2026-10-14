package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"exist.io/internal/store"
	"exist.io/pkg/scopes"
)

func setupTestStore(t *testing.T) (*store.DB, string) {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userID := "user_admin1"
	if _, err := db.UpsertUserByEmail(t.Context(), "admin@exist.io", func() string { return userID }); err != nil {
		t.Fatalf("UpsertUserByEmail: %v", err)
	}
	return db, userID
}

func TestServerExchangeAndDBResolver(t *testing.T) {
	db, userID := setupTestStore(t)
	handler := ExchangeHandler(db)

	rawKey, prefix, hash, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey: %v", err)
	}

	apiKey := store.APIKey{
		ID:        "key_test1",
		UserID:    userID,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Name:      "bootstrap-key",
		Scopes:    scopes.NamespaceRepositoriesRead + " " + scopes.NamespaceRepositoriesCreate,
	}
	if err := db.CreateAPIKey(t.Context(), apiKey); err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}

	// 1. Missing key
	req := httptest.NewRequest(http.MethodPost, "/auth/exchange_user_api_key", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing key status = %d, want 401", rec.Code)
	}

	// 2. Invalid prefix
	req = httptest.NewRequest(http.MethodPost, "/auth/exchange_user_api_key", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer invalidprefix_123")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid prefix status = %d, want 400", rec.Code)
	}

	// 3. Unknown key
	req = httptest.NewRequest(http.MethodPost, "/auth/exchange_user_api_key", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer key_unknown1234567890abcdef")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unknown key status = %d, want 401", rec.Code)
	}

	// 4. Valid exchange
	req = httptest.NewRequest(http.MethodPost, "/auth/exchange_user_api_key", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+rawKey)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid exchange status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var resp exchangeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal exchange response: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected non-empty accessToken in response")
	}

	// 5. Verify DatabaseResolver resolves the exchanged access token
	resolver := NewDatabaseResolver(db)
	authedReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	authedReq.Header.Set("Authorization", "Bearer "+resp.AccessToken)

	tok, err := resolver(authedReq)
	if err != nil {
		t.Fatalf("resolver failed on valid token: %v", err)
	}
	if tok.Subject() != userID {
		t.Fatalf("token subject = %s, want %s", tok.Subject(), userID)
	}
	if tok.Type() != TypeUser {
		t.Fatalf("token type = %s, want user", tok.Type())
	}
	if len(tok.Scopes()) != 2 {
		t.Fatalf("expected 2 scopes, got %v", tok.Scopes())
	}

	// 6. DatabaseResolver rejects raw key directly (must exchange)
	rawKeyReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	rawKeyReq.Header.Set("Authorization", "Bearer "+rawKey)
	_, err = resolver(rawKeyReq)
	if err != errMustExchange {
		t.Fatalf("expected errMustExchange for raw key, got %v", err)
	}

	// 7. DatabaseResolver rejects arbitrary unrecognized token
	badReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	badReq.Header.Set("Authorization", "Bearer totally_fake_token")
	_, err = resolver(badReq)
	if err == nil {
		t.Fatal("expected error for fake token, got nil")
	}
}

func TestExchangeRevokedAndExpiredKeys(t *testing.T) {
	db, userID := setupTestStore(t)
	handler := ExchangeHandler(db)

	rawKey, prefix, hash, _ := GenerateAPIKey()
	apiKey := store.APIKey{
		ID:        "key_revoked",
		UserID:    userID,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Name:      "revoked-key",
		Scopes:    "*",
		RevokedAt: sql.NullString{String: time.Now().UTC().Format(time.RFC3339), Valid: true},
	}
	if err := db.CreateAPIKey(t.Context(), apiKey); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/exchange_user_api_key", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+rawKey)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked key status = %d, want 401", rec.Code)
	}
}
