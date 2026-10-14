-- Migration 0003: Pull request social features (threads, reactions, reviewers, auto-merge, dismissals)

ALTER TABLE pull_requests ADD COLUMN parent_pr_id TEXT REFERENCES pull_requests(id) ON DELETE SET NULL;

CREATE TABLE pr_threads (
    id                  TEXT PRIMARY KEY,
    repo_id             TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    pull_request_id     TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    path                TEXT,
    side                TEXT,
    start_line          INTEGER NOT NULL DEFAULT 0,
    end_line            INTEGER NOT NULL DEFAULT 0,
    version_number      INTEGER NOT NULL DEFAULT 1,
    head_sha            TEXT,
    base_sha            TEXT,
    resolved_at         TEXT,
    resolved_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at          TEXT NOT NULL,
    updated_at          TEXT NOT NULL
);
CREATE INDEX idx_pr_threads_pr ON pr_threads(pull_request_id);

ALTER TABLE pr_comments ADD COLUMN thread_id TEXT REFERENCES pr_threads(id) ON DELETE CASCADE;

CREATE TABLE pr_comment_reactions (
    id         TEXT PRIMARY KEY,
    comment_id TEXT NOT NULL REFERENCES pr_comments(id) ON DELETE CASCADE,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji      TEXT NOT NULL,
    created_at TEXT NOT NULL,
    UNIQUE (comment_id, user_id, emoji)
);
CREATE INDEX idx_reactions_comment ON pr_comment_reactions(comment_id);

CREATE TABLE pr_requested_reviewers (
    pull_request_id TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id         TEXT REFERENCES users(id) ON DELETE CASCADE,
    group_id        TEXT REFERENCES groups(id) ON DELETE CASCADE,
    created_at      TEXT NOT NULL
);
CREATE INDEX idx_pr_req_rev_pr ON pr_requested_reviewers(pull_request_id);

CREATE TABLE pr_auto_merge (
    pull_request_id     TEXT PRIMARY KEY REFERENCES pull_requests(id) ON DELETE CASCADE,
    enabled_by_user_id  TEXT REFERENCES users(id) ON DELETE SET NULL,
    enabled_at          TEXT NOT NULL,
    merge_method        TEXT NOT NULL,
    commit_title        TEXT,
    commit_message      TEXT
);

ALTER TABLE pr_reviews ADD COLUMN dismissal_message TEXT;
ALTER TABLE pr_reviews ADD COLUMN dismissed_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE pr_reviews ADD COLUMN dismissed_at TEXT;
ALTER TABLE pr_reviews ADD COLUMN version_number INTEGER NOT NULL DEFAULT 1;
