# exist.io — Performance & Concurrency Benchmarks (`docs/benchmarks.md`)

Conducted on: November 1, 2026  
Harness: [`bench/main.go`](../bench/main.go)  
Environment: Linux x86_64 (WSL2 kernel 6.6), 16 CPU cores, Go 1.27.2, SQLite WAL mode, Pure-Go Smart-HTTP transport  
Run type: single local run, no other workload. Numbers vary with hardware; reproduce with `go run ./bench/main.go`.

---

## 1. Concurrency Benchmark Summary

The benchmark harness tests autonomous agent lifecycle operations end-to-end:
1. **Lease Issuance:** Request HMAC-SHA256 capability lease via Model Context Protocol (MCP).
2. **Git Clone:** Shallow clone bare repository over HTTP transport under lease credentials.
3. **Branch & Push:** Create isolated `agent/*` branch, commit patch, and push with pre-receive secret scanning.
4. **Pull Request:** Open PR via MCP JSON-RPC protocol against target branch.

### Results Table (Measured Latency, p50 / p95)

| Concurrency Tier | Lease Issuance | Git Clone | Git Push | PR Open | Full Agent Loop |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1 Agent (Baseline)** | 4 ms / 4 ms | 60 ms / 65 ms | 2.852 s / 3.019 s | 8 ms / 8 ms | **2.937 s / 3.097 s** |
| **5 Concurrent Agents** | 7 ms / 29 ms | 61 ms / 70 ms | 3.951 s / 4.242 s | 9 ms / 12 ms | **4.068 s / 4.322 s** |
| **20 Concurrent Agents** | 53 ms / 601 ms | 127 ms / 489 ms | 13.581 s / 14.522 s | 34 ms / 122 ms | **13.870 s / 15.381 s** |
| **50 Concurrent Agents** | 254 ms / 1.310 s | 189 ms / 252 ms | 37.129 s / 40.418 s | 145 ms / 861 ms | **37.880 s / 41.799 s** |

---

## 2. Key Findings & Performance Characteristics

1. **MCP operations stay in the low milliseconds at 1–5 agents:**
   - Capability lease generation (`lease_request`) measured 4–7 ms p50 and 4–29 ms p95 at 1–5 agents.
   - PR creation (`pull_request_create`) measured 8–9 ms p50 and 8–12 ms p95 at 1–5 agents.
   - At 20 agents, lease p95 rises to 601 ms and PR-open p95 to 122 ms.
2. **Smart-HTTP clone remains fast; push dominates the loop:**
   - Shallow clones measured 60–127 ms p50 through 20 agents and 189 ms p50 at 50 agents.
   - Git push accounts for the majority of total loop time: 2.85 s of 2.94 s at 1 agent, and 13.6 s of 13.9 s at 20 agents.
   - At 50 concurrent pushes, p95 push latency measured 40.4 s.
3. **No SQLite lock failures observed:**
   - The SQLite WAL implementation (`modernc.org/sqlite`, zero cgo) completed the run without `SQLITE_BUSY` errors or database locking failures.
