package claims

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	startMarker = "<!-- forbidden-claims:start -->"
	endMarker   = "<!-- forbidden-claims:end -->"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func loadForbidden(t *testing.T, path string) []string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	start := strings.Index(text, startMarker)
	end := strings.Index(text, endMarker)
	if start < 0 || end < 0 || end <= start {
		t.Fatalf("%s: forbidden-claims markers missing", path)
	}
	var phrases []string
	for _, line := range strings.Split(text[start+len(startMarker):end], "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			phrases = append(phrases, line)
		}
	}
	if len(phrases) == 0 {
		t.Fatal("forbidden-claims block is empty")
	}
	return phrases
}

func mdxFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".mdx") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestPublishedClaimsAreCoherent(t *testing.T) {
	root := repoRoot(t)
	phrases := loadForbidden(t, filepath.Join(root, "CLAIMS.md"))
	files := append([]string{filepath.Join(root, "README.md")}, mdxFiles(t, filepath.Join(root, "site", "docs"))...)

	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(body))
		for _, phrase := range phrases {
			if strings.Contains(lower, strings.ToLower(phrase)) {
				rel, _ := filepath.Rel(root, file)
				t.Errorf("%s: contains forbidden claim %q", rel, phrase)
			}
		}
	}
}
