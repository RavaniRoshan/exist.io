package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ConformanceResult records the outcome of an individual API conformance test.
type ConformanceResult struct {
	Operation string        `json:"operation"`
	Method    string        `json:"method"`
	Endpoint  string        `json:"endpoint"`
	Status    string        `json:"status"` // PASS / FAIL / SKIP
	Latency   time.Duration `json:"latency"`
	Details   string        `json:"details,omitempty"`
}

func (c *CLI) runConformance(args []string) error {
	fs := flag.NewFlagSet("conformance", flag.ContinueOnError)
	endpoint := fs.String("endpoint", c.Endpoint, "existd instance URL (e.g. http://127.0.0.1:8080)")
	token := fs.String("token", c.Token, "Admin API key or PAT")
	repoSlug := fs.String("repo", "acme/demo", "Target repository slug (owner/repo)")
	jsonOut := fs.Bool("json", false, "Output results as JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	baseURL := strings.TrimRight(*endpoint, "/")
	// If base URL has /v1/origin, strip for root
	rootURL := strings.TrimSuffix(baseURL, "/v1/origin")

	parts := strings.Split(*repoSlug, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo slug %q, expected owner/repo", *repoSlug)
	}
	owner, repo := parts[0], parts[1]

	results := runConformanceSuite(rootURL, *token, owner, repo)

	if *jsonOut {
		enc := json.NewEncoder(c.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	// Print Markdown Table
	fmt.Fprintln(c.Stdout, "| Operation | Method | Route | Conformance | Latency |")
	fmt.Fprintln(c.Stdout, "| :--- | :--- | :--- | :--- | :--- |")
	passed := 0
	for _, res := range results {
		badge := " PASS "
		if res.Status == "FAIL" {
			badge = "❌ FAIL"
		} else {
			passed++
		}
		fmt.Fprintf(c.Stdout, "| %s | `%s` | `%s` | %s | %v |\n",
			res.Operation, res.Method, res.Endpoint, badge, res.Latency.Round(time.Millisecond))
	}
	fmt.Fprintf(c.Stdout, "\nConformance Score: %d / %d operations verified (%.1f%%)\n",
		passed, len(results), float64(passed)/float64(len(results))*100)

	return nil
}

func runConformanceSuite(rootURL, token, owner, repo string) []ConformanceResult {
	client := &http.Client{Timeout: 10 * time.Second}
	var results []ConformanceResult

	testOp := func(name, method, endpoint string, body any, expectedCodes []int) ([]byte, int) {
		start := time.Now()
		var bodyReader io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(b)
		}
		req, err := http.NewRequestWithContext(context.Background(), method, rootURL+endpoint, bodyReader)
		if err != nil {
			results = append(results, ConformanceResult{
				Operation: name,
				Method:    method,
				Endpoint:  endpoint,
				Status:    "FAIL",
				Latency:   time.Since(start),
				Details:   err.Error(),
			})
			return nil, 0
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		latency := time.Since(start)
		if err != nil {
			results = append(results, ConformanceResult{
				Operation: name,
				Method:    method,
				Endpoint:  endpoint,
				Status:    "FAIL",
				Latency:   latency,
				Details:   err.Error(),
			})
			return nil, 0
		}
		defer resp.Body.Close()
		respBytes, _ := io.ReadAll(resp.Body)

		matched := false
		for _, code := range expectedCodes {
			if resp.StatusCode == code {
				matched = true
				break
			}
		}

		status := "PASS"
		details := ""
		if !matched {
			status = "FAIL"
			details = fmt.Sprintf("status %d, want %v: %s", resp.StatusCode, expectedCodes, string(respBytes))
		}

		results = append(results, ConformanceResult{
			Operation: name,
			Method:    method,
			Endpoint:  endpoint,
			Status:    status,
			Latency:   latency,
			Details:   details,
		})
		return respBytes, resp.StatusCode
	}

	// 1. Healthz
	testOp("HealthCheck", "GET", "/healthz", nil, []int{200})

	// 2. System Architecture Info
	testOp("GetSystemInfo", "GET", "/api/v1/system/info", nil, []int{200})

	// 3. User Authentication Profile
	testOp("GetCurrentUser", "GET", "/v1/origin/user", nil, []int{200})

	// 4. Token Exchange
	testOp("ExchangeAPIKey", "POST", "/auth/exchange_user_api_key", nil, []int{200, 400})

	// 5. List Namespaces
	testOp("ListNamespaces", "GET", "/v1/origin/namespaces", nil, []int{200})

	// 6. List Repositories
	testOp("ListRepositories", "GET", "/api/v1/repos", nil, []int{200})

	// 7. Get Repository Metadata
	testOp("GetRepository", "GET", fmt.Sprintf("/v1/origin/repos/%s/%s", owner, repo), nil, []int{200})

	// 8. Smart-HTTP Git Discovery
	testOp("GitHTTPDiscovery", "GET", fmt.Sprintf("/%s/%s.git/info/refs?service=git-upload-pack", owner, repo), nil, []int{200})

	// 9. Issue Agent Capability Lease
	leaseReq := map[string]any{
		"agentId":            "conformance-runner",
		"repoId":             fmt.Sprintf("%s/%s", owner, repo),
		"allowedRefPrefixes": []string{"refs/heads/agent/"},
		"ttlSeconds":         1800,
	}
	testOp("IssueAgentLease", "POST", "/api/v1/agent/leases", leaseReq, []int{200})

	// 10. List Pull Requests
	testOp("ListPullRequests", "GET", fmt.Sprintf("/v1/origin/repos/%s/%s/pulls", owner, repo), nil, []int{200})

	// 11. Model Context Protocol Tools List
	mcpToolsReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	}
	testOp("MCPToolsList", "POST", "/mcp", mcpToolsReq, []int{200})

	// 12. Model Context Protocol Tool Call (repo_tree)
	mcpTreeReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "repo_tree",
			"arguments": map[string]any{
				"ownerSlug": owner,
				"repoName":  repo,
			},
		},
	}
	testOp("MCPCallRepoTree", "POST", "/mcp", mcpTreeReq, []int{200})

	// 13. Model Context Protocol Read-Only Sandbox Isolation
	mcpRoReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "pull_request_create",
			"arguments": map[string]any{
				"ownerSlug": owner,
				"repoName":  repo,
				"title":     "Unauthorized Sandbox Write",
			},
		},
	}
	// On /mcp/readonly, write operations must return isError: true
	roOut, _ := testOp("MCPReadOnlySandboxGuard", "POST", "/mcp/readonly", mcpRoReq, []int{200})
	if len(results) > 0 && results[len(results)-1].Status == "PASS" {
		if !strings.Contains(string(roOut), `"isError":true`) {
			results[len(results)-1].Status = "FAIL"
			results[len(results)-1].Details = "sandbox permitted write tool invocation"
		}
	}

	return results
}
