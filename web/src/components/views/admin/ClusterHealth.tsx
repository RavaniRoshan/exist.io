import React, { useState, useEffect } from 'react'
import {
  Server,
  Activity,
  Database,
  Cloud,
  HardDrive,
  Cpu,
  RefreshCw,
  CheckCircle2,
  Terminal,
  Layers,
  Bot,
  Copy,
  Check,
} from 'lucide-react'
import { Box, BoxHeader, BoxBody } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { getClusterMetrics, type ClusterMetrics } from '@/lib/api'

export const ClusterHealth: React.FC = () => {
  const [metrics, setMetrics] = useState<ClusterMetrics | null>(null)
  const [refreshing, setRefreshing] = useState(false)
  const [copiedAction, setCopiedAction] = useState<string | null>(null)

  const loadMetrics = async () => {
    setRefreshing(true)
    try {
      const data = await getClusterMetrics()
      setMetrics(data)
    } finally {
      setRefreshing(false)
    }
  }

  useEffect(() => {
    loadMetrics()
  }, [])

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopiedAction(id)
    setTimeout(() => setCopiedAction(null), 2000)
  }

  if (!metrics) {
    return (
      <Box className="p-12 text-center text-muted-foreground font-mono text-xs">
        <Activity className="w-4 h-4 animate-spin mr-2 inline text-primary" />
        <span>Loading cluster health telemetry...</span>
      </Box>
    )
  }

  const formatBytes = (bytes: number) => {
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  return (
    <div className="space-y-4">
      {/* Primary Status Header Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        <div className="space-y-0.5">
          <div className="flex items-center gap-2">
            <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
              <Server className="w-4 h-4 text-primary" />
              <span>Primary Forge Leader</span>
            </h2>
            <Badge variant="success" pulse>
              <CheckCircle2 className="w-3 h-3 mr-1" /> System Operational
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground font-mono">
            {metrics.version} • Uptime: {metrics.uptime} • Architecture: linux/amd64 pure Go
          </p>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={loadMetrics}
          disabled={refreshing}
          className="gap-1.5 self-start sm:self-auto text-xs"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${refreshing ? 'animate-spin' : ''}`} />
          <span>Refresh Telemetry</span>
        </Button>
      </div>

      {/* Operator Metric Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        {/* Metric 1: Database WAL State */}
        <Box className="p-3.5 space-y-1.5">
          <div className="flex items-center justify-between text-muted-foreground text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <Database className="w-3.5 h-3.5 text-primary" /> SQLite WAL Size
            </span>
            <span className="text-[10px] bg-secondary px-1.5 py-0.5 rounded border border-border font-mono text-foreground">
              WAL MODE
            </span>
          </div>
          <div className="text-xl font-bold font-mono text-foreground tracking-tight">
            {formatBytes(metrics.sqliteWalSizeBytes)}
          </div>
          <p className="text-[11px] text-muted-foreground">
            Zero-lock pure Go SQLite storage engine with automatic WAL checkpointing.
          </p>
        </Box>

        {/* Metric 2: Litestream Continuous Sync */}
        <Box className="p-3.5 space-y-1.5">
          <div className="flex items-center justify-between text-muted-foreground text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <Cloud className="w-3.5 h-3.5 text-primary" /> Hot Replication
            </span>
            <Badge variant={metrics.litestreamSyncStatus === 'active' ? 'success' : 'outline'} className="text-[9px] py-0">
              {metrics.litestreamSyncStatus}
            </Badge>
          </div>
          <div className="text-xl font-bold font-mono text-foreground tracking-tight">
            Continuous
          </div>
          <p className="text-[11px] text-muted-foreground truncate" title={metrics.lastBackupTimestamp}>
            Backup: {metrics.lastBackupTimestamp}
          </p>
        </Box>

        {/* Metric 3: Active Git Streams */}
        <Box className="p-3.5 space-y-1.5">
          <div className="flex items-center justify-between text-muted-foreground text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <HardDrive className="w-3.5 h-3.5 text-primary" /> Git Transports
            </span>
            <span className="text-[10px] text-emerald-500 bg-secondary px-1.5 py-0.5 rounded border border-border font-mono">
              HTTP & SSH :2222
            </span>
          </div>
          <div className="text-xl font-bold font-mono text-foreground tracking-tight">
            {metrics.activeGitSessions} <span className="text-xs font-normal text-muted-foreground">concurrent</span>
          </div>
          <p className="text-[11px] text-muted-foreground">
            Pure-Go authenticated packfile streaming over Smart-HTTP and port 2222 SSH.
          </p>
        </Box>

        {/* Metric 4: Runtime Load & Memory */}
        <Box className="p-3.5 space-y-1.5">
          <div className="flex items-center justify-between text-muted-foreground text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <Cpu className="w-3.5 h-3.5 text-primary" /> Resources
            </span>
            <span className="text-[10px] text-muted-foreground font-mono">
              {metrics.goroutines} goroutines
            </span>
          </div>
          <div className="text-xl font-bold font-mono text-foreground tracking-tight">
            {metrics.allocatedMemoryMb} <span className="text-xs font-normal text-muted-foreground">MB RAM</span>
          </div>
          <p className="text-[11px] text-muted-foreground">
            CPU: {metrics.cpuUsagePercent}% • Zero cgo overhead.
          </p>
        </Box>
      </div>

      {/* Subsystem Architecture Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Subsystem 1: Pure-Go SSH Server & Git LFS */}
        <Box>
          <BoxHeader className="bg-secondary/35 py-2">
            <div className="flex items-center gap-2">
              <Terminal className="w-4 h-4 text-primary" />
              <span className="text-xs font-semibold text-foreground">Pure-Go SSH & Git LFS</span>
            </div>
            <Badge variant="success" className="text-[10px] font-mono py-0">Port 2222</Badge>
          </BoxHeader>
          <BoxBody className="p-3.5 space-y-2.5">
            <p className="text-xs text-muted-foreground">
              Autonomous standalone SSH daemon with OpenSSH SIGSPEC signature parsing and Git LFS v1 Batch object engine.
            </p>
            <div className="space-y-1 font-mono text-[11px] text-foreground pt-1">
              <div className="flex justify-between p-1.5 rounded bg-secondary/40 border border-border">
                <span className="text-muted-foreground">SSH Listener:</span>
                <span>0.0.0.0:2222</span>
              </div>
              <div className="flex justify-between p-1.5 rounded bg-secondary/40 border border-border">
                <span className="text-muted-foreground">Git LFS CAS:</span>
                <span className="text-emerald-500">SHA-256 Batch API</span>
              </div>
              <div className="flex justify-between p-1.5 rounded bg-secondary/40 border border-border">
                <span className="text-muted-foreground">Cert Auth:</span>
                <span>SSH CA Supported</span>
              </div>
            </div>
          </BoxBody>
        </Box>

        {/* Subsystem 2: Model Context Protocol (MCP) */}
        <Box>
          <BoxHeader className="bg-secondary/35 py-2">
            <div className="flex items-center gap-2">
              <Bot className="w-4 h-4 text-primary" />
              <span className="text-xs font-semibold text-foreground">MCP Agent Runtime</span>
            </div>
            <Badge variant="agent" className="text-[10px] font-mono py-0">34 Tools</Badge>
          </BoxHeader>
          <BoxBody className="p-3.5 space-y-2.5">
            <p className="text-xs text-muted-foreground">
              Native Model Context Protocol JSON-RPC endpoints mounted at <code>/mcp</code> for autonomous agents.
            </p>
            <div className="space-y-1 font-mono text-[11px] text-foreground pt-1">
              <div className="flex justify-between p-1.5 rounded bg-secondary/40 border border-border">
                <span className="text-muted-foreground">MCP Protocol:</span>
                <span>JSON-RPC 2.0</span>
              </div>
              <div className="flex justify-between p-1.5 rounded bg-secondary/40 border border-border">
                <span className="text-muted-foreground">Review Threading:</span>
                <span className="text-emerald-500">Inline Check Runs</span>
              </div>
              <div className="flex justify-between p-1.5 rounded bg-secondary/40 border border-border">
                <span className="text-muted-foreground">Tool Scope Auth:</span>
                <span>Ambient & Token</span>
              </div>
            </div>
          </BoxBody>
        </Box>

        {/* Subsystem 3: Storage & DR Replications */}
        <Box>
          <BoxHeader className="bg-secondary/35 py-2">
            <div className="flex items-center gap-2">
              <Layers className="w-4 h-4 text-primary" />
              <span className="text-xs font-semibold text-foreground">Zero-Cgo Storage Store</span>
            </div>
            <Badge variant="outline" className="text-[10px] font-mono py-0">
              modernc.org/sqlite
            </Badge>
          </BoxHeader>
          <BoxBody className="p-3.5 space-y-2.5">
            <p className="text-xs text-muted-foreground">
              Single-binary deployment architecture with zero external daemon dependencies and continuous WAL sync.
            </p>
            <div className="space-y-1 font-mono text-[11px] text-foreground pt-1">
              <div className="flex justify-between p-1.5 rounded bg-secondary/40 border border-border">
                <span className="text-muted-foreground">Journal Mode:</span>
                <span>WAL (PRAGMA wal)</span>
              </div>
              <div className="flex justify-between p-1.5 rounded bg-secondary/40 border border-border">
                <span className="text-muted-foreground">Git Backend:</span>
                <span className="text-emerald-500">go-git/v6 Bare Repos</span>
              </div>
              <div className="flex justify-between p-1.5 rounded bg-secondary/40 border border-border">
                <span className="text-muted-foreground">Sync Target:</span>
                <span className="truncate max-w-[120px]">s3://backups</span>
              </div>
            </div>
          </BoxBody>
        </Box>
      </div>

      {/* Operator Scrape & Quick Actions Bar */}
      <Box>
        <BoxHeader className="bg-secondary/35 py-2.5">
          <div className="flex items-center gap-2">
            <Activity className="w-4 h-4 text-emerald-500" />
            <span className="text-xs font-semibold text-foreground">Enterprise Observability & Scrape Endpoints</span>
          </div>
          <Badge variant="outline" className="text-[10px] font-mono">
            Kubernetes / Prometheus Ready
          </Badge>
        </BoxHeader>

        <BoxBody className="p-4 space-y-3">
          <p className="text-xs text-muted-foreground">
            Standard endpoints configured for enterprise Kubernetes liveness probes and Grafana dashboards.
          </p>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-2.5 font-mono text-xs">
            <div className="flex items-center justify-between p-2.5 rounded-md bg-secondary/30 border border-border">
              <div>
                <span className="text-foreground block font-semibold">GET /healthz</span>
                <span className="text-muted-foreground text-[11px]">HTTP Liveness Probe</span>
              </div>
              <Badge variant="success" className="text-[10px] py-0 font-mono">200 OK</Badge>
            </div>

            <div className="flex items-center justify-between p-2.5 rounded-md bg-secondary/30 border border-border">
              <div>
                <span className="text-foreground block font-semibold">GET /metrics</span>
                <span className="text-muted-foreground text-[11px]">Prometheus Exposition</span>
              </div>
              <button
                onClick={() => handleCopy('http://localhost:8080/metrics', 'prom-url')}
                className="text-muted-foreground hover:text-foreground transition-colors"
                title="Copy Prometheus URL"
              >
                {copiedAction === 'prom-url' ? (
                  <Check className="w-4 h-4 text-emerald-500" />
                ) : (
                  <Copy className="w-4 h-4" />
                )}
              </button>
            </div>

            <div className="flex items-center justify-between p-2.5 rounded-md bg-secondary/30 border border-border">
              <div>
                <span className="text-foreground block font-semibold">GET /v1/origin/openapi.yaml</span>
                <span className="text-muted-foreground text-[11px]">OpenAPI 3.1 Contract</span>
              </div>
              <Badge variant="outline" className="text-[10px] py-0 font-mono">
                34 Routes
              </Badge>
            </div>
          </div>
        </BoxBody>
      </Box>
    </div>
  )
}
