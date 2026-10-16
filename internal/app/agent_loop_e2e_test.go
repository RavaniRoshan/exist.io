package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"exist.io/internal/app"
	"exist.io/internal/config"
	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/agentlease"
	"exist.io/pkg/idgen"
	"github.com/go-git/go-git/v6/plumbing/object"
)

type mcpReq struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpResp struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type toolCallResult struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"isError"`
}

func callMCP(t *testing.T, client *http.Client, endpoint, toolName string, args map[string]any) toolCallResult {
	t.Helper()
	paramsBytes, err := json.Marshal(map[string]any{
		"name":      toolName,
		"arguments": args,
	})
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}

	reqEnvelope := mcpReq{
		JSONRPC: "2.0",
		ID:      time.Now().UnixNano(),
		Method:  "tools/call",
		Params:  paramsBytes,
	}
	bodyBytes, _ := json.Marshal(reqEnvelope)

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		t.Fatalf("mcp returned status %d", httpResp.StatusCode)
	}

	var envelope mcpResp
	if err := json.NewDecoder(httpResp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode mcp response: %v", err)
	}
	if envelope.Error != nil {
		t.Fatalf("mcp rpc error: code=%d message=%s", envelope.Error.Code, envelope.Error.Message)
	}

	var tcRes toolCallResult
	if err := json.Unmarshal(envelope.Result, &tcRes); err != nil {
		t.Fatalf("unmarshal tool call result: %v", err)
	}
	return tcRes
}

func TestAgentLoop_EndToEnd_RealServer(t *testing.T) {
	ctx := context.Background()
	dataDir := t.TempDir()

	// 1. Build real existd binary for git pre-receive hook execution
	existdBin := filepath.Join(t.TempDir(), "existd")
	buildCmd := exec.Command("go", "build", "-o", existdBin, "exist.io/cmd/existd")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build existd: %v\nOutput: %s", err, string(out))
	}
	t.Setenv("EXISTD_PATH", existdBin)

	// 2. Start full exist.io forge daemon with in-memory / temp state
	jwtSecret := "super-secret-agent-forge-jwt-key-32chars!"
	cfg := config.Config{
		ListenAddr: "127.0.0.1:0",
		DataDir:    dataDir,
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
		JWTSecret:  jwtSecret,
	}

	application, err := app.Open(cfg)
	if err != nil {
		t.Fatalf("app.Open: %v", err)
	}
	defer application.Close()

	srv := httptest.NewServer(application.Handler())
	defer srv.Close()

	mcpRWURL := srv.URL + "/mcp"
	mcpROURL := srv.URL + "/mcp/readonly"

	// 3. Seed repository with initial main commit
	nsID := idgen.NewNS()
	if err := application.Store.CreateNamespace(ctx, store.Namespace{
		ID:   nsID,
		Slug: "acme",
		Type: "team",
		Name: "Acme",
	}); err != nil {
		t.Fatal(err)
	}

	repoID := idgen.NewRepo()
	if err := application.Store.CreateRepo(ctx, store.Repo{
		ID:            repoID,
		NamespaceID:   nsID,
		Name:          "core",
		FullName:      "acme/core",
		DefaultBranch: "main",
		Visibility:    "public",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := application.Git.Init("acme", "core", "main"); err != nil {
		t.Fatal(err)
	}

	initSig := object.Signature{Name: "Dev", Email: "dev@acme.corp", When: time.Now().UTC()}
	if _, err := application.Git.WriteFiles("acme", "core", "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# Acme Core Library\n"},
		{Path: "service.go", Content: "package main\n\nfunc Version() string { return \"1.0.0\" }\n"},
	}, initSig, "Initial commit"); err != nil {
		t.Fatal(err)
	}

	httpClient := http.DefaultClient

	// 4. STEP 1: Agent requests capability lease via MCP
	leaseRes := callMCP(t, httpClient, mcpRWURL, "lease_request", map[string]any{
		"agentId":        "agent-claude",
		"repoId":         "acme/core",
		"allowedRefs":    []string{"refs/heads/agent/"},
		"disallowedRefs": []string{"refs/heads/main"},
		"ttlSeconds":     600,
	})
	if leaseRes.IsError {
		t.Fatalf("lease_request failed: %+v", leaseRes)
	}

	var leaseData struct {
		Token   string `json:"token"`
		ID      string `json:"id"`
		AgentID string `json:"agentId"`
	}
	if err := json.Unmarshal([]byte(leaseRes.Content[0].Text), &leaseData); err != nil {
		t.Fatalf("unmarshal lease response: %v", err)
	}
	if leaseData.Token == "" || leaseData.ID == "" {
		t.Fatalf("expected non-empty lease token and id: %+v", leaseData)
	}

	// 5. STEP 2: Agent inspects capability lease via MCP
	inspectRes := callMCP(t, httpClient, mcpRWURL, "lease_inspect", map[string]any{
		"token": leaseData.Token,
	})
	if inspectRes.IsError {
		t.Fatalf("lease_inspect failed: %+v", inspectRes)
	}

	// 6. NEGATIVE TEST: Attempting mutating operation on /mcp/readonly MUST FAIL
	roRes := callMCP(t, httpClient, mcpROURL, "pull_request_create", map[string]any{
		"ownerSlug": "acme",
		"repoName":  "core",
		"title":     "Unauthorized PR",
		"headRef":   "agent/test",
		"baseRef":   "main",
	})
	if !roRes.IsError {
		t.Fatalf("expected write tool on /mcp/readonly to be blocked, but succeeded")
	}

	// 7. STEP 3: Agent clones repo over HTTP using capability lease as password
	cloneDir := filepath.Join(t.TempDir(), "agent-worktree")
	authCloneURL := strings.Replace(srv.URL, "http://", fmt.Sprintf("http://agent:%s@", leaseData.Token), 1) + "/acme/core.git"

	cmd := exec.Command("git", "clone", authCloneURL, cloneDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone under capability lease failed: %v\nOutput: %s", err, string(out))
	}

	gitExec := func(args ...string) (string, error) {
		c := exec.Command("git", args...)
		c.Dir = cloneDir
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Claude Agent",
			"GIT_AUTHOR_EMAIL=claude@agent.exist.io",
			"GIT_COMMITTER_NAME=Claude Agent",
			"GIT_COMMITTER_EMAIL=claude@agent.exist.io",
		)
		out, err := c.CombinedOutput()
		return string(out), err
	}

	// 8. STEP 4: Agent branches to refs/heads/agent/patch-v1, adds code, and pushes
	if _, err := gitExec("checkout", "-b", "agent/patch-v1"); err != nil {
		t.Fatal(err)
	}
	newFile := filepath.Join(cloneDir, "agent_impl.go")
	if err := os.WriteFile(newFile, []byte("package main\n\nfunc AgentFeature() string { return \"ready\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := gitExec("add", "agent_impl.go"); err != nil {
		t.Fatal(err)
	}
	if _, err := gitExec("commit", "-m", "feat: implement autonomous agent feature"); err != nil {
		t.Fatal(err)
	}

	pushOut, err := gitExec("push", "origin", "agent/patch-v1:refs/heads/agent/patch-v1")
	if err != nil {
		t.Fatalf("push to refs/heads/agent/patch-v1 failed: %v\nOutput: %s", err, pushOut)
	}

	// 9. NEGATIVE TEST: Agent attempts to push to refs/heads/main -> MUST FAIL with GH007
	mainPushOut, err := gitExec("push", "origin", "agent/patch-v1:refs/heads/main")
	if err == nil {
		t.Fatalf("expected push to refs/heads/main to be rejected under lease, but succeeded: %s", mainPushOut)
	}
	if !strings.Contains(mainPushOut, "GH007") {
		t.Fatalf("expected GH007 lease violation in push output, got: %s", mainPushOut)
	}

	// 10. STEP 5: Agent opens PR via MCP
	prRes := callMCP(t, httpClient, mcpRWURL, "pull_request_create", map[string]any{
		"ownerSlug": "acme",
		"repoName":  "core",
		"title":     "feat: autonomous agent implementation",
		"headRef":   "agent/patch-v1",
		"baseRef":   "main",
		"body":      "Implements AgentFeature and unit verification.",
	})
	if prRes.IsError {
		t.Fatalf("pull_request_create failed: %+v", prRes)
	}

	// 11. STEP 6: Agent posts review thread comment via MCP
	commentRes := callMCP(t, httpClient, mcpRWURL, "thread_comment", map[string]any{
		"ownerSlug":  "acme",
		"repoName":   "core",
		"pullNumber": 1,
		"body":       "Self-review: AST symbol check passed, zero compiler errors.",
	})
	if commentRes.IsError {
		t.Fatalf("thread_comment failed: %+v", commentRes)
	}

	// 12. STEP 7: Add protection ruleset requiring check 'ci/build' to pass
	headSHA, _ := application.Git.ResolveRef("acme", "core", "agent/patch-v1")
	_, err = application.Store.CreateRuleset(ctx, store.Ruleset{
		ID:               idgen.New("rs"),
		RepoID:           repoID,
		Name:             "require-ci",
		Kind:             "merge_branch",
		Enforcement:      "active",
		IncludedRefNames: []string{"~DEFAULT_BRANCH"},
		Rules: []store.RulesetRule{
			{
				RuleType: "require_status_checks",
				Parameters: map[string]any{
					"requiredChecks": []any{
						map[string]any{"name": "ci/build", "context": "ci/build"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create ruleset: %v", err)
	}

	// 13. NEGATIVE TEST: Attempt merge BEFORE status check passes -> MUST FAIL
	blockedMerge := callMCP(t, httpClient, mcpRWURL, "pull_request_merge", map[string]any{
		"ownerSlug":  "acme",
		"repoName":   "core",
		"pullNumber": 1,
	})
	if !blockedMerge.IsError {
		t.Fatalf("expected merge to be blocked by failing status check, but succeeded: %+v", blockedMerge)
	}

	// 14. STEP 8: Report passing status check run via MCP
	checkReportRes := callMCP(t, httpClient, mcpRWURL, "status_check_report", map[string]any{
		"ownerSlug":  "acme",
		"repoName":   "core",
		"headSha":    headSHA.String(),
		"name":       "ci/build",
		"status":     "completed",
		"conclusion": "success",
		"title":      "Continuous Integration",
		"summary":    "All checks passed successfully",
	})
	if checkReportRes.IsError {
		t.Fatalf("status_check_report failed: %+v", checkReportRes)
	}

	// 15. STEP 9: Merge pull request through policy gate via MCP -> MUST SUCCEED
	mergeRes := callMCP(t, httpClient, mcpRWURL, "pull_request_merge", map[string]any{
		"ownerSlug":   "acme",
		"repoName":    "core",
		"pullNumber":  1,
		"mergeMethod": "merge",
	})
	if mergeRes.IsError {
		t.Fatalf("pull_request_merge failed: %+v", mergeRes)
	}

	// Verify PR state in database
	prObj, err := application.Store.GetPullRequest(ctx, repoID, 1)
	if err != nil {
		t.Fatalf("get pr: %v", err)
	}
	if !prObj.Merged || prObj.MergedAt == "" || prObj.MergeCommitSHA == "" {
		t.Fatalf("expected PR merged=true with non-empty mergedAt and mergeCommitSHA, got: %+v", prObj)
	}

	// 16. STEP 10: Revoke capability lease via MCP
	revokeRes := callMCP(t, httpClient, mcpRWURL, "lease_revoke", map[string]any{
		"leaseId": leaseData.ID,
		"reason":  "work finished successfully",
	})
	if revokeRes.IsError {
		t.Fatalf("lease_revoke failed: %+v", revokeRes)
	}

	// 17. NEGATIVE TEST: Push with revoked lease token -> MUST FAIL with 401
	revokedPushOut, err := gitExec("push", authCloneURL, "agent/patch-v1:refs/heads/agent/patch-v2")
	if err == nil {
		t.Fatalf("expected push with revoked lease to fail, but succeeded: %s", revokedPushOut)
	}

	// 18. NEGATIVE TEST: Push with expired lease token -> MUST FAIL with 401
	expiredTok, _, _ := agentlease.Issue([]byte(jwtSecret), "agent-claude", repoID, []string{"refs/heads/agent/"}, nil, -1*time.Minute)
	expiredURL := strings.Replace(srv.URL, "http://", fmt.Sprintf("http://agent:%s@", expiredTok), 1) + "/acme/core.git"
	expiredPushOut, err := gitExec("push", expiredURL, "agent/patch-v1:refs/heads/agent/patch-v3")
	if err == nil {
		t.Fatalf("expected push with expired lease to fail, but succeeded: %s", expiredPushOut)
	}
}
