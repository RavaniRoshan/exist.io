package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"exist.io/internal/app"
	"exist.io/internal/config"
	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func TestCommentsReviewsAndAutoMergeAPI(t *testing.T) {
	cfg := config.Config{
		DataDir: ":memory:",
	}
	application, err := app.Open(cfg)
	if err != nil {
		t.Fatalf("Open app: %v", err)
	}
	defer application.Close()

	ctx := context.Background()

	// 1. Setup namespace, repo, branches, and PR
	ns := store.Namespace{
		ID:   idgen.NewNS(),
		Slug: "acme",
		Name: "Acme Corp",
	}
	if err := application.Store.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("create namespace: %v", err)
	}

	repo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "widget",
		FullName:      "acme/widget",
		DefaultBranch: "main",
	}
	if err := application.Store.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("create repo: %v", err)
	}

	if _, err := application.Git.Init("acme", "widget", "main"); err != nil {
		t.Fatalf("git init: %v", err)
	}

	sig := object.Signature{Name: "Tester", Email: "test@exist.io", When: time.Now()}
	baseHash, err := application.Git.WriteFiles("acme", "widget", "main", []gitstore.FileChange{
		{Path: "main.go", Content: "package main\n\nfunc main() {\n\tprintln(\"hello world\")\n}\n"},
	}, sig, "initial commit")
	if err != nil {
		t.Fatalf("write main: %v", err)
	}

	headHash, err := application.Git.WriteFiles("acme", "widget", "feature", []gitstore.FileChange{
		{Path: "main.go", Content: "package main\n\nfunc main() {\n\tprintln(\"hello exist.io\")\n}\n"},
	}, sig, "feature commit")
	if err != nil {
		t.Fatalf("write feature: %v", err)
	}

	pr := &store.PullRequest{
		ID:        idgen.NewPR(),
		RepoID:    repo.ID,
		Number:    1,
		Title:     "Update greeting",
		Body:      "Uses exist.io",
		State:     "open",
		HeadRef:   "feature",
		HeadSHA:   headHash.String(),
		BaseRef:   "main",
		BaseSHA:   baseHash.String(),
		Mergeable: "mergeable",
	}
	if err := application.Store.CreatePullRequest(ctx, pr); err != nil {
		t.Fatalf("create PR: %v", err)
	}

	handler := application.Handler()

	// 2. Test Inline Comment Creation
	commentReqBody := map[string]any{
		"body":       "```suggestion\n\tprintln(\"hello universe\")\n```",
		"path":       "main.go",
		"line":       4,
		"side":       "RIGHT",
		"start_line": 4,
		"end_line":   4,
	}
	bodyBytes, _ := json.Marshal(commentReqBody)
	req := httptest.NewRequest("POST", "/api/v1/repos/acme/widget/pulls/1/comments", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create inline comment expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var commentResp struct {
		Ok      bool `json:"ok"`
		Comment struct {
			ID       string `json:"id"`
			Path     string `json:"path"`
			Line     int    `json:"line"`
			ThreadID string `json:"threadId"`
		} `json:"comment"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&commentResp)
	if !commentResp.Ok || commentResp.Comment.ID == "" || commentResp.Comment.ThreadID == "" {
		t.Fatalf("unexpected comment response: %+v", commentResp)
	}
	commentID := commentResp.Comment.ID
	threadID := commentResp.Comment.ThreadID

	// 3. Test List Comments
	req = httptest.NewRequest("GET", "/api/v1/repos/acme/widget/pulls/1/comments", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list comments expected 200, got %d", rec.Code)
	}
	var listResp struct {
		Ok       bool             `json:"ok"`
		Comments []map[string]any `json:"comments"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&listResp)
	if !listResp.Ok || len(listResp.Comments) != 1 {
		t.Fatalf("expected 1 comment in list, got %+v", listResp)
	}
	if listResp.Comments[0]["path"] != "main.go" || listResp.Comments[0]["resolved"] != false {
		t.Fatalf("unexpected comment list item: %+v", listResp.Comments[0])
	}

	// 4. Test Resolve Thread
	patchBody, _ := json.Marshal(map[string]any{"resolved": true})
	req = httptest.NewRequest("PATCH", "/api/v1/repos/acme/widget/pulls/1/threads/"+threadID, bytes.NewReader(patchBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("resolve thread expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 5. Test Code Suggestion Apply
	req = httptest.NewRequest("POST", "/api/v1/repos/acme/widget/pulls/1/comments/"+commentID+"/apply", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("apply suggestion expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var applyResp struct {
		Ok        bool   `json:"ok"`
		CommitSHA string `json:"commitSha"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&applyResp)
	if !applyResp.Ok || applyResp.CommitSHA == "" {
		t.Fatalf("unexpected apply response: %+v", applyResp)
	}

	// 6. Test Review Batching
	reviewBody, _ := json.Marshal(map[string]any{
		"verdict": "approve",
		"body":    "LGTM! Clean change.",
		"comments": []map[string]any{
			{
				"path": "main.go",
				"line": 4,
				"side": "RIGHT",
				"body": "Nice improvement!",
			},
		},
	})
	req = httptest.NewRequest("POST", "/api/v1/repos/acme/widget/pulls/1/reviews", bytes.NewReader(reviewBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create review expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// 7. Test Auto-Merge Configuration (PUT, GET, DELETE)
	amPutBody, _ := json.Marshal(map[string]any{
		"merge_method":   "squash",
		"commit_title":   "feat: greeting update (#1)",
		"commit_message": "Automated merge by exist.io",
	})
	req = httptest.NewRequest("PUT", "/api/v1/repos/acme/widget/pulls/1/auto-merge", bytes.NewReader(amPutBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("enable auto-merge expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest("GET", "/api/v1/repos/acme/widget/pulls/1/auto-merge", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get auto-merge expected 200, got %d", rec.Code)
	}
	var amGetResp struct {
		Ok      bool `json:"ok"`
		Enabled bool `json:"enabled"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&amGetResp)
	if !amGetResp.Ok || !amGetResp.Enabled {
		t.Fatalf("expected auto-merge enabled, got %+v", amGetResp)
	}

	req = httptest.NewRequest("DELETE", "/api/v1/repos/acme/widget/pulls/1/auto-merge", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete auto-merge expected 200, got %d", rec.Code)
	}
}
