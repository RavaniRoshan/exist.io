package origin

import "net/http"

// registerRulesets wires the handler group's routes and operations.
// Implemented in the rulesets task; currently a no-op placeholder so the
// router compiles before the group lands.
func (s *Server) registerRulesets(mux *http.ServeMux) {
	_ = mux
}
