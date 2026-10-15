-- Migration 0011: Repository Lifecycle (Description & Archival Read-Only Freeze)

ALTER TABLE repos ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE repos ADD COLUMN archived INTEGER NOT NULL DEFAULT 0;
