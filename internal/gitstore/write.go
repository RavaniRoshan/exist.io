package gitstore

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/filemode"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// Sentinel errors for ref operations.
var (
	ErrRefExists   = errors.New("gitstore: ref already exists")
	ErrRefNotFound = errors.New("gitstore: ref not found")
)

// FileChange is one file to write in WriteFiles. Path is relative to the
// repository root. Delete removes the path from the tree when true.
type FileChange struct {
	Path    string
	Content string
	Delete  bool
}

// WriteFiles builds a new commit containing the given file changes on top of
// baseRef (empty = current branch tip, allowing an empty-tree root commit),
// then updates the branch. It returns the new commit hash. This backs both the
// createCommitFromFiles API and tests that need seed content.
func (s *Store) WriteFiles(namespaceSlug, repoName, branchRef string, changes []FileChange, author object.Signature, message string) (plumbing.Hash, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	// Resolve the base tree (root tree for a brand-new repo).
	var baseTree *object.Tree
	refName := plumbing.NewBranchReferenceName(branchRef)
	parents := []plumbing.Hash{}
	if ref, err := repo.Reference(refName, true); err == nil {
		parents = append(parents, ref.Hash())
		commit, cerr := repo.CommitObject(ref.Hash())
		if cerr == nil {
			if t, terr := commit.Tree(); terr == nil {
				baseTree = t
			}
		}
	}

	// Build the new tree in memory.
	entries, err := materializeEntries(baseTree, changes, repo)
	if err != nil {
		return plumbing.ZeroHash, err
	}
	tree := object.Tree{Entries: entries}
	treeObj := repo.Storer.NewEncodedObject()
	if err := tree.Encode(treeObj); err != nil {
		return plumbing.ZeroHash, fmt.Errorf("gitstore: encode tree: %w", err)
	}
	treeHash, err := repo.Storer.SetEncodedObject(treeObj)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	// Commit and update the branch + HEAD.
	commitHash, err := CommitObject(repo, message, author, author, treeHash, parents)
	if err != nil {
		return plumbing.ZeroHash, err
	}
	if err := repo.Storer.SetReference(plumbing.NewHashReference(refName, commitHash)); err != nil {
		return plumbing.ZeroHash, fmt.Errorf("gitstore: update ref %s: %w", branchRef, err)
	}
	// Keep HEAD symbolic so default-branch resolution keeps working.
	if _, err := repo.Reference(plumbing.HEAD, true); err != nil {
		_ = repo.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, refName))
	}
	return commitHash, nil
}

// materializeEntries merges the base tree entries with the changes into a
// tree, building intermediate trees for nested paths bottom-up.
func materializeEntries(baseTree *object.Tree, changes []FileChange, repo *git.Repository) ([]object.TreeEntry, error) {
	root := &treenode{name: "", isDir: true, children: map[string]*treenode{}}

	// Seed from the base tree recursively.
	if baseTree != nil {
		if err := seedNode(root, baseTree, repo); err != nil {
			return nil, err
		}
	}

	// Apply changes.
	for _, fc := range changes {
		parts := splitPath(fc.Path)
		if len(parts) == 0 {
			return nil, fmt.Errorf("gitstore: empty path")
		}
		cur := root
		for _, p := range parts[:len(parts)-1] {
			child, ok := cur.children[p]
			if !ok {
				child = &treenode{name: p, isDir: true, children: map[string]*treenode{}}
				cur.children[p] = child
				cur.order = append(cur.order, p)
			}
			cur = child
		}
		leaf := parts[len(parts)-1]
		if fc.Delete {
			if _, ok := cur.children[leaf]; ok {
				delete(cur.children, leaf)
				cur.order = removeStr(cur.order, leaf)
			}
			continue
		}
		blob := repo.Storer.NewEncodedObject()
		blob.SetType(plumbing.BlobObject)
		w, err := blob.Writer()
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(fc.Content)); err != nil {
			w.Close()
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		bh, err := repo.Storer.SetEncodedObject(blob)
		if err != nil {
			return nil, err
		}
		if child, ok := cur.children[leaf]; ok && child.isDir {
			// Replace a directory with a file: drop the subtree.
			cur.order = removeStr(cur.order, leaf)
			delete(cur.children, leaf)
		}
		if _, ok := cur.children[leaf]; !ok {
			cur.order = append(cur.order, leaf)
		}
		cur.children[leaf] = &treenode{name: leaf, blobHash: bh, children: map[string]*treenode{}}
	}

	// Encode trees bottom-up.
	return encodeNode(root, repo)
}

