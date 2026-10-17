package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

var suggestionRegex = regexp.MustCompile("(?s)```suggestion[ \\t]*(?:\\r?\\n)?(.*?)(?:\\r?\\n)?```")

// ParseSuggestion extracts replacement code from a markdown ```suggestion block in a comment body.
// It handles both "```suggestion" and "```suggestion\n" variants.
// Returns ok=false if no suggestion block is found.
func ParseSuggestion(body string) (string, bool) {
	m := suggestionRegex.FindStringSubmatch(body)
	if len(m) < 2 {
		return "", false
	}
	return m[1], true
}

// ApplySuggestion applies a code suggestion from a comment onto the pull request head branch.
// It loads the comment and thread anchors, reads the target file from the head branch,
// replaces the lines, commits the changes, and returns the new commit SHA.
func (a *App) ApplySuggestion(ctx context.Context, owner, repo string, prNumber int, commentID string) (string, error) {
	repoObj, err := a.Store.GetRepo(ctx, owner, repo)
	if err != nil {
		if rByID, err2 := a.Store.GetRepoByID(ctx, repo); err2 == nil {
			repoObj = rByID
		} else {
			return "", fmt.Errorf("repo not found: %w", err)
		}
	}
	repoName := repoObj.Name

	comment, err := a.Store.GetComment(ctx, commentID)
	if err != nil {
		return "", fmt.Errorf("comment not found: %w", err)
	}

	suggestion, ok := ParseSuggestion(comment.Body)
	if !ok {
		return "", fmt.Errorf("comment does not contain a code suggestion")
	}

	var path string
	var startLine, endLine int

	if comment.ThreadID != "" {
		thread, err := a.Store.GetThread(ctx, comment.ThreadID)
		if err != nil {
			return "", fmt.Errorf("thread not found: %w", err)
		}
		path = thread.Path
		startLine = thread.StartLine
		endLine = thread.EndLine
	}

	if path == "" {
		path = comment.Path
	}
	if startLine <= 0 && endLine > 0 {
		startLine = endLine
	} else if endLine <= 0 && startLine > 0 {
		endLine = startLine
	} else if startLine <= 0 && endLine <= 0 && comment.Line > 0 {
		startLine = comment.Line
		endLine = comment.Line
	}

	if path == "" {
		return "", fmt.Errorf("suggestion has no target file path")
	}
	if startLine <= 0 {
		return "", fmt.Errorf("invalid line range: start=%d end=%d", startLine, endLine)
	}
	if endLine < startLine {
		startLine, endLine = endLine, startLine
	}

	pr, err := a.Store.GetPullRequest(ctx, repoObj.ID, prNumber)
	if err != nil {
		return "", fmt.Errorf("pull request not found: %w", err)
	}

	if pr.Merged || pr.State != "open" {
		return "", fmt.Errorf("pull request is not open (state=%s, merged=%v)", pr.State, pr.Merged)
	}

	headBranch := strings.TrimPrefix(pr.HeadRef, "refs/heads/")
	gitRepo, err := a.Git.Open(owner, repoName)
	if err != nil {
		return "", fmt.Errorf("open git repo: %w", err)
	}

	ref, err := gitRepo.Reference(plumbing.NewBranchReferenceName(headBranch), true)
	if err != nil {
		return "", fmt.Errorf("branch ref %s: %w", headBranch, err)
	}

	commit, err := gitRepo.CommitObject(ref.Hash())
	if err != nil {
		return "", fmt.Errorf("git commit %s: %w", ref.Hash(), err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return "", fmt.Errorf("git tree: %w", err)
	}

	file, err := tree.File(path)
	if err != nil {
		return "", fmt.Errorf("git file %s: %w", path, err)
	}

	content, err := file.Contents()
	if err != nil {
		return "", fmt.Errorf("git file contents %s: %w", path, err)
	}

	hasTrailingNewline := strings.HasSuffix(content, "\n")
	contentToSplit := strings.TrimSuffix(content, "\n")
	var lines []string
	if contentToSplit != "" {
		lines = strings.Split(contentToSplit, "\n")
	}

	startIdx := startLine - 1
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx > len(lines) {
		startIdx = len(lines)
	}

	endIdx := endLine
	if endIdx < startIdx {
		endIdx = startIdx
	}
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	before := lines[:startIdx]
	after := lines[endIdx:]

	cleanSuggestion := strings.ReplaceAll(suggestion, "\r\n", "\n")
	cleanSuggestion = strings.TrimSuffix(cleanSuggestion, "\n")

	var suggestionLines []string
	if cleanSuggestion != "" {
		suggestionLines = strings.Split(cleanSuggestion, "\n")
	}

	var resultLines []string
	resultLines = append(resultLines, before...)
	resultLines = append(resultLines, suggestionLines...)
	resultLines = append(resultLines, after...)

	newContent := strings.Join(resultLines, "\n")
	if hasTrailingNewline && (len(resultLines) > 0 || len(content) > 0) {
		newContent += "\n"
	}

	author := object.Signature{
		Name:  "exist.io",
		Email: "support@exist.io",
		When:  time.Now().UTC(),
	}
	if comment.AuthorUserID != "" {
		if u, err := a.Store.GetUser(ctx, comment.AuthorUserID); err == nil {
			if u.Name != "" {
				author.Name = u.Name
			} else if u.Login != "" {
				author.Name = u.Login
			}
			if u.Email != "" {
				author.Email = u.Email
			}
		}
	}

	message := fmt.Sprintf("Apply suggestion to %s", path)
	changes := []gitstore.FileChange{
		{
			Path:    path,
			Content: newContent,
		},
	}

	newHash, err := a.Git.WriteFiles(owner, repoName, headBranch, changes, author, message)
	if err != nil {
		return "", fmt.Errorf("write files: %w", err)
	}

	newSHA := newHash.String()
	_ = a.Store.UpdatePullRequest(ctx, pr.ID, store.PullRequestUpdate{
		HeadSHA: &newSHA,
	})

	return newSHA, nil
}

func (a *App) registerSuggestionRoutes(mux *http.ServeMux, db *store.DB, git *gitstore.Store) {
	mux.HandleFunc("POST /api/v1/repos/{owner}/{name}/pulls/{number}/comments/{commentId}/apply", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		owner := r.PathValue("owner")
		name := r.PathValue("name")
		prNumberStr := r.PathValue("number")
		commentID := r.PathValue("commentId")

		prNumber, err := strconv.Atoi(prNumberStr)
		if err != nil {
			http.Error(w, "Invalid PR number", http.StatusBadRequest)
			return
		}

		appInstance := a
		if appInstance == nil {
			appInstance = &App{Store: db, Git: git}
		} else {
			if appInstance.Store == nil {
				appInstance.Store = db
			}
			if appInstance.Git == nil {
				appInstance.Git = git
			}
		}

		newSHA, err := appInstance.ApplySuggestion(r.Context(), owner, name, prNumber, commentID)
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "not found") {
				http.Error(w, errStr, http.StatusNotFound)
				return
			}
			if strings.Contains(errStr, "not open") || strings.Contains(errStr, "merged") || strings.Contains(errStr, "closed") {
				http.Error(w, errStr, http.StatusConflict)
				return
			}
			if strings.Contains(errStr, "suggestion") || strings.Contains(errStr, "invalid") {
				http.Error(w, errStr, http.StatusBadRequest)
				return
			}
			http.Error(w, errStr, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":        true,
			"commitSha": newSHA,
			"sha":       newSHA,
		})
	})
}
