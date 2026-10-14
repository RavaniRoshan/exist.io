package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ---------- Pull Requests ----------

// PullRequest represents a pull request row.
type PullRequest struct {
	ID             string
	RepoID         string
	Number         int
	State          string // open | closed
	Draft          bool
	Merged         bool
	Title          string
	Body           string
	HeadRef        string
	HeadSHA        string
	BaseRef        string
	BaseSHA        string
	AuthorUserID   string
	MergeCommitSHA string
	Mergeable      string // mergeable | stale | blocked
	Additions      int
	Deletions      int
	ChangedFiles   int
	AutoMerge      bool
	ParentPRID     string
	CreatedAt      string
	UpdatedAt      string
	ClosedAt       string
	MergedAt       string
}

// PullRequestUpdate contains optional fields to update on a pull request.
type PullRequestUpdate struct {
	Title          *string
	Body           *string
	State          *string
	Draft          *bool
	Merged         *bool
	HeadRef        *string
	HeadSHA        *string
	BaseRef        *string
	BaseSHA        *string
	MergeCommitSHA *string
	Mergeable      *string
	Additions      *int
	Deletions      *int
	ChangedFiles   *int
	ParentPRID     *string
	ClosedAt       *string
	MergedAt       *string
}

// CreatePullRequest inserts a new pull request, allocating the next sequential
// number for the repo in an atomic transaction.
func (db *DB) CreatePullRequest(ctx context.Context, pr *PullRequest) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var maxNum sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(number) FROM pull_requests WHERE repo_id = ?`, pr.RepoID).Scan(&maxNum); err != nil {
		return err
	}
	pr.Number = 1
	if maxNum.Valid {
		pr.Number = int(maxNum.Int64) + 1
	}

	t := now()
	pr.CreatedAt = t
	pr.UpdatedAt = t

	_, err = tx.ExecContext(ctx, `INSERT INTO pull_requests
		(id, repo_id, number, state, draft, merged, title, body,
		 head_ref, head_sha, base_ref, base_sha, author_user_id,
		 merge_commit_sha, mergeable, additions, deletions, changed_files,
		 auto_merge, parent_pr_id, created_at, updated_at, closed_at, merged_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		pr.ID, pr.RepoID, pr.Number, pr.State, b2i(pr.Draft), b2i(pr.Merged),
		pr.Title, nullStr(pr.Body), pr.HeadRef, pr.HeadSHA, pr.BaseRef, pr.BaseSHA,
		nullStr(pr.AuthorUserID), nullStr(pr.MergeCommitSHA), nullStr(pr.Mergeable),
		pr.Additions, pr.Deletions, pr.ChangedFiles, b2i(pr.AutoMerge),
		nullStr(pr.ParentPRID), pr.CreatedAt, pr.UpdatedAt, nullStr(pr.ClosedAt), nullStr(pr.MergedAt))
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetPullRequest loads a pull request by repo ID and number.
func (db *DB) GetPullRequest(ctx context.Context, repoID string, number int) (PullRequest, error) {
	row := db.sql.QueryRowContext(ctx, `SELECT id, repo_id, number, state, draft, merged, title, body,
		head_ref, head_sha, base_ref, base_sha, author_user_id,
		merge_commit_sha, mergeable, additions, deletions, changed_files,
		auto_merge, parent_pr_id, created_at, updated_at, closed_at, merged_at
		FROM pull_requests WHERE repo_id = ? AND number = ?`, repoID, number)
	return scanPullRequest(row)
}

// GetPullRequestByID loads a pull request by its unique ID (pr_...).
func (db *DB) GetPullRequestByID(ctx context.Context, id string) (PullRequest, error) {
	row := db.sql.QueryRowContext(ctx, `SELECT id, repo_id, number, state, draft, merged, title, body,
		head_ref, head_sha, base_ref, base_sha, author_user_id,
		merge_commit_sha, mergeable, additions, deletions, changed_files,
		auto_merge, parent_pr_id, created_at, updated_at, closed_at, merged_at
		FROM pull_requests WHERE id = ?`, id)
	return scanPullRequest(row)
}

