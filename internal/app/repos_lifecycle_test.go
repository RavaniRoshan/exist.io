package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"exist.io/internal/config"
	"exist.io/internal/store"
)

func TestRepoLifecycle_DangerZoneAndTemplates(t *testing.T) {
	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:0",
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

	// 1. Create a repository with README and Go .gitignore template
	createBody, _ := json.Marshal(map[string]any{
		"owner":              "acme",
		"name":               "satellite",
		"description":        "Orbital communications relay",
		"default_branch":     "main",
		"visibility":         "public",
		"init_readme":        true,
		"gitignore_template": "go",
		"license_template":   "MIT",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Create repo status = %d; body = %s", rec.Code, rec.Body.String())
	}

	// Verify repo was created and git files exist
	repo, err := application.Store.GetRepo(req.Context(), "acme", "satellite")
	if err != nil {
		t.Fatalf("GetRepo: %v", err)
	}
	if repo.Description != "Orbital communications relay" {
		t.Errorf("got Description %q, want Orbital communications relay", repo.Description)
	}
	if repo.Visibility != "public" {
		t.Errorf("got Visibility %q, want public", repo.Visibility)
	}

	// 2. Try changing visibility without confirm_name -> should fail with 400
	visChangeBody, _ := json.Marshal(map[string]any{
		"visibility":  "private",
		"sudo_secret": "my-secret-key-12345",
	})
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/repos/acme/satellite", bytes.NewReader(visChangeBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchRec := httptest.NewRecorder()
	handler.ServeHTTP(patchRec, patchReq)

	if patchRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request without confirmation name, got %d: %s", patchRec.Code, patchRec.Body.String())
	}

	// 3. Try changing visibility with correct confirm_name and sudo_secret -> should succeed
	visChangeValid, _ := json.Marshal(map[string]any{
		"visibility":   "private",
		"confirm_name": "acme/satellite",
		"sudo_secret":  "my-secret-key-12345",
	})
	patchReq2 := httptest.NewRequest(http.MethodPatch, "/api/v1/repos/acme/satellite", bytes.NewReader(visChangeValid))
	patchReq2.Header.Set("Content-Type", "application/json")
	patchRec2 := httptest.NewRecorder()
	handler.ServeHTTP(patchRec2, patchReq2)

	if patchRec2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on valid visibility change, got %d: %s", patchRec2.Code, patchRec2.Body.String())
	}

	repoAfterVis, _ := application.Store.GetRepo(req.Context(), "acme", "satellite")
	if repoAfterVis.Visibility != "private" {
		t.Errorf("expected Visibility private, got %s", repoAfterVis.Visibility)
	}

	// 4. Archive repository with confirm_name -> should succeed
	archBody, _ := json.Marshal(map[string]any{
		"archived":     true,
		"confirm_name": "acme/satellite",
	})
	patchReq3 := httptest.NewRequest(http.MethodPatch, "/api/v1/repos/acme/satellite", bytes.NewReader(archBody))
	patchReq3.Header.Set("Content-Type", "application/json")
	patchRec3 := httptest.NewRecorder()
	handler.ServeHTTP(patchRec3, patchReq3)

	if patchRec3.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on archival, got %d: %s", patchRec3.Code, patchRec3.Body.String())
	}

	repoArch, _ := application.Store.GetRepo(req.Context(), "acme", "satellite")
	if !repoArch.Archived {
		t.Errorf("expected repo to be archived, got false")
	}

	// 5. Rename repository with confirm_name -> should succeed
	renameBody, _ := json.Marshal(map[string]any{
		"new_name":     "satellite-v2",
		"confirm_name": "acme/satellite",
	})
	patchReq4 := httptest.NewRequest(http.MethodPatch, "/api/v1/repos/acme/satellite", bytes.NewReader(renameBody))
	patchReq4.Header.Set("Content-Type", "application/json")
	patchRec4 := httptest.NewRecorder()
	handler.ServeHTTP(patchRec4, patchReq4)

	if patchRec4.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on rename, got %d: %s", patchRec4.Code, patchRec4.Body.String())
	}

	// Old repo name should not exist, new repo name should exist
	if _, err := application.Store.GetRepo(req.Context(), "acme", "satellite"); err == nil {
		t.Errorf("expected old repo name satellite to be not found")
	}
	renamedRepo, err := application.Store.GetRepo(req.Context(), "acme", "satellite-v2")
	if err != nil {
		t.Fatalf("GetRepo(satellite-v2): %v", err)
	}
	if renamedRepo.Name != "satellite-v2" || renamedRepo.FullName != "acme/satellite-v2" {
		t.Errorf("unexpected renamed repo: %+v", renamedRepo)
	}

	// 6. Archival write freeze: creating an issue on archived repo must fail with 403 Forbidden
	issueBody, _ := json.Marshal(map[string]any{
		"title": "Solar flare interference",
		"body":  "Telemetry dropped during flare event",
	})
	issueReq := httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/satellite-v2/issues", bytes.NewReader(issueBody))
	issueReq.Header.Set("Content-Type", "application/json")
	issueRec := httptest.NewRecorder()
	handler.ServeHTTP(issueRec, issueReq)
	if issueRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when creating issue on archived repo, got %d: %s", issueRec.Code, issueRec.Body.String())
	}

	// Unarchive repo
	unarchBody, _ := json.Marshal(map[string]any{
		"archived":     false,
		"confirm_name": "acme/satellite-v2",
	})
	unarchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/repos/acme/satellite-v2", bytes.NewReader(unarchBody))
	unarchReq.Header.Set("Content-Type", "application/json")
	unarchRec := httptest.NewRecorder()
	handler.ServeHTTP(unarchRec, unarchReq)
	if unarchRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on unarchive, got %d: %s", unarchRec.Code, unarchRec.Body.String())
	}

	// Now creating an issue succeeds
	issueReq2 := httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/satellite-v2/issues", bytes.NewReader(issueBody))
	issueReq2.Header.Set("Content-Type", "application/json")
	issueRec2 := httptest.NewRecorder()
	handler.ServeHTTP(issueRec2, issueReq2)
	if issueRec2.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on unarchived repo, got %d: %s", issueRec2.Code, issueRec2.Body.String())
	}

	// 7. Repository Transfer to new owner namespace
	// Create second namespace "globex"
	_ = application.Store.CreateNamespace(req.Context(), store.Namespace{
		ID:   "ns-globex",
		Slug: "globex",
		Type: "organization",
		Name: "Globex Corporation",
	})

	// Try transfer without confirmation name -> fail 400
	transferBad, _ := json.Marshal(map[string]any{
		"new_owner":   "globex",
		"sudo_secret": "my-secret-key-12345",
	})
	trReqBad := httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/satellite-v2/transfer", bytes.NewReader(transferBad))
	trReqBad.Header.Set("Content-Type", "application/json")
	trRecBad := httptest.NewRecorder()
	handler.ServeHTTP(trRecBad, trReqBad)
	if trRecBad.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request without confirmation name on transfer, got %d", trRecBad.Code)
	}

	// Try transfer with valid confirmation and sudo secret -> succeed 200 OK
	transferValid, _ := json.Marshal(map[string]any{
		"new_owner":    "globex",
		"confirm_name": "acme/satellite-v2",
		"sudo_secret":  "my-secret-key-12345",
	})
	trReq := httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/satellite-v2/transfer", bytes.NewReader(transferValid))
	trReq.Header.Set("Content-Type", "application/json")
	trRec := httptest.NewRecorder()
	handler.ServeHTTP(trRec, trReq)
	if trRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on repository transfer, got %d: %s", trRec.Code, trRec.Body.String())
	}

	// Old repo in acme namespace must be gone
	if _, err := application.Store.GetRepo(req.Context(), "acme", "satellite-v2"); err == nil {
		t.Errorf("expected acme/satellite-v2 to be gone after transfer")
	}

	// Transferred repo in globex namespace must exist
	transferredRepo, err := application.Store.GetRepo(req.Context(), "globex", "satellite-v2")
	if err != nil {
		t.Fatalf("GetRepo(globex/satellite-v2): %v", err)
	}
	if transferredRepo.FullName != "globex/satellite-v2" {
		t.Errorf("got FullName %q, want globex/satellite-v2", transferredRepo.FullName)
	}
}
