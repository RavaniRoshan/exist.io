// Package scopes implements Origin's scope catalog and the enforcement rules
// derived from each operation's x-origin-scopes extension.
//
// The authoritative source is the Origin OpenAPI 3.1 spec. Every operation
// carries an x-origin-scopes object with four fields:
//
//	tokenTypes         credential types allowed to call the operation
//	                   (app, installation, user)
//	scopes             scopes the credential must hold
//	ambient            when true, the listed scopes come with the credential
//	                   itself, so a bare app/installation token satisfies them
//	installationScopes additional scopes an installation token must hold
//
// This package deliberately does NOT import internal/auth: auth imports
// scopes (for the middleware), so the reverse dependency would form a cycle.
// The Token interface below is satisfied structurally by auth.Token.
package scopes

import (
	"fmt"
	"strings"
)

// Token types, matching auth.Type{App,Installation,User}.
const (
	TypeApp          = "app"
	TypeInstallation = "installation"
	TypeUser         = "user"
)

// The 33-scope catalog, transcribed verbatim from the Origin OpenAPI spec's
// x-origin-scopes extensions. Wire values are exact; do not rename.
const (
	AppMetadataRead           = "app:metadata:read"
	AppMetadataWrite          = "app:metadata:write"
	AppSettingsWrite          = "app:settings:write"
	AppWebhookDeliveriesRead  = "app:webhook_deliveries:read"
	AppWebhookDeliveriesWrite = "app:webhook_deliveries:write"

	InstallationMetadataRead = "installation:metadata:read"

	NamespaceAppsCreate         = "namespace:apps:create"
	NamespaceAppsRead           = "namespace:apps:read"
	NamespaceInstallationsWrite = "namespace:installations:write"
	NamespaceRepositoriesCreate = "namespace:repositories:create"
	NamespaceRepositoriesRead   = "namespace:repositories:read"
	NamespaceSettingsRead       = "namespace:settings:read"
	NamespaceSettingsWrite      = "namespace:settings:write"
	NamespaceUserTokensWrite    = "namespace:user_tokens:write"

	RepositoryChecksRead               = "repository:checks:read"
	RepositoryChecksWrite              = "repository:checks:write"
	RepositoryContentsRead             = "repository:contents:read"
	RepositoryContentsWrite            = "repository:contents:write"
	RepositoryLabelsRead               = "repository:labels:read"
	RepositoryLabelsWrite              = "repository:labels:write"
	RepositoryMembersRead              = "repository:members:read"
	RepositoryMetadataRead             = "repository:metadata:read"
	RepositoryMirrorDelete             = "repository:mirror:delete"
	RepositoryMirrorRead               = "repository:mirror:read"
	RepositoryMirrorSync               = "repository:mirror:sync"
	RepositoryMirrorWrite              = "repository:mirror:write"
	RepositoryPullRequestsRead         = "repository:pull_requests:read"
	RepositoryPullRequestsReviewsRead  = "repository:pull_requests:reviews:read"
	RepositoryPullRequestsReviewsWrite = "repository:pull_requests:reviews:write"
	RepositoryPullRequestsWrite        = "repository:pull_requests:write"
	RepositoryRulesetsRead             = "repository:rulesets:read"
	RepositoryRulesetsWrite            = "repository:rulesets:write"
	RepositorySettingsRead             = "repository:settings:read"
	RepositorySettingsWrite            = "repository:settings:write"
)

// catalog is the full ordered set of known scopes.
var catalog = []string{
	AppMetadataRead, AppMetadataWrite, AppSettingsWrite,
	AppWebhookDeliveriesRead, AppWebhookDeliveriesWrite,
	InstallationMetadataRead,
	NamespaceAppsCreate, NamespaceAppsRead, NamespaceInstallationsWrite,
	NamespaceRepositoriesCreate, NamespaceRepositoriesRead,
	NamespaceSettingsRead, NamespaceSettingsWrite, NamespaceUserTokensWrite,
	RepositoryChecksRead, RepositoryChecksWrite,
	RepositoryContentsRead, RepositoryContentsWrite,
	RepositoryLabelsRead, RepositoryLabelsWrite,
	RepositoryMembersRead, RepositoryMetadataRead,
	RepositoryMirrorDelete, RepositoryMirrorRead, RepositoryMirrorSync, RepositoryMirrorWrite,
	RepositoryPullRequestsRead, RepositoryPullRequestsReviewsRead,
	RepositoryPullRequestsReviewsWrite, RepositoryPullRequestsWrite,
	RepositoryRulesetsRead, RepositoryRulesetsWrite,
	RepositorySettingsRead, RepositorySettingsWrite,
}

