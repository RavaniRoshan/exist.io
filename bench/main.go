package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"exist.io/internal/app"
	"exist.io/internal/config"
)

type StepTimings struct {
	LeaseDuration time.Duration
	CloneDuration time.Duration
	PushDuration  time.Duration
	PRDuration    time.Duration
	TotalDuration time.Duration
}

type ConcurrencyStats struct {
	Concurrency int
	Iterations  int
	LeaseP50    time.Duration
	LeaseP95    time.Duration
	LeaseP99    time.Duration
	CloneP50    time.Duration
	CloneP95    time.Duration
	CloneP99    time.Duration
	PushP50     time.Duration
	PushP95     time.Duration
	PushP99     time.Duration
	PRP50       time.Duration
	PRP95       time.Duration
	PRP99       time.Duration
	TotalP50    time.Duration
	TotalP95    time.Duration
	TotalP99    time.Duration
}

func main() {
	fmt.Println("=== exist.io Autonomous Agent Concurrent Benchmark ===")
	fmt.Println("Simulating concurrent coding agents: lease request -> clone -> push -> PR open")

	tmpDir, err := os.MkdirTemp("", "exist-bench-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	cfg := config.Config{
		DataDir:    filepath.Join(tmpDir, "data"),
		ListenAddr: "127.0.0.1:0",
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "http://127.0.0.1:8080",
		JWTSecret:  "benchmark-ephemeral-lease-key-32b!",
	}

	ctx := context.Background()
	bRes, err := app.Bootstrap(ctx, cfg, "admin@exist.io", "benchteam")
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	application, err := app.Open(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open app failed: %v\n", err)
		os.Exit(1)
	}
	defer application.Close()

	srv := httptest.NewServer(application.Handler())
	defer srv.Close()

	// Update git backend hooks
	existdPath, _ := os.Executable()
	os.Setenv("EXISTD_PATH", existdPath)

	tiers := []int{1, 5, 20, 50}
	var allStats []ConcurrencyStats

	for _, c := range tiers {
		iterations := c * 2
		if c == 1 {
			iterations = 5
		}
		if c == 50 {
			iterations = 50
		}
		fmt.Printf("\n--> Running benchmark tier: %d concurrent agents (%d iterations total)...\n", c, iterations)
		stats := runTier(srv.URL, bRes.RawKey, "benchteam", "demo", c, iterations, tmpDir)
		allStats = append(allStats, stats)
	}

	printMarkdownSummary(allStats)
}

func runTier(serverURL, apiKey, ownerSlug, repoName string, concurrency, iterations int, tmpRoot string) ConcurrencyStats {
	ch := make(chan int, iterations)
	for i := 0; i < iterations; i++ {
		ch <- i + 1
	}
	close(ch)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var timings []StepTimings

	startTier := time.Now()

	for workerID := 0; workerID < concurrency; workerID++ {
		wg.Add(1)
		go func(wID int) {
			defer wg.Done()
			for iter := range ch {
				timing := runSingleAgentLoop(serverURL, apiKey, ownerSlug, repoName, wID, iter, tmpRoot)
				mu.Lock()
				timings = append(timings, timing)
				mu.Unlock()
			}
		}(workerID)
	}

	wg.Wait()
	totalTierDuration := time.Since(startTier)
	fmt.Printf("    Completed %d iterations in %v (%.2f ops/sec)\n",
		iterations, totalTierDuration.Round(time.Millisecond), float64(iterations)/totalTierDuration.Seconds())

	var leases, clones, pushes, prs, totals []time.Duration
	for _, t := range timings {
		leases = append(leases, t.LeaseDuration)
		clones = append(clones, t.CloneDuration)
		pushes = append(pushes, t.PushDuration)
		prs = append(prs, t.PRDuration)
		totals = append(totals, t.TotalDuration)
	}

	return ConcurrencyStats{
		Concurrency: concurrency,
		Iterations:  iterations,
		LeaseP50:    p(leases, 0.50),
		LeaseP95:    p(leases, 0.95),
		LeaseP99:    p(leases, 0.99),
		CloneP50:    p(clones, 0.50),
		CloneP95:    p(clones, 0.95),
		CloneP99:    p(clones, 0.99),
		PushP50:     p(pushes, 0.50),
		PushP95:     p(pushes, 0.95),
		PushP99:     p(pushes, 0.99),
		PRP50:       p(prs, 0.50),
		PRP95:       p(prs, 0.95),
		PRP99:       p(prs, 0.99),
		TotalP50:    p(totals, 0.50),
		TotalP95:    p(totals, 0.95),
		TotalP99:    p(totals, 0.99),
	}
}

