package origin

import "net/http"

// registerContents wires the handler group's routes and operations.
// Implemented in the contents task; currently a no-op placeholder so the
// router compiles before the group lands.
func (s *Server) registerContents(mux *http.ServeMux) {
	_ = mux
}
