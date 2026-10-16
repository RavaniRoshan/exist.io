package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"exist.io/pkg/idgen"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrPrecondition  = errors.New("precondition failed")
)

type CheckSuite struct {
	ID          string
	RepoID      string
	SHA         string
	BaseSHA     string
	Key         string
	Name        string
	DetailsURL  string
	ExternalID  string
	ActorUserID string
	CreatedAt   string
	UpdatedAt   string
}

type CheckRun struct {
	ID                  string
	RepoID              string
	CheckSuiteID        string
	SHA                 string
	BaseSHA             string
	Key                 string
	Name                string
	Status              string // queued | in_progress | completed | rerequested
	Conclusion          string // success | failure | neutral | cancelled | timed_out | action_required | stale | skipped
	DetailsURL          string
	ExternalID          string
	OutputTitle         string
	OutputSummary       string
	OutputText          string
	ExternalUpdatedAt   string
	StartedAt           string
	CompletedAt         string
	DeadlineAt          string
	IsRerequestable     bool
	RerequestedAt       string
	RerequestedByUserID string
	ActorUserID         string
	CreatedAt           string
	UpdatedAt           string
}

type CheckRunAnnotation struct {
	ID              string
	CheckRunID      string
	Path            string
	StartLine       int
	EndLine         int
	StartColumn     int
	EndColumn       int
	AnnotationLevel string // notice | warning | failure
	Message         string
	Title           string
	RawDetails      string
	CreatedAt       string
	UpdatedAt       string
}

// UpsertCheckSuite inserts or updates a check suite by (repo_id, sha, key, base_sha).
func (db *DB) UpsertCheckSuite(ctx context.Context, s CheckSuite) (CheckSuite, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	if s.CreatedAt == "" {
		s.CreatedAt = now
	}
	s.UpdatedAt = now

	// Check if already exists
	var existing CheckSuite
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repo_id, sha, IFNULL(base_sha, ''), IFNULL(key, ''), IFNULL(name, ''),
		       IFNULL(details_url, ''), IFNULL(external_id, ''), IFNULL(actor_user_id, ''),
		       created_at, updated_at
		FROM check_suites
		WHERE repo_id = ? AND sha = ? AND key = ? AND IFNULL(base_sha, '') = IFNULL(?, '')
	`, s.RepoID, s.SHA, s.Key, s.BaseSHA).Scan(
		&existing.ID, &existing.RepoID, &existing.SHA, &existing.BaseSHA, &existing.Key,
		&existing.Name, &existing.DetailsURL, &existing.ExternalID, &existing.ActorUserID,
		&existing.CreatedAt, &existing.UpdatedAt,
	)

	if err == nil {
		// Update
		_, err = db.sql.ExecContext(ctx, `
			UPDATE check_suites
			SET name = ?, details_url = ?, external_id = ?, actor_user_id = COALESCE(NULLIF(?, ''), actor_user_id), updated_at = ?
			WHERE id = ?
		`, s.Name, s.DetailsURL, s.ExternalID, s.ActorUserID, now, existing.ID)
		if err != nil {
			return CheckSuite{}, fmt.Errorf("update check_suite: %w", err)
		}
		existing.Name = s.Name
		existing.DetailsURL = s.DetailsURL
		existing.ExternalID = s.ExternalID
		if s.ActorUserID != "" {
			existing.ActorUserID = s.ActorUserID
		}
		existing.UpdatedAt = now
		return existing, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return CheckSuite{}, fmt.Errorf("query check_suite: %w", err)
	}

	// Insert new
	if s.ID == "" {
		s.ID = idgen.NewCheckRunGroup()
	}
	_, err = db.sql.ExecContext(ctx, `
		INSERT INTO check_suites (
			id, repo_id, sha, base_sha, key, name, details_url, external_id, actor_user_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?)
	`, s.ID, s.RepoID, s.SHA, s.BaseSHA, s.Key, s.Name, s.DetailsURL, s.ExternalID, s.ActorUserID, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return CheckSuite{}, fmt.Errorf("insert check_suite: %w", err)
	}
	return s, nil
}

// GetCheckSuite retrieves a check suite by ID.
func (db *DB) GetCheckSuite(ctx context.Context, id string) (CheckSuite, error) {
	var s CheckSuite
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repo_id, sha, IFNULL(base_sha, ''), IFNULL(key, ''), IFNULL(name, ''),
		       IFNULL(details_url, ''), IFNULL(external_id, ''), IFNULL(actor_user_id, ''),
		       created_at, updated_at
		FROM check_suites
		WHERE id = ?
	`, id).Scan(
		&s.ID, &s.RepoID, &s.SHA, &s.BaseSHA, &s.Key,
		&s.Name, &s.DetailsURL, &s.ExternalID, &s.ActorUserID,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return CheckSuite{}, ErrNotFound
	}
	if err != nil {
		return CheckSuite{}, fmt.Errorf("get check_suite: %w", err)
	}
	return s, nil
}

