package store

import (
	"context"
	"database/sql"
)

// InAppNotification represents an alert rendered in the forge web UI.
type InAppNotification struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	EventType string `json:"eventType"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	URL       string `json:"url"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"createdAt"`
}

// CreateInAppNotification persists a new in-app notification.
func (db *DB) CreateInAppNotification(ctx context.Context, n InAppNotification) error {
	ts := now()
	if n.CreatedAt == "" {
		n.CreatedAt = ts
	}
	readInt := 0
	if n.Read {
		readInt = 1
	}
	_, err := db.sql.ExecContext(ctx, `INSERT INTO in_app_notifications
		(id, user_id, event_type, title, body, url, read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID, n.UserID, n.EventType, n.Title, n.Body, n.URL, readInt, n.CreatedAt)
	return err
}

// ListInAppNotifications retrieves notifications for a user, ordered from newest to oldest.
func (db *DB) ListInAppNotifications(ctx context.Context, userID string, unreadOnly bool) ([]InAppNotification, error) {
	query := `SELECT id, user_id, event_type, title, body, url, read, created_at
		FROM in_app_notifications WHERE user_id = ?`
	if unreadOnly {
		query += ` AND read = 0`
	}
	query += ` ORDER BY created_at DESC LIMIT 50`

	rows, err := db.sql.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []InAppNotification
	for rows.Next() {
		var n InAppNotification
		var readInt int
		if err := rows.Scan(&n.ID, &n.UserID, &n.EventType, &n.Title, &n.Body, &n.URL, &readInt, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.Read = (readInt == 1)
		list = append(list, n)
	}
	return list, rows.Err()
}

// MarkInAppNotificationRead marks an individual notification as read.
func (db *DB) MarkInAppNotificationRead(ctx context.Context, id, userID string) error {
	query := `UPDATE in_app_notifications SET read = 1 WHERE id = ?`
	args := []any{id}
	if userID != "" {
		query += ` AND user_id = ?`
		args = append(args, userID)
	}
	res, err := db.sql.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// MarkAllInAppNotificationsRead marks all notifications for a user as read.
func (db *DB) MarkAllInAppNotificationsRead(ctx context.Context, userID string) error {
	_, err := db.sql.ExecContext(ctx, `UPDATE in_app_notifications SET read = 1 WHERE user_id = ? AND read = 0`, userID)
	return err
}

// CountUnreadInAppNotifications returns the number of unread notifications for a user.
func (db *DB) CountUnreadInAppNotifications(ctx context.Context, userID string) (int, error) {
	var count int
	err := db.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM in_app_notifications WHERE user_id = ? AND read = 0`, userID).Scan(&count)
	return count, err
}
