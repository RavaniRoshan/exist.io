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

func TestPreReceiveHook_ProtectedBranchProtection(t *testing.T) {
	// Attempting to delete protected default branch refs/heads/main
	input := bytes.NewBufferString("1111111111111111111111111111111111111111 0000000000000000000000000000000000000000 refs/heads/main\n")
	var stderr bytes.Buffer
	err := PreReceiveHook(input, &stderr)
	if err == nil {
		t.Fatalf("expected error rejecting deletion of protected branch main, got nil")
	}
	if !strings.Contains(stderr.String(), "GH006") {
		t.Errorf("expected GH006 error code in stderr, got: %s", stderr.String())
	}

	// Deleting a feature branch should not trigger the protected branch deletion error
	inputFeat := bytes.NewBufferString("1111111111111111111111111111111111111111 0000000000000000000000000000000000000000 refs/heads/feature-x\n")
	var stderrFeat bytes.Buffer
	errFeat := PreReceiveHook(inputFeat, &stderrFeat)
	if errFeat != nil {
		t.Fatalf("expected feature branch deletion to succeed without error, got: %v", errFeat)
	}
}
