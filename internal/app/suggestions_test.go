package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func TestParseSuggestion_Basic(t *testing.T) {
	body := "```suggestion\nfmt.Println(\"hello world\")\n```"
	got, ok := ParseSuggestion(body)
	if !ok {
		t.Fatal("expected ok = true, got false")
	}
	want := "fmt.Println(\"hello world\")"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseSuggestion_MultiLine(t *testing.T) {
	body := "```suggestion\nline 1\nline 2\nline 3\n```"
	got, ok := ParseSuggestion(body)
	if !ok {
		t.Fatal("expected ok = true, got false")
	}
	want := "line 1\nline 2\nline 3"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseSuggestion_NoSuggestion(t *testing.T) {
	cases := []string{
		"This is just a normal comment.",
		"```go\nfunc main() {}\n```",
		"```\ncode block\n```",
		"",
	}
	for _, body := range cases {
		got, ok := ParseSuggestion(body)
		if ok {
			t.Fatalf("expected ok = false for body %q, got %q", body, got)
		}
	}
}

func TestParseSuggestion_EmptySuggestion(t *testing.T) {
	bodyWithNewline := "```suggestion\n```"
	got, ok := ParseSuggestion(bodyWithNewline)
	if !ok {
		t.Fatal("expected ok = true for empty suggestion with newline")
	}
	if got != "" {
		t.Fatalf("expected empty replacement, got %q", got)
	}

	bodyInline := "```suggestion```"
	got, ok = ParseSuggestion(bodyInline)
	if !ok {
		t.Fatal("expected ok = true for inline empty suggestion")
	}
	if got != "" {
		t.Fatalf("expected empty replacement, got %q", got)
	}
}

func TestParseSuggestion_WithSurroundingText(t *testing.T) {
	body := "Before the suggestion:\n\n```suggestion\nconst answer = 42\n```\n\nAfter the suggestion."
	got, ok := ParseSuggestion(body)
	if !ok {
		t.Fatal("expected ok = true, got false")
	}
	want := "const answer = 42"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func setupTestApp(t *testing.T) (*App, *store.Namespace, *store.Repo) {
	t.Helper()
	dataDir := t.TempDir()
	db, err := store.Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	git, err := gitstore.New(filepath.Join(dataDir, "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	ctx := context.Background()
	ns := store.Namespace{
		ID:   idgen.NewNS(),
		Slug: "acme",
		Type: "team",
		Name: "Acme Corp",
	}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("create namespace: %v", err)
	}

	repo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "widget",
		FullName:      "acme/widget",
		DefaultBranch: "main",
		Visibility:    "public",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("create repo: %v", err)
	}

	if _, err := git.Init("acme", "widget", "main"); err != nil {
		t.Fatalf("git.Init: %v", err)
	}

	app := &App{
		Store: db,
		Git:   git,
	}
	return app, &ns, &repo
}

func TestApplySuggestion_Lifecycle(t *testing.T) {
	app, _, repo := setupTestApp(t)
	ctx := context.Background()

	authorSig := object.Signature{
		Name:  "Developer",
		Email: "dev@acme.com",
		When:  time.Now().UTC(),
	}

	// 1. Initial main commit
	_, err := app.Git.WriteFiles("acme", "widget", "main", []gitstore.FileChange{
		{
			Path:    "calc.go",
			Content: "package calc\n\nfunc Add(a, b int) int {\n\treturn 0\n}\n",
		},
	}, authorSig, "initial commit")
	if err != nil {
		t.Fatalf("WriteFiles main: %v", err)
	}

	// 2. Branch feature-calc
	_, err = app.Git.WriteFiles("acme", "widget", "feature-calc", []gitstore.FileChange{
		{
			Path:    "calc.go",
			Content: "package calc\n\nfunc Add(a, b int) int {\n\treturn 1\n}\n",
		},
	}, authorSig, "feature branch")
	if err != nil {
		t.Fatalf("WriteFiles feature-calc: %v", err)
	}

	// 3. Create PR
	pr := store.PullRequest{
		ID:           idgen.NewPR(),
		RepoID:       repo.ID,
		Number:       1,
		State:        "open",
		Title:        "feat: improve add",
		Body:         "Implements add",
		HeadRef:      "feature-calc",
		BaseRef:      "main",
		Mergeable:    "mergeable",
		AuthorUserID: "",
	}
	if err := app.Store.CreatePullRequest(ctx, &pr); err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}

	// 4. Create Thread on line 4
	thread := store.PRThread{
		ID:            idgen.NewPRThread(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		Path:          "calc.go",
		StartLine:     4,
		EndLine:       4,
	}
	if err := app.Store.CreateThread(ctx, thread); err != nil {
		t.Fatalf("CreateThread: %v", err)
	}

	// 5. Create Comment with suggestion
	comment := store.PRComment{
		ID:            idgen.NewPRComment(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		ThreadID:      thread.ID,
		Path:          "calc.go",
		Line:          4,
		Body:          "```suggestion\n\treturn a + b\n```",
	}
	if err := app.Store.CreateComment(ctx, comment); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	// 6. Apply suggestion
	commitSHA, err := app.ApplySuggestion(ctx, "acme", "widget", 1, comment.ID)
	if err != nil {
		t.Fatalf("ApplySuggestion: %v", err)
	}
	if commitSHA == "" {
		t.Fatal("expected non-empty commit SHA")
	}

	// 7. Verify the file content on feature-calc branch
	gitRepo, err := app.Git.Open("acme", "widget")
	if err != nil {
		t.Fatalf("Git.Open: %v", err)
	}
	ref, err := gitRepo.Reference(plumbing.NewBranchReferenceName("feature-calc"), true)
	if err != nil {
		t.Fatalf("Reference feature-calc: %v", err)
	}
	commitObj, err := gitRepo.CommitObject(ref.Hash())
	if err != nil {
		t.Fatalf("CommitObject: %v", err)
	}
	treeObj, err := commitObj.Tree()
	if err != nil {
		t.Fatalf("Tree: %v", err)
	}
	fileObj, err := treeObj.File("calc.go")
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	updatedContent, err := fileObj.Contents()
	if err != nil {
		t.Fatalf("Contents: %v", err)
	}

	expectedContent := "package calc\n\nfunc Add(a, b int) int {\n\treturn a + b\n}\n"
	if updatedContent != expectedContent {
		t.Fatalf("unexpected updated content:\ngot:\n%s\nwant:\n%s", updatedContent, expectedContent)
	}
}

func TestApplySuggestion_Endpoint(t *testing.T) {
	app, _, repo := setupTestApp(t)
	ctx := context.Background()

	authorSig := object.Signature{
		Name:  "Developer",
		Email: "dev@acme.com",
		When:  time.Now().UTC(),
	}

	_, err := app.Git.WriteFiles("acme", "widget", "main", []gitstore.FileChange{
		{
			Path:    "greeting.txt",
			Content: "Hello world\n",
		},
	}, authorSig, "initial")
	if err != nil {
		t.Fatalf("WriteFiles: %v", err)
	}

	_, err = app.Git.WriteFiles("acme", "widget", "feat-greet", []gitstore.FileChange{
		{
			Path:    "greeting.txt",
			Content: "Hello original\n",
		},
	}, authorSig, "branch")
	if err != nil {
		t.Fatalf("WriteFiles feat: %v", err)
	}

	pr := store.PullRequest{
		ID:        idgen.NewPR(),
		RepoID:    repo.ID,
		Number:    42,
		State:     "open",
		Title:     "feat: update greeting",
		HeadRef:   "feat-greet",
		BaseRef:   "main",
		Mergeable: "mergeable",
	}
	if err := app.Store.CreatePullRequest(ctx, &pr); err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}

	thread := store.PRThread{
		ID:            idgen.NewPRThread(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		Path:          "greeting.txt",
		StartLine:     1,
		EndLine:       1,
	}
	if err := app.Store.CreateThread(ctx, thread); err != nil {
		t.Fatalf("CreateThread: %v", err)
	}

	comment := store.PRComment{
		ID:            idgen.NewPRComment(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		ThreadID:      thread.ID,
		Path:          "greeting.txt",
		Line:          1,
		Body:          "```suggestion\nHello from suggestion!\n```",
	}
	if err := app.Store.CreateComment(ctx, comment); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	mux := http.NewServeMux()
	app.registerSuggestionRoutes(mux, app.Store, app.Git)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/repos/acme/widget/pulls/42/comments/"+comment.ID+"/apply", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		OK        bool   `json:"ok"`
		CommitSHA string `json:"commitSha"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Decode response: %v", err)
	}
	if !resp.OK || resp.CommitSHA == "" {
		t.Fatalf("expected ok=true and non-empty commitSha, got %+v", resp)
	}
}

func TestApplySuggestion_ClosedPR(t *testing.T) {
	app, _, repo := setupTestApp(t)
	ctx := context.Background()

	pr := store.PullRequest{
		ID:      idgen.NewPR(),
		RepoID:  repo.ID,
		Number:  99,
		State:   "closed",
		Title:   "closed pr",
		HeadRef: "some-branch",
		BaseRef: "main",
	}
	if err := app.Store.CreatePullRequest(ctx, &pr); err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}

	thread := store.PRThread{
		ID:            idgen.NewPRThread(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		Path:          "file.txt",
		StartLine:     1,
		EndLine:       1,
	}
	if err := app.Store.CreateThread(ctx, thread); err != nil {
		t.Fatalf("CreateThread: %v", err)
	}

	comment := store.PRComment{
		ID:            idgen.NewPRComment(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		ThreadID:      thread.ID,
		Body:          "```suggestion\nnew\n```",
	}
	if err := app.Store.CreateComment(ctx, comment); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	_, err := app.ApplySuggestion(ctx, "acme", "widget", 99, comment.ID)
	if err == nil {
		t.Fatal("expected error applying suggestion on closed PR, got nil")
	}
}
