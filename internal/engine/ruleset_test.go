package engine_test

import (
	"context"
	"testing"

	"exist.io/internal/engine"
	"exist.io/internal/store"
)

func TestEvaluateMerge_Ruleset(t *testing.T) {
	ctx := context.Background()

	// 1. No rulesets -> allowed
	res := engine.EvaluateMerge(ctx, nil, "main", "main", "user_1", "", 0, nil)
	if !res.Allowed || res.Verdict != "mergeable" {
		t.Fatalf("expected allowed, got %+v", res)
	}

	// 2. Pull Request rule requiring 2 approvals
	rsPR := store.Ruleset{
		Enforcement:      "active",
		Kind:             "merge_branch",
		IncludedRefNames: []string{"~DEFAULT_BRANCH"},
		Rules: []store.RulesetRule{
			{
				RuleType: "pull_request",
				Parameters: map[string]any{
					"requiredApprovingReviewCount": float64(2),
				},
			},
		},
	}

	// 0 approvals -> blocked
	res = engine.EvaluateMerge(ctx, []store.Ruleset{rsPR}, "main", "main", "user_1", "", 0, nil)
	if res.Allowed || res.Verdict != "blocked" || len(res.Blockers) != 1 {
		t.Fatalf("expected blocked with 1 blocker, got %+v", res)
	}
	if res.Blockers[0].Kind != "required_approvals" {
		t.Fatalf("expected kind required_approvals, got %s", res.Blockers[0].Kind)
	}

	// 2 approvals -> allowed
	res = engine.EvaluateMerge(ctx, []store.Ruleset{rsPR}, "main", "main", "user_1", "", 2, nil)
	if !res.Allowed || res.Verdict != "mergeable" {
		t.Fatalf("expected allowed with 2 approvals, got %+v", res)
	}

	// 3. Status checks rule
	rsChecks := store.Ruleset{
		Enforcement:      "active",
		Kind:             "merge_branch",
		IncludedRefNames: []string{"refs/heads/main"},
		Rules: []store.RulesetRule{
			{
				RuleType: "require_status_checks",
				Parameters: map[string]any{
					"requiredChecks": []any{
						map[string]any{
							"groupKey": "ci",
							"runKey":   "build",
							"name":     "CI Build",
						},
					},
				},
			},
		},
	}

	// Missing check run -> blocked
	res = engine.EvaluateMerge(ctx, []store.Ruleset{rsChecks}, "main", "main", "user_1", "", 0, nil)
	if res.Allowed || len(res.Blockers) != 1 || res.Blockers[0].Kind != "required_checks" {
		t.Fatalf("expected blocked on required_checks, got %+v", res)
	}

	// Passing check run -> allowed
	passingRun := store.CheckRun{
		Key:        "build",
		Name:       "CI Build",
		Status:     "completed",
		Conclusion: "success",
	}
	res = engine.EvaluateMerge(ctx, []store.Ruleset{rsChecks}, "main", "main", "user_1", "", 0, []store.CheckRun{passingRun})
	if !res.Allowed || res.Verdict != "mergeable" {
		t.Fatalf("expected allowed with passing check run, got %+v", res)
	}

	// 4. Bypass actor
	rsWithBypass := rsPR
	rsWithBypass.BypassActors = []store.RulesetBypassActor{
		{
			BypassMode: "always",
			User:       &store.BypassUser{ID: "admin_user"},
		},
	}

	res = engine.EvaluateMerge(ctx, []store.Ruleset{rsWithBypass}, "main", "main", "admin_user", "", 0, nil)
	if !res.Allowed || res.Verdict != "mergeable" {
		t.Fatalf("expected bypass actor to be allowed, got %+v", res)
	}
}
