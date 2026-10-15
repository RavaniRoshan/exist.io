package agentactivity

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHub_RecordAndRecent(t *testing.T) {
	hub := NewHub(3)

	hub.Record(Activity{AgentName: "claude-1", Action: "repo_tree"})
	hub.Record(Activity{AgentName: "cursor-2", Action: "file_read"})
	hub.Record(Activity{AgentName: "codex-3", Action: "pull_request_create"})
	hub.Record(Activity{AgentName: "claude-4", Action: "diff_compare"})

	recent := hub.Recent(10)
	if len(recent) != 3 {
		t.Fatalf("expected 3 entries in bounded buffer, got %d", len(recent))
	}

	// Newest first
	if recent[0].AgentName != "claude-4" {
		t.Errorf("expected newest entry to be claude-4, got %s", recent[0].AgentName)
	}
	if recent[1].AgentName != "codex-3" {
		t.Errorf("expected second entry to be codex-3, got %s", recent[1].AgentName)
	}
	if recent[2].AgentName != "cursor-2" {
		t.Errorf("expected third entry to be cursor-2, got %s", recent[2].AgentName)
	}
}

func TestHub_Handler(t *testing.T) {
	hub := NewHub(10)
	hub.Record(Activity{AgentName: "agent-test", Action: "checks_read", Status: "success"})

	handler := Handler(hub)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/agent/activities", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Activities []Activity `json:"activities"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(resp.Activities))
	}
	if resp.Activities[0].AgentName != "agent-test" {
		t.Errorf("expected agent-test, got %s", resp.Activities[0].AgentName)
	}
}
