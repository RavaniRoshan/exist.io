-- 0012_deploy_keys_and_repo_webhooks.sql — Repository-level Deploy Keys and Webhooks

CREATE TABLE IF NOT EXISTS repo_deploy_keys (
    id           TEXT PRIMARY KEY,
    repo_id      TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    title        TEXT NOT NULL,
    public_key   TEXT NOT NULL,
    fingerprint  TEXT NOT NULL UNIQUE,
    read_only    INTEGER NOT NULL DEFAULT 1,
    last_used_at TEXT,
    created_at   TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_repo_deploy_keys_repo ON repo_deploy_keys(repo_id);
CREATE INDEX IF NOT EXISTS idx_repo_deploy_keys_fp ON repo_deploy_keys(fingerprint);

CREATE TABLE IF NOT EXISTS repo_webhooks (
    id           TEXT PRIMARY KEY,
    repo_id      TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    url          TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT 'json',
    secret       TEXT NOT NULL DEFAULT '',
    events       TEXT NOT NULL DEFAULT '["push"]',
    active       INTEGER NOT NULL DEFAULT 1,
    insecure_ssl INTEGER NOT NULL DEFAULT 0,
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_repo_webhooks_repo ON repo_webhooks(repo_id);

CREATE TABLE IF NOT EXISTS repo_webhook_deliveries (
    id                   TEXT PRIMARY KEY,
    webhook_id           TEXT NOT NULL REFERENCES repo_webhooks(id) ON DELETE CASCADE,
    event                TEXT NOT NULL,
    guid                 TEXT NOT NULL,
    request_headers      TEXT NOT NULL,
    request_payload      TEXT NOT NULL,
    response_status_code INTEGER DEFAULT 0,
    response_headers     TEXT DEFAULT '',
    response_body        TEXT DEFAULT '',
    duration_ms          INTEGER DEFAULT 0,
    delivered            INTEGER NOT NULL DEFAULT 0,
    created_at           TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_repo_whd_hook ON repo_webhook_deliveries(webhook_id, created_at DESC);
