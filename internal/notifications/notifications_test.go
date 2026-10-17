package notifications

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
)

func TestRenderEmail_PROpened(t *testing.T) {
	evt := NotificationEvent{
		Type:      EventPROpened,
		RepoOwner: "octocat",
		RepoName:  "hello-world",
		ActorName: "alice",
		Recipients: []string{
			"bob@example.com",
		},
		Payload: map[string]string{
			"title":       "Add exciting new feature",
			"branch":      "feature/exciting",
			"base_branch": "main",
			"url":         "https://exist.io/octocat/hello-world/pull/42",
			"body":        "This PR introduces a brand new feature for exist.io users.",
			"number":      "42",
		},
	}

	subject, htmlBody, textBody, err := RenderEmail(evt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify Subject
	if !strings.Contains(subject, "octocat/hello-world") {
		t.Errorf("subject missing repo: %s", subject)
	}
	if !strings.Contains(subject, "Add exciting new feature") {
		t.Errorf("subject missing title: %s", subject)
	}
	if !strings.Contains(subject, "#42") {
		t.Errorf("subject missing PR number: %s", subject)
	}

	// Verify HTML Body
	expectedHTMLSubstrings := []string{
		"exist",
		".io",
		"max-width: 600px",
		"alice",
		"Add exciting new feature",
		"feature/exciting",
		"https://exist.io/octocat/hello-world/pull/42",
		"This PR introduces a brand new feature",
		"View Pull Request",
		"Unsubscribe",
	}
	for _, substr := range expectedHTMLSubstrings {
		if !strings.Contains(htmlBody, substr) {
			t.Errorf("htmlBody missing expected substring %q", substr)
		}
	}

	// Verify Plain Text Body
	expectedTextSubstrings := []string{
		"exist.io - octocat/hello-world",
		"Pull Request Opened: Add exciting new feature",
		"Author: alice",
		"Branch: feature/exciting -> main",
		"This PR introduces a brand new feature",
		"View Pull Request: https://exist.io/octocat/hello-world/pull/42",
		"Unsubscribe",
	}
	for _, substr := range expectedTextSubstrings {
		if !strings.Contains(textBody, substr) {
			t.Errorf("textBody missing expected substring %q", substr)
		}
	}
}

func TestRenderEmail_PRMerged(t *testing.T) {
	evt := NotificationEvent{
		Type:      EventPRMerged,
		RepoOwner: "octocat",
		RepoName:  "hello-world",
		ActorName: "carol",
		Recipients: []string{
			"devs@example.com",
		},
		Payload: map[string]string{
			"title":        "Fix critical edge-case crash",
			"merger":       "carol",
			"merge_method": "squash",
			"url":          "https://exist.io/octocat/hello-world/pull/55",
			"number":       "55",
		},
	}

	subject, htmlBody, textBody, err := RenderEmail(evt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(subject, "merged") {
		t.Errorf("subject missing 'merged': %s", subject)
	}
	if !strings.Contains(subject, "Fix critical edge-case crash") {
		t.Errorf("subject missing title: %s", subject)
	}

	// HTML assertions
	if !strings.Contains(htmlBody, "carol") {
		t.Errorf("htmlBody missing merger name carol")
	}
	if !strings.Contains(htmlBody, "squash") {
		t.Errorf("htmlBody missing merge method squash")
	}
	if !strings.Contains(htmlBody, "View Merged Pull Request") {
		t.Errorf("htmlBody missing View Merged Pull Request button")
	}
	if !strings.Contains(htmlBody, "https://exist.io/octocat/hello-world/pull/55") {
		t.Errorf("htmlBody missing URL")
	}

	// Plain text assertions
	if !strings.Contains(textBody, "Merged by: carol") {
		t.Errorf("textBody missing 'Merged by: carol'")
	}
	if !strings.Contains(textBody, "Method: squash") {
		t.Errorf("textBody missing 'Method: squash'")
	}
}

func TestRenderEmail_PRReviewed(t *testing.T) {
	evt := NotificationEvent{
		Type:      EventPRReviewed,
		RepoOwner: "octocat",
		RepoName:  "hello-world",
		ActorName: "reviewer1",
		Recipients: []string{
			"author@example.com",
		},
		Payload: map[string]string{
			"title":    "Refactor storage engine",
			"reviewer": "reviewer1",
			"verdict":  "approved",
			"body":     "Looks solid and passes all benchmarks!",
			"url":      "https://exist.io/octocat/hello-world/pull/60#review-12",
			"number":   "60",
		},
	}

	subject, htmlBody, textBody, err := RenderEmail(evt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(subject, "Review") && !strings.Contains(subject, "Refactor storage engine") {
		t.Errorf("unexpected subject: %s", subject)
	}

	// HTML assertions
	if !strings.Contains(htmlBody, "reviewer1") {
		t.Errorf("htmlBody missing reviewer name")
	}
	if !strings.Contains(htmlBody, "approved") {
		t.Errorf("htmlBody missing verdict approved")
	}
	if !strings.Contains(htmlBody, "Looks solid and passes all benchmarks!") {
		t.Errorf("htmlBody missing review comment body")
	}
	if !strings.Contains(htmlBody, "View Review") {
		t.Errorf("htmlBody missing View Review button")
	}

	// Plain text assertions
	if !strings.Contains(textBody, "Reviewer: reviewer1") {
		t.Errorf("textBody missing 'Reviewer: reviewer1'")
	}
	if !strings.Contains(textBody, "Verdict: approved") {
		t.Errorf("textBody missing 'Verdict: approved'")
	}
	if !strings.Contains(textBody, "Looks solid and passes all benchmarks!") {
		t.Errorf("textBody missing comment body")
	}
}

func TestRenderEmail_PRClosed(t *testing.T) {
	evt := NotificationEvent{
		Type:      EventPRClosed,
		RepoOwner: "octocat",
		RepoName:  "hello-world",
		ActorName: "maintainer",
		Payload: map[string]string{
			"title":  "Superseded cleanup branch",
			"closer": "maintainer",
			"url":    "https://exist.io/octocat/hello-world/pull/70",
			"number": "70",
		},
	}

	subject, htmlBody, textBody, err := RenderEmail(evt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(subject, "closed") {
		t.Errorf("subject missing closed: %s", subject)
	}
	if !strings.Contains(htmlBody, "maintainer") || !strings.Contains(htmlBody, "closed this pull request without merging") {
		t.Errorf("htmlBody missing closer info")
	}
	if !strings.Contains(textBody, "Closed by: maintainer") {
		t.Errorf("textBody missing closer info")
	}
}

func TestRenderEmail_IssueAssigned(t *testing.T) {
	evt := NotificationEvent{
		Type:      EventIssueAssigned,
		RepoOwner: "octocat",
		RepoName:  "hello-world",
		ActorName: "lead",
		Payload: map[string]string{
			"title":    "Fix memory leak in websocket loop",
			"assignee": "shiva",
			"labels":   "bug, performance",
			"url":      "https://exist.io/octocat/hello-world/issues/101",
			"number":   "101",
		},
	}

	subject, htmlBody, textBody, err := RenderEmail(evt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(subject, "assigned") || !strings.Contains(subject, "Fix memory leak") {
		t.Errorf("unexpected subject: %s", subject)
	}
	if !strings.Contains(htmlBody, "shiva") || !strings.Contains(htmlBody, "was assigned to this issue") {
		t.Errorf("htmlBody missing assignee")
	}
	if !strings.Contains(htmlBody, "bug, performance") {
		t.Errorf("htmlBody missing labels")
	}
	if !strings.Contains(textBody, "Assignee: shiva") || !strings.Contains(textBody, "Labels: bug, performance") {
		t.Errorf("textBody missing assignee/labels")
	}
}

func TestRenderEmail_UnsupportedEvent(t *testing.T) {
	evt := NotificationEvent{
		Type: "unknown.event",
	}

	_, _, _, err := RenderEmail(evt)
	if err == nil {
		t.Fatalf("expected error for unsupported event type, got nil")
	}
}

func TestMailer_NoopWhenUnconfigured(t *testing.T) {
	// Empty Host should immediately return nil
	mailer := NewMailer(SMTPConfig{
		Host: "",
	})

	err := mailer.Send([]string{"user@example.com"}, "Hello", "<p>World</p>", "World")
	if err != nil {
		t.Fatalf("expected nil when mailer host is empty, got: %v", err)
	}

	// Empty recipients should also return nil
	configuredMailer := NewMailer(SMTPConfig{
		Host: "smtp.example.com",
		Port: 587,
	})
	err = configuredMailer.Send([]string{}, "Hello", "<p>World</p>", "World")
	if err != nil {
		t.Fatalf("expected nil when recipients slice is empty, got: %v", err)
	}
}

func TestWorker_StartStop(t *testing.T) {
	mailer := NewMailer(SMTPConfig{})
	worker := NewWorker(mailer, 10)

	worker.Start()
	// Idempotent start
	worker.Start()

	worker.Stop()
	// Idempotent stop
	worker.Stop()
}

func TestWorker_Enqueue(t *testing.T) {
	mailer := NewMailer(SMTPConfig{})
	worker := NewWorker(mailer, 20)

	worker.Start()

	events := []NotificationEvent{
		{
			Type:       EventPROpened,
			RepoOwner:  "acme",
			RepoName:   "rocket",
			ActorName:  "wile",
			Recipients: []string{"coyote@example.com"},
			Payload: map[string]string{
				"title":  "Deploy explosive package",
				"branch": "feature/bomb",
				"url":    "https://exist.io/acme/rocket/pull/1",
			},
		},
		{
			Type:       EventPRReviewed,
			RepoOwner:  "acme",
			RepoName:   "rocket",
			ActorName:  "roadrunner",
			Recipients: []string{"wile@example.com"},
			Payload: map[string]string{
				"title":    "Deploy explosive package",
				"reviewer": "roadrunner",
				"verdict":  "changes_requested",
				"body":     "Beep beep!",
				"url":      "https://exist.io/acme/rocket/pull/1#review-1",
			},
		},
		{
			Type:       EventPRMerged,
			RepoOwner:  "acme",
			RepoName:   "rocket",
			ActorName:  "admin",
			Recipients: []string{"wile@example.com"},
			Payload: map[string]string{
				"title":        "Deploy explosive package",
				"merger":       "admin",
				"merge_method": "rebase",
				"url":          "https://exist.io/acme/rocket/pull/1",
			},
		},
	}

	for _, evt := range events {
		worker.Enqueue(evt)
	}

	// Graceful stop should drain all events without issue
	worker.Stop()

	// Enqueue after stop should be safely dropped without panic
	worker.Enqueue(events[0])
}

func TestMailer_BuildMessage(t *testing.T) {
	from := "Exist Notifications <notifications@exist.io>"
	to := []string{"Alice <alice@example.com>", "bob@example.com"}
	subject := "Test Subject Line"
	htmlBody := "<h1>Hello</h1><p>This is a test HTML email.</p>"
	textBody := "Hello\nThis is a test plain text email."
	host := "mail.example.com"

	msgBytes := BuildMessage(from, to, subject, htmlBody, textBody, host)
	msg := string(msgBytes)

	// Check headers
	if !strings.Contains(msg, "From: Exist Notifications <notifications@exist.io>\r\n") {
		t.Errorf("missing or invalid From header:\n%s", msg)
	}
	if !strings.Contains(msg, "To: Alice <alice@example.com>, bob@example.com\r\n") {
		t.Errorf("missing or invalid To header:\n%s", msg)
	}
	if !strings.Contains(msg, "Subject: Test Subject Line\r\n") {
		t.Errorf("missing or invalid Subject header:\n%s", msg)
	}
	if !strings.Contains(msg, "Date: ") {
		t.Errorf("missing Date header:\n%s", msg)
	}
	if !strings.Contains(msg, "Message-ID: <msg_") || !strings.Contains(msg, "@mail.example.com>\r\n") {
		t.Errorf("missing or invalid Message-ID header:\n%s", msg)
	}
	if !strings.Contains(msg, "MIME-Version: 1.0\r\n") {
		t.Errorf("missing MIME-Version header:\n%s", msg)
	}
	if !strings.Contains(msg, "Content-Type: multipart/alternative; boundary=\"----=_ExistBoundary_") {
		t.Errorf("missing multipart/alternative Content-Type:\n%s", msg)
	}

	// Check parts
	if !strings.Contains(msg, "Content-Type: text/plain; charset=UTF-8\r\n") {
		t.Errorf("missing text/plain part:\n%s", msg)
	}
	if !strings.Contains(msg, "Content-Type: text/html; charset=UTF-8\r\n") {
		t.Errorf("missing text/html part:\n%s", msg)
	}
	if !strings.Contains(msg, "This is a test plain text email.") {
		t.Errorf("missing plain text body content:\n%s", msg)
	}
	if !strings.Contains(msg, "This is a test HTML email.") {
		t.Errorf("missing HTML body content:\n%s", msg)
	}
}

func TestMailer_MockSMTPServer(t *testing.T) {
	// Spin up a local mock SMTP server to verify net/smtp protocol interaction
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	serverHost, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to parse listener addr: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	var receivedData strings.Builder
	var serverWg sync.WaitGroup
	serverWg.Add(1)

	go func() {
		defer serverWg.Done()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)

		// Send initial greeting
		fmt.Fprintf(writer, "220 %s ESMTP Mock\r\n", serverHost)
		writer.Flush()

		inData := false
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			trimmed := strings.TrimRight(line, "\r\n")

			if inData {
				if trimmed == "." {
					inData = false
					fmt.Fprintf(writer, "250 2.0.0 OK: queued\r\n")
					writer.Flush()
					continue
				}
				receivedData.WriteString(line)
				continue
			}

			upper := strings.ToUpper(trimmed)
			if strings.HasPrefix(upper, "EHLO") || strings.HasPrefix(upper, "HELO") {
				// Don't advertise STARTTLS here to test plain handshake
				fmt.Fprintf(writer, "250-%s\r\n250 8BITMIME\r\n", serverHost)
				writer.Flush()
			} else if strings.HasPrefix(upper, "MAIL FROM:") {
				fmt.Fprintf(writer, "250 2.1.0 OK\r\n")
				writer.Flush()
			} else if strings.HasPrefix(upper, "RCPT TO:") {
				fmt.Fprintf(writer, "250 2.1.5 OK\r\n")
				writer.Flush()
			} else if upper == "DATA" {
				inData = true
				fmt.Fprintf(writer, "354 Start mail input; end with <CRLF>.<CRLF>\r\n")
				writer.Flush()
			} else if upper == "QUIT" {
				fmt.Fprintf(writer, "221 2.0.0 Bye\r\n")
				writer.Flush()
				break
			} else {
				fmt.Fprintf(writer, "500 Unrecognized command\r\n")
				writer.Flush()
			}
		}
	}()

	mailer := NewMailer(SMTPConfig{
		Host: serverHost,
		Port: port,
		From: "forge@exist.io",
		TLS:  false,
	})

	err = mailer.Send(
		[]string{"reviewer@example.com"},
		"PR #123 merged successfully",
		"<p>The pull request has been merged.</p>",
		"The pull request has been merged.",
	)
	if err != nil {
		t.Fatalf("mailer.Send failed: %v", err)
	}

	serverWg.Wait()

	dataStr := receivedData.String()
	if !strings.Contains(dataStr, "Subject: PR #123 merged successfully") {
		t.Errorf("mock server did not receive subject:\n%s", dataStr)
	}
	if !strings.Contains(dataStr, "From: forge@exist.io") {
		t.Errorf("mock server did not receive From header:\n%s", dataStr)
	}
	if !strings.Contains(dataStr, "To: reviewer@example.com") {
		t.Errorf("mock server did not receive To header:\n%s", dataStr)
	}
	if !strings.Contains(dataStr, "The pull request has been merged.") {
		t.Errorf("mock server did not receive expected body:\n%s", dataStr)
	}
}

func TestEvent_Repo(t *testing.T) {
	e1 := NotificationEvent{RepoOwner: "alice", RepoName: "forge"}
	if e1.Repo() != "alice/forge" {
		t.Errorf("expected alice/forge, got %s", e1.Repo())
	}

	e2 := NotificationEvent{RepoName: "standalone"}
	if e2.Repo() != "standalone" {
		t.Errorf("expected standalone, got %s", e2.Repo())
	}

	e3 := NotificationEvent{}
	if e3.Repo() != "" {
		t.Errorf("expected empty string, got %s", e3.Repo())
	}
}

func TestMailer_ExtractEmail(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"alice@example.com", "alice@example.com"},
		{"Alice Smith <alice@example.com>", "alice@example.com"},
		{"  <bob@example.com>  ", "bob@example.com"},
		{"plain", "plain"},
	}
	for _, c := range cases {
		got := extractEmail(c.input)
		if got != c.expected {
			t.Errorf("extractEmail(%q) = %q; expected %q", c.input, got, c.expected)
		}
	}
}

func TestWorker_BufferDefaultsAndDraining(t *testing.T) {
	// Worker with 0 bufferSize should default to 100
	w := NewWorker(nil, 0)
	if cap(w.eventCh) != 100 {
		t.Errorf("expected default buffer capacity 100, got %d", cap(w.eventCh))
	}

	// Worker should safely handle events with no recipients
	w.processEvent(NotificationEvent{Type: EventPROpened})

	// Worker should safely handle events with invalid type
	w.processEvent(NotificationEvent{
		Type:       "invalid.type",
		Recipients: []string{"test@example.com"},
	})

	w.Start()
	w.Stop()
}

func TestWorker_FullDeliveryWithMockSMTP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	serverHost, portStr, _ := net.SplitHostPort(ln.Addr().String())
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	var serverWg sync.WaitGroup
	serverWg.Add(1)

	var receivedData strings.Builder
	go func() {
		defer serverWg.Done()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)

		fmt.Fprintf(writer, "220 %s ESMTP Mock\r\n", serverHost)
		writer.Flush()

		inData := false
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			trimmed := strings.TrimRight(line, "\r\n")

			if inData {
				if trimmed == "." {
					inData = false
					fmt.Fprintf(writer, "250 OK queued\r\n")
					writer.Flush()
					continue
				}
				receivedData.WriteString(line)
				continue
			}

			upper := strings.ToUpper(trimmed)
			if strings.HasPrefix(upper, "EHLO") || strings.HasPrefix(upper, "HELO") {
				fmt.Fprintf(writer, "250-%s\r\n250 8BITMIME\r\n", serverHost)
				writer.Flush()
			} else if strings.HasPrefix(upper, "MAIL FROM:") {
				fmt.Fprintf(writer, "250 OK\r\n")
				writer.Flush()
			} else if strings.HasPrefix(upper, "RCPT TO:") {
				fmt.Fprintf(writer, "250 OK\r\n")
				writer.Flush()
			} else if upper == "DATA" {
				inData = true
				fmt.Fprintf(writer, "354 Go ahead\r\n")
				writer.Flush()
			} else if upper == "QUIT" {
				fmt.Fprintf(writer, "221 Bye\r\n")
				writer.Flush()
				break
			} else {
				fmt.Fprintf(writer, "500 Error\r\n")
				writer.Flush()
			}
		}
	}()

	mailer := NewMailer(SMTPConfig{
		Host: serverHost,
		Port: port,
		From: "forge@exist.io",
	})
	worker := NewWorker(mailer, 10)
	worker.Start()

	worker.Enqueue(NotificationEvent{
		Type:       EventPROpened,
		RepoOwner:  "exist",
		RepoName:   "exist.io",
		ActorName:  "dev",
		Recipients: []string{"team@example.com"},
		Payload: map[string]string{
			"title":  "Worker live delivery test",
			"branch": "feat/notifications",
			"url":    "https://exist.io/exist/exist.io/pull/99",
		},
	})

	worker.Stop()
	serverWg.Wait()

	if !strings.Contains(receivedData.String(), "Worker live delivery test") {
		t.Errorf("mock server did not receive delivered email:\n%s", receivedData.String())
	}
}

