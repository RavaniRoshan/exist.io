// Package gitstore wraps go-git/v6 (pure Go, no cgo) to provide the git object
// store exist.io uses for repository data. The metadata DB (internal/store)
// holds Origin entities; this package holds the actual commits, trees, blobs,
// and refs on disk.
//
// Layout: each Origin repo maps to a bare git repository under
// <root>/<namespace-slug>/<repo-name>.git. Bare because Origin repos are
// server-side; clients push/pull over smart HTTP/SSH served elsewhere.
package gitstore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"exist.io/pkg/slug"
	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// Store manages git repositories rooted at a base directory.
type Store struct {
	root      string
	repoLocks sync.Map // map[string]*sync.RWMutex
}

// New returns a Store rooted at root, creating the directory if needed.
func New(root string) (*Store, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("gitstore: create root %s: %w", root, err)
	}
	return &Store{root: root}, nil
}

// Root returns the store's base directory.
func (s *Store) Root() string { return s.root }

// RepoLock returns the synchronization RWMutex for a specific repository.
func (s *Store) RepoLock(namespaceSlug, repoName string) *sync.RWMutex {
	key := namespaceSlug + "/" + repoName
	val, _ := s.repoLocks.LoadOrStore(key, &sync.RWMutex{})
	return val.(*sync.RWMutex)
}

// repoPath returns the on-disk path for a repo, e.g. <root>/acme/widgets.git.
// It verifies slug validity and path containment within the store root.
func (s *Store) repoPath(namespaceSlug, repoName string) string {
	if err := slug.Validate(namespaceSlug); err != nil {
		return ""
	}
	if err := slug.Validate(repoName); err != nil {
		return ""
	}
	target := filepath.Join(s.root, namespaceSlug, repoName+".git")
	cleanTarget := filepath.Clean(target)
	cleanRoot := filepath.Clean(s.root)
	if !strings.HasPrefix(cleanTarget, cleanRoot+string(filepath.Separator)) {
		return ""
	}
	return cleanTarget
}

// Init creates a new bare repository for the given owner/name with the given
// default branch, and returns it. It errors if the repo already exists.
func (s *Store) Init(namespaceSlug, repoName, defaultBranch string) (*git.Repository, error) {
	path := s.repoPath(namespaceSlug, repoName)
	if path == "" {
		return nil, fmt.Errorf("gitstore: invalid or unsafe repository slug %s/%s", namespaceSlug, repoName)
	}
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("gitstore: repo %s/%s already exists", namespaceSlug, repoName)
	}
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	repo, err := git.PlainInit(path, true) // bare
	if err != nil {
		return nil, fmt.Errorf("gitstore: init %s/%s: %w", namespaceSlug, repoName, err)
	}
	// Set the default branch via HEAD symbolic ref.
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName(defaultBranch))
	if err := repo.Storer.SetReference(headRef); err != nil {
		return nil, fmt.Errorf("gitstore: set HEAD for %s/%s: %w", namespaceSlug, repoName, err)
	}

	_ = EnsureHooks(path)

	return repo, nil
}

// EnsureHooks writes standard enterprise Git server hooks (including pre-receive secret scanning).
func EnsureHooks(repoPath string) error {
	hooksDir := filepath.Join(repoPath, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return err
	}
	preReceive := filepath.Join(hooksDir, "pre-receive")

	exe, err := os.Executable()
	if err != nil || exe == "" || strings.HasSuffix(exe, ".test") {
		exe = "existd"
	}

	content := fmt.Sprintf("#!/bin/sh\nif command -v \"%s\" >/dev/null 2>&1; then\n  exec \"%s\" hook pre-receive\nfi\nexit 0\n", exe, exe)
	return os.WriteFile(preReceive, []byte(content), 0o755)
}

// Open opens an existing bare repository.
func (s *Store) Open(namespaceSlug, repoName string) (*git.Repository, error) {
	path := s.repoPath(namespaceSlug, repoName)
	if path == "" {
		return nil, fmt.Errorf("gitstore: invalid or unsafe repository slug %s/%s", namespaceSlug, repoName)
	}
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("gitstore: open %s/%s: %w", namespaceSlug, repoName, err)
	}
	return repo, nil
}

