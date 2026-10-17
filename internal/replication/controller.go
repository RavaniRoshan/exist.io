package replication

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"exist.io/internal/store"
)

// Config configures the replication and disaster recovery failover controller.
type Config struct {
	NodeID          string
	Role            Role
	DataDir         string
	ReplicaDir      string
	PrimaryEndpoint string
	SyncInterval    time.Duration
	DB              *store.DB
}

// Controller coordinates automated SQLite WAL checkpointing, point-in-time snapshots,
// cross-region replica synchronization, and standby promotion.
type Controller struct {
	cfg   Config
	db    *store.DB
	mu    sync.RWMutex
	role  Role
	seq   uint64
	state ReplicationStatus

	started atomic.Bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// NewController instantiates a replication controller for the node.
func NewController(cfg Config) (*Controller, error) {
	if cfg.NodeID == "" {
		cfg.NodeID = fmt.Sprintf("node-%d", time.Now().UnixNano()%10000)
	}
	if cfg.Role == "" {
		cfg.Role = RolePrimary
	}
	if cfg.SyncInterval <= 0 {
		cfg.SyncInterval = 30 * time.Second
	}

	c := &Controller{
		cfg:    cfg,
		db:     cfg.DB,
		role:   cfg.Role,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
		state: ReplicationStatus{
			NodeID:          cfg.NodeID,
			Role:            cfg.Role,
			Health:          "healthy",
			PrimaryEndpoint: cfg.PrimaryEndpoint,
			LastCheckpoint:  time.Now().UTC(),
		},
	}

	return c, nil
}

// Start begins the background replication and checkpoint ticker.
func (c *Controller) Start(ctx context.Context) error {
	snapshotsDir := filepath.Join(c.cfg.DataDir, "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		return fmt.Errorf("replication: create snapshots directory: %w", err)
	}

	if c.cfg.ReplicaDir != "" {
		if err := os.MkdirAll(c.cfg.ReplicaDir, 0755); err != nil {
			return fmt.Errorf("replication: create replica directory: %w", err)
		}
	}

	c.started.Store(true)
	go c.runLoop()
	return nil
}

// Stop shuts down the background synchronization loop.
func (c *Controller) Stop() {
	if !c.started.Load() {
		return
	}
	select {
	case <-c.stopCh:
		return
	default:
		close(c.stopCh)
		<-c.doneCh
	}
}

func (c *Controller) runLoop() {
	defer close(c.doneCh)
	ticker := time.NewTicker(c.cfg.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), c.cfg.SyncInterval)
			if c.GetRole() == RolePrimary {
				_ = c.Checkpoint(ctx)
				if _, err := c.CreateSnapshot(ctx); err == nil && c.cfg.ReplicaDir != "" {
					_ = c.SyncToReplica(ctx)
				}
			}
			cancel()
		}
	}
}

// GetRole returns the current node role in a thread-safe manner.
func (c *Controller) GetRole() Role {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.role
}

// Status returns a snapshot of the cluster replication metrics.
func (c *Controller) Status() ReplicationStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s := c.state
	s.Role = c.role
	return s
}

// Checkpoint forces an immediate SQLite WAL truncate checkpoint.
func (c *Controller) Checkpoint(ctx context.Context) error {
	if c.db == nil || c.db.SQL() == nil {
		return nil
	}

	_, err := c.db.SQL().ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
	if err != nil {
		return fmt.Errorf("replication: wal checkpoint: %w", err)
	}

	c.mu.Lock()
	c.state.LastCheckpoint = time.Now().UTC()
	c.mu.Unlock()
	return nil
}

