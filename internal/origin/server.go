package origin

import (
	"net/http"

	"exist.io/internal/auth"
	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/scopes"
)

// Server holds the shared dependencies every Origin handler needs: the
// metadata store, the git object store, and the operation registry the auth
// middleware consults. Handlers are methods on Server.
type Server struct {
	store    *store.DB
	git      *gitstore.Store
	registry *auth.Registry
	// endpoint/gitHost feed derived URLs (cloneUrl, webUrl) on responses.
	endpoint string
	gitHost  string
	// resolve, when set, overrides BearerResolver in guard(). Tests inject a
	// resolver that returns a fixed credential with known scopes.
	resolve auth.Resolver
}

// New builds a Server. registry may be nil (a fresh one is created).
func New(db *store.DB, git *gitstore.Store, registry *auth.Registry, endpoint, gitHost string) *Server {
	if registry == nil {
		registry = auth.NewRegistry()
	}
	return &Server{store: db, git: git, registry: registry, endpoint: endpoint, gitHost: gitHost}
}

// Registry exposes the operation registry so callers can register operations
// and inspect enforcement metadata.
func (s *Server) Registry() *auth.Registry { return s.registry }

// Store exposes the metadata store for handler groups in other files.
func (s *Server) Store() *store.DB { return s.store }

// Git exposes the git object store.
func (s *Server) Git() *gitstore.Store { return s.git }

// register is a small helper for handler groups to declare an operation's
// scope requirements.
func (s *Server) register(operationID string, reqs scopes.Requirements) {
	s.registry.Register(operationID, reqs)
}

// guard wraps a handler with auth enforcement for a registered operation. The
// handler runs only if the resolved credential satisfies the operation's
// x-origin-scopes requirements. If s.resolve is set (tests), it is used;
// otherwise the production BearerResolver applies.
func (s *Server) guard(operationID string, h http.HandlerFunc) http.Handler {
	resolve := s.resolve
	if resolve == nil {
		resolve = auth.BearerResolver
	}
	return auth.RequireWith(s.registry, operationID, resolve, h)
}

// cloneURL builds the HTTPS clone URL for a repo.
func (s *Server) cloneURL(ownerSlug, repoName string) string {
	return "https://" + s.gitHost + "/" + ownerSlug + "/" + repoName + ".git"
}

// webURL builds the browser URL for a repo.
func (s *Server) webURL(ownerSlug, repoName string) string {
	return "https://cursor.com/codebase/" + ownerSlug + "/" + repoName
}
