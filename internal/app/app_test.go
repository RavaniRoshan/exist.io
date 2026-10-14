package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"exist.io/internal/config"
)

func TestAppLifecycle(t *testing.T) {
	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:8080",
		DataDir:    dataDir,
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
	}

	application, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() {
		if err := application.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	// 1. Health check endpoint works
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "ok\n" {
		t.Fatalf("GET /healthz body = %q, want %q", rec.Body.String(), "ok\n")
	}

	// 2. Origin route is mounted (GET /v1/origin/namespaces returns 401 unauthenticated when no token provided)
	reqRepo := httptest.NewRequest(http.MethodGet, "/v1/origin/namespaces", nil)
	recRepo := httptest.NewRecorder()
	application.Handler().ServeHTTP(recRepo, reqRepo)

	if recRepo.Code != http.StatusUnauthorized {
		t.Fatalf("GET /v1/origin/namespaces without auth = %d, want 401", recRepo.Code)
	}
}

func TestAppOpenFailure(t *testing.T) {
	f := t.TempDir() + "/blocked"
	if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		DataDir: f + "/cannot-mkdir",
	}

	_, err := Open(cfg)
	if err == nil {
		t.Fatal("expected Open to fail for blocked directory path, but got nil")
	}
}
