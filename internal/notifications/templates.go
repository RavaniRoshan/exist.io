package notifications

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

type emailViewData struct {
	Event            EventType
	Subject          string
	Repo             string
	RepoOwner        string
	RepoName         string
	ActorName        string
	Author           string
	Reviewer         string
	Merger           string
	Closer           string
	Assignee         string
	Verdict          string
	VerdictBgColor   string
	VerdictTextColor string
	MergeMethod      string
	Title            string
	Number           string
	Branch           string
	BaseBranch       string
	Body             string
	Labels           string
	URL              string
	UnsubscribeURL   string
}

const emailHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Subject}}</title>
</head>
<body style="margin: 0; padding: 24px 0; background-color: #0d1117; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif; color: #c9d1d9; line-height: 1.5;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background-color: #0d1117; width: 100%;">
    <tr>
      <td align="center" style="padding: 0 16px;">
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="max-width: 600px; width: 100%; background-color: #161b22; border: 1px solid #30363d; border-radius: 6px; overflow: hidden; border-collapse: separate;">
          <!-- Dark Header with exist.io Branding -->
          <tr>
            <td style="background-color: #0d1117; padding: 16px 24px; border-bottom: 1px solid #30363d;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td>
                    <span style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; font-size: 18px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px;">exist<span style="color: #38bdf8;">.io</span></span>
                  </td>
                  {{if .Repo}}
                  <td align="right">
                    <span style="font-family: ui-monospace, SFMono-Regular, monospace; font-size: 12px; color: #8b949e;">{{.Repo}}</span>
                  </td>
                  {{end}}
                </tr>
              </table>
            </td>
          </tr>
          <!-- Body Content Area -->
          <tr>
            <td style="padding: 24px; font-size: 14px; color: #c9d1d9;">
              {{if eq .Event "pr.opened"}}
                <div style="font-size: 13px; color: #8b949e; margin-bottom: 8px;">
                  <strong style="color: #f0f6fc;">{{.Author}}</strong> opened a pull request
                  {{if .Branch}}into <code style="background-color: #21262d; padding: 2px 6px; border-radius: 4px; font-size: 12px; color: #58a6ff; font-family: ui-monospace, monospace;">{{.BaseBranch}}</code> from <code style="background-color: #21262d; padding: 2px 6px; border-radius: 4px; font-size: 12px; color: #58a6ff; font-family: ui-monospace, monospace;">{{.Branch}}</code>{{end}}
                </div>
                <h1 style="margin: 0 0 16px 0; font-size: 20px; font-weight: 600; color: #f0f6fc; line-height: 1.3;">{{.Title}}</h1>
                {{if .Body}}
                <div style="padding: 16px; background-color: #0d1117; border: 1px solid #30363d; border-radius: 6px; margin-bottom: 20px; white-space: pre-wrap; font-size: 13px; color: #c9d1d9; line-height: 1.6;">{{.Body}}</div>
                {{end}}
                <div style="margin-top: 20px;">
                  <a href="{{.URL}}" style="display: inline-block; background-color: #238636; color: #ffffff; text-decoration: none; padding: 10px 18px; border-radius: 6px; font-weight: 600; font-size: 14px;">View Pull Request</a>
                </div>

              {{else if eq .Event "pr.reviewed"}}
                <div style="font-size: 13px; color: #8b949e; margin-bottom: 8px;">
                  <strong style="color: #f0f6fc;">{{.Reviewer}}</strong> reviewed this pull request:
                  <span style="display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 11px; font-weight: 600; text-transform: uppercase; background-color: {{.VerdictBgColor}}; color: {{.VerdictTextColor}};">{{.Verdict}}</span>
                </div>
                <h1 style="margin: 0 0 16px 0; font-size: 20px; font-weight: 600; color: #f0f6fc; line-height: 1.3;">{{.Title}}</h1>
                {{if .Body}}
                <div style="padding: 16px; background-color: #0d1117; border: 1px solid #30363d; border-radius: 6px; margin-bottom: 20px; white-space: pre-wrap; font-size: 13px; color: #c9d1d9; line-height: 1.6;">{{.Body}}</div>
                {{end}}
                <div style="margin-top: 20px;">
                  <a href="{{.URL}}" style="display: inline-block; background-color: #238636; color: #ffffff; text-decoration: none; padding: 10px 18px; border-radius: 6px; font-weight: 600; font-size: 14px;">View Review</a>
                </div>

              {{else if eq .Event "pr.merged"}}
                <div style="font-size: 13px; color: #8b949e; margin-bottom: 8px;">
                  <strong style="color: #f0f6fc;">{{.Merger}}</strong> merged this pull request
                  {{if .MergeMethod}}via <code style="background-color: #21262d; padding: 2px 6px; border-radius: 4px; font-size: 12px; color: #a371f7; font-family: ui-monospace, monospace;">{{.MergeMethod}}</code>{{end}}
                </div>
                <h1 style="margin: 0 0 16px 0; font-size: 20px; font-weight: 600; color: #f0f6fc; line-height: 1.3;">{{.Title}}</h1>
                {{if .Body}}
                <div style="padding: 16px; background-color: #0d1117; border: 1px solid #30363d; border-radius: 6px; margin-bottom: 20px; white-space: pre-wrap; font-size: 13px; color: #c9d1d9; line-height: 1.6;">{{.Body}}</div>
                {{end}}
                <div style="margin-top: 20px;">
                  <a href="{{.URL}}" style="display: inline-block; background-color: #8957e5; color: #ffffff; text-decoration: none; padding: 10px 18px; border-radius: 6px; font-weight: 600; font-size: 14px;">View Merged Pull Request</a>
                </div>

              {{else if eq .Event "pr.closed"}}
                <div style="font-size: 13px; color: #8b949e; margin-bottom: 8px;">
                  <strong style="color: #f0f6fc;">{{.Closer}}</strong> closed this pull request without merging
                </div>
                <h1 style="margin: 0 0 16px 0; font-size: 20px; font-weight: 600; color: #f0f6fc; line-height: 1.3;">{{.Title}}</h1>
                {{if .Body}}
                <div style="padding: 16px; background-color: #0d1117; border: 1px solid #30363d; border-radius: 6px; margin-bottom: 20px; white-space: pre-wrap; font-size: 13px; color: #c9d1d9; line-height: 1.6;">{{.Body}}</div>
                {{end}}
                <div style="margin-top: 20px;">
                  <a href="{{.URL}}" style="display: inline-block; background-color: #30363d; color: #c9d1d9; text-decoration: none; padding: 10px 18px; border-radius: 6px; font-weight: 600; font-size: 14px;">View Pull Request</a>
                </div>

              {{else if eq .Event "issue.assigned"}}
                <div style="font-size: 13px; color: #8b949e; margin-bottom: 8px;">
                  <strong style="color: #f0f6fc;">{{.Assignee}}</strong> was assigned to this issue
                  {{if .ActorName}}by <strong style="color: #f0f6fc;">{{.ActorName}}</strong>{{end}}
                </div>
                <h1 style="margin: 0 0 16px 0; font-size: 20px; font-weight: 600; color: #f0f6fc; line-height: 1.3;">{{.Title}}</h1>
                {{if .Labels}}
                <div style="margin-bottom: 16px; font-size: 13px; color: #8b949e;">
                  Labels: <span style="display: inline-block; padding: 2px 8px; background-color: #21262d; border-radius: 12px; font-size: 12px; color: #58a6ff;">{{.Labels}}</span>
                </div>
                {{end}}
                {{if .Body}}
                <div style="padding: 16px; background-color: #0d1117; border: 1px solid #30363d; border-radius: 6px; margin-bottom: 20px; white-space: pre-wrap; font-size: 13px; color: #c9d1d9; line-height: 1.6;">{{.Body}}</div>
                {{end}}
                <div style="margin-top: 20px;">
                  <a href="{{.URL}}" style="display: inline-block; background-color: #238636; color: #ffffff; text-decoration: none; padding: 10px 18px; border-radius: 6px; font-weight: 600; font-size: 14px;">View Issue</a>
                </div>
              {{end}}
            </td>
          </tr>
          <!-- Footer with Unsubscribe Placeholder -->
          <tr>
            <td style="padding: 16px 24px; background-color: #0d1117; border-top: 1px solid #30363d; text-align: center; font-size: 12px; color: #8b949e;">
              {{if .Repo}}<p style="margin: 0 0 6px 0;">You are receiving this email because you are subscribed to notifications for <strong style="color: #c9d1d9;">{{.Repo}}</strong>.</p>{{end}}
              <p style="margin: 0;"><a href="{{.UnsubscribeURL}}" style="color: #58a6ff; text-decoration: underline;">Unsubscribe</a> or manage notification settings</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>
