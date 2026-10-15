package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"exist.io/pkg/idgen"
)

// Release represents a Git release/tag with notes, assets, and publish metadata.
type Release struct {
	ID          string `json:"id"`
	RepoID      string `json:"repoId"`
	TagName     string `json:"tagName"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	Author      string `json:"author"`
	CommitSHA   string `json:"commitSha"`
	CreatedAt   string `json:"createdAt"`
	PublishedAt string `json:"publishedAt"`
}

// CreateRelease inserts a new release record into SQLite WAL.
func (db *DB) CreateRelease(ctx context.Context, r Release) (Release, error) {
	if r.ID == "" {
		r.ID = idgen.NewRelease()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if r.CreatedAt == "" {
		r.CreatedAt = now
	}
	if r.PublishedAt == "" {
		r.PublishedAt = now
	}

	draftInt := 0
	if r.Draft {
		draftInt = 1
	}
	preInt := 0
	if r.Prerelease {
		preInt = 1
	}

	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO releases (
			id, repo_id, tag_name, name, body, draft, prerelease, author, commit_sha, created_at, published_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, r.ID, r.RepoID, r.TagName, r.Name, r.Body, draftInt, preInt, r.Author, r.CommitSHA, r.CreatedAt, r.PublishedAt)
	if err != nil {
		return Release{}, fmt.Errorf("insert release: %w", err)
	}
	return r, nil
}

// ListReleases returns all releases for a repository sorted newest published first.
func (db *DB) ListReleases(ctx context.Context, repoID string) ([]Release, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, repo_id, tag_name, name, body, draft, prerelease, author, commit_sha, created_at, published_at
		FROM releases
		WHERE repo_id = ?
		ORDER BY published_at DESC, id DESC
	`, repoID)
	if err != nil {
		return nil, fmt.Errorf("list releases: %w", err)
	}
	defer rows.Close()

	var list []Release
	for rows.Next() {
		var r Release
		var draftInt, preInt int
		if err := rows.Scan(
			&r.ID, &r.RepoID, &r.TagName, &r.Name, &r.Body,
			&draftInt, &preInt, &r.Author, &r.CommitSHA, &r.CreatedAt, &r.PublishedAt,
		); err != nil {
			return nil, err
		}
		r.Draft = draftInt == 1
		r.Prerelease = preInt == 1
		list = append(list, r)
	}
	if list == nil {
		list = []Release{}
	}
	return list, rows.Err()
}

// GetRelease retrieves a single release by its ID.
func (db *DB) GetRelease(ctx context.Context, id string) (Release, error) {
	var r Release
	var draftInt, preInt int
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repo_id, tag_name, name, body, draft, prerelease, author, commit_sha, created_at, published_at
		FROM releases
		WHERE id = ?
	`, id).Scan(
		&r.ID, &r.RepoID, &r.TagName, &r.Name, &r.Body,
		&draftInt, &preInt, &r.Author, &r.CommitSHA, &r.CreatedAt, &r.PublishedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Release{}, ErrNotFound
	}
	if err != nil {
		return Release{}, fmt.Errorf("get release: %w", err)
	}
	r.Draft = draftInt == 1
	r.Prerelease = preInt == 1
	return r, nil
}

// DeleteRelease removes a release by ID.
func (db *DB) DeleteRelease(ctx context.Context, id string) error {
	res, err := db.sql.ExecContext(ctx, `DELETE FROM releases WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete release: %w", err)
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
