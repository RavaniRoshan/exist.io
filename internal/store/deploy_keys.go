package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"exist.io/pkg/idgen"
	"golang.org/x/crypto/ssh"
)

// RepoDeployKey represents an SSH public key granted read or read-write access to a specific repository.
type RepoDeployKey struct {
	ID          string `json:"id"`
	RepoID      string `json:"repo_id"`
	Title       string `json:"title"`
	PublicKey   string `json:"key"`
	Fingerprint string `json:"fingerprint"`
	ReadOnly    bool   `json:"read_only"`
	LastUsedAt  string `json:"last_used_at,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// CreateDeployKey validates an SSH public key, computes its SHA256 fingerprint, and records it for a repository.
func (db *DB) CreateDeployKey(ctx context.Context, repoID, title, rawKey string, readOnly bool) (*RepoDeployKey, error) {
	trimmed := strings.TrimSpace(rawKey)
	if trimmed == "" {
		return nil, errors.New("public key cannot be empty")
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(trimmed))
	if err != nil {
		return nil, fmt.Errorf("invalid ssh public key format: %w", err)
	}

	fingerprint := ssh.FingerprintSHA256(pubKey)
	id := idgen.New("depk")
	now := time.Now().UTC().Format(time.RFC3339)

	readOnlyInt := 0
	if readOnly {
		readOnlyInt = 1
	}

	_, err = db.sql.ExecContext(ctx, `
		INSERT INTO repo_deploy_keys (id, repo_id, title, public_key, fingerprint, read_only, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, repoID, title, trimmed, fingerprint, readOnlyInt, now)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, errors.New("key is already in use")
		}
		return nil, fmt.Errorf("insert deploy key: %w", err)
	}

	return &RepoDeployKey{
		ID:          id,
		RepoID:      repoID,
		Title:       title,
		PublicKey:   trimmed,
		Fingerprint: fingerprint,
		ReadOnly:    readOnly,
		CreatedAt:   now,
	}, nil
}

// ListDeployKeys returns all deploy keys attached to a repository.
func (db *DB) ListDeployKeys(ctx context.Context, repoID string) ([]*RepoDeployKey, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, repo_id, title, public_key, fingerprint, read_only, COALESCE(last_used_at, ''), created_at
		FROM repo_deploy_keys
		WHERE repo_id = ?
		ORDER BY created_at DESC
	`, repoID)
	if err != nil {
		return nil, fmt.Errorf("list deploy keys: %w", err)
	}
	defer rows.Close()

	var keys []*RepoDeployKey
	for rows.Next() {
		var k RepoDeployKey
		var readOnlyInt int
		if err := rows.Scan(&k.ID, &k.RepoID, &k.Title, &k.PublicKey, &k.Fingerprint, &readOnlyInt, &k.LastUsedAt, &k.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan deploy key: %w", err)
		}
		k.ReadOnly = readOnlyInt == 1
		keys = append(keys, &k)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

// GetDeployKey returns a single deploy key by its ID.
func (db *DB) GetDeployKey(ctx context.Context, id string) (*RepoDeployKey, error) {
	var k RepoDeployKey
	var readOnlyInt int
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repo_id, title, public_key, fingerprint, read_only, COALESCE(last_used_at, ''), created_at
		FROM repo_deploy_keys
		WHERE id = ?
	`, id).Scan(&k.ID, &k.RepoID, &k.Title, &k.PublicKey, &k.Fingerprint, &readOnlyInt, &k.LastUsedAt, &k.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get deploy key: %w", err)
	}
	k.ReadOnly = readOnlyInt == 1
	return &k, nil
}

// GetDeployKeyByFingerprint looks up a deploy key by its SHA256 fingerprint, along with its associated Repo.
func (db *DB) GetDeployKeyByFingerprint(ctx context.Context, fingerprint string) (*RepoDeployKey, *Repo, error) {
	var k RepoDeployKey
	var readOnlyInt int
	var repoID string

	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repo_id, title, public_key, fingerprint, read_only, COALESCE(last_used_at, ''), created_at
		FROM repo_deploy_keys
		WHERE fingerprint = ?
	`, fingerprint).Scan(&k.ID, &repoID, &k.Title, &k.PublicKey, &k.Fingerprint, &readOnlyInt, &k.LastUsedAt, &k.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, fmt.Errorf("get deploy key by fingerprint: %w", err)
	}
	k.ReadOnly = readOnlyInt == 1
	k.RepoID = repoID

	repo, err := db.GetRepoByID(ctx, repoID)
	if err != nil {
		return nil, nil, fmt.Errorf("get deploy key repo: %w", err)
	}

	return &k, &repo, nil
}

// DeleteDeployKey removes a deploy key by ID.
func (db *DB) DeleteDeployKey(ctx context.Context, id string) error {
	res, err := db.sql.ExecContext(ctx, `DELETE FROM repo_deploy_keys WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete deploy key: %w", err)
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

// TouchDeployKey records the current timestamp as last_used_at.
func (db *DB) TouchDeployKey(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.sql.ExecContext(ctx, `UPDATE repo_deploy_keys SET last_used_at = ? WHERE id = ?`, now, id)
	return err
}
