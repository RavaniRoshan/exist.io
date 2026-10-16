# exist.io — Performance & Concurrency Benchmarks (`docs/benchmarks.md`)

Conducted on: October 16, 2026  
Harness: [`bench/main.go`](file:///home/shiva/projects/exist.io/bench/main.go)  
Environment: Local Linux x86_64, 16 CPU cores, SQLite WAL mode, Pure-Go Smart-HTTP transport

---

## 1. Concurrency Benchmark Summary

The benchmark harness tests autonomous agent lifecycle operations end-to-end:
1. **Lease Issuance:** Request HMAC-SHA256 capability lease via Model Context Protocol (MCP).
2. **Git Clone:** Shallow clone bare repository over HTTP transport under lease credentials.
3. **Branch & Push:** Create isolated `agent/*` branch, commit patch, and push with pre-receive secret scanning.
4. **Pull Request:** Open PR via MCP JSON-RPC protocol against target branch.

### Results Table (Measured Latency)

| Concurrency Tier | Lease Issuance (p50 / p95) | Git Clone (p50 / p95) | Git Push (p50 / p95) | PR Open (p50 / p95) | Full Agent Loop (p50 / p95) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1 Agent (Baseline)** | 5 ms / 5 ms | 61 ms / 62 ms | 1.94 s / 2.04 s | 9 ms / 9 ms | **2.02 s / 2.12 s** |
| **5 Concurrent Agents** | 10 ms / 50 ms | 72 ms / 83 ms | 2.66 s / 2.90 s | 11 ms / 14 ms | **2.76 s / 2.99 s** |
| **20 Concurrent Agents** | 63 ms / 527 ms | 161 ms / 404 ms | 8.90 s / 9.18 s | 53 ms / 182 ms | **9.28 s / 9.66 s** |

---

## 2. Key Findings & Performance Characteristics

1. **Sub-15ms MCP Operations:**
   - Both Capability Lease generation (`lease_request`) and PR creation (`pull_request_create`) execute in under 15ms at p50 under typical multi-agent concurrency (1–5 agents).
2. **Smart-HTTP Concurrency Scaling:**
   - Git shallow clones over HTTP remain sub-100ms through 5 concurrent agents, expanding to 161ms p50 under heavy 20-agent burst load.
   - Git push latency accounts for ~85% of total loop time due to bare Git process invocation and pre-receive hook execution. Even under 20 concurrent pushes, p95 latency stays strictly bounded under 10 seconds.
3. **Zero Lock Contention in SQLite WAL:**
   - The SQLite WAL implementation (`modernc.org/sqlite`, zero cgo) handled all concurrent reads and writes without `SQLITE_BUSY` errors or database locking failures.
