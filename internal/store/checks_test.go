package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

func TestChecks_SuitesAndRuns(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// Seed user for foreign keys
	_, err = db.UpsertUserByEmail(ctx, "test@acme.dev", func() string { return "user_test" })
	if err != nil {
		t.Fatalf("UpsertUserByEmail: %v", err)
	}

	// Seed namespace and repo
	ns := store.Namespace{ID: idgen.NewNS(), Slug: "acme", Type: "team", Name: "Acme"}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	repo := store.Repo{
		ID:          idgen.NewRepo(),
		NamespaceID: ns.ID,
		Name:        "rocket",
		FullName:    "acme/rocket",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	headSHA := "9a41f0c3d2b8e7f6a5c4d3e2f1b0a9c8d7e6f5a4"

	// 1. Upsert Check Suite
	suiteInput := store.CheckSuite{
		RepoID:     repo.ID,
		SHA:        headSHA,
		Key:        "ci-8842",
		Name:       "CI",
		DetailsURL: "https://ci.acme.dev/runs/8842",
		ExternalID: "build-8842",
	}
	suite, err := db.UpsertCheckSuite(ctx, suiteInput)
	if err != nil {
		t.Fatalf("UpsertCheckSuite: %v", err)
	}
	if suite.ID == "" {
		t.Fatalf("expected suite ID to be populated")
	}

	// Re-upsert suite with updated details
	suiteInput.Name = "CI Suite v2"
	updatedSuite, err := db.UpsertCheckSuite(ctx, suiteInput)
	if err != nil {
		t.Fatalf("re-upsert CheckSuite: %v", err)
	}
	if updatedSuite.ID != suite.ID {
		t.Fatalf("expected same suite ID %s, got %s", suite.ID, updatedSuite.ID)
	}
	if updatedSuite.Name != "CI Suite v2" {
		t.Fatalf("expected updated suite name, got %s", updatedSuite.Name)
	}

	// GetCheckSuite
	gotSuite, err := db.GetCheckSuite(ctx, suite.ID)
	if err != nil {
		t.Fatalf("GetCheckSuite: %v", err)
	}
	if gotSuite.Name != "CI Suite v2" {
		t.Fatalf("unexpected suite name: %s", gotSuite.Name)
	}

	// ListCheckSuitesForCommit
	suites, err := db.ListCheckSuitesForCommit(ctx, repo.ID, headSHA, 10, 0)
	if err != nil {
		t.Fatalf("ListCheckSuitesForCommit: %v", err)
	}
	if len(suites) != 1 {
		t.Fatalf("expected 1 suite, got %d", len(suites))
	}

	// 2. Check Run Creation
	now := time.Now().UTC()
	extUpdate1 := now.Format(time.RFC3339)
	runInput := store.CheckRun{
		RepoID:            repo.ID,
		CheckSuiteID:      suite.ID,
		SHA:               headSHA,
		Key:               "ci-8842-unit",
		Name:              "unit-tests",
		Status:            "in_progress",
		ExternalID:        "run-8842-1",
		ExternalUpdatedAt: extUpdate1,
		OutputTitle:       "Unit tests",
		OutputSummary:     "Running tests...",
		IsRerequestable:   true,
	}

	run, outcome, err := db.UpsertCheckRun(ctx, runInput)
	if err != nil {
		t.Fatalf("UpsertCheckRun: %v", err)
	}
	if outcome != "created" {
		t.Fatalf("expected outcome 'created', got %s", outcome)
	}
	if run.ID == "" {
		t.Fatalf("expected non-empty run ID")
	}

	// Unchanged test
	_, outcomeUnchanged, err := db.UpsertCheckRun(ctx, runInput)
	if err != nil {
		t.Fatalf("UpsertCheckRun (unchanged): %v", err)
	}
	if outcomeUnchanged != "unchanged" {
		t.Fatalf("expected outcome 'unchanged', got %s", outcomeUnchanged)
	}

	// Stale test (older external_updated_at)
	olderInput := runInput
	olderInput.ExternalUpdatedAt = now.Add(-10 * time.Minute).Format(time.RFC3339)
	olderInput.OutputSummary = "Should be ignored"
	_, outcomeStale, err := db.UpsertCheckRun(ctx, olderInput)
	if err != nil {
		t.Fatalf("UpsertCheckRun (stale): %v", err)
	}
	if outcomeStale != "ignored_stale" {
		t.Fatalf("expected outcome 'ignored_stale', got %s", outcomeStale)
	}

	// Valid update (newer external_updated_at)
	newerInput := runInput
	extUpdate2 := now.Add(5 * time.Minute).Format(time.RFC3339)
	newerInput.ExternalUpdatedAt = extUpdate2
	newerInput.Status = "completed"
	newerInput.Conclusion = "success"
	newerInput.OutputSummary = "All 128 tests passed."
	runUpdated, outcomeUpdated, err := db.UpsertCheckRun(ctx, newerInput)
	if err != nil {
		t.Fatalf("UpsertCheckRun (update): %v", err)
	}
	if outcomeUpdated != "updated" {
		t.Fatalf("expected outcome 'updated', got %s", outcomeUpdated)
	}
	if runUpdated.Status != "completed" || runUpdated.Conclusion != "success" {
		t.Fatalf("expected completed/success, got %s/%s", runUpdated.Status, runUpdated.Conclusion)
	}

	// 3. Rerequest
	rerequested, err := db.RerequestCheckRun(ctx, run.ID, "user_test")
	if err != nil {
		t.Fatalf("RerequestCheckRun: %v", err)
	}
	if rerequested.Status != "rerequested" || rerequested.RerequestedAt == "" {
		t.Fatalf("expected status rerequested with timestamp, got %s/%s", rerequested.Status, rerequested.RerequestedAt)
	}

	// Repeated rerequest fails with ErrAlreadyExists
	_, err = db.RerequestCheckRun(ctx, run.ID, "user_test")
	if !errors.Is(err, store.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	// 4. Annotations
	annotationsToCreate := []store.CheckRunAnnotation{
		{
			Path:            "src/app.go",
			StartLine:       10,
			EndLine:         10,
			StartColumn:     1,
			EndColumn:       15,
			AnnotationLevel: "warning",
			Message:         "Unchecked error return",
			Title:           "Lint Warning",
		},
	}
	createdAnns, err := db.CreateCheckRunAnnotations(ctx, run.ID, annotationsToCreate)
	if err != nil {
		t.Fatalf("CreateCheckRunAnnotations: %v", err)
	}
	if len(createdAnns) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(createdAnns))
	}
	if createdAnns[0].ID == "" {
		t.Fatalf("expected annotation ID")
	}

	listedAnns, err := db.ListCheckRunAnnotations(ctx, run.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListCheckRunAnnotations: %v", err)
	}
	if len(listedAnns) != 1 {
		t.Fatalf("expected 1 listed annotation, got %d", len(listedAnns))
	}
	if listedAnns[0].Message != "Unchecked error return" {
		t.Fatalf("unexpected message: %s", listedAnns[0].Message)
	}
}