func TestMailer_PlainAuthWithMockSMTP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	serverHost, portStr, _ := net.SplitHostPort(ln.Addr().String())
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	authSeen := false
	var serverWg sync.WaitGroup
	serverWg.Add(1)

	go func() {
		defer serverWg.Done()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)

		fmt.Fprintf(writer, "220 %s ESMTP AuthMock\r\n", serverHost)
		writer.Flush()

		inData := false
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			trimmed := strings.TrimRight(line, "\r\n")

			if inData {
				if trimmed == "." {
					inData = false
					fmt.Fprintf(writer, "250 OK\r\n")
					writer.Flush()
					continue
				}
				continue
			}

			upper := strings.ToUpper(trimmed)
			if strings.HasPrefix(upper, "EHLO") || strings.HasPrefix(upper, "HELO") {
				// Advertise PLAIN AUTH
				fmt.Fprintf(writer, "250-%s\r\n250-AUTH PLAIN\r\n250 8BITMIME\r\n", serverHost)
				writer.Flush()
			} else if strings.HasPrefix(upper, "AUTH PLAIN") {
				authSeen = true
				fmt.Fprintf(writer, "235 2.7.0 Authentication successful\r\n")
				writer.Flush()
			} else if strings.HasPrefix(upper, "MAIL FROM:") {
				fmt.Fprintf(writer, "250 OK\r\n")
				writer.Flush()
			} else if strings.HasPrefix(upper, "RCPT TO:") {
				fmt.Fprintf(writer, "250 OK\r\n")
				writer.Flush()
			} else if upper == "DATA" {
				inData = true
				fmt.Fprintf(writer, "354 End data with <CRLF>.<CRLF>\r\n")
				writer.Flush()
			} else if upper == "QUIT" {
				fmt.Fprintf(writer, "221 Bye\r\n")
				writer.Flush()
				break
			} else {
				fmt.Fprintf(writer, "500 Command unrecognized\r\n")
				writer.Flush()
			}
		}
	}()

	mailer := NewMailer(SMTPConfig{
		Host:     serverHost,
		Port:     port,
		Username: "forgeuser",
		Password: "supersecretpassword",
		From:     "forge@exist.io",
	})

	err = mailer.Send(
		[]string{"user@example.com"},
		"Auth Test",
		"<p>Body</p>",
		"Body",
	)
	if err != nil {
		t.Fatalf("mailer.Send with auth failed: %v", err)
	}

	serverWg.Wait()

	if !authSeen {
		t.Errorf("expected AUTH PLAIN command to be sent to mock SMTP server")
	}
}
