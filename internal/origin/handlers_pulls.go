package origin

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"exist.io/internal/agentreviewer"
	"exist.io/internal/auth"
	"exist.io/internal/engine"
	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"exist.io/pkg/scopes"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// Scope requirements for pull requests.
var (
	reqPullsRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryPullRequestsRead},
	}
	reqPullsWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryPullRequestsWrite},
	}
	reqPRContentsWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryContentsWrite},
	}
)

func (s *Server) registerPulls(mux *http.ServeMux) {
	s.registerPRSocial(mux)

	// Register operation scopes
	s.register("OriginService_ListPullRequests", reqPullsRead)
	s.register("OriginService_CreatePullRequest", reqPullsWrite)
	s.register("OriginService_GetPullRequest", reqPullsRead)
	s.register("OriginService_UpdatePullRequest", reqPullsWrite)
	s.register("OriginService_MergePullRequest", reqPRContentsWrite)
	s.register("OriginService_GetPullRequestMergeability", reqPullsRead)
	s.register("OriginService_PreparePullRequestMergeRef", reqPullsRead)
	s.register("OriginService_EnablePullRequestAutoMerge", reqPRContentsWrite)
	s.register("OriginService_DisablePullRequestAutoMerge", reqPullsWrite)
	s.register("OriginService_ListPullRequestFiles", reqPullsRead)
	s.register("OriginService_ListPullRequestCommits", reqPullsRead)
	s.register("OriginService_CreateAgentReview", reqPullsWrite)

	// Routes
	dispatcher := s.pullsDispatcher()
	base := "/v1/origin/repos/{ownerSlug}/{repoName}/pulls"
	mux.Handle(base, dispatcher)
	mux.Handle(base+"/{rest...}", dispatcher)
}

