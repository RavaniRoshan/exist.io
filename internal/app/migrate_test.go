package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"exist.io/internal/config"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// TestMigrateRequest_Validation verifies parameter validation for migration requests.
func TestMigrateRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     MigrateRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "empty clone url",
			req: MigrateRequest{
				CloneURL: "",
				Owner:    "acme",
				RepoName: "widgets",
			},
			wantErr: true,
			errMsg:  "clone_url is required",
		},
		{
			name: "invalid clone url scheme",
			req: MigrateRequest{
				CloneURL: "ftp://example.com/repo.git",
				Owner:    "acme",
				RepoName: "widgets",
			},
			wantErr: true,
			errMsg:  "clone_url must be a valid HTTP, HTTPS, Git, SSH, or file URL",
		},
		{
			name: "empty owner",
			req: MigrateRequest{
				CloneURL: "https://github.com/acme/widgets.git",
				Owner:    "",
				RepoName: "widgets",
			},
			wantErr: true,
			errMsg:  "owner is required",
		},
		{
			name: "invalid owner slug",
			req: MigrateRequest{
				CloneURL: "https://github.com/acme/widgets.git",
				Owner:    "invalid owner with spaces",
				RepoName: "widgets",
			},
			wantErr: true,
			errMsg:  "invalid owner",
		},
		{
			name: "empty repo name",
			req: MigrateRequest{
				CloneURL: "https://github.com/acme/widgets.git",
				Owner:    "acme",
				RepoName: "",
			},
			wantErr: true,
			errMsg:  "repo_name is required",
		},
		{
			name: "repo name ending with .git",
			req: MigrateRequest{
				CloneURL: "https://github.com/acme/widgets.git",
				Owner:    "acme",
				RepoName: "widgets.git",
			},
			wantErr: true,
			errMsg:  "invalid repo_name",
		},
		{
			name: "valid https request",
			req: MigrateRequest{
				CloneURL: "https://github.com/octocat/Hello-World.git",
				Owner:    "octocat",
				RepoName: "Hello-World",
			},
			wantErr: false,
		},
		{
			name: "valid git ssh request",
			req: MigrateRequest{
				CloneURL: "git@github.com:torvalds/linux.git",
				Owner:    "torvalds",
				RepoName: "linux",
			},
			wantErr: false,
		},
		{
			name: "valid file url",
			req: MigrateRequest{
				CloneURL: "file:///tmp/git/repo.git",
				Owner:    "test-org",
				RepoName: "local-repo",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Validate() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("Validate() error %q, want substring %q", err.Error(), tc.errMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("Validate() unexpected error: %v", err)
				}
			}
		})
	}

	// Test HTTP handler validation behaviors
	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:0",
		DataDir:    dataDir,
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
	}

	app, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer app.Close()

	// 1. Invalid JSON body -> 400
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repos/migrate", strings.NewReader("{invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rec.Code)
	}

	// 2. Non-existent owner -> 400
	body, _ := json.Marshal(MigrateRequest{
		CloneURL: "https://github.com/acme/project.git",
		Owner:    "nonexistent-owner",
		RepoName: "project",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/repos/migrate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-existent owner, got %d", rec.Code)
	}
}

