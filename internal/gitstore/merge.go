package gitstore

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/utils/merkletrie"
)

// ErrMergeConflict is returned when a merge cannot be completed cleanly.
var ErrMergeConflict = errors.New("gitstore: merge conflict")

// PRFileChange represents a file modified in a pull request diff.
type PRFileChange struct {
	Filename         string `json:"filename"`
	PreviousFilename string `json:"previousFilename,omitempty"`
	Status           string `json:"status"` // added | removed | modified | renamed
	Additions        int    `json:"additions"`
	Deletions        int    `json:"deletions"`
	Changes          int    `json:"changes"`
	Patch            string `json:"patch,omitempty"`
}

// PRDiffStats contains the total line and file change counts and file list.
type PRDiffStats struct {
	Additions    int
	Deletions    int
	ChangedFiles int
	Files        []PRFileChange
}

// DiffCommits computes the differences between baseSHA and headSHA.
func (s *Store) DiffCommits(namespaceSlug, repoName, baseSHA, headSHA string) (*PRDiffStats, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return nil, err
	}

	baseCommit, err := repo.CommitObject(plumbing.NewHash(baseSHA))
	if err != nil {
		return nil, fmt.Errorf("gitstore: base commit %s: %w", baseSHA, err)
	}
	headCommit, err := repo.CommitObject(plumbing.NewHash(headSHA))
	if err != nil {
		return nil, fmt.Errorf("gitstore: head commit %s: %w", headSHA, err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, err
	}
	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, err
	}

	changes, err := baseTree.Diff(headTree)
	if err != nil {
		return nil, err
	}

	stats := &PRDiffStats{}
	for _, ch := range changes {
		action, err := ch.Action()
		if err != nil {
			continue
		}
		from, to, err := ch.Files()
		if err != nil {
			continue
		}

		var fc PRFileChange
		switch action {
		case merkletrie.Insert:
			fc.Status = "added"
			if to != nil {
				fc.Filename = to.Name
			}
		case merkletrie.Delete:
			fc.Status = "removed"
			if from != nil {
				fc.Filename = from.Name
			}
		case merkletrie.Modify:
			fc.Status = "modified"
			if to != nil {
				fc.Filename = to.Name
			} else if from != nil {
				fc.Filename = from.Name
			}
			if from != nil && to != nil && from.Name != to.Name {
				fc.Status = "renamed"
				fc.PreviousFilename = from.Name
			}
		}

		if patch, err := ch.Patch(); err == nil && patch != nil {
			fc.Patch = patch.String()
			fileStats := patch.Stats()
			for _, fs := range fileStats {
				fc.Additions += fs.Addition
				fc.Deletions += fs.Deletion
			}
			fc.Changes = fc.Additions + fc.Deletions
		}

		stats.Additions += fc.Additions
		stats.Deletions += fc.Deletions
		stats.Files = append(stats.Files, fc)
	}
	stats.ChangedFiles = len(stats.Files)
	return stats, nil
}

