package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// LFSObject represents a Git LFS binary object stored on disk.
type LFSObject struct {
	OID         string `json:"oid"`
	Size        int64  `json:"size"`
	StoragePath string `json:"storage_path"`
	CreatedAt   string `json:"created_at"`
}

// LFSLock represents a Git LFS file lock.
type LFSLock struct {
	ID           string `json:"id"`
	RepositoryID string `json:"repository_id"`
	Path         string `json:"path"`
	OwnerName    string `json:"owner_name"`
	LockedAt     string `json:"locked_at"`
}

var (
	ErrLFSNotFound     = errors.New("lfs object not found")
	ErrLFSLockConflict = errors.New("lfs lock already exists for this path")
	ErrLFSLockNotFound = errors.New("lfs lock not found")
)

// GetLFSObject retrieves metadata for an LFS object by its SHA-256 OID.
func (db *DB) GetLFSObject(ctx context.Context, oid string) (LFSObject, error) {
	row := db.sql.QueryRowContext(ctx, `
		SELECT oid, size, storage_path, created_at
		FROM lfs_objects
		WHERE oid = ?
	`, oid)

	var obj LFSObject
	err := row.Scan(&obj.OID, &obj.Size, &obj.StoragePath, &obj.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return LFSObject{}, ErrLFSNotFound
	}
	if err != nil {
		return LFSObject{}, fmt.Errorf("get lfs object: %w", err)
	}
	return obj, nil
}

// PutLFSObject records a new LFS object or updates its storage path.
func (db *DB) PutLFSObject(ctx context.Context, obj LFSObject) error {
	if obj.CreatedAt == "" {
		obj.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO lfs_objects (oid, size, storage_path, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(oid) DO UPDATE SET
			size = excluded.size,
			storage_path = excluded.storage_path
	`, obj.OID, obj.Size, obj.StoragePath, obj.CreatedAt)
	if err != nil {
		return fmt.Errorf("put lfs object: %w", err)
	}
	return nil
}

// HasLFSObject checks if an object exists by OID and matches expected size.
func (db *DB) HasLFSObject(ctx context.Context, oid string) (bool, error) {
	var count int
	err := db.sql.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM lfs_objects WHERE oid = ?
	`, oid).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("has lfs object: %w", err)
	}
	return count > 0, nil
}

// CreateLFSLock attempts to acquire a lock on a repository path.
func (db *DB) CreateLFSLock(ctx context.Context, lock LFSLock) error {
	if lock.LockedAt == "" {
		lock.LockedAt = time.Now().UTC().Format(time.RFC3339)
	}
	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO lfs_locks (id, repository_id, path, owner_name, locked_at)
		VALUES (?, ?, ?, ?, ?)
	`, lock.ID, lock.RepositoryID, lock.Path, lock.OwnerName, lock.LockedAt)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrLFSLockConflict, err)
	}
	return nil
}

// DeleteLFSLock removes an existing lock.
func (db *DB) DeleteLFSLock(ctx context.Context, id string) (LFSLock, error) {
	var lock LFSLock
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repository_id, path, owner_name, locked_at
		FROM lfs_locks
		WHERE id = ?
	`, id).Scan(&lock.ID, &lock.RepositoryID, &lock.Path, &lock.OwnerName, &lock.LockedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return LFSLock{}, ErrLFSLockNotFound
	}
	if err != nil {
		return LFSLock{}, err
	}

	_, err = db.sql.ExecContext(ctx, `DELETE FROM lfs_locks WHERE id = ?`, id)
	if err != nil {
		return LFSLock{}, fmt.Errorf("delete lfs lock: %w", err)
	}
	return lock, nil
}

// ListLFSLocks queries locks for a repository, optionally filtered by path.
func (db *DB) ListLFSLocks(ctx context.Context, repoID, path string) ([]LFSLock, error) {
	query := `SELECT id, repository_id, path, owner_name, locked_at FROM lfs_locks WHERE repository_id = ?`
	args := []any{repoID}

	if path != "" {
		query += ` AND path = ?`
		args = append(args, path)
	}
	query += ` ORDER BY locked_at DESC`

	rows, err := db.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list lfs locks: %w", err)
	}
	defer rows.Close()

	var locks []LFSLock
	for rows.Next() {
		var l LFSLock
		if err := rows.Scan(&l.ID, &l.RepositoryID, &l.Path, &l.OwnerName, &l.LockedAt); err != nil {
			return nil, err
		}
		locks = append(locks, l)
	}
	return locks, rows.Err()
}
