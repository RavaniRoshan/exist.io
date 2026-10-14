package origin

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"exist.io/internal/auth"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"exist.io/pkg/scopes"
)

// Scope requirements for PR discussions, reviews, and labels.
var (
	reqReviewsRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryPullRequestsReviewsRead},
	}
	reqReviewsWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser},
		Scopes:     []string{scopes.RepositoryPullRequestsReviewsWrite},
	}
)

func (s *Server) registerPRSocial(mux *http.ServeMux) {
	// Comments & Threads
	s.register("OriginService_CreatePullRequestComment", reqReviewsWrite)
	s.register("OriginService_GetPullRequestComment", reqReviewsRead)
	s.register("OriginService_ListPullRequestComments", reqReviewsRead)
	s.register("OriginService_UpdatePullRequestComment", reqReviewsWrite)
	s.register("OriginService_DeletePullRequestComment", reqReviewsWrite)
	s.register("OriginService_UpdatePullRequestThread", reqReviewsWrite)
	s.register("OriginService_AddPullRequestCommentReaction", reqReviewsWrite)
	s.register("OriginService_RemovePullRequestCommentReaction", reqReviewsWrite)

	// Reviews
	s.register("OriginService_CreatePullRequestReview", reqReviewsWrite)
	s.register("OriginService_UpdatePullRequestReview", reqReviewsWrite)
	s.register("OriginService_DismissPullRequestReview", reqReviewsWrite)
	s.register("OriginService_ListPullRequestReviews", reqReviewsRead)

	// Reviewers
	s.register("OriginService_ListPullRequestRequestedReviewers", reqReviewsRead)
	s.register("OriginService_RequestPullRequestReviewers", reqReviewsWrite)
	s.register("OriginService_RemovePullRequestRequestedReviewers", reqReviewsWrite)

	// Labels
	s.register("OriginService_ListPullRequestLabels", reqPullsRead)
	s.register("OriginService_SetPullRequestLabels", reqPullsWrite)
	s.register("OriginService_AddPullRequestLabels", reqPullsWrite)
	s.register("OriginService_RemovePullRequestLabel", reqPullsWrite)
	s.register("OriginService_RemoveAllPullRequestLabels", reqPullsWrite)
}

// Wire DTOs for PR Social

