package store

import (
	"database/sql"
	"embed"
	"fmt"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// migrate applies all embedded migrations in order, recording each in the
// migrations table. It is idempotent: already-applied versions are skipped.
func (db *DB) migrate() error {
	if _, err := db.sql.Exec(`CREATE TABLE IF NOT EXISTS migrations (
		version    INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
	)`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		var version int
		if _, err := fmt.Sscanf(e.Name(), "%d_", &version); err != nil {
			return fmt.Errorf("parse migration filename %q: %w", e.Name(), err)
		}

		var applied int
		if err := db.sql.QueryRow(`SELECT COUNT(*) FROM migrations WHERE version = ?`, version).Scan(&applied); err != nil {
			return err
		}
		if applied > 0 {
			continue
		}

		body, err := migrationFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return fmt.Errorf("read migration %q: %w", e.Name(), err)
		}

		tx, err := db.sql.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(body)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %q: %w", e.Name(), err)
		}
		if _, err := tx.Exec(`INSERT INTO migrations (version) VALUES (?)`, version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %q: %w", e.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %q: %w", e.Name(), err)
		}
	}
	return nil
}

// version returns the highest applied migration version, or 0 if none.
func (db *DB) version() (int, error) {
	var v int
	err := db.sql.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM migrations`).Scan(&v)
	return v, err
}

var _ = sql.ErrNoRows
