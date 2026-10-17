package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"exist.io/pkg/idgen"
)

// Issue is a stored repository issue.
type Issue struct {
	ID           string `json:"id"`
	RepoID       string `json:"repo_id"`
	Number       int    `json:"number"`
	State        string `json:"state"` // open | closed
	Title        string `json:"title"`
	Body         string `json:"body"`
	AuthorUserID string `json:"author_user_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	ClosedAt     string `json:"closed_at,omitempty"`
}

// IssueComment is a comment on a repository issue.
type IssueComment struct {
	ID           string `json:"id"`
	RepoID       string `json:"repo_id"`
	IssueID      string `json:"issue_id"`
	AuthorUserID string `json:"author_user_id"`
	Body         string `json:"body"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// NextIssueNumber computes the next issue number for a repository.
func (db *DB) NextIssueNumber(ctx context.Context, repoID string) (int, error) {
	var maxNumber sql.NullInt64
	err := db.sql.QueryRowContext(ctx, `SELECT MAX(number) FROM (
		SELECT number FROM issues WHERE repo_id = ?
		UNION ALL
		SELECT number FROM pull_requests WHERE repo_id = ?
	)`, repoID, repoID).Scan(&maxNumber)
	if err != nil {
		return 0, fmt.Errorf("next issue number: %w", err)
	}
	if !maxNumber.Valid {
		return 1, nil
	}
	return int(maxNumber.Int64) + 1, nil
}

// CreateIssue persists a new issue record.
func (db *DB) CreateIssue(ctx context.Context, issue *Issue) error {
	if issue.ID == "" {
		issue.ID = idgen.NewIssue()
	}
	t := now()
	if issue.CreatedAt == "" {
		issue.CreatedAt = t
	}
	if issue.UpdatedAt == "" {
		issue.UpdatedAt = t
	}
	if issue.State == "" {
		issue.State = "open"
	}
	if issue.Number <= 0 {
		num, err := db.NextIssueNumber(ctx, issue.RepoID)
		if err != nil {
			return err
		}
		issue.Number = num
	}

	_, err := db.sql.ExecContext(ctx, `INSERT INTO issues
		(id, repo_id, number, state, title, body, author_user_id, created_at, updated_at, closed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		issue.ID, issue.RepoID, issue.Number, issue.State, issue.Title,
		nullStr(issue.Body), nullStr(issue.AuthorUserID), issue.CreatedAt, issue.UpdatedAt, nullStr(issue.ClosedAt))
	if err != nil {
		return fmt.Errorf("create issue: %w", err)
	}
	return nil
}

// GetIssue fetches an issue by repository ID and number.
func (db *DB) GetIssue(ctx context.Context, repoID string, number int) (*Issue, error) {
	row := db.sql.QueryRowContext(ctx, `SELECT id, repo_id, number, state, title, IFNULL(body, ''), IFNULL(author_user_id, ''), created_at, updated_at, IFNULL(closed_at, '')
		FROM issues WHERE repo_id = ? AND number = ?`, repoID, number)
	var i Issue
	if err := row.Scan(&i.ID, &i.RepoID, &i.Number, &i.State, &i.Title, &i.Body, &i.AuthorUserID, &i.CreatedAt, &i.UpdatedAt, &i.ClosedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get issue: %w", err)
	}
	return &i, nil
}

// ListIssues returns issues for a repository, optionally filtered by state ("open", "closed", or "" for all).
func (db *DB) ListIssues(ctx context.Context, repoID string, state string) ([]Issue, error) {
	query := `SELECT id, repo_id, number, state, title, IFNULL(body, ''), IFNULL(author_user_id, ''), created_at, updated_at, IFNULL(closed_at, '')
		FROM issues WHERE repo_id = ?`
	args := []any{repoID}
	if state != "" && state != "all" {
		query += ` AND state = ?`
		args = append(args, state)
	}
	query += ` ORDER BY number DESC`

	rows, err := db.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}
	defer rows.Close()

	var out []Issue
	for rows.Next() {
		var i Issue
		if err := rows.Scan(&i.ID, &i.RepoID, &i.Number, &i.State, &i.Title, &i.Body, &i.AuthorUserID, &i.CreatedAt, &i.UpdatedAt, &i.ClosedAt); err != nil {
			return nil, fmt.Errorf("scan issue: %w", err)
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

// UpdateIssueState transitions an issue between "open" and "closed".
func (db *DB) UpdateIssueState(ctx context.Context, repoID string, number int, state string) error {
	t := now()
	var closedAt any = nil
	if state == "closed" {
		closedAt = t
	}
	res, err := db.sql.ExecContext(ctx, `UPDATE issues
		SET state = ?, updated_at = ?, closed_at = ?
		WHERE repo_id = ? AND number = ?`, state, t, closedAt, repoID, number)
	if err != nil {
		return fmt.Errorf("update issue state: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateIssueComment adds a comment to an issue.
func (db *DB) CreateIssueComment(ctx context.Context, comment *IssueComment) error {
	if comment.ID == "" {
		comment.ID = idgen.NewIssueComment()
	}
	t := now()
	if comment.CreatedAt == "" {
		comment.CreatedAt = t
	}
	if comment.UpdatedAt == "" {
		comment.UpdatedAt = t
	}

	_, err := db.sql.ExecContext(ctx, `INSERT INTO issue_comments
		(id, repo_id, issue_id, author_user_id, body, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		comment.ID, comment.RepoID, comment.IssueID, nullStr(comment.AuthorUserID), comment.Body, comment.CreatedAt, comment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create issue comment: %w", err)
	}
	return nil
}

// ListIssueComments retrieves all comments for a given issue ID.
func (db *DB) ListIssueComments(ctx context.Context, issueID string) ([]IssueComment, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT id, repo_id, issue_id, IFNULL(author_user_id, ''), body, created_at, updated_at
		FROM issue_comments WHERE issue_id = ? ORDER BY created_at ASC`, issueID)
	if err != nil {
		return nil, fmt.Errorf("list issue comments: %w", err)
	}
	defer rows.Close()

	var out []IssueComment
	for rows.Next() {
		var c IssueComment
		if err := rows.Scan(&c.ID, &c.RepoID, &c.IssueID, &c.AuthorUserID, &c.Body, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan issue comment: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
