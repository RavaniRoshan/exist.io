-- Migration 0008: Enterprise Federation (SSO and SCIM 2.0)
CREATE TABLE IF NOT EXISTS enterprise_sso_configs (
    id            TEXT PRIMARY KEY,
    provider_type TEXT NOT NULL, -- 'oidc' | 'saml'
    issuer_url    TEXT NOT NULL,
    client_id     TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    metadata_xml  TEXT,
    enabled       INTEGER NOT NULL DEFAULT 1,
    enforce_sso   INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS scim_tokens (
    id          TEXT PRIMARY KEY,
    token_hash  TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    expires_at  TEXT
);

CREATE INDEX IF NOT EXISTS idx_scim_tokens_hash ON scim_tokens(token_hash);
