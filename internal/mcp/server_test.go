package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func setupTestMCP(t *testing.T) (*Server, *store.DB, *gitstore.Store) {
	t.Helper()
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	git, err := gitstore.New(filepath.Join(dir, "git"))
	if err != nil {
		t.Fatalf("init gitstore: %v", err)
	}

	srv := New(db, git)
	return srv, db, git
}

func callRPC(handler http.Handler, reqBody map[string]any) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestMCP_Initialize(t *testing.T) {
	srv, _, _ := setupTestMCP(t)
	h := srv.Handler(false)

	rec := callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]any{},
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("initialize status: %d", rec.Code)
	}

	var resp rpcResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	res, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result: %+v", resp)
	}
	if res["protocolVersion"] != "2024-11-05" {
		t.Fatalf("expected protocolVersion 2024-11-05, got: %v", res["protocolVersion"])
	}
}

func TestMCP_ToolsList_ReadWrite_And_ReadOnly(t *testing.T) {
	srv, _, _ := setupTestMCP(t)

	// 1. Read-write handler (/mcp)
	rwHandler := srv.Handler(false)
	rec := callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("tools/list status: %d", rec.Code)
	}

	var resp rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res := resp.Result.(map[string]any)
	tools := res["tools"].([]any)
	if len(tools) < 33 {
		t.Fatalf("expected at least 33 tools, got %d", len(tools))
	}

	// 2. Read-only handler (/mcp/readonly)
	roHandler := srv.Handler(true)
	recRO := callRPC(roHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/list",
	})
	if recRO.Code != http.StatusOK {
		t.Fatalf("tools/list readonly status: %d", recRO.Code)
	}

	var respRO rpcResponse
	_ = json.Unmarshal(recRO.Body.Bytes(), &respRO)
	resRO := respRO.Result.(map[string]any)
	toolsRO := resRO["tools"].([]any)

	// Ensure no write tools are present in readonly list
	for _, rawT := range toolsRO {
		tDef := rawT.(map[string]any)
		name := tDef["name"].(string)
		if name == "create_repository" || name == "merge_pull_request" || name == "create_pull_request" {
			t.Fatalf("found write tool %q in /mcp/readonly tools/list", name)
		}
	}
}

func TestMCP_ToolCall_RepositoryAndReadOnlyGuard(t *testing.T) {
	srv, db, _ := setupTestMCP(t)
	ctx := context.Background()

	// Seed namespace
	ns := store.Namespace{
		ID:   idgen.NewNS(),
		Slug: "acme",
		Type: "team",
		Name: "Acme Corp",
	}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("create namespace: %v", err)
	}

	// 1. Call create_repository on read-only endpoint -> should return isError: true
	roHandler := srv.Handler(true)
	rec := callRPC(roHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      10,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_repository",
			"arguments": map[string]any{
				"ownerSlug": "acme",
				"name":      "rocket",
			},
		},
	})
	var resp rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res := resp.Result.(map[string]any)
	if res["isError"] != true {
		t.Fatalf("expected write tool on readonly to fail with isError=true, got: %+v", res)
	}

	// 2. Call create_repository on read-write endpoint -> should succeed
	rwHandler := srv.Handler(false)
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      11,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_repository",
			"arguments": map[string]any{
				"ownerSlug": "acme",
				"name":      "rocket",
			},
		},
	})
	var respCreate rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &respCreate)
	resCreate := respCreate.Result.(map[string]any)
	if resCreate["isError"] == true {
		t.Fatalf("create_repository failed: %+v", resCreate)
	}

	// 3. Call get_repository -> should return repo metadata
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      12,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "get_repository",
			"arguments": map[string]any{
				"ownerSlug": "acme",
				"repoName":  "rocket",
			},
		},
	})
	var respGet rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &respGet)
	resGet := respGet.Result.(map[string]any)
	if resGet["isError"] == true {
		t.Fatalf("get_repository failed: %+v", resGet)
	}

	// 4. Call list_repositories
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      13,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "list_repositories",
			"arguments": map[string]any{
				"ownerSlug": "acme",
			},
		},
	})
	var respList rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &respList)
	resList := respList.Result.(map[string]any)
	if resList["isError"] == true {
		t.Fatalf("list_repositories failed: %+v", resList)
	}
}

func TestMCP_LeaseTools(t *testing.T) {
	srv, _, _ := setupTestMCP(t)
	srv.SetLeaseSecret([]byte("test-super-secret-hmac-key-12345678"))
	rwHandler := srv.Handler(false)

	// 1. Request lease
	rec := callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      20,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "lease_request",
			"arguments": map[string]any{
				"agentId":    "agent-claude",
				"repoId":     "repo-123",
				"ttlSeconds": 300,
			},
		},
	})
	var resp rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res := resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("lease_request failed: %+v", res)
	}

	content := res["content"].([]any)[0].(map[string]any)["text"].(string)
	var leaseData map[string]any
	if err := json.Unmarshal([]byte(content), &leaseData); err != nil {
		t.Fatalf("unmarshal leaseData: %v", err)
	}
	token := leaseData["token"].(string)
	leaseID := leaseData["id"].(string)
	if token == "" || leaseID == "" {
		t.Fatalf("empty token or leaseID in lease_request response")
	}

	// 2. Inspect lease
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      21,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "lease_inspect",
			"arguments": map[string]any{
				"token": token,
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("lease_inspect failed: %+v", res)
	}
	content = res["content"].([]any)[0].(map[string]any)["text"].(string)
	var inspectData map[string]any
	_ = json.Unmarshal([]byte(content), &inspectData)
	if inspectData["agentId"] != "agent-claude" || inspectData["revoked"] != false {
		t.Fatalf("unexpected inspect data: %+v", inspectData)
	}

	// 3. Revoke lease
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      22,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "lease_revoke",
			"arguments": map[string]any{
				"leaseId": leaseID,
				"reason":  "task completed",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("lease_revoke failed: %+v", res)
	}

	// 4. Re-inspect -> revoked should be true
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      23,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "lease_inspect",
			"arguments": map[string]any{
				"token": token,
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	content = res["content"].([]any)[0].(map[string]any)["text"].(string)
	_ = json.Unmarshal([]byte(content), &inspectData)
	if inspectData["revoked"] != true {
		t.Fatalf("expected revoked=true, got %+v", inspectData)
	}
}

