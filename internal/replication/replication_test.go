package replication_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"exist.io/internal/replication"
	"exist.io/internal/store"
)

func TestReplicationController_Lifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "meta.db")

	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer db.Close()

	replicaDir := filepath.Join(tmpDir, "replica")

	ctrl, err := replication.NewController(replication.Config{
		NodeID:       "test-primary-01",
		Role:         replication.RolePrimary,
		DataDir:      tmpDir,
		ReplicaDir:   replicaDir,
		SyncInterval: 100 * time.Millisecond,
		DB:           db,
	})
	if err != nil {
		t.Fatalf("failed to create replication controller: %v", err)
	}

	ctx := context.Background()
	if err := ctrl.Start(ctx); err != nil {
		t.Fatalf("failed to start controller: %v", err)
	}
	defer ctrl.Stop()

	// 1. Initial status checks
	status := ctrl.Status()
	if status.NodeID != "test-primary-01" {
		t.Errorf("NodeID = %s, want test-primary-01", status.NodeID)
	}
	if status.Role != replication.RolePrimary {
		t.Errorf("Role = %s, want primary", status.Role)
	}
	if status.Health != "healthy" {
		t.Errorf("Health = %s, want healthy", status.Health)
	}

	// 2. Explicit Checkpoint
	if err := ctrl.Checkpoint(ctx); err != nil {
		t.Fatalf("Checkpoint failed: %v", err)
	}

	// 3. Create point-in-time snapshot
	meta, err := ctrl.CreateSnapshot(ctx)
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	if meta.Seq != 1 {
		t.Errorf("Snapshot Seq = %d, want 1", meta.Seq)
	}
	if meta.ChecksumSHA256 == "" {
		t.Errorf("expected non-empty checksum")
	}

	snapshotFile := filepath.Join(tmpDir, "snapshots", meta.Filename)
	f, err := os.Open(snapshotFile)
	if err != nil {
		t.Fatalf("snapshot file missing: %v", err)
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		t.Fatalf("failed to hash snapshot: %v", err)
	}
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	if actualChecksum != meta.ChecksumSHA256 {
		t.Errorf("Checksum mismatch: got %s, want %s", actualChecksum, meta.ChecksumSHA256)
	}

	// 4. Sync to replica
	if err := ctrl.SyncToReplica(ctx); err != nil {
		t.Fatalf("SyncToReplica failed: %v", err)
	}

	replicaSnapshot := filepath.Join(replicaDir, meta.Filename)
	if _, err := os.Stat(replicaSnapshot); err != nil {
		t.Fatalf("replica snapshot missing in replicaDir: %v", err)
	}

	// 5. Restore from snapshot into fresh directory
	restoreDir := t.TempDir()
	restoredDBPath := filepath.Join(restoreDir, "meta.db")

	if err := replication.RestoreFromSnapshot(replicaSnapshot, restoredDBPath); err != nil {
		t.Fatalf("RestoreFromSnapshot failed: %v", err)
	}

	restoredDB, err := store.Open(restoredDBPath)
	if err != nil {
		t.Fatalf("failed to open restored database: %v", err)
	}
	defer restoredDB.Close()

	// Verify restored db can execute queries
	var count int
	if err := restoredDB.SQL().QueryRow("SELECT COUNT(*) FROM namespaces").Scan(&count); err != nil {
		t.Fatalf("failed to query restored db: %v", err)
	}
}

func TestReplicationController_StandbyPromotion(t *testing.T) {
	tmpDir := t.TempDir()
	ctrl, err := replication.NewController(replication.Config{
		NodeID:          "test-standby-02",
		Role:            replication.RoleStandby,
		DataDir:         tmpDir,
		PrimaryEndpoint: "http://primary.corp.exist.io:8080",
	})
	if err != nil {
		t.Fatalf("failed to create standby controller: %v", err)
	}

	if ctrl.GetRole() != replication.RoleStandby {
		t.Errorf("Role = %s, want standby", ctrl.GetRole())
	}
	if ctrl.Status().PrimaryEndpoint != "http://primary.corp.exist.io:8080" {
		t.Errorf("PrimaryEndpoint = %s, want http://primary.corp.exist.io:8080", ctrl.Status().PrimaryEndpoint)
	}

	// Promote to primary
	ctx := context.Background()
	if err := ctrl.PromoteToPrimary(ctx); err != nil {
		t.Fatalf("PromoteToPrimary failed: %v", err)
	}

	if ctrl.GetRole() != replication.RolePrimary {
		t.Errorf("After promotion Role = %s, want primary", ctrl.GetRole())
	}
	if ctrl.Status().PrimaryEndpoint != "" {
		t.Errorf("PrimaryEndpoint after promotion should be empty, got %s", ctrl.Status().PrimaryEndpoint)
	}
}
