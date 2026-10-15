package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsRegistry(t *testing.T) {
	reg := NewRegistry()
	reg.IncHTTPRequests()
	reg.IncGitTransfers()
	reg.IncMCPCalls()
	reg.IncSecretPushBlocks()

	if reg.HTTPRequestsTotal() != 1 {
		t.Errorf("expected 1 http request, got %d", reg.HTTPRequestsTotal())
	}
	if reg.GitTransfersTotal() != 1 {
		t.Errorf("expected 1 git transfer, got %d", reg.GitTransfersTotal())
	}
	if reg.MCPCallsTotal() != 1 {
		t.Errorf("expected 1 mcp call, got %d", reg.MCPCallsTotal())
	}
	if reg.SecretPushBlocksTotal() != 1 {
		t.Errorf("expected 1 secret push block, got %d", reg.SecretPushBlocksTotal())
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	reg.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"exist_uptime_seconds",
		"exist_http_requests_total 1",
		"exist_git_transfers_total 1",
		"exist_mcp_tool_calls_total 1",
		"exist_secret_push_blocks_total 1",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected metric output to contain %q, got: %s", expected, body)
		}
	}
}