`

var compiledHTMLTemplate = template.Must(template.New("notificationEmail").Parse(emailHTMLTemplate))

// RenderEmail renders HTML and plaintext bodies alongside the subject for a notification event.
func RenderEmail(evt NotificationEvent) (subject, htmlBody, textBody string, err error) {
	switch evt.Type {
	case EventPROpened, EventPRReviewed, EventPRMerged, EventPRClosed, EventIssueAssigned:
		// Supported types
	default:
		return "", "", "", fmt.Errorf("unsupported notification event type: %s", evt.Type)
	}

	data := buildViewData(evt)

	// Determine final subject
	if evt.Subject != "" {
		subject = evt.Subject
	} else {
		repoPrefix := ""
		if data.Repo != "" {
			repoPrefix = "[" + data.Repo + "] "
		}
		numPrefix := ""
		if data.Number != "" {
			numPrefix = "#" + data.Number + " "
		}

		switch evt.Type {
		case EventPROpened:
			subject = fmt.Sprintf("%sPull request %sopened: %s", repoPrefix, numPrefix, data.Title)
		case EventPRReviewed:
			subject = fmt.Sprintf("%sReview submitted on PR %s: %s", repoPrefix, numPrefix, data.Title)
		case EventPRMerged:
			subject = fmt.Sprintf("%sPull request %smerged: %s", repoPrefix, numPrefix, data.Title)
		case EventPRClosed:
			subject = fmt.Sprintf("%sPull request %sclosed: %s", repoPrefix, numPrefix, data.Title)
		case EventIssueAssigned:
			subject = fmt.Sprintf("%sIssue %sassigned: %s", repoPrefix, numPrefix, data.Title)
		}
	}
	data.Subject = subject

	// Render HTML
	var htmlBuf bytes.Buffer
	if err := compiledHTMLTemplate.Execute(&htmlBuf, data); err != nil {
		return "", "", "", fmt.Errorf("render html template: %w", err)
	}
	htmlBody = htmlBuf.String()

	// Render plain text
	textBody = renderText(data)

	return subject, htmlBody, textBody, nil
}

func buildViewData(evt NotificationEvent) emailViewData {
	p := evt.Payload
	if p == nil {
		p = make(map[string]string)
	}

	repo := evt.Repo()

	title := p["title"]
	if title == "" {
		title = p["issue_title"]
	}
	if title == "" {
		title = "Untitled"
	}

	url := p["url"]
	if url == "" {
		url = p["pr_url"]
	}
	if url == "" {
		url = p["issue_url"]
	}
	if url == "" {
		url = "https://exist.io"
	}

	unsub := p["unsubscribe_url"]
	if unsub == "" {
		unsub = "https://exist.io/settings/notifications"
	}

	author := p["author"]
	if author == "" {
		author = evt.ActorName
	}

	reviewer := p["reviewer"]
	if reviewer == "" {
		reviewer = evt.ActorName
	}

	merger := p["merger"]
	if merger == "" {
		merger = evt.ActorName
	}

	closer := p["closer"]
	if closer == "" {
		closer = evt.ActorName
	}

	assignee := p["assignee"]
	if assignee == "" {
		assignee = evt.ActorName
	}

	verdict := p["verdict"]
	if verdict == "" {
		verdict = p["state"]
	}
	if verdict == "" {
		verdict = "commented"
	}

	verdictBgColor := "#1f6feb"
	verdictTextColor := "#ffffff"
	if strings.EqualFold(verdict, "approved") {
		verdictBgColor = "#238636"
	} else if strings.EqualFold(verdict, "changes_requested") {
		verdictBgColor = "#da3633"
	}

	mergeMethod := p["merge_method"]
	if mergeMethod == "" {
		mergeMethod = "merge commit"
	}

	body := p["body"]
	if body == "" {
		body = p["comment"]
	}
	if body == "" {
		body = p["description"]
	}

	branch := p["branch"]
	if branch == "" {
		branch = p["head_branch"]
	}

	baseBranch := p["base_branch"]
	if baseBranch == "" {
		baseBranch = "main"
	}

	number := p["number"]
	if number == "" {
		number = p["pr_number"]
	}
	if number == "" {
		number = p["issue_number"]
	}

	labels := p["labels"]

	return emailViewData{
		Event:            evt.Type,
		Subject:          evt.Subject,
		Repo:             repo,
		RepoOwner:        evt.RepoOwner,
		RepoName:         evt.RepoName,
		ActorName:        evt.ActorName,
		Author:           author,
		Reviewer:         reviewer,
		Merger:           merger,
		Closer:           closer,
		Assignee:         assignee,
		Verdict:          verdict,
		VerdictBgColor:   verdictBgColor,
		VerdictTextColor: verdictTextColor,
		MergeMethod:      mergeMethod,
		Title:            title,
		Number:           number,
		Branch:           branch,
		BaseBranch:       baseBranch,
		Body:             body,
		Labels:           labels,
		URL:              url,
		UnsubscribeURL:   unsub,
	}
}

func renderText(data emailViewData) string {
	var b strings.Builder
	b.WriteString("exist.io")
	if data.Repo != "" {
		b.WriteString(" - ")
		b.WriteString(data.Repo)
	}
	b.WriteString("\n\n")

	switch data.Event {
	case EventPROpened:
		b.WriteString(fmt.Sprintf("Pull Request Opened: %s\n", data.Title))
		b.WriteString(fmt.Sprintf("Author: %s\n", data.Author))
		if data.Branch != "" {
			b.WriteString(fmt.Sprintf("Branch: %s -> %s\n", data.Branch, data.BaseBranch))
		}
		if data.Body != "" {
			b.WriteString("\n")
			b.WriteString(data.Body)
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("\nView Pull Request: %s\n", data.URL))

	case EventPRReviewed:
		b.WriteString(fmt.Sprintf("Pull Request Reviewed: %s\n", data.Title))
		b.WriteString(fmt.Sprintf("Reviewer: %s\n", data.Reviewer))
		b.WriteString(fmt.Sprintf("Verdict: %s\n", data.Verdict))
		if data.Body != "" {
			b.WriteString("\n")
			b.WriteString(data.Body)
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("\nView Review: %s\n", data.URL))

	case EventPRMerged:
		b.WriteString(fmt.Sprintf("Pull Request Merged: %s\n", data.Title))
		b.WriteString(fmt.Sprintf("Merged by: %s\n", data.Merger))
		b.WriteString(fmt.Sprintf("Method: %s\n", data.MergeMethod))
		b.WriteString(fmt.Sprintf("\nView Merged Pull Request: %s\n", data.URL))

	case EventPRClosed:
		b.WriteString(fmt.Sprintf("Pull Request Closed: %s\n", data.Title))
		b.WriteString(fmt.Sprintf("Closed by: %s\n", data.Closer))
		b.WriteString(fmt.Sprintf("\nView Pull Request: %s\n", data.URL))

	case EventIssueAssigned:
		b.WriteString(fmt.Sprintf("Issue Assigned: %s\n", data.Title))
		b.WriteString(fmt.Sprintf("Assignee: %s\n", data.Assignee))
		if data.Labels != "" {
			b.WriteString(fmt.Sprintf("Labels: %s\n", data.Labels))
		}
		if data.Body != "" {
			b.WriteString("\n")
			b.WriteString(data.Body)
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("\nView Issue: %s\n", data.URL))
	}

	b.WriteString("\n---\n")
	if data.Repo != "" {
		b.WriteString(fmt.Sprintf("You are receiving this because you are subscribed to %s.\n", data.Repo))
	}
	b.WriteString(fmt.Sprintf("Unsubscribe: %s\n", data.UnsubscribeURL))

	return b.String()
}
