// Package engine provides policy evaluation, branch/tag protection ruleset checking,
// and mergeability gate validation for exist.io.
package engine

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"exist.io/internal/store"
)

// MergeBlocker describes a condition preventing a pull request from being merged.
type MergeBlocker struct {
	Kind              string             `json:"kind"`
	Message           string             `json:"message"`
	RequiredApprovals *RequiredApprovals `json:"requiredApprovals,omitempty"`
	RequiredChecks    *RequiredChecks    `json:"requiredChecks,omitempty"`
}

// RequiredApprovals summarizes approval count requirements and current state.
type RequiredApprovals struct {
	RequiredCount int `json:"requiredCount"`
	ApprovedCount int `json:"approvedCount"`
}

// RequiredChecks summarizes required check runs and overall state.
type RequiredChecks struct {
	State  string         `json:"state"`
	Checks []CheckDetails `json:"checks"`
}

// CheckDetails holds details for an individual status check run.
type CheckDetails struct {
	Name     string         `json:"name"`
	CheckRun map[string]any `json:"checkRun,omitempty"`
}

// EvaluationResult contains the outcome of evaluating rulesets against a pull request or ref.
type EvaluationResult struct {
	Allowed  bool
	Verdict  string // "mergeable" | "blocked"
	Blockers []MergeBlocker
}

// EvaluateMerge evaluates merge protection rulesets for a target pull request.
func EvaluateMerge(
	ctx context.Context,
	rulesets []store.Ruleset,
	defaultBranch string,
	baseRef string,
	actingUserID string,
	actingAppID string,
	approvedReviewCount int,
	checkRuns []store.CheckRun,
) EvaluationResult {
	var blockers []MergeBlocker

	for _, rs := range rulesets {
		if rs.Enforcement != "active" {
			continue
		}
		if rs.Kind != "merge_branch" {
			continue
		}

		if !refApplies(baseRef, rs.IncludedRefNames, rs.ExcludedRefNames, defaultBranch) {
			continue
		}

		if canBypass(rs.BypassActors, actingUserID, actingAppID, true) {
			continue
		}

		for _, rule := range rs.Rules {
			switch rule.RuleType {
			case "pull_request":
				reqApprovals := 1
				if rawCount, ok := rule.Parameters["requiredApprovingReviewCount"]; ok {
					switch v := rawCount.(type) {
					case float64:
						reqApprovals = int(v)
					case int:
						reqApprovals = v
					}
				}
				if approvedReviewCount < reqApprovals {
					blockers = append(blockers, MergeBlocker{
						Kind:    "required_approvals",
						Message: fmt.Sprintf("Approving review count is %d; %d required. Request reviews and wait for the required approvals.", approvedReviewCount, reqApprovals),
						RequiredApprovals: &RequiredApprovals{
							RequiredCount: reqApprovals,
							ApprovedCount: approvedReviewCount,
						},
					})
				}

			case "require_status_checks":
				rawChecks, ok := rule.Parameters["requiredChecks"]
				if !ok {
					continue
				}
				checkList, ok := rawChecks.([]any)
				if !ok || len(checkList) == 0 {
					continue
				}

				var missing []CheckDetails
				for _, rc := range checkList {
					m, ok := rc.(map[string]any)
					if !ok {
						continue
					}
					checkName, _ := m["name"].(string)
					if checkName == "" {
						checkName, _ = m["context"].(string)
					}
					groupKey, _ := m["groupKey"].(string)
					runKey, _ := m["runKey"].(string)

					passed := false
					var matchedRun *store.CheckRun
					for _, cr := range checkRuns {
						// Match on runKey or groupKey or name
						keyMatch := false
						if runKey != "" && cr.Key == runKey {
							keyMatch = true
						} else if runKey == "" && groupKey != "" && cr.Key == groupKey {
							keyMatch = true
						} else if checkName != "" && cr.Name == checkName {
							keyMatch = true
						}

						if keyMatch {
							matchedRun = &cr
							if cr.Status == "completed" && (cr.Conclusion == "success" || cr.Conclusion == "neutral" || cr.Conclusion == "skipped") {
								passed = true
							}
							break
						}
					}

					if !passed {
						name := checkName
						if name == "" {
							name = runKey
						}
						if name == "" {
							name = groupKey
						}
						cd := CheckDetails{Name: name}
						if matchedRun != nil {
							cd.CheckRun = map[string]any{
								"id":         matchedRun.ID,
								"name":       matchedRun.Name,
								"status":     matchedRun.Status,
								"conclusion": matchedRun.Conclusion,
							}
						}
						missing = append(missing, cd)
					}
				}

				if len(missing) > 0 {
					blockers = append(blockers, MergeBlocker{
						Kind:    "required_checks",
						Message: "Required status checks are pending or failing. Wait for checks to finish or fix the failing checks.",
						RequiredChecks: &RequiredChecks{
							State:  "pending",
							Checks: missing,
						},
					})
				}
			}
		}
	}

	if len(blockers) > 0 {
		return EvaluationResult{
			Allowed:  false,
			Verdict:  "blocked",
			Blockers: blockers,
		}
	}

	return EvaluationResult{
		Allowed: true,
		Verdict: "mergeable",
	}
}

