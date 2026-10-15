package store

import (
	"context"
	"testing"
)

func TestUsersAndSessionsStore(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Create User
	u := User{
		ID:           "user_test_admin",
		Email:        "admin@exist.io",
		Login:        "admin",
		Name:         "Cluster Admin",
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=2$mockhash",
		Role:         "admin",
		Status:       "active",
	}
	if err := db.CreateUserWithPassword(ctx, u); err != nil {
		t.Fatalf("CreateUserWithPassword: %v", err)
	}

	// 2. Fetch User by Email and Login
	byEmail, err := db.GetUserByEmail(ctx, "admin@exist.io")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if byEmail.Role != "admin" || byEmail.Login != "admin" {
		t.Errorf("expected role=admin login=admin, got role=%s login=%s", byEmail.Role, byEmail.Login)
	}

	byLogin, err := db.GetUserByLogin(ctx, "admin")
	if err != nil {
		t.Fatalf("GetUserByLogin: %v", err)
	}
	if byLogin.ID != u.ID {
		t.Errorf("expected ID=%s, got %s", u.ID, byLogin.ID)
	}

	// 3. List Users
	users, err := db.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) < 1 {
		t.Fatalf("expected at least 1 user, got %d", len(users))
	}

	// 4. Session Lifecycle
	sess := Session{
		ID:        "sess_01M4TEST",
		UserID:    u.ID,
		TokenHash: "mock_session_hash_1234",
		CSRFToken: "csrf_secret_token_abcd",
		IPAddress: "127.0.0.1",
		UserAgent: "Mozilla/5.0",
		ExpiresAt: "2099-01-01T00:00:00Z",
	}
	if err := db.CreateSession(ctx, sess); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	s, userFromSess, err := db.GetSessionByHash(ctx, sess.TokenHash)
	if err != nil {
		t.Fatalf("GetSessionByHash: %v", err)
	}
	if s.CSRFToken != sess.CSRFToken {
		t.Errorf("expected csrf=%s, got %s", sess.CSRFToken, s.CSRFToken)
	}
	if userFromSess.ID != u.ID {
		t.Errorf("expected user=%s, got %s", u.ID, userFromSess.ID)
	}

	if err := db.DeleteSession(ctx, sess.ID); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if _, _, err := db.GetSessionByHash(ctx, sess.TokenHash); err != ErrNotFound {
		t.Errorf("expected ErrNotFound after DeleteSession, got %v", err)
	}

	// 5. SSH Key Lifecycle
	k := SSHKey{
		ID:          "sshkey_01M4TEST",
		UserID:      u.ID,
		Name:        "Personal Laptop",
		Fingerprint: "SHA256:abcd1234mockfingerprint",
		PublicKey:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMockKey user@laptop",
		KeyType:     "ssh-ed25519",
	}
	if err := db.AddUserSSHKey(ctx, k); err != nil {
		t.Fatalf("AddUserSSHKey: %v", err)
	}

	keys, err := db.ListUserSSHKeys(ctx, u.ID)
	if err != nil {
		t.Fatalf("ListUserSSHKeys: %v", err)
	}
	if len(keys) != 1 || keys[0].Name != "Personal Laptop" {
		t.Errorf("expected 1 key named 'Personal Laptop', got %+v", keys)
	}

	foundKey, keyOwner, err := db.GetSSHKeyByFingerprint(ctx, k.Fingerprint)
	if err != nil {
		t.Fatalf("GetSSHKeyByFingerprint: %v", err)
	}
	if foundKey.ID != k.ID || keyOwner.ID != u.ID {
		t.Errorf("foundKey or keyOwner mismatch: key=%+v owner=%+v", foundKey, keyOwner)
	}

	if err := db.DeleteUserSSHKey(ctx, u.ID, k.ID); err != nil {
		t.Fatalf("DeleteUserSSHKey: %v", err)
	}
}
