// Package security provides enterprise security scanning and push protection for exist.io.
package security

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// SecretPattern defines a compiled regular expression and metadata for secret detection.
type SecretPattern struct {
	ID          string
	Description string
	Regex       *regexp.Regexp
}

// SecretViolation records an identified credential in scanned commit diffs or file contents.
type SecretViolation struct {
	PatternID   string `json:"patternId"`
	Description string `json:"description"`
	LineNumber  int    `json:"lineNumber"`
	Snippet     string `json:"snippet"`
	Path        string `json:"path"`
}

// Scanner evaluates diffs or file buffers against enterprise push-protection rules.
type Scanner struct {
	patterns []SecretPattern
}

// NewScanner constructs a Scanner pre-configured with industry standard secret detection rules.
func NewScanner() *Scanner {
	return &Scanner{
		patterns: []SecretPattern{
			{
				ID:          "aws-access-key",
				Description: "AWS Access Key ID",
				Regex:       regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
			},
			{
				ID:          "aws-secret-key",
				Description: "AWS Secret Access Key Assignment",
				Regex:       regexp.MustCompile(`(?i)\baws_secret_access_key\s*[:=]\s*["']?[A-Za-z0-9/+=]{40}["']?`),
			},
			{
				ID:          "slack-token",
				Description: "Slack API or Bot Token",
				Regex:       regexp.MustCompile(`\bxox[baprs]-[0-9a-zA-Z]{10,48}\b`),
			},
			{
				ID:          "private-key",
				Description: "Unencrypted Private Key Block",
				Regex:       regexp.MustCompile(`-----BEGIN (?:[A-Z0-9_-]+ )?PRIVATE KEY`),
			},
			{
				ID:          "github-token",
				Description: "GitHub Token",
				Regex:       regexp.MustCompile(`\b(?:ghp|gho|ghu|ghs|ghr)_[0-9a-zA-Z]{36}\b`),
			},
			{
				ID:          "github-fine-grained-pat",
				Description: "GitHub Fine-Grained Personal Access Token",
				Regex:       regexp.MustCompile(`\bgithub_pat_[0-9a-zA-Z_]{82}\b`),
			},
			{
				ID:          "generic-high-entropy",
				Description: "Generic High-Entropy API Key or Secret Assignment",
				Regex:       regexp.MustCompile(`(?i)\b(?:api[_-]?key|secret[_-]?key|auth[_-]?token|access[_-]?token)\s*[:=]\s*["'][A-Za-z0-9+/=_-]{32,}["']`),
			},
			{
				ID:          "generic-jwt-secret",
				Description: "High-Entropy JSON Web Token (JWT)",
				Regex:       regexp.MustCompile(`\bey[A-Za-z0-9_-]{20,}\.ey[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}\b`),
			},
		},
	}
}

// DefaultScanner is the package-level secret scanner instance.
var DefaultScanner = NewScanner()

// ScanFile inspects raw file contents for credentials.
func (s *Scanner) ScanFile(path string, content []byte) []SecretViolation {
	var violations []SecretViolation
	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		for _, p := range s.patterns {
			if loc := p.Regex.FindStringIndex(trimmed); loc != nil {
				match := trimmed[loc[0]:loc[1]]
				violations = append(violations, SecretViolation{
					PatternID:   p.ID,
					Description: p.Description,
					LineNumber:  lineNum,
					Snippet:     redactMatch(match),
					Path:        path,
				})
			}
		}
	}
	return violations
}

// ScanDiff analyzes unified diff text or commit content, flagging newly introduced secrets.
func (s *Scanner) ScanDiff(path string, content []byte) []SecretViolation {
	var violations []SecretViolation

	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// If scanning diff format, only inspect added lines (starts with '+', but not '+++')
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(line, "+++") {
			continue
		}
		if strings.HasPrefix(line, "+") {
			trimmed = strings.TrimSpace(line[1:])
		}

		for _, p := range s.patterns {
			if loc := p.Regex.FindStringIndex(trimmed); loc != nil {
				// Redact snippet for safe logging and presentation
				match := trimmed[loc[0]:loc[1]]
				redacted := redactMatch(match)

				violations = append(violations, SecretViolation{
					PatternID:   p.ID,
					Description: p.Description,
					LineNumber:  lineNum,
					Snippet:     redacted,
					Path:        path,
				})
			}
		}
	}

	return violations
}

// redactMatch masks intermediate characters in detected secrets for audit output.
func redactMatch(s string) string {
	if len(s) <= 8 {
		return "********"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

// Error formats secret violations into a Git pre-receive rejection message.
func FormatPushRejection(violations []SecretViolation) string {
	var sb strings.Builder
	sb.WriteString("\n[EXIST.IO ENTERPRISE SECURITY: PUSH REJECTED]\n")
	sb.WriteString("----------------------------------------------------------------\n")
	sb.WriteString("Push protection detected sensitive credentials in your commits.\n")
	sb.WriteString("Please remove the secrets, amend your commits, and push again.\n\n")

	for _, v := range violations {
		sb.WriteString(fmt.Sprintf("  • File: %s:%d\n", v.Path, v.LineNumber))
		sb.WriteString(fmt.Sprintf("    Type: %s (%s)\n", v.Description, v.PatternID))
		sb.WriteString(fmt.Sprintf("    Credential: %s\n\n", v.Snippet))
	}
	sb.WriteString("----------------------------------------------------------------\n")
	sb.WriteString("To configure enterprise bypass exceptions, contact your forge administrator.\n")
	return sb.String()
}