func seedNode(parent *treenode, tree *object.Tree, repo *git.Repository) error {
	for _, e := range tree.Entries {
		if e.Mode == filemode.Dir {
			sub, err := tree.Tree(e.Name)
			if err != nil {
				return err
			}
			child := &treenode{name: e.Name, isDir: true, children: map[string]*treenode{}}
			parent.children[e.Name] = child
			parent.order = append(parent.order, e.Name)
			if err := seedNode(child, sub, repo); err != nil {
				return err
			}
		} else {
			parent.children[e.Name] = &treenode{name: e.Name, blobHash: e.Hash, children: map[string]*treenode{}}
			parent.order = append(parent.order, e.Name)
		}
	}
	return nil
}

func encodeNode(n *treenode, repo *git.Repository) ([]object.TreeEntry, error) {
	var entries []object.TreeEntry
	for _, name := range n.order {
		child := n.children[name]
		if child.isDir {
			subEntries, err := encodeNode(child, repo)
			if err != nil {
				return nil, err
			}
			subTree := object.Tree{Entries: subEntries}
			obj := repo.Storer.NewEncodedObject()
			if err := subTree.Encode(obj); err != nil {
				return nil, err
			}
			h, err := repo.Storer.SetEncodedObject(obj)
			if err != nil {
				return nil, err
			}
			entries = append(entries, object.TreeEntry{Name: name, Mode: filemode.Dir, Hash: h})
		} else {
			entries = append(entries, object.TreeEntry{Name: name, Mode: filemode.Regular, Hash: child.blobHash})
		}
	}
	// Git requires tree entries sorted with directories compared as "name/".
	sort.Slice(entries, func(i, j int) bool {
		ki, kj := entries[i].Name, entries[j].Name
		if entries[i].Mode == filemode.Dir {
			ki += "/"
		}
		if entries[j].Mode == filemode.Dir {
			kj += "/"
		}
		return ki < kj
	})
	return entries, nil
}

// treenode is the tree-builder node type used to assemble trees bottom-up.
type treenode struct {
	name     string
	isDir    bool
	blobHash plumbing.Hash
	children map[string]*treenode
	order    []string
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

// RefInfo is a resolved git reference.
type RefInfo struct {
	Ref    string
	SHA    string
	IsHash bool // true when the ref resolves directly to a hash ref
}

// GetRef resolves a branch reference by short or full name.
func (s *Store) GetRef(namespaceSlug, repoName, refName string) (RefInfo, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return RefInfo{}, err
	}
	r := normalizeBranchRef(refName)
	ref, err := repo.Reference(plumbing.ReferenceName(r), true)
	if err != nil {
		return RefInfo{}, fmt.Errorf("gitstore: ref %s: %w", r, err)
	}
	return RefInfo{Ref: ref.Name().Short(), SHA: ref.Hash().String(), IsHash: ref.Type() == plumbing.HashReference}, nil
}