// ListCheckSuitesForCommit retrieves suites for a commit SHA.
func (db *DB) ListCheckSuitesForCommit(ctx context.Context, repoID, sha string, limit, offset int) ([]CheckSuite, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, repo_id, sha, IFNULL(base_sha, ''), IFNULL(key, ''), IFNULL(name, ''),
		       IFNULL(details_url, ''), IFNULL(external_id, ''), IFNULL(actor_user_id, ''),
		       created_at, updated_at
		FROM check_suites
		WHERE repo_id = ? AND sha = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, repoID, sha, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list check_suites for commit: %w", err)
	}
	defer rows.Close()

	var suites []CheckSuite
	for rows.Next() {
		var s CheckSuite
		if err := rows.Scan(
			&s.ID, &s.RepoID, &s.SHA, &s.BaseSHA, &s.Key,
			&s.Name, &s.DetailsURL, &s.ExternalID, &s.ActorUserID,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		suites = append(suites, s)
	}
	return suites, rows.Err()
}

// UpsertCheckRun creates or updates a check run.
// Reusing an external_id within the suite updates in place.
// Different external_id creates a new attempt.
func (db *DB) UpsertCheckRun(ctx context.Context, r CheckRun) (CheckRun, string, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	if r.CreatedAt == "" {
		r.CreatedAt = now
	}
	r.UpdatedAt = now

	// Look up by suite and external_id
	var existing CheckRun
	var isRerequestableInt int
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repo_id, IFNULL(check_suite_id, ''), sha, IFNULL(base_sha, ''), IFNULL(key, ''), IFNULL(name, ''),
		       status, IFNULL(conclusion, ''), IFNULL(details_url, ''), IFNULL(external_id, ''),
		       IFNULL(output_title, ''), IFNULL(output_summary, ''), IFNULL(output_text, ''),
		       IFNULL(external_updated_at, ''), IFNULL(started_at, ''), IFNULL(completed_at, ''),
		       IFNULL(deadline_at, ''), is_rerequestable, IFNULL(rerequested_at, ''),
		       IFNULL(rerequested_by_user_id, ''), IFNULL(actor_user_id, ''),
		       created_at, updated_at
		FROM check_runs
		WHERE check_suite_id = ? AND external_id = ?
	`, r.CheckSuiteID, r.ExternalID).Scan(
		&existing.ID, &existing.RepoID, &existing.CheckSuiteID, &existing.SHA, &existing.BaseSHA, &existing.Key, &existing.Name,
		&existing.Status, &existing.Conclusion, &existing.DetailsURL, &existing.ExternalID,
		&existing.OutputTitle, &existing.OutputSummary, &existing.OutputText,
		&existing.ExternalUpdatedAt, &existing.StartedAt, &existing.CompletedAt,
		&existing.DeadlineAt, &isRerequestableInt, &existing.RerequestedAt,
		&existing.RerequestedByUserID, &existing.ActorUserID,
		&existing.CreatedAt, &existing.UpdatedAt,
	)
	existing.IsRerequestable = isRerequestableInt == 1

	if err == nil {
		// Existing run found with this external_id.
		// Staleness check against stored external_updated_at or rerequested_at
		threshold := existing.ExternalUpdatedAt
		if existing.RerequestedAt != "" && existing.RerequestedAt > threshold {
			threshold = existing.RerequestedAt
		}

		if r.ExternalUpdatedAt != "" && threshold != "" {
			if r.ExternalUpdatedAt < threshold {
				return existing, "ignored_stale", nil
			}
			if r.ExternalUpdatedAt == threshold {
				if r.Status != "completed" && existing.Status == "completed" {
					return existing, "ignored_stale", nil
				}
				if existing.RerequestedAt != "" {
					return existing, "ignored_stale", nil
				}
			}
		}

		// Check if unchanged
		if existing.Status == r.Status &&
			existing.Conclusion == r.Conclusion &&
			existing.OutputTitle == r.OutputTitle &&
			existing.OutputSummary == r.OutputSummary &&
			existing.OutputText == r.OutputText &&
			existing.StartedAt == r.StartedAt &&
			existing.CompletedAt == r.CompletedAt &&
			existing.ExternalUpdatedAt == r.ExternalUpdatedAt {
			return existing, "unchanged", nil
		}

		rerequestedAtVal := existing.RerequestedAt
		if r.ExternalUpdatedAt > threshold {
			// Newer post clears outstanding rerequest
			rerequestedAtVal = ""
		}

		isReqInt := 0
		if r.IsRerequestable {
			isReqInt = 1
		}

		_, err = db.sql.ExecContext(ctx, `
			UPDATE check_runs
			SET status = ?, conclusion = NULLIF(?, ''), details_url = ?,
			    output_title = ?, output_summary = ?, output_text = ?,
			    external_updated_at = ?, started_at = NULLIF(?, ''), completed_at = NULLIF(?, ''),
			    deadline_at = NULLIF(?, ''), is_rerequestable = ?, rerequested_at = NULLIF(?, ''),
			    actor_user_id = COALESCE(NULLIF(?, ''), actor_user_id), updated_at = ?
			WHERE id = ?
		`, r.Status, r.Conclusion, r.DetailsURL,
			r.OutputTitle, r.OutputSummary, r.OutputText,
			r.ExternalUpdatedAt, r.StartedAt, r.CompletedAt,
			r.DeadlineAt, isReqInt, rerequestedAtVal,
			r.ActorUserID, now, existing.ID)
		if err != nil {
			return CheckRun{}, "", fmt.Errorf("update check_run: %w", err)
		}

		r.ID = existing.ID
		r.CreatedAt = existing.CreatedAt
		r.UpdatedAt = now
		r.RerequestedAt = rerequestedAtVal
		r.RerequestedByUserID = existing.RerequestedByUserID
		if r.ActorUserID == "" {
			r.ActorUserID = existing.ActorUserID
		}
		return r, "updated", nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return CheckRun{}, "", fmt.Errorf("query check_run: %w", err)
	}

	// Insert new check run attempt
	if r.ID == "" {
		r.ID = idgen.NewCheckRun()
	}
	isReqInt := 0
	if r.IsRerequestable {
		isReqInt = 1
	}

	_, err = db.sql.ExecContext(ctx, `
		INSERT INTO check_runs (
			id, repo_id, check_suite_id, sha, base_sha, key, name,
			status, conclusion, details_url, external_id,
			output_title, output_summary, output_text,
			external_updated_at, started_at, completed_at, deadline_at,
			is_rerequestable, rerequested_at, rerequested_by_user_id,
			actor_user_id, created_at, updated_at
		) VALUES (
			?, ?, NULLIF(?, ''), ?, ?, ?, ?,
			?, NULLIF(?, ''), ?, ?,
			?, ?, ?,
			?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''),
			?, NULLIF(?, ''), NULLIF(?, ''),
			NULLIF(?, ''), ?, ?
		)
	`, r.ID, r.RepoID, r.CheckSuiteID, r.SHA, r.BaseSHA, r.Key, r.Name,
		r.Status, r.Conclusion, r.DetailsURL, r.ExternalID,
		r.OutputTitle, r.OutputSummary, r.OutputText,
		r.ExternalUpdatedAt, r.StartedAt, r.CompletedAt, r.DeadlineAt,
		isReqInt, r.RerequestedAt, r.RerequestedByUserID,
		r.ActorUserID, r.CreatedAt, r.UpdatedAt)
	if err != nil {
		return CheckRun{}, "", fmt.Errorf("insert check_run: %w", err)
	}

	return r, "created", nil
}

// GetCheckRun retrieves a check run by ID.
func (db *DB) GetCheckRun(ctx context.Context, id string) (CheckRun, error) {
	var r CheckRun
	var isRerequestableInt int
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repo_id, IFNULL(check_suite_id, ''), sha, IFNULL(base_sha, ''), IFNULL(key, ''), IFNULL(name, ''),
		       status, IFNULL(conclusion, ''), IFNULL(details_url, ''), IFNULL(external_id, ''),
		       IFNULL(output_title, ''), IFNULL(output_summary, ''), IFNULL(output_text, ''),
		       IFNULL(external_updated_at, ''), IFNULL(started_at, ''), IFNULL(completed_at, ''),
		       IFNULL(deadline_at, ''), is_rerequestable, IFNULL(rerequested_at, ''),
		       IFNULL(rerequested_by_user_id, ''), IFNULL(actor_user_id, ''),
		       created_at, updated_at
		FROM check_runs
		WHERE id = ?
	`, id).Scan(
		&r.ID, &r.RepoID, &r.CheckSuiteID, &r.SHA, &r.BaseSHA, &r.Key, &r.Name,
		&r.Status, &r.Conclusion, &r.DetailsURL, &r.ExternalID,
		&r.OutputTitle, &r.OutputSummary, &r.OutputText,
		&r.ExternalUpdatedAt, &r.StartedAt, &r.CompletedAt,
		&r.DeadlineAt, &isRerequestableInt, &r.RerequestedAt,
		&r.RerequestedByUserID, &r.ActorUserID,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return CheckRun{}, ErrNotFound
	}
	if err != nil {
		return CheckRun{}, fmt.Errorf("get check_run: %w", err)
	}
	r.IsRerequestable = isRerequestableInt == 1
	return r, nil
}

