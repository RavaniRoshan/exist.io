package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"exist.io/internal/config"
)

func TestPullRequestsLifecycle(t *testing.T) {
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

	// 1. Create a repository
	createRepoBody, _ := json.Marshal(map[string]any{
		"owner":          "acme",
		"name":           "rocket",
		"default_branch": "main",
		"visibility":     "public",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewReader(createRepoBody))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/repos code = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 2. Commit initial README.md to main branch
	commitMainBody, _ := json.Marshal(map[string]any{
		"branch":  "main",
		"path":    "README.md",
		"content": "# Acme Rocket\nInitial main commit.\n",
		"message": "docs: initial readme",
		"author":  "Operator",
		"email":   "operator@acme.com",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/rocket/files", bytes.NewReader(commitMainBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/repos/acme/rocket/files code = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 3. List branches
	req = httptest.NewRequest(http.MethodGet, "/api/v1/repos/acme/rocket/branches", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/repos/acme/rocket/branches code = %d", rec.Code)
	}
	var branchesResp struct {
		Branches []struct {
			Name      string `json:"name"`
			IsDefault bool   `json:"is_default"`
		} `json:"branches"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &branchesResp)
	if len(branchesResp.Branches) != 1 || branchesResp.Branches[0].Name != "main" {
		t.Fatalf("expected 1 main branch, got %+v", branchesResp.Branches)
	}

	// 4. Create a new branch 'feature/engine'
	createBranchBody, _ := json.Marshal(map[string]any{
		"name": "feature/engine",
		"from": "main",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/rocket/branches", bytes.NewReader(createBranchBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/repos/acme/rocket/branches code = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 5. Commit file on 'feature/engine'
	commitFeatBody, _ := json.Marshal(map[string]any{
		"branch":  "feature/engine",
		"path":    "engine.go",
		"content": "package rocket\n\nfunc Thrust() int { return 1000 }\n",
		"message": "feat: add rocket thrust engine",
		"author":  "Engineer",
		"email":   "eng@acme.com",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/rocket/files", bytes.NewReader(commitFeatBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST file on feature/engine code = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 6. Create Pull Request
	createPRBody, _ := json.Marshal(map[string]any{
		"title": "feat: implement rocket propulsion thrust",
		"head":  "feature/engine",
		"base":  "main",
		"body":  "This PR implements 1000kN rocket thrust engine.",
		"draft": false,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/rocket/pulls", bytes.NewReader(createPRBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/repos/acme/rocket/pulls code = %d, body = %s", rec.Code, rec.Body.String())
	}
	var prResp struct {
		OK   bool `json:"ok"`
		Pull struct {
			ID     string `json:"id"`
			Number int    `json:"number"`
			State  string `json:"state"`
		} `json:"pull"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &prResp)
	if !prResp.OK || prResp.Pull.Number != 1 || prResp.Pull.State != "open" {
		t.Fatalf("unexpected PR response: %+v", prResp)
	}

	// 7. List Pull Requests and verify diff hunks
	req = httptest.NewRequest(http.MethodGet, "/api/v1/repos/acme/rocket/pulls", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/repos/acme/rocket/pulls code = %d", rec.Code)
	}
	var listPRsResp struct {
		Pulls []struct {
			Number    int    `json:"number"`
			State     string `json:"state"`
			Title     string `json:"title"`
			DiffHunks []struct {
				NewPath string `json:"newPath"`
			} `json:"diffHunks"`
		} `json:"pulls"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &listPRsResp)
	if len(listPRsResp.Pulls) != 1 {
		t.Fatalf("expected 1 pull request, got %d", len(listPRsResp.Pulls))
	}
	if len(listPRsResp.Pulls[0].DiffHunks) != 1 || listPRsResp.Pulls[0].DiffHunks[0].NewPath != "engine.go" {
		t.Fatalf("expected diffHunk for engine.go, got %+v", listPRsResp.Pulls[0].DiffHunks)
	}

	// 8. Add a comment on the PR
	commentBody, _ := json.Marshal(map[string]any{
		"body": "Looks great, ready for deployment!",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/rocket/pulls/1/comments", bytes.NewReader(commentBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST comment code = %d, body = %s", rec.Code, rec.Body.String())
	}

	// 9. Run automated Agent Review
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/rocket/pulls/1/reviews/agent", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST agent review code = %d, body = %s", rec.Code, rec.Body.String())
	}
	var agentReviewResp struct {
		Verdict    string `json:"verdict"`
		Conclusion string `json:"conclusion"`
		Score      int    `json:"score"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &agentReviewResp)
	if agentReviewResp.Verdict == "" {
		t.Fatalf("expected non-empty verdict from agent review: %+v", agentReviewResp)
	}

	// 10. Merge the Pull Request
	mergeBody, _ := json.Marshal(map[string]any{
		"mergeMethod": "squash",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/rocket/pulls/1/merge", bytes.NewReader(mergeBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST merge code = %d, body = %s", rec.Code, rec.Body.String())
	}
	var mergeResp struct {
		OK             bool   `json:"ok"`
		MergeCommitSHA string `json:"mergeCommitSha"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &mergeResp)
	if !mergeResp.OK || mergeResp.MergeCommitSHA == "" {
		t.Fatalf("unexpected merge response: %+v", mergeResp)
	}

	// 11. Verify PR is now marked 'merged'
	req = httptest.NewRequest(http.MethodGet, "/api/v1/repos/acme/rocket/pulls", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &listPRsResp)
	if listPRsResp.Pulls[0].State != "merged" {
		t.Fatalf("expected PR state to be 'merged', got %q", listPRsResp.Pulls[0].State)
	}
}
