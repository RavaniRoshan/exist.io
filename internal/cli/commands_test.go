package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLI_VersionAndHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	c := &CLI{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	if err := c.Run([]string{"--version"}); err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(stdout.String(), "exist version") {
		t.Errorf("expected version string, got: %s", stdout.String())
	}

	stdout.Reset()
	if err := c.Run([]string{"--help"}); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(stdout.String(), "exist <command>") {
		t.Errorf("expected help string, got: %s", stdout.String())
	}
}

func TestCLI_RepoCommands(t *testing.T) {
	var requestedPath, requestedMethod string
	var requestedBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedMethod = r.Method
		requestedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:   ts.URL,
		Token:      "test_tok_123",
		HTTPClient: ts.Client(),
		Stdout:     &stdout,
		Stderr:     &stderr,
	}

	// 1. repo list
	if err := c.Run([]string{"repo", "list", "--namespace", "acme"}); err != nil {
		t.Fatalf("repo list: %v", err)
	}
	if requestedPath != "/repos/acme" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /repos/acme, got %s %s", requestedMethod, requestedPath)
	}

	// 2. repo view
	if err := c.Run([]string{"repo", "view", "acme/rocket"}); err != nil {
		t.Fatalf("repo view: %v", err)
	}
	if requestedPath != "/repos/acme/rocket" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /repos/acme/rocket, got %s %s", requestedMethod, requestedPath)
	}

	// 3. repo create
	if err := c.Run([]string{"repo", "create", "acme/rocket", "--default-branch", "develop"}); err != nil {
		t.Fatalf("repo create: %v", err)
	}
	if requestedPath != "/repos/acme" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /repos/acme, got %s %s", requestedMethod, requestedPath)
	}
	var createReq map[string]any
	_ = json.Unmarshal(requestedBody, &createReq)
	repoMap, ok := createReq["repo"].(map[string]any)
	if !ok || repoMap["name"] != "rocket" || repoMap["defaultBranch"] != "develop" {
		t.Errorf("unexpected body: %+v", createReq)
	}

	// 4. repo mirror sync
	if err := c.Run([]string{"repo", "mirror", "sync", "acme/rocket", "--ref", "refs/heads/main"}); err != nil {
		t.Fatalf("repo mirror sync: %v", err)
	}
	if requestedPath != "/repos/acme/rocket:syncMirror" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /repos/acme/rocket:syncMirror, got %s %s", requestedMethod, requestedPath)
	}
}

func TestCLI_PRCommands(t *testing.T) {
	var requestedPath, requestedMethod string
	var requestedBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedMethod = r.Method
		requestedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:    ts.URL,
		Token:       "test_tok_123",
		DefaultRepo: "acme/rocket",
		HTTPClient:  ts.Client(),
		Stdout:      &stdout,
		Stderr:      &stderr,
	}

	// 1. pr list
	if err := c.Run([]string{"pr", "list"}); err != nil {
		t.Fatalf("pr list: %v", err)
	}
	if requestedPath != "/repos/acme/rocket/pulls" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /repos/acme/rocket/pulls, got %s %s", requestedMethod, requestedPath)
	}

	// 2. pr merge posts to .../merge
	if err := c.Run([]string{"pr", "merge", "42", "--method", "squash"}); err != nil {
		t.Fatalf("pr merge: %v", err)
	}
	if requestedPath != "/repos/acme/rocket/pulls/42/merge" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /repos/acme/rocket/pulls/42/merge, got %s %s", requestedMethod, requestedPath)
	}
	var mergeReq map[string]any
	_ = json.Unmarshal(requestedBody, &mergeReq)
	if mergeReq["mergeMethod"] != "squash" {
		t.Errorf("expected mergeMethod squash, got %+v", mergeReq)
	}
}

func TestCLI_APICommand(t *testing.T) {
	var requestedPath, requestedMethod string
	var requestedHeader string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedMethod = r.Method
		requestedHeader = r.Header.Get("Authorization")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"rateLimit": {"limit": 5000}}`))
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:   ts.URL,
		Token:      "my_secret_token",
		HTTPClient: ts.Client(),
		Stdout:     &stdout,
		Stderr:     &stderr,
	}

	if err := c.Run([]string{"api", "/rate_limit"}); err != nil {
		t.Fatalf("api: %v", err)
	}
	if requestedPath != "/rate_limit" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /rate_limit, got %s %s", requestedMethod, requestedPath)
	}
	if requestedHeader != "Bearer my_secret_token" {
		t.Errorf("expected bearer token header, got: %s", requestedHeader)
	}
}
