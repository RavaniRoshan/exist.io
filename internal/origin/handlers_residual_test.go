package origin

import (
	"encoding/json"
	"net/http"
	"testing"

	"exist.io/internal/gitstore"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func TestResidual_Handlers(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	headSha := seedRepo(t, srv, "acme", "widgets")

	// 1. Rate limit
	rec := do(h, http.MethodGet, "/v1/origin/rate_limit", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("rate_limit = %d; body=%s", rec.Code, rec.Body.String())
	}
	var rlResp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &rlResp)
	if rlResp["rateLimit"] == nil {
		t.Fatalf("expected rateLimit in response: %+v", rlResp)
	}

	// 2. Search pulls
	rec = do(h, http.MethodGet, "/v1/origin/search/pulls", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("search/pulls = %d; body=%s", rec.Code, rec.Body.String())
	}

	// 3. Collaborator permission
	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/collaborators/usr_alice/permission", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("collaborator permission = %d; body=%s", rec.Code, rec.Body.String())
	}

	// 4. Git tags
	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/git/tags/"+headSha, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("git tags = %d; body=%s", rec.Code, rec.Body.String())
	}

	// 5. Matching refs
	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/git/matching-refs?ref=refs/heads/", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("matching refs = %d; body=%s", rec.Code, rec.Body.String())
	}
	var refsResp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &refsResp)
	refs, ok := refsResp["gitRefs"].([]any)
	if !ok || len(refs) == 0 {
		t.Fatalf("expected at least 1 gitRef: %+v", refsResp)
	}

	// 6. Tarball
	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/tarball/main", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("tarball = %d; body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "application/x-gzip" {
		t.Fatalf("expected application/x-gzip, got %q", rec.Header().Get("Content-Type"))
	}

	// 7. Comparison files
	// Create another commit on a new branch so we can compare
	sig := object.Signature{Name: "Seed", Email: "seed@example.com"}
	_, err := srv.Git().CreateBranchRef("acme", "widgets", "refs/heads/feat", headSha)
	if err != nil {
		t.Fatalf("CreateBranchRef: %v", err)
	}
	featSha, err := srv.Git().WriteFiles("acme", "widgets", "feat", []gitstore.FileChange{
		{Path: "NEW_FILE.md", Content: "New file content\n"},
	}, sig, "add new file on feat")
	if err != nil {
		t.Fatalf("WriteFiles on feat: %v", err)
	}

	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/compare/"+headSha+"..."+featSha.String()+"/files", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("compare files = %d; body=%s", rec.Code, rec.Body.String())
	}
	var compResp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &compResp)
	files, ok := compResp["files"].([]any)
	if !ok || len(files) != 1 {
		t.Fatalf("expected 1 file change, got: %+v", compResp)
	}
}
