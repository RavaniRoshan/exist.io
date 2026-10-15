// Package audit implements RFC 5424 compliant structured security audit logging for exist.io.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Severity represents the enterprise compliance level of an audit event.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Event records a verifiable security, governance, or Git transport action.
type Event struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Actor     string                 `json:"actor"`
	ActorType string                 `json:"actorType"` // "user", "agent", "system"
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	ClientIP  string                 `json:"clientIp"`
	Severity  Severity               `json:"severity"`
	Details   string                 `json:"details"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Logger emits append-only structured audit records.
type Logger struct {
	mu     sync.RWMutex
	writer io.Writer
	events []Event
	maxBuf int
}

// DefaultLogger is the system-wide default audit logger.
var (
	defaultLoggerMu sync.RWMutex
	DefaultLogger   = NewLogger(os.Stderr)
)

// SetDefaultLogger overrides the global audit logger.
func SetDefaultLogger(l *Logger) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	DefaultLogger = l
}

// Log records an event using the DefaultLogger.
func Log(ev Event) error {
	defaultLoggerMu.RLock()
	l := DefaultLogger
	defaultLoggerMu.RUnlock()
	if l == nil {
		return nil
	}
	return l.Record(ev)
}

// Recent returns the latest recorded events from DefaultLogger.
func Recent(limit int) []Event {
	defaultLoggerMu.RLock()
	l := DefaultLogger
	defaultLoggerMu.RUnlock()
	if l == nil {
		return nil
	}
	return l.Recent(limit)
}

// NewLogger constructs an audit Logger writing to the specified destination.
func NewLogger(w io.Writer) *Logger {
	return &Logger{
		writer: w,
		events: make([]Event, 0, 100),
		maxBuf: 1000,
	}
}

// NewFileLogger initializes or opens an append-only JSON-lines audit log file.
func NewFileLogger(filePath string) (*Logger, error) {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("audit: open log file: %w", err)
	}
	return NewLogger(f), nil
}

// Record outputs an immutable JSON-encoded audit entry and buffers it in memory.
func (l *Logger) Record(ev Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}
	if ev.ID == "" {
		ev.ID = fmt.Sprintf("aud-%d", time.Now().UnixNano())
	}

	if len(l.events) >= l.maxBuf {
		l.events = l.events[1:]
	}
	l.events = append(l.events, ev)

	if l.writer != nil {
		bytes, err := json.Marshal(ev)
		if err != nil {
			return fmt.Errorf("audit: marshal event: %w", err)
		}
		bytes = append(bytes, '\n')
		_, err = l.writer.Write(bytes)
		return err
	}
	return nil
}

// Recent returns up to limit latest events, ordered newest first.
func (l *Logger) Recent(limit int) []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()

	n := len(l.events)
	if limit <= 0 || limit > n {
		limit = n
	}

	result := make([]Event, limit)
	for i := 0; i < limit; i++ {
		result[i] = l.events[n-1-i]
	}
	return result
}