type commentThreadDTO struct {
	ID         string            `json:"id"`
	Version    versionRefDTO     `json:"version"`
	Path       string            `json:"path,omitempty"`
	Side       string            `json:"side,omitempty"`
	StartLine  int               `json:"startLine,omitempty"`
	EndLine    int               `json:"endLine,omitempty"`
	ResolvedAt string            `json:"resolvedAt,omitempty"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
}

type versionRefDTO struct {
	Number  string `json:"number"`
	HeadSHA string `json:"headSha"`
	BaseSHA string `json:"baseSha"`
}

type commentReactionDTO struct {
	Emoji     string      `json:"emoji"`
	Reactor   *prActorDTO `json:"reactor,omitempty"`
	CreatedAt string      `json:"createdAt"`
}

type pullRequestCommentDTO struct {
	ID        string               `json:"id"`
	Thread    commentThreadDTO     `json:"thread"`
	Body      string               `json:"body"`
	Author    *prActorDTO          `json:"author,omitempty"`
	CreatedAt string               `json:"createdAt"`
	UpdatedAt string               `json:"updatedAt"`
	Reactions []commentReactionDTO `json:"reactions"`
}

type pullRequestReviewDTO struct {
	ID                 string             `json:"id"`
	Author             *prActorDTO        `json:"author,omitempty"`
	Verdict            string             `json:"verdict"`
	Body               string             `json:"body"`
	SubmittedAt        string             `json:"submittedAt"`
	PullRequestVersion versionRefDTO      `json:"pullRequestVersion"`
	Dismissal          *reviewDismissalDTO `json:"dismissal,omitempty"`
}

type reviewDismissalDTO struct {
	Message     string      `json:"message"`
	DismissedBy *prActorDTO `json:"dismissedBy,omitempty"`
	DismissedAt string      `json:"dismissedAt"`
}

func (s *Server) toCommentDTO(r *http.Request, c store.PRComment, t store.PRThread, rxns []store.PRCommentReaction) pullRequestCommentDTO {
	dto := pullRequestCommentDTO{
		ID: c.ID,
		Thread: commentThreadDTO{
			ID: t.ID,
			Version: versionRefDTO{
				Number:  strconv.Itoa(t.VersionNumber),
				HeadSHA: t.HeadSHA,
				BaseSHA: t.BaseSHA,
			},
			Path:       t.Path,
			Side:       t.Side,
			StartLine:  t.StartLine,
			EndLine:    t.EndLine,
			ResolvedAt: t.ResolvedAt,
			CreatedAt:  t.CreatedAt,
			UpdatedAt:  t.UpdatedAt,
		},
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Reactions: make([]commentReactionDTO, 0, len(rxns)),
	}

	if c.AuthorUserID != "" {
		dto.Author = &prActorDTO{User: &prUserActorDTO{ID: c.AuthorUserID}}
	}

	for _, rxn := range rxns {
		var reactor *prActorDTO
		if rxn.UserID != "" {
			reactor = &prActorDTO{User: &prUserActorDTO{ID: rxn.UserID}}
		}
		dto.Reactions = append(dto.Reactions, commentReactionDTO{
			Emoji:     rxn.Emoji,
			Reactor:   reactor,
			CreatedAt: rxn.CreatedAt,
		})
	}

	return dto
}

// Handlers — Comments & Threads

type createCommentReq struct {
	Body          string `json:"body"`
	ThreadID      string `json:"threadId"`
	VersionNumber string `json:"versionNumber"`
	File          *struct {
		Path string `json:"path"`
		Side string `json:"side"`
	} `json:"file"`
	Line *struct {
		Path      string `json:"path"`
		Side      string `json:"side"`
		StartLine int    `json:"startLine"`
		EndLine   int    `json:"endLine"`
	} `json:"line"`
}

func (s *Server) createPullRequestComment(w http.ResponseWriter, r *http.Request) {
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

	var req createCommentReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Body) == "" {
		badRequest(w, "comment body cannot be empty")
		return
	}

	var thread store.PRThread
	if req.ThreadID != "" {
		thread, err = s.store.GetThread(r.Context(), req.ThreadID)
		if err != nil {
			badRequest(w, "thread not found: "+req.ThreadID)
			return
		}
	} else {
		// Create new thread
		threadID := idgen.NewPRThread()
		verNum := 1
		if req.VersionNumber != "" {
			if n, err := strconv.Atoi(req.VersionNumber); err == nil && n > 0 {
				verNum = n
			}
		}

		path := ""
		side := ""
		startLine := 0
		endLine := 0
		if req.Line != nil {
			path = req.Line.Path
			side = req.Line.Side
			startLine = req.Line.StartLine
			endLine = req.Line.EndLine
		} else if req.File != nil {
			path = req.File.Path
			side = req.File.Side
		}

		thread = store.PRThread{
			ID:            threadID,
			RepoID:        repo.ID,
			PullRequestID: pr.ID,
			Path:          path,
			Side:          side,
			StartLine:     startLine,
			EndLine:       endLine,
			VersionNumber: verNum,
			HeadSHA:       pr.HeadSHA,
			BaseSHA:       pr.BaseSHA,
		}
		if err := s.store.CreateThread(r.Context(), thread); err != nil {
			internalError(w, "create thread: "+err.Error())
			return
		}
	}

	authorID := ""
	if tok, ok := auth.ContextToken(r.Context()); ok && tok.Type() == scopes.TypeUser {
		authorID = tok.Subject()
	}

	comment := store.PRComment{
		ID:            idgen.NewPRComment(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		ThreadID:      thread.ID,
		AuthorUserID:  authorID,
		Body:          req.Body,
		Path:          thread.Path,
		Line:          thread.StartLine,
		Side:          thread.Side,
		CommitSHA:     thread.HeadSHA,
	}

	if err := s.store.CreateComment(r.Context(), comment); err != nil {
		internalError(w, "create comment: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, s.toCommentDTO(r, comment, thread, nil))
}

func (s *Server) getPullRequestComment(w http.ResponseWriter, r *http.Request) {
	_, _, ok := s.openRepo(w, r)
	if !ok {
		return
	}

	commentID := r.PathValue("commentId")
	comment, err := s.store.GetComment(r.Context(), commentID)
	if err != nil {
		notFound(w, "comment not found")
		return
	}

	thread, err := s.store.GetThread(r.Context(), comment.ThreadID)
	if err != nil {
		internalError(w, "load thread: "+err.Error())
		return
	}

	rxns, _ := s.store.ListReactions(r.Context(), comment.ID)
	writeJSON(w, http.StatusOK, s.toCommentDTO(r, comment, thread, rxns))
}

func (s *Server) listPullRequestComments(w http.ResponseWriter, r *http.Request) {
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

	comments, err := s.store.ListComments(r.Context(), pr.ID)
	if err != nil {
		internalError(w, "list comments: "+err.Error())
		return
	}

	dtos := make([]pullRequestCommentDTO, 0, len(comments))
	threadCache := make(map[string]store.PRThread)
	for _, c := range comments {
		thread, ok := threadCache[c.ThreadID]
		if !ok {
			thread, err = s.store.GetThread(r.Context(), c.ThreadID)
			if err == nil {
				threadCache[c.ThreadID] = thread
			}
		}
		rxns, _ := s.store.ListReactions(r.Context(), c.ID)
		dtos = append(dtos, s.toCommentDTO(r, c, thread, rxns))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"comments":      dtos,
		"nextPageToken": "",
	})
}

func (s *Server) updatePullRequestComment(w http.ResponseWriter, r *http.Request) {
	_, _, ok := s.openRepo(w, r)
	if !ok {
		return
	}

	commentID := r.PathValue("commentId")
	comment, err := s.store.GetComment(r.Context(), commentID)
	if err != nil {
		notFound(w, "comment not found")
		return
	}

	var req struct {
		Body string `json:"body"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Body) == "" {
		badRequest(w, "comment body cannot be empty")
		return
	}

	if err := s.store.UpdateComment(r.Context(), commentID, req.Body); err != nil {
		internalError(w, "update comment: "+err.Error())
		return
	}

	comment.Body = req.Body
	thread, _ := s.store.GetThread(r.Context(), comment.ThreadID)
	rxns, _ := s.store.ListReactions(r.Context(), comment.ID)
	writeJSON(w, http.StatusOK, s.toCommentDTO(r, comment, thread, rxns))
}

