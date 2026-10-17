package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"exist.io/pkg/slug"
)

// MigrateRequest defines the parameters for a repository migration.
type MigrateRequest struct {
	CloneURL     string `json:"clone_url"`
	Owner        string `json:"owner"`
	RepoName     string `json:"repo_name"`
	GitHubToken  string `json:"github_token,omitempty"`
	ImportIssues bool   `json:"import_issues"`
	ImportPRs    bool   `json:"import_prs"`
	Private      bool   `json:"private"`
}

// MigrateStatus tracks migration progress.
type MigrateStatus struct {
	ID       string `json:"id"`
	Phase    string `json:"phase"` // cloning | importing_issues | importing_prs | complete | failed
	Progress int    `json:"progress"` // 0-100
	Error    string `json:"error,omitempty"`
}

const (
	// PhaseCloning indicates repository cloning from the remote forge.
	PhaseCloning = "cloning"
	// PhaseImportingIssues indicates issue importing is in progress.
	PhaseImportingIssues = "importing_issues"
	// PhaseImportingPRs indicates pull request importing is in progress.
	PhaseImportingPRs = "importing_prs"
	// PhaseComplete indicates the migration has finished successfully.
	PhaseComplete = "complete"
	// PhaseFailed indicates the migration failed with an error.
	PhaseFailed = "failed"
)

var (
	migrationJobs       sync.Map
	customGitHubBaseURL string
)

// SetGitHubBaseURLForTesting overrides the default GitHub API base URL during testing.
func SetGitHubBaseURLForTesting(u string) {
	customGitHubBaseURL = u
}

// getMigrationJob retrieves the current status of a migration job by ID.
func getMigrationJob(id string) (MigrateStatus, bool) {
	val, ok := migrationJobs.Load(id)
	if !ok {
		return MigrateStatus{}, false
	}
	return val.(MigrateStatus), true
}

// setMigrationJob records or updates the status of a migration job.
func setMigrationJob(status MigrateStatus) {
	migrationJobs.Store(status.ID, status)
}

// Validate checks that the migration request contains valid parameters.
func (r *MigrateRequest) Validate() error {
	r.CloneURL = strings.TrimSpace(r.CloneURL)
	if r.CloneURL == "" {
		return errors.New("clone_url is required")
	}

	isValidURL := strings.HasPrefix(r.CloneURL, "http://") ||
		strings.HasPrefix(r.CloneURL, "https://") ||
		strings.HasPrefix(r.CloneURL, "git://") ||
		strings.HasPrefix(r.CloneURL, "ssh://") ||
		strings.HasPrefix(r.CloneURL, "git@") ||
		strings.HasPrefix(r.CloneURL, "file://") ||
		filepath.IsAbs(r.CloneURL)

	if !isValidURL {
		return errors.New("clone_url must be a valid HTTP, HTTPS, Git, SSH, or file URL")
	}

	r.Owner = strings.TrimSpace(r.Owner)
	if r.Owner == "" {
		return errors.New("owner is required")
	}
	if err := slug.Validate(r.Owner); err != nil {
		return fmt.Errorf("invalid owner: %w", err)
	}

	r.RepoName = strings.TrimSpace(r.RepoName)
	if r.RepoName == "" {
		return errors.New("repo_name is required")
	}
	if err := slug.Validate(r.RepoName); err != nil {
		return fmt.Errorf("invalid repo_name: %w", err)
	}

	return nil
}

// parseGitHubOwnerRepo attempts to extract owner and repository names from a clone URL.
func parseGitHubOwnerRepo(cloneURL string) (string, string) {
	s := strings.TrimSpace(cloneURL)
	s = strings.TrimSuffix(s, ".git")
	s = strings.TrimSuffix(s, "/")

	// Handle SCP-style: git@github.com:owner/repo
	if strings.Contains(s, ":") && !strings.Contains(s, "://") {
		parts := strings.Split(s, ":")
		if len(parts) == 2 {
			sub := strings.Trim(parts[1], "/")
			segments := strings.Split(sub, "/")
			if len(segments) >= 2 {
				return segments[len(segments)-2], segments[len(segments)-1]
			}
		}
	}

	// Handle standard URL schemes
	if u, err := url.Parse(s); err == nil && u.Path != "" {
		path := strings.Trim(u.Path, "/")
		segments := strings.Split(path, "/")
		if len(segments) >= 2 {
			return segments[len(segments)-2], segments[len(segments)-1]
		}
	}

	// Generic slash-delimited fallback
	segments := strings.Split(strings.Trim(s, "/"), "/")
	if len(segments) >= 2 {
		return segments[len(segments)-2], segments[len(segments)-1]
	}

	return "", ""
}

