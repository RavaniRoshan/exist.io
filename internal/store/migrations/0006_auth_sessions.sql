-- 0006_auth_sessions.sql — User authentication, sessions, and SSH keys for enterprise self-hosting.

ALTER TABLE users ADD COLUMN password_hash TEXT;
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'member';
ALTER TABLE users ADD COLUMN status TEXT NOT NULL DEFAULT 'active';

CREATE TABLE user_sessions (
    id                 TEXT PRIMARY KEY,              -- sess_... (ULID)
    user_id            TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash         TEXT NOT NULL UNIQUE,          -- SHA-256 of raw session token
    csrf_token         TEXT NOT NULL,                 -- Synchronizer token for mutating operations
    ip_address         TEXT,
    user_agent         TEXT,
    created_at         TEXT NOT NULL,
    expires_at         TEXT NOT NULL
);
CREATE INDEX idx_user_sessions_hash ON user_sessions(token_hash);
CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);

CREATE TABLE user_ssh_keys (
    id                 TEXT PRIMARY KEY,              -- sshkey_... (ULID)
    user_id            TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name               TEXT NOT NULL,                 -- e.g. "Workstation Key"
    fingerprint        TEXT NOT NULL UNIQUE,          -- SHA256:Base64...
    public_key         TEXT NOT NULL,                 -- raw authorized_keys format
    key_type           TEXT NOT NULL,                 -- ssh-ed25519, ssh-rsa, etc.
    created_at         TEXT NOT NULL
);
CREATE INDEX idx_user_ssh_keys_user ON user_ssh_keys(user_id);
CREATE INDEX idx_user_ssh_keys_fingerprint ON user_ssh_keys(fingerprint);
