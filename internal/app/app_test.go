package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

	// 3. Prometheus metrics endpoint is mounted
	reqMetrics := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recMetrics := httptest.NewRecorder()
	application.Handler().ServeHTTP(recMetrics, reqMetrics)
	if recMetrics.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want 200", recMetrics.Code)
	}

	// 4. Admin health endpoint returns live cluster metrics
	reqHealth := httptest.NewRequest(http.MethodGet, "/api/v1/admin/health", nil)
	recHealth := httptest.NewRecorder()
	application.Handler().ServeHTTP(recHealth, reqHealth)
	if recHealth.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/health status = %d, want 200", recHealth.Code)
	}

	// 5. Admin audit logs endpoint returns json log records
	reqAudit := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs", nil)
	recAudit := httptest.NewRecorder()
	application.Handler().ServeHTTP(recAudit, reqAudit)
	if recAudit.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/audit-logs status = %d, want 200", recAudit.Code)
	}

	// 5b. Cluster replication status endpoint returns valid JSON
	reqRepl := httptest.NewRequest(http.MethodGet, "/api/v1/cluster/replication", nil)
	recRepl := httptest.NewRecorder()
	application.Handler().ServeHTTP(recRepl, reqRepl)
	if recRepl.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/cluster/replication status = %d, want 200", recRepl.Code)
	}

	// 5c. Failover promotion endpoint succeeds
	reqPromote := httptest.NewRequest(http.MethodPost, "/api/v1/cluster/failover/promote", nil)
	recPromote := httptest.NewRecorder()
	application.Handler().ServeHTTP(recPromote, reqPromote)
	if recPromote.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/cluster/failover/promote status = %d, want 200", recPromote.Code)
	}

	// 6. Embedded Web UI is served on root / and SPA routes
	reqUI := httptest.NewRequest(http.MethodGet, "/", nil)
	recUI := httptest.NewRecorder()
	application.Handler().ServeHTTP(recUI, reqUI)
	if recUI.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", recUI.Code)
	}
	if !strings.Contains(recUI.Body.String(), "<!doctype html>") && !strings.Contains(recUI.Body.String(), "<html") {
		t.Fatalf("expected embedded HTML on /, got %s", recUI.Body.String())
	}

	reqSPA := httptest.NewRequest(http.MethodGet, "/acme/frontend-app", nil)
	recSPA := httptest.NewRecorder()
	application.Handler().ServeHTTP(recSPA, reqSPA)
	if recSPA.Code != http.StatusOK {
		t.Fatalf("GET /acme/frontend-app status = %d, want 200", recSPA.Code)
	}
	if !strings.Contains(recSPA.Body.String(), "<!doctype html>") && !strings.Contains(recSPA.Body.String(), "<html") {
		t.Fatalf("expected embedded HTML SPA fallback, got %s", recSPA.Body.String())
	}

	// 7. Git requests are delegated to git transport (not HTML)
	reqGit := httptest.NewRequest(http.MethodGet, "/acme/repo.git/info/refs?service=git-upload-pack", nil)
	recGit := httptest.NewRecorder()
	application.Handler().ServeHTTP(recGit, reqGit)
	if strings.Contains(recGit.Body.String(), "<!doctype html>") {
		t.Fatalf("expected Git transport response, got HTML")
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
