package store_test

import (
	"context"
	"path/filepath"
	"testing"

	"exist.io/internal/store"
)

func TestAgentLeases_Revocation(t *testing.T) {
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	leaseID := "lease_test_123"

	// 1. Initial lease is not revoked
	revoked, err := db.IsAgentLeaseRevoked(ctx, leaseID)
	if err != nil {
		t.Fatalf("IsAgentLeaseRevoked: %v", err)
	}
	if revoked {
		t.Fatal("expected lease to not be revoked")
	}

	// 2. Revoke lease
	err = db.RevokeAgentLease(ctx, leaseID, "agent-claude", "repo-1", "work completed")
	if err != nil {
		t.Fatalf("RevokeAgentLease: %v", err)
	}

	// 3. Now it should be revoked
	revoked, err = db.IsAgentLeaseRevoked(ctx, leaseID)
	if err != nil {
		t.Fatalf("IsAgentLeaseRevoked: %v", err)
	}
	if !revoked {
		t.Fatal("expected lease to be marked as revoked")
	}
}
