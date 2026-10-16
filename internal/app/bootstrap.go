package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-git/go-git/v6/plumbing/object"

	"exist.io/internal/auth"
	"exist.io/internal/config"
	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

// BootstrapResult contains the details created by a successful initial bootstrap.
type BootstrapResult struct {
	User      store.User
	Namespace store.Namespace
	Repo      store.Repo
	RawKey    string
	KeyPrefix string
}

// Bootstrap initializes the first administrator account, namespace, admin grant,
// initial API key, and starter repository with an initial commit.
// It fails if any users already exist in the database.
func Bootstrap(ctx context.Context, cfg config.Config, email, namespaceSlug string) (*BootstrapResult, error) {
	if email == "" {
		return nil, errors.New("bootstrap: email cannot be empty")
	}
	if namespaceSlug == "" {
		return nil, errors.New("bootstrap: namespace cannot be empty")
	}

	application, err := Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	defer application.Close()

	count, err := application.Store.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: count users: %w", err)
	}
	if count > 0 {
		return nil, errors.New("bootstrap: instance is already initialized (users exist)")
	}

	// 1. Create User
	userID := idgen.NewUser()
	if _, err := application.Store.UpsertUserByEmail(ctx, email, func() string { return userID }); err != nil {
		return nil, fmt.Errorf("bootstrap: create user: %w", err)
	}
	user, err := application.Store.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: get user: %w", err)
	}

	// 2. Create Namespace
	nsID := idgen.NewNS()
	ns := store.Namespace{
		ID:          nsID,
		Slug:        namespaceSlug,
		Type:        "user",
		Name:        namespaceSlug,
		OwnerUserID: userID,
	}
	if err := application.Store.CreateNamespace(ctx, ns); err != nil {
		return nil, fmt.Errorf("bootstrap: create namespace: %w", err)
	}

	// 3. Create Admin Grant
	grant := store.Grant{
		ID:            idgen.New("grant"),
		Scope:         "namespace",
		NamespaceID:   nsID,
		PrincipalKind: "user",
		UserID:        userID,
		Permission:    "PERMISSION_ADMIN",
	}
	if _, err := application.Store.UpsertGrant(ctx, grant); err != nil {
		return nil, fmt.Errorf("bootstrap: create admin grant: %w", err)
	}

	// 4. Generate High-Entropy API Key
	rawKey, prefix, hash, err := auth.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("bootstrap: generate key: %w", err)
	}

	keyID := idgen.New("key")
	apiKey := store.APIKey{
		ID:        keyID,
		UserID:    userID,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Name:      "bootstrap-admin-key",
		Scopes:    "*",
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := application.Store.CreateAPIKey(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("bootstrap: store key: %w", err)
	}

	// 5. Create First Repository ("demo")
	repoName := "demo"
	fullName := namespaceSlug + "/" + repoName
	repoID := idgen.NewRepo()
	repoRecord := store.Repo{
		ID:                  repoID,
		NamespaceID:         nsID,
		Name:                repoName,
		FullName:            fullName,
		Description:         "Initial repository created during bootstrap",
		DefaultBranch:       "main",
		Visibility:          "public",
		CloneURL:            fmt.Sprintf("%s/%s.git", cfg.GitURL, fullName),
		WebURL:              fmt.Sprintf("%s/%s", cfg.PublicURL, fullName),
		AllowMergeCommit:    true,
		AllowSquashMerge:    true,
		DeleteBranchOnMerge: false,
	}
	if err := application.Store.CreateRepo(ctx, repoRecord); err != nil {
		return nil, fmt.Errorf("bootstrap: create initial repo: %w", err)
	}

	if _, err := application.Git.Init(namespaceSlug, repoName, "main"); err != nil {
		return nil, fmt.Errorf("bootstrap: git init: %w", err)
	}

	initialChanges := []gitstore.FileChange{
		{
			Path:    "README.md",
			Content: fmt.Sprintf("# %s\n\nWelcome to exist.io — the neutral, self-hosted Git forge for coding agents.\n", repoName),
		},
	}
	sig := object.Signature{
		Name:  "exist.io",
		Email: email,
		When:  time.Now().UTC(),
	}
	if _, err := application.Git.WriteFiles(namespaceSlug, repoName, "main", initialChanges, sig, "Initial commit"); err != nil {
		return nil, fmt.Errorf("bootstrap: write initial commit: %w", err)
	}

	return &BootstrapResult{
		User:      user,
		Namespace: ns,
		Repo:      repoRecord,
		RawKey:    rawKey,
		KeyPrefix: prefix,
	}, nil
}
