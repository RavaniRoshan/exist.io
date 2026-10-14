package origin

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func seedRepoWithBranches(t *testing.T, srv *Server, db *store.DB, ownerSlug, repoName string) {
	t.Helper()
	seedNS(t, db, ownerSlug)
	ns, err := db.GetNamespaceBySlug(t.Context(), ownerSlug)
	if err != nil {
		t.Fatalf("get ns: %v", err)
	}

	repo := store.Repo{
		ID:                  idgen.NewRepo(),
		NamespaceID:         ns.ID,
		Name:                repoName,
		FullName:            ownerSlug + "/" + repoName,
		DefaultBranch:       "main",
		Visibility:          "private",
		AllowMergeCommit:    true,
		AllowSquashMerge:    true,
		DeleteBranchOnMerge: false,
	}
	if err := db.CreateRepo(t.Context(), repo); err != nil {
		t.Fatalf("create repo: %v", err)
	}
	if _, err := srv.git.Init(ownerSlug, repoName, "main"); err != nil {
		t.Fatalf("init git: %v", err)
	}

	sig := object.Signature{Name: "Tester", Email: "test@example.com", When: time.Now()}
	// Initial commit on main
	c1, err := srv.git.WriteFiles(ownerSlug, repoName, "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# Rocket\nWelcome"},
	}, sig, "initial commit")
	if err != nil {
		t.Fatalf("write main: %v", err)
	}

	// Create feature branch
	if _, err := srv.git.CreateBranchRef(ownerSlug, repoName, "refs/heads/feature-1", c1.String()); err != nil {
		t.Fatalf("create branch: %v", err)
	}

	// Commit on feature branch
	_, err = srv.git.WriteFiles(ownerSlug, repoName, "feature-1", []gitstore.FileChange{
		{Path: "telemetry.txt", Content: "sensor=1\nstatus=ok\n"},
	}, sig, "add telemetry")
	if err != nil {
		t.Fatalf("write feat: %v", err)
	}
}

