package security

import (
	"fmt"
	"testing"
)

func TestScannerDetectsSecrets(t *testing.T) {
	scanner := NewScanner()

	dummySlack := "xoxb" + "-123456789012-345678901234-abcdefghijklmnopqrstuvwx"
	dummyAWS := "AKIA" + "IOSFODNN7EXAMPLE"

	diff := []byte(fmt.Sprintf(`
diff --git a/config.go b/config.go
index 1234..5678 100644
--- a/config.go
+++ b/config.go
@@ -10,3 +10,4 @@
-	awsKey = "placeholder"
+	awsKey = %q
+	slack = %q
+	safe = "hello-world"
`, dummyAWS, dummySlack))

	violations := scanner.ScanDiff("config.go", diff)

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if violations[0].PatternID != "aws-access-key" {
		t.Errorf("expected aws-access-key, got %s", violations[0].PatternID)
	}

	if violations[1].PatternID != "slack-token" {
		t.Errorf("expected slack-token, got %s", violations[1].PatternID)
	}

	rejectionMsg := FormatPushRejection(violations)
	if rejectionMsg == "" {
		t.Errorf("expected non-empty rejection message")
	}
}

func TestScannerCleanDiffPasses(t *testing.T) {
	scanner := NewScanner()

	cleanDiff := []byte(`
+func Add(a, b int) int {
+	return a + b
+}
`)

	violations := scanner.ScanDiff("math.go", cleanDiff)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations on clean diff, got %d", len(violations))
	}
}

func TestScannerDetectsExtendedSecrets(t *testing.T) {
	scanner := NewScanner()

	cases := []struct {
		name      string
		diff      string
		wantID    string
	}{
		{
			name:   "PEM Private Key",
			diff:   "+-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA0...",
			wantID: "private-key",
		},
		{
			name:   "OpenSSH Private Key",
			diff:   "+-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXk...",
			wantID: "private-key",
		},
		{
			name:   "Generic PEM Private Key",
			diff:   "+-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASC...",
			wantID: "private-key",
		},
		{
			name:   "GitHub Personal Access Token",
			diff:   "+var token = \"ghp_123456789012345678901234567890123456\"",
			wantID: "github-token",
		},
		{
			name:   "GitHub OAuth Token",
			diff:   "+var oauth = \"gho_123456789012345678901234567890123456\"",
			wantID: "github-token",
		},
		{
			name:   "GitHub Fine-Grained Token",
			diff:   "+var fg = \"github_pat_11AAAAAAA0123456789012345678901234567890123456789012345678901234567890123456789012\"",
			wantID: "github-fine-grained-pat",
		},
		{
			name:   "AWS Secret Access Key",
			diff:   "+aws_secret_access_key = \"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\"",
			wantID: "aws-secret-key",
		},
		{
			name:   "Generic High Entropy API Key",
			diff:   "+api_key = \"ab12cd34ef56ab12cd34ef56ab12cd34ef56\"",
			wantID: "generic-high-entropy",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			violations := scanner.ScanDiff("secret.txt", []byte(tc.diff))
			if len(violations) == 0 {
				t.Fatalf("expected violation for %s, got none", tc.name)
			}
			if violations[0].PatternID != tc.wantID {
				t.Errorf("got pattern ID %q, want %q", violations[0].PatternID, tc.wantID)
			}
		})
	}
}
