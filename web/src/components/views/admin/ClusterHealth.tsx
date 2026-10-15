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
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
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
      <div className="flex items-center justify-center p-12 text-slate-500 font-mono text-xs">
        <Activity className="w-4 h-4 animate-spin mr-2 text-sky-400" />
        <span>Loading cluster health telemetry...</span>
      </div>
    )
  }

  const formatBytes = (bytes: number) => {
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  return (
    <div className="space-y-6">
      {/* Primary Status Banner */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-5 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md">
        <div className="flex items-center gap-3.5">
          <div className="flex items-center justify-center w-11 h-11 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400">
            <Server className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="text-base font-semibold text-white">Cluster Node: Primary Forge Leader</h3>
              <Badge variant="success" pulse>
                <CheckCircle2 className="w-3 h-3 mr-1" /> System Operational
              </Badge>
            </div>
            <p className="text-xs text-slate-400 font-mono mt-0.5">
              {metrics.version} • Uptime: {metrics.uptime} • Architecture: linux/amd64 pure Go
            </p>
          </div>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={loadMetrics}
          loading={refreshing}
          className="gap-1.5 self-start sm:self-auto text-slate-300"
        >
          <RefreshCw className="w-3.5 h-3.5" />
          <span>Refresh Telemetry</span>
        </Button>
      </div>

      {/* Operator Metric Cards Grid (Answers: Is it healthy? What is the bottleneck?) */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Metric 1: Database WAL State */}
        <Card className="border-slate-800 bg-slate-950/60 p-4 space-y-2">
          <div className="flex items-center justify-between text-slate-400 text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <Database className="w-4 h-4 text-sky-400" /> SQLite WAL Size
            </span>
            <span className="text-[10px] bg-slate-900 px-1.5 py-0.5 rounded border border-slate-800 font-mono text-sky-300">
              WAL MODE
            </span>
          </div>
          <div className="text-2xl font-bold font-mono text-white tracking-tight">
            {formatBytes(metrics.sqliteWalSizeBytes)}
          </div>
          <p className="text-[11px] text-slate-500">
            Zero-lock pure Go SQLite storage engine with automatic WAL checkpointing.
          </p>
        </Card>

        {/* Metric 2: Litestream Continuous Sync */}
        <Card className="border-slate-800 bg-slate-950/60 p-4 space-y-2">
          <div className="flex items-center justify-between text-slate-400 text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <Cloud className="w-4 h-4 text-purple-400" /> Hot Replication
            </span>
            <Badge variant={metrics.litestreamSyncStatus === 'active' ? 'success' : 'outline'} className="text-[9px] py-0">
              {metrics.litestreamSyncStatus}
            </Badge>
          </div>
          <div className="text-2xl font-bold font-mono text-white tracking-tight">
            Continuous
          </div>
          <p className="text-[11px] text-slate-500 truncate" title={metrics.lastBackupTimestamp}>
            Backup: {metrics.lastBackupTimestamp}
          </p>
        </Card>

        {/* Metric 3: Active Git Streams */}
        <Card className="border-slate-800 bg-slate-950/60 p-4 space-y-2">
          <div className="flex items-center justify-between text-slate-400 text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <HardDrive className="w-4 h-4 text-emerald-400" /> Git Transports
            </span>
            <span className="text-[10px] text-emerald-400 bg-emerald-950/40 px-1.5 py-0.5 rounded border border-emerald-800 font-mono">
              HTTP & SSH :2222
            </span>
          </div>
          <div className="text-2xl font-bold font-mono text-white tracking-tight">
            {metrics.activeGitSessions} <span className="text-xs font-normal text-slate-400">concurrent</span>
          </div>
          <p className="text-[11px] text-slate-500">
            Pure-Go authenticated packfile streaming over Smart-HTTP and port 2222 SSH.
          </p>
        </Card>

        {/* Metric 4: Runtime Load & Memory */}
        <Card className="border-slate-800 bg-slate-950/60 p-4 space-y-2">
          <div className="flex items-center justify-between text-slate-400 text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <Cpu className="w-4 h-4 text-amber-400" /> Process Resources
            </span>
            <span className="text-[10px] text-slate-400 font-mono">
              {metrics.goroutines} goroutines
            </span>
          </div>
          <div className="text-2xl font-bold font-mono text-white tracking-tight">
            {metrics.allocatedMemoryMb} <span className="text-xs font-normal text-slate-400">MB RAM</span>
          </div>
          <p className="text-[11px] text-slate-500">
            CPU: {metrics.cpuUsagePercent}% • Hermetic memory footprint with zero cgo overhead.
          </p>
        </Card>
      </div>

      {/* Subsystem Architecture Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Subsystem 1: Pure-Go SSH Server & Git LFS */}
        <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Terminal className="w-4 h-4 text-sky-400" />
              <h4 className="text-sm font-semibold text-white">Pure-Go SSH & Git LFS</h4>
            </div>
            <Badge variant="success" className="text-[10px] font-mono py-0">Port 2222</Badge>
          </div>
          <p className="text-xs text-slate-400">
            Autonomous standalone SSH daemon with OpenSSH SIGSPEC signature parsing and Git LFS v1 Batch object engine.
          </p>
          <div className="space-y-1.5 font-mono text-[11px] text-slate-300 pt-1">
            <div className="flex justify-between p-2 rounded bg-slate-900/60 border border-slate-800">
              <span className="text-slate-500">SSH Listener:</span>
              <span className="text-sky-300">0.0.0.0:2222</span>
            </div>
            <div className="flex justify-between p-2 rounded bg-slate-900/60 border border-slate-800">
              <span className="text-slate-500">Git LFS CAS:</span>
              <span className="text-emerald-400">SHA-256 Batch API</span>
            </div>
            <div className="flex justify-between p-2 rounded bg-slate-900/60 border border-slate-800">
              <span className="text-slate-500">Cert Authentication:</span>
              <span className="text-purple-300">SSH CA Supported</span>
            </div>
          </div>
        </Card>

        {/* Subsystem 2: Model Context Protocol (MCP) */}
        <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Bot className="w-4 h-4 text-sky-400" />
              <h4 className="text-sm font-semibold text-white">MCP Agent Runtime</h4>
            </div>
            <Badge variant="agent" className="text-[10px] font-mono py-0">34 Tools</Badge>
          </div>
          <p className="text-xs text-slate-400">
            Native Model Context Protocol JSON-RPC endpoints mounted at <code>/mcp</code> for autonomous agents.
          </p>
          <div className="space-y-1.5 font-mono text-[11px] text-slate-300 pt-1">
            <div className="flex justify-between p-2 rounded bg-slate-900/60 border border-slate-800">
              <span className="text-slate-500">MCP Protocol:</span>
              <span className="text-sky-300">JSON-RPC 2.0</span>
            </div>
            <div className="flex justify-between p-2 rounded bg-slate-900/60 border border-slate-800">
              <span className="text-slate-500">Review Threading:</span>
              <span className="text-emerald-400">Inline Check Runs</span>
            </div>
            <div className="flex justify-between p-2 rounded bg-slate-900/60 border border-slate-800">
              <span className="text-slate-500">Tool Scope Auth:</span>
              <span className="text-purple-300">Ambient & Token</span>
            </div>
          </div>
        </Card>

        {/* Subsystem 3: Storage & DR Replications */}
        <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Layers className="w-4 h-4 text-purple-400" />
              <h4 className="text-sm font-semibold text-white">Zero-Cgo Storage Store</h4>
            </div>
            <Badge variant="outline" className="text-[10px] text-purple-300 border-purple-500/40 font-mono py-0">
              modernc.org/sqlite
            </Badge>
          </div>
          <p className="text-xs text-slate-400">
            Single-binary deployment architecture with zero external daemon dependencies and continuous WAL sync.
          </p>
          <div className="space-y-1.5 font-mono text-[11px] text-slate-300 pt-1">
            <div className="flex justify-between p-2 rounded bg-slate-900/60 border border-slate-800">
              <span className="text-slate-500">Journal Mode:</span>
              <span className="text-purple-300">WAL (PRAGMA wal)</span>
            </div>
            <div className="flex justify-between p-2 rounded bg-slate-900/60 border border-slate-800">
              <span className="text-slate-500">Git Backend:</span>
              <span className="text-emerald-400">go-git/v6 Bare Repos</span>
            </div>
            <div className="flex justify-between p-2 rounded bg-slate-900/60 border border-slate-800">
              <span className="text-slate-500">Sync Target:</span>
              <span className="text-sky-300 truncate max-w-[120px]">s3://backups</span>
            </div>
          </div>
        </Card>
      </div>

      {/* Operator Scrape & Quick Actions Bar */}
      <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
        <CardHeader className="p-0">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm text-white flex items-center gap-2">
              <Activity className="w-4 h-4 text-emerald-400" />
              <span>Enterprise Observability & Scrape Endpoints</span>
            </CardTitle>
            <Badge variant="outline" className="text-[10px] text-slate-400 font-mono">
              Kubernetes / Prometheus Ready
            </Badge>
          </div>
          <CardDescription>
            Standard endpoints configured for enterprise Kubernetes liveness probes and Grafana dashboards.
          </CardDescription>
        </CardHeader>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 font-mono text-xs">
          <div className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800">
            <div>
              <span className="text-slate-300 block font-semibold">GET /healthz</span>
              <span className="text-slate-500 text-[11px]">HTTP Liveness Probe</span>
            </div>
            <Badge variant="success" className="text-[10px] py-0 font-mono">200 OK</Badge>
          </div>

          <div className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800">
            <div>
              <span className="text-slate-300 block font-semibold">GET /metrics</span>
              <span className="text-slate-500 text-[11px]">Prometheus Exposition</span>
            </div>
            <button
              onClick={() => handleCopy('http://localhost:8080/metrics', 'prom-url')}
              className="text-slate-400 hover:text-white transition-colors"
              title="Copy Prometheus URL"
            >
              {copiedAction === 'prom-url' ? (
                <Check className="w-4 h-4 text-emerald-400" />
              ) : (
                <Copy className="w-4 h-4" />
              )}
            </button>
          </div>

          <div className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800">
            <div>
              <span className="text-slate-300 block font-semibold">GET /v1/origin/openapi.yaml</span>
              <span className="text-slate-500 text-[11px]">OpenAPI 3.1 Contract</span>
            </div>
            <Badge variant="outline" className="text-[10px] text-sky-300 border-sky-500/40 py-0 font-mono">
              34 Routes
            </Badge>
          </div>
        </div>
      </Card>
    </div>
  )
}
