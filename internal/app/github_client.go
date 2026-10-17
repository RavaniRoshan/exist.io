package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitHubClient is a lightweight GitHub REST API client.
type GitHubClient struct {
	token   string
	baseURL string
	http    *http.Client
}

// NewGitHubClient creates a new GitHub API client.
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		token:   token,
		baseURL: "https://api.github.com",
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// SetBaseURL overrides the API base URL (useful for testing against mock servers).
func (c *GitHubClient) SetBaseURL(baseURL string) {
	c.baseURL = strings.TrimRight(baseURL, "/")
}

// SetHTTPClient sets the underlying HTTP client.
func (c *GitHubClient) SetHTTPClient(client *http.Client) {
	c.http = client
}

// GitHubIssue represents an issue object in GitHub's REST API.
type GitHubIssue struct {
	Number    int           `json:"number"`
	Title     string        `json:"title"`
	Body      string        `json:"body"`
	State     string        `json:"state"` // open | closed
	User      GitHubUser    `json:"user"`
	Labels    []GitHubLabel `json:"labels"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
	ClosedAt  *string       `json:"closed_at"`
	PR        *GitHubPRRef  `json:"pull_request"` // non-nil means this is a PR, not an issue
}

// GitHubPR represents a pull request object in GitHub's REST API.
type GitHubPR struct {
	Number    int          `json:"number"`
	Title     string       `json:"title"`
	Body      string       `json:"body"`
	State     string       `json:"state"`
	User      GitHubUser   `json:"user"`
	Head      GitHubBranch `json:"head"`
	Base      GitHubBranch `json:"base"`
	Merged    bool         `json:"merged"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
	ClosedAt  *string      `json:"closed_at"`
	MergedAt  *string      `json:"merged_at"`
}

// GitHubUser represents a user login on GitHub.
type GitHubUser struct {
	Login string `json:"login"`
}

// GitHubLabel represents an issue/PR label on GitHub.
type GitHubLabel struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// GitHubComment represents a comment on an issue or pull request.
type GitHubComment struct {
	Body      string     `json:"body"`
	User      GitHubUser `json:"user"`
	CreatedAt string     `json:"created_at"`
}

// GitHubBranch represents a branch ref and SHA in a PR head/base.
type GitHubBranch struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// GitHubPRRef links an issue to its pull request representation.
type GitHubPRRef struct {
	URL string `json:"url"`
}

// parseNextLink parses an RFC 5988 / RFC 8288 Link header and extracts the URL with rel="next".
func parseNextLink(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}
	parts := strings.Split(linkHeader, ",")
	for _, part := range parts {
		start := strings.Index(part, "<")
		end := strings.Index(part, ">")
		if start == -1 || end == -1 || start >= end {
			continue
		}
		targetURL := strings.TrimSpace(part[start+1 : end])
		params := strings.Split(part[end+1:], ";")
		for _, p := range params {
			p = strings.TrimSpace(p)
			if strings.EqualFold(p, `rel="next"`) ||
				strings.EqualFold(p, `rel='next'`) ||
				strings.EqualFold(p, `rel=next`) {
				return targetURL
			}
		}
	}
	return ""
}

func (c *GitHubClient) resolveURL(ref string) string {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	return strings.TrimRight(c.baseURL, "/") + "/" + strings.TrimLeft(ref, "/")
}

func fetchAllPages[T any](c *GitHubClient, initialURL string) ([]T, error) {
	var allItems []T
	currentURL := initialURL

	for currentURL != "" {
		req, err := http.NewRequest(http.MethodGet, currentURL, nil)
		if err != nil {
			return nil, fmt.Errorf("create github request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", "exist.io-migration")
		if c.token != "" {
			if strings.HasPrefix(c.token, "Bearer ") || strings.HasPrefix(c.token, "token ") {
				req.Header.Set("Authorization", c.token)
			} else {
				req.Header.Set("Authorization", "Bearer "+c.token)
			}
		}

		httpClient := c.http
		if httpClient == nil {
			httpClient = &http.Client{Timeout: 30 * time.Second}
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("github api request %s: %w", currentURL, err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			return nil, fmt.Errorf("github api %s returned status %d: %s", currentURL, resp.StatusCode, string(body))
		}

		var pageItems []T
		err = json.NewDecoder(resp.Body).Decode(&pageItems)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decode github response %s: %w", currentURL, err)
		}

		allItems = append(allItems, pageItems...)

		if len(pageItems) == 0 {
			break
		}

		linkHeader := resp.Header.Get("Link")
		next := parseNextLink(linkHeader)
		if next != "" {
			currentURL = c.resolveURL(next)
		} else {
			currentURL = ""
		}
	}

	return allItems, nil
}

// ListIssues fetches ALL issues (paginated, state=all) for owner/repo.
func (c *GitHubClient) ListIssues(owner, repo string) ([]GitHubIssue, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/issues?state=all&per_page=100", strings.TrimRight(c.baseURL, "/"), owner, repo)
	return fetchAllPages[GitHubIssue](c, u)
}

// ListIssueComments fetches comments for an issue.
func (c *GitHubClient) ListIssueComments(owner, repo string, number int) ([]GitHubComment, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments?per_page=100", strings.TrimRight(c.baseURL, "/"), owner, repo, number)
	return fetchAllPages[GitHubComment](c, u)
}

// ListPullRequests fetches ALL pull requests (paginated, state=all) for owner/repo.
func (c *GitHubClient) ListPullRequests(owner, repo string) ([]GitHubPR, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/pulls?state=all&per_page=100", strings.TrimRight(c.baseURL, "/"), owner, repo)
	return fetchAllPages[GitHubPR](c, u)
}

// ListPRComments fetches PR review comments for owner/repo.
func (c *GitHubClient) ListPRComments(owner, repo string, number int) ([]GitHubComment, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/comments?per_page=100", strings.TrimRight(c.baseURL, "/"), owner, repo, number)
	return fetchAllPages[GitHubComment](c, u)
}

// ListLabels fetches repository labels for owner/repo.
func (c *GitHubClient) ListLabels(owner, repo string) ([]GitHubLabel, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/labels?per_page=100", strings.TrimRight(c.baseURL, "/"), owner, repo)
	return fetchAllPages[GitHubLabel](c, u)
}
