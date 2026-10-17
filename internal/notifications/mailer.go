package notifications

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"exist.io/pkg/idgen"
)

// SMTPConfig holds SMTP server connection settings.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	TLS      bool
}

// Mailer sends emails via SMTP.
type Mailer struct {
	cfg SMTPConfig
}

// NewMailer creates a new SMTP mailer.
func NewMailer(cfg SMTPConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

// Send dispatches an email to the given recipients.
// If config has an empty Host or no recipients are specified, Send is a no-op returning nil.
func (m *Mailer) Send(to []string, subject, htmlBody, textBody string) error {
	if m.cfg.Host == "" || len(to) == 0 {
		return nil
	}

	hostName := m.cfg.Host
	if h, _, err := net.SplitHostPort(hostName); err == nil {
		hostName = h
	}

	port := m.cfg.Port
	if port <= 0 {
		if m.cfg.TLS {
			port = 465
		} else {
			port = 587
		}
	}

	addr := net.JoinHostPort(hostName, strconv.Itoa(port))
	tlsConfig := &tls.Config{
		ServerName: hostName,
	}

	var client *smtp.Client

	// Support direct TLS on port 465 or when direct TLS is explicitly configured on non-standard ports
	if port == 465 || (m.cfg.TLS && port != 587 && port != 25) {
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("smtp direct tls dial %s: %w", addr, err)
		}
		c, err := smtp.NewClient(conn, hostName)
		if err != nil {
			_ = conn.Close()
			return fmt.Errorf("smtp client init: %w", err)
		}
		client = c
	} else {
		// Plain TCP connection with opportunistic or enforced STARTTLS
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return fmt.Errorf("smtp dial %s: %w", addr, err)
		}
		c, err := smtp.NewClient(conn, hostName)
		if err != nil {
			_ = conn.Close()
			return fmt.Errorf("smtp client init: %w", err)
		}
		client = c

		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				_ = client.Close()
				return fmt.Errorf("smtp starttls: %w", err)
			}
		} else if m.cfg.TLS {
			_ = client.Close()
			return fmt.Errorf("smtp: server does not support STARTTLS")
		}
	}
	defer client.Close()

	if m.cfg.Username != "" {
		auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, hostName)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	from := m.cfg.From
	if from == "" {
		from = fmt.Sprintf("noreply@%s", hostName)
	}

	envelopeFrom := extractEmail(from)
	if err := client.Mail(envelopeFrom); err != nil {
		return fmt.Errorf("smtp mail from %s: %w", envelopeFrom, err)
	}

	for _, recipient := range to {
		envelopeTo := extractEmail(recipient)
		if err := client.Rcpt(envelopeTo); err != nil {
			return fmt.Errorf("smtp rcpt to %s: %w", envelopeTo, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	msg := BuildMessage(from, to, subject, htmlBody, textBody, hostName)
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return fmt.Errorf("smtp write data: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	_ = client.Quit()
	return nil
}

// BuildMessage constructs an RFC 5322 MIME multipart/alternative email message with headers.
func BuildMessage(from string, to []string, subject, htmlBody, textBody, host string) []byte {
	if host == "" {
		host = "exist.io"
	}
	if from == "" {
		from = fmt.Sprintf("noreply@%s", host)
	}

	cleanSubject := strings.ReplaceAll(subject, "\r", "")
	cleanSubject = strings.ReplaceAll(cleanSubject, "\n", " ")

	boundary := fmt.Sprintf("----=_ExistBoundary_%s", idgen.New("mime"))
	msgID := fmt.Sprintf("<%s@%s>", idgen.New("msg"), host)
	date := time.Now().UTC().Format(time.RFC1123Z)

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", cleanSubject))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", date))
	buf.WriteString(fmt.Sprintf("Message-ID: %s\r\n", msgID))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))

	// Plain text part
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	qpText := quotedprintable.NewWriter(&buf)
	_, _ = qpText.Write([]byte(textBody))
	_ = qpText.Close()
	buf.WriteString("\r\n")

	// HTML part
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	qpHtml := quotedprintable.NewWriter(&buf)
	_, _ = qpHtml.Write([]byte(htmlBody))
	_ = qpHtml.Close()
	buf.WriteString("\r\n")

	// Closing boundary
	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return buf.Bytes()
}

func extractEmail(addr string) string {
	addr = strings.TrimSpace(addr)
	if start := strings.Index(addr, "<"); start >= 0 {
		if end := strings.Index(addr[start:], ">"); end >= 0 {
			return addr[start+1 : start+end]
		}
	}
	return addr
}
