package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"exist.io/internal/config"
)

func TestReleasesRulesetsAndWebhooksLifecycle(t *testing.T) {
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
	defer application.Close()

	handler := application.Handler()

	// 1. Create repository
	createRepoBody, _ := json.Marshal(map[string]any{
		"owner":          "acme",
		"name":           "satellite",
		"default_branch": "main",
		"visibility":     "public",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewReader(createRepoBody))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/repos code = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 2. Commit a file to main so HEAD exists
	commitBody, _ := json.Marshal(map[string]any{
		"branch":  "main",
		"path":    "main.go",
		"content": "package main\n\nfunc main() {}\n",
		"message": "feat: initial commit",
		"author":  "Operator",
		"email":   "operator@acme.com",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/satellite/files", bytes.NewReader(commitBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST files code = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 3. Test Releases API: GET empty releases
	req = httptest.NewRequest(http.MethodGet, "/api/v1/repos/acme/satellite/releases", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /releases code = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 4. Test Releases API: POST new release
	pubRelBody, _ := json.Marshal(map[string]any{
		"tag_name":   "v1.0.0",
		"name":       "v1.0.0 Satellite Launch",
		"body":       "Initial orbital telemetry deployment.",
		"draft":      false,
		"prerelease": false,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/satellite/releases", bytes.NewReader(pubRelBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /releases code = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 5. Verify release in list
	req = httptest.NewRequest(http.MethodGet, "/api/v1/repos/acme/satellite/releases", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	var relListResp struct {
		Releases []struct {
			TagName string `json:"tagName"`
			Name    string `json:"name"`
		} `json:"releases"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&relListResp)
	if len(relListResp.Releases) != 1 || relListResp.Releases[0].TagName != "v1.0.0" {
		t.Fatalf("expected 1 release v1.0.0, got: %+v", relListResp.Releases)
	}

	// 6. Test Rulesets API: POST create ruleset
	rulesetBody, _ := json.Marshal(map[string]any{
		"name":             "Mainline Gate",
		"enforcement":      "active",
		"kind":             "branch",
		"includedRefNames": []string{"refs/heads/main"},
		"rules": []map[string]any{
			{"type": "pull_request", "parameters": map[string]any{"required_approving_review_count": 1}},
		},
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/satellite/rulesets", bytes.NewReader(rulesetBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /rulesets code = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 7. Test Rulesets API: GET list rulesets
	req = httptest.NewRequest(http.MethodGet, "/api/v1/repos/acme/satellite/rulesets", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	var rsResp struct {
		Rulesets []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"rulesets"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&rsResp)
	if len(rsResp.Rulesets) != 1 || rsResp.Rulesets[0].Name != "Mainline Gate" {
		t.Fatalf("expected 1 ruleset 'Mainline Gate', got: %+v", rsResp.Rulesets)
	}

	// 8. Test Webhooks API: POST ping webhook
	req = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/pings", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /webhooks/pings code = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 9. Test Webhooks API: GET deliveries
	req = httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/deliveries", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /webhooks/deliveries code = %d, body = %s", rec.Code, rec.Body.String())
	}
	var whResp struct {
		Deliveries []struct {
			ID    string `json:"id"`
			Event string `json:"event"`
		} `json:"deliveries"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&whResp)
	if len(whResp.Deliveries) < 1 {
		t.Fatalf("expected at least 1 webhook delivery, got %d", len(whResp.Deliveries))
	}
}