func (s *Server) pullsDispatcher() http.Handler {
	// Pre-create PR guarded handlers
	hListPRs := s.guard("OriginService_ListPullRequests", s.listPullRequests)
	hCreatePR := s.guard("OriginService_CreatePullRequest", s.createPullRequest)
	hGetPR := s.guard("OriginService_GetPullRequest", s.getPullRequest)
	hUpdatePR := s.guard("OriginService_UpdatePullRequest", s.updatePullRequest)
	hMergePR := s.guard("OriginService_MergePullRequest", s.mergePullRequest)
	hMergeability := s.guard("OriginService_GetPullRequestMergeability", s.getPullRequestMergeability)
	hMergeRef := s.guard("OriginService_PreparePullRequestMergeRef", s.preparePullRequestMergeRef)
	hEnableAutoMerge := s.guard("OriginService_EnablePullRequestAutoMerge", s.enablePullRequestAutoMerge)
	hDisableAutoMerge := s.guard("OriginService_DisablePullRequestAutoMerge", s.disablePullRequestAutoMerge)
	hListFiles := s.guard("OriginService_ListPullRequestFiles", s.listPullRequestFiles)
	hListCommits := s.guard("OriginService_ListPullRequestCommits", s.listPullRequestCommits)

	// Pre-create social guarded handlers
	hCreateComment := s.guard("OriginService_CreatePullRequestComment", s.createPullRequestComment)
	hGetComment := s.guard("OriginService_GetPullRequestComment", s.getPullRequestComment)
	hListComments := s.guard("OriginService_ListPullRequestComments", s.listPullRequestComments)
	hUpdateComment := s.guard("OriginService_UpdatePullRequestComment", s.updatePullRequestComment)
	hDeleteComment := s.guard("OriginService_DeletePullRequestComment", s.deletePullRequestComment)
	hUpdateThread := s.guard("OriginService_UpdatePullRequestThread", s.updatePullRequestThread)
	hAddReaction := s.guard("OriginService_AddPullRequestCommentReaction", s.addPullRequestCommentReaction)
	hRemoveReaction := s.guard("OriginService_RemovePullRequestCommentReaction", s.removePullRequestCommentReaction)

	hCreateReview := s.guard("OriginService_CreatePullRequestReview", s.createPullRequestReview)
	hAgentReview := s.guard("OriginService_CreateAgentReview", s.triggerAgentReview)
	hListReviews := s.guard("OriginService_ListPullRequestReviews", s.listPullRequestReviews)
	hUpdateReview := s.guard("OriginService_UpdatePullRequestReview", s.updatePullRequestReview)
	hDismissReview := s.guard("OriginService_DismissPullRequestReview", s.dismissPullRequestReview)

	hListReviewers := s.guard("OriginService_ListPullRequestRequestedReviewers", s.listPullRequestRequestedReviewers)
	hRequestReviewers := s.guard("OriginService_RequestPullRequestReviewers", s.requestPullRequestReviewers)
	hRemoveReviewers := s.guard("OriginService_RemovePullRequestRequestedReviewers", s.removePullRequestRequestedReviewers)

	hListLabels := s.guard("OriginService_ListPullRequestLabels", s.listPullRequestLabels)
	hSetLabels := s.guard("OriginService_SetPullRequestLabels", s.setPullRequestLabels)
	hAddLabels := s.guard("OriginService_AddPullRequestLabels", s.addPullRequestLabels)
	hRemoveLabel := s.guard("OriginService_RemovePullRequestLabel", s.removePullRequestLabel)
	hRemoveAllLabels := s.guard("OriginService_RemoveAllPullRequestLabels", s.removeAllPullRequestLabels)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rest := strings.Trim(r.PathValue("rest"), "/")
		method := r.Method

		// 1. Root: /pulls
		if rest == "" {
			switch method {
			case http.MethodGet:
				hListPRs.ServeHTTP(w, r)
			case http.MethodPost:
				hCreatePR.ServeHTTP(w, r)
			default:
				notFound(w, "method not allowed")
			}
			return
		}

		// 2. Comments: comments/{commentId}...
		if strings.HasPrefix(rest, "comments/") {
			parts := strings.Split(rest, "/")
			if len(parts) == 2 {
				r.SetPathValue("commentId", parts[1])
				switch method {
				case http.MethodGet:
					hGetComment.ServeHTTP(w, r)
				case http.MethodPatch:
					hUpdateComment.ServeHTTP(w, r)
				case http.MethodDelete:
					hDeleteComment.ServeHTTP(w, r)
				default:
					notFound(w, "method not allowed")
				}
				return
			}
			if len(parts) == 3 && parts[2] == "reactions" && method == http.MethodPost {
				r.SetPathValue("commentId", parts[1])
				hAddReaction.ServeHTTP(w, r)
				return
			}
			if len(parts) == 4 && parts[2] == "reactions" && method == http.MethodDelete {
				r.SetPathValue("commentId", parts[1])
				r.SetPathValue("emoji", parts[3])
				hRemoveReaction.ServeHTTP(w, r)
				return
			}
			notFound(w, "not found")
			return
		}

		// 3. Threads: threads/{threadId}
		if strings.HasPrefix(rest, "threads/") {
			parts := strings.Split(rest, "/")
			if len(parts) == 2 && method == http.MethodPatch {
				r.SetPathValue("threadId", parts[1])
				hUpdateThread.ServeHTTP(w, r)
				return
			}
			notFound(w, "not found")
			return
		}

		// 4. PR-specific: {pullNumber}/...
		parts := strings.Split(rest, "/")
		r.SetPathValue("pullNumber", parts[0])

		if len(parts) == 1 {
			switch method {
			case http.MethodGet:
				hGetPR.ServeHTTP(w, r)
			case http.MethodPatch:
				hUpdatePR.ServeHTTP(w, r)
			default:
				notFound(w, "method not allowed")
			}
			return
		}

		if len(parts) == 2 {
			switch parts[1] {
			case "merge":
				if method == http.MethodPost {
					hMergePR.ServeHTTP(w, r)
					return
				}
			case "mergeability":
				if method == http.MethodGet {
					hMergeability.ServeHTTP(w, r)
					return
				}
			case "merge_ref":
				if method == http.MethodPost {
					hMergeRef.ServeHTTP(w, r)
					return
				}
			case "auto_merge":
				switch method {
				case http.MethodPut:
					hEnableAutoMerge.ServeHTTP(w, r)
					return
				case http.MethodDelete:
					hDisableAutoMerge.ServeHTTP(w, r)
					return
				}
			case "files":
				if method == http.MethodGet {
					hListFiles.ServeHTTP(w, r)
					return
				}
			case "commits":
				if method == http.MethodGet {
					hListCommits.ServeHTTP(w, r)
					return
				}
			case "comments":
				switch method {
				case http.MethodGet:
					hListComments.ServeHTTP(w, r)
					return
				case http.MethodPost:
					hCreateComment.ServeHTTP(w, r)
					return
				}
			case "reviews":
				switch method {
				case http.MethodGet:
					hListReviews.ServeHTTP(w, r)
					return
				case http.MethodPost:
					hCreateReview.ServeHTTP(w, r)
					return
				}
			case "requested_reviewers":
				switch method {
				case http.MethodGet:
					hListReviewers.ServeHTTP(w, r)
					return
				case http.MethodPost:
					hRequestReviewers.ServeHTTP(w, r)
					return
				case http.MethodDelete:
					hRemoveReviewers.ServeHTTP(w, r)
					return
				}
			case "labels":
				switch method {
				case http.MethodGet:
					hListLabels.ServeHTTP(w, r)
					return
				case http.MethodPut:
					hSetLabels.ServeHTTP(w, r)
					return
				case http.MethodPost:
					hAddLabels.ServeHTTP(w, r)
					return
				case http.MethodDelete:
					hRemoveAllLabels.ServeHTTP(w, r)
					return
				}
			}
			notFound(w, "not found")
			return
		}

		if len(parts) == 3 {
			switch parts[1] {
			case "reviews":
				if parts[2] == "agent" && method == http.MethodPost {
					hAgentReview.ServeHTTP(w, r)
					return
				}
				if method == http.MethodPatch {
					r.SetPathValue("reviewId", parts[2])
					hUpdateReview.ServeHTTP(w, r)
					return
				}
			case "labels":
				if method == http.MethodDelete {
					r.SetPathValue("labelName", parts[2])
					hRemoveLabel.ServeHTTP(w, r)
					return
				}
			}
			notFound(w, "not found")
			return
		}

		if len(parts) == 4 && parts[1] == "reviews" && parts[3] == "dismissals" && method == http.MethodPut {
			r.SetPathValue("reviewId", parts[2])
			hDismissReview.ServeHTTP(w, r)
			return
		}

		notFound(w, "not found")
	})
}

