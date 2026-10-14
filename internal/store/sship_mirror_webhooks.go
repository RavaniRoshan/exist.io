package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"exist.io/pkg/idgen"
	"golang.org/x/crypto/ssh"
)

// SSHCertificateAuthority represents an authorized SSH certificate authority.
type SSHCertificateAuthority struct {
	ID          string
	NamespaceID string
	Name        string
	KeyType     string
	PublicKey   string
	Fingerprint string
	Required    bool
	CreatedAt   string
}

// IPAllowlistEntry represents one CIDR on a namespace's inbound IP allowlist.
type IPAllowlistEntry struct {
	ID          string
	NamespaceID string
	CIDR        string
	Description string
	Enabled     bool
	CreatedAt   string
	UpdatedAt   string
}

// MirrorTransitionJob tracks a repository's mirror state transition.
type MirrorTransitionJob struct {
	ID               string
	RepoID           string
	Transition       string
	Status           string // queued | running | succeeded | failed_rolled_back | requires_attention | superseded
	Phase            string
	AttemptCount     int
	DrainUntil       string
	LastErrorCode    string
	LastErrorMessage string
	StartedAt        string
	CompletedAt      string
	CreatedAt        string
	UpdatedAt        string
}

// WebhookDelivery represents a webhook payload owed to an application.
type WebhookDelivery struct {
	ID             string
	AppID          string
	InstallationID string
	Event          string
	Payload        string
	Signature      string
	Delivered      bool
	DeliveredAt    string
	LastAttemptAt  string
	AttemptCount   int
	CreatedAt      string
}

// ---------- SSH Certificate Authorities ----------

