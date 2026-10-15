package gitstore

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"regexp"
	"strings"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/filemode"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// StreamTarball streams a gzip-compressed tar of the tree at ref directly to w.
// This prevents in-memory heap buffering of large repository snapshots.
func (s *Store) StreamTarball(w io.Writer, namespaceSlug, repoName, ref string) error {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return err
	}
	h, err := resolveRefInRepo(repo, ref)
	if err != nil {
		return err
	}
	commit, err := repo.CommitObject(h)
	if err != nil {
		return err
	}
	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	gz := gzip.NewWriter(w)
	tw := tar.NewWriter(gz)

	err = walkTreeFiles(tree, "", func(path string, mode filemode.FileMode, data []byte) error {
		osMode, _ := mode.ToOSFileMode()
		hdr := &tar.Header{Name: path, Mode: int64(osMode.Perm()), Size: int64(len(data)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err := tw.Write(data)
		return err
	})
	if err != nil {
		_ = tw.Close()
		_ = gz.Close()
		return err
	}
	if err := tw.Close(); err != nil {
		_ = gz.Close()
		return err
	}
	return gz.Close()
}

// Tarball produces a gzip-compressed tar of the tree at ref. Equivalent to
// GitHub's tarball endpoint.
func (s *Store) Tarball(namespaceSlug, repoName, ref string) ([]byte, error) {
	var buf bytes.Buffer
	if err := s.StreamTarball(&buf, namespaceSlug, repoName, ref); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func walkTreeFiles(tree *object.Tree, prefix string, fn func(path string, mode filemode.FileMode, data []byte) error) error {
	for _, e := range tree.Entries {
		path := joinPath(prefix, e.Name)
		switch e.Mode {
		case filemode.Dir:
			sub, err := tree.Tree(e.Name)
			if err != nil {
				return err
			}
			if err := walkTreeFiles(sub, path, fn); err != nil {
				return err
			}
		case filemode.Submodule:
			continue // skip submodules
		default:
			file, err := tree.File(e.Name)
			if err != nil {
				return err
			}
			contents, err := file.Contents()
			if err != nil {
				return err
			}
			if err := fn(path, e.Mode, []byte(contents)); err != nil {
				return err
			}
		}
	}
	return nil
}

// GrepMatch is one matching line with optional context.
type GrepMatch struct {
	Path       string   `json:"path"`
	LineNumber int      `json:"lineNumber"`
	Line       string   `json:"line"`
	Context    []string `json:"context,omitempty"`
}

// Grep searches file contents at ref for query, returning matching lines. It
// supports literal (optionally case-insensitive) and regexp queries.
func (s *Store) Grep(namespaceSlug, repoName, commitHash, query string, pathPrefixes []string, caseInsensitive, literal bool, maxResults int) ([]GrepMatch, bool, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return nil, false, err
	}
	commit, err := repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		return nil, false, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, false, err
	}

	var matcher func(string) bool
	if literal {
		q := query
		if caseInsensitive {
			q = strings.ToLower(q)
			matcher = func(line string) bool { return strings.Contains(strings.ToLower(line), q) }
		} else {
			matcher = func(line string) bool { return strings.Contains(line, q) }
		}
	} else {
		pat := query
		if caseInsensitive && !strings.HasPrefix(pat, "(?i)") {
			pat = "(?i)" + pat
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, false, fmt.Errorf("gitstore: bad regexp: %w", err)
		}
		matcher = re.MatchString
	}

	if maxResults <= 0 {
		maxResults = 100
	}
	var matches []GrepMatch
	limitHit := false

	err = walkTreeFiles(tree, "", func(path string, _ filemode.FileMode, data []byte) error {
		if !pathAllowed(path, pathPrefixes) {
			return nil
		}
		if isBinary(data) {
			return nil
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if matcher(line) {
				matches = append(matches, GrepMatch{Path: path, LineNumber: i + 1, Line: line})
				if len(matches) >= maxResults {
					limitHit = true
					return errStopGrep
				}
			}
		}
		return nil
	})
	if err == errStopGrep {
		err = nil
	}
	return matches, limitHit, err
}

var errStopGrep = fmt.Errorf("stop")

func pathAllowed(path string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return true
	}
	for _, p := range prefixes {
		if path == p || strings.HasPrefix(path, strings.TrimSuffix(p, "/")+"/") {
			return true
		}
	}
	return false
}

func isBinary(data []byte) bool {
	n := len(data)
	if n > 8000 {
		n = 8000
	}
	return bytes.IndexByte(data[:n], 0) >= 0
}

// CommitFile is one file changed in a commit.
type CommitFile struct {
	Path      string `json:"path"`
	Status    string `json:"status"` // added | removed | modified
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

// CommitFiles returns the files changed by a commit versus its first parent.
func (s *Store) CommitFiles(namespaceSlug, repoName, commitHash string) ([]CommitFile, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return nil, err
	}
	commit, err := repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var parentTree *object.Tree
	if commit.NumParents() > 0 {
		parent, err := commit.Parent(0)
		if err != nil {
			return nil, err
		}
		parentTree, err = parent.Tree()
		if err != nil {
			return nil, err
		}
	}

	changes, err := tree.Diff(parentTree)
	if err != nil {
		return nil, err
	}
	var files []CommitFile
	for _, ch := range changes {
		action, err := ch.Action()
		if err != nil {
			continue
		}
		from, to, err := ch.Files()
		if err != nil {
			continue
		}
		path := ""
		if to != nil && to.Name != "" {
			path = to.Name
		} else if from != nil {
			path = from.Name
		}
		files = append(files, CommitFile{Path: path, Status: action.String()})
	}
	return files, nil
}

var _ = git.NoErrAlreadyUpToDate