func refApplies(refName string, included, excluded []string, defaultBranch string) bool {
	// Canonicalize
	canon := refName
	if !strings.HasPrefix(canon, "refs/") {
		canon = "refs/heads/" + canon
	}

	incMatch := false
	for _, inc := range included {
		if refMatches(canon, inc, defaultBranch) {
			incMatch = true
			break
		}
	}
	if !incMatch && len(included) > 0 {
		return false
	}

	for _, exc := range excluded {
		if refMatches(canon, exc, defaultBranch) {
			return false
		}
	}

	return true
}

func refMatches(canonRef, pattern, defaultBranch string) bool {
	if pattern == "~ALL" {
		return true
	}
	if pattern == "~DEFAULT_BRANCH" {
		return canonRef == "refs/heads/"+defaultBranch
	}

	patCanon := pattern
	if !strings.HasPrefix(patCanon, "refs/") {
		patCanon = "refs/heads/" + patCanon
	}

	if patCanon == canonRef {
		return true
	}

	matched, err := filepath.Match(patCanon, canonRef)
	if err == nil && matched {
		return true
	}
	return false
}

func canBypass(actors []store.RulesetBypassActor, actingUserID, actingAppID string, isPR bool) bool {
	for _, a := range actors {
		if a.BypassMode == "always" || (isPR && a.BypassMode == "pull_request_only") {
			if a.User != nil && a.User.ID != "" && a.User.ID == actingUserID {
				return true
			}
			if a.App != nil && a.App.ID != "" && a.App.ID == actingAppID {
				return true
			}
		}
	}
	return false
}

// Engine evaluates rulesets and policies.
type Engine struct {
	db *store.DB
}

// New creates a new ruleset Engine.
func New(db *store.DB) *Engine {
	return &Engine{db: db}
}

// RequiredReviewCount returns the required approving review count for a target ref based on rulesets.
func (e *Engine) RequiredReviewCount(ctx context.Context, repoID, baseRef string) int {
	if e == nil || e.db == nil {
		return 0
	}
	repo, err := e.db.GetRepoByID(ctx, repoID)
	if err != nil {
		return 0
	}
	rulesets, err := e.db.ListRulesets(ctx, repoID)
	if err != nil {
		return 0
	}
	maxReq := 0
	for _, rs := range rulesets {
		if rs.Enforcement != "active" || rs.Kind != "merge_branch" {
			continue
		}
		if !refApplies(baseRef, rs.IncludedRefNames, rs.ExcludedRefNames, repo.DefaultBranch) {
			continue
		}
		for _, rule := range rs.Rules {
			if rule.RuleType == "pull_request" {
				reqApprovals := 1
				if rawCount, ok := rule.Parameters["requiredApprovingReviewCount"]; ok {
					switch v := rawCount.(type) {
					case float64:
						reqApprovals = int(v)
					case int:
						reqApprovals = v
					}
				}
				if reqApprovals > maxReq {
					maxReq = reqApprovals
				}
			}
		}
	}
	return maxReq
}

// EvaluateMerge evaluates merge protection rulesets for a target pull request.
func (e *Engine) EvaluateMerge(
	ctx context.Context,
	rulesets []store.Ruleset,
	defaultBranch string,
	baseRef string,
	actingUserID string,
	actingAppID string,
	approvedReviewCount int,
	checkRuns []store.CheckRun,
) EvaluationResult {
	return EvaluateMerge(ctx, rulesets, defaultBranch, baseRef, actingUserID, actingAppID, approvedReviewCount, checkRuns)
}
