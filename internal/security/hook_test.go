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

func TestPreReceiveHook_AgentLeaseEnforcement(t *testing.T) {
	t.Setenv("EXIST_LEASE_TOKEN", "dummy.jwt.token")
	t.Setenv("EXIST_LEASE_ALLOWED_REFS", "refs/heads/agent/")
	t.Setenv("EXIST_LEASE_DISALLOWED_REFS", "refs/heads/main,refs/heads/master")
	t.Setenv("EXIST_LEASE_EXPIRES_AT", "2099-01-01T00:00:00Z")

	// 1. Push to refs/heads/agent/branch passes
	inputValid := bytes.NewBufferString("1111111111111111111111111111111111111111 2222222222222222222222222222222222222222 refs/heads/agent/feature-fix\n")
	var stderrValid bytes.Buffer
	if err := PreReceiveHook(inputValid, &stderrValid); err != nil {
		t.Fatalf("expected valid agent lease push to pass, got: %v (stderr: %s)", err, stderrValid.String())
	}

	// 2. Push to refs/heads/main under lease fails
	inputMain := bytes.NewBufferString("1111111111111111111111111111111111111111 2222222222222222222222222222222222222222 refs/heads/main\n")
	var stderrMain bytes.Buffer
	if err := PreReceiveHook(inputMain, &stderrMain); err == nil {
		t.Fatal("expected lease push to main to be rejected, but got nil error")
	} else if !strings.Contains(stderrMain.String(), "GH007") {
		t.Fatalf("expected GH007 error code, got: %s", stderrMain.String())
	}

	// 3. Push to refs/heads/user-branch (outside allowed prefixes) fails
	inputOther := bytes.NewBufferString("1111111111111111111111111111111111111111 2222222222222222222222222222222222222222 refs/heads/user-branch\n")
	var stderrOther bytes.Buffer
	if err := PreReceiveHook(inputOther, &stderrOther); err == nil {
		t.Fatal("expected lease push to unauthorized ref to be rejected, but got nil error")
	} else if !strings.Contains(stderrOther.String(), "outside allowed prefixes") {
		t.Fatalf("expected outside allowed prefixes error, got: %s", stderrOther.String())
	}

	// 4. Ref deletion under lease fails
	inputDel := bytes.NewBufferString("1111111111111111111111111111111111111111 0000000000000000000000000000000000000000 refs/heads/agent/feature-fix\n")
	var stderrDel bytes.Buffer
	if err := PreReceiveHook(inputDel, &stderrDel); err == nil {
		t.Fatal("expected lease ref deletion to be rejected, got nil error")
	} else if !strings.Contains(stderrDel.String(), "does not permit deleting refs") {
		t.Fatalf("expected deletion prohibited error, got: %s", stderrDel.String())
	}

	// 5. Expired lease fails
	t.Setenv("EXIST_LEASE_EXPIRES_AT", "2020-01-01T00:00:00Z")
	inputExpired := bytes.NewBufferString("1111111111111111111111111111111111111111 2222222222222222222222222222222222222222 refs/heads/agent/feature-fix\n")
	var stderrExpired bytes.Buffer
	if err := PreReceiveHook(inputExpired, &stderrExpired); err == nil {
		t.Fatal("expected expired lease push to be rejected, got nil error")
	} else if !strings.Contains(stderrExpired.String(), "lease has expired") {
		t.Fatalf("expected lease has expired error, got: %s", stderrExpired.String())
	}
}
