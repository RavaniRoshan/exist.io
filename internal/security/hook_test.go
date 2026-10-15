package security

import (
	"bytes"
	"strings"
	"testing"
)

func TestPreReceiveHookRejection(t *testing.T) {
	// Simulate input to hook where git diff would return diff content
	// If input is empty lines, hook should pass cleanly
	input := bytes.NewBufferString("\n")
	var stderr bytes.Buffer

	err := PreReceiveHook(input, &stderr)
	if err != nil {
		t.Fatalf("expected nil error on empty input, got: %v", err)
	}

	// Test ScanFile directly
	dummyKey := "AKIA" + "1234567890ABCDEF"
	sampleCodeWithSecret := []byte("package main\n\nconst AWS_KEY = \"" + dummyKey + "\"\n")
	violations := DefaultScanner.ScanFile("main.go", sampleCodeWithSecret)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].PatternID != "aws-access-key" {
		t.Errorf("expected aws-access-key pattern, got %s", violations[0].PatternID)
	}

	rejectionMsg := FormatPushRejection(violations)
	if !strings.Contains(rejectionMsg, "PUSH REJECTED") {
		t.Errorf("expected rejection header, got: %s", rejectionMsg)
	}
}