// ListPullRequests returns pull requests for a repo, optionally filtered by state ("open" or "closed").
// Pagination is cursor-based on PR number.
func (db *DB) ListPullRequests(ctx context.Context, repoID, state string, pageSize int, pageToken string) ([]PullRequest, string, error) {
	if pageSize <= 0 {
		pageSize = 30
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var cursorNumber int = 1<<31 - 1
	if pageToken != "" {
		if n, err := strconv.Atoi(pageToken); err == nil {
			cursorNumber = n
		}
	}

	var query strings.Builder
	var args []any
	query.WriteString(`SELECT id, repo_id, number, state, draft, merged, title, body,
		head_ref, head_sha, base_ref, base_sha, author_user_id,
		merge_commit_sha, mergeable, additions, deletions, changed_files,
		auto_merge, parent_pr_id, created_at, updated_at, closed_at, merged_at
		FROM pull_requests WHERE repo_id = ? AND number <= ?`)
	args = append(args, repoID, cursorNumber)

	if state != "" && state != "all" {
		query.WriteString(` AND state = ?`)
		args = append(args, state)
	}
	query.WriteString(` ORDER BY number DESC LIMIT ?`)
	args = append(args, pageSize+1)

	rows, err := db.sql.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var prs []PullRequest
	for rows.Next() {
		pr, err := scanPullRequest(rows)
		if err != nil {
			return nil, "", err
		}
		prs = append(prs, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(prs) > pageSize {
		nextToken = strconv.Itoa(prs[pageSize].Number)
		prs = prs[:pageSize]
	}

	return prs, nextToken, nil
}

// UpdatePullRequest applies partial updates to a pull request.
func (db *DB) UpdatePullRequest(ctx context.Context, prID string, u PullRequestUpdate) error {
	var sets []string
	var args []any

	if u.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *u.Title)
	}
	if u.Body != nil {
		sets = append(sets, "body = ?")
		args = append(args, nullStr(*u.Body))
	}
	if u.State != nil {
		sets = append(sets, "state = ?")
		args = append(args, *u.State)
	}
	if u.Draft != nil {
		sets = append(sets, "draft = ?")
		args = append(args, b2i(*u.Draft))
	}
	if u.Merged != nil {
		sets = append(sets, "merged = ?")
		args = append(args, b2i(*u.Merged))
	}
	if u.HeadRef != nil {
		sets = append(sets, "head_ref = ?")
		args = append(args, *u.HeadRef)
	}
	if u.HeadSHA != nil {
		sets = append(sets, "head_sha = ?")
		args = append(args, *u.HeadSHA)
	}
	if u.BaseRef != nil {
		sets = append(sets, "base_ref = ?")
		args = append(args, *u.BaseRef)
	}
	if u.BaseSHA != nil {
		sets = append(sets, "base_sha = ?")
		args = append(args, *u.BaseSHA)
	}
	if u.MergeCommitSHA != nil {
		sets = append(sets, "merge_commit_sha = ?")
		args = append(args, nullStr(*u.MergeCommitSHA))
	}
	if u.Mergeable != nil {
		sets = append(sets, "mergeable = ?")
		args = append(args, nullStr(*u.Mergeable))
	}
	if u.Additions != nil {
		sets = append(sets, "additions = ?")
		args = append(args, *u.Additions)
	}
	if u.Deletions != nil {
		sets = append(sets, "deletions = ?")
		args = append(args, *u.Deletions)
	}
	if u.ChangedFiles != nil {
		sets = append(sets, "changed_files = ?")
		args = append(args, *u.ChangedFiles)
	}
	if u.ParentPRID != nil {
		sets = append(sets, "parent_pr_id = ?")
		args = append(args, nullStr(*u.ParentPRID))
	}
	if u.ClosedAt != nil {
		sets = append(sets, "closed_at = ?")
		args = append(args, nullStr(*u.ClosedAt))
	}
	if u.MergedAt != nil {
		sets = append(sets, "merged_at = ?")
		args = append(args, nullStr(*u.MergedAt))
	}

	if len(sets) == 0 {
		return nil
	}

	sets = append(sets, "updated_at = ?")
	args = append(args, now())
	args = append(args, prID)

	query := fmt.Sprintf("UPDATE pull_requests SET %s WHERE id = ?", strings.Join(sets, ", "))
	res, err := db.sql.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// SetPullRequestMergeResult records the merge outcome on a PR.
func (db *DB) SetPullRequestMergeResult(ctx context.Context, prID, mergeCommitSHA string) error {
	t := now()
	res, err := db.sql.ExecContext(ctx, `UPDATE pull_requests SET
		merged = 1, state = 'closed', merge_commit_sha = ?, merged_at = ?, closed_at = ?, updated_at = ?
		WHERE id = ?`, mergeCommitSHA, t, t, t, prID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- Auto-Merge ----------

// PRAutoMerge represents auto-merge configuration on a PR.
type PRAutoMerge struct {
	PullRequestID   string
	EnabledByUserID string
	EnabledAt       string
	MergeMethod     string // merge | squash
	CommitTitle     string
	CommitMessage   string
}

// SetAutoMerge enables auto-merge on a pull request.
func (db *DB) SetAutoMerge(ctx context.Context, am PRAutoMerge) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if am.EnabledAt == "" {
		am.EnabledAt = now()
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO pr_auto_merge
		(pull_request_id, enabled_by_user_id, enabled_at, merge_method, commit_title, commit_message)
		VALUES (?,?,?,?,?,?)
		ON CONFLICT(pull_request_id) DO UPDATE SET
		enabled_by_user_id = excluded.enabled_by_user_id,
		enabled_at = excluded.enabled_at,
		merge_method = excluded.merge_method,
		commit_title = excluded.commit_title,
		commit_message = excluded.commit_message`,
		am.PullRequestID, nullStr(am.EnabledByUserID), am.EnabledAt, am.MergeMethod,
		nullStr(am.CommitTitle), nullStr(am.CommitMessage))
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE pull_requests SET auto_merge = 1, updated_at = ? WHERE id = ?`, now(), am.PullRequestID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetAutoMerge returns the auto-merge settings for a PR, or nil if not enabled.
func (db *DB) GetAutoMerge(ctx context.Context, prID string) (*PRAutoMerge, error) {
	row := db.sql.QueryRowContext(ctx, `SELECT pull_request_id, enabled_by_user_id, enabled_at, merge_method, commit_title, commit_message
		FROM pr_auto_merge WHERE pull_request_id = ?`, prID)
	var am PRAutoMerge
	var enabledBy, title, msg sql.NullString
	if err := row.Scan(&am.PullRequestID, &enabledBy, &am.EnabledAt, &am.MergeMethod, &title, &msg); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	am.EnabledByUserID = enabledBy.String
	am.CommitTitle = title.String
	am.CommitMessage = msg.String
	return &am, nil
}

// DeleteAutoMerge disables auto-merge on a pull request.
func (db *DB) DeleteAutoMerge(ctx context.Context, prID string) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM pr_auto_merge WHERE pull_request_id = ?`, prID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE pull_requests SET auto_merge = 0, updated_at = ? WHERE id = ?`, now(), prID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// ---------- Threads & Comments ----------

// PRThread represents a discussion thread (general or anchored).
type PRThread struct {
	ID                string
	RepoID            string
	PullRequestID     string
	Path              string
	Side              string
	StartLine         int
	EndLine           int
	VersionNumber     int
	HeadSHA           string
	BaseSHA           string
	ResolvedAt        string
	ResolvedByUserID  string
	CreatedAt         string
	UpdatedAt         string
}

// CreateThread inserts a new discussion thread.
func (db *DB) CreateThread(ctx context.Context, t PRThread) error {
	ts := now()
	if t.CreatedAt == "" {
		t.CreatedAt = ts
	}
	if t.UpdatedAt == "" {
		t.UpdatedAt = ts
	}
	if t.VersionNumber <= 0 {
		t.VersionNumber = 1
	}

	_, err := db.sql.ExecContext(ctx, `INSERT INTO pr_threads
		(id, repo_id, pull_request_id, path, side, start_line, end_line,
		 version_number, head_sha, base_sha, resolved_at, resolved_by_user_id,
		 created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.RepoID, t.PullRequestID, nullStr(t.Path), nullStr(t.Side),
		t.StartLine, t.EndLine, t.VersionNumber, nullStr(t.HeadSHA), nullStr(t.BaseSHA),
		nullStr(t.ResolvedAt), nullStr(t.ResolvedByUserID), t.CreatedAt, t.UpdatedAt)
	return err
}

// GetThread returns a discussion thread by ID.
func (db *DB) GetThread(ctx context.Context, id string) (PRThread, error) {
	row := db.sql.QueryRowContext(ctx, `SELECT id, repo_id, pull_request_id, path, side, start_line, end_line,
		version_number, head_sha, base_sha, resolved_at, resolved_by_user_id, created_at, updated_at
		FROM pr_threads WHERE id = ?`, id)
	return scanThread(row)
}

// UpdateThread updates the resolution state of a thread.
func (db *DB) UpdateThread(ctx context.Context, id string, resolvedAt *string, resolvedByUserID *string) error {
	sets := []string{"updated_at = ?"}
	args := []any{now()}
	if resolvedAt != nil {
		sets = append(sets, "resolved_at = ?")
		args = append(args, nullStr(*resolvedAt))
	}
	if resolvedByUserID != nil {
		sets = append(sets, "resolved_by_user_id = ?")
		args = append(args, nullStr(*resolvedByUserID))
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE pr_threads SET %s WHERE id = ?", strings.Join(sets, ", "))
	res, err := db.sql.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// PRComment represents a single comment in a thread.
type PRComment struct {
	ID             string
	RepoID         string
	PullRequestID  string
	ThreadID       string
	AuthorUserID   string
	Body           string
	Path           string
	Line           int
	Side           string
	CommitSHA      string
	InReplyTo      string
	CreatedAt      string
	UpdatedAt      string
}

// CreateComment inserts a comment on a pull request.
func (db *DB) CreateComment(ctx context.Context, c PRComment) error {
	ts := now()
	if c.CreatedAt == "" {
		c.CreatedAt = ts
	}
	if c.UpdatedAt == "" {
		c.UpdatedAt = ts
	}
	_, err := db.sql.ExecContext(ctx, `INSERT INTO pr_comments
		(id, repo_id, pull_request_id, thread_id, author_user_id, body,
		 path, line, side, commit_sha, in_reply_to, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		c.ID, c.RepoID, c.PullRequestID, nullStr(c.ThreadID), nullStr(c.AuthorUserID),
		c.Body, nullStr(c.Path), c.Line, nullStr(c.Side), nullStr(c.CommitSHA),
		nullStr(c.InReplyTo), c.CreatedAt, c.UpdatedAt)
	return err
}

// GetComment returns a comment by ID.
func (db *DB) GetComment(ctx context.Context, id string) (PRComment, error) {
	row := db.sql.QueryRowContext(ctx, `SELECT id, repo_id, pull_request_id, COALESCE(thread_id, ''),
		COALESCE(author_user_id, ''), body, COALESCE(path, ''), COALESCE(line, 0),
		COALESCE(side, ''), COALESCE(commit_sha, ''), COALESCE(in_reply_to, ''),
		created_at, updated_at
		FROM pr_comments WHERE id = ?`, id)
	return scanComment(row)
}

// UpdateComment updates a comment's body.
func (db *DB) UpdateComment(ctx context.Context, id, body string) error {
	res, err := db.sql.ExecContext(ctx, `UPDATE pr_comments SET body = ?, updated_at = ? WHERE id = ?`, body, now(), id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteComment removes a comment.
func (db *DB) DeleteComment(ctx context.Context, id string) error {
	res, err := db.sql.ExecContext(ctx, `DELETE FROM pr_comments WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListComments returns all comments for a PR in chronological order.
func (db *DB) ListComments(ctx context.Context, prID string) ([]PRComment, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT id, repo_id, pull_request_id, COALESCE(thread_id, ''),
		COALESCE(author_user_id, ''), body, COALESCE(path, ''), COALESCE(line, 0),
		COALESCE(side, ''), COALESCE(commit_sha, ''), COALESCE(in_reply_to, ''),
		created_at, updated_at
		FROM pr_comments WHERE pull_request_id = ? ORDER BY created_at ASC`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []PRComment
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// ---------- Reactions ----------

// PRCommentReaction represents an emoji reaction on a PR comment.
type PRCommentReaction struct {
	ID        string
	CommentID string
	UserID    string
	Emoji     string
	CreatedAt string
}

// AddReaction adds an emoji reaction to a comment.
func (db *DB) AddReaction(ctx context.Context, rxn PRCommentReaction) error {
	if rxn.CreatedAt == "" {
		rxn.CreatedAt = now()
	}
	_, err := db.sql.ExecContext(ctx, `INSERT INTO pr_comment_reactions
		(id, comment_id, user_id, emoji, created_at)
		VALUES (?,?,?,?,?)
		ON CONFLICT(comment_id, user_id, emoji) DO NOTHING`,
		rxn.ID, rxn.CommentID, rxn.UserID, rxn.Emoji, rxn.CreatedAt)
	return err
}

// RemoveReaction removes a user's emoji reaction from a comment.
func (db *DB) RemoveReaction(ctx context.Context, commentID, userID, emoji string) error {
	_, err := db.sql.ExecContext(ctx, `DELETE FROM pr_comment_reactions
		WHERE comment_id = ? AND user_id = ? AND emoji = ?`,
		commentID, userID, emoji)
	return err
}

// ListReactions returns all reactions for a comment ordered oldest first.
func (db *DB) ListReactions(ctx context.Context, commentID string) ([]PRCommentReaction, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT id, comment_id, user_id, emoji, created_at
		FROM pr_comment_reactions WHERE comment_id = ? ORDER BY created_at ASC`, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rxns []PRCommentReaction
	for rows.Next() {
		var r PRCommentReaction
		if err := rows.Scan(&r.ID, &r.CommentID, &r.UserID, &r.Emoji, &r.CreatedAt); err != nil {
			return nil, err
		}
		rxns = append(rxns, r)
	}
	return rxns, rows.Err()
}

// ---------- Reviews ----------

// PRReview represents a pull request review.
type PRReview struct {
	ID               string
	RepoID           string
	PullRequestID    string
	AuthorUserID     string
	State            string // approved | changes_requested | commented | dismissed
	Body             string
	CommitSHA        string
	Dismissed        bool
	DismissalMessage string
	DismissedByUserID string
	DismissedAt      string
	VersionNumber    int
	CreatedAt        string
	UpdatedAt        string
}

// CreateReview inserts a new review.
func (db *DB) CreateReview(ctx context.Context, rev PRReview) error {
	ts := now()
	if rev.CreatedAt == "" {
		rev.CreatedAt = ts
	}
	if rev.UpdatedAt == "" {
		rev.UpdatedAt = ts
	}
	if rev.VersionNumber <= 0 {
		rev.VersionNumber = 1
	}

	_, err := db.sql.ExecContext(ctx, `INSERT INTO pr_reviews
		(id, repo_id, pull_request_id, author_user_id, state, body,
		 commit_sha, dismissed, dismissal_message, dismissed_by_user_id,
		 dismissed_at, version_number, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		rev.ID, rev.RepoID, rev.PullRequestID, nullStr(rev.AuthorUserID),
		rev.State, nullStr(rev.Body), nullStr(rev.CommitSHA), b2i(rev.Dismissed),
		nullStr(rev.DismissalMessage), nullStr(rev.DismissedByUserID),
		nullStr(rev.DismissedAt), rev.VersionNumber, rev.CreatedAt, rev.UpdatedAt)
	return err
}

// GetReview returns a review by ID.
func (db *DB) GetReview(ctx context.Context, id string) (PRReview, error) {
	row := db.sql.QueryRowContext(ctx, `SELECT id, repo_id, pull_request_id, COALESCE(author_user_id, ''),
		state, COALESCE(body, ''), COALESCE(commit_sha, ''), dismissed,
		COALESCE(dismissal_message, ''), COALESCE(dismissed_by_user_id, ''),
		COALESCE(dismissed_at, ''), version_number, created_at, updated_at
		FROM pr_reviews WHERE id = ?`, id)
	return scanReview(row)
}

// UpdateReview updates a review's summary body.
func (db *DB) UpdateReview(ctx context.Context, id, body string) error {
	res, err := db.sql.ExecContext(ctx, `UPDATE pr_reviews SET body = ?, updated_at = ? WHERE id = ?`, body, now(), id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DismissReview marks a review as dismissed with a message.
func (db *DB) DismissReview(ctx context.Context, id, message, dismissedByUserID string) error {
	t := now()
	res, err := db.sql.ExecContext(ctx, `UPDATE pr_reviews SET
		dismissed = 1, state = 'dismissed', dismissal_message = ?, dismissed_by_user_id = ?, dismissed_at = ?, updated_at = ?
		WHERE id = ?`, message, nullStr(dismissedByUserID), t, t, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListReviews returns all reviews for a PR ordered oldest first.
func (db *DB) ListReviews(ctx context.Context, prID string) ([]PRReview, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT id, repo_id, pull_request_id, COALESCE(author_user_id, ''),
		state, COALESCE(body, ''), COALESCE(commit_sha, ''), dismissed,
		COALESCE(dismissal_message, ''), COALESCE(dismissed_by_user_id, ''),
		COALESCE(dismissed_at, ''), version_number, created_at, updated_at
		FROM pr_reviews WHERE pull_request_id = ? ORDER BY created_at ASC`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []PRReview
	for rows.Next() {
		rev, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, rows.Err()
}

// ---------- Requested Reviewers ----------

// AddRequestedReviewers associates users and/or groups as requested reviewers.
func (db *DB) AddRequestedReviewers(ctx context.Context, prID string, userIDs, groupIDs []string) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	t := now()
	for _, uid := range userIDs {
		if uid == "" {
			continue
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO pr_requested_reviewers
			(pull_request_id, user_id, group_id, created_at)
			VALUES (?, ?, NULL, ?)`, prID, uid, t)
		if err != nil {
			return err
		}
	}
	for _, gid := range groupIDs {
		if gid == "" {
			continue
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO pr_requested_reviewers
			(pull_request_id, user_id, group_id, created_at)
			VALUES (?, NULL, ?, ?)`, prID, gid, t)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RemoveRequestedReviewers removes requested reviewer records.
func (db *DB) RemoveRequestedReviewers(ctx context.Context, prID string, userIDs, groupIDs []string) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, uid := range userIDs {
		if uid == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM pr_requested_reviewers WHERE pull_request_id = ? AND user_id = ?`, prID, uid); err != nil {
			return err
		}
	}
	for _, gid := range groupIDs {
		if gid == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM pr_requested_reviewers WHERE pull_request_id = ? AND group_id = ?`, prID, gid); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ListRequestedReviewers returns requested user IDs and group IDs for a PR.
func (db *DB) ListRequestedReviewers(ctx context.Context, prID string) (userIDs, groupIDs []string, err error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT COALESCE(user_id, ''), COALESCE(group_id, '')
		FROM pr_requested_reviewers WHERE pull_request_id = ?`, prID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid, gid string
		if err := rows.Scan(&uid, &gid); err != nil {
			return nil, nil, err
		}
		if uid != "" {
			userIDs = append(userIDs, uid)
		}
		if gid != "" {
			groupIDs = append(groupIDs, gid)
		}
	}
	return userIDs, groupIDs, rows.Err()
}

// ---------- Labels ----------

// Label represents a label row in the namespace.
type Label struct {
	ID          string
	NamespaceID string
	Name        string
	Color       string
	Description string
	CreatedAt   string
}

// GetOrCreateLabel finds or creates a label in the namespace.
func (db *DB) GetOrCreateLabel(ctx context.Context, id, namespaceID, name, color, description string) (Label, error) {
	row := db.sql.QueryRowContext(ctx, `SELECT id, namespace_id, name, COALESCE(color, ''), COALESCE(description, ''), created_at
		FROM labels WHERE namespace_id = ? AND LOWER(name) = LOWER(?)`, namespaceID, name)
	var lbl Label
	if err := row.Scan(&lbl.ID, &lbl.NamespaceID, &lbl.Name, &lbl.Color, &lbl.Description, &lbl.CreatedAt); err == nil {
		return lbl, nil
	}

	t := now()
	_, err := db.sql.ExecContext(ctx, `INSERT INTO labels
		(id, namespace_id, name, color, description, created_at)
		VALUES (?,?,?,?,?,?)`,
		id, namespaceID, name, nullStr(color), nullStr(description), t)
	if err != nil {
		return Label{}, err
	}
	return Label{ID: id, NamespaceID: namespaceID, Name: name, Color: color, Description: description, CreatedAt: t}, nil
}

// ListLabels returns all labels in a namespace.
func (db *DB) ListLabels(ctx context.Context, namespaceID string) ([]Label, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT id, namespace_id, name, COALESCE(color, ''), COALESCE(description, ''), created_at
		FROM labels WHERE namespace_id = ? ORDER BY name ASC`, namespaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		var l Label
		if err := rows.Scan(&l.ID, &l.NamespaceID, &l.Name, &l.Color, &l.Description, &l.CreatedAt); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

// GetLabel retrieves a label by namespace ID and name.
func (db *DB) GetLabel(ctx context.Context, namespaceID, name string) (Label, error) {
	row := db.sql.QueryRowContext(ctx, `SELECT id, namespace_id, name, COALESCE(color, ''), COALESCE(description, ''), created_at
		FROM labels WHERE namespace_id = ? AND LOWER(name) = LOWER(?)`, namespaceID, name)
	var lbl Label
	if err := row.Scan(&lbl.ID, &lbl.NamespaceID, &lbl.Name, &lbl.Color, &lbl.Description, &lbl.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Label{}, ErrNotFound
		}
		return Label{}, err
	}
	return lbl, nil
}

// UpdateLabel updates an existing label in a namespace.
func (db *DB) UpdateLabel(ctx context.Context, namespaceID, name, newName, newColor, newDesc string) (Label, error) {
	lbl, err := db.GetLabel(ctx, namespaceID, name)
	if err != nil {
		return Label{}, err
	}
	if newName != "" {
		lbl.Name = newName
	}
	if newColor != "" {
		lbl.Color = newColor
	}
	if newDesc != "" {
		lbl.Description = newDesc
	}
	_, err = db.sql.ExecContext(ctx, `UPDATE labels SET name = ?, color = ?, description = ? WHERE id = ?`,
		lbl.Name, nullStr(lbl.Color), nullStr(lbl.Description), lbl.ID)
	if err != nil {
		return Label{}, err
	}
	return lbl, nil
}

// DeleteLabel deletes a label in a namespace.
func (db *DB) DeleteLabel(ctx context.Context, namespaceID, name string) error {
	res, err := db.sql.ExecContext(ctx, `DELETE FROM labels WHERE namespace_id = ? AND LOWER(name) = LOWER(?)`, namespaceID, name)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// GetPRLabels returns the labels assigned to a PR.
func (db *DB) GetPRLabels(ctx context.Context, prID string) ([]Label, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT l.id, l.namespace_id, l.name, COALESCE(l.color, ''), COALESCE(l.description, ''), l.created_at
		FROM labels l
		INNER JOIN pr_labels pl ON pl.label_id = l.id
		WHERE pl.pull_request_id = ?
		ORDER BY l.name ASC`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		var l Label
		if err := rows.Scan(&l.ID, &l.NamespaceID, &l.Name, &l.Color, &l.Description, &l.CreatedAt); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

// SetPRLabels replaces all labels assigned to a PR.
func (db *DB) SetPRLabels(ctx context.Context, prID string, labelIDs []string) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM pr_labels WHERE pull_request_id = ?`, prID); err != nil {
		return err
	}
	for _, lid := range labelIDs {
		if lid == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO pr_labels (pull_request_id, label_id) VALUES (?,?)`, prID, lid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// AddPRLabels adds labels to a PR if not already present.
func (db *DB) AddPRLabels(ctx context.Context, prID string, labelIDs []string) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, lid := range labelIDs {
		if lid == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO pr_labels (pull_request_id, label_id) VALUES (?,?)`, prID, lid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// RemovePRLabel removes a specific label from a PR.
func (db *DB) RemovePRLabel(ctx context.Context, prID, labelID string) error {
	_, err := db.sql.ExecContext(ctx, `DELETE FROM pr_labels WHERE pull_request_id = ? AND label_id = ?`, prID, labelID)
	return err
}

// RemoveAllPRLabels removes all labels from a PR.
func (db *DB) RemoveAllPRLabels(ctx context.Context, prID string) error {
	_, err := db.sql.ExecContext(ctx, `DELETE FROM pr_labels WHERE pull_request_id = ?`, prID)
	return err
}

// ---------- Scanning Helpers ----------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPullRequest(s rowScanner) (PullRequest, error) {
	var pr PullRequest
	var draft, merged, autoMerge int
	var body, author, mergeCommit, mergeable, parent, closedAt, mergedAt sql.NullString
	err := s.Scan(&pr.ID, &pr.RepoID, &pr.Number, &pr.State, &draft, &merged,
		&pr.Title, &body, &pr.HeadRef, &pr.HeadSHA, &pr.BaseRef, &pr.BaseSHA,
		&author, &mergeCommit, &mergeable, &pr.Additions, &pr.Deletions, &pr.ChangedFiles,
		&autoMerge, &parent, &pr.CreatedAt, &pr.UpdatedAt, &closedAt, &mergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PullRequest{}, ErrNotFound
		}
		return PullRequest{}, err
	}
	pr.Draft = draft == 1
	pr.Merged = merged == 1
	pr.AutoMerge = autoMerge == 1
	pr.Body = body.String
	pr.AuthorUserID = author.String
	pr.MergeCommitSHA = mergeCommit.String
	pr.Mergeable = mergeable.String
	pr.ParentPRID = parent.String
	pr.ClosedAt = closedAt.String
	pr.MergedAt = mergedAt.String
	return pr, nil
}

func scanThread(s rowScanner) (PRThread, error) {
	var t PRThread
	var path, side, headSHA, baseSHA, resolvedAt, resolvedBy sql.NullString
	err := s.Scan(&t.ID, &t.RepoID, &t.PullRequestID, &path, &side, &t.StartLine, &t.EndLine,
		&t.VersionNumber, &headSHA, &baseSHA, &resolvedAt, &resolvedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PRThread{}, ErrNotFound
		}
		return PRThread{}, err
	}
	t.Path = path.String
	t.Side = side.String
	t.HeadSHA = headSHA.String
	t.BaseSHA = baseSHA.String
	t.ResolvedAt = resolvedAt.String
	t.ResolvedByUserID = resolvedBy.String
	return t, nil
}

func scanComment(s rowScanner) (PRComment, error) {
	var c PRComment
	err := s.Scan(&c.ID, &c.RepoID, &c.PullRequestID, &c.ThreadID, &c.AuthorUserID,
		&c.Body, &c.Path, &c.Line, &c.Side, &c.CommitSHA, &c.InReplyTo,
		&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PRComment{}, ErrNotFound
		}
		return PRComment{}, err
	}
	return c, nil
}

func scanReview(s rowScanner) (PRReview, error) {
	var r PRReview
	var dismissed int
	err := s.Scan(&r.ID, &r.RepoID, &r.PullRequestID, &r.AuthorUserID, &r.State,
		&r.Body, &r.CommitSHA, &dismissed, &r.DismissalMessage, &r.DismissedByUserID,
		&r.DismissedAt, &r.VersionNumber, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PRReview{}, ErrNotFound
		}
		return PRReview{}, err
	}
	r.Dismissed = dismissed == 1
	return r, nil
}