// All returns a copy of the full scope catalog.
func All() []string {
	out := make([]string, len(catalog))
	copy(out, catalog)
	return out
}

// Known reports whether s is a member of the Origin scope catalog.
func Known(s string) bool {
	for _, c := range catalog {
		if c == s {
			return true
		}
	}
	return false
}

// Token is the minimal credential view enforcement needs. auth.Token (with
// Type/Scopes/Subject) satisfies it structurally.
type Token interface {
	// Type returns the credential type: TypeApp, TypeInstallation, or TypeUser.
	Type() string
	// Scopes returns the scopes granted to the credential, in order.
	Scopes() []string
}

// Requirements is an operation's x-origin-scopes extension.
type Requirements struct {
	// TokenTypes lists the credential types permitted to call the operation.
	TokenTypes []string
	// Scopes lists the scopes the credential must hold (unless Ambient).
	Scopes []string
	// Ambient reports whether the Scopes come with the credential itself, so a
	// bare app/installation token satisfies them without listing them.
	Ambient bool
	// InstallationScopes lists extra scopes an installation token must hold in
	// addition to Scopes.
	InstallationScopes []string
}

// google.rpc.Code values used by the Origin Status envelope.
const (
	CodeOK               = 0
	CodePermissionDenied = 7  // PERMISSION_DENIED
	CodeUnauthenticated  = 16 // UNAUTHENTICATED
)

// Denied is returned when a credential fails scope enforcement. Code carries
// the google.rpc.Code value for the Status envelope.
type Denied struct {
	Code   int
	Reason string
}

func (e *Denied) Error() string {
	return fmt.Sprintf("scopes: permission denied (code %d): %s", e.Code, e.Reason)
}

// Require enforces req against tok. It returns nil when tok may call the
// operation, or a *Denied error otherwise.
//
// The checks, in order:
//  1. The token type must appear in req.TokenTypes.
//  2. Every scope in req.Scopes must be held by the token, unless req.Ambient
//     is set (ambient scopes come with the credential itself).
//  3. When the token is an installation token, every scope in
//     req.InstallationScopes must also be held.
func Require(tok Token, req Requirements) error {
	if !typeAllowed(tok.Type(), req.TokenTypes) {
		return &Denied{
			Code:   CodePermissionDenied,
			Reason: fmt.Sprintf("token type %q is not permitted for this operation (allowed: %s)", tok.Type(), strings.Join(req.TokenTypes, ", ")),
		}
	}

	held := scopeSet(tok.Scopes())

	// Base scopes. Ambient scopes are granted implicitly with the credential,
	// so they are not required to appear in the token's scope list.
	if !req.Ambient {
		for _, s := range req.Scopes {
			if !held[s] {
				return &Denied{
					Code:   CodePermissionDenied,
					Reason: fmt.Sprintf("missing required scope %q", s),
				}
			}
		}
	}

	// Installation tokens must additionally hold InstallationScopes.
	if tok.Type() == TypeInstallation {
		for _, s := range req.InstallationScopes {
			if !held[s] {
				return &Denied{
					Code:   CodePermissionDenied,
					Reason: fmt.Sprintf("installation token missing required scope %q", s),
				}
			}
		}
	}

	return nil
}

// Has reports whether tok satisfies req (Require returns nil).
func Has(tok Token, req Requirements) bool {
	return Require(tok, req) == nil
}

func typeAllowed(typ string, allowed []string) bool {
	for _, a := range allowed {
		if a == typ {
			return true
		}
	}
	return false
}

func scopeSet(scopes []string) map[string]bool {
	m := make(map[string]bool, len(scopes))
	for _, s := range scopes {
		m[s] = true
	}
	return m
}
