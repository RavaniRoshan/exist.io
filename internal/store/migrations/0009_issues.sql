-- ============================================================================
-- Migration 0009: Repository Issues and Issue Comments
-- ============================================================================

CREATE TABLE IF NOT EXISTS issues (
    id             TEXT PRIMARY KEY,
    repo_id        TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    number         INTEGER NOT NULL,
    state          TEXT NOT NULL DEFAULT 'open', -- open | closed
    title          TEXT NOT NULL,
    body           TEXT,
    author_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at     TEXT NOT NULL,
    updated_at     TEXT NOT NULL,
    closed_at      TEXT,
    UNIQUE (repo_id, number)
);

CREATE INDEX IF NOT EXISTS idx_issues_repo_state ON issues(repo_id, state);

CREATE TABLE IF NOT EXISTS issue_comments (
    id             TEXT PRIMARY KEY,
    repo_id        TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    issue_id       TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    body           TEXT NOT NULL,
    created_at     TEXT NOT NULL,
    updated_at     TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_issue_comments_issue ON issue_comments(issue_id);
