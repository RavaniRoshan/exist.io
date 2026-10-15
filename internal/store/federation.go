package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SSOConfig holds enterprise Single Sign-On configuration.
type SSOConfig struct {
	ID           string `json:"id"`
	ProviderType string `json:"provider_type"` // "oidc" or "saml"
	IssuerURL    string `json:"issuer_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	MetadataXML  string `json:"metadata_xml,omitempty"`
	Enabled      bool   `json:"enabled"`
	EnforceSSO   bool   `json:"enforce_sso"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// SCIMToken stores a bearer token used by identity providers (Okta, Azure AD) for SCIM synchronization.
type SCIMToken struct {
	ID          string `json:"id"`
	TokenHash   string `json:"token_hash"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

var (
	ErrSSONotConfigured = errors.New("sso is not configured")
)

// GetSSOConfig returns the active SSO configuration.
func (db *DB) GetSSOConfig(ctx context.Context) (SSOConfig, error) {
	row := db.sql.QueryRowContext(ctx, `
		SELECT id, provider_type, issuer_url, client_id, client_secret, IFNULL(metadata_xml, ''), enabled, enforce_sso, created_at, updated_at
		FROM enterprise_sso_configs
		LIMIT 1
	`)

	var cfg SSOConfig
	var enabledInt, enforceInt int
	err := row.Scan(&cfg.ID, &cfg.ProviderType, &cfg.IssuerURL, &cfg.ClientID, &cfg.ClientSecret, &cfg.MetadataXML, &enabledInt, &enforceInt, &cfg.CreatedAt, &cfg.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return SSOConfig{}, ErrSSONotConfigured
	}
	if err != nil {
		return SSOConfig{}, fmt.Errorf("get sso config: %w", err)
	}

	cfg.Enabled = enabledInt == 1
	cfg.EnforceSSO = enforceInt == 1
	return cfg, nil
}

// SetSSOConfig inserts or updates the enterprise SSO configuration.
func (db *DB) SetSSOConfig(ctx context.Context, cfg SSOConfig) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if cfg.CreatedAt == "" {
		cfg.CreatedAt = now
	}
	cfg.UpdatedAt = now

	enabledInt := 0
	if cfg.Enabled {
		enabledInt = 1
	}
	enforceInt := 0
	if cfg.EnforceSSO {
		enforceInt = 1
	}

	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO enterprise_sso_configs (id, provider_type, issuer_url, client_id, client_secret, metadata_xml, enabled, enforce_sso, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			provider_type = excluded.provider_type,
			issuer_url = excluded.issuer_url,
			client_id = excluded.client_id,
			client_secret = excluded.client_secret,
			metadata_xml = excluded.metadata_xml,
			enabled = excluded.enabled,
			enforce_sso = excluded.enforce_sso,
			updated_at = excluded.updated_at
	`, cfg.ID, cfg.ProviderType, cfg.IssuerURL, cfg.ClientID, cfg.ClientSecret, cfg.MetadataXML, enabledInt, enforceInt, cfg.CreatedAt, cfg.UpdatedAt)
	if err != nil {
		return fmt.Errorf("set sso config: %w", err)
	}
	return nil
}

// CreateSCIMToken records a new SCIM API token.
func (db *DB) CreateSCIMToken(ctx context.Context, t SCIMToken) error {
	if t.CreatedAt == "" {
		t.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO scim_tokens (id, token_hash, description, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`, t.ID, t.TokenHash, t.Description, t.CreatedAt, nullStr(t.ExpiresAt))
	if err != nil {
		return fmt.Errorf("create scim token: %w", err)
	}
	return nil
}

// ValidateSCIMToken validates a token hash.
func (db *DB) ValidateSCIMToken(ctx context.Context, hash string) (bool, error) {
	var count int
	err := db.sql.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM scim_tokens
		WHERE token_hash = ?
	`, hash).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("validate scim token: %w", err)
	}
	return count > 0, nil
}