// registerMigrateRoutes registers the repository migration HTTP endpoints.
func (a *App) registerMigrateRoutes(mux *http.ServeMux, db *store.DB, git *gitstore.Store) {
	// POST /api/v1/repos/migrate - starts migration async, returns job ID
	mux.HandleFunc("POST /api/v1/repos/migrate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var req MigrateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ns, err := db.GetNamespaceBySlug(r.Context(), req.Owner)
		if err != nil {
			http.Error(w, fmt.Sprintf("owner %q not found", req.Owner), http.StatusBadRequest)
			return
		}

		if _, err := db.GetRepo(r.Context(), req.Owner, req.RepoName); err == nil {
			http.Error(w, fmt.Sprintf("repository %s/%s already exists", req.Owner, req.RepoName), http.StatusConflict)
			return
		}

		jobID := idgen.New("mig")
		initialStatus := MigrateStatus{
			ID:       jobID,
			Phase:    PhaseCloning,
			Progress: 0,
		}
		setMigrationJob(initialStatus)

		go a.runMigration(jobID, req, ns, db, git)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":       jobID,
			"job_id":   jobID,
			"phase":    initialStatus.Phase,
			"progress": initialStatus.Progress,
		})
	})

	// GET /api/v1/repos/migrate/{jobId}/status - returns current status
	statusHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		jobID := r.PathValue("jobId")
		if jobID == "" {
			jobID = r.PathValue("id")
		}

		status, ok := getMigrationJob(jobID)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "migration job not found"})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(status)
	}

	mux.HandleFunc("GET /api/v1/repos/migrate/{jobId}/status", statusHandler)
	mux.HandleFunc("GET /api/v1/repos/migrate/{jobId}", statusHandler)
}

func (a *App) runMigration(jobID string, req MigrateRequest, ns store.Namespace, db *store.DB, git *gitstore.Store) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	a.runMigrationWithContext(ctx, jobID, req, ns, db, git)
}

