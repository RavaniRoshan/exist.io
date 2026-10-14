package origin

import "net/http"

// registerRefs wires the handler group's routes and operations.
// Implemented in the refs task; currently a no-op placeholder so the
// router compiles before the group lands.
func (s *Server) registerRefs(mux *http.ServeMux) {
	_ = mux
}
