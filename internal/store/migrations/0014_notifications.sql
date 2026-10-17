CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    channel    TEXT NOT NULL DEFAULT 'email',
    enabled    INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (user_id, event_type, channel)
);

CREATE TABLE IF NOT EXISTS notification_log (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    subject    TEXT NOT NULL,
    sent_at    TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'sent'
);
CREATE INDEX IF NOT EXISTS idx_notification_log_user ON notification_log(user_id);