func (db *DB) ListSSHCAs(ctx context.Context, nsID string) ([]SSHCertificateAuthority, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, namespace_id, IFNULL(name, ''), key_type, public_key, fingerprint, required, created_at
		FROM ssh_cas
		WHERE namespace_id = ?
		ORDER BY created_at ASC, id ASC
	`, nsID)
	if err != nil {
		return nil, fmt.Errorf("list ssh cas: %w", err)
	}
	defer rows.Close()

	var list []SSHCertificateAuthority
	for rows.Next() {
		var ca SSHCertificateAuthority
		var reqInt int
		if err := rows.Scan(&ca.ID, &ca.NamespaceID, &ca.Name, &ca.KeyType, &ca.PublicKey, &ca.Fingerprint, &reqInt, &ca.CreatedAt); err != nil {
			return nil, err
		}
		ca.Required = reqInt == 1
		list = append(list, ca)
	}
	return list, rows.Err()
}

func (db *DB) AddSSHCA(ctx context.Context, nsID, name, rawPublicKey string) (SSHCertificateAuthority, error) {
	pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(strings.TrimSpace(rawPublicKey)))
	if err != nil {
		return SSHCertificateAuthority{}, fmt.Errorf("invalid public key: %w", err)
	}

	keyType := pub.Type()
	wireKey := fmt.Sprintf("%s %s", keyType, ssh.FingerprintSHA256(pub))
	fingerprint := ssh.FingerprintSHA256(pub)
	// OpenSSH wire format for public key without comments
	marshaledKey := fmt.Sprintf("%s %s", keyType, strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub))))
	parts := strings.Split(marshaledKey, " ")
	cleanPubKey := parts[0]
	if len(parts) > 1 {
		cleanPubKey = parts[0] + " " + parts[1]
	}

	ca := SSHCertificateAuthority{
		ID:          idgen.NewNSCA(),
		NamespaceID: nsID,
		Name:        name,
		KeyType:     keyType,
		PublicKey:   cleanPubKey,
		Fingerprint: fingerprint,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	_ = wireKey

	_, err = db.sql.ExecContext(ctx, `
		INSERT INTO ssh_cas (id, namespace_id, name, key_type, public_key, fingerprint, required, created_at)
		VALUES (?, ?, ?, ?, ?, ?, 0, ?)
	`, ca.ID, ca.NamespaceID, ca.Name, ca.KeyType, ca.PublicKey, ca.Fingerprint, ca.CreatedAt)
	if err != nil {
		return SSHCertificateAuthority{}, fmt.Errorf("insert ssh ca: %w", err)
	}
	return ca, nil
}

func (db *DB) DeleteSSHCA(ctx context.Context, nsID, caID string) error {
	res, err := db.sql.ExecContext(ctx, `
		DELETE FROM ssh_cas WHERE id = ? AND namespace_id = ?
	`, caID, nsID)
	if err != nil {
		return fmt.Errorf("delete ssh ca: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (db *DB) SetSSHCARequirement(ctx context.Context, nsID string, required bool) error {
	reqInt := 0
	if required {
		reqInt = 1
	}
	res, err := db.sql.ExecContext(ctx, `
		UPDATE namespaces SET ssh_cert_required = ? WHERE id = ?
	`, reqInt, nsID)
	if err != nil {
		return fmt.Errorf("set ssh cert requirement: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (db *DB) GetSSHCARequirement(ctx context.Context, nsID string) (bool, error) {
	var reqInt int
	err := db.sql.QueryRowContext(ctx, `
		SELECT ssh_cert_required FROM namespaces WHERE id = ?
	`, nsID).Scan(&reqInt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, ErrNotFound
	}
	if err != nil {
		return false, err
	}
	return reqInt == 1, nil
}

// ---------- Inbound IP Allowlist ----------

func (db *DB) GetIPAllowlist(ctx context.Context, nsID string) (bool, []IPAllowlistEntry, string, error) {
	var enabledInt int
	err := db.sql.QueryRowContext(ctx, `
		SELECT ip_allowlist_enabled FROM namespaces WHERE id = ?
	`, nsID).Scan(&enabledInt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil, "", ErrNotFound
	}
	if err != nil {
		return false, nil, "", err
	}

	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, namespace_id, cidr, IFNULL(description, ''), enabled, created_at, IFNULL(updated_at, created_at)
		FROM ip_allowlist
		WHERE namespace_id = ?
		ORDER BY created_at ASC, id ASC
	`, nsID)
	if err != nil {
		return false, nil, "", fmt.Errorf("query ip allowlist entries: %w", err)
	}
	defer rows.Close()

	var entries []IPAllowlistEntry
	h := sha256.New()
	for rows.Next() {
		var e IPAllowlistEntry
		var enInt int
		if err := rows.Scan(&e.ID, &e.NamespaceID, &e.CIDR, &e.Description, &enInt, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return false, nil, "", err
		}
		e.Enabled = enInt == 1
		entries = append(entries, e)
		h.Write([]byte(fmt.Sprintf("%s:%s:%s:%t;", e.ID, e.CIDR, e.Description, e.Enabled)))
	}
	etag := hex.EncodeToString(h.Sum(nil))
	return enabledInt == 1, entries, etag, nil
}

