package gitstore

import (
	"os"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(filepath.Join(t.TempDir(), "git"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestInitCreatesBareRepo(t *testing.T) {
	s := newStore(t)

	repo, err := s.Init("acme", "widgets", "main")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	// A bare repo has no worktree.
	wt, err := repo.Worktree()
	if err == nil && wt != nil {
		t.Error("expected bare repo (no worktree)")
	}

	// The directory exists on disk with a HEAD.
	path := s.repoPath("acme", "widgets")
	if _, err := os.Stat(filepath.Join(path, "HEAD")); err != nil {
		t.Errorf("HEAD not created: %v", err)
	}
	if !s.Exists("acme", "widgets") {
		t.Error("Exists = false after Init")
	}

	// HEAD points at the default branch.
	head, err := repo.Reference(plumbing.HEAD, false)
	if err != nil {
		t.Fatalf("Reference(HEAD): %v", err)
	}
	if head.Target().Short() != "main" {
		t.Errorf("HEAD target = %q, want main", head.Target().Short())
	}
}

func TestInitIdempotentConflict(t *testing.T) {
	s := newStore(t)
	if _, err := s.Init("acme", "widgets", "main"); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	if _, err := s.Init("acme", "widgets", "main"); err == nil {
		t.Error("second Init = nil, want already-exists error")
	}
}

func TestOpenRoundtrip(t *testing.T) {
	s := newStore(t)
	if _, err := s.Init("acme", "widgets", "main"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	repo, err := s.Open("acme", "widgets")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if repo == nil {
		t.Fatal("Open returned nil repo")
	}
	// Opening a nonexistent repo errors.
	if _, err := s.Open("acme", "ghost"); err == nil {
		t.Error("Open(ghost) = nil, want error")
	}
}

// TestCloneFromLocalPath roundtrips commits: build a source repo with a commit,
// clone it via the store, and confirm the commit hash matches.
func TestCloneFromLocalPath(t *testing.T) {
	// Build a source repo with one commit.
	srcDir := filepath.Join(t.TempDir(), "src")
	src, err := git.PlainInit(srcDir, false)
	if err != nil {
		t.Fatalf("PlainInit src: %v", err)
	}
	wt, err := src.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}
	f := filepath.Join(srcDir, "README.md")
	if err := os.WriteFile(f, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	sig := &object.Signature{Name: "Test", Email: "t@example.com"}
	srcHash, err := wt.Commit("initial", &git.CommitOptions{Author: sig, Committer: sig})
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// Clone via the store (local path, no auth).
	s := newStore(t)
	cloned, err := s.Clone("acme", "mirror", srcDir, nil)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	head, err := cloned.Head()
	if err != nil {
		t.Fatalf("cloned Head: %v", err)
	}
	if head.Hash() != srcHash {
		t.Errorf("cloned HEAD = %s, want %s", head.Hash(), srcHash)
	}

	// The object is retrievable.
	_, err = cloned.CommitObject(head.Hash())
	if err != nil {
		t.Errorf("CommitObject: %v", err)
	}
}

func TestCommitObject(t *testing.T) {
	s := newStore(t)
	repo, err := s.Init("acme", "widgets", "main")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Build a tree with one blob.
	blob := repo.Storer.NewEncodedObject()
	blob.SetType(plumbing.BlobObject)
	w, err := blob.Writer()
	if err != nil {
		t.Fatalf("blob writer: %v", err)
	}
	if _, err := w.Write([]byte("content")); err != nil {
		t.Fatalf("blob write: %v", err)
	}
	w.Close()
	blobHash, err := repo.Storer.SetEncodedObject(blob)
	if err != nil {
		t.Fatalf("SetEncodedObject blob: %v", err)
	}

	tree := &object.Tree{Entries: []object.TreeEntry{{
		Name: "file.txt", Mode: 0o100644, Hash: blobHash,
	}}}
	treeObj := repo.Storer.NewEncodedObject()
	if err := tree.Encode(treeObj); err != nil {
		t.Fatalf("tree encode: %v", err)
	}
	treeHash, err := repo.Storer.SetEncodedObject(treeObj)
	if err != nil {
		t.Fatalf("SetEncodedObject tree: %v", err)
	}

	sig := object.Signature{Name: "T", Email: "t@e.com"}
	h, err := CommitObject(repo, "root commit", sig, sig, treeHash, nil)
	if err != nil {
		t.Fatalf("CommitObject: %v", err)
	}
	if h == plumbing.ZeroHash {
		t.Fatal("CommitObject returned zero hash")
	}

	// The commit is retrievable and decodes to the tree we built.
	c, err := repo.CommitObject(h)
	if err != nil {
		t.Fatalf("CommitObject decode: %v", err)
	}
	if c.TreeHash != treeHash {
		t.Errorf("commit tree = %s, want %s", c.TreeHash, treeHash)
	}
	if c.Message != "root commit" {
		t.Errorf("commit message = %q", c.Message)
	}
}
