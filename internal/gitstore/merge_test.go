package gitstore

import (
	"errors"
	"testing"
	"time"

	"github.com/go-git/go-git/v6/plumbing/object"
)

func author() object.Signature {
	return object.Signature{
		Name:  "Test Dev",
		Email: "dev@exist.io",
		When:  time.Now(),
	}
}

func TestGitstore_MergeCleanAndSquash(t *testing.T) {
	s := newStore(t)
	owner, repo := "acme", "widgets"
	if _, err := s.Init(owner, repo, "main"); err != nil {
		t.Fatalf("init repo: %v", err)
	}

	// Commit 1 on main
	c1, err := s.WriteFiles(owner, repo, "main", []FileChange{
		{Path: "README.md", Content: "# Widgets\nInitial"},
	}, author(), "initial commit")
	if err != nil {
		t.Fatalf("commit 1: %v", err)
	}

	// Branch feat from main
	if _, err := s.CreateBranchRef(owner, repo, "refs/heads/feat", c1.String()); err != nil {
		t.Fatalf("create branch feat: %v", err)
	}

	// Commit on feat
	c2, err := s.WriteFiles(owner, repo, "feat", []FileChange{
		{Path: "feature.txt", Content: "feature line 1\nfeature line 2\n"},
	}, author(), "add feature")
	if err != nil {
		t.Fatalf("commit 2 on feat: %v", err)
	}

	// Diff commits
	diff, err := s.DiffCommits(owner, repo, c1.String(), c2.String())
	if err != nil {
		t.Fatalf("diff commits: %v", err)
	}
	if diff.ChangedFiles != 1 || len(diff.Files) != 1 {
		t.Fatalf("expected 1 changed file, got %+v", diff)
	}
	if diff.Files[0].Filename != "feature.txt" || diff.Files[0].Status != "added" {
		t.Fatalf("unexpected file diff: %+v", diff.Files[0])
	}
	if diff.Additions != 2 {
		t.Fatalf("expected 2 additions, got %d", diff.Additions)
	}

	// PR Commits
	prCommits, err := s.ListPRCommits(owner, repo, c1.String(), c2.String(), 10)
	if err != nil {
		t.Fatalf("list pr commits: %v", err)
	}
	if len(prCommits) != 1 || prCommits[0].SHA != c2.String() {
		t.Fatalf("expected 1 pr commit %s, got %+v", c2.String(), prCommits)
	}

	// Check mergeable
	mergeable, treeSHA, conflicts, err := s.CheckMergeable(owner, repo, "main", "feat")
	if err != nil {
		t.Fatalf("check mergeable: %v", err)
	}
	if !mergeable || treeSHA == "" || len(conflicts) > 0 {
		t.Fatalf("expected mergeable, got mergeable=%v, tree=%s, conflicts=%v", mergeable, treeSHA, conflicts)
	}

	// Prepare preview ref
	previewSHA, err := s.PrepareMergeRef(owner, repo, "main", "feat", 42, author())
	if err != nil {
		t.Fatalf("prepare merge ref: %v", err)
	}
	if previewSHA == "" {
		t.Fatalf("expected non-empty preview SHA")
	}

	// Check that refs/pull/42/merge exists
	previewRef, err := s.GetRef(owner, repo, "refs/pull/42/merge")
	if err != nil || previewRef.SHA != previewSHA {
		t.Fatalf("preview ref mismatch: %v, ref=%+v", err, previewRef)
	}

	// Perform merge
	mergeCommitSHA, err := s.Merge(owner, repo, "main", "feat", "merge", author(), "Merge feat into main")
	if err != nil {
		t.Fatalf("merge: %v", err)
	}

	mainRef, err := s.GetRef(owner, repo, "main")
	if err != nil {
		t.Fatalf("get main ref: %v", err)
	}
	if mainRef.SHA != mergeCommitSHA {
		t.Fatalf("expected main to advance to %s, got %s", mergeCommitSHA, mainRef.SHA)
	}

	// Verify both files exist on main now
	data, _, _, isDir, err := s.ReadFileAtPath(owner, repo, "main", "feature.txt")
	if err != nil || isDir || string(data) != "feature line 1\nfeature line 2\n" {
		t.Fatalf("feature.txt not in main after merge: %v, data=%s", err, string(data))
	}
}

func TestGitstore_MergeConflict(t *testing.T) {
	s := newStore(t)
	owner, repo := "acme", "widgets"
	if _, err := s.Init(owner, repo, "main"); err != nil {
		t.Fatalf("init repo: %v", err)
	}

	// Base commit on main
	baseCommit, err := s.WriteFiles(owner, repo, "main", []FileChange{
		{Path: "file.txt", Content: "line 1\nline 2\n"},
	}, author(), "base commit")
	if err != nil {
		t.Fatalf("base commit: %v", err)
	}

	// Branch feat from base
	if _, err := s.CreateBranchRef(owner, repo, "refs/heads/feat", baseCommit.String()); err != nil {
		t.Fatalf("create branch feat: %v", err)
	}

	// Divergent commit on feat
	_, err = s.WriteFiles(owner, repo, "feat", []FileChange{
		{Path: "file.txt", Content: "feat edit\nline 2\n"},
	}, author(), "feat change")
	if err != nil {
		t.Fatalf("feat commit: %v", err)
	}

	// Divergent conflicting commit on main
	_, err = s.WriteFiles(owner, repo, "main", []FileChange{
		{Path: "file.txt", Content: "main edit\nline 2\n"},
	}, author(), "main change")
	if err != nil {
		t.Fatalf("main commit: %v", err)
	}

	// Check mergeable -> false
	mergeable, _, conflicts, err := s.CheckMergeable(owner, repo, "main", "feat")
	if err != nil {
		t.Fatalf("check mergeable: %v", err)
	}
	if mergeable {
		t.Fatalf("expected mergeable = false on conflict")
	}
	if len(conflicts) == 0 {
		t.Fatalf("expected non-empty conflicts list on conflict")
	}

	// Merge attempt should return ErrMergeConflict
	_, err = s.Merge(owner, repo, "main", "feat", "merge", author(), "Merge conflict")
	if !errors.Is(err, ErrMergeConflict) {
		t.Fatalf("expected ErrMergeConflict, got %v", err)
	}
}
