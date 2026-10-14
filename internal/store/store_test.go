package store

import (
	"path/filepath"
	"testing"
)

// openMem opens an in-memory DB and runs migrations.
func openMem(t *testing.T) *DB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestOpenAndMigrate(t *testing.T) {
	db := openMem(t)

	// Every planned table must exist after migration.
	want := []string{
		"namespaces", "repos", "grants", "rulesets", "ssh_cas",
		"ip_allowlist", "pull_requests", "check_runs", "apps",
		"installations", "users", "webhook_deliveries",
		// Additional tables introduced by later phases but created now so the
		// schema is stable from the first migration.
		"check_suites", "groups", "check_run_annotations",
		"pr_comments", "pr_reviews", "labels", "pr_labels",
		"pr_threads", "pr_comment_reactions", "pr_requested_reviewers", "pr_auto_merge",
		"migrations",
	}
	for _, tbl := range want {
		if !tableExists(t, db, tbl) {
			t.Errorf("table %q missing after migration", tbl)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := openMem(t)
	// Running migrate again must be a no-op, not an error.
	if err := db.migrate(); err != nil {
		t.Fatalf("second migrate() = %v, want nil", err)
	}
	if !tableExists(t, db, "repos") {
		t.Error("repos table missing after second migrate")
	}
}

func TestMigrateRecordsVersion(t *testing.T) {
	db := openMem(t)
	var n int
	if err := db.sql.QueryRow(`SELECT COUNT(*) FROM migrations`).Scan(&n); err != nil {
		t.Fatalf("query migrations: %v", err)
	}
	if n == 0 {
		t.Error("no migration versions recorded")
	}
}

func TestOpenFileOnDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exist.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open(%s) = %v", path, err)
	}
	defer db.Close()
	if !tableExists(t, db, "repos") {
		t.Error("repos table missing in on-disk DB")
	}
}

func TestOpenEnablesForeignKeys(t *testing.T) {
	db := openMem(t)
	var fk int
	if err := db.sql.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}
}

func tableExists(t *testing.T, db *DB, name string) bool {
	t.Helper()
	var n int
	err := db.sql.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, name,
	).Scan(&n)
	if err != nil {
		t.Fatalf("tableExists(%q): %v", name, err)
	}
	return n == 1
}
