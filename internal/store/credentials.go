package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// APIKey represents a stored high-entropy API key.
type APIKey struct {
	ID         string
	UserID     string
	KeyHash    string
	KeyPrefix  string
	Name       string
	Scopes     string
	CreatedAt  string
	ExpiresAt  sql.NullString
	RevokedAt  sql.NullString
	LastUsedAt sql.NullString
}

// AccessToken represents an active session or exchanged token.
type AccessToken struct {
	ID        string
	TokenHash string
	UserID    string
	APIKeyID  sql.NullString
	TokenType string
	Scopes    string
	CreatedAt string
	ExpiresAt string
	RevokedAt sql.NullString
}

// CreateAPIKey stores a new API key record.
func (db *DB) CreateAPIKey(ctx context.Context, k APIKey) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if k.CreatedAt == "" {
		k.CreatedAt = now
	}
	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO api_keys (id, user_id, key_hash, key_prefix, name, scopes, created_at, expires_at, revoked_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		k.ID, k.UserID, k.KeyHash, k.KeyPrefix, k.Name, k.Scopes, k.CreatedAt,
		k.ExpiresAt, k.RevokedAt, k.LastUsedAt,
	)
	if err != nil {
		if isConflict(err) {
			return fmt.Errorf("%w: api_key", ErrConflict)
		}
		return err
	}
	return nil
}

// GetAPIKeyByHash retrieves an API key by its hash.
func (db *DB) GetAPIKeyByHash(ctx context.Context, hash string) (APIKey, error) {
	var k APIKey
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, user_id, key_hash, key_prefix, name, scopes, created_at, expires_at, revoked_at, last_used_at
		FROM api_keys WHERE key_hash = ?`, hash).Scan(
		&k.ID, &k.UserID, &k.KeyHash, &k.KeyPrefix, &k.Name, &k.Scopes,
		&k.CreatedAt, &k.ExpiresAt, &k.RevokedAt, &k.LastUsedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return APIKey{}, ErrNotFound
	}
	return k, err
}

// RevokeAPIKey marks an API key as revoked.
func (db *DB) RevokeAPIKey(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := db.sql.ExecContext(ctx, `UPDATE api_keys SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL`, now, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListAPIKeysByUserID returns all active API keys for a user.
func (db *DB) ListAPIKeysByUserID(ctx context.Context, userID string) ([]APIKey, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, user_id, key_hash, key_prefix, name, scopes, created_at, expires_at, revoked_at, last_used_at
		FROM api_keys
		WHERE user_id = ? AND revoked_at IS NULL
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.KeyHash, &k.KeyPrefix, &k.Name, &k.Scopes,
			&k.CreatedAt, &k.ExpiresAt, &k.RevokedAt, &k.LastUsedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// TouchAPIKey updates the last_used_at timestamp on an API key.
func (db *DB) TouchAPIKey(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := db.sql.ExecContext(ctx, `UPDATE api_keys SET last_used_at = ? WHERE id = ?`, now, id)
	return err
}

// CreateAccessToken stores a newly issued access token.
func (db *DB) CreateAccessToken(ctx context.Context, tok AccessToken) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if tok.CreatedAt == "" {
		tok.CreatedAt = now
	}
	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO access_tokens (id, token_hash, user_id, api_key_id, token_type, scopes, created_at, expires_at, revoked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tok.ID, tok.TokenHash, tok.UserID, tok.APIKeyID, tok.TokenType, tok.Scopes,
		tok.CreatedAt, tok.ExpiresAt, tok.RevokedAt,
	)
	if err != nil {
		if isConflict(err) {
			return fmt.Errorf("%w: access_token", ErrConflict)
		}
		return err
	}
	return nil
}

// GetAccessTokenByHash retrieves an access token by its hash.
func (db *DB) GetAccessTokenByHash(ctx context.Context, hash string) (AccessToken, error) {
	var tok AccessToken
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, token_hash, user_id, api_key_id, token_type, scopes, created_at, expires_at, revoked_at
		FROM access_tokens WHERE token_hash = ?`, hash).Scan(
		&tok.ID, &tok.TokenHash, &tok.UserID, &tok.APIKeyID, &tok.TokenType, &tok.Scopes,
		&tok.CreatedAt, &tok.ExpiresAt, &tok.RevokedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return AccessToken{}, ErrNotFound
	}
	return tok, err
}

// CountUsers reports the total number of registered users. Used during bootstrap.
func (db *DB) CountUsers(ctx context.Context) (int, error) {
	var count int
	err := db.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}