// ListCheckRunsForSuite retrieves all check runs belonging to a check suite.
func (db *DB) ListCheckRunsForSuite(ctx context.Context, suiteID string, limit, offset int) ([]CheckRun, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, repo_id, IFNULL(check_suite_id, ''), sha, IFNULL(base_sha, ''), IFNULL(key, ''), IFNULL(name, ''),
		       status, IFNULL(conclusion, ''), IFNULL(details_url, ''), IFNULL(external_id, ''),
		       IFNULL(output_title, ''), IFNULL(output_summary, ''), IFNULL(output_text, ''),
		       IFNULL(external_updated_at, ''), IFNULL(started_at, ''), IFNULL(completed_at, ''),
		       IFNULL(deadline_at, ''), is_rerequestable, IFNULL(rerequested_at, ''),
		       IFNULL(rerequested_by_user_id, ''), IFNULL(actor_user_id, ''),
		       created_at, updated_at
		FROM check_runs
		WHERE check_suite_id = ?
		ORDER BY created_at ASC, id ASC
		LIMIT ? OFFSET ?
	`, suiteID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list check_runs for suite: %w", err)
	}
	defer rows.Close()

	var runs []CheckRun
	for rows.Next() {
		var r CheckRun
		var isReqInt int
		if err := rows.Scan(
			&r.ID, &r.RepoID, &r.CheckSuiteID, &r.SHA, &r.BaseSHA, &r.Key, &r.Name,
			&r.Status, &r.Conclusion, &r.DetailsURL, &r.ExternalID,
			&r.OutputTitle, &r.OutputSummary, &r.OutputText,
			&r.ExternalUpdatedAt, &r.StartedAt, &r.CompletedAt,
			&r.DeadlineAt, &isReqInt, &r.RerequestedAt,
			&r.RerequestedByUserID, &r.ActorUserID,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		r.IsRerequestable = isReqInt == 1
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// ListCheckRunsForCommit retrieves check runs for a commit SHA.
// Default returns the latest run per key.
func (db *DB) ListCheckRunsForCommit(ctx context.Context, repoID, sha, checkName, status string, limit, offset int) ([]CheckRun, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, repo_id, IFNULL(check_suite_id, ''), sha, IFNULL(base_sha, ''), IFNULL(key, ''), IFNULL(name, ''),
		       status, IFNULL(conclusion, ''), IFNULL(details_url, ''), IFNULL(external_id, ''),
		       IFNULL(output_title, ''), IFNULL(output_summary, ''), IFNULL(output_text, ''),
		       IFNULL(external_updated_at, ''), IFNULL(started_at, ''), IFNULL(completed_at, ''),
		       IFNULL(deadline_at, ''), is_rerequestable, IFNULL(rerequested_at, ''),
		       IFNULL(rerequested_by_user_id, ''), IFNULL(actor_user_id, ''),
		       created_at, updated_at
		FROM check_runs
		WHERE repo_id = ? AND sha = ?
	`
	var args []any
	args = append(args, repoID, sha)

	if checkName != "" {
		query += " AND name = ?"
		args = append(args, checkName)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list check_runs for commit: %w", err)
	}
	defer rows.Close()

	var runs []CheckRun
	for rows.Next() {
		var r CheckRun
		var isReqInt int
		if err := rows.Scan(
			&r.ID, &r.RepoID, &r.CheckSuiteID, &r.SHA, &r.BaseSHA, &r.Key, &r.Name,
			&r.Status, &r.Conclusion, &r.DetailsURL, &r.ExternalID,
			&r.OutputTitle, &r.OutputSummary, &r.OutputText,
			&r.ExternalUpdatedAt, &r.StartedAt, &r.CompletedAt,
			&r.DeadlineAt, &isReqInt, &r.RerequestedAt,
			&r.RerequestedByUserID, &r.ActorUserID,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		r.IsRerequestable = isReqInt == 1
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// RerequestCheckRun flags a check run as rerequested.
// Must be completed, must be rerequestable, and not already rerequested.
func (db *DB) RerequestCheckRun(ctx context.Context, checkRunID, actingUserID string) (CheckRun, error) {
	run, err := db.GetCheckRun(ctx, checkRunID)
	if err != nil {
		return CheckRun{}, err
	}

	if run.RerequestedAt != "" {
		return CheckRun{}, fmt.Errorf("%w: check run rerequest is already outstanding", ErrAlreadyExists)
	}
	if run.Status != "completed" {
		return CheckRun{}, fmt.Errorf("%w: check run status is %q, expected 'completed'", ErrPrecondition, run.Status)
	}
	if !run.IsRerequestable {
		return CheckRun{}, fmt.Errorf("%w: check run is not rerequestable", ErrPrecondition)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.sql.ExecContext(ctx, `
		UPDATE check_runs
		SET status = 'rerequested', rerequested_at = ?, rerequested_by_user_id = NULLIF(?, ''), updated_at = ?
		WHERE id = ?
	`, now, actingUserID, now, checkRunID)
	if err != nil {
		return CheckRun{}, fmt.Errorf("update rerequest check_run: %w", err)
	}

	run.Status = "rerequested"
	run.RerequestedAt = now
	run.RerequestedByUserID = actingUserID
	run.UpdatedAt = now
	return run, nil
}

// CreateCheckRunAnnotations appends annotations to a check run.
func (db *DB) CreateCheckRunAnnotations(ctx context.Context, checkRunID string, annotations []CheckRunAnnotation) ([]CheckRunAnnotation, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var created []CheckRunAnnotation

	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO check_run_annotations (
			id, check_run_id, path, start_line, end_line, start_column, end_column,
			annotation_level, message, title, raw_details, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare annotation stmt: %w", err)
	}
	defer stmt.Close()

	for _, a := range annotations {
		if a.ID == "" {
			a.ID = idgen.NewCheckRunAnnotation()
		}
		a.CheckRunID = checkRunID
		a.CreatedAt = now
		a.UpdatedAt = now

		_, err := stmt.ExecContext(ctx,
			a.ID, a.CheckRunID, a.Path, a.StartLine, a.EndLine, a.StartColumn, a.EndColumn,
			a.AnnotationLevel, a.Message, a.Title, a.RawDetails, a.CreatedAt, a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("insert annotation: %w", err)
		}
		created = append(created, a)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return created, nil
}

// ListCheckRunAnnotations retrieves annotations for a check run.
func (db *DB) ListCheckRunAnnotations(ctx context.Context, checkRunID string, limit, offset int) ([]CheckRunAnnotation, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, check_run_id, IFNULL(path, ''), IFNULL(start_line, 0), IFNULL(end_line, 0),
		       IFNULL(start_column, 0), IFNULL(end_column, 0), IFNULL(annotation_level, ''),
		       IFNULL(message, ''), IFNULL(title, ''), IFNULL(raw_details, ''),
		       created_at, IFNULL(updated_at, created_at)
		FROM check_run_annotations
		WHERE check_run_id = ?
		ORDER BY created_at ASC, id ASC
		LIMIT ? OFFSET ?
	`, checkRunID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list annotations: %w", err)
	}
	defer rows.Close()

	var list []CheckRunAnnotation
	for rows.Next() {
		var a CheckRunAnnotation
		if err := rows.Scan(
			&a.ID, &a.CheckRunID, &a.Path, &a.StartLine, &a.EndLine,
			&a.StartColumn, &a.EndColumn, &a.AnnotationLevel,
			&a.Message, &a.Title, &a.RawDetails,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}
