// Package metrics provides Prometheus-compatible telemetry for exist.io.
package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// Registry maintains in-process operational metrics.
type Registry struct {
	startTime             time.Time
	httpRequestsTotal     atomic.Uint64
	gitTransfersTotal     atomic.Uint64
	mcpCallsTotal         atomic.Uint64
	secretPushBlocksTotal atomic.Uint64
}

// DefaultRegistry is the global metric registry for the daemon.
var DefaultRegistry = NewRegistry()

// NewRegistry constructs an initialized telemetry Registry.
func NewRegistry() *Registry {
	return &Registry{
		startTime: time.Now(),
	}
}

// IncHTTPRequests records an incoming HTTP request.
func (r *Registry) IncHTTPRequests() {
	r.httpRequestsTotal.Add(1)
}

// IncGitTransfers records a completed Smart-HTTP or SSH transfer.
func (r *Registry) IncGitTransfers() {
	r.gitTransfersTotal.Add(1)
}

// IncMCPCalls records a JSON-RPC tool invocation by an autonomous agent.
func (r *Registry) IncMCPCalls() {
	r.mcpCallsTotal.Add(1)
}

// IncSecretPushBlocks records a blocked commit during push protection.
func (r *Registry) IncSecretPushBlocks() {
	r.secretPushBlocksTotal.Add(1)
}

// UptimeSeconds returns daemon uptime in seconds.
func (r *Registry) UptimeSeconds() float64 {
	return time.Since(r.startTime).Seconds()
}

// HTTPRequestsTotal returns the count of HTTP requests processed.
func (r *Registry) HTTPRequestsTotal() uint64 {
	return r.httpRequestsTotal.Load()
}

// GitTransfersTotal returns the count of Git operations processed.
func (r *Registry) GitTransfersTotal() uint64 {
	return r.gitTransfersTotal.Load()
}

// MCPCallsTotal returns the count of MCP tool invocations.
func (r *Registry) MCPCallsTotal() uint64 {
	return r.mcpCallsTotal.Load()
}

// SecretPushBlocksTotal returns the count of push protection blocks.
func (r *Registry) SecretPushBlocksTotal() uint64 {
	return r.secretPushBlocksTotal.Load()
}

// Handler returns an HTTP handler serving Prometheus text exposition format.
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		uptime := time.Since(r.startTime).Seconds()

		fmt.Fprintln(w, "# HELP exist_uptime_seconds Total runtime of the existd daemon.")
		fmt.Fprintln(w, "# TYPE exist_uptime_seconds gauge")
		fmt.Fprintf(w, "exist_uptime_seconds %.1f\n", uptime)

		fmt.Fprintln(w, "# HELP exist_http_requests_total Total number of HTTP router requests.")
		fmt.Fprintln(w, "# TYPE exist_http_requests_total counter")
		fmt.Fprintf(w, "exist_http_requests_total %d\n", r.httpRequestsTotal.Load())

		fmt.Fprintln(w, "# HELP exist_git_transfers_total Total number of Smart-HTTP git clone, fetch, or push operations.")
		fmt.Fprintln(w, "# TYPE exist_git_transfers_total counter")
		fmt.Fprintf(w, "exist_git_transfers_total %d\n", r.gitTransfersTotal.Load())

		fmt.Fprintln(w, "# HELP exist_mcp_tool_calls_total Total JSON-RPC tool calls executed by autonomous coding agents.")
		fmt.Fprintln(w, "# TYPE exist_mcp_tool_calls_total counter")
		fmt.Fprintf(w, "exist_mcp_tool_calls_total %d\n", r.mcpCallsTotal.Load())

		fmt.Fprintln(w, "# HELP exist_secret_push_blocks_total Total number of push attempts blocked by pre-receive secret scanning.")
		fmt.Fprintln(w, "# TYPE exist_secret_push_blocks_total counter")
		fmt.Fprintf(w, "exist_secret_push_blocks_total %d\n", r.secretPushBlocksTotal.Load())
	})
}