func (a *App) runMigrationWithContext(ctx context.Context, jobID string, req MigrateRequest, ns store.Namespace, db *store.DB, git *gitstore.Store) {
	status := MigrateStatus{
		ID:       jobID,
		Phase:    PhaseCloning,
		Progress: 10,
	}
	setMigrationJob(status)

	// 2. Create repo record in SQLite with store.CreateRepo
	fullName := req.Owner + "/" + req.RepoName
	var publicURL string
	if a != nil {
		publicURL = a.Config.PublicURL
	}
	cloneHTTP := fmt.Sprintf("%s/%s/%s.git", strings.TrimRight(publicURL, "/"), req.Owner, req.RepoName)
	webURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(publicURL, "/"), req.Owner, req.RepoName)

	visibility := "public"
	if req.Private {
		visibility = "private"
	}

	repoRecord := store.Repo{
		ID:                  idgen.NewRepo(),
		NamespaceID:         ns.ID,
		Name:                req.RepoName,
		FullName:            fullName,
		Description:         "",
		DefaultBranch:       "main",
		Visibility:          visibility,
		CloneURL:            cloneHTTP,
		WebURL:              webURL,
		AllowMergeCommit:    true,
		AllowSquashMerge:    true,
		DeleteBranchOnMerge: false,
	}

	if err := db.CreateRepo(ctx, repoRecord); err != nil {
		status.Phase = PhaseFailed
		status.Error = fmt.Sprintf("failed to create repo record: %v", err)
		setMigrationJob(status)
		return
	}

	// 3. Clone the repository using gitstore.Clone(owner, repoName, cloneURL, auth)
	var gitAuth *gitstore.Auth
	if req.GitHubToken != "" {
		gitAuth = &gitstore.Auth{
			Username: "git",
			Password: req.GitHubToken,
		}
	}

	gitRepo, err := git.Clone(req.Owner, req.RepoName, req.CloneURL, gitAuth)
	if err != nil {
		status.Phase = PhaseFailed
		status.Error = fmt.Sprintf("git clone failed: %v", err)
		setMigrationJob(status)
		return
	}

	// Update default branch if detected from cloned repo HEAD
	if headRef, err := gitRepo.Head(); err == nil && headRef.Name().IsBranch() {
		branch := headRef.Name().Short()
		if branch != "" && branch != repoRecord.DefaultBranch {
			_ = db.UpdateRepo(ctx, req.Owner, req.RepoName, store.RepoUpdate{
				DefaultBranch: &branch,
			})
			repoRecord.DefaultBranch = branch
		}
	}

	status.Progress = 30
	setMigrationJob(status)

	// Determine if GitHub API import is requested with a token
	needsGitHubAPI := (req.ImportIssues || req.ImportPRs) && req.GitHubToken != ""
	if !needsGitHubAPI {
		status.Phase = PhaseComplete
		status.Progress = 100
		setMigrationJob(status)
		return
	}

	ghOwner, ghRepo := parseGitHubOwnerRepo(req.CloneURL)
	if ghOwner == "" || ghRepo == "" {
		ghOwner = req.Owner
		ghRepo = req.RepoName
	}

	ghClient := NewGitHubClient(req.GitHubToken)
	if customGitHubBaseURL != "" {
		ghClient.SetBaseURL(customGitHubBaseURL)
	} else if strings.HasPrefix(req.CloneURL, "http://127.0.0.1:") || strings.HasPrefix(req.CloneURL, "http://localhost:") {
		if u, err := url.Parse(req.CloneURL); err == nil {
			ghClient.SetBaseURL(fmt.Sprintf("%s://%s", u.Scheme, u.Host))
		}
	}

	// 4. If GitHub token provided and ImportIssues=true, fetch and create issues
	if req.ImportIssues {
		status.Phase = PhaseImportingIssues
		status.Progress = 40
		setMigrationJob(status)

		// Pre-populate repository labels if available
		if labels, err := ghClient.ListLabels(ghOwner, ghRepo); err == nil {
			for _, lbl := range labels {
				_, _ = db.GetOrCreateLabel(ctx, idgen.NewLabel(), ns.ID, lbl.Name, lbl.Color, "")
			}
		}

		issues, err := ghClient.ListIssues(ghOwner, ghRepo)
		if err != nil {
			status.Phase = PhaseFailed
			status.Error = fmt.Sprintf("failed to fetch issues: %v", err)
			setMigrationJob(status)
			return
		}

		// Filter out pull requests which GitHub includes in /issues
		var actualIssues []GitHubIssue
		for _, issue := range issues {
			if issue.PR == nil {
				actualIssues = append(actualIssues, issue)
			}
		}

		totalIssues := len(actualIssues)
		for idx, ghIssue := range actualIssues {
			var authorUserID string
			if ghIssue.User.Login != "" {
				if u, err := db.GetUserByLogin(ctx, ghIssue.User.Login); err == nil {
					authorUserID = u.ID
				}
			}

			closedAt := ""
			if ghIssue.ClosedAt != nil {
				closedAt = *ghIssue.ClosedAt
			} else if ghIssue.State == "closed" {
				closedAt = ghIssue.UpdatedAt
			}

			issueRecord := &store.Issue{
				ID:           idgen.NewIssue(),
				RepoID:       repoRecord.ID,
				Number:       ghIssue.Number,
				State:        ghIssue.State,
				Title:        ghIssue.Title,
				Body:         ghIssue.Body,
				AuthorUserID: authorUserID,
				CreatedAt:    ghIssue.CreatedAt,
				UpdatedAt:    ghIssue.UpdatedAt,
				ClosedAt:     closedAt,
			}

			if err := db.CreateIssue(ctx, issueRecord); err != nil {
				log.Printf("migrate: create issue #%d failed: %v", ghIssue.Number, err)
			}

			// Pre-populate labels for this issue into namespace
			for _, lbl := range ghIssue.Labels {
				_, _ = db.GetOrCreateLabel(ctx, idgen.NewLabel(), ns.ID, lbl.Name, lbl.Color, "")
			}

			// Fetch and import comments for this issue
			comments, err := ghClient.ListIssueComments(ghOwner, ghRepo, ghIssue.Number)
			if err == nil {
				for _, c := range comments {
					var commentAuthorID string
					if c.User.Login != "" {
						if u, err := db.GetUserByLogin(ctx, c.User.Login); err == nil {
							commentAuthorID = u.ID
						}
					}
					commentRecord := &store.IssueComment{
						ID:           idgen.NewIssueComment(),
						RepoID:       repoRecord.ID,
						IssueID:      issueRecord.ID,
						AuthorUserID: commentAuthorID,
						Body:         c.Body,
						CreatedAt:    c.CreatedAt,
						UpdatedAt:    c.CreatedAt,
					}
					_ = db.CreateIssueComment(ctx, commentRecord)
				}
			}

			if totalIssues > 0 {
				status.Progress = 40 + int(float64(idx+1)/float64(totalIssues)*25)
				setMigrationJob(status)
			}
		}
	}

	// 5. If ImportPRs=true, fetch and create PRs
	if req.ImportPRs {
		status.Phase = PhaseImportingPRs
		if status.Progress < 65 {
			status.Progress = 65
		}
		setMigrationJob(status)

		prs, err := ghClient.ListPullRequests(ghOwner, ghRepo)
		if err != nil {
			status.Phase = PhaseFailed
			status.Error = fmt.Sprintf("failed to fetch PRs: %v", err)
			setMigrationJob(status)
			return
		}

		totalPRs := len(prs)
		for idx, ghPR := range prs {
			var authorUserID string
			if ghPR.User.Login != "" {
				if u, err := db.GetUserByLogin(ctx, ghPR.User.Login); err == nil {
					authorUserID = u.ID
				}
			}

			closedAt := ""
			if ghPR.ClosedAt != nil {
				closedAt = *ghPR.ClosedAt
			}
			mergedAt := ""
			if ghPR.MergedAt != nil {
				mergedAt = *ghPR.MergedAt
			}

			prRecord := &store.PullRequest{
				ID:           idgen.NewPR(),
				RepoID:       repoRecord.ID,
				Number:       ghPR.Number,
				State:        ghPR.State,
				Draft:        false,
				Merged:       ghPR.Merged || mergedAt != "",
				Title:        ghPR.Title,
				Body:         ghPR.Body,
				HeadRef:      ghPR.Head.Ref,
				HeadSHA:      ghPR.Head.SHA,
				BaseRef:      ghPR.Base.Ref,
				BaseSHA:      ghPR.Base.SHA,
				AuthorUserID: authorUserID,
				Mergeable:    "mergeable",
				CreatedAt:    ghPR.CreatedAt,
				UpdatedAt:    ghPR.UpdatedAt,
				ClosedAt:     closedAt,
				MergedAt:     mergedAt,
			}

			if err := db.CreatePullRequest(ctx, prRecord); err != nil {
				log.Printf("migrate: create PR #%d failed: %v", ghPR.Number, err)
			}

			// Fetch PR review comments
			comments, err := ghClient.ListPRComments(ghOwner, ghRepo, ghPR.Number)
			if err == nil {
				for _, c := range comments {
					var commentAuthorID string
					if c.User.Login != "" {
						if u, err := db.GetUserByLogin(ctx, c.User.Login); err == nil {
							commentAuthorID = u.ID
						}
					}
					commentRecord := store.PRComment{
						ID:            idgen.NewPRComment(),
						RepoID:        repoRecord.ID,
						PullRequestID: prRecord.ID,
						AuthorUserID:  commentAuthorID,
						Body:          c.Body,
						CreatedAt:     c.CreatedAt,
						UpdatedAt:     c.CreatedAt,
					}
					_ = db.CreateComment(ctx, commentRecord)
				}
			}

			if totalPRs > 0 {
				status.Progress = 65 + int(float64(idx+1)/float64(totalPRs)*30)
				setMigrationJob(status)
			}
		}
	}

	// 6. Track progress in an in-memory map -> complete
	status.Phase = PhaseComplete
	status.Progress = 100
	status.Error = ""
	setMigrationJob(status)
}
