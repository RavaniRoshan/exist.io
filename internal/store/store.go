// Package store is exist.io's metadata store: a pure-Go SQLite database
// (modernc.org/sqlite, no cgo) holding namespaces, repos, grants, pull
// requests, checks, rulesets, apps, and every other Origin entity that is not
// itself a git object. Git object data lives in internal/gitstore.
package store

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, registers name "sqlite"
)

// DB wraps the metadata database and exposes the underlying *sql.DB for
// repository/query code in later phases.
type DB struct {
	sql *sql.DB
}

// Open opens (creating if necessary) the SQLite database at path, applies
// pragmas, and runs migrations. Use ":memory:" for an ephemeral database.
func Open(path string) (*DB, error) {
	// modernc's driver expects a file: DSN with a few pragmas for durability.
	dsn := "file:" + path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	if path == ":memory:" {
		dsn = "file::memory:?_pragma=busy_timeout(5000)"
	}

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", path, err)
	}
	// Configure connection pooling: in-memory requires 1 connection to preserve state,
	// while on-disk WAL mode supports concurrent readers with busy_timeout queuing.
	if path == ":memory:" {
		sqlDB.SetMaxOpenConns(1)
	} else {
		maxConns := max(4, runtime.NumCPU()*2)
		sqlDB.SetMaxOpenConns(maxConns)
		sqlDB.SetMaxIdleConns(max(2, runtime.NumCPU()))
	}

	if _, err := sqlDB.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("store: enable foreign_keys: %w", err)
	}

	db := &DB{sql: sqlDB}
	if err := db.migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("store: migrate: %w", err)
	}
	return db, nil
}

// Close closes the underlying database.
func (db *DB) Close() error {
	if db == nil || db.sql == nil {
		return nil
	}
	return db.sql.Close()
}

// SQL exposes the raw database for query code in later phases. Prefer typed
// helpers as they are added; this is the escape hatch.
func (db *DB) SQL() *sql.DB { return db.sql }

// Dir returns the directory that should hold git object stores for repos backed
// by this metadata DB (i.e. the parent of the sqlite file). For ":memory:" it
// returns "".
func (db *DB) Dir() string { return "" }

// ensureDir is a small helper for later phases that lay git repos next to the
// metadata file.
func dataDir(path string) string {
	if path == ":memory:" {
		return ""
	}
	return filepath.Dir(path)
}
