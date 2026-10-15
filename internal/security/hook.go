package security

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"exist.io/internal/audit"
	"exist.io/internal/metrics"
)

// PreReceiveHook processes standard Git pre-receive input lines:
// <old-rev> <new-rev> <ref-name>
// It diffs commits and rejects any push containing detected secrets.
func PreReceiveHook(r io.Reader, stderr io.Writer) error {
	scanner := bufio.NewScanner(r)
	var allViolations []SecretViolation

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		oldRev, newRev, refName := parts[0], parts[1], parts[2]

		// Ref deletion: all zeros
		if strings.Trim(newRev, "0") == "" {
			continue
		}

		var cmd *exec.Cmd
		if strings.Trim(oldRev, "0") == "" {
			// New branch/ref: diff against Git's canonical empty tree hash
			cmd = exec.Command("git", "diff", "4b825dc642cb6eb9a060e54bf8d69288fbee4904", newRev)
		} else {
			cmd = exec.Command("git", "diff", oldRev, newRev)
		}

		diffOut, err := cmd.Output()
		if err != nil {
			// Fallback to git log -p if diff against empty tree fails
			cmd = exec.Command("git", "log", "-p", "-n", "25", newRev)
			diffOut, _ = cmd.Output()
		}

		if len(diffOut) > 0 {
			violations := DefaultScanner.ScanDiff(refName, diffOut)
			allViolations = append(allViolations, violations...)
		}
	}

	if len(allViolations) > 0 {
		metrics.DefaultRegistry.IncSecretPushBlocks()
		_ = audit.Log(audit.Event{
			Actor:     "git-client",
			ActorType: "user",
			Action:    "security.secret_push_blocked",
			Resource:  allViolations[0].Path,
			Severity:  audit.SeverityCritical,
			Details:   fmt.Sprintf("Blocked push containing %d detected secret(s)", len(allViolations)),
		})
		fmt.Fprintln(stderr, FormatPushRejection(allViolations))
		return fmt.Errorf("push rejected: sensitive credentials detected")
	}

	return nil
}
