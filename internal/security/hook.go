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
			if refName == "refs/heads/main" || refName == "refs/heads/master" {
				fmt.Fprintf(stderr, "remote: error: GH006: Protected branch update failed for %s\n", refName)
				fmt.Fprintf(stderr, "remote: error: Deletion of protected default branch is prohibited\n")
				return fmt.Errorf("protected branch deletion rejected")
			}
			continue
		}

		// Force-push detection (non-fast-forward) on protected default branches
		if strings.Trim(oldRev, "0") != "" && (refName == "refs/heads/main" || refName == "refs/heads/master") {
			ancestorCheck := exec.Command("git", "merge-base", "--is-ancestor", oldRev, newRev)
			if err := ancestorCheck.Run(); err != nil {
				fmt.Fprintf(stderr, "remote: error: GH006: Protected branch update failed for %s\n", refName)
				fmt.Fprintf(stderr, "remote: error: Force-push (non-fast-forward update) to protected branch is not permitted\n")
				return fmt.Errorf("protected branch force-push rejected")
			}
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
