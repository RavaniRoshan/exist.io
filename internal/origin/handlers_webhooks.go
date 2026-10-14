package origin

import "net/http"

// registerWebhooks wires the handler group's routes and operations.
// Implemented in the webhooks task; currently a no-op placeholder so the
// router compiles before the group lands.
func (s *Server) registerWebhooks(mux *http.ServeMux) {
	_ = mux
}
