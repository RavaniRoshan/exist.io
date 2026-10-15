package store_test

import (
	"testing"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

func TestIssueLifecycle(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open store: %v", err)
	}
	defer db.Close()

	ctx := t.Context()
	ns := store.Namespace{ID: idgen.NewNS(), Slug: "testteam", Type: "team", Name: "Test Team"}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	repo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "forge",
		FullName:      "testteam/forge",
		DefaultBranch: "main",
		Visibility:    "public",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	// 1. Create Issue #1
	issue1 := &store.Issue{
		RepoID: repo.ID,
		Title:  "First issue: Setup continuous delivery",
		Body:   "We need air-gapped CI/CD with automated status checks.",
	}
	if err := db.CreateIssue(ctx, issue1); err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	if issue1.Number != 1 {
		t.Errorf("issue1.Number = %d, want 1", issue1.Number)
	}
	if issue1.State != "open" {
		t.Errorf("issue1.State = %q, want open", issue1.State)
	}

	// 2. Create Issue #2
	issue2 := &store.Issue{
		RepoID: repo.ID,
		Title:  "Second issue: Model Context Protocol agent lease",
		Body:   "Expose git store tools to autonomous agents.",
	}
	if err := db.CreateIssue(ctx, issue2); err != nil {
		t.Fatalf("CreateIssue 2: %v", err)
	}
	if issue2.Number != 2 {
		t.Errorf("issue2.Number = %d, want 2", issue2.Number)
	}

	// 3. List Issues
	issues, err := db.ListIssues(ctx, repo.ID, "")
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("ListIssues count = %d, want 2", len(issues))
	}

	// 4. Add Comment to Issue #1
	comment := &store.IssueComment{
		RepoID:  repo.ID,
		IssueID: issue1.ID,
		Body:    "Agent reviewer approved the architecture plan.",
	}
	if err := db.CreateIssueComment(ctx, comment); err != nil {
		t.Fatalf("CreateIssueComment: %v", err)
	}

	comments, err := db.ListIssueComments(ctx, issue1.ID)
	if err != nil {
		t.Fatalf("ListIssueComments: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("ListIssueComments count = %d, want 1", len(comments))
	}
	if comments[0].Body != comment.Body {
		t.Errorf("comment body = %q, want %q", comments[0].Body, comment.Body)
	}

	// 5. Close Issue #1
	if err := db.UpdateIssueState(ctx, repo.ID, 1, "closed"); err != nil {
		t.Fatalf("UpdateIssueState: %v", err)
	}
	got1, err := db.GetIssue(ctx, repo.ID, 1)
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}
	if got1.State != "closed" {
		t.Errorf("got1.State = %q, want closed", got1.State)
	}
	if got1.ClosedAt == "" {
		t.Errorf("got1.ClosedAt empty, want RFC3339 timestamp")
	}

	// 6. Filter by open
	openIssues, err := db.ListIssues(ctx, repo.ID, "open")
	if err != nil {
		t.Fatalf("ListIssues open: %v", err)
	}
	if len(openIssues) != 1 {
		t.Errorf("open count = %d, want 1", len(openIssues))
	}
}
