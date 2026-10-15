package origin

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCreateCommitFromFiles(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	seedRepo(t, srv, "acme", "widgets")

	// Get the current head sha to satisfy expectedHeadSha.
	treeRec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/git/ref/main", "")
	if treeRec.Code != http.StatusOK {
		t.Fatalf("get ref: %d", treeRec.Code)
	}
	var refResp gitRefDTO
	_ = json.Unmarshal(treeRec.Body.Bytes(), &refResp)
	head := refResp.Object.SHA

	body := `{
		"targetBranch": "main",
		"expectedHeadSha": "` + head + `",
		"message": "add file via API",
		"author": {"name": "API", "email": "api@example.com"},
		"files": [{"path": "docs/new.md", "content": "# New\n"}]
	}`
	rec := do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/git/commits:createFromFiles", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("createFromFiles = %d; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		SHA             string `json:"sha"`
		TreeSha         string `json:"treeSha"`
		PreviousHeadSha string `json:"previousHeadSha"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.SHA == "" || resp.TreeSha == "" {
		t.Errorf("resp = %+v, want sha and treeSha", resp)
	}
	if resp.PreviousHeadSha != head {
		t.Errorf("previousHeadSha = %q, want %q", resp.PreviousHeadSha, head)
	}

	// The new file is readable via contents.
	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/contents?path=docs/new.md", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("read new file = %d", rec.Code)
	}
	var c contentDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &c)
	if c.Type != "file" {
		t.Fatalf("new file type = %q, want file", c.Type)
	}
	raw, _ := base64.StdEncoding.DecodeString(c.Content)
	if !strings.Contains(string(raw), "# New") {
		t.Errorf("new file content = %q, want '# New'", string(raw))
	}
}

func TestCreateCommitFromFilesStaleHead(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	seedRepo(t, srv, "acme", "widgets")

	body := `{
		"targetBranch": "main",
		"expectedHeadSha": "0000000000000000000000000000000000000000",
		"message": "stale",
		"author": {"name": "A", "email": "a@b.com"},
		"files": [{"path": "x.txt", "content": "x"}]
	}`
	rec := do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/git/commits:createFromFiles", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("stale head = %d, want 409 FAILED_PRECONDITION; body=%s", rec.Code, rec.Body.String())
	}
	var env struct {
		Code int `json:"code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	if env.Code != 9 {
		t.Errorf("envelope code = %d, want 9 (FAILED_PRECONDITION)", env.Code)
	}
}

func TestCreateCommitFromFilesDelete(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	head := seedRepo(t, srv, "acme", "widgets")

	body := `{
		"targetBranch": "main",
		"expectedHeadSha": "` + head + `",
		"message": "delete readme",
		"author": {"name": "A", "email": "a@b.com"},
		"files": [{"path": "README.md", "delete": true}]
	}`
	rec := do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/git/commits:createFromFiles", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete commit = %d; body=%s", rec.Code, rec.Body.String())
	}
	// README.md is gone.
	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/contents?path=README.md", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("README.md after delete = %d, want 404", rec.Code)
	}
}

func TestCreateCommitFromFilesSecretPushProtection(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	head := seedRepo(t, srv, "acme", "widgets")

	dummyKey := "AKIA" + "IOSFODNN7EXAMPLE"
	body := `{
		"targetBranch": "main",
		"expectedHeadSha": "` + head + `",
		"message": "commit with aws key",
		"author": {"name": "Dev", "email": "dev@acme.com"},
		"files": [{"path": "config.env", "content": "AWS_KEY=` + dummyKey + `\n"}]
	}`
	rec := do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/git/commits:createFromFiles", body)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 UnprocessableEntity on secret, got %d; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "PUSH REJECTED") {
		t.Errorf("expected push rejected banner in body, got: %s", rec.Body.String())
	}
}

func TestCreateGitRef(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	head := seedRepo(t, srv, "acme", "widgets")

	// Create a branch.
	rec := do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/git/refs",
		`{"ref":"refs/heads/feature-x","sha":"`+head+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("create ref = %d; body=%s", rec.Code, rec.Body.String())
	}
	var ref gitRefDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &ref)
	if ref.Ref != "refs/heads/feature-x" || ref.Object.SHA != head {
		t.Errorf("ref = %+v, want refs/heads/feature-x -> %s", ref, head)
	}

	// Duplicate → 409.
	if rec := do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/git/refs", `{"ref":"heads/feature-x","sha":"`+head+`"}`); rec.Code != http.StatusConflict {
		t.Errorf("dup ref = %d, want 409", rec.Code)
	}

	// Non-branch ref → 400.
	if rec := do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/git/refs", `{"ref":"refs/tags/v1","sha":"`+head+`"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("tag ref = %d, want 400", rec.Code)
	}

	// Get the ref.
	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/git/ref/feature-x", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get ref = %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &ref)
	if ref.Object.SHA != head {
		t.Errorf("get ref sha = %q, want %q", ref.Object.SHA, head)
	}
}

func TestListMatchingRefs(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	head := seedRepo(t, srv, "acme", "widgets")
	do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/git/refs", `{"ref":"refs/heads/feat/a","sha":"`+head+`"}`)
	do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/git/refs", `{"ref":"refs/heads/feat/b","sha":"`+head+`"}`)

	rec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/git/matching-refs/heads/feat", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("matching-refs = %d; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Refs []gitRefDTO `json:"refs"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Refs) != 2 {
		t.Fatalf("refs = %d, want 2", len(resp.Refs))
	}
	if resp.Refs[0].Ref != "refs/heads/feat/a" || resp.Refs[1].Ref != "refs/heads/feat/b" {
		t.Errorf("refs = %+v, want sorted feat/a, feat/b", resp.Refs)
	}
}

func TestDeleteGitRef(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	head := seedRepo(t, srv, "acme", "widgets")
	do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/git/refs", `{"ref":"heads/temp","sha":"`+head+`"}`)

	rec := do(h, http.MethodDelete, "/v1/origin/repos/acme/widgets/git/refs/temp", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete ref = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	// Gone.
	if rec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/git/ref/temp", ""); rec.Code != http.StatusNotFound {
		t.Errorf("get deleted ref = %d, want 404", rec.Code)
	}
	// Deleting missing → 404.
	if rec := do(h, http.MethodDelete, "/v1/origin/repos/acme/widgets/git/refs/missing", ""); rec.Code != http.StatusNotFound {
		t.Errorf("delete missing ref = %d, want 404", rec.Code)
	}
}
