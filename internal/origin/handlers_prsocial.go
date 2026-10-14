package origin

import "net/http"

// registerPrsocial wires the handler group's routes and operations.
// Implemented in the prsocial task; currently a no-op placeholder so the
// router compiles before the group lands.
func (s *Server) registerPRSocial(mux *http.ServeMux) {
	_ = mux
}
