package origin

import "net/http"

// registerApps wires the handler group's routes and operations.
// Implemented in the apps task; currently a no-op placeholder so the
// router compiles before the group lands.
func (s *Server) registerApps(mux *http.ServeMux) {
	_ = mux
}
