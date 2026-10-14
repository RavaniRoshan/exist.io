package store_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

func TestRulesets_CRUD(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

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

	// 1. Create Ruleset
	rsInput := store.Ruleset{
		RepoID:           repo.ID,
		Name:             "require-review",
		Description:      "Require an approving review before merging to main.",
		Enforcement:      "active",
		Kind:             "merge_branch",
		IncludedRefNames: []string{"refs/heads/main", "~DEFAULT_BRANCH"},
		ExcludedRefNames: []string{"refs/heads/hotfix/*"},
		Rules: []store.RulesetRule{
			{
				RuleType: "pull_request",
				Parameters: map[string]any{
					"requiredApprovingReviewCount": float64(1),
				},
			},
		},
		BypassActors: []store.RulesetBypassActor{
			{
				BypassMode: "always",
				User:       &store.BypassUser{ID: "user_01k2ja2000e0080000000000c3"},
			},
		},
	}

	created, err := db.CreateRuleset(ctx, rsInput)
	if err != nil {
		t.Fatalf("CreateRuleset: %v", err)
	}
	if !strings.HasPrefix(created.ID, "rs_") {
		t.Fatalf("expected rs_ prefix, got %s", created.ID)
	}
	if len(created.Rules) != 1 || !strings.HasPrefix(created.Rules[0].ID, "rsr_") {
		t.Fatalf("expected rsr_ prefix on rule, got %+v", created.Rules)
	}
	if len(created.BypassActors) != 1 || !strings.HasPrefix(created.BypassActors[0].ID, "rsba_") {
		t.Fatalf("expected rsba_ prefix on bypass actor, got %+v", created.BypassActors)
	}

	// 2. Get Ruleset
	got, err := db.GetRuleset(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetRuleset: %v", err)
	}
	if got.Name != "require-review" || got.Enforcement != "active" {
		t.Fatalf("unexpected got ruleset: %+v", got)
	}
	if len(got.IncludedRefNames) != 2 || len(got.ExcludedRefNames) != 1 {
		t.Fatalf("unexpected ref names: inc=%+v, exc=%+v", got.IncludedRefNames, got.ExcludedRefNames)
	}

	// 3. List Rulesets
	list, err := db.ListRulesets(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListRulesets: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 ruleset, got %d", len(list))
	}

	// 4. Update Ruleset
	got.Description = "Updated description"
	got.Enforcement = "evaluate"
	updated, err := db.UpdateRuleset(ctx, got)
	if err != nil {
		t.Fatalf("UpdateRuleset: %v", err)
	}
	if updated.Description != "Updated description" || updated.Enforcement != "evaluate" {
		t.Fatalf("unexpected updated ruleset: %+v", updated)
	}

	// 5. Delete Ruleset
	if err := db.DeleteRuleset(ctx, repo.ID, created.ID); err != nil {
		t.Fatalf("DeleteRuleset: %v", err)
	}

	// 6. Verify Delete
	_, err = db.GetRuleset(ctx, created.ID)
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
