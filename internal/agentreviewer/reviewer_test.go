package agentreviewer

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func author() object.Signature {
	return object.Signature{
		Name:  "Test Dev",
		Email: "dev@exist.io",
		When:  time.Now(),
	}
}

func TestAgentReviewer_CleanPR(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	gs, err := gitstore.New(filepath.Join(t.TempDir(), "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	owner, repoName := "corp", "services"
	if _, err := gs.Init(owner, repoName, "main"); err != nil {
		t.Fatalf("gs.Init: %v", err)
	}

	// Base commit on main
	c1, err := gs.WriteFiles(owner, repoName, "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# Services\nInitial setup"},
	}, author(), "initial commit")
	if err != nil {
		t.Fatalf("commit 1: %v", err)
	}

	// Branch feat
	if _, err := gs.CreateBranchRef(owner, repoName, "refs/heads/feat", c1.String()); err != nil {
		t.Fatalf("create branch: %v", err)
	}

	// Commit on feat with tests
	c2, err := gs.WriteFiles(owner, repoName, "feat", []gitstore.FileChange{
		{Path: "service.go", Content: "package services\n\nfunc Hello() string { return \"world\" }\n"},
		{Path: "service_test.go", Content: "package services\n\nimport \"testing\"\n\nfunc TestHello(t *testing.T) {}\n"},
	}, author(), "add service and test")
	if err != nil {
		t.Fatalf("commit 2: %v", err)
	}

	ns := store.Namespace{ID: "ns_1", Slug: owner, Type: "organization", Name: owner}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	repo := store.Repo{
		ID:            "repo_1",
		NamespaceID:   ns.ID,
		FullName:      owner + "/" + repoName,
		Name:          repoName,
		DefaultBranch: "main",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	pr := store.PullRequest{
		ID:      "pr_1",
		RepoID:  repo.ID,
		Number:  1,
		Title:   "Add Hello Service",
		HeadRef: "feat",
		HeadSHA: c2.String(),
		BaseRef: "main",
		BaseSHA: c1.String(),
		State:   "open",
	}
	if err := db.CreatePullRequest(ctx, &pr); err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}

	res, err := ReviewPullRequest(ctx, db, gs, owner, repo, pr)
	if err != nil {
		t.Fatalf("ReviewPullRequest: %v", err)
	}

	if res.Verdict != "approve" {
		t.Errorf("expected verdict 'approve', got %q", res.Verdict)
	}
	if res.Conclusion != "success" {
		t.Errorf("expected conclusion 'success', got %q", res.Conclusion)
	}
	if res.Score != 100 {
		t.Errorf("expected score 100, got %d", res.Score)
	}

	// Verify review was inserted in DB
	reviews, err := db.ListReviews(ctx, pr.ID)
	if err != nil {
		t.Fatalf("ListReviews: %v", err)
	}
	if len(reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(reviews))
	}
	if reviews[0].State != "approve" {
		t.Errorf("expected review state 'approve', got %q", reviews[0].State)
	}
}

func TestAgentReviewer_DetectLeakedSecret(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	gs, err := gitstore.New(filepath.Join(t.TempDir(), "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	owner, repoName := "security", "vault"
	if _, err := gs.Init(owner, repoName, "main"); err != nil {
		t.Fatalf("gs.Init: %v", err)
	}

	c1, err := gs.WriteFiles(owner, repoName, "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# Vault"},
	}, author(), "initial commit")
	if err != nil {
		t.Fatalf("commit 1: %v", err)
	}

	if _, err := gs.CreateBranchRef(owner, repoName, "refs/heads/leak-branch", c1.String()); err != nil {
		t.Fatalf("create branch: %v", err)
	}

	// Commit containing an AWS key and a Slack token
	c2, err := gs.WriteFiles(owner, repoName, "leak-branch", []gitstore.FileChange{
		{Path: "config.go", Content: "package config\n\nconst AWSKey = \"AKIA1234567890ABCDEF\"\nconst Slack = \"xoxb-123456789012\"\n"},
	}, author(), "add config")
	if err != nil {
		t.Fatalf("commit 2: %v", err)
	}

	ns := store.Namespace{ID: "ns_2", Slug: owner, Type: "organization", Name: owner}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	repo := store.Repo{
		ID:            "repo_2",
		NamespaceID:   ns.ID,
		FullName:      owner + "/" + repoName,
		Name:          repoName,
		DefaultBranch: "main",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	pr := store.PullRequest{
		ID:      "pr_2",
		RepoID:  repo.ID,
		Number:  2,
		Title:   "Add Config",
		HeadRef: "leak-branch",
		HeadSHA: c2.String(),
		BaseRef: "main",
		BaseSHA: c1.String(),
		State:   "open",
	}
	if err := db.CreatePullRequest(ctx, &pr); err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}

	res, err := ReviewPullRequest(ctx, db, gs, owner, repo, pr)
	if err != nil {
		t.Fatalf("ReviewPullRequest: %v", err)
	}

	if res.Verdict != "request_changes" {
		t.Errorf("expected verdict 'request_changes', got %q", res.Verdict)
	}
	if res.Conclusion != "failure" {
		t.Errorf("expected conclusion 'failure', got %q", res.Conclusion)
	}
	if len(res.Violations) < 2 {
		t.Fatalf("expected at least 2 violations, got %d", len(res.Violations))
	}
}
