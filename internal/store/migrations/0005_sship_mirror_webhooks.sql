-- 0005_sship_mirror_webhooks.sql — Schema additions for SSH CAs, IP allowlist, mirror transition jobs, and app installation links.

ALTER TABLE namespaces ADD COLUMN ip_allowlist_enabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE namespaces ADD COLUMN ssh_cert_required INTEGER NOT NULL DEFAULT 0;
ALTER TABLE ip_allowlist ADD COLUMN updated_at TEXT;

CREATE TABLE IF NOT EXISTS mirror_transition_jobs (
    id                 TEXT PRIMARY KEY,
    repo_id            TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    transition         TEXT NOT NULL,
    status             TEXT NOT NULL,
    phase              TEXT NOT NULL,
    attempt_count      INTEGER NOT NULL DEFAULT 1,
    drain_until        TEXT,
    last_error_code    TEXT,
    last_error_message TEXT,
    started_at         TEXT,
    completed_at       TEXT,
    created_at         TEXT NOT NULL,
    updated_at         TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_mirror_jobs_repo ON mirror_transition_jobs(repo_id, status);

CREATE TABLE IF NOT EXISTS app_signing_keys (
    id           TEXT PRIMARY KEY,
    app_id       TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    kid          TEXT NOT NULL,
    public_key   TEXT NOT NULL,
    created_at   TEXT NOT NULL,
    UNIQUE(app_id, kid)
);

CREATE TABLE IF NOT EXISTS installation_repos (
    installation_id TEXT NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
    repo_id         TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    created_at      TEXT NOT NULL,
    PRIMARY KEY (installation_id, repo_id)
);