func TestMCP_CanonicalAgentLoopTools(t *testing.T) {
	srv, db, git := setupTestMCP(t)
	ctx := context.Background()
	rwHandler := srv.Handler(false)

	// 1. Seed namespace and repo
	nsID := idgen.NewNS()
	if err := db.CreateNamespace(ctx, store.Namespace{
		ID:   nsID,
		Slug: "agentcorp",
		Type: "team",
		Name: "Agent Corp",
	}); err != nil {
		t.Fatal(err)
	}

	repoID := idgen.NewRepo()
	if err := db.CreateRepo(ctx, store.Repo{
		ID:            repoID,
		NamespaceID:   nsID,
		Name:          "webforge",
		FullName:      "agentcorp/webforge",
		DefaultBranch: "main",
		Visibility:    "public",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := git.Init("agentcorp", "webforge", "main"); err != nil {
		t.Fatal(err)
	}

	sig := object.Signature{Name: "Dev", Email: "dev@exist.io", When: time.Now().UTC()}
	c1, err := git.WriteFiles("agentcorp", "webforge", "main", []gitstore.FileChange{
		{Path: "main.go", Content: "package main\n\nfunc main() {}\n"},
	}, sig, "Initial commit")
	if err != nil {
		t.Fatal(err)
	}

	// 2. Branch refs/heads/agent/feature
	if _, err := git.CreateBranchRef("agentcorp", "webforge", "refs/heads/agent/feature", c1.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := git.WriteFiles("agentcorp", "webforge", "agent/feature", []gitstore.FileChange{
		{Path: "agent.go", Content: "package main\n\nfunc RunAgent() string {\n\treturn \"autonomous\"\n}\n"},
	}, sig, "Add agent.go"); err != nil {
		t.Fatal(err)
	}

	// 3. Test repo_tree
	rec := callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      30,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "repo_tree",
			"arguments": map[string]any{
				"ownerSlug": "agentcorp",
				"repoName":  "webforge",
				"ref":       "agent/feature",
			},
		},
	})
	var resp rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res := resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("repo_tree failed: %+v", res)
	}

	// 4. Test file_read
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      31,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "file_read",
			"arguments": map[string]any{
				"ownerSlug": "agentcorp",
				"repoName":  "webforge",
				"ref":       "agent/feature",
				"path":      "agent.go",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("file_read failed: %+v", res)
	}

	// 5. Test diff_compare
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      32,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "diff_compare",
			"arguments": map[string]any{
				"ownerSlug": "agentcorp",
				"repoName":  "webforge",
				"base":      "main",
				"head":      "agent/feature",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("diff_compare failed: %+v", res)
	}

	// 6. Test symbol_search
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      33,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "symbol_search",
			"arguments": map[string]any{
				"ownerSlug": "agentcorp",
				"repoName":  "webforge",
				"query":     "RunAgent",
				"ref":       "agent/feature",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("symbol_search failed: %+v", res)
	}

	// 7. Test pull_request_create
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      34,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "pull_request_create",
			"arguments": map[string]any{
				"ownerSlug": "agentcorp",
				"repoName":  "webforge",
				"title":     "Add autonomous agent",
				"headRef":   "agent/feature",
				"baseRef":   "main",
				"body":      "Implements RunAgent function",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("pull_request_create failed: %+v", res)
	}

	// 8. Test thread_comment
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      35,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "thread_comment",
			"arguments": map[string]any{
				"ownerSlug":  "agentcorp",
				"repoName":   "webforge",
				"pullNumber": 1,
				"body":       "Looks great, verified AST symbols and signatures.",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("thread_comment failed: %+v", res)
	}

	// 9. Add ruleset requiring check 'ci/test'
	headSHA, _ := git.ResolveRef("agentcorp", "webforge", "agent/feature")
	_, _ = db.CreateRuleset(ctx, store.Ruleset{
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
						map[string]any{"name": "ci/test", "context": "ci/test"},
					},
				},
			},
		},
	})

	// 10. Test pull_request_merge BEFORE status check -> MUST FAIL due to blocker!
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      36,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "pull_request_merge",
			"arguments": map[string]any{
				"ownerSlug":  "agentcorp",
				"repoName":   "webforge",
				"pullNumber": 1,
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] != true {
		t.Fatalf("expected pull_request_merge to be blocked by missing check run, but it succeeded: %+v", res)
	}

	// 11. Test status_check_report -> report passing ci/test check run
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      37,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "status_check_report",
			"arguments": map[string]any{
				"ownerSlug":  "agentcorp",
				"repoName":   "webforge",
				"headSha":    headSHA.String(),
				"name":       "ci/test",
				"status":     "completed",
				"conclusion": "success",
				"title":      "Unit tests",
				"summary":    "All 42 tests passed in 12ms",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("status_check_report failed: %+v", res)
	}

	// 12. Test pull_request_merge AFTER status check passes -> MUST SUCCEED!
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      38,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "pull_request_merge",
			"arguments": map[string]any{
				"ownerSlug":  "agentcorp",
				"repoName":   "webforge",
				"pullNumber": 1,
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("pull_request_merge after passing check failed: %+v", res)
	}
}
