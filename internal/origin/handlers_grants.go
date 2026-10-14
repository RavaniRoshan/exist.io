package origin

import "net/http"

// registerGrants wires the handler group's routes and operations.
// Implemented in the grants task; currently a no-op placeholder so the
// router compiles before the group lands.
func (s *Server) registerGrants(mux *http.ServeMux) {
	_ = mux
}
