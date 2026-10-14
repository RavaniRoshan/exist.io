package origin

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestPRSocial_CommentsAndThreads(t *testing.T) {
	srv, db, h := testServer(t)
	seedRepoWithBranches(t, srv, db, "acme", "rocket")

	// Create PR
	do(h, "POST", "/v1/origin/repos/acme/rocket/pulls", `{"title":"Telemetry PR","head":"feature-1","base":"main"}`)

	// Create line-anchored comment (starts a thread)
	c1Body := `{"body":"Can we optimize this line?","line":{"path":"telemetry.txt","side":"right","startLine":1,"endLine":2}}`
	res := do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/1/comments", c1Body)
	if res.Code != http.StatusOK {
		t.Fatalf("create comment expected 200, got %d: %s", res.Code, res.Body.String())
	}

	var c1 pullRequestCommentDTO
	if err := json.Unmarshal(res.Body.Bytes(), &c1); err != nil {
		t.Fatalf("unmarshal c1: %v", err)
	}
	if c1.Body != "Can we optimize this line?" || c1.Thread.ID == "" || c1.Thread.StartLine != 1 {
		t.Fatalf("unexpected comment 1: %+v", c1)
	}

	// Reply to thread
	c2Body := `{"body":"Optimized in latest commit","threadId":"` + c1.Thread.ID + `"}`
	res = do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/1/comments", c2Body)
	if res.Code != http.StatusOK {
		t.Fatalf("reply comment expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var c2 pullRequestCommentDTO
	if err := json.Unmarshal(res.Body.Bytes(), &c2); err != nil {
		t.Fatalf("unmarshal c2: %v", err)
	}
	if c2.Thread.ID != c1.Thread.ID {
		t.Fatalf("expected same thread ID, got %s vs %s", c2.Thread.ID, c1.Thread.ID)
	}

	// List comments
	res = do(h, "GET", "/v1/origin/repos/acme/rocket/pulls/1/comments", "")
	if res.Code != http.StatusOK {
		t.Fatalf("list comments expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var listResp struct {
		Comments []pullRequestCommentDTO `json:"comments"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal list comments: %v", err)
	}
	if len(listResp.Comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(listResp.Comments))
	}

	// Resolve thread
	res = do(h, "PATCH", "/v1/origin/repos/acme/rocket/pulls/threads/"+c1.Thread.ID, `{"resolved":true}`)
	if res.Code != http.StatusOK {
		t.Fatalf("resolve thread expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var threadResp commentThreadDTO
	if err := json.Unmarshal(res.Body.Bytes(), &threadResp); err != nil {
		t.Fatalf("unmarshal thread: %v", err)
	}
	if threadResp.ResolvedAt == "" {
		t.Fatalf("expected non-empty resolvedAt")
	}

	// Add Reaction
	res = do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/comments/"+c1.ID+"/reactions", `{"emoji":"🚀"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("add reaction expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var c1WithRxn pullRequestCommentDTO
	if err := json.Unmarshal(res.Body.Bytes(), &c1WithRxn); err != nil {
		t.Fatalf("unmarshal rxn: %v", err)
	}
	if len(c1WithRxn.Reactions) != 1 || c1WithRxn.Reactions[0].Emoji != "🚀" {
		t.Fatalf("unexpected reactions: %+v", c1WithRxn.Reactions)
	}

	// Remove Reaction
	res = do(h, "DELETE", "/v1/origin/repos/acme/rocket/pulls/comments/"+c1.ID+"/reactions/🚀", "")
	if res.Code != http.StatusOK {
		t.Fatalf("remove reaction expected 200, got %d: %s", res.Code, res.Body.String())
	}

	// Delete comment 2
	res = do(h, "DELETE", "/v1/origin/repos/acme/rocket/pulls/comments/"+c2.ID, "")
	if res.Code != http.StatusOK {
		t.Fatalf("delete comment expected 200, got %d: %s", res.Code, res.Body.String())
	}
}

func TestPRSocial_ReviewsAndDismissals(t *testing.T) {
	srv, db, h := testServer(t)
	seedRepoWithBranches(t, srv, db, "acme", "rocket")

	do(h, "POST", "/v1/origin/repos/acme/rocket/pulls", `{"title":"Review PR","head":"feature-1","base":"main"}`)

	// Create Review: request_changes
	res := do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/1/reviews", `{"verdict":"request_changes","body":"Needs more tests"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("create review expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var rev pullRequestReviewDTO
	if err := json.Unmarshal(res.Body.Bytes(), &rev); err != nil {
		t.Fatalf("unmarshal rev: %v", err)
	}
	if rev.Verdict != "request_changes" || rev.Body != "Needs more tests" {
		t.Fatalf("unexpected review: %+v", rev)
	}

	// List reviews
	res = do(h, "GET", "/v1/origin/repos/acme/rocket/pulls/1/reviews", "")
	if res.Code != http.StatusOK {
		t.Fatalf("list reviews expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var listResp struct {
		Reviews []pullRequestReviewDTO `json:"reviews"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(listResp.Reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(listResp.Reviews))
	}

	// Update Review
	res = do(h, "PATCH", "/v1/origin/repos/acme/rocket/pulls/1/reviews/"+rev.ID, `{"body":"Needs integration tests"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("update review expected 200, got %d: %s", res.Code, res.Body.String())
	}

	// Dismiss Review
	res = do(h, "PUT", "/v1/origin/repos/acme/rocket/pulls/1/reviews/"+rev.ID+"/dismissals", `{"message":"Resolved in follow-up"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("dismiss review expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var dismissed pullRequestReviewDTO
	if err := json.Unmarshal(res.Body.Bytes(), &dismissed); err != nil {
		t.Fatalf("unmarshal dismissed: %v", err)
	}
	if dismissed.Dismissal == nil || dismissed.Dismissal.Message != "Resolved in follow-up" {
		t.Fatalf("unexpected dismissal: %+v", dismissed.Dismissal)
	}
}

func TestPRSocial_LabelsAndReviewers(t *testing.T) {
	srv, db, h := testServer(t)
	seedRepoWithBranches(t, srv, db, "acme", "rocket")

	do(h, "POST", "/v1/origin/repos/acme/rocket/pulls", `{"title":"Labels PR","head":"feature-1","base":"main"}`)

	// Set labels
	res := do(h, "PUT", "/v1/origin/repos/acme/rocket/pulls/1/labels", `{"labels":["bug","documentation"],"createMissing":true}`)
	if res.Code != http.StatusOK {
		t.Fatalf("set labels expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var labelsResp struct {
		Labels []labelDTO `json:"labels"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &labelsResp); err != nil {
		t.Fatalf("unmarshal labels: %v", err)
	}
	if len(labelsResp.Labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(labelsResp.Labels))
	}

	// Add label
	res = do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/1/labels", `{"labels":["enhancement"],"createMissing":true}`)
	if res.Code != http.StatusOK {
		t.Fatalf("add label expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if err := json.Unmarshal(res.Body.Bytes(), &labelsResp); err != nil {
		t.Fatalf("unmarshal labels: %v", err)
	}
	if len(labelsResp.Labels) != 3 {
		t.Fatalf("expected 3 labels, got %d", len(labelsResp.Labels))
	}

	// Remove one label
	res = do(h, "DELETE", "/v1/origin/repos/acme/rocket/pulls/1/labels/bug", "")
	if res.Code != http.StatusOK {
		t.Fatalf("remove label expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if err := json.Unmarshal(res.Body.Bytes(), &labelsResp); err != nil {
		t.Fatalf("unmarshal labels: %v", err)
	}
	if len(labelsResp.Labels) != 2 {
		t.Fatalf("expected 2 labels after remove, got %d", len(labelsResp.Labels))
	}

	// Remove all labels
	res = do(h, "DELETE", "/v1/origin/repos/acme/rocket/pulls/1/labels", "")
	if res.Code != http.StatusOK {
		t.Fatalf("remove all labels expected 200, got %d: %s", res.Code, res.Body.String())
	}

	// Request Reviewers
	res = do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/1/requested_reviewers", `{"users":["user_test"]}`)
	if res.Code != http.StatusOK {
		t.Fatalf("request reviewers expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var revResp struct {
		Users []struct {
			ID string `json:"id"`
		} `json:"users"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &revResp); err != nil {
		t.Fatalf("unmarshal reviewers: %v", err)
	}
	if len(revResp.Users) != 1 || revResp.Users[0].ID != "user_test" {
		t.Fatalf("unexpected reviewers: %+v", revResp)
	}

	// Remove Reviewers
	res = do(h, "DELETE", "/v1/origin/repos/acme/rocket/pulls/1/requested_reviewers", `{"users":["user_test"]}`)
	if res.Code != http.StatusOK {
		t.Fatalf("remove reviewers expected 200, got %d: %s", res.Code, res.Body.String())
	}
}
