package gitstore

import (
	"fmt"
	"io"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/filemode"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// ResolveRef resolves a ref name (commit SHA, branch, tag, or symbolic like
// "HEAD") to a commit hash. An empty ref resolves to the default branch (HEAD).
func (s *Store) ResolveRef(namespaceSlug, repoName, ref string) (plumbing.Hash, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return plumbing.ZeroHash, err
	}
	return resolveRefInRepo(repo, ref)
}

func resolveRefInRepo(repo *git.Repository, ref string) (plumbing.Hash, error) {
	if ref == "" || ref == "HEAD" {
		head, err := repo.Head()
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf("gitstore: resolve HEAD: %w", err)
		}
		return head.Hash(), nil
	}
	// Try as a full hash first.
	if h := plumbing.NewHash(ref); h.String() == ref || len(ref) == 40 || len(ref) == 64 {
		if _, err := repo.CommitObject(h); err == nil {
			return h, nil
		}
	}
	// Try branch, then tag, then generic ref.
	for _, candidate := range []plumbing.ReferenceName{
		plumbing.NewBranchReferenceName(ref),
		plumbing.NewTagReferenceName(ref),
		plumbing.ReferenceName(ref),
	} {
		if r, err := repo.Reference(candidate, true); err == nil {
			return r.Hash(), nil
		}
	}
	return plumbing.ZeroHash, fmt.Errorf("gitstore: ref %q not found", ref)
}

// TreeEntry is one entry in a resolved tree, flattened for the contents API.
type TreeEntry struct {
	Mode string // octal git mode, e.g. "100644"
	Path string // path relative to the requested root
	SHA  string
	Size int64
	Type string // "blob" | "tree" | "commit"
}

// ReadTree reads a git tree at the given commit/tree hash (or ref name, like
// a branch or tag) and path. When recursive is true, it walks the whole
// subtree; otherwise only immediate children. An empty path means the root tree.
func (s *Store) ReadTree(namespaceSlug, repoName, commitHash, path string, recursive bool) (treeSHA string, entries []TreeEntry, err error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return "", nil, err
	}
	// The hash may be a commit, a tree, or a ref name. Try as a tree first, then
	// resolve to a commit and read its root tree (GitHub-compatible behavior).
	th := plumbing.NewHash(commitHash)
	if tree, terr := repo.TreeObject(th); terr == nil {
		return treeAt(tree, path, recursive)
	}
	ch, err := resolveRefInRepo(repo, commitHash)
	if err != nil {
		return "", nil, fmt.Errorf("gitstore: ref %q not found", commitHash)
	}
	commit, err := repo.CommitObject(ch)
	if err != nil {
		return "", nil, fmt.Errorf("gitstore: commit %s: %w", commitHash, err)
	}
	tree, err := commit.Tree()
	if err != nil {
		return "", nil, err
	}
	return treeAt(tree, path, recursive)
}

func treeAt(tree *object.Tree, path string, recursive bool) (string, []TreeEntry, error) {
	var entries []TreeEntry
	if path != "" {
		var err error
		tree, err = tree.Tree(path)
		if err != nil {
			return "", nil, fmt.Errorf("gitstore: tree path %q: %w", path, err)
		}
	}
	treeSHA := tree.Hash.String()
	if recursive {
		if err := walkTree(tree, path, &entries); err != nil {
			return "", nil, err
		}
	} else {
		for _, e := range tree.Entries {
			entries = append(entries, entryAt(path, e))
		}
	}
	return treeSHA, entries, nil
}

func walkTree(tree *object.Tree, prefix string, out *[]TreeEntry) error {
	for _, e := range tree.Entries {
		te := entryAt(prefix, e)
		*out = append(*out, te)
		if e.Mode == filemode.Dir {
			sub, err := tree.Tree(e.Name)
			if err != nil {
				return err
			}
			if err := walkTree(sub, joinPath(prefix, e.Name), out); err != nil {
				return err
			}
		}
	}
	return nil
}

func entryAt(prefix string, e object.TreeEntry) TreeEntry {
	te := TreeEntry{
		Mode: fmt.Sprintf("%06o", uint32(e.Mode)),
		Path: joinPath(prefix, e.Name),
		SHA:  e.Hash.String(),
		Type: "blob",
	}
	switch e.Mode {
	case filemode.Dir:
		te.Type = "tree"
	case filemode.Submodule:
		te.Type = "commit"
	}
	return te
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "/" + name
}

// ReadBlob reads a blob's raw bytes and size by hash.
func (s *Store) ReadBlob(namespaceSlug, repoName, blobHash string) (data []byte, size int64, err error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return nil, 0, err
	}
	return readBlobInRepo(repo, blobHash)
}