// CreateBranchRef creates refs/heads/<branch> pointing at sha. Only branch
// refs may be created. Returns ErrRefExists if the branch already exists.
func (s *Store) CreateBranchRef(namespaceSlug, repoName, refName, sha string) (RefInfo, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return RefInfo{}, err
	}
	if !strings.HasPrefix(refName, "refs/heads/") && !strings.HasPrefix(refName, "heads/") {
		return RefInfo{}, fmt.Errorf("gitstore: only branch references can be created, got %q", refName)
	}
	r := normalizeBranchRef(refName)
	// Verify the commit exists.
	if _, err := repo.CommitObject(plumbing.NewHash(sha)); err != nil {
		return RefInfo{}, fmt.Errorf("gitstore: commit %s not found", sha)
	}
	if _, err := repo.Reference(plumbing.ReferenceName(r), true); err == nil {
		return RefInfo{}, ErrRefExists
	}
	h := plumbing.NewHash(sha)
	if err := repo.Storer.SetReference(plumbing.NewHashReference(plumbing.ReferenceName(r), h)); err != nil {
		return RefInfo{}, err
	}
	return RefInfo{Ref: strings.TrimPrefix(r, "refs/heads/"), SHA: sha, IsHash: true}, nil
}

// CreateTagRef creates refs/tags/<tag> pointing at sha.
// Returns ErrRefExists if the tag already exists.
func (s *Store) CreateTagRef(namespaceSlug, repoName, tagName, sha string) (RefInfo, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return RefInfo{}, err
	}
	tagRef := tagName
	if !strings.HasPrefix(tagRef, "refs/tags/") {
		tagRef = "refs/tags/" + strings.TrimPrefix(tagRef, "tags/")
	}
	// Verify the commit exists.
	if _, err := repo.CommitObject(plumbing.NewHash(sha)); err != nil {
		return RefInfo{}, fmt.Errorf("gitstore: commit %s not found", sha)
	}
	if _, err := repo.Reference(plumbing.ReferenceName(tagRef), true); err == nil {
		return RefInfo{}, ErrRefExists
	}
	h := plumbing.NewHash(sha)
	if err := repo.Storer.SetReference(plumbing.NewHashReference(plumbing.ReferenceName(tagRef), h)); err != nil {
		return RefInfo{}, err
	}
	return RefInfo{Ref: strings.TrimPrefix(tagRef, "refs/tags/"), SHA: sha, IsHash: true}, nil
}

// DeleteRef removes refs/heads/<branch>. Returns ErrRefNotFound when absent.
func (s *Store) DeleteRef(namespaceSlug, repoName, refName string) error {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return err
	}
	r := normalizeBranchRef(refName)
	rn := plumbing.ReferenceName(r)
	ref, err := repo.Reference(rn, true)
	if err != nil {
		return ErrRefNotFound
	}
	// Refuse deleting the HEAD target (git convention).
	if head, herr := repo.Reference(plumbing.HEAD, true); herr == nil && ref.Hash() == head.Hash() && head.Name() == rn {
		return ErrRefNotFound
	}
	return repo.Storer.RemoveReference(rn)
}

// ListMatchingRefs returns refs whose full name starts with the given prefix
// (normalized to refs/heads/ style), sorted by full ref name.
func (s *Store) ListMatchingRefs(namespaceSlug, repoName, prefix string) ([]RefInfo, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return nil, err
	}
	p := normalizeBranchRef(prefix)
	iter, err := repo.References()
	if err != nil {
		return nil, err
	}
	var out []RefInfo
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().String()
		if strings.HasPrefix(name, p) {
			out = append(out, RefInfo{Ref: strings.TrimPrefix(name, "refs/heads/"), SHA: ref.Hash().String(), IsHash: ref.Type() == plumbing.HashReference})
		}
		return nil
	})
	return out, err
}

// normalizeBranchRef ensures the ref is in refs/heads/... form.
func normalizeBranchRef(ref string) string {
	if strings.HasPrefix(ref, "refs/") {
		return ref
	}
	if strings.HasPrefix(ref, "heads/") {
		return "refs/" + ref
	}
	return "refs/heads/" + ref
}

func removeStr(list []string, s string) []string {
	out := list[:0]
	for _, v := range list {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}
