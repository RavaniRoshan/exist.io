package agentlease

import (
	"encoding/json"
	"net/http"
	"time"
)

// LeaseRequest payload for issuing a short-lived capability lease.
type LeaseRequest struct {
	AgentID               string   `json:"agentId"`
	RepoID                string   `json:"repoId"`
	AllowedRefPrefixes    []string `json:"allowedRefPrefixes"`
	DisallowedRefPrefixes []string `json:"disallowedRefPrefixes"`
	TTLSeconds            int      `json:"ttlSeconds"`
}

// LeaseResponse returns the signed capability lease token and metadata.
type LeaseResponse struct {
	Token     string `json:"token"`
	ID        string `json:"id"`
	AgentID   string `json:"agentId"`
	RepoID    string `json:"repoId"`
	IssuedAt  string `json:"issuedAt"`
	ExpiresAt string `json:"expiresAt"`
}

// Handler returns an HTTP handler for issuing signed agent capability leases.
func Handler(signingSecret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LeaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.AgentID == "" {
			req.AgentID = "agent:generic"
		}
		if req.RepoID == "" {
			req.RepoID = "*"
		}
		if len(req.AllowedRefPrefixes) == 0 {
			req.AllowedRefPrefixes = []string{"refs/heads/agent/", "refs/heads/feat/"}
		}

		ttl := 15 * time.Minute
		if req.TTLSeconds > 0 {
			ttl = time.Duration(req.TTLSeconds) * time.Second
		}

		token, lease, err := Issue(signingSecret, req.AgentID, req.RepoID, req.AllowedRefPrefixes, req.DisallowedRefPrefixes, ttl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LeaseResponse{
			Token:     token,
			ID:        lease.ID,
			AgentID:   lease.AgentID,
			RepoID:    lease.RepoID,
			IssuedAt:  lease.IssuedAt.Format(time.RFC3339),
			ExpiresAt: lease.ExpiresAt.Format(time.RFC3339),
		})
	}
}
