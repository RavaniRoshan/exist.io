// Package telemetry implements the consent-gated activation ping.
package telemetry

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Config controls the telemetry client. Enabled defaults to false.
type Config struct {
	Enabled  bool
	Endpoint string
	Version  string
	DataDir  string
}

// Payload is the complete set of fields that may leave the instance.
type Payload struct {
	InstanceID string `json:"instance_id"`
	Event      string `json:"event"`
	Version    string `json:"version"`
	OS         string `json:"os"`
	Date       string `json:"date"`
}

// Client records activation events. A zero-value client is disabled.
type Client struct {
	cfg  Config
	http *http.Client
}

// New returns a telemetry client.
func New(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 5 * time.Second},
	}
}

// Record sends one event when telemetry is enabled. It never blocks the caller
// for more than the HTTP timeout and never returns an error to the request path.
func (c *Client) Record(ctx context.Context, event string) error {
	if c == nil || !c.cfg.Enabled {
		return nil
	}

	instanceID, err := c.instanceID()
	if err != nil {
		return err
	}

	payload := Payload{
		InstanceID: instanceID,
		Event:      event,
		Version:    c.cfg.Version,
		OS:         runtime.GOOS,
		Date:       time.Now().UTC().Format("2006-01-02"),
	}

	if c.cfg.Endpoint == "" {
		log.Printf("telemetry: %s recorded locally (no endpoint configured)", event)
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	_, _ = io.Copy(io.Discard, response.Body)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("telemetry: endpoint returned HTTP %d", response.StatusCode)
	}
	return nil
}

func (c *Client) instanceID() (string, error) {
	path := filepath.Join(c.cfg.DataDir, "instance_id")
	if body, err := os.ReadFile(path); err == nil { // #nosec G304 -- path is derived from the operator-configured data directory
		if id := strings.TrimSpace(string(body)); id != "" {
			return id, nil
		}
	}

	if err := os.MkdirAll(c.cfg.DataDir, 0o700); err != nil {
		return "", err
	}
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	id := hex.EncodeToString(buf)
	if err := os.WriteFile(path, []byte(id), 0o600); err != nil {
		return "", err
	}
	return id, nil
}
