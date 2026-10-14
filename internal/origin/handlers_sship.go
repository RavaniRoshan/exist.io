package origin

import "net/http"

// registerSship wires the handler group's routes and operations.
// Implemented in the sship task; currently a no-op placeholder so the
// router compiles before the group lands.
func (s *Server) registerSSHAndIP(mux *http.ServeMux) {
	_ = mux
}
