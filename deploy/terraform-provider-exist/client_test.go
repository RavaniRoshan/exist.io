package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"exist.io/internal/app"
	"exist.io/internal/auth"
	"exist.io/internal/config"
	"exist.io/pkg/scopes"
)

type omnipotentToken struct{}

func (t omnipotentToken) Type() string     { return scopes.TypeUser }
func (t omnipotentToken) Scopes() []string { return scopes.All() }
func (t omnipotentToken) Subject() string  { return "user_tf_admin" }

func TestTerraformClientIntegration(t *testing.T) {
	dataDir := t.TempDir()
	cfg := config.Config{
		DataDir:    filepath.Join(dataDir, "meta"),
		PublicURL:  "http://127.0.0.1:8080/v1/origin",
		GitURL:     "127.0.0.1:8080",
		ListenAddr: "127.0.0.1:8080",
		SSHAddr:    "none",
	}

	application, err := app.Open(cfg)
	if err != nil {
		t.Fatalf("app.Open: %v", err)
	}
	defer application.Close()

	application.Origin.SetResolver(func(*http.Request) (auth.Token, error) {
		return omnipotentToken{}, nil
	})

	server := httptest.NewServer(application.Handler())
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	// 1. Create Namespace
	ns, err := client.CreateNamespace(ctx, "acmecorp", "Acme Corporation", "team")
	if err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	if ns.Slug != "acmecorp" {
		t.Fatalf("unexpected slug: %s", ns.Slug)
	}

	// 2. Get Namespace
	fetchedNS, err := client.GetNamespace(ctx, "acmecorp")
	if err != nil {
		t.Fatalf("GetNamespace: %v", err)
	}
	if fetchedNS.Name != "Acme Corporation" {
		t.Fatalf("unexpected name: %s", fetchedNS.Name)
	}

	// 3. Create Repository
	repo, err := client.CreateRepository(ctx, "acmecorp", "core-engine", "main", "private")
	if err != nil {
		t.Fatalf("CreateRepository: %v", err)
	}
	if repo.Name != "core-engine" {
		t.Fatalf("unexpected repo name: %s", repo.Name)
	}

	// 4. Get Repository
	fetchedRepo, err := client.GetRepository(ctx, "acmecorp", "core-engine")
	if err != nil {
		t.Fatalf("GetRepository: %v", err)
	}
	if fetchedRepo.DefaultBranch != "main" {
		t.Fatalf("unexpected default branch: %s", fetchedRepo.DefaultBranch)
	}

	// 5. Add IP Allowlist entry
	_, err = client.AddIPAllowlistEntry(ctx, "acmecorp", "192.168.1.0/24", "Corporate VPN")
	if err != nil {
		t.Fatalf("AddIPAllowlistEntry: %v", err)
	}

	// 6. Create Ruleset
	rulesetPayload := map[string]any{
		"name":             "protect-main",
		"enforcement":      "active",
		"kind":             "push_branch",
		"includedRefNames": []string{"refs/heads/main"},
		"rules": []map[string]any{
			{"type": "require_pull_request", "parameters": map[string]any{"requiredApprovals": 1}},
			{"type": "require_signed_commits"},
			{"type": "prevent_force_push"},
		},
	}
	_, err = client.CreateRuleset(ctx, "acmecorp", "core-engine", rulesetPayload)
	if err != nil {
		t.Fatalf("CreateRuleset: %v", err)
	}
}
