-- 0002_credentials.sql — exist.io local credentials schema.
--
-- Enables local, self-hosted API keys and access tokens so exist.io operates
-- completely independently of Cursor's authentication infrastructure.

CREATE TABLE api_keys (
    id           TEXT PRIMARY KEY,              -- key_... (ULID prefixed)
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash     TEXT NOT NULL UNIQUE,          -- SHA-256 of raw secret key
    key_prefix   TEXT NOT NULL,                 -- e.g. "key_12345678" for identification
    name         TEXT NOT NULL,                 -- label, e.g. "bootstrap-admin-key"
    scopes       TEXT NOT NULL DEFAULT '*',     -- space-separated scopes or '*' for all
    created_at   TEXT NOT NULL,
    expires_at   TEXT,                          -- NULL = no expiration
    revoked_at   TEXT,                          -- NULL = active
    last_used_at TEXT
);
CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);

CREATE TABLE access_tokens (
    id           TEXT PRIMARY KEY,              -- tok_... (ULID prefixed)
    token_hash   TEXT NOT NULL UNIQUE,          -- SHA-256 of raw access token
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id   TEXT REFERENCES api_keys(id) ON DELETE SET NULL,
    token_type   TEXT NOT NULL,                 -- 'user' | 'installation' | 'app'
    scopes       TEXT NOT NULL DEFAULT '*',     -- space-separated scopes or '*' for all
    created_at   TEXT NOT NULL,
    expires_at   TEXT NOT NULL,
    revoked_at   TEXT                           -- NULL = active
);
CREATE INDEX idx_access_tokens_hash ON access_tokens(token_hash);
CREATE INDEX idx_access_tokens_user ON access_tokens(user_id);
