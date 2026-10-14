package scopes

import (
	"strings"
	"testing"
)

// fakeToken is a minimal Token for exercising enforcement without depending on
// internal/auth (which imports pkg/scopes, so the reverse import would cycle).
type fakeToken struct {
	typ    string
	scopes []string
}

func (f fakeToken) Type() string     { return f.typ }
func (f fakeToken) Scopes() []string { return f.scopes }

// Representative operations transcribed from the Origin OpenAPI spec's
// x-origin-scopes extension.
var (
	listRepos = Requirements{ // OriginService_ListRepos
		TokenTypes: []string{TypeUser},
		Scopes:     []string{NamespaceRepositoriesRead},
	}
	createRepo = Requirements{ // OriginService_CreateRepo
		TokenTypes: []string{TypeUser},
		Scopes:     []string{NamespaceRepositoriesCreate},
	}
	getRepo = Requirements{ // OriginService_GetRepo
		TokenTypes: []string{TypeInstallation, TypeUser},
		Scopes:     []string{RepositoryMetadataRead},
	}
	getAuthenticatedApp = Requirements{ // GET /v1/origin/app
		TokenTypes: []string{TypeApp},
		Scopes:     []string{AppMetadataRead},
		Ambient:    true,
	}
	listInstallationRepos = Requirements{ // GET /v1/origin/installation/repos
		TokenTypes: []string{TypeInstallation},
		Scopes:     []string{InstallationMetadataRead},
		Ambient:    true,
	}
	createInstallationUserToken = Requirements{ // POST .../installations/{id}/user_access_tokens
		TokenTypes:         []string{TypeApp},
		Scopes:             []string{AppMetadataRead},
		Ambient:            true,
		InstallationScopes: []string{NamespaceUserTokensWrite},
	}
)

func TestAllCatalog(t *testing.T) {
	all := All()
	// Origin's OpenAPI spec lists 33 scopes in the x-origin-scopes.scopes field
	// across all 120 operations, plus namespace:user_tokens:write which appears
	// only in installationScopes (it gates user_access_tokens). The full catalog
	// enforcement recognizes is therefore 34.
	if len(all) != 34 {
		t.Fatalf("All() returned %d scopes, want 34 (33 in scopes + 1 installation-only)", len(all))
	}
	seen := map[string]bool{}
	for _, s := range all {
		if seen[s] {
			t.Errorf("duplicate scope in catalog: %q", s)
		}
		seen[s] = true
		if !strings.Contains(s, ":") {
			t.Errorf("scope %q is not in resource:action[:sub] form", s)
		}
	}
	for _, want := range []string{
		"app:metadata:read", "app:webhook_deliveries:write",
		"installation:metadata:read", "namespace:repositories:create",
		"repository:pull_requests:reviews:read", "repository:mirror:sync",
		"repository:settings:write",
	} {
		if !seen[want] {
			t.Errorf("catalog missing scope %q", want)
		}
	}
}

// TestRequireUserListVsCreate is the brief's core assertion: a user token with
// namespace:repositories:read passes ListRepos but fails CreateRepo.
func TestRequireUserListVsCreate(t *testing.T) {
	reader := fakeToken{typ: TypeUser, scopes: []string{NamespaceRepositoriesRead}}

	if err := Require(reader, listRepos); err != nil {
		t.Errorf("Require(reader, ListRepos) = %v, want nil", err)
	}
	err := Require(reader, createRepo)
	if err == nil {
		t.Fatal("Require(reader, CreateRepo) = nil, want PERMISSION_DENIED")
	}
	var denied *Denied
	if !asDenied(err, &denied) {
		t.Fatalf("Require error type = %T, want *Denied", err)
	}
	if denied.Code != CodePermissionDenied {
		t.Errorf("denied.Code = %d, want %d", denied.Code, CodePermissionDenied)
	}
	if !strings.Contains(err.Error(), NamespaceRepositoriesCreate) {
		t.Errorf("error = %q, want it to name missing scope %q", err.Error(), NamespaceRepositoriesCreate)
	}
}

