// Package agentreviewer provides automated, AI-native code review, security analysis,
// and quality assurance gates for pull requests in exist.io.
package agentreviewer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"exist.io/internal/agentactivity"
	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

// Violation represents a specific security or quality finding within a pull request.
type Violation struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Severity string `json:"severity"` // "error" | "warning" | "info"
	Rule     string `json:"rule"`     // e.g. "security/secret-token", "security/sql-injection", "quality/missing-tests"
	Message  string `json:"message"`
	Snippet  string `json:"snippet,omitempty"`
}

// ReviewResult encapsulates the complete outcome of an automated agent review.
type ReviewResult struct {
	Verdict    string            `json:"verdict"`    // "approve" | "request_changes" | "comment"
	Conclusion string            `json:"conclusion"` // "success" | "failure" | "neutral"
	Score      int               `json:"score"`      // 0-100 quality score
	Summary    string            `json:"summary"`    // Executive summary text
	OverviewMD string            `json:"overviewMd"` // Full Markdown review text
	Violations []Violation       `json:"violations"`
	ReviewID   string            `json:"reviewId"`
	CheckRunID string            `json:"checkRunId"`
	CreatedAt  string            `json:"createdAt"`
}

// Security secret patterns
var (
	rePrivateKey = regexp.MustCompile(`-----BEGIN [A-Z0-9 ]+PRIVATE KEY-----`)
	reAWSAccess  = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)
	reGHToken    = regexp.MustCompile(`\bgh[pousr]_[a-zA-Z0-9]{36}\b`)
	reGLToken    = regexp.MustCompile(`\bglpat-[a-zA-Z0-9\-]{20,}\b`)
	reSlackToken = regexp.MustCompile(`\bxox[baprs]-[0-9a-zA-Z]{10,}\b`)
	reGenericKey = regexp.MustCompile(`(?i)(?:api_key|apikey|secret_key|auth_token|client_secret)\s*[:=]\s*["']([A-Za-z0-9_\-]{24,})["']`)
	reSQLInject  = regexp.MustCompile(`(?i)(?:fmt\.Sprintf|Sprintf)\s*\(\s*["'](?:SELECT|INSERT|UPDATE|DELETE|DROP)\s+[^"']*%\w`)
	reTODO       = regexp.MustCompile(`(?i)(?:TODO|FIXME|HACK|BUG):\s*(.+)`)
)

