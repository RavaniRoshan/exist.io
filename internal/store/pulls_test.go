package store

import (
	"context"
	"testing"

	"exist.io/pkg/idgen"
)

func seedNamespaceAndRepo(t *testing.T, db *DB) (Namespace, Repo) {
	t.Helper()
	ctx := context.Background()
	ns := Namespace{
		ID:   idgen.NewNS(),
		Slug: "acme-" + idgen.New("slug"),
		Type: "team",
		Name: "Acme Corp",
	}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("create namespace: %v", err)
	}

	repo := Repo{
		ID:                  idgen.NewRepo(),
		NamespaceID:         ns.ID,
		Name:                "rocket",
		FullName:            ns.Slug + "/rocket",
		DefaultBranch:       "main",
		Visibility:          "private",
		AllowMergeCommit:    true,
		AllowSquashMerge:    true,
		DeleteBranchOnMerge: false,
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("create repo: %v", err)
	}
	return ns, repo
}

func seedUser(t *testing.T, db *DB, id, email string) {
	t.Helper()
	_, err := db.sql.Exec(`INSERT INTO users (id, email, created_at) VALUES (?, ?, ?)`, id, email, now())
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func TestPullRequests_Lifecycle(t *testing.T) {
	db := openMem(t)
	ctx := context.Background()
	_, repo := seedNamespaceAndRepo(t, db)

	// Create PR 1
	pr1 := PullRequest{
		ID:        idgen.NewPR(),
		RepoID:    repo.ID,
		State:     "open",
		Draft:     false,
		Title:     "Add launch telemetry",
		Body:      "Adds structured launch telemetry",
		HeadRef:   "feature/telemetry",
		HeadSHA:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BaseRef:   "main",
		BaseSHA:   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Mergeable: "mergeable",
	}
	if err := db.CreatePullRequest(ctx, &pr1); err != nil {
		t.Fatalf("create pr1: %v", err)
	}
	if pr1.Number != 1 {
		t.Fatalf("expected PR #1, got %d", pr1.Number)
	}

	// Create PR 2
	pr2 := PullRequest{
		ID:        idgen.NewPR(),
		RepoID:    repo.ID,
		State:     "open",
		Draft:     true,
		Title:     "Draft feature",
		HeadRef:   "feature/draft",
		HeadSHA:   "cccccccccccccccccccccccccccccccccccccccc",
		BaseRef:   "main",
		BaseSHA:   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Mergeable: "blocked",
	}
	if err := db.CreatePullRequest(ctx, &pr2); err != nil {
		t.Fatalf("create pr2: %v", err)
	}
	if pr2.Number != 2 {
		t.Fatalf("expected PR #2, got %d", pr2.Number)
	}

	// Fetch PR 1 by number
	got1, err := db.GetPullRequest(ctx, repo.ID, 1)
	if err != nil {
		t.Fatalf("get pr1: %v", err)
	}
	if got1.Title != "Add launch telemetry" || got1.Draft {
		t.Fatalf("unexpected pr1: %+v", got1)
	}

	// Fetch PR 2 by ID
	got2, err := db.GetPullRequestByID(ctx, pr2.ID)
	if err != nil {
		t.Fatalf("get pr2 by id: %v", err)
	}
	if got2.Number != 2 || !got2.Draft {
		t.Fatalf("unexpected pr2: %+v", got2)
	}

	// Update PR 1
	newTitle := "Add launch telemetry (v2)"
	newDraft := true
	if err := db.UpdatePullRequest(ctx, pr1.ID, PullRequestUpdate{
		Title: &newTitle,
		Draft: &newDraft,
	}); err != nil {
		t.Fatalf("update pr1: %v", err)
	}
	updated1, err := db.GetPullRequest(ctx, repo.ID, 1)
	if err != nil {
		t.Fatalf("get updated pr1: %v", err)
	}
	if updated1.Title != newTitle || !updated1.Draft {
		t.Fatalf("pr1 not updated: %+v", updated1)
	}

	// Merge PR 1
	mergeSha := "dddddddddddddddddddddddddddddddddddddddd"
	if err := db.SetPullRequestMergeResult(ctx, pr1.ID, mergeSha); err != nil {
		t.Fatalf("merge pr1: %v", err)
	}
	merged1, err := db.GetPullRequest(ctx, repo.ID, 1)
	if err != nil {
		t.Fatalf("get merged pr1: %v", err)
	}
	if !merged1.Merged || merged1.State != "closed" || merged1.MergeCommitSHA != mergeSha {
		t.Fatalf("unexpected merged pr1: %+v", merged1)
	}

	// List PRs by state
	openPRs, _, err := db.ListPullRequests(ctx, repo.ID, "open", 10, "")
	if err != nil {
		t.Fatalf("list open prs: %v", err)
	}
	if len(openPRs) != 1 || openPRs[0].Number != 2 {
		t.Fatalf("expected 1 open PR (#2), got %+v", openPRs)
	}

	closedPRs, _, err := db.ListPullRequests(ctx, repo.ID, "closed", 10, "")
	if err != nil {
		t.Fatalf("list closed prs: %v", err)
	}
	if len(closedPRs) != 1 || closedPRs[0].Number != 1 {
		t.Fatalf("expected 1 closed PR (#1), got %+v", closedPRs)
	}
}

func TestPullRequests_AutoMerge(t *testing.T) {
	db := openMem(t)
	ctx := context.Background()
	_, repo := seedNamespaceAndRepo(t, db)

	pr := PullRequest{
		ID:      idgen.NewPR(),
		RepoID:  repo.ID,
		State:   "open",
		Title:   "Auto-merge test",
		HeadRef: "feat",
		HeadSHA: "1111111111111111111111111111111111111111",
		BaseRef: "main",
		BaseSHA: "2222222222222222222222222222222222222222",
	}
	if err := db.CreatePullRequest(ctx, &pr); err != nil {
		t.Fatalf("create pr: %v", err)
	}

	am, err := db.GetAutoMerge(ctx, pr.ID)
	if err != nil || am != nil {
		t.Fatalf("expected nil auto-merge, got: %v, err: %v", am, err)
	}

	if err := db.SetAutoMerge(ctx, PRAutoMerge{
		PullRequestID: pr.ID,
		MergeMethod:   "squash",
		CommitTitle:   "Auto squashed",
	}); err != nil {
		t.Fatalf("set auto-merge: %v", err)
	}

	am, err = db.GetAutoMerge(ctx, pr.ID)
	if err != nil || am == nil {
		t.Fatalf("expected auto-merge, got nil, err: %v", err)
	}
	if am.MergeMethod != "squash" || am.CommitTitle != "Auto squashed" {
		t.Fatalf("unexpected auto-merge: %+v", am)
	}

	if err := db.DeleteAutoMerge(ctx, pr.ID); err != nil {
		t.Fatalf("delete auto-merge: %v", err)
	}
	am, err = db.GetAutoMerge(ctx, pr.ID)
	if err != nil || am != nil {
		t.Fatalf("expected deleted auto-merge, got: %v, err: %v", am, err)
	}
}

func TestPullRequests_ThreadsCommentsReactions(t *testing.T) {
	db := openMem(t)
	ctx := context.Background()
	_, repo := seedNamespaceAndRepo(t, db)
	seedUser(t, db, "user_123", "alice@example.com")

	pr := PullRequest{
		ID:      idgen.NewPR(),
		RepoID:  repo.ID,
		State:   "open",
		Title:   "PR with comments",
		HeadRef: "feat",
		HeadSHA: "1111111111111111111111111111111111111111",
		BaseRef: "main",
		BaseSHA: "2222222222222222222222222222222222222222",
	}
	if err := db.CreatePullRequest(ctx, &pr); err != nil {
		t.Fatalf("create pr: %v", err)
	}

	threadID := idgen.NewPRThread()
	thread := PRThread{
		ID:            threadID,
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		Path:          "pkg/telemetry.go",
		Side:          "right",
		StartLine:     12,
		EndLine:       15,
		VersionNumber: 1,
	}
	if err := db.CreateThread(ctx, thread); err != nil {
		t.Fatalf("create thread: %v", err)
	}

	gotThread, err := db.GetThread(ctx, threadID)
	if err != nil {
		t.Fatalf("get thread: %v", err)
	}
	if gotThread.Path != "pkg/telemetry.go" || gotThread.StartLine != 12 {
		t.Fatalf("unexpected thread: %+v", gotThread)
	}

	// Create root comment
	c1ID := idgen.NewPRComment()
	c1 := PRComment{
		ID:            c1ID,
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		ThreadID:      threadID,
		Body:          "Could we use an atomic counter here?",
		Path:          "pkg/telemetry.go",
		Line:          15,
		Side:          "right",
	}
	if err := db.CreateComment(ctx, c1); err != nil {
		t.Fatalf("create comment 1: %v", err)
	}

	// Create reply comment
	c2ID := idgen.NewPRComment()
	c2 := PRComment{
		ID:            c2ID,
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		ThreadID:      threadID,
		Body:          "Good catch, updated!",
		InReplyTo:     c1ID,
	}
	if err := db.CreateComment(ctx, c2); err != nil {
		t.Fatalf("create comment 2: %v", err)
	}

	// List comments
	comments, err := db.ListComments(ctx, pr.ID)
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	// Reactions
	rxn := PRCommentReaction{
		ID:        idgen.NewReaction(),
		CommentID: c1ID,
		UserID:    "user_123",
		Emoji:     "👍",
	}
	if err := db.AddReaction(ctx, rxn); err != nil {
		t.Fatalf("add reaction: %v", err)
	}
	reactions, err := db.ListReactions(ctx, c1ID)
	if err != nil {
		t.Fatalf("list reactions: %v", err)
	}
	if len(reactions) != 1 || reactions[0].Emoji != "👍" {
		t.Fatalf("unexpected reactions: %+v", reactions)
	}
	if err := db.RemoveReaction(ctx, c1ID, "user_123", "👍"); err != nil {
		t.Fatalf("remove reaction: %v", err)
	}
	reactions, err = db.ListReactions(ctx, c1ID)
	if err != nil || len(reactions) != 0 {
		t.Fatalf("expected 0 reactions after remove, got %+v", reactions)
	}
}

func TestPullRequests_ReviewsAndDismissals(t *testing.T) {
	db := openMem(t)
	ctx := context.Background()
	_, repo := seedNamespaceAndRepo(t, db)
	seedUser(t, db, "user_admin", "admin@example.com")

	pr := PullRequest{
		ID:      idgen.NewPR(),
		RepoID:  repo.ID,
		State:   "open",
		Title:   "PR with review",
		HeadRef: "feat",
		HeadSHA: "1111111111111111111111111111111111111111",
		BaseRef: "main",
		BaseSHA: "2222222222222222222222222222222222222222",
	}
	if err := db.CreatePullRequest(ctx, &pr); err != nil {
		t.Fatalf("create pr: %v", err)
	}

	revID := idgen.NewPRReview()
	rev := PRReview{
		ID:            revID,
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		State:         "changes_requested",
		Body:          "Needs better tests",
		VersionNumber: 1,
	}
	if err := db.CreateReview(ctx, rev); err != nil {
		t.Fatalf("create review: %v", err)
	}

	gotRev, err := db.GetReview(ctx, revID)
	if err != nil {
		t.Fatalf("get review: %v", err)
	}
	if gotRev.State != "changes_requested" || gotRev.Body != "Needs better tests" {
		t.Fatalf("unexpected review: %+v", gotRev)
	}

	// Dismiss review
	if err := db.DismissReview(ctx, revID, "Addressed in v2", "user_admin"); err != nil {
		t.Fatalf("dismiss review: %v", err)
	}
	dismissedRev, err := db.GetReview(ctx, revID)
	if err != nil {
		t.Fatalf("get dismissed review: %v", err)
	}
	if !dismissedRev.Dismissed || dismissedRev.DismissalMessage != "Addressed in v2" {
		t.Fatalf("unexpected dismissed review: %+v", dismissedRev)
	}
}

func TestPullRequests_Labels(t *testing.T) {
	db := openMem(t)
	ctx := context.Background()
	ns, repo := seedNamespaceAndRepo(t, db)

	pr := PullRequest{
		ID:      idgen.NewPR(),
		RepoID:  repo.ID,
		State:   "open",
		Title:   "PR with labels",
		HeadRef: "feat",
		HeadSHA: "1111111111111111111111111111111111111111",
		BaseRef: "main",
		BaseSHA: "2222222222222222222222222222222222222222",
	}
	if err := db.CreatePullRequest(ctx, &pr); err != nil {
		t.Fatalf("create pr: %v", err)
	}

	l1, err := db.GetOrCreateLabel(ctx, idgen.NewLabel(), ns.ID, "bug", "d73a4a", "Bug fix")
	if err != nil {
		t.Fatalf("create label bug: %v", err)
	}
	l2, err := db.GetOrCreateLabel(ctx, idgen.NewLabel(), ns.ID, "enhancement", "a2eeef", "New feature")
	if err != nil {
		t.Fatalf("create label enhancement: %v", err)
	}

	// Set labels on PR
	if err := db.SetPRLabels(ctx, pr.ID, []string{l1.ID, l2.ID}); err != nil {
		t.Fatalf("set pr labels: %v", err)
	}
	prLabels, err := db.GetPRLabels(ctx, pr.ID)
	if err != nil {
		t.Fatalf("get pr labels: %v", err)
	}
	if len(prLabels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(prLabels))
	}

	// Remove one label
	if err := db.RemovePRLabel(ctx, pr.ID, l1.ID); err != nil {
		t.Fatalf("remove pr label: %v", err)
	}
	prLabels, err = db.GetPRLabels(ctx, pr.ID)
	if err != nil || len(prLabels) != 1 || prLabels[0].ID != l2.ID {
		t.Fatalf("expected only enhancement label, got %+v", prLabels)
	}

	// Remove all
	if err := db.RemoveAllPRLabels(ctx, pr.ID); err != nil {
		t.Fatalf("remove all: %v", err)
	}
	prLabels, err = db.GetPRLabels(ctx, pr.ID)
	if err != nil || len(prLabels) != 0 {
		t.Fatalf("expected 0 labels, got %+v", prLabels)
	}
}
