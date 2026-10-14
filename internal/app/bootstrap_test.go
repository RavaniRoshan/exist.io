package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"exist.io/internal/config"
)

func TestBootstrapAndEndToEndExchange(t *testing.T) {
	ctx := context.Background()
	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:8080",
		DataDir:    dataDir,
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
	}

	// 1. Initial bootstrap succeeds
	result, err := Bootstrap(ctx, cfg, "admin@exist.io", "admin-org")
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	if result.User.Email != "admin@exist.io" {
		t.Fatalf("user email = %s, want admin@exist.io", result.User.Email)
	}
	if result.Namespace.Slug != "admin-org" {
		t.Fatalf("namespace slug = %s, want admin-org", result.Namespace.Slug)
	}
	if !strings.HasPrefix(result.RawKey, "key_") {
		t.Fatalf("rawKey = %s, want key_ prefix", result.RawKey)
	}

	// 2. Second bootstrap on the same data dir fails
	_, err = Bootstrap(ctx, cfg, "second@exist.io", "second-org")
	if err == nil {
		t.Fatal("expected second bootstrap to fail, but got nil")
	}

	// 3. Start App over the bootstrapped data dir
	application, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer application.Close()

	// A. Raw key cannot be used directly as bearer token on protected routes
	unexchangedReq := httptest.NewRequest(http.MethodGet, "/v1/origin/namespaces", nil)
	unexchangedReq.Header.Set("Authorization", "Bearer "+result.RawKey)
	unexchangedRec := httptest.NewRecorder()
	application.Handler().ServeHTTP(unexchangedRec, unexchangedReq)
	if unexchangedRec.Code != http.StatusUnauthorized {
		t.Fatalf("raw key without exchange status = %d, want 401", unexchangedRec.Code)
	}

	// B. Exchange raw key for access token
	exchangeReq := httptest.NewRequest(http.MethodPost, "/auth/exchange_user_api_key", strings.NewReader("{}"))
	exchangeReq.Header.Set("Authorization", "Bearer "+result.RawKey)
	exchangeRec := httptest.NewRecorder()
	application.Handler().ServeHTTP(exchangeRec, exchangeReq)

	if exchangeRec.Code != http.StatusOK {
		t.Fatalf("exchange status = %d, want 200; body=%s", exchangeRec.Code, exchangeRec.Body.String())
	}

	var exchangeResp struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.Unmarshal(exchangeRec.Body.Bytes(), &exchangeResp); err != nil {
		t.Fatalf("unmarshal exchange response: %v", err)
	}
	if exchangeResp.AccessToken == "" {
		t.Fatal("expected non-empty accessToken")
	}

	// C. Call /v1/origin/namespaces with exchanged access token -> succeeds!
	authedReq := httptest.NewRequest(http.MethodGet, "/v1/origin/namespaces", nil)
	authedReq.Header.Set("Authorization", "Bearer "+exchangeResp.AccessToken)
	authedRec := httptest.NewRecorder()
	application.Handler().ServeHTTP(authedRec, authedReq)

	if authedRec.Code != http.StatusOK {
		t.Fatalf("authenticated GET /v1/origin/namespaces status = %d, want 200; body=%s",
			authedRec.Code, authedRec.Body.String())
	}

	var nsResp struct {
		Namespaces []struct {
			Slug string `json:"slug"`
		} `json:"namespaces"`
	}
	if err := json.Unmarshal(authedRec.Body.Bytes(), &nsResp); err != nil {
		t.Fatalf("unmarshal namespaces: %v", err)
	}
	if len(nsResp.Namespaces) != 1 || nsResp.Namespaces[0].Slug != "admin-org" {
		t.Fatalf("unexpected namespaces items: %+v", nsResp.Namespaces)
	}

	// D. Arbitrary fake token fails
	fakeReq := httptest.NewRequest(http.MethodGet, "/v1/origin/namespaces", nil)
	fakeReq.Header.Set("Authorization", "Bearer arbitrary_fake_token")
	fakeRec := httptest.NewRecorder()
	application.Handler().ServeHTTP(fakeRec, fakeReq)

	if fakeRec.Code != http.StatusUnauthorized {
		t.Fatalf("fake token status = %d, want 401", fakeRec.Code)
	}
}
