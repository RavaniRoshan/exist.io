package origin

import "net/http"

// registerMirror wires the handler group's routes and operations.
// Implemented in the mirror task; currently a no-op placeholder so the
// router compiles before the group lands.
func (s *Server) registerMirror(mux *http.ServeMux) {
	_ = mux
}
