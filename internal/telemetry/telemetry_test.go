package telemetry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestDisabledByDefault(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(Config{Enabled: false, Endpoint: server.URL, DataDir: t.TempDir(), Version: "v0.6.0-beta"})
	if err := client.Record(context.Background(), "activation.first_merge"); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&requests); got != 0 {
		t.Fatalf("disabled client sent %d request(s)", got)
	}
}

func TestEnabledSendsWhitelistedPayload(t *testing.T) {
	payloads := make(chan Payload, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload Payload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		payloads <- payload
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := New(Config{Enabled: true, Endpoint: server.URL, DataDir: t.TempDir(), Version: "v0.6.0-beta"})
	if err := client.Record(context.Background(), "activation.first_merge"); err != nil {
		t.Fatal(err)
	}

	select {
	case payload := <-payloads:
		if payload.Event != "activation.first_merge" {
			t.Fatalf("event = %q", payload.Event)
		}
		if payload.Version != "v0.6.0-beta" {
			t.Fatalf("version = %q", payload.Version)
		}
		if payload.OS == "" || payload.Date == "" {
			t.Fatalf("missing os/date: %+v", payload)
		}
		if len(payload.InstanceID) != 32 {
			t.Fatalf("instance id length = %d, want 32", len(payload.InstanceID))
		}
	default:
		t.Fatal("endpoint received no payload")
	}
}

func TestPayloadHasNoHostOrPathData(t *testing.T) {
	payload := Payload{
		InstanceID: "0123456789abcdef0123456789abcdef",
		Event:      "activation.first_merge",
		Version:    "v0.6.0-beta",
		OS:         "linux",
		Date:       "2026-11-01",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"hostname", "repo", "repository", "path", "user", "email"} {
		if _, ok := fields[forbidden]; ok {
			t.Fatalf("payload contains forbidden field %q", forbidden)
		}
	}
	if len(fields) != 5 {
		t.Fatalf("payload has %d fields, want 5: %v", len(fields), fields)
	}
}

func TestInstanceIDIsStable(t *testing.T) {
	dir := t.TempDir()
	first := New(Config{Enabled: true, DataDir: dir, Version: "v0.6.0-beta"})
	id1, err := first.instanceID()
	if err != nil {
		t.Fatal(err)
	}
	second := New(Config{Enabled: true, DataDir: dir, Version: "v0.6.0-beta"})
	id2, err := second.instanceID()
	if err != nil {
		t.Fatal(err)
	}
	if id1 != id2 {
		t.Fatalf("instance id changed: %q != %q", id1, id2)
	}

	info, err := os.Stat(filepath.Join(dir, "instance_id"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("instance_id mode = %o, want 600", info.Mode().Perm())
	}
}

func TestEnabledWithoutEndpointRecordsLocally(t *testing.T) {
	client := New(Config{Enabled: true, DataDir: t.TempDir(), Version: "v0.6.0-beta"})
	if err := client.Record(context.Background(), "activation.first_merge"); err != nil {
		t.Fatal(err)
	}
}