// Wire DTOs

type prRefDTO struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

type prActorDTO struct {
	User *prUserActorDTO `json:"user,omitempty"`
}

type prUserActorDTO struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
}

type prVersionDTO struct {
	Number    string `json:"number"`
	HeadSHA   string `json:"headSha"`
	BaseSHA   string `json:"baseSha"`
	CreatedAt string `json:"createdAt"`
}

type prAutoMergeDTO struct {
	EnabledAt     string `json:"enabledAt"`
	MergeMethod   string `json:"mergeMethod"`
	CommitTitle   string `json:"commitTitle,omitempty"`
	CommitMessage string `json:"commitMessage,omitempty"`
}

type pullRequestDTO struct {
	ID             string          `json:"id"`
	Number         string          `json:"number"`
	State          string          `json:"state"`
	Draft          bool            `json:"draft"`
	Merged         bool            `json:"merged"`
	Title          string          `json:"title"`
	Body           string          `json:"body"`
	Head           prRefDTO        `json:"head"`
	Base           prRefDTO        `json:"base"`
	Author         *prActorDTO     `json:"author,omitempty"`
	CreatedAt      string          `json:"createdAt"`
	UpdatedAt      string          `json:"updatedAt"`
	ClosedAt       string          `json:"closedAt,omitempty"`
	MergedAt       string          `json:"mergedAt,omitempty"`
	MergeCommitSHA string          `json:"mergeCommitSha,omitempty"`
	Additions      int             `json:"additions"`
	Deletions      int             `json:"deletions"`
	ChangedFiles   int             `json:"changedFiles"`
	WebURL         string          `json:"webUrl"`
	Version        prVersionDTO    `json:"version"`
	AutoMerge      *prAutoMergeDTO `json:"autoMerge,omitempty"`
	Labels         []labelDTO      `json:"labels"`
}

type labelDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color,omitempty"`
	Description string `json:"description,omitempty"`
}