// TestGitHubClient_PaginationParsing verifies Link header parsing and paginated fetching.
func TestGitHubClient_PaginationParsing(t *testing.T) {
	// 1. Direct unit tests for parseNextLink
	linkTests := []struct {
		name       string
		header     string
		wantTarget string
	}{
		{
			name:       "empty header",
			header:     "",
			wantTarget: "",
		},
		{
			name:       "single next link",
			header:     `<https://api.github.com/repos/octocat/Hello-World/issues?page=2>; rel="next"`,
			wantTarget: "https://api.github.com/repos/octocat/Hello-World/issues?page=2",
		},
		{
			name:       "next and last links",
			header:     `<https://api.github.com/repos/octocat/Hello-World/issues?page=2>; rel="next", <https://api.github.com/repos/octocat/Hello-World/issues?page=5>; rel="last"`,
			wantTarget: "https://api.github.com/repos/octocat/Hello-World/issues?page=2",
		},
		{
			name:       "prev and next links",
			header:     `<https://api.github.com/repos/octocat/Hello-World/issues?page=1>; rel="prev", <https://api.github.com/repos/octocat/Hello-World/issues?page=3>; rel="next"`,
			wantTarget: "https://api.github.com/repos/octocat/Hello-World/issues?page=3",
		},
		{
			name:       "last link only (terminal page)",
			header:     `<https://api.github.com/repos/octocat/Hello-World/issues?page=5>; rel="last"`,
			wantTarget: "",
		},
		{
			name:       "unquoted rel with spaces",
			header:     ` <https://api.github.com/issues?page=2> ; rel=next `,
			wantTarget: "https://api.github.com/issues?page=2",
		},
	}

	for _, tt := range linkTests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNextLink(tt.header)
			if got != tt.wantTarget {
				t.Errorf("parseNextLink(%q) = %q, want %q", tt.header, got, tt.wantTarget)
			}
		})
	}

	// 2. Integration test with mock HTTP server returning 3 paginated pages
	var requestsReceived []string
	var authHeaders []string
	var userAgentHeaders []string

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsReceived = append(requestsReceived, r.URL.RequestURI())
		authHeaders = append(authHeaders, r.Header.Get("Authorization"))
		userAgentHeaders = append(userAgentHeaders, r.Header.Get("User-Agent"))
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/repos/octocat/Hello-World/issues":
			page := r.URL.Query().Get("page")
			switch page {
			case "", "1":
				w.Header().Set("Link", fmt.Sprintf(`<%s/repos/octocat/Hello-World/issues?page=2>; rel="next", <%s/repos/octocat/Hello-World/issues?page=3>; rel="last"`, "http://"+r.Host, "http://"+r.Host))
				_ = json.NewEncoder(w).Encode([]GitHubIssue{
					{Number: 1, Title: "Issue 1", State: "open"},
					{Number: 2, Title: "Issue 2", State: "open"},
				})
			case "2":
				w.Header().Set("Link", fmt.Sprintf(`<%s/repos/octocat/Hello-World/issues?page=3>; rel="next", <%s/repos/octocat/Hello-World/issues?page=1>; rel="prev"`, "http://"+r.Host, "http://"+r.Host))
				_ = json.NewEncoder(w).Encode([]GitHubIssue{
					{Number: 3, Title: "Issue 3", State: "closed"},
					{Number: 4, Title: "Issue 4", State: "open"},
				})
			case "3":
				// Final page, no next link
				w.Header().Set("Link", fmt.Sprintf(`<%s/repos/octocat/Hello-World/issues?page=2>; rel="prev"`, "http://"+r.Host))
				_ = json.NewEncoder(w).Encode([]GitHubIssue{
					{Number: 5, Title: "Issue 5", State: "closed"},
				})
			default:
				http.NotFound(w, r)
			}
		case "/repos/octocat/Hello-World/labels":
			_ = json.NewEncoder(w).Encode([]GitHubLabel{
				{Name: "bug", Color: "ff0000"},
				{Name: "feature", Color: "00ff00"},
			})
		case "/repos/octocat/Hello-World/pulls":
			_ = json.NewEncoder(w).Encode([]GitHubPR{
				{Number: 10, Title: "PR 10", State: "open", Head: GitHubBranch{Ref: "feat", SHA: "abc"}, Base: GitHubBranch{Ref: "main", SHA: "123"}},
			})
		case "/repos/octocat/Hello-World/issues/1/comments":
			_ = json.NewEncoder(w).Encode([]GitHubComment{
				{Body: "Comment on issue 1", User: GitHubUser{Login: "reviewer1"}},
			})
		case "/repos/octocat/Hello-World/pulls/10/comments":
			_ = json.NewEncoder(w).Encode([]GitHubComment{
				{Body: "PR review comment", User: GitHubUser{Login: "reviewer2"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	client := NewGitHubClient("test_token_xyz")
	client.SetBaseURL(mockServer.URL)

	// Fetch all issues across the 3 pages
	issues, err := client.ListIssues("octocat", "Hello-World")
	if err != nil {
		t.Fatalf("ListIssues failed: %v", err)
	}

	if len(issues) != 5 {
		t.Fatalf("expected 5 paginated issues, got %d", len(issues))
	}
	if issues[0].Title != "Issue 1" || issues[4].Title != "Issue 5" {
		t.Errorf("issues order mismatch: first=%s, last=%s", issues[0].Title, issues[4].Title)
	}

	// Verify headers sent to mock server
	if len(authHeaders) == 0 || authHeaders[0] != "Bearer test_token_xyz" {
		t.Errorf("expected Bearer token in Authorization header, got %q", authHeaders[0])
	}
	if len(userAgentHeaders) == 0 || userAgentHeaders[0] != "exist.io-migration" {
		t.Errorf("expected exist.io-migration User-Agent, got %q", userAgentHeaders[0])
	}

	// Verify Labels
	labels, err := client.ListLabels("octocat", "Hello-World")
	if err != nil || len(labels) != 2 {
		t.Fatalf("ListLabels: err=%v, count=%d", err, len(labels))
	}

	// Verify PRs
	prs, err := client.ListPullRequests("octocat", "Hello-World")
	if err != nil || len(prs) != 1 {
		t.Fatalf("ListPullRequests: err=%v, count=%d", err, len(prs))
	}

	// Verify Issue Comments
	comments, err := client.ListIssueComments("octocat", "Hello-World", 1)
	if err != nil || len(comments) != 1 {
		t.Fatalf("ListIssueComments: err=%v, count=%d", err, len(comments))
	}

	// Verify PR Comments
	prComments, err := client.ListPRComments("octocat", "Hello-World", 10)
	if err != nil || len(prComments) != 1 {
		t.Fatalf("ListPRComments: err=%v, count=%d", err, len(prComments))
	}
}

// TestMigrateStatus_Lifecycle tests tracking migration status transitions from start to completion and on failure.
func TestMigrateStatus_Lifecycle(t *testing.T) {
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

	ctx := context.Background()

	// 1. Verify querying non-existent migration returns 404
	reqNonExistent := httptest.NewRequest(http.MethodGet, "/api/v1/repos/migrate/mig_nonexistent_123/status", nil)
	recNonExistent := httptest.NewRecorder()
	application.Handler().ServeHTTP(recNonExistent, reqNonExistent)
	if recNonExistent.Code != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent job, got %d", recNonExistent.Code)
	}

	// 2. Prepare namespace 'acme'
	ns := store.Namespace{
		ID:   idgen.NewNS(),
		Slug: "acme",
		Name: "Acme Corp",
		Type: "team",
	}
	if err := application.Store.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	// 3. Prepare a local git repository to serve as the migration clone source
	srcDir := filepath.Join(t.TempDir(), "source-repo")
	srcRepo, err := git.PlainInit(srcDir, false)
	if err != nil {
		t.Fatalf("PlainInit: %v", err)
	}
	wt, err := srcRepo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}
	readmePath := filepath.Join(srcDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Migrated Project\nHello from source forge!"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	sig := &object.Signature{
		Name:  "Migration Author",
		Email: "author@exist.io",
		When:  time.Now(),
	}
	commitHash, err := wt.Commit("Initial import commit", &git.CommitOptions{Author: sig, Committer: sig})
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// 4. Mock GitHub API server for issues, PRs, comments, and labels
	closedAt := "2026-01-05T12:00:00Z"
	mergedAt := "2026-01-06T15:00:00Z"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/labels"):
			_ = json.NewEncoder(w).Encode([]GitHubLabel{
				{Name: "bug", Color: "d73a4a"},
				{Name: "enhancement", Color: "a2eeef"},
			})
		case strings.Contains(r.URL.Path, "/issues/1/comments"):
			_ = json.NewEncoder(w).Encode([]GitHubComment{
				{Body: "Investigating the crash in production", User: GitHubUser{Login: "octocat"}, CreatedAt: "2026-01-04T12:00:00Z"},
			})
		case strings.Contains(r.URL.Path, "/issues"):
			_ = json.NewEncoder(w).Encode([]GitHubIssue{
				{
					Number:    1,
					Title:     "Imported Bug Issue",
					Body:      "Issue description text",
					State:     "open",
					User:      GitHubUser{Login: "octocat"},
					Labels:    []GitHubLabel{{Name: "bug", Color: "d73a4a"}},
					CreatedAt: "2026-01-04T10:00:00Z",
					UpdatedAt: "2026-01-04T11:00:00Z",
				},
				{
					Number:    2,
					Title:     "Closed Feature Request",
					Body:      "Already resolved",
					State:     "closed",
					User:      GitHubUser{Login: "octocat"},
					CreatedAt: "2026-01-04T10:00:00Z",
					UpdatedAt: "2026-01-05T12:00:00Z",
					ClosedAt:  &closedAt,
				},
				{
					Number: 3,
					Title:  "PR in issues endpoint",
					Body:   "Must be skipped during issue import",
					State:  "open",
					User:   GitHubUser{Login: "octocat"},
					PR:     &GitHubPRRef{URL: "https://api.github.com/repos/acme/migrated-repo/pulls/3"},
				},
			})
		case strings.Contains(r.URL.Path, "/pulls/10/comments"):
			_ = json.NewEncoder(w).Encode([]GitHubComment{
				{Body: "LGTM! Ready to merge.", User: GitHubUser{Login: "octocat"}, CreatedAt: "2026-01-06T14:00:00Z"},
			})
		case strings.Contains(r.URL.Path, "/pulls"):
			_ = json.NewEncoder(w).Encode([]GitHubPR{
				{
					Number:    10,
					Title:     "Support OAuth2 login",
					Body:      "PR description text",
					State:     "closed",
					Merged:    true,
					MergedAt:  &mergedAt,
					User:      GitHubUser{Login: "octocat"},
					Head:      GitHubBranch{Ref: "feature/oauth2", SHA: commitHash.String()},
					Base:      GitHubBranch{Ref: "main", SHA: commitHash.String()},
					CreatedAt: "2026-01-06T10:00:00Z",
					UpdatedAt: "2026-01-06T15:00:00Z",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	SetGitHubBaseURLForTesting(mockServer.URL)
	defer SetGitHubBaseURLForTesting("")

	// 5. Start migration request
	migratePayload := MigrateRequest{
		CloneURL:     srcDir,
		Owner:        "acme",
		RepoName:     "migrated-repo",
		GitHubToken:  "ghp_test_token_12345",
		ImportIssues: true,
		ImportPRs:    true,
		Private:      false,
	}

	bodyBytes, _ := json.Marshal(migratePayload)
	reqPost := httptest.NewRequest(http.MethodPost, "/api/v1/repos/migrate", bytes.NewReader(bodyBytes))
	reqPost.Header.Set("Content-Type", "application/json")
	recPost := httptest.NewRecorder()
	application.Handler().ServeHTTP(recPost, reqPost)

	if recPost.Code != http.StatusAccepted {
		t.Fatalf("POST /api/v1/repos/migrate status = %d, want 202; body = %s", recPost.Code, recPost.Body.String())
	}

	var respMap map[string]any
	if err := json.NewDecoder(recPost.Body).Decode(&respMap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	jobID, ok := respMap["id"].(string)
	if !ok || jobID == "" {
		t.Fatalf("missing or invalid job id in response: %+v", respMap)
	}

	// 6. Poll status until complete
	var finalStatus MigrateStatus
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		reqStatus := httptest.NewRequest(http.MethodGet, "/api/v1/repos/migrate/"+jobID+"/status", nil)
		recStatus := httptest.NewRecorder()
		application.Handler().ServeHTTP(recStatus, reqStatus)

		if recStatus.Code == http.StatusOK {
			var st MigrateStatus
			if err := json.NewDecoder(recStatus.Body).Decode(&st); err == nil {
				if st.Phase == PhaseComplete || st.Phase == PhaseFailed {
					finalStatus = st
					break
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	if finalStatus.Phase != PhaseComplete {
		t.Fatalf("migration failed to complete: phase = %q, error = %q, progress = %d",
			finalStatus.Phase, finalStatus.Error, finalStatus.Progress)
	}
	if finalStatus.Progress != 100 {
		t.Errorf("expected progress 100, got %d", finalStatus.Progress)
	}

	// 7. Verify repository metadata and cloned git objects
	repo, err := application.Store.GetRepo(ctx, "acme", "migrated-repo")
	if err != nil {
		t.Fatalf("GetRepo: %v", err)
	}
	if repo.FullName != "acme/migrated-repo" {
		t.Errorf("repo FullName = %q, want acme/migrated-repo", repo.FullName)
	}

	// Verify bare repo exists in gitstore
	clonedGitRepo, err := application.Git.Open("acme", "migrated-repo")
	if err != nil {
		t.Fatalf("Git.Open failed: %v", err)
	}
	headRef, err := clonedGitRepo.Head()
	if err != nil {
		t.Fatalf("Head() failed: %v", err)
	}
	if headRef.Hash() != commitHash {
		t.Errorf("cloned HEAD hash = %s, want %s", headRef.Hash(), commitHash)
	}

	// 8. Verify issues and comments imported
	issues, err := application.Store.ListIssues(ctx, repo.ID, "all")
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("expected 2 imported issues, got %d", len(issues))
	}

	// Check issue 1 comments
	issue1Comments, err := application.Store.ListIssueComments(ctx, issues[1].ID) // Ordered by number desc so index 1 is #1
	if err == nil && len(issue1Comments) == 0 {
		// Try the other index if order differs
		issue1Comments, _ = application.Store.ListIssueComments(ctx, issues[0].ID)
	}
	if len(issue1Comments) != 1 {
		t.Errorf("expected 1 comment for issue #1, got %d", len(issue1Comments))
	}

	// 9. Verify PR and comments imported
	pr, err := application.Store.GetPullRequest(ctx, repo.ID, 10)
	if err != nil {
		t.Fatalf("GetPullRequest(#10): %v", err)
	}
	if pr.Title != "Support OAuth2 login" {
		t.Errorf("PR title = %q, want 'Support OAuth2 login'", pr.Title)
	}
	if !pr.Merged {
		t.Errorf("expected PR #10 to be marked as merged")
	}

	prComments, err := application.Store.ListComments(ctx, pr.ID)
	if err != nil || len(prComments) != 1 {
		t.Errorf("expected 1 PR review comment, got %d (err: %v)", len(prComments), err)
	}

	// 10. Verify labels imported
	labels, err := application.Store.ListLabels(ctx, ns.ID)
	if err != nil {
		t.Fatalf("ListLabels: %v", err)
	}
	if len(labels) < 2 {
		t.Errorf("expected at least 2 labels, got %d", len(labels))
	}

	// 11. Test failure lifecycle when cloning an invalid remote
	failPayload := MigrateRequest{
		CloneURL: "file:///path/to/nonexistent/repository.git",
		Owner:    "acme",
		RepoName: "failed-repo",
	}
	failBytes, _ := json.Marshal(failPayload)
	reqFail := httptest.NewRequest(http.MethodPost, "/api/v1/repos/migrate", bytes.NewReader(failBytes))
	reqFail.Header.Set("Content-Type", "application/json")
	recFail := httptest.NewRecorder()
	application.Handler().ServeHTTP(recFail, reqFail)

	if recFail.Code != http.StatusAccepted {
		t.Fatalf("POST for failing clone status = %d, want 202", recFail.Code)
	}
	var failResp map[string]any
	_ = json.NewDecoder(recFail.Body).Decode(&failResp)
	failJobID := failResp["id"].(string)

	var failStatus MigrateStatus
	failDeadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(failDeadline) {
		reqStatus := httptest.NewRequest(http.MethodGet, "/api/v1/repos/migrate/"+failJobID+"/status", nil)
		recStatus := httptest.NewRecorder()
		application.Handler().ServeHTTP(recStatus, reqStatus)
		if recStatus.Code == http.StatusOK {
			var st MigrateStatus
			if err := json.NewDecoder(recStatus.Body).Decode(&st); err == nil && st.Phase == PhaseFailed {
				failStatus = st
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	if failStatus.Phase != PhaseFailed {
		t.Fatalf("expected failed phase, got %q", failStatus.Phase)
	}
	if failStatus.Error == "" {
		t.Errorf("expected non-empty error in failed migration status")
	}
}
