package origin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

func TestChecks_Handlers(t *testing.T) {
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

	headSHA := "9a41f0c3d2b8e7f6a5c4d3e2f1b0a9c8d7e6f5a4"

	// 1. PostCheckRun
	now := time.Now().UTC()
	extUpdate := now.Format(time.RFC3339)
	postBody := map[string]any{
		"identifier": map[string]string{
			"ownerSlug": "acme",
			"name":      "rocket",
		},
		"headSha": headSHA,
		"checkSuite": map[string]any{
			"key":        "ci-8842",
			"name":       "CI",
			"detailsUrl": "https://ci.acme.dev/runs/8842",
			"externalId": "build-8842",
		},
		"checkRun": map[string]any{
			"key":               "ci-8842-unit",
			"name":              "unit-tests",
			"status":            "completed",
			"conclusion":        "success",
			"detailsUrl":        "https://ci.acme.dev/runs/8842",
			"externalId":        "run-8842-1",
			"externalUpdatedAt": extUpdate,
			"isRerequestable":   true,
			"output": map[string]string{
				"title":   "Unit Tests",
				"summary": "128 tests passed.",
				"text":    "All suites green.",
			},
		},
	}
	raw, _ := json.Marshal(postBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/origin/repos/acme/rocket/check-runs", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("postCheckRun expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var postResp struct {
		CheckSuite map[string]any `json:"checkSuite"`
		CheckRun   map[string]any `json:"checkRun"`
		Outcome    string         `json:"outcome"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &postResp); err != nil {
		t.Fatalf("unmarshal postCheckRun response: %v", err)
	}
	if postResp.Outcome != "created" {
		t.Fatalf("expected outcome created, got %s", postResp.Outcome)
	}
	runID, _ := postResp.CheckRun["id"].(string)
	suiteID, _ := postResp.CheckSuite["id"].(string)
	if runID == "" || suiteID == "" {
		t.Fatalf("expected populated IDs: run=%s, suite=%s", runID, suiteID)
	}

	// 2. PostCheckRun with newer update -> updated
	newerUpdate := now.Add(2 * time.Minute).Format(time.RFC3339)
	postBody["checkRun"].(map[string]any)["externalUpdatedAt"] = newerUpdate
	postBody["checkRun"].(map[string]any)["output"].(map[string]string)["summary"] = "130 tests passed."
	raw, _ = json.Marshal(postBody)
	req = httptest.NewRequest(http.MethodPost, "/v1/origin/repos/acme/rocket/check-runs", bytes.NewReader(raw))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("postCheckRun update expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &postResp)
	if postResp.Outcome != "updated" {
		t.Fatalf("expected outcome updated, got %s", postResp.Outcome)
	}

	// 3. PostCheckRun with older update -> ignored_stale
	olderUpdate := now.Add(-10 * time.Minute).Format(time.RFC3339)
	postBody["checkRun"].(map[string]any)["externalUpdatedAt"] = olderUpdate
	raw, _ = json.Marshal(postBody)
	req = httptest.NewRequest(http.MethodPost, "/v1/origin/repos/acme/rocket/check-runs", bytes.NewReader(raw))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("postCheckRun stale expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &postResp)
	if postResp.Outcome != "ignored_stale" {
		t.Fatalf("expected outcome ignored_stale, got %s", postResp.Outcome)
	}

	// 4. GetCheckRun
	req = httptest.NewRequest(http.MethodGet, "/v1/origin/repos/acme/rocket/check-runs/"+runID, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("getCheckRun expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var getRun map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &getRun); err != nil {
		t.Fatalf("unmarshal getCheckRun: %v", err)
	}
	if getRun["id"] != runID || getRun["status"] != "completed" {
		t.Fatalf("unexpected getCheckRun response: %+v", getRun)
	}

	// 5. GetCheckSuite
	req = httptest.NewRequest(http.MethodGet, "/v1/origin/repos/acme/rocket/check-suites/"+suiteID, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("getCheckSuite expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var getSuite map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &getSuite); err != nil {
		t.Fatalf("unmarshal getCheckSuite: %v", err)
	}
	if getSuite["id"] != suiteID || getSuite["name"] != "CI" {
		t.Fatalf("unexpected getCheckSuite response: %+v", getSuite)
	}

	// 6. BatchUpsertCheckRuns
	batchBody := map[string]any{
		"headSha": headSHA,
		"checkSuite": map[string]any{
			"key":        "ci-8842",
			"name":       "CI",
			"externalId": "build-8842",
		},
		"checkRuns": []map[string]any{
			{
				"key":               "ci-lint",
				"name":              "linter",
				"status":            "completed",
				"conclusion":        "success",
				"externalId":        "run-8842-lint",
				"externalUpdatedAt": extUpdate,
			},
			{
				"key":               "ci-e2e",
				"name":              "e2e-tests",
				"status":            "in_progress",
				"externalId":        "run-8842-e2e",
				"externalUpdatedAt": extUpdate,
			},
		},
	}
	raw, _ = json.Marshal(batchBody)
	req = httptest.NewRequest(http.MethodPost, "/v1/origin/repos/acme/rocket/check-runs:batchUpsert", bytes.NewReader(raw))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("batchUpsertCheckRuns expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var batchResp struct {
		CheckSuite map[string]any   `json:"checkSuite"`
		CheckRuns  []map[string]any `json:"checkRuns"`
		Results    []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &batchResp); err != nil {
		t.Fatalf("unmarshal batchUpsertCheckRuns: %v", err)
	}
	if len(batchResp.CheckRuns) != 2 || len(batchResp.Results) != 2 {
		t.Fatalf("expected 2 check runs, got %d", len(batchResp.CheckRuns))
	}

	// 7. ListCheckRunsForSuite
	req = httptest.NewRequest(http.MethodGet, "/v1/origin/repos/acme/rocket/check-suites/"+suiteID+"/check-runs", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("listCheckRunsForSuite expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var suiteRunsResp struct {
		CheckRuns []map[string]any `json:"checkRuns"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &suiteRunsResp)
	if len(suiteRunsResp.CheckRuns) != 3 {
		t.Fatalf("expected 3 check runs in suite, got %d", len(suiteRunsResp.CheckRuns))
	}

	// 8. ListCheckRunsForCommit & ListCheckSuitesForCommit
	req = httptest.NewRequest(http.MethodGet, "/v1/origin/repos/acme/rocket/commits/"+headSHA+"/check-runs?status=completed", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("listCheckRunsForCommit expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var commitRunsResp struct {
		CheckRuns []map[string]any `json:"checkRuns"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &commitRunsResp)
	if len(commitRunsResp.CheckRuns) != 2 {
		t.Fatalf("expected 2 completed runs, got %d", len(commitRunsResp.CheckRuns))
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/origin/repos/acme/rocket/commits/"+headSHA+"/check-suites", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("listCheckSuitesForCommit expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var commitSuitesResp struct {
		CheckSuites []map[string]any `json:"checkSuites"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &commitSuitesResp)
	if len(commitSuitesResp.CheckSuites) != 1 {
		t.Fatalf("expected 1 suite for commit, got %d", len(commitSuitesResp.CheckSuites))
	}

	// 9. Annotations
	annBody := map[string]any{
		"annotations": []map[string]any{
			{
				"annotationLevel": "warning",
				"message":         "Deprecated method call",
				"title":           "Deprecation",
				"location": map[string]any{
					"path":      "main.go",
					"startLine": 42,
					"endLine":   42,
					"columns": map[string]any{
						"startColumn": 5,
						"endColumn":   20,
					},
				},
			},
		},
	}
	raw, _ = json.Marshal(annBody)
	req = httptest.NewRequest(http.MethodPost, "/v1/origin/repos/acme/rocket/check-runs/"+runID+"/annotations", bytes.NewReader(raw))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("createCheckRunAnnotations expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var createAnnResp struct {
		Annotations []map[string]any `json:"annotations"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &createAnnResp)
	if len(createAnnResp.Annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(createAnnResp.Annotations))
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/origin/repos/acme/rocket/check-runs/"+runID+"/annotations", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("listCheckRunAnnotations expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var listAnnResp struct {
		Annotations []map[string]any `json:"annotations"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &listAnnResp)
	if len(listAnnResp.Annotations) != 1 || listAnnResp.Annotations[0]["message"] != "Deprecated method call" {
		t.Fatalf("unexpected annotations: %+v", listAnnResp.Annotations)
	}

	// 10. Rerequest
	req = httptest.NewRequest(http.MethodPost, "/v1/origin/repos/acme/rocket/check-runs/"+runID+"/rerequest", bytes.NewReader([]byte("{}")))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("rerequestCheckRun expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var rerequestResp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &rerequestResp)
	if rerequestResp["status"] != "rerequested" || rerequestResp["rerequestedAt"] == nil {
		t.Fatalf("expected rerequested status, got %+v", rerequestResp)
	}

	// Repeat rerequest fails with 409 Conflict
	req = httptest.NewRequest(http.MethodPost, "/v1/origin/repos/acme/rocket/check-runs/"+runID+"/rerequest", bytes.NewReader([]byte("{}")))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("repeat rerequestCheckRun expected 409, got %d", rec.Code)
	}
}
