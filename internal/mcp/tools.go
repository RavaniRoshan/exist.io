package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"exist.io/internal/store"
	"exist.io/internal/symbolsearch"
	"exist.io/pkg/idgen"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// ToolDef defines an MCP tool and its execution handler.
type ToolDef struct {
	Name        string                                                                 `json:"name"`
	Description string                                                                 `json:"description"`
	ReadOnly    bool                                                                   `json:"-"`
	InputSchema map[string]any                                                         `json:"inputSchema"`
	Handler     func(ctx context.Context, s *Server, args map[string]any) (any, error) `json:"-"`
}

func getString(args map[string]any, key string) string {
	if v, ok := args[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getInt(args map[string]any, key string) int {
	if v, ok := args[key]; ok && v != nil {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		case string:
			if i, err := strconv.Atoi(n); err == nil {
				return i
			}
		}
	}
	return 0
}

func getBool(args map[string]any, key string) bool {
	if v, ok := args[key]; ok && v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
		if s, ok := v.(string); ok {
			return s == "true" || s == "1"
		}
	}
	return false
}

func getStringSlice(args map[string]any, key string) []string {
	if v, ok := args[key]; ok && v != nil {
		if list, ok := v.([]any); ok {
			var out []string
			for _, item := range list {
				out = append(out, fmt.Sprintf("%v", item))
			}
			return out
		}
		if list, ok := v.([]string); ok {
			return list
		}
	}
	return nil
}

func makeSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

