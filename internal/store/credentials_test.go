package store

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestCredentialsStore(t *testing.T) {
	ctx := context.Background()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// Verify migrations reached version 2
	v, err := db.version()
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if v < 2 {
		t.Fatalf("expected migration version >= 2, got %d", v)
	}

	// Create user
	userID := "user_test1"
	if _, err := db.sql.Exec(`INSERT INTO users (id, email, created_at) VALUES (?, ?, ?)`,
		userID, "admin@example.com", time.Now().UTC().Format(time.RFC3339)); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	count, err := db.CountUsers(ctx)
	if err != nil {
		t.Fatalf("CountUsers: %v", err)
	}
	if count != 1 {
		t.Fatalf("CountUsers = %d, want 1", count)
	}

	// Create API Key
	apiKey := APIKey{
		ID:        "key_01",
		UserID:    userID,
		KeyHash:   "hash_abc123",
		KeyPrefix: "key_abc12345",
		Name:      "admin-key",
		Scopes:    "*",
	}
	if err := db.CreateAPIKey(ctx, apiKey); err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}

	// Fetch API Key
	fetched, err := db.GetAPIKeyByHash(ctx, "hash_abc123")
	if err != nil {
		t.Fatalf("GetAPIKeyByHash: %v", err)
	}
	if fetched.ID != "key_01" || fetched.UserID != userID || fetched.Name != "admin-key" {
		t.Fatalf("unexpected fetched key: %+v", fetched)
	}

	// Touch API Key
	if err := db.TouchAPIKey(ctx, "key_01"); err != nil {
		t.Fatalf("TouchAPIKey: %v", err)
	}
	fetched, _ = db.GetAPIKeyByHash(ctx, "hash_abc123")
	if !fetched.LastUsedAt.Valid || fetched.LastUsedAt.String == "" {
		t.Fatalf("expected valid last_used_at after touch")
	}

	// Revoke API Key
	if err := db.RevokeAPIKey(ctx, "key_01"); err != nil {
		t.Fatalf("RevokeAPIKey: %v", err)
	}
	fetched, _ = db.GetAPIKeyByHash(ctx, "hash_abc123")
	if !fetched.RevokedAt.Valid || fetched.RevokedAt.String == "" {
		t.Fatalf("expected valid revoked_at after revoke")
	}

	// Create Access Token
	now := time.Now().UTC()
	tok := AccessToken{
		ID:        "tok_01",
		TokenHash: "tokhash_xyz789",
		UserID:    userID,
		APIKeyID:  sql.NullString{String: "key_01", Valid: true},
		TokenType: "user",
		Scopes:    "repository:metadata:read namespace:repositories:read",
		ExpiresAt: now.Add(24 * time.Hour).Format(time.RFC3339),
	}
	if err := db.CreateAccessToken(ctx, tok); err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}

	fetchedTok, err := db.GetAccessTokenByHash(ctx, "tokhash_xyz789")
	if err != nil {
		t.Fatalf("GetAccessTokenByHash: %v", err)
	}
	if fetchedTok.ID != "tok_01" || fetchedTok.UserID != userID || fetchedTok.TokenType != "user" {
		t.Fatalf("unexpected fetched token: %+v", fetchedTok)
	}
}