func TestPullRequests_CreateAndGet(t *testing.T) {
	srv, db, h := testServer(t)
	seedRepoWithBranches(t, srv, db, "acme", "rocket")

	// Create PR
	body := `{"title":"Add telemetry","body":"Enables telemetry collection","head":"feature-1","base":"main","draft":false}`
	res := do(h, "POST", "/v1/origin/repos/acme/rocket/pulls", body)
	if res.Code != http.StatusOK {
		t.Fatalf("create PR expected 200, got %d: %s", res.Code, res.Body.String())
	}

	var created pullRequestDTO
	if err := json.Unmarshal(res.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal PR: %v", err)
	}

	if created.Number != "1" {
		t.Fatalf("expected PR #1, got %s", created.Number)
	}
	if created.Title != "Add telemetry" || created.State != "open" || created.Draft {
		t.Fatalf("unexpected PR: %+v", created)
	}
	if created.Additions != 2 || created.ChangedFiles != 1 {
		t.Fatalf("expected 2 additions and 1 changed file, got %d / %d", created.Additions, created.ChangedFiles)
	}

	// Get PR by number
	res = do(h, "GET", "/v1/origin/repos/acme/rocket/pulls/1", "")
	if res.Code != http.StatusOK {
		t.Fatalf("get PR expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var fetched pullRequestDTO
	if err := json.Unmarshal(res.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("unmarshal get PR: %v", err)
	}
	if fetched.ID != created.ID || fetched.Head.Ref != "feature-1" {
		t.Fatalf("unexpected fetched PR: %+v", fetched)
	}
}

func TestPullRequests_ListAndUpdate(t *testing.T) {
	srv, db, h := testServer(t)
	seedRepoWithBranches(t, srv, db, "acme", "rocket")

	// Create PR 1
	do(h, "POST", "/v1/origin/repos/acme/rocket/pulls", `{"title":"PR 1","head":"feature-1","base":"main","draft":false}`)

	// Update PR 1
	res := do(h, "PATCH", "/v1/origin/repos/acme/rocket/pulls/1", `{"title":"PR 1 Updated","state":"closed"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("patch PR expected 200, got %d: %s", res.Code, res.Body.String())
	}

	var updated pullRequestDTO
	if err := json.Unmarshal(res.Body.Bytes(), &updated); err != nil {
		t.Fatalf("unmarshal updated: %v", err)
	}
	if updated.Title != "PR 1 Updated" || updated.State != "closed" || updated.ClosedAt == "" {
		t.Fatalf("unexpected updated PR: %+v", updated)
	}

	// List PRs
	res = do(h, "GET", "/v1/origin/repos/acme/rocket/pulls?state=closed", "")
	if res.Code != http.StatusOK {
		t.Fatalf("list PRs expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var listResp struct {
		PullRequests []pullRequestDTO `json:"pullRequests"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(listResp.PullRequests) != 1 || listResp.PullRequests[0].Number != "1" {
		t.Fatalf("expected 1 closed PR, got %+v", listResp)
	}
}

func TestPullRequests_MergeAndMergeability(t *testing.T) {
	srv, db, h := testServer(t)
	seedRepoWithBranches(t, srv, db, "acme", "rocket")

	// Create PR
	do(h, "POST", "/v1/origin/repos/acme/rocket/pulls", `{"title":"Telemetry","head":"feature-1","base":"main"}`)

	// Get Mergeability
	res := do(h, "GET", "/v1/origin/repos/acme/rocket/pulls/1/mergeability", "")
	if res.Code != http.StatusOK {
		t.Fatalf("mergeability expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var mergeability struct {
		Verdict string `json:"verdict"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &mergeability); err != nil {
		t.Fatalf("unmarshal mergeability: %v", err)
	}
	if mergeability.Verdict != "mergeable" {
		t.Fatalf("expected mergeable verdict, got %s", mergeability.Verdict)
	}

	// Prepare merge preview ref
	res = do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/1/merge_ref", "{}")
	if res.Code != http.StatusOK {
		t.Fatalf("merge_ref expected 200, got %d: %s", res.Code, res.Body.String())
	}

	// Merge PR
	res = do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/1/merge", `{"mergeMethod":"merge"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("merge expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var mergeResp struct {
		MergeCommitSHA    string         `json:"mergeCommitSha"`
		MergedPullNumbers []string       `json:"mergedPullNumbers"`
		PullRequest       pullRequestDTO `json:"pullRequest"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &mergeResp); err != nil {
		t.Fatalf("unmarshal merge response: %v", err)
	}
	if mergeResp.MergeCommitSHA == "" || len(mergeResp.MergedPullNumbers) != 1 || mergeResp.MergedPullNumbers[0] != "1" {
		t.Fatalf("unexpected merge response: %+v", mergeResp)
	}
	if !mergeResp.PullRequest.Merged || mergeResp.PullRequest.State != "closed" {
		t.Fatalf("PR not marked merged: %+v", mergeResp.PullRequest)
	}

	// Idempotent retry
	res2 := do(h, "POST", "/v1/origin/repos/acme/rocket/pulls/1/merge", `{"mergeMethod":"merge"}`)
	if res2.Code != http.StatusOK {
		t.Fatalf("idempotent retry expected 200, got %d", res2.Code)
	}
}

func TestPullRequests_AutoMerge(t *testing.T) {
	srv, db, h := testServer(t)
	seedRepoWithBranches(t, srv, db, "acme", "rocket")

	do(h, "POST", "/v1/origin/repos/acme/rocket/pulls", `{"title":"AutoMerge PR","head":"feature-1","base":"main"}`)

	// Enable auto-merge
	res := do(h, "PUT", "/v1/origin/repos/acme/rocket/pulls/1/auto_merge", `{"mergeMethod":"squash","commitTitle":"Auto commit"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("enable auto-merge expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var pr pullRequestDTO
	if err := json.Unmarshal(res.Body.Bytes(), &pr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if pr.AutoMerge == nil || pr.AutoMerge.MergeMethod != "squash" || pr.AutoMerge.CommitTitle != "Auto commit" {
		t.Fatalf("unexpected auto-merge dto: %+v", pr.AutoMerge)
	}

	// Disable auto-merge
	res = do(h, "DELETE", "/v1/origin/repos/acme/rocket/pulls/1/auto_merge", "")
	if res.Code != http.StatusOK {
		t.Fatalf("disable auto-merge expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var prAfter pullRequestDTO
	if err := json.Unmarshal(res.Body.Bytes(), &prAfter); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if prAfter.AutoMerge != nil {
		t.Fatalf("expected auto-merge nil after delete, got %+v", prAfter.AutoMerge)
	}
}

func TestPullRequests_FilesAndCommits(t *testing.T) {
	srv, db, h := testServer(t)
	seedRepoWithBranches(t, srv, db, "acme", "rocket")

	do(h, "POST", "/v1/origin/repos/acme/rocket/pulls", `{"title":"Telemetry","head":"feature-1","base":"main"}`)

	// Files
	res := do(h, "GET", "/v1/origin/repos/acme/rocket/pulls/1/files", "")
	if res.Code != http.StatusOK {
		t.Fatalf("files expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var filesResp struct {
		Files []gitstore.PRFileChange `json:"files"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &filesResp); err != nil {
		t.Fatalf("unmarshal files: %v", err)
	}
	if len(filesResp.Files) != 1 || filesResp.Files[0].Filename != "telemetry.txt" {
		t.Fatalf("unexpected files response: %+v", filesResp)
	}

	// Commits
	res = do(h, "GET", "/v1/origin/repos/acme/rocket/pulls/1/commits", "")
	if res.Code != http.StatusOK {
		t.Fatalf("commits expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var commitsResp struct {
		Commits []struct {
			SHA    string `json:"sha"`
			Commit struct {
				Message string `json:"message"`
			} `json:"commit"`
		} `json:"commits"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &commitsResp); err != nil {
		t.Fatalf("unmarshal commits: %v", err)
	}
	if len(commitsResp.Commits) != 1 || commitsResp.Commits[0].Commit.Message != "add telemetry" {
		t.Fatalf("unexpected commits response: %+v", commitsResp)
	}
}
