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
