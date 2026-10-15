package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestFederationStore(t *testing.T) {
	dataDir := t.TempDir()
	db, err := Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Initial SSO config not found
	_, err = db.GetSSOConfig(ctx)
	if err == nil {
		t.Fatalf("expected ErrSSONotConfigured, got nil")
	}

	// 2. Set SSO config
	cfg := SSOConfig{
		ID:           "sso_global",
		ProviderType: "oidc",
		IssuerURL:    "https://login.microsoftonline.com/corp/v2.0",
		ClientID:     "client_123456",
		ClientSecret: "secret_abcdef",
		Enabled:      true,
		EnforceSSO:   false,
	}
	if err := db.SetSSOConfig(ctx, cfg); err != nil {
		t.Fatalf("SetSSOConfig: %v", err)
	}

	// 3. Get SSO config
	fetched, err := db.GetSSOConfig(ctx)
	if err != nil {
		t.Fatalf("GetSSOConfig: %v", err)
	}
	if fetched.IssuerURL != cfg.IssuerURL || !fetched.Enabled {
		t.Fatalf("unexpected fetched SSO config: %+v", fetched)
	}

	// 4. SCIM token
	tok := SCIMToken{
		ID:          "scim_tok_1",
		TokenHash:   "hash_okta_token_999",
		Description: "Okta SCIM provisioning",
	}
	if err := db.CreateSCIMToken(ctx, tok); err != nil {
		t.Fatalf("CreateSCIMToken: %v", err)
	}

	valid, err := db.ValidateSCIMToken(ctx, "hash_okta_token_999")
	if err != nil || !valid {
		t.Fatalf("ValidateSCIMToken valid token failed: %v", err)
	}

	invalid, err := db.ValidateSCIMToken(ctx, "hash_unknown")
	if err != nil || invalid {
		t.Fatalf("ValidateSCIMToken invalid token expected false, got true")
	}
}
