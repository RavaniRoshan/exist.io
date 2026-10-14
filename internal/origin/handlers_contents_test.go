package origin

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"exist.io/internal/gitstore"

	"github.com/go-git/go-git/v6/plumbing/object"
)

// seedRepo creates a repo via the API and seeds two commits on its default
// branch, returning the HEAD commit SHA.
func seedRepo(t *testing.T, srv *Server, owner, name string) string {
	t.Helper()
	if rec := do(srv.Handler(), http.MethodPost, "/v1/origin/repos/"+owner, `{"repo":{"name":"`+name+`"}}`); rec.Code != http.StatusOK {
		t.Fatalf("create repo: %d %s", rec.Code, rec.Body.String())
	}
	sig := object.Signature{Name: "Seed", Email: "seed@example.com"}
	sha, err := srv.Git().WriteFiles(owner, name, "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# Hello\nline two mentions needle\n"},
		{Path: "src/main.go", Content: "package main\n\nfunc main() {}\n"},
		{Path: "src/util/helper.go", Content: "package util\n\nfunc Help() {}\n"},
	}, sig, "initial seed")
	if err != nil {
		t.Fatalf("WriteFiles: %v", err)
	}
	sha, err = srv.Git().WriteFiles(owner, name, "main", []gitstore.FileChange{
		{Path: "CHANGELOG.md", Content: "v1 needle found here\n"},
	}, sig, "add changelog")
	if err != nil {
		t.Fatalf("WriteFiles 2: %v", err)
	}
	return sha.String()
}

func TestGetContentsFile(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	seedRepo(t, srv, "acme", "widgets")

	rec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/contents?path=README.md", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get contents = %d; body=%s", rec.Code, rec.Body.String())
	}
	var c contentDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &c)
	if c.Type != "file" || c.Encoding != "base64" {
		t.Errorf("content = %+v, want file/base64", c)
	}
	raw, err := base64.StdEncoding.DecodeString(c.Content)
	if err != nil || !strings.Contains(string(raw), "# Hello") {
		t.Errorf("decoded content = %q, err=%v; want '# Hello'", string(raw), err)
	}
	if c.SHA == "" {
		t.Error("sha empty")
	}
}

func TestGetContentsDirectory(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	seedRepo(t, srv, "acme", "widgets")

	rec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/contents?path=src", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get dir = %d; body=%s", rec.Code, rec.Body.String())
	}
	var c contentDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &c)
	if c.Type != "dir" {
		t.Fatalf("type = %q, want dir", c.Type)
	}
	names := map[string]bool{}
	for _, e := range c.Entries {
		names[e.Name] = true
	}
	if !names["main.go"] || !names["util"] {
		t.Errorf("dir entries = %v, want main.go and util", names)
	}
}

func TestBatchGetContents(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	seedRepo(t, srv, "acme", "widgets")

	rec := do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/contents:batchGet",
		`{"paths":["README.md","ghost.txt","src/main.go"]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("batchGet = %d; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ResolvedCommitSha string `json:"resolvedCommitSha"`
		Results           []struct {
			Path    string      `json:"path"`
			Found   bool        `json:"found"`
			Content *contentDTO `json:"content"`
		} `json:"results"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.ResolvedCommitSha == "" {
		t.Error("resolvedCommitSha empty")
	}
	if len(resp.Results) != 3 {
		t.Fatalf("results = %d, want 3", len(resp.Results))
	}
	if !resp.Results[0].Found || resp.Results[0].Content == nil {
		t.Error("README.md should be found")
	}
	if resp.Results[1].Found {
		t.Error("ghost.txt should not be found")
	}
	if !resp.Results[2].Found {
		t.Error("src/main.go should be found")
	}
}