func (s *Server) toPullRequestDTO(pr store.PullRequest, ownerSlug, repoName string, am *store.PRAutoMerge, labels []store.Label) pullRequestDTO {
	numStr := strconv.Itoa(pr.Number)
	dto := pullRequestDTO{
		ID:             pr.ID,
		Number:         numStr,
		State:          pr.State,
		Draft:          pr.Draft,
		Merged:         pr.Merged,
		Title:          pr.Title,
		Body:           pr.Body,
		Head:           prRefDTO{Ref: pr.HeadRef, SHA: pr.HeadSHA},
		Base:           prRefDTO{Ref: pr.BaseRef, SHA: pr.BaseSHA},
		CreatedAt:      pr.CreatedAt,
		UpdatedAt:      pr.UpdatedAt,
		ClosedAt:       pr.ClosedAt,
		MergedAt:       pr.MergedAt,
		MergeCommitSHA: pr.MergeCommitSHA,
		Additions:      pr.Additions,
		Deletions:      pr.Deletions,
		ChangedFiles:   pr.ChangedFiles,
		WebURL:         s.webURL(ownerSlug, repoName) + "/pulls/" + numStr,
		Version: prVersionDTO{
			Number:    "1",
			HeadSHA:   pr.HeadSHA,
			BaseSHA:   pr.BaseSHA,
			CreatedAt: pr.CreatedAt,
		},
		Labels: make([]labelDTO, 0, len(labels)),
	}

	if pr.AuthorUserID != "" {
		dto.Author = &prActorDTO{User: &prUserActorDTO{ID: pr.AuthorUserID}}
	}

	if am != nil {
		dto.AutoMerge = &prAutoMergeDTO{
			EnabledAt:     am.EnabledAt,
			MergeMethod:   am.MergeMethod,
			CommitTitle:   am.CommitTitle,
			CommitMessage: am.CommitMessage,
		}
	}

	for _, l := range labels {
		dto.Labels = append(dto.Labels, labelDTO{
			ID:          l.ID,
			Name:        l.Name,
			Color:       l.Color,
			Description: l.Description,
		})
	}

	return dto
}

// Handlers

func (s *Server) listPullRequests(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	state := r.URL.Query().Get("state")
	pageSize := parsePageSize(r)
	prs, nextToken, err := s.store.ListPullRequests(r.Context(), repo.ID, state, pageSize, r.URL.Query().Get("pageToken"))
	if err != nil {
		internalError(w, "list pull requests: "+err.Error())
		return
	}

	dtos := make([]pullRequestDTO, 0, len(prs))
	for _, pr := range prs {
		am, _ := s.store.GetAutoMerge(r.Context(), pr.ID)
		labels, _ := s.store.GetPRLabels(r.Context(), pr.ID)
		dtos = append(dtos, s.toPullRequestDTO(pr, owner, name, am, labels))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"pullRequests":  dtos,
		"nextPageToken": nextToken,
	})
}

type createPRRequest struct {
	Title             string `json:"title"`
	Body              string `json:"body"`
	Head              string `json:"head"`
	Base              string `json:"base"`
	Draft             bool   `json:"draft"`
	ParentPullRequest *struct {
		Number string `json:"number"`
		ID     string `json:"id"`
	} `json:"parentPullRequest"`
}

