package origin

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestRulesets_HandlersAndMergeGate(t *testing.T) {
	srv, db, h := testServer(t)
	seedRepoWithBranches(t, srv, db, "acme", "rocket")

	// 1. ListRulesets (initially empty)
	res := do(h, "GET", "/v1/origin/repos/acme/rocket/rulesets", "")
	if res.Code != http.StatusOK {
		t.Fatalf("listRulesets expected 200, got %d", res.Code)
	}
	var listResp struct {
		Rulesets []map[string]any `json:"rulesets"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal listRulesets: %v", err)
	}
	if len(listResp.Rulesets) != 0 {
		t.Fatalf("expected 0 rulesets, got %d", len(listResp.Rulesets))
	}

	// 2. CreateRuleset with required reviews rule
	createBody := `{
		"name": "require-review",
		"description": "Require 1 review before merging to main.",
		"enforcement": "active",
		"kind": "merge_branch",
		"includedRefNames": ["refs/heads/main", "~DEFAULT_BRANCH"],
		"rules": [
			{
				"ruleType": "pull_request",
				"parameters": {
					"requiredApprovingReviewCount": 1
				}
			}
		],
		"bypassActors": [
			{
				"bypassMode": "always",
				"user": {
					"id": "admin_bypass_user"
				}
			}
		]
	}`
	res = do(h, "POST", "/v1/origin/repos/acme/rocket/rulesets", createBody)
	if res.Code != http.StatusOK {
		t.Fatalf("createRuleset expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var createdRS map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &createdRS); err != nil {
		t.Fatalf("unmarshal createRuleset: %v", err)
	}
	rsID, _ := createdRS["id"].(string)
	if rsID == "" || createdRS["name"] != "require-review" {
		t.Fatalf("unexpected created ruleset: %+v", createdRS)
	}

	// 3. GetRuleset
	res = do(h, "GET", "/v1/origin/repos/acme/rocket/rulesets/"+rsID, "")
	if res.Code != http.StatusOK {
		t.Fatalf("getRuleset expected 200, got %d", res.Code)
	}

	// 4. UpdateRuleset
	updateBody := `{
		"name": "require-review",
		"description": "Updated description",
		"enforcement": "active",
		"kind": "merge_branch",
		"includedRefNames": ["refs/heads/main", "~DEFAULT_BRANCH"],
		"rules": [
			{
				"ruleType": "pull_request",
				"parameters": {
					"requiredApprovingReviewCount": 1
				}
			}
		]
	}`
	res = do(h, "PUT", "/v1/origin/repos/acme/rocket/rulesets/"+rsID, updateBody)
	if res.Code != http.StatusOK {
		t.Fatalf("updateRuleset expected 200, got %d: %s", res.Code, res.Body.String())
	}

	// 5. Create Pull Request and verify ruleset blocks merge until approved
	prBody := `{"title":"Add telemetry","body":"Enables telemetry","head":"feature-1","base":"main","draft":false}`
	res = do(h, "POST", "/v1/origin/repos/acme/rocket/pulls", prBody)
	if res.Code != http.StatusOK {
		t.Fatalf("create PR expected 200, got %d: %s", res.Code, res.Body.String())
	}

	// Check mergeability: should report blocked by required approvals
	res = do(h, "GET", "/v1/origin/repos/acme/rocket/pulls/1/mergeability", "")
	if res.Code != http.StatusOK {
		t.Fatalf("getMergeability expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var mergeabilityResp struct {
		Verdict  string           `json:"verdict"`
		Blockers []map[string]any `json:"blockers"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &mergeabilityResp); err != nil {
		t.Fatalf("unmarshal mergeability: %v", err)
	}
	if mergeabilityResp.Verdict != "blocked" || len(mergeabilityResp.Blockers) == 0 {
		t.Fatalf("expected blocked verdict with blockers, got %+v", mergeabilityResp)
	}

	// Attempt merge: should fail with 409 Conflict due to ruleset
	mergeBody := `{"mergeMethod":"merge"}`
	res = do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/1/merge", mergeBody)
	if res.Code != http.StatusConflict {
		t.Fatalf("merge expected 409 Conflict when blocked by ruleset, got %d: %s", res.Code, res.Body.String())
	}

	// Add an approving review
	reviewBody := `{"verdict":"approve","body":"LGTM!"}`
	res = do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/1/reviews", reviewBody)
	if res.Code != http.StatusOK {
		t.Fatalf("create review expected 200, got %d: %s", res.Code, res.Body.String())
	}

	// Now merge should succeed!
	res = do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/1/merge", mergeBody)
	if res.Code != http.StatusOK {
		t.Fatalf("merge expected 200 after approval, got %d: %s", res.Code, res.Body.String())
	}
	var mergeResp struct {
		MergeCommitSHA string `json:"mergeCommitSha"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &mergeResp); err != nil {
		t.Fatalf("unmarshal merge response: %v", err)
	}
	if mergeResp.MergeCommitSHA == "" {
		t.Fatalf("expected merge commit SHA to be populated")
	}

	// 6. DeleteRuleset
	res = do(h, "DELETE", "/v1/origin/repos/acme/rocket/rulesets/"+rsID, "")
	if res.Code != http.StatusNoContent {
		t.Fatalf("deleteRuleset expected 204, got %d", res.Code)
	}

	// Verify 404 after delete
	res = do(h, "GET", "/v1/origin/repos/acme/rocket/rulesets/"+rsID, "")
	if res.Code != http.StatusNotFound {
		t.Fatalf("getRuleset after delete expected 404, got %d", res.Code)
	}
}
