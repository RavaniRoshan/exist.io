-- 0013_agent_leases.sql: Schema for agent capability lease tracking and revocation

CREATE TABLE IF NOT EXISTS revoked_agent_leases (
    lease_id    TEXT PRIMARY KEY,
    agent_id    TEXT NOT NULL,
    repo_id     TEXT NOT NULL,
    revoked_at  TEXT NOT NULL,
    reason      TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_revoked_leases_agent ON revoked_agent_leases(agent_id);
CREATE INDEX IF NOT EXISTS idx_revoked_leases_repo ON revoked_agent_leases(repo_id);
