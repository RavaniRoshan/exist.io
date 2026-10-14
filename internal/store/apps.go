package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"exist.io/pkg/idgen"
)

type App struct {
	ID            string   `json:"id"`
	NamespaceID   string   `json:"namespaceId"`
	DisplayName   string   `json:"displayName"`
	Description   string   `json:"description,omitempty"`
	WebsiteURL    string   `json:"websiteUrl,omitempty"`
	WebhookURL    string   `json:"webhookUrl,omitempty"`
	Events        []string `json:"events,omitempty"`
	DefaultScopes []string `json:"defaultScopes,omitempty"`
	SigningKey    string   `json:"signingKey,omitempty"`
	Visibility    string   `json:"visibility,omitempty"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
}

type Installation struct {
	ID            string   `json:"id"`
	AppID         string   `json:"appId"`
	NamespaceID   string   `json:"namespaceId"`
	Suspended     bool     `json:"suspended"`
	GrantedScopes []string `json:"grantedScopes,omitempty"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
}

type AppSigningKey struct {
	ID        string `json:"id"`
	AppID     string `json:"appId"`
	KID       string `json:"kid"`
	PublicKey string `json:"publicKey"`
	CreatedAt string `json:"createdAt"`
}

func (db *DB) CreateApp(ctx context.Context, a App) (App, error) {
	if a.ID == "" {
		a.ID = idgen.NewApp()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if a.CreatedAt == "" {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	if a.Visibility == "" {
		a.Visibility = "private"
	}

	eventsJSON, _ := json.Marshal(a.Events)
	scopesJSON, _ := json.Marshal(a.DefaultScopes)

	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO apps (
			id, namespace_id, display_name, description, website_url, webhook_url,
			events, default_scopes, signing_key, visibility, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, a.ID, a.NamespaceID, a.DisplayName, a.Description, a.WebsiteURL, a.WebhookURL,
		string(eventsJSON), string(scopesJSON), a.SigningKey, a.Visibility, a.CreatedAt, a.UpdatedAt)
	if err != nil {
		return App{}, fmt.Errorf("insert app: %w", err)
	}
	return a, nil
}

func (db *DB) GetApp(ctx context.Context, id string) (App, error) {
	var a App
	var eventsRaw, scopesRaw sql.NullString
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, namespace_id, display_name, IFNULL(description, ''),
		       IFNULL(website_url, ''), IFNULL(webhook_url, ''),
		       events, default_scopes, IFNULL(signing_key, ''), visibility,
		       created_at, updated_at
		FROM apps
		WHERE id = ?
	`, id).Scan(
		&a.ID, &a.NamespaceID, &a.DisplayName, &a.Description,
		&a.WebsiteURL, &a.WebhookURL,
		&eventsRaw, &scopesRaw, &a.SigningKey, &a.Visibility,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return App{}, ErrNotFound
	}
	if err != nil {
		return App{}, err
	}
	if eventsRaw.Valid && eventsRaw.String != "" {
		_ = json.Unmarshal([]byte(eventsRaw.String), &a.Events)
	}
	if scopesRaw.Valid && scopesRaw.String != "" {
		_ = json.Unmarshal([]byte(scopesRaw.String), &a.DefaultScopes)
	}
	return a, nil
}

func (db *DB) ListNamespaceApps(ctx context.Context, nsID string) ([]App, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, namespace_id, display_name, IFNULL(description, ''),
		       IFNULL(website_url, ''), IFNULL(webhook_url, ''),
		       events, default_scopes, IFNULL(signing_key, ''), visibility,
		       created_at, updated_at
		FROM apps
		WHERE namespace_id = ?
		ORDER BY created_at ASC, id ASC
	`, nsID)
	if err != nil {
		return nil, fmt.Errorf("list namespace apps: %w", err)
	}
	defer rows.Close()

	var list []App
	for rows.Next() {
		var a App
		var eventsRaw, scopesRaw sql.NullString
		if err := rows.Scan(
			&a.ID, &a.NamespaceID, &a.DisplayName, &a.Description,
			&a.WebsiteURL, &a.WebhookURL,
			&eventsRaw, &scopesRaw, &a.SigningKey, &a.Visibility,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if eventsRaw.Valid && eventsRaw.String != "" {
			_ = json.Unmarshal([]byte(eventsRaw.String), &a.Events)
		}
		if scopesRaw.Valid && scopesRaw.String != "" {
			_ = json.Unmarshal([]byte(scopesRaw.String), &a.DefaultScopes)
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (db *DB) UpdateApp(ctx context.Context, a App) (App, error) {
	existing, err := db.GetApp(ctx, a.ID)
	if err != nil {
		return App{}, err
	}

	if a.DisplayName != "" {
		existing.DisplayName = a.DisplayName
	}
	if a.Description != "" {
		existing.Description = a.Description
	}
	if a.WebsiteURL != "" {
		existing.WebsiteURL = a.WebsiteURL
	}
	if a.WebhookURL != "" {
		existing.WebhookURL = a.WebhookURL
	}
	if len(a.Events) > 0 {
		existing.Events = a.Events
	}
	if len(a.DefaultScopes) > 0 {
		existing.DefaultScopes = a.DefaultScopes
	}
	existing.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	eventsJSON, _ := json.Marshal(existing.Events)
	scopesJSON, _ := json.Marshal(existing.DefaultScopes)

	_, err = db.sql.ExecContext(ctx, `
		UPDATE apps
		SET display_name = ?, description = ?, website_url = ?, webhook_url = ?,
		    events = ?, default_scopes = ?, updated_at = ?
		WHERE id = ?
	`, existing.DisplayName, existing.Description, existing.WebsiteURL, existing.WebhookURL,
		string(eventsJSON), string(scopesJSON), existing.UpdatedAt, existing.ID)
	if err != nil {
		return App{}, fmt.Errorf("update app: %w", err)
	}
	return existing, nil
}

func (db *DB) AddAppSigningKey(ctx context.Context, appID, kid, pubKey string) (AppSigningKey, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	key := AppSigningKey{
		ID:        idgen.New("ask"),
		AppID:     appID,
		KID:       kid,
		PublicKey: pubKey,
		CreatedAt: now,
	}
	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO app_signing_keys (id, app_id, kid, public_key, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, key.ID, key.AppID, key.KID, key.PublicKey, key.CreatedAt)
	if err != nil {
		return AppSigningKey{}, fmt.Errorf("insert signing key: %w", err)
	}
	return key, nil
}

func (db *DB) RevokeAppSigningKey(ctx context.Context, appID, kid string) error {
	res, err := db.sql.ExecContext(ctx, `
		DELETE FROM app_signing_keys WHERE app_id = ? AND kid = ?
	`, appID, kid)
	if err != nil {
		return fmt.Errorf("delete signing key: %w", err)
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

func (db *DB) CreateInstallation(ctx context.Context, inst Installation) (Installation, error) {
	if inst.ID == "" {
		inst.ID = idgen.NewInstallation()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if inst.CreatedAt == "" {
		inst.CreatedAt = now
	}
	inst.UpdatedAt = now

	susInt := 0
	if inst.Suspended {
		susInt = 1
	}
	scopesJSON, _ := json.Marshal(inst.GrantedScopes)

	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO installations (id, app_id, namespace_id, suspended, granted_scopes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, inst.ID, inst.AppID, inst.NamespaceID, susInt, string(scopesJSON), inst.CreatedAt, inst.UpdatedAt)
	if err != nil {
		return Installation{}, fmt.Errorf("insert installation: %w", err)
	}
	return inst, nil
}

func (db *DB) GetInstallation(ctx context.Context, id string) (Installation, error) {
	var inst Installation
	var susInt int
	var scopesRaw sql.NullString
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, app_id, namespace_id, suspended, granted_scopes, created_at, updated_at
		FROM installations
		WHERE id = ?
	`, id).Scan(&inst.ID, &inst.AppID, &inst.NamespaceID, &susInt, &scopesRaw, &inst.CreatedAt, &inst.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Installation{}, ErrNotFound
	}
	if err != nil {
		return Installation{}, err
	}
	inst.Suspended = susInt == 1
	if scopesRaw.Valid && scopesRaw.String != "" {
		_ = json.Unmarshal([]byte(scopesRaw.String), &inst.GrantedScopes)
	}
	return inst, nil
}

func (db *DB) ListAppInstallations(ctx context.Context, appID string) ([]Installation, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, app_id, namespace_id, suspended, granted_scopes, created_at, updated_at
		FROM installations
		WHERE app_id = ?
		ORDER BY created_at ASC, id ASC
	`, appID)
	if err != nil {
		return nil, fmt.Errorf("list installations: %w", err)
	}
	defer rows.Close()

	var list []Installation
	for rows.Next() {
		var inst Installation
		var susInt int
		var scopesRaw sql.NullString
		if err := rows.Scan(&inst.ID, &inst.AppID, &inst.NamespaceID, &susInt, &scopesRaw, &inst.CreatedAt, &inst.UpdatedAt); err != nil {
			return nil, err
		}
		inst.Suspended = susInt == 1
		if scopesRaw.Valid && scopesRaw.String != "" {
			_ = json.Unmarshal([]byte(scopesRaw.String), &inst.GrantedScopes)
		}
		list = append(list, inst)
	}
	return list, rows.Err()
}

func (db *DB) DeleteInstallation(ctx context.Context, id string) error {
	res, err := db.sql.ExecContext(ctx, `
		DELETE FROM installations WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("delete installation: %w", err)
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

func (db *DB) AddInstallationRepos(ctx context.Context, instID string, repoIDs []string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	for _, rid := range repoIDs {
		_, _ = db.sql.ExecContext(ctx, `
			INSERT OR IGNORE INTO installation_repos (installation_id, repo_id, created_at)
			VALUES (?, ?, ?)
		`, instID, rid, now)
	}
	return nil
}

func (db *DB) ListInstallationRepos(ctx context.Context, instID string) ([]Repo, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT r.id, r.namespace_id, r.name, r.full_name, r.default_branch, r.visibility,
		       IFNULL(r.clone_url, ''), IFNULL(r.web_url, ''),
		       r.allow_merge_commit, r.allow_squash_merge, r.delete_branch_on_merge,
		       r.created_at, r.updated_at, IFNULL(r.pushed_at, '')
		FROM repos r
		JOIN installation_repos ir ON r.id = ir.repo_id
		WHERE ir.installation_id = ?
		ORDER BY r.name ASC
	`, instID)
	if err != nil {
		return nil, fmt.Errorf("list installation repos: %w", err)
	}
	defer rows.Close()

	var list []Repo
	for rows.Next() {
		var r Repo
		var amc, asm, dbo int
		if err := rows.Scan(
			&r.ID, &r.NamespaceID, &r.Name, &r.FullName, &r.DefaultBranch, &r.Visibility,
			&r.CloneURL, &r.WebURL,
			&amc, &asm, &dbo,
			&r.CreatedAt, &r.UpdatedAt, &r.PushedAt,
		); err != nil {
			return nil, err
		}
		r.AllowMergeCommit = amc == 1
		r.AllowSquashMerge = asm == 1
		r.DeleteBranchOnMerge = dbo == 1
		list = append(list, r)
	}
	return list, rows.Err()
}
