package agentactivity

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"exist.io/pkg/idgen"
)

// Activity represents an event executed by an autonomous AI coding agent.
type Activity struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	AgentName  string    `json:"agentName"`
	AgentType  string    `json:"agentType"` // "claude-code", "cursor", "codex", "exist-mcp", "reviewer"
	Action     string    `json:"action"`
	ToolName   string    `json:"toolName,omitempty"`
	Resource   string    `json:"resource"`
	Status     string    `json:"status"` // "success", "failed", "in_progress"
	DurationMs int       `json:"durationMs,omitempty"`
	Details    string    `json:"details,omitempty"`
}

// Hub maintains an in-memory thread-safe bounded buffer of recent agent operations.
type Hub struct {
	mu         sync.RWMutex
	maxEntries int
	entries    []Activity
}

// NewHub creates a new agent activity hub with the given capacity limit.
func NewHub(maxEntries int) *Hub {
	if maxEntries <= 0 {
		maxEntries = 200
	}
	return &Hub{
		maxEntries: maxEntries,
		entries:    make([]Activity, 0, maxEntries),
	}
}

// DefaultHub is the singleton activity hub used by existd.
var DefaultHub = NewHub(200)

// Record adds an activity entry to the buffer.
func (h *Hub) Record(act Activity) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if act.ID == "" {
		act.ID = idgen.New("act")
	}
	if act.Timestamp.IsZero() {
		act.Timestamp = time.Now().UTC()
	}

	h.entries = append(h.entries, act)
	if len(h.entries) > h.maxEntries {
		h.entries = h.entries[len(h.entries)-h.maxEntries:]
	}
}

// Recent returns the most recent agent activities up to limit, newest first.
func (h *Hub) Recent(limit int) []Activity {
	h.mu.RLock()
	defer h.mu.RUnlock()

	n := len(h.entries)
	if n == 0 {
		return []Activity{}
	}
	if limit <= 0 || limit > n {
		limit = n
	}

	out := make([]Activity, limit)
	for i := 0; i < limit; i++ {
		out[i] = h.entries[n-1-i]
	}
	return out
}

// Clear resets the activity buffer (used in testing).
func (h *Hub) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = make([]Activity, 0, h.maxEntries)
}

// Handler returns an HTTP handler for streaming recent agent operations.
func Handler(h *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		activities := h.Recent(100)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"activities": activities,
		})
	}
}
