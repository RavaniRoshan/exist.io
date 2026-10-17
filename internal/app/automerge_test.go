package app

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/go-git/go-git/v6/plumbing/object"

	"exist.io/internal/engine"
	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

func setupAutoMergeTestEnv(t *testing.T) (*store.DB, *gitstore.Store, store.Repo, *store.User) {
	t.Helper()
	dataDir := t.TempDir()
	db, err := store.Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	git, err := gitstore.New(filepath.Join(dataDir, "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	ctx := context.Background()

	// 1. Create User
	u := store.User{
		ID:    idgen.NewUser(),
		Login: "developer",
		Name:  "Test Developer",
		Email: "dev@exist.io",
	}
	if err := db.CreateUserWithPassword(ctx, u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// 2. Create Namespace
	ns := store.Namespace{
		ID:   idgen.NewNS(),
		Slug: "acme",
		Type: "user",
	}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	// 3. Create Repo in DB
	repo := store.Repo{
		ID:                  idgen.NewRepo(),
		NamespaceID:         ns.ID,
		Name:                "rocket",
		FullName:            "acme/rocket",
		DefaultBranch:       "main",
		Visibility:          "public",
		AllowMergeCommit:    true,
		AllowSquashMerge:    true,
		DeleteBranchOnMerge: false,
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	// 4. Initialize git repo on disk
	if _, err := git.Init("acme", "rocket", "main"); err != nil {
		t.Fatalf("git.Init: %v", err)
	}

	return db, git, repo, &u
}

func createTestPRWithGit(t *testing.T, db *store.DB, git *gitstore.Store, repo store.Repo, u *store.User) store.PullRequest {
	t.Helper()
	ctx := context.Background()

	authorSig := object.Signature{
		Name:  u.Name,
		Email: u.Email,
		When:  time.Now().UTC(),
	}

	// 1. Initial commit on main
	c1, err := git.WriteFiles("acme", repo.Name, "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# Rocket\nInitial main commit.\n"},
	}, authorSig, "initial main commit")
	if err != nil {
		t.Fatalf("WriteFiles main: %v", err)
	}

	// 2. Create feature branch refs/heads/feature-thrust
	if _, err := git.CreateBranchRef("acme", repo.Name, "refs/heads/feature-thrust", c1.String()); err != nil {
		t.Fatalf("CreateBranchRef: %v", err)
	}

	// 3. Commit on feature branch
	c2, err := git.WriteFiles("acme", repo.Name, "feature-thrust", []gitstore.FileChange{
		{Path: "engine.go", Content: "package rocket\n\nconst Thrust = 1000\n"},
	}, authorSig, "feat: add thrust engine")
	if err != nil {
		t.Fatalf("WriteFiles feature branch: %v", err)
	}

	// 4. Create PR record in DB
	pr := store.PullRequest{
		ID:           idgen.NewPR(),
		RepoID:       repo.ID,
		Number:       1,
		State:        "open",
		Title:        "feat: add thrust engine",
		Body:         "Implements 1000kN thrust engine",
		HeadRef:      "feature-thrust",
		HeadSHA:      c2.String(),
		BaseRef:      "main",
		BaseSHA:      c1.String(),
		AuthorUserID: u.ID,
		Mergeable:    "mergeable",
	}
	if err := db.CreatePullRequest(ctx, &pr); err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}

	return pr
}

func TestAutoMergeEngine_StartStop(t *testing.T) {
	db, git, _, _ := setupAutoMergeTestEnv(t)
	engine := NewAutoMergeEngine(db, git, nil)
	engine.Start()
	// Stop should complete without hanging
	engine.Stop()
}

func TestAutoMergeEngine_Notify(t *testing.T) {
	engine := NewAutoMergeEngine(nil, nil, nil)
	evt := AutoMergeEvent{
		Type:          "check_completed",
		RepoID:        "repo_test_123",
		PullRequestID: "pr_test_456",
	}
	engine.Notify(evt)

	select {
	case received := <-engine.eventCh:
		if received.Type != evt.Type || received.RepoID != evt.RepoID || received.PullRequestID != evt.PullRequestID {
			t.Fatalf("expected event %+v, got %+v", evt, received)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event on channel")
	}
}

func TestAutoMergeEngine_IgnoresWithoutAutoMerge(t *testing.T) {
	db, git, repo, u := setupAutoMergeTestEnv(t)
	pr := createTestPRWithGit(t, db, git, repo, u)
	ctx := context.Background()

	// Add passing check run and approval review, but do NOT set auto-merge
	checkRun := store.CheckRun{
		ID:         idgen.NewCheckRun(),
		RepoID:     repo.ID,
		SHA:        pr.HeadSHA,
		Name:       "ci/test",
		Status:     "completed",
		Conclusion: "success",
	}
	if _, _, err := db.UpsertCheckRun(ctx, checkRun); err != nil {
		t.Fatalf("UpsertCheckRun: %v", err)
	}

	review := store.PRReview{
		ID:            idgen.NewPRReview(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		AuthorUserID:  u.ID,
		State:         "approved",
	}
	if err := db.CreateReview(ctx, review); err != nil {
		t.Fatalf("CreateReview: %v", err)
	}

	eng := NewAutoMergeEngine(db, git, nil)
	err := eng.Evaluate(ctx, AutoMergeEvent{
		Type:          "check_completed",
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	// PR must remain unmerged
	updatedPR, err := db.GetPullRequestByID(ctx, pr.ID)
	if err != nil {
		t.Fatalf("GetPullRequestByID: %v", err)
	}
	if updatedPR.Merged {
		t.Fatalf("expected PR to not be merged, but Merged=true")
	}
}

func TestAutoMergeEngine_MergesWhenReady(t *testing.T) {
	db, git, repo, u := setupAutoMergeTestEnv(t)
	pr := createTestPRWithGit(t, db, git, repo, u)
	ctx := context.Background()

	// 1. Enable auto-merge
	am := store.PRAutoMerge{
		PullRequestID:   pr.ID,
		EnabledByUserID: u.ID,
		MergeMethod:     "squash",
		CommitTitle:     "feat: thrust engine auto-merged",
	}
	if err := db.SetAutoMerge(ctx, am); err != nil {
		t.Fatalf("SetAutoMerge: %v", err)
	}

	// 2. Add passing check run
	checkRun := store.CheckRun{
		ID:         idgen.NewCheckRun(),
		RepoID:     repo.ID,
		SHA:        pr.HeadSHA,
		Name:       "ci/test",
		Status:     "completed",
		Conclusion: "success",
	}
	if _, _, err := db.UpsertCheckRun(ctx, checkRun); err != nil {
		t.Fatalf("UpsertCheckRun: %v", err)
	}

	// 3. Add approving review
	review := store.PRReview{
		ID:            idgen.NewPRReview(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		AuthorUserID:  u.ID,
		State:         "approved",
	}
	if err := db.CreateReview(ctx, review); err != nil {
		t.Fatalf("CreateReview: %v", err)
	}

	// 4. Start engine and notify
	eng := NewAutoMergeEngine(db, git, nil)
	eng.Start()
	defer eng.Stop()

	eng.Notify(AutoMergeEvent{
		Type:          "review_submitted",
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
	})

	// Wait for async processing
	var mergedPR store.PullRequest
	for i := 0; i < 50; i++ {
		time.Sleep(50 * time.Millisecond)
		var err error
		mergedPR, err = db.GetPullRequestByID(ctx, pr.ID)
		if err == nil && mergedPR.Merged {
			break
		}
	}

	if !mergedPR.Merged {
		t.Fatalf("expected PR to be merged by AutoMergeEngine, but Merged=false")
	}
	if mergedPR.MergeCommitSHA == "" {
		t.Fatalf("expected non-empty MergeCommitSHA")
	}

	// Auto-merge config must be cleaned up
	remainingAM, err := db.GetAutoMerge(ctx, pr.ID)
	if err != nil {
		t.Fatalf("GetAutoMerge: %v", err)
	}
	if remainingAM != nil {
		t.Fatalf("expected auto-merge config to be deleted, got %+v", remainingAM)
	}
}

func TestAutoMergeEngine_BlockedByPendingChecks(t *testing.T) {
	db, git, repo, u := setupAutoMergeTestEnv(t)
	pr := createTestPRWithGit(t, db, git, repo, u)
	ctx := context.Background()

	// 1. Enable auto-merge
	am := store.PRAutoMerge{
		PullRequestID:   pr.ID,
		EnabledByUserID: u.ID,
		MergeMethod:     "squash",
	}
	if err := db.SetAutoMerge(ctx, am); err != nil {
		t.Fatalf("SetAutoMerge: %v", err)
	}

	// 2. Add an in_progress check run
	checkRun := store.CheckRun{
		ID:     idgen.NewCheckRun(),
		RepoID: repo.ID,
		SHA:    pr.HeadSHA,
		Name:   "ci/test",
		Status: "in_progress",
	}
	if _, _, err := db.UpsertCheckRun(ctx, checkRun); err != nil {
		t.Fatalf("UpsertCheckRun: %v", err)
	}

	// 3. Add approving review
	review := store.PRReview{
		ID:            idgen.NewPRReview(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		AuthorUserID:  u.ID,
		State:         "approved",
	}
	if err := db.CreateReview(ctx, review); err != nil {
		t.Fatalf("CreateReview: %v", err)
	}

	eng := NewAutoMergeEngine(db, git, nil)
	if err := eng.Evaluate(ctx, AutoMergeEvent{
		Type:          "check_completed",
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
	}); err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}

	// PR must NOT be merged
	updatedPR, err := db.GetPullRequestByID(ctx, pr.ID)
	if err != nil {
		t.Fatalf("GetPullRequestByID: %v", err)
	}
	if updatedPR.Merged {
		t.Fatalf("expected PR to remain unmerged while check is in progress")
	}
}

func TestAutoMergeEngine_BlockedByChangesRequested(t *testing.T) {
	db, git, repo, u := setupAutoMergeTestEnv(t)
	pr := createTestPRWithGit(t, db, git, repo, u)
	ctx := context.Background()

	// 1. Enable auto-merge
	am := store.PRAutoMerge{
		PullRequestID:   pr.ID,
		EnabledByUserID: u.ID,
		MergeMethod:     "squash",
	}
	if err := db.SetAutoMerge(ctx, am); err != nil {
		t.Fatalf("SetAutoMerge: %v", err)
	}

	// 2. Add passing check run
	checkRun := store.CheckRun{
		ID:         idgen.NewCheckRun(),
		RepoID:     repo.ID,
		SHA:        pr.HeadSHA,
		Name:       "ci/test",
		Status:     "completed",
		Conclusion: "success",
	}
	if _, _, err := db.UpsertCheckRun(ctx, checkRun); err != nil {
		t.Fatalf("UpsertCheckRun: %v", err)
	}

	// 3. Add changes_requested review
	review := store.PRReview{
		ID:            idgen.NewPRReview(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		AuthorUserID:  u.ID,
		State:         "changes_requested",
	}
	if err := db.CreateReview(ctx, review); err != nil {
		t.Fatalf("CreateReview: %v", err)
	}

	eng := NewAutoMergeEngine(db, git, nil)
	if err := eng.Evaluate(ctx, AutoMergeEvent{
		Type:          "review_submitted",
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
	}); err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}

	// PR must NOT be merged
	updatedPR, err := db.GetPullRequestByID(ctx, pr.ID)
	if err != nil {
		t.Fatalf("GetPullRequestByID: %v", err)
	}
	if updatedPR.Merged {
		t.Fatalf("expected PR to remain unmerged when changes are requested")
	}
}

func TestAutoMergeEngine_BlockedByRequiredApprovals(t *testing.T) {
	db, git, repo, u := setupAutoMergeTestEnv(t)
	pr := createTestPRWithGit(t, db, git, repo, u)
	ctx := context.Background()

	// 1. Enable auto-merge
	am := store.PRAutoMerge{
		PullRequestID:   pr.ID,
		EnabledByUserID: u.ID,
		MergeMethod:     "squash",
	}
	if err := db.SetAutoMerge(ctx, am); err != nil {
		t.Fatalf("SetAutoMerge: %v", err)
	}

	// 2. Add passing check run
	checkRun := store.CheckRun{
		ID:         idgen.NewCheckRun(),
		RepoID:     repo.ID,
		SHA:        pr.HeadSHA,
		Name:       "ci/test",
		Status:     "completed",
		Conclusion: "success",
	}
	if _, _, err := db.UpsertCheckRun(ctx, checkRun); err != nil {
		t.Fatalf("UpsertCheckRun: %v", err)
	}

	// 3. Add ruleset requiring 2 approvals
	rs := store.Ruleset{
		ID:               idgen.NewRuleset(),
		RepoID:           repo.ID,
		Name:             "two-approvals",
		Enforcement:      "active",
		Kind:             "merge_branch",
		IncludedRefNames: []string{"refs/heads/main"},
		Rules: []store.RulesetRule{
			{
				ID:       idgen.NewRulesetRule(),
				RuleType: "pull_request",
				Parameters: map[string]any{
					"requiredApprovingReviewCount": 2,
				},
			},
		},
	}
	if _, err := db.CreateRuleset(ctx, rs); err != nil {
		t.Fatalf("CreateRuleset: %v", err)
	}

	// 4. Create ruleset engine
	rulesEng := engine.New(db)
	eng := NewAutoMergeEngine(db, git, rulesEng)

	// 5. Add 1 approval review (second user)
	u2 := store.User{
		ID:    idgen.NewUser(),
		Login: "reviewer2",
		Email: "reviewer2@exist.io",
	}
	if err := db.CreateUserWithPassword(ctx, u2); err != nil {
		t.Fatalf("CreateUser u2: %v", err)
	}

	review1 := store.PRReview{
		ID:            idgen.NewPRReview(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		AuthorUserID:  u2.ID,
		State:         "approved",
	}
	if err := db.CreateReview(ctx, review1); err != nil {
		t.Fatalf("CreateReview 1: %v", err)
	}

	// Evaluate with only 1 approval
	if err := eng.Evaluate(ctx, AutoMergeEvent{
		Type:          "review_submitted",
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
	}); err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}

	// PR must NOT be merged
	updatedPR, err := db.GetPullRequestByID(ctx, pr.ID)
	if err != nil {
		t.Fatalf("GetPullRequestByID: %v", err)
	}
	if updatedPR.Merged {
		t.Fatalf("expected PR to not be merged with 1 of 2 required approvals")
	}

	// 6. Add 2nd approval review (third user)
	u3 := store.User{
		ID:    idgen.NewUser(),
		Login: "reviewer3",
		Email: "reviewer3@exist.io",
	}
	if err := db.CreateUserWithPassword(ctx, u3); err != nil {
		t.Fatalf("CreateUser u3: %v", err)
	}

	review2 := store.PRReview{
		ID:            idgen.NewPRReview(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		AuthorUserID:  u3.ID,
		State:         "approved",
	}
	if err := db.CreateReview(ctx, review2); err != nil {
		t.Fatalf("CreateReview 2: %v", err)
	}

	// Evaluate with 2 approvals
	if err := eng.Evaluate(ctx, AutoMergeEvent{
		Type:          "review_submitted",
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
	}); err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}

	// PR must now be merged
	updatedPR, err = db.GetPullRequestByID(ctx, pr.ID)
	if err != nil {
		t.Fatalf("GetPullRequestByID: %v", err)
	}
	if !updatedPR.Merged {
		t.Fatalf("expected PR to be merged once 2 approvals were met")
	}
}

func TestAutoMergeEngine_ConcurrentEvents(t *testing.T) {
	db, git, repo, u := setupAutoMergeTestEnv(t)
	pr := createTestPRWithGit(t, db, git, repo, u)
	ctx := context.Background()

	// 1. Enable auto-merge
	am := store.PRAutoMerge{
		PullRequestID:   pr.ID,
		EnabledByUserID: u.ID,
		MergeMethod:     "squash",
	}
	if err := db.SetAutoMerge(ctx, am); err != nil {
		t.Fatalf("SetAutoMerge: %v", err)
	}

	// 2. Add passing check run
	checkRun := store.CheckRun{
		ID:         idgen.NewCheckRun(),
		RepoID:     repo.ID,
		SHA:        pr.HeadSHA,
		Name:       "ci/test",
		Status:     "completed",
		Conclusion: "success",
	}
	if _, _, err := db.UpsertCheckRun(ctx, checkRun); err != nil {
		t.Fatalf("UpsertCheckRun: %v", err)
	}

	// 3. Add approving review
	review := store.PRReview{
		ID:            idgen.NewPRReview(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		AuthorUserID:  u.ID,
		State:         "approved",
	}
	if err := db.CreateReview(ctx, review); err != nil {
		t.Fatalf("CreateReview: %v", err)
	}

	// 4. Start engine
	eng := NewAutoMergeEngine(db, git, nil)
	eng.Start()
	defer eng.Stop()

	// Fire 10 concurrent notifications
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			evtType := "check_completed"
			if idx%2 == 0 {
				evtType = "review_submitted"
			}
			eng.Notify(AutoMergeEvent{
				Type:          evtType,
				RepoID:        repo.ID,
				PullRequestID: pr.ID,
			})
		}(i)
	}
	wg.Wait()

	// Wait for engine to finish processing
	var mergedPR store.PullRequest
	for i := 0; i < 50; i++ {
		time.Sleep(50 * time.Millisecond)
		var err error
		mergedPR, err = db.GetPullRequestByID(ctx, pr.ID)
		if err == nil && mergedPR.Merged {
			break
		}
	}

	if !mergedPR.Merged {
		t.Fatalf("expected PR to be merged after concurrent notifications")
	}
}
