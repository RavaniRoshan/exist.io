package origin

import (
	"net/http"

	"exist.io/pkg/scopes"
)

// Handler returns the fully wired Origin HTTP handler, mounting every resource
// group under /v1/origin/. Handler groups register their own operations and
// routes; this is the single entry point the server binary mounts.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	s.registerRepos(mux)    // Task 6
	s.registerGrants(mux)   // Task 7
	s.registerContents(mux) // Task 9
	s.registerRefs(mux)     // Task 10
	s.registerPulls(mux)    // Task 11
	s.registerPRSocial(mux) // Task 12
	s.registerChecks(mux)   // Task 13
	s.registerRulesets(mux) // Task 14
	s.registerSSHAndIP(mux) // Task 15
	s.registerMirror(mux)   // Task 16
	s.registerApps(mux)     // Task 17
	s.registerWebhooks(mux) // Task 18

	return mux
}

// The registerX methods are defined in their respective handler files. Until a
// group is implemented, its file provides a no-op that registers nothing, so
// the router always compiles. Each group registers operations (via s.register)
// and guarded routes (via s.guard).

// Common operation scope requirements, shared across groups. These mirror the
// Origin OpenAPI x-origin-scopes extensions verbatim.

var (
	reqListRepos      = scopes.Requirements{TokenTypes: []string{scopes.TypeUser}, Scopes: []string{scopes.NamespaceRepositoriesRead}}
	reqCreateRepo     = scopes.Requirements{TokenTypes: []string{scopes.TypeUser}, Scopes: []string{scopes.NamespaceRepositoriesCreate}}
	reqGetRepo        = scopes.Requirements{TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser}, Scopes: []string{scopes.RepositoryMetadataRead}}
	reqUpdateRepo     = scopes.Requirements{TokenTypes: []string{scopes.TypeUser}, Scopes: []string{scopes.NamespaceSettingsWrite}}
	reqListNamespaces = scopes.Requirements{TokenTypes: []string{scopes.TypeInstallation, scopes.TypeUser}, Scopes: []string{scopes.InstallationMetadataRead}}
)