func (db *DB) SetIPAllowlistEnabled(ctx context.Context, nsID string, enabled bool) error {
	enInt := 0
	if enabled {
		enInt = 1
	}
	res, err := db.sql.ExecContext(ctx, `
		UPDATE namespaces SET ip_allowlist_enabled = ? WHERE id = ?
	`, enInt, nsID)
	if err != nil {
		return fmt.Errorf("update ip allowlist enabled: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (db *DB) AddIPAllowlistEntry(ctx context.Context, nsID, cidr, description string, enabled bool) (IPAllowlistEntry, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	entry := IPAllowlistEntry{
		ID:          idgen.NewNSIP(),
		NamespaceID: nsID,
		CIDR:        strings.TrimSpace(cidr),
		Description: description,
		Enabled:     enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	enInt := 0
	if enabled {
		enInt = 1
	}
	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO ip_allowlist (id, namespace_id, cidr, description, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, entry.ID, entry.NamespaceID, entry.CIDR, entry.Description, enInt, entry.CreatedAt, entry.UpdatedAt)
	if err != nil {
		return IPAllowlistEntry{}, fmt.Errorf("insert ip allowlist entry: %w", err)
	}
	return entry, nil
}

func (db *DB) GetIPAllowlistEntry(ctx context.Context, nsID, entryID string) (IPAllowlistEntry, error) {
	var e IPAllowlistEntry
	var enInt int
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, namespace_id, cidr, IFNULL(description, ''), enabled, created_at, IFNULL(updated_at, created_at)
		FROM ip_allowlist
		WHERE id = ? AND namespace_id = ?
	`, entryID, nsID).Scan(&e.ID, &e.NamespaceID, &e.CIDR, &e.Description, &enInt, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return IPAllowlistEntry{}, ErrNotFound
	}
	if err != nil {
		return IPAllowlistEntry{}, err
	}
	e.Enabled = enInt == 1
	return e, nil
}

func (db *DB) UpdateIPAllowlistEntry(ctx context.Context, nsID, entryID, newCIDR, newDesc string, enabled *bool) (IPAllowlistEntry, error) {
	existing, err := db.GetIPAllowlistEntry(ctx, nsID, entryID)
	if err != nil {
		return IPAllowlistEntry{}, err
	}

	if newCIDR != "" {
		existing.CIDR = strings.TrimSpace(newCIDR)
	}
	if newDesc != "" {
		existing.Description = newDesc
	}
	if enabled != nil {
		existing.Enabled = *enabled
	}
	existing.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	enInt := 0
	if existing.Enabled {
		enInt = 1
	}
	_, err = db.sql.ExecContext(ctx, `
		UPDATE ip_allowlist
		SET cidr = ?, description = ?, enabled = ?, updated_at = ?
		WHERE id = ? AND namespace_id = ?
	`, existing.CIDR, existing.Description, enInt, existing.UpdatedAt, entryID, nsID)
	if err != nil {
		return IPAllowlistEntry{}, fmt.Errorf("update ip allowlist entry: %w", err)
	}
	return existing, nil
}

func (db *DB) DeleteIPAllowlistEntry(ctx context.Context, nsID, entryID string) error {
	res, err := db.sql.ExecContext(ctx, `
		DELETE FROM ip_allowlist WHERE id = ? AND namespace_id = ?
	`, entryID, nsID)
	if err != nil {
		return fmt.Errorf("delete ip allowlist entry: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- Repository Mirroring ----------

func (db *DB) DetachMirror(ctx context.Context, repoID string) error {
	res, err := db.sql.ExecContext(ctx, `
		UPDATE repos
		SET mirror_source = NULL, mirror_source_id = NULL, mirror_status = 'detached', updated_at = ?
		WHERE id = ?
	`, time.Now().UTC().Format(time.RFC3339), repoID)
	if err != nil {
		return fmt.Errorf("detach mirror: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (db *DB) TransitionMirror(ctx context.Context, repoID, transition string) (MirrorTransitionJob, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	job := MirrorTransitionJob{
		ID:           "mtj_" + idgen.New("job"),
		RepoID:       repoID,
		Transition:   transition,
		Status:       "succeeded",
		Phase:        "completed",
		AttemptCount: 1,
		StartedAt:    now,
		CompletedAt:  now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO mirror_transition_jobs (
			id, repo_id, transition, status, phase, attempt_count,
			started_at, completed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID, job.RepoID, job.Transition, job.Status, job.Phase, job.AttemptCount,
		job.StartedAt, job.CompletedAt, job.CreatedAt, job.UpdatedAt)
	if err != nil {
		return MirrorTransitionJob{}, fmt.Errorf("insert mirror transition job: %w", err)
	}

	_, _ = db.sql.ExecContext(ctx, `
		UPDATE repos SET mirror_status = 'active', updated_at = ? WHERE id = ?
	`, now, repoID)

	return job, nil
}

func (db *DB) GetMirrorTransitionJob(ctx context.Context, repoID, jobID string) (MirrorTransitionJob, error) {
	var j MirrorTransitionJob
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repo_id, transition, status, phase, attempt_count,
		       IFNULL(drain_until, ''), IFNULL(last_error_code, ''), IFNULL(last_error_message, ''),
		       IFNULL(started_at, ''), IFNULL(completed_at, ''), created_at, updated_at
		FROM mirror_transition_jobs
		WHERE id = ? AND repo_id = ?
	`, jobID, repoID).Scan(
		&j.ID, &j.RepoID, &j.Transition, &j.Status, &j.Phase, &j.AttemptCount,
		&j.DrainUntil, &j.LastErrorCode, &j.LastErrorMessage,
		&j.StartedAt, &j.CompletedAt, &j.CreatedAt, &j.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return MirrorTransitionJob{}, ErrNotFound
	}
	if err != nil {
		return MirrorTransitionJob{}, err
	}
	return j, nil
}

func (db *DB) GetActiveMirrorTransitionJob(ctx context.Context, repoID string) (*MirrorTransitionJob, *MirrorTransitionJob, error) {
	// Query latest terminal job
	var lastJob *MirrorTransitionJob
	var j MirrorTransitionJob
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repo_id, transition, status, phase, attempt_count,
		       IFNULL(drain_until, ''), IFNULL(last_error_code, ''), IFNULL(last_error_message, ''),
		       IFNULL(started_at, ''), IFNULL(completed_at, ''), created_at, updated_at
		FROM mirror_transition_jobs
		WHERE repo_id = ? AND status IN ('succeeded', 'failed_rolled_back', 'superseded')
		ORDER BY created_at DESC LIMIT 1
	`, repoID).Scan(
		&j.ID, &j.RepoID, &j.Transition, &j.Status, &j.Phase, &j.AttemptCount,
		&j.DrainUntil, &j.LastErrorCode, &j.LastErrorMessage,
		&j.StartedAt, &j.CompletedAt, &j.CreatedAt, &j.UpdatedAt,
	)
	if err == nil {
		lastJob = &j
	}

	return nil, lastJob, nil
}

// ---------- Webhooks ----------

func (db *DB) CreateWebhookDelivery(ctx context.Context, d WebhookDelivery) (WebhookDelivery, error) {
	if d.ID == "" {
		d.ID = idgen.NewWebhookDelivery()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if d.CreatedAt == "" {
		d.CreatedAt = now
	}

	delInt := 0
	if d.Delivered {
		delInt = 1
	}

	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO webhook_deliveries (
			id, app_id, installation_id, event, payload, signature,
			delivered, delivered_at, last_attempt_at, attempt_count, created_at
		) VALUES (?, ?, NULLIF(?, ''), ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?)
	`, d.ID, d.AppID, d.InstallationID, d.Event, d.Payload, d.Signature,
		delInt, d.DeliveredAt, d.LastAttemptAt, d.AttemptCount, d.CreatedAt)
	if err != nil {
		return WebhookDelivery{}, fmt.Errorf("insert webhook delivery: %w", err)
	}
	return d, nil
}

func (db *DB) ListWebhookDeliveries(ctx context.Context, appID string, limit, offset int) ([]WebhookDelivery, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, app_id, IFNULL(installation_id, ''), event, payload, signature,
		       delivered, IFNULL(delivered_at, ''), IFNULL(last_attempt_at, ''), attempt_count, created_at
		FROM webhook_deliveries
		WHERE app_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, appID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list webhook deliveries: %w", err)
	}
	defer rows.Close()

	var list []WebhookDelivery
	for rows.Next() {
		var d WebhookDelivery
		var delInt int
		if err := rows.Scan(
			&d.ID, &d.AppID, &d.InstallationID, &d.Event, &d.Payload, &d.Signature,
			&delInt, &d.DeliveredAt, &d.LastAttemptAt, &d.AttemptCount, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		d.Delivered = delInt == 1
		list = append(list, d)
	}
	return list, rows.Err()
}