func readBlobInRepo(repo *git.Repository, blobHash string) ([]byte, int64, error) {
	blob, err := repo.BlobObject(plumbing.NewHash(blobHash))
	if err != nil {
		return nil, 0, fmt.Errorf("gitstore: blob %s: %w", blobHash, err)
	}
	r, err := blob.Reader()
	if err != nil {
		return nil, 0, err
	}
	defer r.Close()
	buf := make([]byte, blob.Size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, 0, err
	}
	return buf, blob.Size, nil
}

// ReadFileAtPath reads the file content at path in the given commit.
func (s *Store) ReadFileAtPath(namespaceSlug, repoName, commitHash, path string) (data []byte, sha string, size int64, isDir bool, err error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return nil, "", 0, false, err
	}
	commit, err := repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		return nil, "", 0, false, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, "", 0, false, err
	}
	if path == "" {
		return nil, tree.Hash.String(), 0, true, nil
	}
	// Directory?
	if _, err := tree.Tree(path); err == nil {
		return nil, "", 0, true, nil
	}
	file, err := tree.File(path)
	if err != nil {
		return nil, "", 0, false, fmt.Errorf("gitstore: file %q: %w", path, err)
	}
	r, err := file.Reader()
	if err != nil {
		return nil, "", 0, false, err
	}
	defer r.Close()
	buf := make([]byte, file.Size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, "", 0, false, err
	}
	return buf, file.Hash.String(), file.Size, false, nil
}

// CommitInfo is the git-object view of a commit for the contents API.
type CommitInfo struct {
	SHA                                          string
	TreeSHA                                      string
	Message                                      string
	Parents                                      []string
	AuthorName, AuthorEmail, AuthorDate          string
	CommitterName, CommitterEmail, CommitterDate string
}

// ReadCommit reads a commit object by hash.
func (s *Store) ReadCommit(namespaceSlug, repoName, commitHash string) (CommitInfo, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return CommitInfo{}, err
	}
	c, err := repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		return CommitInfo{}, fmt.Errorf("gitstore: commit %s: %w", commitHash, err)
	}
	parents := make([]string, 0, len(c.ParentHashes))
	for _, p := range c.ParentHashes {
		parents = append(parents, p.String())
	}
	return CommitInfo{
		SHA: c.Hash.String(), TreeSHA: c.TreeHash.String(), Message: c.Message, Parents: parents,
		AuthorName: c.Author.Name, AuthorEmail: c.Author.Email, AuthorDate: c.Author.When.Format("2006-01-02T15:04:05Z07:00"),
		CommitterName: c.Committer.Name, CommitterEmail: c.Committer.Email, CommitterDate: c.Committer.When.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// ListBranches returns branch name→tip SHA for the repo.
func (s *Store) ListBranches(namespaceSlug, repoName string) ([]BranchInfo, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return nil, err
	}
	iter, err := repo.Branches()
	if err != nil {
		return nil, err
	}
	var out []BranchInfo
	err = iter.ForEach(func(r *plumbing.Reference) error {
		out = append(out, BranchInfo{Name: r.Name().Short(), SHA: r.Hash().String()})
		return nil
	})
	return out, err
}

// BranchInfo is a branch name and its tip commit SHA.
type BranchInfo struct {
	Name string
	SHA  string
}

// ListCommits returns commits reachable from ref, newest first, up to limit.
func (s *Store) ListCommits(namespaceSlug, repoName, ref string, limit int) ([]CommitInfo, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return nil, err
	}
	h, err := resolveRefInRepo(repo, ref)
	if err != nil {
		return nil, err
	}
	iter, err := repo.Log(&git.LogOptions{From: h})
	if err != nil {
		return nil, err
	}
	var out []CommitInfo
	count := 0
	err = iter.ForEach(func(c *object.Commit) error {
		if count >= limit {
			return storerIterDone
		}
		parents := make([]string, 0, len(c.ParentHashes))
		for _, p := range c.ParentHashes {
			parents = append(parents, p.String())
		}
		out = append(out, CommitInfo{
			SHA: c.Hash.String(), TreeSHA: c.TreeHash.String(), Message: c.Message, Parents: parents,
			AuthorName: c.Author.Name, AuthorEmail: c.Author.Email, AuthorDate: c.Author.When.Format("2006-01-02T15:04:05Z07:00"),
			CommitterName: c.Committer.Name, CommitterEmail: c.Committer.Email, CommitterDate: c.Committer.When.Format("2006-01-02T15:04:05Z07:00"),
		})
		count++
		return nil
	})
	if err == storerIterDone {
		err = nil
	}
	return out, err
}

// storerIterDone is a sentinel to stop commit iteration early.
var storerIterDone = errStopIter{}

type errStopIter struct{}

func (errStopIter) Error() string { return "stop" }
