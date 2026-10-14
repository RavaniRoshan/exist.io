package origin

import "net/http"

// registerPulls wires the handler group's routes and operations.
// Implemented in the pulls task; currently a no-op placeholder so the
// router compiles before the group lands.
func (s *Server) registerPulls(mux *http.ServeMux) {
	_ = mux
}