// Exists reports whether a repo exists on disk.
func (s *Store) Exists(namespaceSlug, repoName string) bool {
	path := s.repoPath(namespaceSlug, repoName)
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// Rename renames a bare repository directory on disk from oldName.git to newName.git.
func (s *Store) Rename(namespaceSlug, oldName, newName string) error {
	oldPath := s.repoPath(namespaceSlug, oldName)
	newPath := s.repoPath(namespaceSlug, newName)
	if oldPath == "" || newPath == "" {
		return fmt.Errorf("gitstore: invalid or unsafe repository slug")
	}
	if _, err := os.Stat(oldPath); err != nil {
		return fmt.Errorf("gitstore: source repo %s/%s does not exist: %w", namespaceSlug, oldName, err)
	}
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("gitstore: destination repo %s/%s already exists", namespaceSlug, newName)
	}
	return os.Rename(oldPath, newPath)
}

// Transfer moves a bare repository directory on disk from oldNamespaceSlug/repoName.git to newNamespaceSlug/repoName.git.
func (s *Store) Transfer(oldNamespaceSlug, newNamespaceSlug, repoName string) error {
	oldPath := s.repoPath(oldNamespaceSlug, repoName)
	newPath := s.repoPath(newNamespaceSlug, repoName)
	if oldPath == "" || newPath == "" {
		return fmt.Errorf("gitstore: invalid or unsafe repository slug")
	}
	newParent := filepath.Join(s.root, newNamespaceSlug)
	if err := os.MkdirAll(newParent, 0o755); err != nil {
		return fmt.Errorf("gitstore: create target namespace dir: %w", err)
	}
	if _, err := os.Stat(oldPath); err != nil {
		return fmt.Errorf("gitstore: source repo %s/%s does not exist: %w", oldNamespaceSlug, repoName, err)
	}
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("gitstore: destination repo %s/%s already exists", newNamespaceSlug, repoName)
	}
	return os.Rename(oldPath, newPath)
}

// Clone clones a remote repository into the store as a bare repo and returns it.
// This is used by GitHub mirroring (Phase 5).
func (s *Store) Clone(namespaceSlug, repoName, remoteURL string, auth *Auth) (*git.Repository, error) {
	path := s.repoPath(namespaceSlug, repoName)
	if path == "" {
		return nil, fmt.Errorf("gitstore: invalid or unsafe repository slug %s/%s", namespaceSlug, repoName)
	}
	repo, err := git.PlainClone(path, &git.CloneOptions{
		URL:           remoteURL,
		Mirror:        true, // mirror so all refs come across
		Bare:          true,
		SingleBranch:  false,
		ClientOptions: auth.clientOptions(),
	})
	if err != nil {
		return nil, fmt.Errorf("gitstore: clone %s -> %s/%s: %w", remoteURL, namespaceSlug, repoName, err)
	}
	return repo, nil
}

// Push pushes local refs to a remote.
func (s *Store) Push(namespaceSlug, repoName, remoteURL, refspec string, auth *Auth) error {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return err
	}
	po := &git.PushOptions{RemoteName: remoteURL, RefSpecs: []config.RefSpec{config.RefSpec(refspec)}}
	po.ClientOptions = auth.clientOptions()
	if err := repo.Push(po); err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("gitstore: push %s/%s: %w", namespaceSlug, repoName, err)
	}
	return nil
}

// Commit builds a commit object in the repo's object store without touching a
// worktree. parentHashes may be empty (root commit). It returns the new commit
// hash. Used by createCommitFromFiles and merge.
func CommitObject(repo *git.Repository, message string, author object.Signature, committer object.Signature, treeHash plumbing.Hash, parentHashes []plumbing.Hash) (plumbing.Hash, error) {
	parents := make([]plumbing.Hash, 0, len(parentHashes))
	parents = append(parents, parentHashes...)
	c := &object.Commit{
		Author:       author,
		Committer:    committer,
		Message:      message,
		TreeHash:     treeHash,
		ParentHashes: parents,
	}
	obj := repo.Storer.NewEncodedObject()
	if err := c.Encode(obj); err != nil {
		return plumbing.ZeroHash, fmt.Errorf("gitstore: encode commit: %w", err)
	}
	h, err := repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("gitstore: store commit: %w", err)
	}
	return h, nil
}
