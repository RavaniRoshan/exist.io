-- Migration 0007: Git LFS Objects and File Locks
CREATE TABLE IF NOT EXISTS lfs_objects (
    oid          TEXT PRIMARY KEY,
    size         INTEGER NOT NULL,
    storage_path TEXT NOT NULL,
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX IF NOT EXISTS idx_lfs_objects_created_at ON lfs_objects(created_at);

CREATE TABLE IF NOT EXISTS lfs_locks (
    id            TEXT PRIMARY KEY,
    repository_id TEXT NOT NULL,
    path          TEXT NOT NULL,
    owner_name    TEXT NOT NULL,
    locked_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    UNIQUE(repository_id, path)
);

CREATE INDEX IF NOT EXISTS idx_lfs_locks_repo ON lfs_locks(repository_id);