// ListPRCommits returns the commits belonging to the PR (reachable from headSHA
// but not reachable from baseSHA), newest first.
func (s *Store) ListPRCommits(namespaceSlug, repoName, baseSHA, headSHA string, limit int) ([]CommitInfo, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return nil, err
	}

	headH := plumbing.NewHash(headSHA)
	baseH := plumbing.NewHash(baseSHA)

	if limit <= 0 {
		limit = 100
	}

	visited := make(map[plumbing.Hash]bool)
	// Mark base and all ancestors as visited/stop
	if baseCommit, err := repo.CommitObject(baseH); err == nil {
		_ = walkAncestors(repo, baseCommit, visited)
	}

	var commits []CommitInfo
	curr, err := repo.CommitObject(headH)
	if err != nil {
		return nil, fmt.Errorf("gitstore: head commit %s: %w", headSHA, err)
	}

	queue := []*object.Commit{curr}
	for len(queue) > 0 && len(commits) < limit {
		c := queue[0]
		queue = queue[1:]

		if visited[c.Hash] {
			continue
		}
		visited[c.Hash] = true

		parents := make([]string, 0, len(c.ParentHashes))
		for _, p := range c.ParentHashes {
			parents = append(parents, p.String())
			if !visited[p] {
				if parentCommit, err := repo.CommitObject(p); err == nil {
					queue = append(queue, parentCommit)
				}
			}
		}

		commits = append(commits, CommitInfo{
			SHA:           c.Hash.String(),
			TreeSHA:       c.TreeHash.String(),
			Message:       c.Message,
			Parents:       parents,
			AuthorName:    c.Author.Name,
			AuthorEmail:   c.Author.Email,
			AuthorDate:    c.Author.When.Format("2006-01-02T15:04:05Z07:00"),
			CommitterName: c.Committer.Name,
			CommitterEmail: c.Committer.Email,
			CommitterDate: c.Committer.When.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return commits, nil
}

func walkAncestors(repo *git.Repository, c *object.Commit, visited map[plumbing.Hash]bool) error {
	if visited[c.Hash] {
		return nil
	}
	visited[c.Hash] = true
	for _, p := range c.ParentHashes {
		if !visited[p] {
			if parent, err := repo.CommitObject(p); err == nil {
				_ = walkAncestors(repo, parent, visited)
			}
		}
	}
	return nil
}

// CheckMergeable evaluates whether baseRef and headRef can be merged cleanly
// without conflict. It returns the resulting tree SHA if clean, or conflicts if blocked.
func (s *Store) CheckMergeable(namespaceSlug, repoName, baseRef, headRef string) (mergeable bool, treeSHA string, conflicts []string, err error) {
	repoPath := s.repoPath(namespaceSlug, repoName)

	baseSHA, err := s.ResolveRef(namespaceSlug, repoName, baseRef)
	if err != nil {
		return false, "", nil, fmt.Errorf("resolve base ref %s: %w", baseRef, err)
	}
	headSHA, err := s.ResolveRef(namespaceSlug, repoName, headRef)
	if err != nil {
		return false, "", nil, fmt.Errorf("resolve head ref %s: %w", headRef, err)
	}

	cmd := exec.Command("git", "-C", repoPath, "merge-tree", "--write-tree", baseSHA.String(), headSHA.String())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if runErr == nil && len(lines) > 0 && len(lines[0]) == 40 {
		return true, lines[0], nil, nil
	}

	var conflictList []string
	for _, l := range lines {
		if strings.HasPrefix(l, "CONFLICT") {
			conflictList = append(conflictList, strings.TrimSpace(l))
		}
	}
	if len(conflictList) == 0 && runErr != nil {
		conflictList = append(conflictList, "Merge conflict detected")
	}

	return false, "", conflictList, nil
}

// Merge performs a 3-way merge or squash merge of headRef into baseRef and advances
// baseRef to the newly created commit.
func (s *Store) Merge(namespaceSlug, repoName, baseRef, headRef, mergeMethod string, author object.Signature, message string) (string, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return "", err
	}

	baseSHA, err := s.ResolveRef(namespaceSlug, repoName, baseRef)
	if err != nil {
		return "", fmt.Errorf("resolve base %s: %w", baseRef, err)
	}
	headSHA, err := s.ResolveRef(namespaceSlug, repoName, headRef)
	if err != nil {
		return "", fmt.Errorf("resolve head %s: %w", headRef, err)
	}

	mergeable, treeSHA, conflicts, err := s.CheckMergeable(namespaceSlug, repoName, baseRef, headRef)
	if err != nil {
		return "", err
	}
	if !mergeable {
		return "", fmt.Errorf("%w: %s", ErrMergeConflict, strings.Join(conflicts, "; "))
	}

	treeHash := plumbing.NewHash(treeSHA)
	var parents []plumbing.Hash
	if mergeMethod == "squash" {
		parents = []plumbing.Hash{baseSHA}
	} else {
		parents = []plumbing.Hash{baseSHA, headSHA}
	}

	commitHash, err := CommitObject(repo, message, author, author, treeHash, parents)
	if err != nil {
		return "", fmt.Errorf("create merge commit: %w", err)
	}

	// Update base branch ref
	refName := plumbing.ReferenceName(normalizeBranchRef(baseRef))
	if err := repo.Storer.SetReference(plumbing.NewHashReference(refName, commitHash)); err != nil {
		return "", fmt.Errorf("update branch %s: %w", baseRef, err)
	}

	return commitHash.String(), nil
}

// PrepareMergeRef creates a preview merge commit and updates refs/pull/<number>/merge.
func (s *Store) PrepareMergeRef(namespaceSlug, repoName, baseRef, headRef string, pullNumber int, author object.Signature) (string, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return "", err
	}

	baseSHA, err := s.ResolveRef(namespaceSlug, repoName, baseRef)
	if err != nil {
		return "", fmt.Errorf("resolve base %s: %w", baseRef, err)
	}
	headSHA, err := s.ResolveRef(namespaceSlug, repoName, headRef)
	if err != nil {
		return "", fmt.Errorf("resolve head %s: %w", headRef, err)
	}

	mergeable, treeSHA, conflicts, err := s.CheckMergeable(namespaceSlug, repoName, baseRef, headRef)
	if err != nil {
		return "", err
	}
	if !mergeable {
		return "", fmt.Errorf("%w: %s", ErrMergeConflict, strings.Join(conflicts, "; "))
	}

	msg := fmt.Sprintf("Merge %s into %s (preview for PR #%d)", headSHA.String()[:8], baseRef, pullNumber)
	treeHash := plumbing.NewHash(treeSHA)
	parents := []plumbing.Hash{baseSHA, headSHA}

	commitHash, err := CommitObject(repo, msg, author, author, treeHash, parents)
	if err != nil {
		return "", fmt.Errorf("create preview merge commit: %w", err)
	}

	previewRef := plumbing.ReferenceName(fmt.Sprintf("refs/pull/%d/merge", pullNumber))
	if err := repo.Storer.SetReference(plumbing.NewHashReference(previewRef, commitHash)); err != nil {
		return "", fmt.Errorf("set preview ref: %w", err)
	}

	return commitHash.String(), nil
}