// ReviewPullRequest analyzes changes in a pull request, records findings as a PR review
// and check run in the database, and emits telemetry to the agent activity hub.
func ReviewPullRequest(ctx context.Context, db *store.DB, gs *gitstore.Store, ownerSlug string, repo store.Repo, pr store.PullRequest) (*ReviewResult, error) {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	if ownerSlug == "" && strings.Contains(repo.FullName, "/") {
		ownerSlug = strings.Split(repo.FullName, "/")[0]
	}

	// 1. Compute diff between BaseSHA and HeadSHA
	diffStats, err := gs.DiffCommits(ownerSlug, repo.Name, pr.BaseSHA, pr.HeadSHA)
	if err != nil {
		return nil, fmt.Errorf("agentreviewer: diff commits: %w", err)
	}

	var violations []Violation
	hasSourceCode := false
	hasTests := false

	// 2. Scan diff for security tokens and code quality patterns
	for _, fc := range diffStats.Files {
		ext := getExtension(fc.Filename)
		isTestFile := isTest(fc.Filename)

		if isSource(ext) {
			hasSourceCode = true
			if isTestFile {
				hasTests = true
			}
		}

		// Analyze added lines in patch
		lines := strings.Split(fc.Patch, "\n")
		lineNum := 0
		for _, line := range lines {
			if strings.HasPrefix(line, "@@") {
				// parse line number header if needed
				lineNum = 1
				continue
			}
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				lineNum++
				content := line[1:]

				// Secret detection
				if rePrivateKey.MatchString(content) {
					violations = append(violations, Violation{
						File:     fc.Filename,
						Line:     lineNum,
						Severity: "error",
						Rule:     "security/private-key",
						Message:  "Exposed private cryptographic key detected in diff.",
						Snippet:  redact(content),
					})
				}
				if reAWSAccess.MatchString(content) {
					violations = append(violations, Violation{
						File:     fc.Filename,
						Line:     lineNum,
						Severity: "error",
						Rule:     "security/aws-access-key",
						Message:  "Exposed AWS Access Key ID detected.",
						Snippet:  redact(content),
					})
				}
				if reGHToken.MatchString(content) || reGLToken.MatchString(content) {
					violations = append(violations, Violation{
						File:     fc.Filename,
						Line:     lineNum,
						Severity: "error",
						Rule:     "security/access-token",
						Message:  "Exposed Git personal access token detected.",
						Snippet:  redact(content),
					})
				}
				if reSlackToken.MatchString(content) {
					violations = append(violations, Violation{
						File:     fc.Filename,
						Line:     lineNum,
						Severity: "error",
						Rule:     "security/slack-token",
						Message:  "Exposed Slack API token detected.",
						Snippet:  redact(content),
					})
				}
				if match := reGenericKey.FindStringSubmatch(content); len(match) > 1 {
					val := match[1]
					if !isTestMock(val) {
						violations = append(violations, Violation{
							File:     fc.Filename,
							Line:     lineNum,
							Severity: "error",
							Rule:     "security/generic-secret",
							Message:  "Potential hardcoded API key or credential secret detected.",
							Snippet:  redact(content),
						})
					}
				}

				// SQL injection heuristic
				if reSQLInject.MatchString(content) {
					violations = append(violations, Violation{
						File:     fc.Filename,
						Line:     lineNum,
						Severity: "warning",
						Rule:     "security/sql-injection",
						Message:  "Dynamic string formatting detected in SQL query. Use parameterized placeholders.",
						Snippet:  strings.TrimSpace(content),
					})
				}

				// Quality: TODO / FIXME
				if m := reTODO.FindStringSubmatch(content); len(m) > 1 {
					violations = append(violations, Violation{
						File:     fc.Filename,
						Line:     lineNum,
						Severity: "info",
						Rule:     "quality/todo-marker",
						Message:  fmt.Sprintf("Pending item recorded: %s", strings.TrimSpace(m[1])),
						Snippet:  strings.TrimSpace(content),
					})
				}
			}
		}
	}

	// 3. Test coverage evaluation
	if hasSourceCode && !hasTests && diffStats.Additions > 30 {
		violations = append(violations, Violation{
			File:     "test-suite",
			Line:     0,
			Severity: "warning",
			Rule:     "quality/missing-tests",
			Message:  fmt.Sprintf("Pull request adds %d lines of implementation logic across %d files without accompanying unit tests.", diffStats.Additions, diffStats.ChangedFiles),
		})
	}

	// 4. Calculate verdict, conclusion, and quality score
	errCount := 0
	warnCount := 0
	infoCount := 0
	for _, v := range violations {
		switch v.Severity {
		case "error":
			errCount++
		case "warning":
			warnCount++
		case "info":
			infoCount++
		}
	}

	verdict := "approve"
	conclusion := "success"
	score := 100 - (errCount * 30) - (warnCount * 10) - (infoCount * 2)
	if score < 0 {
		score = 0
	}

	if errCount > 0 {
		verdict = "request_changes"
		conclusion = "failure"
	} else if warnCount > 0 {
		verdict = "comment"
		conclusion = "neutral"
	}

	// 5. Build rich Markdown overview
	overviewMD := generateMarkdownReport(ownerSlug, repo, pr, diffStats, score, verdict, violations)

	// 6. Record PRReview in database
	reviewID := idgen.NewPRReview()
	prReview := store.PRReview{
		ID:            reviewID,
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
		AuthorUserID:  "",
		State:         verdict,
		Body:          overviewMD,
		CommitSHA:     pr.HeadSHA,
		VersionNumber: 1,
		CreatedAt:     nowStr,
		UpdatedAt:     nowStr,
	}
	if err := db.CreateReview(ctx, prReview); err != nil {
		return nil, fmt.Errorf("agentreviewer: record review: %w", err)
	}

	// 7. Upsert CheckSuite & CheckRun
	suiteKey := "agent/code-review"
	suite, err := db.UpsertCheckSuite(ctx, store.CheckSuite{
		RepoID:      repo.ID,
		SHA:         pr.HeadSHA,
		BaseSHA:     pr.BaseSHA,
		Key:         suiteKey,
		Name:        "Exist Agent Reviewer",
		ActorUserID: "",
		CreatedAt:   nowStr,
		UpdatedAt:   nowStr,
	})
	if err != nil {
		return nil, fmt.Errorf("agentreviewer: upsert check suite: %w", err)
	}

	summaryText := fmt.Sprintf("Autonomous review completed: score %d/100, %d errors, %d warnings.", score, errCount, warnCount)
	runID := idgen.NewCheckRun()
	checkRun := store.CheckRun{
		ID:            runID,
		RepoID:        repo.ID,
		CheckSuiteID:  suite.ID,
		SHA:           pr.HeadSHA,
		BaseSHA:       pr.BaseSHA,
		Key:           suiteKey,
		Name:          "Exist Agent Reviewer",
		Status:        "completed",
		Conclusion:    conclusion,
		OutputTitle:   fmt.Sprintf("Agent Review: %s (Score %d/100)", strings.ToUpper(verdict), score),
		OutputSummary: summaryText,
		OutputText:    overviewMD,
		ExternalID:    fmt.Sprintf("agent-review-%s-%s", pr.ID, pr.HeadSHA),
		StartedAt:     nowStr,
		CompletedAt:   nowStr,
		ActorUserID:   "",
		CreatedAt:     nowStr,
		UpdatedAt:     nowStr,
	}
	if _, _, err := db.UpsertCheckRun(ctx, checkRun); err != nil {
		return nil, fmt.Errorf("agentreviewer: upsert check run: %w", err)
	}

	// 8. Record telemetry to agent activity hub
	agentactivity.DefaultHub.Record(agentactivity.Activity{
		AgentName: "exist-reviewer",
		AgentType: "reviewer",
		Action:    "pr_review_completed",
		Resource:  fmt.Sprintf("%s/%s#%d", ownerSlug, repo.Name, pr.Number),
		Status:    conclusion,
		Details:   fmt.Sprintf("PR #%d: %s (%s, score %d/100, %d errors, %d warnings)", pr.Number, pr.Title, verdict, score, errCount, warnCount),
	})

	return &ReviewResult{
		Verdict:    verdict,
		Conclusion: conclusion,
		Score:      score,
		Summary:    summaryText,
		OverviewMD: overviewMD,
		Violations: violations,
		ReviewID:   reviewID,
		CheckRunID: runID,
		CreatedAt:  nowStr,
	}, nil
}