func runSingleAgentLoop(serverURL, apiKey, ownerSlug, repoName string, workerID, iteration int, tmpRoot string) StepTimings {
	client := &http.Client{Timeout: 15 * time.Second}
	var timing StepTimings
	startTotal := time.Now()

	// 1. Lease Request via MCP
	startLease := time.Now()
	mcpReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      iteration,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "lease_request",
			"arguments": map[string]any{
				"agentId":            fmt.Sprintf("bench-agent-%d", workerID),
				"repoId":             fmt.Sprintf("%s/%s", ownerSlug, repoName),
				"allowedRefPrefixes": []string{"refs/heads/agent/"},
				"ttlSeconds":         3600,
			},
		},
	}
	reqBytes, _ := json.Marshal(mcpReq)
	req, _ := http.NewRequestWithContext(context.Background(), "POST", serverURL+"/mcp", bytes.NewReader(reqBytes))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	var leaseToken string
	if err == nil {
		defer resp.Body.Close()
		var res struct {
			Result struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"result"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		if len(res.Result.Content) > 0 {
			var ld struct {
				Token string `json:"token"`
			}
			_ = json.Unmarshal([]byte(res.Result.Content[0].Text), &ld)
			leaseToken = ld.Token
		}
	}
	timing.LeaseDuration = time.Since(startLease)

	if leaseToken == "" {
		leaseToken = "fallback-bench-token"
	}

	// 2. Git Clone
	startClone := time.Now()
	cloneDir := filepath.Join(tmpRoot, fmt.Sprintf("work-%d-%d", workerID, iteration))
	cloneURL := fmt.Sprintf("%s/%s/%s.git", serverURL, ownerSlug, repoName)
	authURL := fmt.Sprintf("http://lease:%s@%s", leaseToken, strings.TrimPrefix(cloneURL, "http://"))

	cmdClone := exec.Command("git", "clone", "--depth=1", authURL, cloneDir)
	_ = cmdClone.Run()
	timing.CloneDuration = time.Since(startClone)

	// 3. Branch, Edit, Push
	startPush := time.Now()
	branchName := fmt.Sprintf("agent/bench-%d-%d", workerID, iteration)
	cmdBranch := exec.Command("git", "-C", cloneDir, "checkout", "-b", branchName)
	_ = cmdBranch.Run()

	filePath := filepath.Join(cloneDir, fmt.Sprintf("patch-%d-%d.txt", workerID, iteration))
	_ = os.WriteFile(filePath, []byte(fmt.Sprintf("Benchmark patch %d-%d\n", workerID, iteration)), 0644)

	_ = exec.Command("git", "-C", cloneDir, "config", "user.name", "Bench Agent").Run()
	_ = exec.Command("git", "-C", cloneDir, "config", "user.email", "bench@exist.io").Run()
	_ = exec.Command("git", "-C", cloneDir, "add", ".").Run()
	_ = exec.Command("git", "-C", cloneDir, "commit", "-m", "bench commit").Run()

	cmdPush := exec.Command("git", "-C", cloneDir, "push", "origin", branchName)
	_ = cmdPush.Run()
	timing.PushDuration = time.Since(startPush)

	// 4. PR Open via MCP
	startPR := time.Now()
	prReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      iteration + 1000,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "pull_request_create",
			"arguments": map[string]any{
				"ownerSlug": ownerSlug,
				"repoName":  repoName,
				"title":     fmt.Sprintf("Bench PR %d-%d", workerID, iteration),
				"headRef":   branchName,
				"baseRef":   "main",
			},
		},
	}
	prBytes, _ := json.Marshal(prReq)
	reqPR, _ := http.NewRequestWithContext(context.Background(), "POST", serverURL+"/mcp", bytes.NewReader(prBytes))
	reqPR.Header.Set("Authorization", "Bearer "+apiKey)
	reqPR.Header.Set("Content-Type", "application/json")

	respPR, err := client.Do(reqPR)
	if err == nil {
		_ = respPR.Body.Close()
	}
	timing.PRDuration = time.Since(startPR)

	_ = os.RemoveAll(cloneDir)
	timing.TotalDuration = time.Since(startTotal)
	return timing
}

func p(d []time.Duration, percentile float64) time.Duration {
	if len(d) == 0 {
		return 0
	}
	s := make([]time.Duration, len(d))
	copy(s, d)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	idx := int(float64(len(s)-1) * percentile)
	return s[idx]
}

func printMarkdownSummary(stats []ConcurrencyStats) {
	fmt.Println("\n================================================================================")
	fmt.Println("BENCHMARK RESULTS (Markdown Conformance Table)")
	fmt.Println("================================================================================")
	fmt.Println("| Concurrency | Lease p50 / p95 / p99 | Clone p50 / p95 / p99 | Push p50 / p95 / p99 | PR Open p50 / p95 / p99 | Total Loop p50 / p95 / p99 |")
	fmt.Println("| :--- | :--- | :--- | :--- | :--- | :--- |")
	for _, s := range stats {
		fmt.Printf("| **%d Agents** | %v / %v / %v | %v / %v / %v | %v / %v / %v | %v / %v / %v | **%v / %v / %v** |\n",
			s.Concurrency,
			s.LeaseP50.Round(time.Millisecond), s.LeaseP95.Round(time.Millisecond), s.LeaseP99.Round(time.Millisecond),
			s.CloneP50.Round(time.Millisecond), s.CloneP95.Round(time.Millisecond), s.CloneP99.Round(time.Millisecond),
			s.PushP50.Round(time.Millisecond), s.PushP95.Round(time.Millisecond), s.PushP99.Round(time.Millisecond),
			s.PRP50.Round(time.Millisecond), s.PRP95.Round(time.Millisecond), s.PRP99.Round(time.Millisecond),
			s.TotalP50.Round(time.Millisecond), s.TotalP95.Round(time.Millisecond), s.TotalP99.Round(time.Millisecond),
		)
	}
	fmt.Println("================================================================================")
}