func (s *Server) createPullRequest(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	var req createPRRequest
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Title == "" || req.Head == "" || req.Base == "" {
		badRequest(w, "title, head, and base are required")
		return
	}

	headSHA, err := s.git.ResolveRef(owner, name, req.Head)
	if err != nil {
		badRequest(w, "head ref not found: "+req.Head)
		return
	}
	baseSHA, err := s.git.ResolveRef(owner, name, req.Base)
	if err != nil {
		badRequest(w, "base ref not found: "+req.Base)
		return
	}

	// Calculate initial diff stats
	var additions, deletions, changedFiles int
	diff, err := s.git.DiffCommits(owner, name, baseSHA.String(), headSHA.String())
	if err == nil && diff != nil {
		additions = diff.Additions
		deletions = diff.Deletions
		changedFiles = diff.ChangedFiles
	}

	// Check mergeability
	mergeable := "mergeable"
	canMerge, _, _, _ := s.git.CheckMergeable(owner, name, req.Base, req.Head)
	if !canMerge {
		mergeable = "blocked"
	}

	var parentID string
	if req.ParentPullRequest != nil {
		if req.ParentPullRequest.ID != "" {
			parentID = req.ParentPullRequest.ID
		} else if req.ParentPullRequest.Number != "" {
			if n, err := strconv.Atoi(req.ParentPullRequest.Number); err == nil {
				if parentPR, err := s.store.GetPullRequest(r.Context(), repo.ID, n); err == nil {
					parentID = parentPR.ID
				}
			}
		}
	}

	authorID := ""
	if tok, ok := auth.ContextToken(r.Context()); ok && tok.Type() == scopes.TypeUser {
		authorID = tok.Subject()
	}

	pr := store.PullRequest{
		ID:           idgen.NewPR(),
		RepoID:       repo.ID,
		State:        "open",
		Draft:        req.Draft,
		Title:        req.Title,
		Body:         req.Body,
		HeadRef:      req.Head,
		HeadSHA:      headSHA.String(),
		BaseRef:      req.Base,
		BaseSHA:      baseSHA.String(),
		AuthorUserID: authorID,
		Mergeable:    mergeable,
		Additions:    additions,
		Deletions:    deletions,
		ChangedFiles: changedFiles,
		ParentPRID:   parentID,
	}

	if err := s.store.CreatePullRequest(r.Context(), &pr); err != nil {
		internalError(w, "create pull request: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, s.toPullRequestDTO(pr, owner, name, nil, nil))
}

func (s *Server) getPullRequest(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	num, err := strconv.Atoi(r.PathValue("pullNumber"))
	if err != nil {
		badRequest(w, "invalid pull number")
		return
	}

	pr, err := s.store.GetPullRequest(r.Context(), repo.ID, num)
	if err != nil {
		notFound(w, "pull request not found")
		return
	}

	am, _ := s.store.GetAutoMerge(r.Context(), pr.ID)
	labels, _ := s.store.GetPRLabels(r.Context(), pr.ID)
	writeJSON(w, http.StatusOK, s.toPullRequestDTO(pr, owner, name, am, labels))
}

type updatePRRequest struct {
	Title             *string `json:"title"`
	Body              *string `json:"body"`
	State             *string `json:"state"`
	Draft             *bool   `json:"draft"`
	Base              *string `json:"base"`
	ParentPullRequest *struct {
		Number string `json:"number"`
		ID     string `json:"id"`
		Clear  bool   `json:"clear"`
	} `json:"parentPullRequest"`
}

func (s *Server) updatePullRequest(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	num, err := strconv.Atoi(r.PathValue("pullNumber"))
	if err != nil {
		badRequest(w, "invalid pull number")
		return
	}

	pr, err := s.store.GetPullRequest(r.Context(), repo.ID, num)
	if err != nil {
		notFound(w, "pull request not found")
		return
	}

	var req updatePRRequest
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	u := store.PullRequestUpdate{
		Title: req.Title,
		Body:  req.Body,
		State: req.State,
		Draft: req.Draft,
	}

	if req.Base != nil && *req.Base != "" && *req.Base != pr.BaseRef {
		baseSHA, err := s.git.ResolveRef(owner, name, *req.Base)
		if err != nil {
			badRequest(w, "new base ref not found: "+*req.Base)
			return
		}
		u.BaseRef = req.Base
		bsha := baseSHA.String()
		u.BaseSHA = &bsha

		// Recompute diff stats with new base
		diff, err := s.git.DiffCommits(owner, name, bsha, pr.HeadSHA)
		if err == nil && diff != nil {
			u.Additions = &diff.Additions
			u.Deletions = &diff.Deletions
			u.ChangedFiles = &diff.ChangedFiles
		}
	}

	if req.ParentPullRequest != nil {
		if req.ParentPullRequest.Clear {
			empty := ""
			u.ParentPRID = &empty
		} else if req.ParentPullRequest.ID != "" {
			u.ParentPRID = &req.ParentPullRequest.ID
		} else if req.ParentPullRequest.Number != "" {
			if n, err := strconv.Atoi(req.ParentPullRequest.Number); err == nil {
				if parentPR, err := s.store.GetPullRequest(r.Context(), repo.ID, n); err == nil {
					u.ParentPRID = &parentPR.ID
				}
			}
		}
	}

	if req.State != nil && *req.State == "closed" {
		t := time.Now().UTC().Format(time.RFC3339Nano)
		u.ClosedAt = &t
	}

	if err := s.store.UpdatePullRequest(r.Context(), pr.ID, u); err != nil {
		internalError(w, "update pull request: "+err.Error())
		return
	}

	updatedPR, err := s.store.GetPullRequest(r.Context(), repo.ID, num)
	if err != nil {
		internalError(w, "reload pull request: "+err.Error())
		return
	}

	am, _ := s.store.GetAutoMerge(r.Context(), pr.ID)
	labels, _ := s.store.GetPRLabels(r.Context(), pr.ID)
	writeJSON(w, http.StatusOK, s.toPullRequestDTO(updatedPR, owner, name, am, labels))
}

type mergePRRequest struct {
	ExpectedHeadSHA string `json:"expectedHeadSha"`
	MergeMethod     string `json:"mergeMethod"` // merge | squash
}

func (s *Server) mergePullRequest(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	num, err := strconv.Atoi(r.PathValue("pullNumber"))
	if err != nil {
		badRequest(w, "invalid pull number")
		return
	}

	pr, err := s.store.GetPullRequest(r.Context(), repo.ID, num)
	if err != nil {
		notFound(w, "pull request not found")
		return
	}

	var req mergePRRequest
	_ = decodeJSON(r, &req)

	// Idempotent success if already merged
	if pr.Merged {
		am, _ := s.store.GetAutoMerge(r.Context(), pr.ID)
		labels, _ := s.store.GetPRLabels(r.Context(), pr.ID)
		writeJSON(w, http.StatusOK, map[string]any{
			"mergeCommitSha":    pr.MergeCommitSHA,
			"mergedPullNumbers": []string{},
			"pullRequest":       s.toPullRequestDTO(pr, owner, name, am, labels),
		})
		return
	}

	// Verify expected head SHA
	if req.ExpectedHeadSHA != "" && req.ExpectedHeadSHA != pr.HeadSHA {
		conflict(w, "head SHA changed: expected "+req.ExpectedHeadSHA+" but was "+pr.HeadSHA)
		return
	}

	// Determine merge method
	method := req.MergeMethod
	if method == "" {
		if repo.AllowMergeCommit {
			method = "merge"
		} else if repo.AllowSquashMerge {
			method = "squash"
		} else {
			method = "merge"
		}
	}
	if method == "merge" && !repo.AllowMergeCommit {
		badRequest(w, "merge commit is not allowed on this repository")
		return
	}
	if method == "squash" && !repo.AllowSquashMerge {
		badRequest(w, "squash merge is not allowed on this repository")
		return
	}

	// Evaluate rulesets before executing merge
	rulesets, _ := s.store.ListRulesets(r.Context(), repo.ID)
	reviews, _ := s.store.ListReviews(r.Context(), pr.ID)
	approvedCount := 0
	for _, rev := range reviews {
		if !rev.Dismissed && (rev.State == "approved" || rev.State == "approve") {
			approvedCount++
		}
	}
	checkRuns, _ := s.store.ListCheckRunsForCommit(r.Context(), repo.ID, pr.HeadSHA, "", "", 100, 0)
	actingUser := ""
	if tok, ok := auth.ContextToken(r.Context()); ok {
		actingUser = tok.Subject()
	}
	evalRes := engine.EvaluateMerge(r.Context(), rulesets, repo.DefaultBranch, pr.BaseRef, actingUser, "", approvedCount, checkRuns)
	if !evalRes.Allowed {
		msg := "blocked by ruleset"
		if len(evalRes.Blockers) > 0 {
			msg = evalRes.Blockers[0].Message
		}
		conflict(w, msg)
		return
	}

	authorSig := object.Signature{
		Name:  "Exist Engine",
		Email: "merge@exist.io",
		When:  time.Now(),
	}
	msg := fmt.Sprintf("Merge pull request #%d from %s\n\n%s", pr.Number, pr.HeadRef, pr.Title)

	mergeCommitSHA, err := s.git.Merge(owner, name, pr.BaseRef, pr.HeadRef, method, authorSig, msg)
	if err != nil {
		if errors.Is(err, gitstore.ErrMergeConflict) {
			conflict(w, "merge conflict: "+err.Error())
			return
		}
		internalError(w, "perform merge: "+err.Error())
		return
	}

	if err := s.store.SetPullRequestMergeResult(r.Context(), pr.ID, mergeCommitSHA); err != nil {
		internalError(w, "record merge: "+err.Error())
		return
	}

	// Reload PR
	mergedPR, _ := s.store.GetPullRequest(r.Context(), repo.ID, num)
	am, _ := s.store.GetAutoMerge(r.Context(), pr.ID)
	labels, _ := s.store.GetPRLabels(r.Context(), pr.ID)

	writeJSON(w, http.StatusOK, map[string]any{
		"mergeCommitSha":    mergeCommitSHA,
		"mergedPullNumbers": []string{strconv.Itoa(pr.Number)},
		"pullRequest":       s.toPullRequestDTO(mergedPR, owner, name, am, labels),
	})
}

func (s *Server) getPullRequestMergeability(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	num, err := strconv.Atoi(r.PathValue("pullNumber"))
	if err != nil {
		badRequest(w, "invalid pull number")
		return
	}

	pr, err := s.store.GetPullRequest(r.Context(), repo.ID, num)
	if err != nil {
		notFound(w, "pull request not found")
		return
	}

	expected := r.URL.Query().Get("expectedHeadSha")
	if expected != "" && expected != pr.HeadSHA {
		conflict(w, "evaluated head differs from expected")
		return
	}

	mergeable, _, conflicts, err := s.git.CheckMergeable(owner, name, pr.BaseRef, pr.HeadRef)
	verdict := "mergeable"
	var blockers []any

	if !mergeable {
		verdict = "conflicting"
		msg := "Merge conflict"
		if len(conflicts) > 0 {
			msg = conflicts[0]
		}
		blockers = append(blockers, map[string]any{
			"kind":    "merge_conflict",
			"message": msg,
		})
	}
	if pr.Draft {
		verdict = "blocked"
		blockers = append(blockers, map[string]any{
			"kind":    "draft",
			"message": "Pull request is a draft",
		})
	}
	if pr.State == "closed" {
		verdict = "blocked"
		blockers = append(blockers, map[string]any{
			"kind":    "closed",
			"message": "Pull request is closed",
		})
	}

	// Evaluate Rulesets
	rulesets, _ := s.store.ListRulesets(r.Context(), repo.ID)
	reviews, _ := s.store.ListReviews(r.Context(), pr.ID)
	approvedCount := 0
	for _, rev := range reviews {
		if !rev.Dismissed && (rev.State == "approved" || rev.State == "approve") {
			approvedCount++
		}
	}
	checkRuns, _ := s.store.ListCheckRunsForCommit(r.Context(), repo.ID, pr.HeadSHA, "", "", 100, 0)
	actingUser := ""
	if tok, ok := auth.ContextToken(r.Context()); ok {
		actingUser = tok.Subject()
	}
	evalRes := engine.EvaluateMerge(r.Context(), rulesets, repo.DefaultBranch, pr.BaseRef, actingUser, "", approvedCount, checkRuns)
	if !evalRes.Allowed {
		verdict = "blocked"
		for _, b := range evalRes.Blockers {
			blockers = append(blockers, b)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"verdict":  verdict,
		"blockers": blockers,
		"pullRequest": map[string]any{
			"id":     pr.ID,
			"number": strconv.Itoa(pr.Number),
		},
		"headSha":     pr.HeadSHA,
		"baseRef":     pr.BaseRef,
		"baseSha":     pr.BaseSHA,
		"evaluatedAt": time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func (s *Server) preparePullRequestMergeRef(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	num, err := strconv.Atoi(r.PathValue("pullNumber"))
	if err != nil {
		badRequest(w, "invalid pull number")
		return
	}

	pr, err := s.store.GetPullRequest(r.Context(), repo.ID, num)
	if err != nil {
		notFound(w, "pull request not found")
		return
	}

	authorSig := object.Signature{
		Name:  "Exist Engine",
		Email: "merge-preview@exist.io",
		When:  time.Now(),
	}

	previewSHA, err := s.git.PrepareMergeRef(owner, name, pr.BaseRef, pr.HeadRef, pr.Number, authorSig)
	if err != nil {
		if errors.Is(err, gitstore.ErrMergeConflict) {
			conflict(w, "merge conflict: "+err.Error())
			return
		}
		internalError(w, "prepare merge ref: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"mergeCommitSha": previewSHA})
}

func (s *Server) enablePullRequestAutoMerge(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	num, err := strconv.Atoi(r.PathValue("pullNumber"))
	if err != nil {
		badRequest(w, "invalid pull number")
		return
	}

	pr, err := s.store.GetPullRequest(r.Context(), repo.ID, num)
	if err != nil {
		notFound(w, "pull request not found")
		return
	}

	var req struct {
		MergeMethod   string `json:"mergeMethod"`
		CommitTitle   string `json:"commitTitle"`
		CommitMessage string `json:"commitMessage"`
	}
	_ = decodeJSON(r, &req)
	method := req.MergeMethod
	if method == "" {
		method = "merge"
	}

	userId := ""
	if tok, ok := auth.ContextToken(r.Context()); ok && tok.Type() == scopes.TypeUser {
		userId = tok.Subject()
	}

	am := store.PRAutoMerge{
		PullRequestID:   pr.ID,
		EnabledByUserID: userId,
		MergeMethod:     method,
		CommitTitle:     req.CommitTitle,
		CommitMessage:   req.CommitMessage,
	}
	if err := s.store.SetAutoMerge(r.Context(), am); err != nil {
		internalError(w, "enable auto-merge: "+err.Error())
		return
	}

	labels, _ := s.store.GetPRLabels(r.Context(), pr.ID)
	writeJSON(w, http.StatusOK, s.toPullRequestDTO(pr, owner, name, &am, labels))
}

func (s *Server) disablePullRequestAutoMerge(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	num, err := strconv.Atoi(r.PathValue("pullNumber"))
	if err != nil {
		badRequest(w, "invalid pull number")
		return
	}

	pr, err := s.store.GetPullRequest(r.Context(), repo.ID, num)
	if err != nil {
		notFound(w, "pull request not found")
		return
	}

	if err := s.store.DeleteAutoMerge(r.Context(), pr.ID); err != nil {
		internalError(w, "disable auto-merge: "+err.Error())
		return
	}

	labels, _ := s.store.GetPRLabels(r.Context(), pr.ID)
	writeJSON(w, http.StatusOK, s.toPullRequestDTO(pr, owner, name, nil, labels))
}

func (s *Server) listPullRequestFiles(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	num, err := strconv.Atoi(r.PathValue("pullNumber"))
	if err != nil {
		badRequest(w, "invalid pull number")
		return
	}

	pr, err := s.store.GetPullRequest(r.Context(), repo.ID, num)
	if err != nil {
		notFound(w, "pull request not found")
		return
	}

	diff, err := s.git.DiffCommits(owner, name, pr.BaseSHA, pr.HeadSHA)
	if err != nil {
		internalError(w, "diff files: "+err.Error())
		return
	}

	files := []gitstore.PRFileChange{}
	if diff != nil && diff.Files != nil {
		files = diff.Files
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"files":         files,
		"nextPageToken": "",
	})
}

func (s *Server) listPullRequestCommits(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	num, err := strconv.Atoi(r.PathValue("pullNumber"))
	if err != nil {
		badRequest(w, "invalid pull number")
		return
	}

	pr, err := s.store.GetPullRequest(r.Context(), repo.ID, num)
	if err != nil {
		notFound(w, "pull request not found")
		return
	}

	pageSize := parsePageSize(r)
	commits, err := s.git.ListPRCommits(owner, name, pr.BaseSHA, pr.HeadSHA, pageSize)
	if err != nil {
		internalError(w, "list commits: "+err.Error())
		return
	}

	type commitEntry struct {
		SHA     string `json:"sha"`
		Commit  any    `json:"commit"`
		Parents any    `json:"parents"`
	}
	out := make([]commitEntry, 0, len(commits))
	for _, c := range commits {
		parents := make([]map[string]string, 0, len(c.Parents))
		for _, p := range c.Parents {
			parents = append(parents, map[string]string{"sha": p})
		}
		out = append(out, commitEntry{
			SHA: c.SHA,
			Commit: map[string]any{
				"author":    map[string]string{"name": c.AuthorName, "email": c.AuthorEmail, "date": c.AuthorDate},
				"committer": map[string]string{"name": c.CommitterName, "email": c.CommitterEmail, "date": c.CommitterDate},
				"message":   c.Message,
				"tree":      map[string]string{"sha": c.TreeSHA},
			},
			Parents: parents,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"commits":       out,
		"nextPageToken": "",
	})
}

func (s *Server) triggerAgentReview(w http.ResponseWriter, r *http.Request) {
	owner, name, ok := s.openRepo(w, r)
	if !ok {
		return
	}
	repo, err := s.store.GetRepo(r.Context(), owner, name)
	if err != nil {
		notFound(w, "repository not found")
		return
	}

	num, err := strconv.Atoi(r.PathValue("pullNumber"))
	if err != nil {
		badRequest(w, "invalid pull number")
		return
	}

	pr, err := s.store.GetPullRequest(r.Context(), repo.ID, num)
	if err != nil {
		notFound(w, "pull request not found")
		return
	}

	res, err := agentreviewer.ReviewPullRequest(r.Context(), s.store, s.git, owner, repo, pr)
	if err != nil {
		internalError(w, "agent review failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

