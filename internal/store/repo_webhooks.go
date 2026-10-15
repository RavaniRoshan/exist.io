package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"exist.io/pkg/idgen"
)

// RepoWebhook represents a repository-specific webhook subscription.
type RepoWebhook struct {
	ID          string `json:"id"`
	RepoID      string `json:"repo_id"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"` // json or form
	Secret      string `json:"secret,omitempty"`
	Events      string `json:"events"` // JSON array e.g. ["push", "pull_request"]
	Active      bool   `json:"active"`
	InsecureSSL bool   `json:"insecure_ssl"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// RepoWebhookDelivery records an HTTP dispatch attempt for a repository webhook.
type RepoWebhookDelivery struct {
	ID                 string `json:"id"`
	WebhookID          string `json:"webhook_id"`
	Event              string `json:"event"`
	GUID               string `json:"guid"`
	RequestHeaders     string `json:"request_headers"`
	RequestPayload     string `json:"request_payload"`
	ResponseStatusCode int    `json:"response_status_code"`
	ResponseHeaders    string `json:"response_headers"`
	ResponseBody       string `json:"response_body"`
	DurationMs         int    `json:"duration_ms"`
	Delivered          bool   `json:"delivered"`
	CreatedAt          string `json:"created_at"`
}

// CreateRepoWebhook registers a new webhook on a repository.
func (db *DB) CreateRepoWebhook(ctx context.Context, repoID, url, secret, contentType, events string, active, insecureSSL bool) (*RepoWebhook, error) {
	if url == "" {
		return nil, errors.New("webhook url cannot be empty")
	}
	if contentType == "" {
		contentType = "json"
	}
	if events == "" {
		events = `["push"]`
	}

	id := idgen.New("rwhk")
	now := time.Now().UTC().Format(time.RFC3339)

	activeInt := 0
	if active {
		activeInt = 1
	}
	insecureSSLInt := 0
	if insecureSSL {
		insecureSSLInt = 1
	}

	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO repo_webhooks (id, repo_id, url, content_type, secret, events, active, insecure_ssl, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, repoID, url, contentType, secret, events, activeInt, insecureSSLInt, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert repo webhook: %w", err)
	}

	return &RepoWebhook{
		ID:          id,
		RepoID:      repoID,
		URL:         url,
		ContentType: contentType,
		Secret:      secret,
		Events:      events,
		Active:      active,
		InsecureSSL: insecureSSL,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ListRepoWebhooks returns all webhooks configured for a repository.
func (db *DB) ListRepoWebhooks(ctx context.Context, repoID string) ([]*RepoWebhook, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, repo_id, url, content_type, secret, events, active, insecure_ssl, created_at, updated_at
		FROM repo_webhooks
		WHERE repo_id = ?
		ORDER BY created_at DESC
	`, repoID)
	if err != nil {
		return nil, fmt.Errorf("list repo webhooks: %w", err)
	}
	defer rows.Close()

	var hooks []*RepoWebhook
	for rows.Next() {
		var h RepoWebhook
		var activeInt, insecureSSLInt int
		if err := rows.Scan(&h.ID, &h.RepoID, &h.URL, &h.ContentType, &h.Secret, &h.Events, &activeInt, &insecureSSLInt, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan repo webhook: %w", err)
		}
		h.Active = activeInt == 1
		h.InsecureSSL = insecureSSLInt == 1
		hooks = append(hooks, &h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hooks, nil
}

// GetRepoWebhook returns a webhook by ID.
func (db *DB) GetRepoWebhook(ctx context.Context, id string) (*RepoWebhook, error) {
	var h RepoWebhook
	var activeInt, insecureSSLInt int
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repo_id, url, content_type, secret, events, active, insecure_ssl, created_at, updated_at
		FROM repo_webhooks
		WHERE id = ?
	`, id).Scan(&h.ID, &h.RepoID, &h.URL, &h.ContentType, &h.Secret, &h.Events, &activeInt, &insecureSSLInt, &h.CreatedAt, &h.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get repo webhook: %w", err)
	}
	h.Active = activeInt == 1
	h.InsecureSSL = insecureSSLInt == 1
	return &h, nil
}

// DeleteRepoWebhook removes a webhook.
func (db *DB) DeleteRepoWebhook(ctx context.Context, id string) error {
	res, err := db.sql.ExecContext(ctx, `DELETE FROM repo_webhooks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete repo webhook: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// RecordRepoWebhookDelivery saves a delivery attempt.
func (db *DB) RecordRepoWebhookDelivery(ctx context.Context, d *RepoWebhookDelivery) error {
	if d.ID == "" {
		d.ID = idgen.New("del")
	}
	if d.CreatedAt == "" {
		d.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	deliveredInt := 0
	if d.Delivered {
		deliveredInt = 1
	}

	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO repo_webhook_deliveries (
			id, webhook_id, event, guid, request_headers, request_payload,
			response_status_code, response_headers, response_body, duration_ms, delivered, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, d.ID, d.WebhookID, d.Event, d.GUID, d.RequestHeaders, d.RequestPayload,
		d.ResponseStatusCode, d.ResponseHeaders, d.ResponseBody, d.DurationMs, deliveredInt, d.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert repo webhook delivery: %w", err)
	}
	return nil
}

// ListRepoWebhookDeliveries returns the recent deliveries for a webhook.
func (db *DB) ListRepoWebhookDeliveries(ctx context.Context, webhookID string, limit int) ([]*RepoWebhookDelivery, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, webhook_id, event, guid, request_headers, request_payload,
		       response_status_code, response_headers, response_body, duration_ms, delivered, created_at
		FROM repo_webhook_deliveries
		WHERE webhook_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, webhookID, limit)
	if err != nil {
		return nil, fmt.Errorf("list repo webhook deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*RepoWebhookDelivery
	for rows.Next() {
		var d RepoWebhookDelivery
		var deliveredInt int
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.Event, &d.GUID, &d.RequestHeaders, &d.RequestPayload,
			&d.ResponseStatusCode, &d.ResponseHeaders, &d.ResponseBody, &d.DurationMs, &deliveredInt, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan repo webhook delivery: %w", err)
		}
		d.Delivered = deliveredInt == 1
		deliveries = append(deliveries, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return deliveries, nil
}

// GetRepoWebhookDelivery returns a specific delivery by ID.
func (db *DB) GetRepoWebhookDelivery(ctx context.Context, id string) (*RepoWebhookDelivery, error) {
	var d RepoWebhookDelivery
	var deliveredInt int
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, webhook_id, event, guid, request_headers, request_payload,
		       response_status_code, response_headers, response_body, duration_ms, delivered, created_at
		FROM repo_webhook_deliveries
		WHERE id = ?
	`, id).Scan(&d.ID, &d.WebhookID, &d.Event, &d.GUID, &d.RequestHeaders, &d.RequestPayload,
		&d.ResponseStatusCode, &d.ResponseHeaders, &d.ResponseBody, &d.DurationMs, &deliveredInt, &d.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get repo webhook delivery: %w", err)
	}
	d.Delivered = deliveredInt == 1
	return &d, nil
}
