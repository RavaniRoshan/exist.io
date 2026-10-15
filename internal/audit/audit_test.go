package audit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestAuditLoggerRecordsEvents(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	ev := Event{
		Actor:     "roshan@exist.io",
		ActorType: "user",
		Action:    "repo.create_ruleset",
		Resource:  "exist/exist.io:rs-main",
		ClientIP:  "192.168.1.100",
		Severity:  SeverityInfo,
		Details:   "Branch ruleset created",
	}

	if err := logger.Record(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "repo.create_ruleset") {
		t.Errorf("expected output to contain action, got: %s", output)
	}

	var parsed Event
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if parsed.Actor != "roshan@exist.io" {
		t.Errorf("expected actor roshan@exist.io, got %s", parsed.Actor)
	}
	if parsed.ID == "" {
		t.Errorf("expected auto-generated ID, got empty")
	}
	if parsed.Timestamp.IsZero() {
		t.Errorf("expected auto-generated Timestamp, got zero")
	}
}

func TestAuditLoggerRecent(t *testing.T) {
	logger := NewLogger(nil)
	SetDefaultLogger(logger)

	for i := 1; i <= 5; i++ {
		_ = Log(Event{
			Actor:    "admin",
			Action:   "test.action",
			Resource: "test",
			Details:  "item",
		})
	}

	recent := Recent(3)
	if len(recent) != 3 {
		t.Fatalf("expected 3 events, got %d", len(recent))
	}
}