// BuiltinTools returns all defined MCP tools.
func BuiltinTools() []ToolDef {
	return []ToolDef{
		// 1. get_repository
		{
			Name:        "get_repository",
			Description: "Get metadata for an Origin repository",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Repository owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				return s.store.GetRepo(ctx, owner, name)
			},
		},
		// 2. list_repositories
		{
			Name:        "list_repositories",
			Description: "List repositories accessible to the account",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Optional owner namespace filter"},
			}, nil),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				if owner != "" {
					repos, _, err := s.store.ListRepos(ctx, owner, 100, "")
					return repos, err
				}
				namespaces, err := s.store.ListNamespaces(ctx, 100)
				if err != nil {
					return nil, err
				}
				var all []store.Repo
				for _, ns := range namespaces {
					repos, _, _ := s.store.ListRepos(ctx, ns.Slug, 100, "")
					all = append(all, repos...)
				}
				return all, nil
			},
		},
		// 3. create_repository
		{
			Name:        "create_repository",
			Description: "Create a new repository in Origin",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":     map[string]any{"type": "string", "description": "Owner namespace slug"},
				"name":          map[string]any{"type": "string", "description": "Repository name"},
				"defaultBranch": map[string]any{"type": "string", "description": "Default branch name (default: main)"},
			}, []string{"ownerSlug", "name"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "name")
				branch := getString(args, "defaultBranch")
				if branch == "" {
					branch = "main"
				}
				ns, err := s.store.GetNamespaceBySlug(ctx, owner)
				if err != nil {
					return nil, fmt.Errorf("namespace %q not found: %w", owner, err)
				}
				repo := store.Repo{
					ID:                  idgen.NewRepo(),
					NamespaceID:         ns.ID,
					Name:                name,
					FullName:            owner + "/" + name,
					DefaultBranch:       branch,
					Visibility:          "private",
					AllowMergeCommit:    true,
					AllowSquashMerge:    true,
					DeleteBranchOnMerge: false,
				}
				if err := s.store.CreateRepo(ctx, repo); err != nil {
					return nil, err
				}
				if _, err := s.git.Init(owner, name, branch); err != nil {
					return nil, fmt.Errorf("init git repo: %w", err)
				}
				return repo, nil
			},
		},
		// 4. get_file_contents
		{
			Name:        "get_file_contents",
			Description: "Read the content of a file in a repository at a given ref",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"path":      map[string]any{"type": "string", "description": "File path in repository"},
				"ref":       map[string]any{"type": "string", "description": "Commit or branch ref (default: HEAD)"},
			}, []string{"ownerSlug", "repoName", "path"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				path := getString(args, "path")
				ref := getString(args, "ref")
				if ref == "" {
					ref = "HEAD"
				}
				data, sha, size, isDir, err := s.git.ReadFileAtPath(owner, name, ref, path)
				if err != nil {
					return nil, err
				}
				if isDir {
					return map[string]any{"type": "dir", "path": path}, nil
				}
				return map[string]any{
					"path":     path,
					"sha":      sha,
					"size":     size,
					"encoding": "base64",
					"content":  base64.StdEncoding.EncodeToString(data),
				}, nil
			},
		},
		// 5. get_git_tree
		{
			Name:        "get_git_tree",
			Description: "Get the directory tree of a repository at a ref",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"ref":       map[string]any{"type": "string", "description": "Git commit SHA, branch or tag (default: HEAD)"},
				"path":      map[string]any{"type": "string", "description": "Subdirectory path to list (default: root)"},
				"recursive": map[string]any{"type": "boolean", "description": "Recursively list subtree"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				ref := getString(args, "ref")
				if ref == "" {
					ref = "HEAD"
				}
				path := getString(args, "path")
				recursive := getBool(args, "recursive")
				sha, entries, err := s.git.ReadTree(owner, name, ref, path, recursive)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"sha":     sha,
					"tree":    entries,
					"entries": len(entries),
				}, nil
			},
		},
		// 6. grep_contents
		{
			Name:        "grep_contents",
			Description: "Search file contents in a repository at a ref",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"query":     map[string]any{"type": "string", "description": "Search query or regex"},
				"ref":       map[string]any{"type": "string", "description": "Git ref (default: HEAD)"},
			}, []string{"ownerSlug", "repoName", "query"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				query := getString(args, "query")
				ref := getString(args, "ref")
				if ref == "" {
					ref = "HEAD"
				}
				matches, truncated, err := s.git.Grep(owner, name, ref, query, nil, false, false, 50)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"matches":   matches,
					"truncated": truncated,
				}, nil
			},
		},
		// 7. list_branches
		{
			Name:        "list_branches",
			Description: "List branches in a repository",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				return s.git.ListBranches(owner, name)
			},
		},
		// 8. list_commits
		{
			Name:        "list_commits",
			Description: "List commits reachable from a ref",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"ref":       map[string]any{"type": "string", "description": "Branch or commit ref (default: HEAD)"},
				"limit":     map[string]any{"type": "integer", "description": "Maximum commits to return (default: 20)"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				ref := getString(args, "ref")
				if ref == "" {
					ref = "HEAD"
				}
				limit := getInt(args, "limit")
				if limit <= 0 {
					limit = 20
				}
				return s.git.ListCommits(owner, name, ref, limit)
			},
		},
		// 9. compare_commits
		{
			Name:        "compare_commits",
			Description: "Compare differences between two commit SHAs or branches",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"base":      map[string]any{"type": "string", "description": "Base ref or commit SHA"},
				"head":      map[string]any{"type": "string", "description": "Head ref or commit SHA"},
			}, []string{"ownerSlug", "repoName", "base", "head"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				base := getString(args, "base")
				head := getString(args, "head")
				return s.git.DiffCommits(owner, name, base, head)
			},
		},
		// 10. sync_mirror
		{
			Name:        "sync_mirror",
			Description: "Trigger an immediate sync of a mirrored repository from upstream",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"synced": true,
					"repoId": repo.ID,
				}, nil
			},
		},
		// 11. list_pull_requests
		{
			Name:        "list_pull_requests",
			Description: "List pull requests in a repository",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"status":    map[string]any{"type": "string", "description": "Optional status filter (open, closed)"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				status := getString(args, "status")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pulls, _, err := s.store.ListPullRequests(ctx, repo.ID, status, 50, "")
				return pulls, err
			},
		},
		// 12. search_pull_requests
		{
			Name:        "search_pull_requests",
			Description: "Search pull requests in a repository by title or query",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"query":     map[string]any{"type": "string", "description": "Search query"},
			}, []string{"ownerSlug", "repoName", "query"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				query := strings.ToLower(getString(args, "query"))
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pulls, _, err := s.store.ListPullRequests(ctx, repo.ID, "", 100, "")
				if err != nil {
					return nil, err
				}
				var matched []store.PullRequest
				for _, pr := range pulls {
					if strings.Contains(strings.ToLower(pr.Title), query) || strings.Contains(strings.ToLower(pr.Body), query) {
						matched = append(matched, pr)
					}
				}
				return matched, nil
			},
		},
		// 13. pull_request_read
		{
			Name:        "pull_request_read",
			Description: "Get details of a pull request by number",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
			}, []string{"ownerSlug", "repoName", "pullNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				return s.store.GetPullRequest(ctx, repo.ID, num)
			},
		},
		// 14. create_pull_request
		{
			Name:        "create_pull_request",
			Description: "Open a new pull request in a repository",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"title":       map[string]any{"type": "string", "description": "Pull request title"},
				"headRef":     map[string]any{"type": "string", "description": "Head branch"},
				"baseRef":     map[string]any{"type": "string", "description": "Base branch (e.g. main)"},
				"description": map[string]any{"type": "string", "description": "Optional body/description"},
			}, []string{"ownerSlug", "repoName", "title", "headRef", "baseRef"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				head := getString(args, "headRef")
				base := getString(args, "baseRef")
				title := getString(args, "title")
				desc := getString(args, "description")

				headHash, err := s.git.ResolveRef(owner, name, head)
				if err != nil {
					return nil, fmt.Errorf("resolve head ref %q: %w", head, err)
				}
				baseHash, err := s.git.ResolveRef(owner, name, base)
				if err != nil {
					return nil, fmt.Errorf("resolve base ref %q: %w", base, err)
				}

				pr := store.PullRequest{
					ID:       idgen.NewPR(),
					RepoID:   repo.ID,
					Title:    title,
					Body:     desc,
					State:    "open",
					HeadRef:  head,
					BaseRef:  base,
					HeadSHA:  headHash.String(),
					BaseSHA:  baseHash.String(),
				}
				if err := s.store.CreatePullRequest(ctx, &pr); err != nil {
					return nil, err
				}
				return pr, nil
			},
		},
		// 15. update_pull_request
		{
			Name:        "update_pull_request",
			Description: "Update pull request title, description, or status",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber":  map[string]any{"type": "integer", "description": "Pull request number"},
				"title":       map[string]any{"type": "string", "description": "New title"},
				"description": map[string]any{"type": "string", "description": "New description"},
				"status":      map[string]any{"type": "string", "description": "New status (open or closed)"},
			}, []string{"ownerSlug", "repoName", "pullNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				var u store.PullRequestUpdate
				if t := getString(args, "title"); t != "" {
					u.Title = &t
				}
				if d := getString(args, "description"); d != "" {
					u.Body = &d
				}
				if st := getString(args, "status"); st != "" {
					u.State = &st
				}
				if err := s.store.UpdatePullRequest(ctx, pr.ID, u); err != nil {
					return nil, err
				}
				return s.store.GetPullRequest(ctx, repo.ID, num)
			},
		},
		// 16. update_pull_request_base
		{
			Name:        "update_pull_request_base",
			Description: "Update the base branch of a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"baseRef":    map[string]any{"type": "string", "description": "New base branch"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "baseRef"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				base := getString(args, "baseRef")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				baseHash, err := s.git.ResolveRef(owner, name, base)
				if err != nil {
					return nil, fmt.Errorf("resolve base ref %q: %w", base, err)
				}
				bSHA := baseHash.String()
				u := store.PullRequestUpdate{
					BaseRef: &base,
					BaseSHA: &bSHA,
				}
				if err := s.store.UpdatePullRequest(ctx, pr.ID, u); err != nil {
					return nil, err
				}
				return s.store.GetPullRequest(ctx, repo.ID, num)
			},
		},
		// 17. merge_pull_request
		{
			Name:        "merge_pull_request",
			Description: "Merge a pull request into its base branch",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber":  map[string]any{"type": "integer", "description": "Pull request number"},
				"mergeMethod": map[string]any{"type": "string", "description": "Merge method (merge, squash)"},
			}, []string{"ownerSlug", "repoName", "pullNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				method := getString(args, "mergeMethod")
				if method == "" {
					method = "merge"
				}
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				sig := object.Signature{Name: "Exist Agent", Email: "agent@exist.io", When: time.Now().UTC()}
				commitSHA, err := s.git.Merge(owner, name, pr.BaseRef, pr.HeadRef, method, sig, fmt.Sprintf("Merge PR #%d: %s", pr.Number, pr.Title))
				if err != nil {
					return nil, fmt.Errorf("git merge: %w", err)
				}
				if err := s.store.SetPullRequestMergeResult(ctx, pr.ID, commitSHA); err != nil {
					return nil, err
				}
				return s.store.GetPullRequest(ctx, repo.ID, num)
			},
		},
		// 18. create_pull_request_review
		{
			Name:        "create_pull_request_review",
			Description: "Submit a review on a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"verdict":    map[string]any{"type": "string", "description": "Review verdict (approved, changes_requested, commented)"},
				"body":       map[string]any{"type": "string", "description": "Review comment body"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "verdict"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				verdict := getString(args, "verdict")
				body := getString(args, "body")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				review := store.PRReview{
					ID:            idgen.NewPRReview(),
					RepoID:        repo.ID,
					PullRequestID: pr.ID,
					AuthorUserID:  "usr_agent",
					State:         verdict,
					Body:          body,
					CommitSHA:     pr.HeadSHA,
				}
				if err := s.store.CreateReview(ctx, review); err != nil {
					return nil, err
				}
				return review, nil
			},
		},
		// 19. update_pull_request_review
		{
			Name:        "update_pull_request_review",
			Description: "Update an existing pull request review",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"reviewId":   map[string]any{"type": "string", "description": "Review ID"},
				"body":       map[string]any{"type": "string", "description": "Updated body"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "reviewId"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				reviewID := getString(args, "reviewId")
				body := getString(args, "body")
				if err := s.store.UpdateReview(ctx, reviewID, body); err != nil {
					return nil, err
				}
				return s.store.GetReview(ctx, reviewID)
			},
		},
		// 20. dismiss_pull_request_review
		{
			Name:        "dismiss_pull_request_review",
			Description: "Dismiss a pull request review",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"reviewId":   map[string]any{"type": "string", "description": "Review ID"},
				"message":    map[string]any{"type": "string", "description": "Dismissal explanation"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "reviewId"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				reviewID := getString(args, "reviewId")
				msg := getString(args, "message")
				if err := s.store.DismissReview(ctx, reviewID, msg, "usr_agent"); err != nil {
					return nil, err
				}
				return s.store.GetReview(ctx, reviewID)
			},
		},
		// 21. create_pull_request_comment
		{
			Name:        "create_pull_request_comment",
			Description: "Add a top-level comment to a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"body":       map[string]any{"type": "string", "description": "Comment text"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "body"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				body := getString(args, "body")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				comment := store.PRComment{
					ID:            idgen.NewPRComment(),
					RepoID:        repo.ID,
					PullRequestID: pr.ID,
					AuthorUserID:  "usr_agent",
					Body:          body,
				}
				if err := s.store.CreateComment(ctx, comment); err != nil {
					return nil, err
				}
				return comment, nil
			},
		},
		// 22. create_pull_request_inline_comment
		{
			Name:        "create_pull_request_inline_comment",
			Description: "Start a review thread on a specific file line in a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"path":       map[string]any{"type": "string", "description": "File path"},
				"line":       map[string]any{"type": "integer", "description": "Line number"},
				"body":       map[string]any{"type": "string", "description": "Comment text"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "path", "line", "body"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				path := getString(args, "path")
				line := getInt(args, "line")
				body := getString(args, "body")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				thread := store.PRThread{
					ID:            idgen.NewPRThread(),
					RepoID:        repo.ID,
					PullRequestID: pr.ID,
					Path:          path,
					StartLine:     line,
					EndLine:       line,
					Side:          "RIGHT",
					HeadSHA:       pr.HeadSHA,
					BaseSHA:       pr.BaseSHA,
				}
				if err := s.store.CreateThread(ctx, thread); err != nil {
					return nil, err
				}
				comment := store.PRComment{
					ID:            idgen.NewPRComment(),
					RepoID:        repo.ID,
					PullRequestID: pr.ID,
					ThreadID:      thread.ID,
					AuthorUserID:  "usr_agent",
					Body:          body,
					Path:          path,
					Line:          line,
					Side:          "RIGHT",
					CommitSHA:     pr.HeadSHA,
				}
				if err := s.store.CreateComment(ctx, comment); err != nil {
					return nil, err
				}
				return map[string]any{
					"thread":  thread,
					"comment": comment,
				}, nil
			},
		},
		// 23. reply_pull_request_review_thread
		{
			Name:        "reply_pull_request_review_thread",
			Description: "Add a reply comment to an existing review thread",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"threadId":   map[string]any{"type": "string", "description": "Thread ID"},
				"body":       map[string]any{"type": "string", "description": "Reply comment body"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "threadId", "body"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				threadID := getString(args, "threadId")
				body := getString(args, "body")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				comment := store.PRComment{
					ID:            idgen.NewPRComment(),
					RepoID:        repo.ID,
					PullRequestID: pr.ID,
					ThreadID:      threadID,
					AuthorUserID:  "usr_agent",
					Body:          body,
				}
				if err := s.store.CreateComment(ctx, comment); err != nil {
					return nil, err
				}
				return comment, nil
			},
		},
		// 24. update_pull_request_review_thread
		{
			Name:        "update_pull_request_review_thread",
			Description: "Resolve or reopen a review thread",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"threadId":   map[string]any{"type": "string", "description": "Thread ID"},
				"resolved":   map[string]any{"type": "boolean", "description": "True to resolve, false to reopen"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "threadId", "resolved"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				threadID := getString(args, "threadId")
				resolved := getBool(args, "resolved")
				var resolvedAt *string
				var resolvedBy *string
				if resolved {
					nowStr := time.Now().UTC().Format(time.RFC3339)
					agentStr := "usr_agent"
					resolvedAt = &nowStr
					resolvedBy = &agentStr
				}
				if err := s.store.UpdateThread(ctx, threadID, resolvedAt, resolvedBy); err != nil {
					return nil, err
				}
				return s.store.GetThread(ctx, threadID)
			},
		},
		// 25. update_pull_request_comment
		{
			Name:        "update_pull_request_comment",
			Description: "Edit the body of an existing pull request comment",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"commentId":  map[string]any{"type": "string", "description": "Comment ID"},
				"body":       map[string]any{"type": "string", "description": "Updated comment body"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "commentId", "body"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				commentID := getString(args, "commentId")
				body := getString(args, "body")
				if err := s.store.UpdateComment(ctx, commentID, body); err != nil {
					return nil, err
				}
				return s.store.GetComment(ctx, commentID)
			},
		},
		// 26. update_pull_request_labels
		{
			Name:        "update_pull_request_labels",
			Description: "Add or remove labels on a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":    map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":     map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber":   map[string]any{"type": "integer", "description": "Pull request number"},
				"addLabels":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Labels to attach"},
				"removeLabels": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Labels to detach"},
			}, []string{"ownerSlug", "repoName", "pullNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				ns, err := s.store.GetNamespace(ctx, repo.NamespaceID)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				add := getStringSlice(args, "addLabels")
				remove := getStringSlice(args, "removeLabels")
				for _, lblName := range add {
					lbl, err := s.store.GetOrCreateLabel(ctx, idgen.NewLabel(), ns.ID, lblName, "ededed", "")
					if err == nil {
						_ = s.store.AddPRLabels(ctx, pr.ID, []string{lbl.ID})
					}
				}
				for _, lblName := range remove {
					lbl, err := s.store.GetLabel(ctx, ns.ID, lblName)
					if err == nil {
						_ = s.store.RemovePRLabel(ctx, pr.ID, lbl.ID)
					}
				}
				return s.store.GetPRLabels(ctx, pr.ID)
			},
		},
		// 27. create_label
		{
			Name:        "create_label",
			Description: "Create a label in a repository namespace",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"name":        map[string]any{"type": "string", "description": "Label name"},
				"color":       map[string]any{"type": "string", "description": "Label hex color (without #)"},
				"description": map[string]any{"type": "string", "description": "Optional description"},
			}, []string{"ownerSlug", "repoName", "name", "color"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				labelName := getString(args, "name")
				color := getString(args, "color")
				desc := getString(args, "description")
				ns, err := s.store.GetNamespaceBySlug(ctx, owner)
				if err != nil {
					return nil, err
				}
				return s.store.GetOrCreateLabel(ctx, idgen.NewLabel(), ns.ID, labelName, color, desc)
			},
		},
		// 28. request_pull_request_reviewers
		{
			Name:        "request_pull_request_reviewers",
			Description: "Request reviews from users on a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"reviewers":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "List of user IDs"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "reviewers"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				reviewers := getStringSlice(args, "reviewers")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				if err := s.store.AddRequestedReviewers(ctx, pr.ID, reviewers, nil); err != nil {
					return nil, err
				}
				uids, _, err := s.store.ListRequestedReviewers(ctx, pr.ID)
				return uids, err
			},
		},
		// 29. remove_pull_request_reviewers
		{
			Name:        "remove_pull_request_reviewers",
			Description: "Remove requested reviewers from a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"reviewers":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "List of user IDs to remove"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "reviewers"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				reviewers := getStringSlice(args, "reviewers")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				if err := s.store.RemoveRequestedReviewers(ctx, pr.ID, reviewers, nil); err != nil {
					return nil, err
				}
				uids, _, err := s.store.ListRequestedReviewers(ctx, pr.ID)
				return uids, err
			},
		},
		// 30. checks_read
		{
			Name:        "checks_read",
			Description: "Get check runs and check suites for a commit SHA",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"headSha":   map[string]any{"type": "string", "description": "Commit SHA to inspect"},
			}, []string{"ownerSlug", "repoName", "headSha"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				headSha := getString(args, "headSha")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				runs, err := s.store.ListCheckRunsForCommit(ctx, repo.ID, headSha, "", "", 100, 0)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"commitSha": headSha,
					"checkRuns": runs,
				}, nil
			},
		},
		// 31. list_rulesets
		{
			Name:        "list_rulesets",
			Description: "List rulesets configured on a repository",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				return s.store.ListRulesets(ctx, repo.ID)
			},
		},
		// 32. get_ruleset
		{
			Name:        "get_ruleset",
			Description: "Get details of a ruleset by ID",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"rulesetId": map[string]any{"type": "string", "description": "Ruleset ID"},
			}, []string{"ownerSlug", "repoName", "rulesetId"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				id := getString(args, "rulesetId")
				return s.store.GetRuleset(ctx, id)
			},
		},
		// 33. get_repository_access
		{
			Name:        "get_repository_access",
			Description: "Get access grants and collaborators on a repository",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				return s.store.ListRepositoryGrants(ctx, repo.ID)
			},
		},
		// 34. list_namespaces
		{
			Name:        "list_namespaces",
			Description: "List all namespaces in exist.io",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{}, nil),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				return s.store.ListNamespaces(ctx, 100)
			},
		},
		// 35. search_symbols
		{
			Name:        "search_symbols",
			Description: "Search code symbols (functions, structs, classes, types, interfaces) across repository source files",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Repository owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"query":     map[string]any{"type": "string", "description": "Symbol name or substring to search for"},
				"ref":       map[string]any{"type": "string", "description": "Optional branch, tag, or commit SHA (defaults to default branch)"},
				"kind":      map[string]any{"type": "string", "description": "Optional kind filter: function, method, struct, class, interface, type, constant"},
				"limit":     map[string]any{"type": "integer", "description": "Maximum number of results to return (default 50)"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				query := getString(args, "query")
				ref := getString(args, "ref")
				kind := getString(args, "kind")
				limit := getInt(args, "limit")
				if limit <= 0 {
					limit = 50
				}
				if ref == "" {
					repo, err := s.store.GetRepo(ctx, owner, name)
					if err == nil && repo.DefaultBranch != "" {
						ref = repo.DefaultBranch
					} else {
						ref = "HEAD"
					}
				}
				return symbolsearch.Search(ctx, s.git, owner, name, ref, query, kind, limit)
			},
		},
	}
}
