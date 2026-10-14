package origin

import "net/http"

// registerChecks wires the handler group's routes and operations.
// Implemented in the checks task; currently a no-op placeholder so the
// router compiles before the group lands.
func (s *Server) registerChecks(mux *http.ServeMux) {
	_ = mux
}
