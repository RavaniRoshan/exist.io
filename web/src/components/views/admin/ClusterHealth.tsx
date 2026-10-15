import React, { useState, useEffect } from 'react'
import { Server, Activity, Database, CloudCheck, HardDrive, Cpu, RefreshCw, CheckCircle2 } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { getClusterMetrics, type ClusterMetrics } from '@/lib/api'

export const ClusterHealth: React.FC = () => {
  const [metrics, setMetrics] = useState<ClusterMetrics | null>(null)
  const [refreshing, setRefreshing] = useState(false)

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
              <h3 className="text-base font-semibold text-white">Cluster Node: Primary Master</h3>
              <Badge variant="success" pulse>
                <CheckCircle2 className="w-3 h-3 mr-1" /> System Operational
              </Badge>
            </div>
            <p className="text-xs text-slate-400 font-mono mt-0.5">
              {metrics.version} • Uptime: {metrics.uptime}
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

      {/* Operator Metric Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Metric 1: Database WAL State */}
        <Card className="border-slate-800 bg-slate-950/60 p-4 space-y-2">
          <div className="flex items-center justify-between text-slate-400 text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <Database className="w-4 h-4 text-sky-400" /> SQLite WAL Size
            </span>
            <span className="text-[10px] bg-slate-900 px-1.5 py-0.5 rounded border border-slate-800 font-mono">
              WAL MODE
            </span>
          </div>
          <div className="text-2xl font-bold font-mono text-white tracking-tight">
            {formatBytes(metrics.sqliteWalSizeBytes)}
          </div>
          <p className="text-[11px] text-slate-500">
            Zero-lock pure Go SQLite engine with automatic checkpointing.
          </p>
        </Card>

        {/* Metric 2: Litestream Continuous Sync */}
        <Card className="border-slate-800 bg-slate-950/60 p-4 space-y-2">
          <div className="flex items-center justify-between text-slate-400 text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <CloudCheck className="w-4 h-4 text-purple-400" /> S3 Replication
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

        {/* Metric 3: Active Git Sessions */}
        <Card className="border-slate-800 bg-slate-950/60 p-4 space-y-2">
          <div className="flex items-center justify-between text-slate-400 text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <HardDrive className="w-4 h-4 text-emerald-400" /> Git Smart-HTTP
            </span>
            <span className="text-[10px] text-emerald-400 bg-emerald-950/40 px-1.5 py-0.5 rounded border border-emerald-800 font-mono">
              Pure-Go v6
            </span>
          </div>
          <div className="text-2xl font-bold font-mono text-white tracking-tight">
            {metrics.activeGitSessions} <span className="text-xs font-normal text-slate-400">concurrent</span>
          </div>
          <p className="text-[11px] text-slate-500">
            Authenticated packfile streaming over HTTP and SSH.
          </p>
        </Card>

        {/* Metric 4: Runtime Load & Goroutines */}
        <Card className="border-slate-800 bg-slate-950/60 p-4 space-y-2">
          <div className="flex items-center justify-between text-slate-400 text-xs">
            <span className="flex items-center gap-1.5 font-medium">
              <Cpu className="w-4 h-4 text-amber-400" /> Process Load
            </span>
            <span className="text-[10px] text-slate-400 font-mono">
              {metrics.goroutines} goroutines
            </span>
          </div>
          <div className="text-2xl font-bold font-mono text-white tracking-tight">
            {metrics.allocatedMemoryMb} <span className="text-xs font-normal text-slate-400">MB RAM</span>
          </div>
          <p className="text-[11px] text-slate-500">
            CPU: {metrics.cpuUsagePercent}% • Hermetic memory footprint.
          </p>
        </Card>
      </div>

      {/* Storage & Observability Subsystems */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Storage Subsystem Details */}
        <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
          <CardHeader className="p-0">
            <CardTitle className="text-sm text-white flex items-center gap-2">
              <Database className="w-4 h-4 text-sky-400" />
              <span>Storage Engine & Integrity</span>
            </CardTitle>
            <CardDescription>
              Underlying database architecture and repository filesystem paths.
            </CardDescription>
          </CardHeader>

          <div className="space-y-3 font-mono text-xs">
            <div className="p-3 rounded-xl bg-slate-900/60 border border-slate-800 space-y-1">
              <span className="text-slate-500 text-[11px] block">Metadata Store:</span>
              <span className="text-slate-200 break-all">{metrics.databasePath}</span>
            </div>

            <div className="p-3 rounded-xl bg-slate-900/60 border border-slate-800 space-y-1">
              <span className="text-slate-500 text-[11px] block">Disaster Recovery (DR) Destination:</span>
              <span className="text-purple-300">s3://exist-enterprise-backups/clusters/primary-node/</span>
            </div>
          </div>
        </Card>

        {/* Observability & Prometheus Probes */}
        <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
          <CardHeader className="p-0">
            <CardTitle className="text-sm text-white flex items-center gap-2">
              <Activity className="w-4 h-4 text-emerald-400" />
              <span>Prometheus & Health Scrape Probes</span>
            </CardTitle>
            <CardDescription>
              Standard endpoints for enterprise Kubernetes and Grafana monitoring.
            </CardDescription>
          </CardHeader>

          <div className="space-y-2.5 font-mono text-xs">
            <div className="flex items-center justify-between p-2.5 rounded-xl bg-slate-900/60 border border-slate-800">
              <span className="text-slate-300">GET /healthz</span>
              <Badge variant="success" className="text-[10px] py-0 font-mono">200 OK</Badge>
            </div>
            <div className="flex items-center justify-between p-2.5 rounded-xl bg-slate-900/60 border border-slate-800">
              <span className="text-slate-300">GET /v1/origin/openapi.yaml</span>
              <Badge variant="outline" className="text-[10px] py-0 font-mono text-sky-300 border-sky-500/40">OpenAPI 3.1</Badge>
            </div>
            <div className="flex items-center justify-between p-2.5 rounded-xl bg-slate-900/60 border border-slate-800">
              <span className="text-slate-300">POST /mcp</span>
              <Badge variant="agent" className="text-[10px] py-0 font-mono">34 Tools Live</Badge>
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}