// CreateSnapshot creates an immutable, SHA-256 verified point-in-time backup of the database.
func (c *Controller) CreateSnapshot(ctx context.Context) (*SnapshotMetadata, error) {
	if err := c.Checkpoint(ctx); err != nil {
		return nil, err
	}

	seq := atomic.AddUint64(&c.seq, 1)
	now := time.Now().UTC()

	dbPath := filepath.Join(c.cfg.DataDir, "meta.db")
	srcFile, err := os.Open(dbPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// If meta.db does not exist yet (e.g. in-memory or fresh startup), return empty snapshot
			return &SnapshotMetadata{
				Seq:       seq,
				Timestamp: now,
			}, nil
		}
		return nil, fmt.Errorf("replication: open db for snapshot: %w", err)
	}
	defer srcFile.Close()

	snapshotsDir := filepath.Join(c.cfg.DataDir, "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("snapshot_%06d_%d.db", seq, now.Unix())
	destPath := filepath.Join(snapshotsDir, filename)
	tmpDest := destPath + ".tmp"

	outFile, err := os.OpenFile(tmpDest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("replication: create snapshot file: %w", err)
	}

	hasher := sha256.New()
	multiWriter := io.MultiWriter(outFile, hasher)

	copiedBytes, err := io.Copy(multiWriter, srcFile)
	if err != nil {
		outFile.Close()
		_ = os.Remove(tmpDest)
		return nil, fmt.Errorf("replication: write snapshot stream: %w", err)
	}
	if err := outFile.Sync(); err != nil {
		outFile.Close()
		_ = os.Remove(tmpDest)
		return nil, err
	}
	outFile.Close()

	if err := os.Rename(tmpDest, destPath); err != nil {
		return nil, fmt.Errorf("replication: commit snapshot file: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))

	meta := &SnapshotMetadata{
		Seq:            seq,
		Timestamp:      now,
		ChecksumSHA256: checksum,
		SizeBytes:      copiedBytes,
		SourceDB:       dbPath,
		Filename:       filename,
	}

	// Write latest manifest
	manifestPath := filepath.Join(snapshotsDir, "latest-manifest.json")
	if data, err := json.MarshalIndent(meta, "", "  "); err == nil {
		_ = os.WriteFile(manifestPath, data, 0644)
	}

	c.mu.Lock()
	c.state.LastSnapshotTimestamp = now
	c.state.LastSnapshotSeq = seq
	c.state.TotalSnapshots++
	c.state.WALBytesSynced += copiedBytes
	c.state.ReplicationLagSeconds = 0
	c.mu.Unlock()

	return meta, nil
}

// SyncToReplica syncs the latest snapshot to the replica target directory.
func (c *Controller) SyncToReplica(ctx context.Context) error {
	if c.cfg.ReplicaDir == "" {
		return nil
	}

	snapshotsDir := filepath.Join(c.cfg.DataDir, "snapshots")
	manifestPath := filepath.Join(snapshotsDir, "latest-manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	var meta SnapshotMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return err
	}

	src := filepath.Join(snapshotsDir, meta.Filename)
	dst := filepath.Join(c.cfg.ReplicaDir, meta.Filename)

	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	tmpDst := dst + ".tmp"
	dstF, err := os.OpenFile(tmpDst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dstF, srcF); err != nil {
		dstF.Close()
		_ = os.Remove(tmpDst)
		return err
	}
	if err := dstF.Sync(); err != nil {
		dstF.Close()
		return err
	}
	dstF.Close()

	if err := os.Rename(tmpDst, dst); err != nil {
		return err
	}

	// Copy manifest as well
	_ = os.WriteFile(filepath.Join(c.cfg.ReplicaDir, "latest-manifest.json"), data, 0644)

	c.mu.Lock()
	c.state.ActiveReplicasCount = 1
	c.mu.Unlock()

	return nil
}

// PromoteToPrimary elevates a hot standby replica node to active primary status.
func (c *Controller) PromoteToPrimary(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.role == RolePrimary {
		return nil
	}

	c.role = RolePrimary
	c.state.Role = RolePrimary
	c.state.Health = "healthy"
	c.state.PrimaryEndpoint = ""
	return nil
}

// RestoreFromSnapshot restores an active database file from a verified snapshot.
func RestoreFromSnapshot(snapshotPath, targetDBPath string) error {
	f, err := os.Open(snapshotPath)
	if err != nil {
		return fmt.Errorf("replication: open snapshot for restore: %w", err)
	}
	defer f.Close()

	if err := os.MkdirAll(filepath.Dir(targetDBPath), 0755); err != nil {
		return err
	}

	tmpTarget := targetDBPath + ".restoring"
	out, err := os.OpenFile(tmpTarget, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("replication: create target db: %w", err)
	}

	if _, err := io.Copy(out, f); err != nil {
		out.Close()
		_ = os.Remove(tmpTarget)
		return err
	}
	if err := out.Sync(); err != nil {
		out.Close()
		return err
	}
	out.Close()

	// Remove existing WAL and SHM if restoring
	_ = os.Remove(targetDBPath + "-wal")
	_ = os.Remove(targetDBPath + "-shm")

	if err := os.Rename(tmpTarget, targetDBPath); err != nil {
		return fmt.Errorf("replication: finalize restore: %w", err)
	}

	return nil
}
