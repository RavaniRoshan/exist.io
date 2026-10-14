package gitstore

import (
	"fmt"
	"sort"
	"strings"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/filemode"
	"github.com/go-git/go-git/v6/plumbing/object"
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

func removeStr(list []string, s string) []string {
	out := list[:0]
	for _, v := range list {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}
