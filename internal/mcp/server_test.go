package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
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
