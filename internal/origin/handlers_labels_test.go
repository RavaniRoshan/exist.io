package origin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

func TestLabels_Handlers(t *testing.T) {
	srv, db, handler := testServer(t)
	_ = srv

	ctx := t.Context()
	ns := seedNS(t, db, "acme")
	repo := store.Repo{
		ID:          idgen.NewRepo(),
		NamespaceID: ns.ID,
		Name:        "rocket",
		FullName:    "acme/rocket",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	// 1. Create label
	createBody := map[string]any{
		"name":        "bug",
		"color":       "d73a4a",
		"description": "Something isn't working",
	}
	raw, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/origin/repos/acme/rocket/labels", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("create label expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created: %v", err)
	}
	if created["name"] != "bug" || created["color"] != "d73a4a" {
		t.Fatalf("unexpected created label: %+v", created)
	}

	// 2. List labels
	req = httptest.NewRequest(http.MethodGet, "/v1/origin/repos/acme/rocket/labels", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("list labels expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var listResp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	labels, ok := listResp["labels"].([]any)
	if !ok || len(labels) != 1 {
		t.Fatalf("expected 1 label, got: %+v", listResp)
	}

	// 3. Get label
	req = httptest.NewRequest(http.MethodGet, "/v1/origin/repos/acme/rocket/labels/bug", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("get label expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 4. Update label
	patchBody := map[string]any{
		"name":        "critical-bug",
		"color":       "b60205",
		"description": "Urgent fix required",
	}
	raw, _ = json.Marshal(patchBody)
	req = httptest.NewRequest(http.MethodPatch, "/v1/origin/repos/acme/rocket/labels/bug", bytes.NewReader(raw))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("patch label expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var patched map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &patched); err != nil {
		t.Fatalf("unmarshal patched: %v", err)
	}
	if patched["name"] != "critical-bug" || patched["color"] != "b60205" {
		t.Fatalf("unexpected patched label: %+v", patched)
	}

	// 5. Delete label
	req = httptest.NewRequest(http.MethodDelete, "/v1/origin/repos/acme/rocket/labels/critical-bug", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete label expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}