func TestRequireWrongTokenType(t *testing.T) {
	user := fakeToken{typ: TypeUser, scopes: []string{AppMetadataRead}}
	err := Require(user, getAuthenticatedApp)
	if err == nil {
		t.Fatal("Require(user, GetAuthenticatedApp) = nil, want PERMISSION_DENIED")
	}
	var denied *Denied
	if !asDenied(err, &denied) {
		t.Fatalf("error type = %T, want *Denied", err)
	}
	if denied.Code != CodePermissionDenied {
		t.Errorf("denied.Code = %d, want %d", denied.Code, CodePermissionDenied)
	}
}

func TestRequireInstallationAndUserShareOp(t *testing.T) {
	inst := fakeToken{typ: TypeInstallation, scopes: []string{RepositoryMetadataRead}}
	usr := fakeToken{typ: TypeUser, scopes: []string{RepositoryMetadataRead}}
	if err := Require(inst, getRepo); err != nil {
		t.Errorf("Require(install, GetRepo) = %v, want nil", err)
	}
	if err := Require(usr, getRepo); err != nil {
		t.Errorf("Require(user, GetRepo) = %v, want nil", err)
	}
	bare := fakeToken{typ: TypeUser, scopes: nil}
	if err := Require(bare, getRepo); err == nil {
		t.Error("Require(scopeless user, GetRepo) = nil, want PERMISSION_DENIED")
	}
}

// TestRequireAmbient: ambient scopes come with the credential itself, so an app
// or installation token satisfies the scope check without listing the scope.
func TestRequireAmbient(t *testing.T) {
	scopelessApp := fakeToken{typ: TypeApp}
	if err := Require(scopelessApp, getAuthenticatedApp); err != nil {
		t.Errorf("Require(scopeless app, ambient GetAuthenticatedApp) = %v, want nil", err)
	}
	scopelessInstall := fakeToken{typ: TypeInstallation}
	if err := Require(scopelessInstall, listInstallationRepos); err != nil {
		t.Errorf("Require(scopeless install, ambient ListInstallationRepos) = %v, want nil", err)
	}
	user := fakeToken{typ: TypeUser}
	if err := Require(user, getAuthenticatedApp); err == nil {
		t.Error("Require(user, ambient app op) = nil, want PERMISSION_DENIED (wrong type)")
	}
}

// TestRequireInstallationScopes: an installation token must additionally hold
// every scope listed in InstallationScopes.
func TestRequireInstallationScopes(t *testing.T) {
	req := Requirements{
		TokenTypes:         []string{TypeInstallation},
		Scopes:             []string{RepositoryMetadataRead},
		InstallationScopes: []string{NamespaceUserTokensWrite},
	}
	withExtra := fakeToken{typ: TypeInstallation, scopes: []string{RepositoryMetadataRead, NamespaceUserTokensWrite}}
	if err := Require(withExtra, req); err != nil {
		t.Errorf("Require(install with extra scope) = %v, want nil", err)
	}
	withoutExtra := fakeToken{typ: TypeInstallation, scopes: []string{RepositoryMetadataRead}}
	err := Require(withoutExtra, req)
	if err == nil {
		t.Fatal("Require(install missing installation scope) = nil, want PERMISSION_DENIED")
	}
	if !strings.Contains(err.Error(), NamespaceUserTokensWrite) {
		t.Errorf("error = %q, want it to name %q", err.Error(), NamespaceUserTokensWrite)
	}
	app := fakeToken{typ: TypeApp}
	if err := Require(app, createInstallationUserToken); err != nil {
		t.Errorf("Require(app, createInstallationUserToken) = %v, want nil (ambient app op)", err)
	}
}

func TestHas(t *testing.T) {
	reader := fakeToken{typ: TypeUser, scopes: []string{NamespaceRepositoriesRead}}
	if !Has(reader, listRepos) {
		t.Error("Has(reader, ListRepos) = false, want true")
	}
	if Has(reader, createRepo) {
		t.Error("Has(reader, CreateRepo) = true, want false")
	}
}

func TestDeniedError(t *testing.T) {
	err := &Denied{Code: CodePermissionDenied, Reason: "missing required scope \"x:y:z\""}
	if !strings.Contains(err.Error(), "7") || !strings.Contains(err.Error(), "x:y:z") {
		t.Errorf("Denied.Error() = %q, want it to include code 7 and the reason", err.Error())
	}
}

// asDenied is a local errors.As to avoid importing "errors" solely for this.
func asDenied(err error, target **Denied) bool {
	for err != nil {
		if d, ok := err.(*Denied); ok {
			*target = d
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