func TestGetTreeRecursive(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	sha := seedRepo(t, srv, "acme", "widgets")

	rec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/git/trees/"+sha+"?recursive=true", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get tree = %d; body=%s", rec.Code, rec.Body.String())
	}
	var treeResp struct {
		SHA  string `json:"sha"`
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
			Mode string `json:"mode"`
		} `json:"tree"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &treeResp)
	paths := map[string]string{}
	for _, e := range treeResp.Tree {
		paths[e.Path] = e.Type
	}
	if paths["README.md"] != "blob" {
		t.Errorf("README.md type = %q, want blob", paths["README.md"])
	}
	if paths["src"] != "tree" {
		t.Errorf("src type = %q, want tree", paths["src"])
	}
	if paths["src/util/helper.go"] != "blob" {
		t.Errorf("src/util/helper.go type = %q, want blob", paths["src/util/helper.go"])
	}
}

func TestGetBlobBySha(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	seedRepo(t, srv, "acme", "widgets")

	// Get the blob sha via the tree.
	treeRec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/git/trees/main?recursive=true", "")
	var treeResp struct {
		Tree []struct {
			Path string `json:"path"`
			SHA  string `json:"sha"`
			Type string `json:"type"`
		} `json:"tree"`
	}
	_ = json.Unmarshal(treeRec.Body.Bytes(), &treeResp)
	var blobSHA string
	for _, e := range treeResp.Tree {
		if e.Path == "README.md" && e.Type == "blob" {
			blobSHA = e.SHA
		}
	}
	if blobSHA == "" {
		t.Fatal("could not find README.md blob sha")
	}

	rec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/git/blobs/"+blobSHA, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get blob = %d; body=%s", rec.Code, rec.Body.String())
	}
	var blob struct {
		SHA      string `json:"sha"`
		Encoding string `json:"encoding"`
		Size     int64  `json:"size"`
		Content  string `json:"content"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &blob)
	if blob.Encoding != "base64" || blob.Size <= 0 {
		t.Errorf("blob = %+v, want base64 with size", blob)
	}
	raw, _ := base64.StdEncoding.DecodeString(blob.Content)
	if !strings.Contains(string(raw), "# Hello") {
		t.Errorf("blob content = %q", string(raw))
	}

	// Unknown sha → 404.
	if rec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/git/blobs/0000000000000000000000000000000000000000", ""); rec.Code != http.StatusNotFound {
		t.Errorf("unknown blob = %d, want 404", rec.Code)
	}
}

func TestListBranchesAndCommits(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	seedRepo(t, srv, "acme", "widgets")

	rec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/branches", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("branches = %d", rec.Code)
	}
	var br struct {
		Branches []struct {
			Name   string `json:"name"`
			Commit struct {
				SHA string `json:"sha"`
			} `json:"commit"`
		} `json:"branches"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &br)
	if len(br.Branches) != 1 || br.Branches[0].Name != "main" || br.Branches[0].Commit.SHA == "" {
		t.Errorf("branches = %+v, want [main with sha]", br.Branches)
	}

	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/commits", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("commits = %d", rec.Code)
	}
	var cm struct {
		Commits []struct {
			SHA    string `json:"sha"`
			Commit struct {
				Message string `json:"message"`
			} `json:"commit"`
		} `json:"commits"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &cm)
	if len(cm.Commits) != 2 {
		t.Fatalf("commits = %d, want 2", len(cm.Commits))
	}
	if cm.Commits[0].Commit.Message != "add changelog" {
		t.Errorf("newest commit message = %q, want 'add changelog'", cm.Commits[0].Commit.Message)
	}

	// GetCommit by sha includes parents.
	rec = do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/commits/"+cm.Commits[0].SHA, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get commit = %d", rec.Code)
	}
}

func TestGrepContents(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	seedRepo(t, srv, "acme", "widgets")

	rec := do(h, http.MethodPost, "/v1/origin/repos/acme/widgets/contents:grep", `{"query":"needle"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("grep = %d; body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Matches []struct {
			Path string `json:"path"`
			Line string `json:"line"`
		} `json:"matches"`
		LimitHit bool `json:"limitHit"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	paths := map[string]bool{}
	for _, m := range resp.Matches {
		paths[m.Path] = true
	}
	if !paths["README.md"] || !paths["CHANGELOG.md"] {
		t.Errorf("grep matches = %v, want README.md and CHANGELOG.md", paths)
	}
}

func TestTarball(t *testing.T) {
	srv, db, h := testServer(t)
	seedNS(t, db, "acme")
	seedRepo(t, srv, "acme", "widgets")

	rec := do(h, http.MethodGet, "/v1/origin/repos/acme/widgets/tarball", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("tarball = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/gzip" {
		t.Errorf("content-type = %q, want application/gzip", ct)
	}
	if !strings.HasPrefix(rec.Body.String(), "\x1f\x8b") {
		t.Error("body is not gzip (missing 1f8b magic)")
	}
}
