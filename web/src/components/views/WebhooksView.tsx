import React, { useState } from 'react'
import { Webhook, RefreshCw, CheckCircle2, ChevronDown, ChevronRight } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'

export const WebhooksView: React.FC = () => {
  const [expandedId, setExpandedId] = useState<string | null>('wh-1')
  const [redelivering, setRedelivering] = useState<string | null>(null)

  const [deliveries, setDeliveries] = useState([
    {
      id: 'wh-1',
      event: 'pull_request.opened',
      targetUrl: 'https://ci.internal.exist.io/hooks/forge',
      statusCode: 200,
      timestamp: '3 hours ago',
      latency: '34ms',
      payload: {
        action: 'opened',
        number: 42,
        repository: 'exist/exist.io',
        sender: 'claude-agent',
      },
    },
    {
      id: 'wh-2',
      event: 'check_run.completed',
      targetUrl: 'https://ci.internal.exist.io/hooks/forge',
      statusCode: 200,
      timestamp: '2 hours ago',
      latency: '22ms',
      payload: {
        check_name: 'Go Build & Test Matrix',
        conclusion: 'success',
        duration_s: 42,
      },
    },
    {
      id: 'wh-3',
      event: 'push',
      targetUrl: 'https://audit.corp.exist.io/log',
      statusCode: 204,
      timestamp: '1 hour ago',
      latency: '18ms',
      payload: {
        ref: 'refs/heads/main',
        after: '7aa844697bf92d4f58c707d8ceca47e1d882b524',
      },
    },
  ])

  const handleRedeliver = (id: string) => {
    setRedelivering(id)
    setTimeout(() => {
      setRedelivering(null)
      setDeliveries((prev) =>
        prev.map((d) => (d.id === id ? { ...d, timestamp: 'Just now' } : d))
      )
    }, 800)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="p-4 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-xl bg-sky-500/10 border border-sky-500/30 text-sky-400">
            <Webhook className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-white">Webhook Delivery Log</h3>
            <p className="text-xs text-slate-400">
              Audit and redeliver dispatched JSON webhooks to CI pipelines and audit services.
            </p>
          </div>
        </div>
        <Badge variant="outline" className="text-xs font-mono text-slate-300">
          3 Deliveries Active
        </Badge>
      </div>

      {/* Deliveries List */}
      <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/60">
        <div className="divide-y divide-slate-850/70">
          {deliveries.map((del) => {
            const isExpanded = expandedId === del.id
            return (
              <div key={del.id} className="p-4 hover:bg-slate-900/40 transition-colors space-y-3">
                <div className="flex items-center justify-between gap-4">
                  <div
                    onClick={() => setExpandedId(isExpanded ? null : del.id)}
                    className="flex items-center gap-3 cursor-pointer flex-1"
                  >
                    {isExpanded ? (
                      <ChevronDown className="w-4 h-4 text-slate-400" />
                    ) : (
                      <ChevronRight className="w-4 h-4 text-slate-400" />
                    )}
                    <span className="font-mono text-xs font-semibold text-sky-300 bg-slate-900 px-2 py-0.5 rounded border border-slate-800">
                      {del.event}
                    </span>
                    <span className="text-xs text-slate-300 font-mono hidden sm:inline truncate max-w-sm">
                      {del.targetUrl}
                    </span>
                  </div>

                  <div className="flex items-center gap-3">
                    <Badge variant="success" className="text-[10px] font-mono py-0">
                      <CheckCircle2 className="w-3 h-3 mr-1" /> HTTP {del.statusCode}
                    </Badge>
                    <span className="text-xs font-mono text-slate-500">{del.latency}</span>
                    <span className="text-[11px] text-slate-500">{del.timestamp}</span>

                    <Button
                      variant="ghost"
                      size="xs"
                      onClick={() => handleRedeliver(del.id)}
                      loading={redelivering === del.id}
                      className="gap-1 h-7 text-slate-300"
                    >
                      <RefreshCw className="w-3 h-3" />
                      <span className="hidden sm:inline">Redeliver</span>
                    </Button>
                  </div>
                </div>

                {isExpanded && (
                  <div className="p-3 bg-slate-950 rounded-xl border border-slate-850 font-mono text-xs text-sky-200 overflow-x-auto space-y-2">
                    <div className="text-slate-500 text-[11px] flex items-center justify-between">
                      <span>HTTP Request Payload (Signed via HMAC-SHA256)</span>
                      <span className="text-slate-400">Target: {del.targetUrl}</span>
                    </div>
                    <pre className="text-slate-300">{JSON.stringify(del.payload, null, 2)}</pre>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </Card>
    </div>
  )
}
