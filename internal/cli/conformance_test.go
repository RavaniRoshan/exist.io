package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLI_ConformanceCommand(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/mcp/readonly" {
			// Readonly sandbox returns isError: true on write attempts
			fmt.Fprintln(w, `{"jsonrpc":"2.0","id":3,"result":{"content":[{"type":"text","text":"read-only sandbox"}],"isError":true}}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"ok": true}`)
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint: ts.URL,
		Token:    "key_test_token_12345",
		Stdout:   &stdout,
		Stderr:   &stderr,
	}

	err := c.Run([]string{"conformance", "--endpoint", ts.URL, "--token", "key_test", "--repo", "acme/demo"})
	if err != nil {
		t.Fatalf("run conformance failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "| Operation | Method | Route | Conformance | Latency |") {
		t.Errorf("expected markdown table header in conformance output, got:\n%s", out)
	}
	if !strings.Contains(out, "Conformance Score:") {
		t.Errorf("expected Conformance Score summary, got:\n%s", out)
	}
}
