-- 0004_checks_rulesets.sql — Add missing columns to check runs, suites, annotations, and indexes.

ALTER TABLE check_runs ADD COLUMN output_text TEXT;
ALTER TABLE check_runs ADD COLUMN external_updated_at TEXT;
ALTER TABLE check_runs ADD COLUMN deadline_at TEXT;
ALTER TABLE check_runs ADD COLUMN is_rerequestable INTEGER NOT NULL DEFAULT 0;
ALTER TABLE check_runs ADD COLUMN rerequested_at TEXT;
ALTER TABLE check_runs ADD COLUMN rerequested_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE check_runs ADD COLUMN actor_user_id TEXT REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE check_suites ADD COLUMN actor_user_id TEXT REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE check_run_annotations ADD COLUMN start_column INTEGER;
ALTER TABLE check_run_annotations ADD COLUMN end_column INTEGER;
ALTER TABLE check_run_annotations ADD COLUMN updated_at TEXT;

CREATE INDEX IF NOT EXISTS idx_check_runs_ext_id ON check_runs(check_suite_id, external_id);
CREATE INDEX IF NOT EXISTS idx_check_runs_key ON check_runs(check_suite_id, key);
CREATE INDEX IF NOT EXISTS idx_check_suites_lookup ON check_suites(repo_id, sha, key);