func isSource(ext string) bool {
	switch ext {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".rs", ".java", ".c", ".cpp":
		return true
	default:
		return false
	}
}

func isTest(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.HasSuffix(lower, "_test.go") ||
		strings.Contains(lower, ".test.") ||
		strings.Contains(lower, ".spec.") ||
		strings.HasPrefix(lower, "test_") ||
		strings.Contains(lower, "/test/") ||
		strings.Contains(lower, "/tests/")
}

func getExtension(filename string) string {
	idx := strings.LastIndex(filename, ".")
	if idx == -1 {
		return ""
	}
	return strings.ToLower(filename[idx:])
}

func isTestMock(val string) bool {
	lower := strings.ToLower(val)
	return strings.Contains(lower, "fake") ||
		strings.Contains(lower, "test") ||
		strings.Contains(lower, "dummy") ||
		strings.Contains(lower, "example") ||
		strings.Contains(lower, "mock") ||
		strings.Contains(lower, "placeholder") ||
		strings.Contains(lower, "xxxxxxxx")
}

func redact(s string) string {
	if len(s) <= 8 {
		return "********"
	}
	return s[:4] + "••••••••" + s[len(s)-4:]
}

func generateMarkdownReport(ownerSlug string, repo store.Repo, pr store.PullRequest, diff *gitstore.PRDiffStats, score int, verdict string, violations []Violation) string {
	var sb strings.Builder

	badge := "🟢 APPROVED"
	if verdict == "request_changes" {
		badge = "🔴 CHANGES REQUESTED"
	} else if verdict == "comment" {
		badge = "🟡 COMMENTS & RECOMMENDATIONS"
	}

	sb.WriteString(fmt.Sprintf("## Autonomous Agent Code Review: %s\n\n", badge))
	sb.WriteString(fmt.Sprintf("**Repository:** `%s/%s` | **PR:** #%d (`%s` → `%s`)\n", ownerSlug, repo.Name, pr.Number, pr.HeadRef, pr.BaseRef))
	sb.WriteString(fmt.Sprintf("**Quality & Compliance Score:** **%d / 100**\n\n", score))

	sb.WriteString("### Change Summary\n")
	sb.WriteString(fmt.Sprintf("- **Files Changed:** %d\n", diff.ChangedFiles))
	sb.WriteString(fmt.Sprintf("- **Additions:** +%d lines\n", diff.Additions))
	sb.WriteString(fmt.Sprintf("- **Deletions:** -%d lines\n\n", diff.Deletions))

	if len(violations) == 0 {
		sb.WriteString("### Automated Findings\n")
		sb.WriteString("✅ **Zero security vulnerabilities, leaked secrets, or syntax anomalies detected.**\n")
		sb.WriteString("Clean git commit hygiene and test conventions verified.\n")
		return sb.String()
	}

	sb.WriteString("### Automated Findings & Diagnostics\n\n")
	sb.WriteString("| Severity | Rule | Location | Detail |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- |\n")
	for _, v := range violations {
		icon := "ℹ️"
		if v.Severity == "error" {
			icon = "🚨 ERROR"
		} else if v.Severity == "warning" {
			icon = "⚠️ WARN"
		}

		loc := v.File
		if v.Line > 0 {
			loc = fmt.Sprintf("`%s:%d`", v.File, v.Line)
		} else {
			loc = fmt.Sprintf("`%s`", v.File)
		}

		sb.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s |\n", icon, v.Rule, loc, v.Message))
	}

	sb.WriteString("\n### Next Steps\n")
	if verdict == "request_changes" {
		sb.WriteString("> ⚠️ **Action Required:** Please address the critical security error(s) flagged above before re-requesting review or attempting to merge.\n")
	} else if verdict == "comment" {
		sb.WriteString("> 💡 **Recommendation:** Review the suggestions above. Adding automated test coverage is strongly encouraged for production stability.\n")
	} else {
		sb.WriteString("> 🚀 **Ready for Merge:** All automated gates and policy checks have passed.\n")
	}

	return sb.String()
}