func (s *Server) deletePullRequestComment(w http.ResponseWriter, r *http.Request) {
	_, _, ok := s.openRepo(w, r)
	if !ok {
		return
	}

	commentID := r.PathValue("commentId")
	if err := s.store.DeleteComment(r.Context(), commentID); err != nil {
		notFound(w, "comment not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *Server) updatePullRequestThread(w http.ResponseWriter, r *http.Request) {
	_, _, ok := s.openRepo(w, r)
	if !ok {
		return
	}

	threadID := r.PathValue("threadId")
	thread, err := s.store.GetThread(r.Context(), threadID)
	if err != nil {
		notFound(w, "thread not found")
		return
	}

	var req struct {
		Resolved bool `json:"resolved"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	var resolvedAt *string
	var resolvedBy *string
	if req.Resolved {
		t := time.Now().UTC().Format(time.RFC3339Nano)
		resolvedAt = &t
		if tok, ok := auth.ContextToken(r.Context()); ok && tok.Type() == scopes.TypeUser {
			s := tok.Subject()
			resolvedBy = &s
		}
	} else {
		empty := ""
		resolvedAt = &empty
		resolvedBy = &empty
	}

	if err := s.store.UpdateThread(r.Context(), threadID, resolvedAt, resolvedBy); err != nil {
		internalError(w, "update thread: "+err.Error())
		return
	}

	thread, _ = s.store.GetThread(r.Context(), threadID)
	writeJSON(w, http.StatusOK, commentThreadDTO{
		ID: thread.ID,
		Version: versionRefDTO{
			Number:  strconv.Itoa(thread.VersionNumber),
			HeadSHA: thread.HeadSHA,
			BaseSHA: thread.BaseSHA,
		},
		Path:       thread.Path,
		Side:       thread.Side,
		StartLine:  thread.StartLine,
		EndLine:    thread.EndLine,
		ResolvedAt: thread.ResolvedAt,
		CreatedAt:  thread.CreatedAt,
		UpdatedAt:  thread.UpdatedAt,
	})
}

func (s *Server) addPullRequestCommentReaction(w http.ResponseWriter, r *http.Request) {
	_, _, ok := s.openRepo(w, r)
	if !ok {
		return
	}

	commentID := r.PathValue("commentId")
	comment, err := s.store.GetComment(r.Context(), commentID)
	if err != nil {
		notFound(w, "comment not found")
		return
	}

	var req struct {
		Emoji string `json:"emoji"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Emoji == "" {
		badRequest(w, "emoji is required")
		return
	}

	userID := ""
	if tok, ok := auth.ContextToken(r.Context()); ok && tok.Type() == scopes.TypeUser {
		userID = tok.Subject()
	}

	rxn := store.PRCommentReaction{
		ID:        idgen.NewReaction(),
		CommentID: commentID,
		UserID:    userID,
		Emoji:     req.Emoji,
	}
	_ = s.store.AddReaction(r.Context(), rxn)

	thread, _ := s.store.GetThread(r.Context(), comment.ThreadID)
	rxns, _ := s.store.ListReactions(r.Context(), comment.ID)
	writeJSON(w, http.StatusOK, s.toCommentDTO(r, comment, thread, rxns))
}

func (s *Server) removePullRequestCommentReaction(w http.ResponseWriter, r *http.Request) {
	_, _, ok := s.openRepo(w, r)
	if !ok {
		return
	}

	commentID := r.PathValue("commentId")
	emoji := r.PathValue("emoji")
	comment, err := s.store.GetComment(r.Context(), commentID)
	if err != nil {
		notFound(w, "comment not found")
		return
	}

	userID := ""
	if tok, ok := auth.ContextToken(r.Context()); ok && tok.Type() == scopes.TypeUser {
		userID = tok.Subject()
	}

	_ = s.store.RemoveReaction(r.Context(), commentID, userID, emoji)

	thread, _ := s.store.GetThread(r.Context(), comment.ThreadID)
	rxns, _ := s.store.ListReactions(r.Context(), comment.ID)
	writeJSON(w, http.StatusOK, s.toCommentDTO(r, comment, thread, rxns))
}

// Handlers — Reviews

type createReviewReq struct {
	Verdict       string `json:"verdict"` // approve | request_changes | comment
	Body          string `json:"body"`
	VersionNumber string `json:"versionNumber"`
}

func (s *Server) createPullRequestReview(w http.ResponseWriter, r *http.Request) {
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

	var req createReviewReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Verdict != "approve" && req.Verdict != "request_changes" && req.Verdict != "comment" {
		badRequest(w, "verdict must be approve, request_changes, or comment")
		return
	}

	verNum := 1
	if req.VersionNumber != "" {
		if n, err := strconv.Atoi(req.VersionNumber); err == nil && n > 0 {
			verNum = n
		}
	}

	authorID := ""
	if tok, ok := auth.ContextToken(r.Context()); ok && tok.Type() == scopes.TypeUser {
		authorID = tok.Subject()
	}

	review := store.PRReview{
		ID:            idgen.NewPRReview(),
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		AuthorUserID:  authorID,
		State:         req.Verdict,
		Body:          req.Body,
		CommitSHA:     pr.HeadSHA,
		VersionNumber: verNum,
	}

	if err := s.store.CreateReview(r.Context(), review); err != nil {
		internalError(w, "create review: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, pullRequestReviewDTO{
		ID:          review.ID,
		Author:      &prActorDTO{User: &prUserActorDTO{ID: review.AuthorUserID}},
		Verdict:     review.State,
		Body:        review.Body,
		SubmittedAt: review.CreatedAt,
		PullRequestVersion: versionRefDTO{
			Number:  strconv.Itoa(review.VersionNumber),
			HeadSHA: review.CommitSHA,
			BaseSHA: pr.BaseSHA,
		},
	})
}

func (s *Server) listPullRequestReviews(w http.ResponseWriter, r *http.Request) {
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

	reviews, err := s.store.ListReviews(r.Context(), pr.ID)
	if err != nil {
		internalError(w, "list reviews: "+err.Error())
		return
	}

	dtos := make([]pullRequestReviewDTO, 0, len(reviews))
	for _, rev := range reviews {
		dto := pullRequestReviewDTO{
			ID:          rev.ID,
			Verdict:     rev.State,
			Body:        rev.Body,
			SubmittedAt: rev.CreatedAt,
			PullRequestVersion: versionRefDTO{
				Number:  strconv.Itoa(rev.VersionNumber),
				HeadSHA: rev.CommitSHA,
				BaseSHA: pr.BaseSHA,
			},
		}
		if rev.AuthorUserID != "" {
			dto.Author = &prActorDTO{User: &prUserActorDTO{ID: rev.AuthorUserID}}
		}
		if rev.Dismissed {
			dto.Dismissal = &reviewDismissalDTO{
				Message:     rev.DismissalMessage,
				DismissedBy: &prActorDTO{User: &prUserActorDTO{ID: rev.DismissedByUserID}},
				DismissedAt: rev.DismissedAt,
			}
		}
		dtos = append(dtos, dto)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"reviews":       dtos,
		"nextPageToken": "",
	})
}

func (s *Server) updatePullRequestReview(w http.ResponseWriter, r *http.Request) {
	_, _, ok := s.openRepo(w, r)
	if !ok {
		return
	}

	reviewID := r.PathValue("reviewId")
	rev, err := s.store.GetReview(r.Context(), reviewID)
	if err != nil {
		notFound(w, "review not found")
		return
	}

	var req struct {
		Body string `json:"body"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Body) == "" {
		badRequest(w, "body must contain non-whitespace text")
		return
	}

	if err := s.store.UpdateReview(r.Context(), reviewID, req.Body); err != nil {
		internalError(w, "update review: "+err.Error())
		return
	}

	rev.Body = req.Body
	writeJSON(w, http.StatusOK, pullRequestReviewDTO{
		ID:          rev.ID,
		Author:      &prActorDTO{User: &prUserActorDTO{ID: rev.AuthorUserID}},
		Verdict:     rev.State,
		Body:        rev.Body,
		SubmittedAt: rev.CreatedAt,
		PullRequestVersion: versionRefDTO{
			Number:  strconv.Itoa(rev.VersionNumber),
			HeadSHA: rev.CommitSHA,
		},
	})
}

func (s *Server) dismissPullRequestReview(w http.ResponseWriter, r *http.Request) {
	_, _, ok := s.openRepo(w, r)
	if !ok {
		return
	}

	reviewID := r.PathValue("reviewId")
	rev, err := s.store.GetReview(r.Context(), reviewID)
	if err != nil {
		notFound(w, "review not found")
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Message) == "" {
		badRequest(w, "dismissal message cannot be empty")
		return
	}

	dismissedBy := ""
	if tok, ok := auth.ContextToken(r.Context()); ok && tok.Type() == scopes.TypeUser {
		dismissedBy = tok.Subject()
	}

	if err := s.store.DismissReview(r.Context(), reviewID, req.Message, dismissedBy); err != nil {
		internalError(w, "dismiss review: "+err.Error())
		return
	}

	rev, _ = s.store.GetReview(r.Context(), reviewID)
	writeJSON(w, http.StatusOK, pullRequestReviewDTO{
		ID:          rev.ID,
		Author:      &prActorDTO{User: &prUserActorDTO{ID: rev.AuthorUserID}},
		Verdict:     rev.State,
		Body:        rev.Body,
		SubmittedAt: rev.CreatedAt,
		PullRequestVersion: versionRefDTO{
			Number:  strconv.Itoa(rev.VersionNumber),
			HeadSHA: rev.CommitSHA,
		},
		Dismissal: &reviewDismissalDTO{
			Message:     rev.DismissalMessage,
			DismissedBy: &prActorDTO{User: &prUserActorDTO{ID: rev.DismissedByUserID}},
			DismissedAt: rev.DismissedAt,
		},
	})
}

// Handlers — Reviewers

func (s *Server) listPullRequestRequestedReviewers(w http.ResponseWriter, r *http.Request) {
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

	userIDs, groupIDs, err := s.store.ListRequestedReviewers(r.Context(), pr.ID)
	if err != nil {
		internalError(w, "list requested reviewers: "+err.Error())
		return
	}

	type userDTO struct {
		ID string `json:"id"`
	}
	type groupDTO struct {
		ID string `json:"id"`
	}

	users := make([]userDTO, 0, len(userIDs))
	for _, uid := range userIDs {
		users = append(users, userDTO{ID: uid})
	}
	groups := make([]groupDTO, 0, len(groupIDs))
	for _, gid := range groupIDs {
		groups = append(groups, groupDTO{ID: gid})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"users":  users,
		"groups": groups,
	})
}

func (s *Server) requestPullRequestReviewers(w http.ResponseWriter, r *http.Request) {
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
		Users  []string `json:"users"`
		Groups []string `json:"groups"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	if err := s.store.AddRequestedReviewers(r.Context(), pr.ID, req.Users, req.Groups); err != nil {
		internalError(w, "add reviewers: "+err.Error())
		return
	}

	s.listPullRequestRequestedReviewers(w, r)
}

func (s *Server) removePullRequestRequestedReviewers(w http.ResponseWriter, r *http.Request) {
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
		Users  []string `json:"users"`
		Groups []string `json:"groups"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	if err := s.store.RemoveRequestedReviewers(r.Context(), pr.ID, req.Users, req.Groups); err != nil {
		internalError(w, "remove reviewers: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{})
}

// Handlers — Labels

func (s *Server) listPullRequestLabels(w http.ResponseWriter, r *http.Request) {
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

	labels, err := s.store.GetPRLabels(r.Context(), pr.ID)
	if err != nil {
		internalError(w, "get labels: "+err.Error())
		return
	}

	dtos := make([]labelDTO, 0, len(labels))
	for _, l := range labels {
		dtos = append(dtos, labelDTO{ID: l.ID, Name: l.Name, Color: l.Color, Description: l.Description})
	}

	writeJSON(w, http.StatusOK, map[string]any{"labels": dtos})
}

func (s *Server) setPullRequestLabels(w http.ResponseWriter, r *http.Request) {
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
		Labels        []string `json:"labels"`
		CreateMissing bool     `json:"createMissing"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	var labelIDs []string
	for _, labelName := range req.Labels {
		if labelName == "" {
			continue
		}
		lbl, err := s.store.GetOrCreateLabel(r.Context(), idgen.NewLabel(), repo.NamespaceID, labelName, "ededed", "")
		if err != nil {
			internalError(w, "get or create label: "+err.Error())
			return
		}
		labelIDs = append(labelIDs, lbl.ID)
	}

	if err := s.store.SetPRLabels(r.Context(), pr.ID, labelIDs); err != nil {
		internalError(w, "set labels: "+err.Error())
		return
	}

	s.listPullRequestLabels(w, r)
}

func (s *Server) addPullRequestLabels(w http.ResponseWriter, r *http.Request) {
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
		Labels        []string `json:"labels"`
		CreateMissing bool     `json:"createMissing"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	var labelIDs []string
	for _, labelName := range req.Labels {
		if labelName == "" {
			continue
		}
		lbl, err := s.store.GetOrCreateLabel(r.Context(), idgen.NewLabel(), repo.NamespaceID, labelName, "ededed", "")
		if err != nil {
			internalError(w, "get or create label: "+err.Error())
			return
		}
		labelIDs = append(labelIDs, lbl.ID)
	}

	if err := s.store.AddPRLabels(r.Context(), pr.ID, labelIDs); err != nil {
		internalError(w, "add labels: "+err.Error())
		return
	}

	s.listPullRequestLabels(w, r)
}

func (s *Server) removePullRequestLabel(w http.ResponseWriter, r *http.Request) {
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

	labelName := r.PathValue("labelName")
	labels, err := s.store.GetPRLabels(r.Context(), pr.ID)
	if err != nil {
		internalError(w, "get labels: "+err.Error())
		return
	}

	for _, l := range labels {
		if strings.EqualFold(l.Name, labelName) {
			_ = s.store.RemovePRLabel(r.Context(), pr.ID, l.ID)
			break
		}
	}

	s.listPullRequestLabels(w, r)
}

func (s *Server) removeAllPullRequestLabels(w http.ResponseWriter, r *http.Request) {
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

	if err := s.store.RemoveAllPRLabels(r.Context(), pr.ID); err != nil {
		internalError(w, "remove all labels: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{})
}
