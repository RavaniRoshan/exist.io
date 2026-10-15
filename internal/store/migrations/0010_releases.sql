-- 0010_releases.sql — Schema for repository releases and tags metadata
CREATE TABLE IF NOT EXISTS releases (
    id           TEXT PRIMARY KEY,
    repo_id      TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    tag_name     TEXT NOT NULL,
    name         TEXT NOT NULL,
    body         TEXT NOT NULL,
    draft        INTEGER NOT NULL DEFAULT 0,
    prerelease   INTEGER NOT NULL DEFAULT 0,
    author       TEXT NOT NULL,
    commit_sha   TEXT NOT NULL DEFAULT '',
    created_at   TEXT NOT NULL,
    published_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_releases_repo ON releases(repo_id, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_releases_tag ON releases(repo_id, tag_name);
