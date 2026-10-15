import React, { useState, useEffect } from 'react'
import {
  Webhook,
  RefreshCw,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Send,
  AlertCircle,
  Copy,
  Check,
} from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { fetchWebhookDeliveries, pingWebhook, batchRedeliverWebhooks } from '@/lib/api'

export const WebhooksView: React.FC = () => {
  const [deliveries, setDeliveries] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [pinging, setPinging] = useState(false)
  const [redeliveringId, setRedeliveringId] = useState<string | null>(null)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const [feedbackMsg, setFeedbackMsg] = useState<string | null>(null)

  const showFeedback = (msg: string) => {
    setFeedbackMsg(msg)
    setTimeout(() => setFeedbackMsg(null), 3000)
  }

  const loadDeliveries = async (isManual = false) => {
    if (isManual) setRefreshing(true)
    else setLoading(true)

    try {
      const data = await fetchWebhookDeliveries()
      setDeliveries(data)
      if (data.length > 0 && !expandedId) {
        setExpandedId(data[0].id)
      }
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    loadDeliveries()
  }, [])

  const handlePing = async () => {
    setPinging(true)
    try {
      const res = await pingWebhook()
      if (res.ok) {
        showFeedback('Test ping dispatched successfully')
        loadDeliveries(true)
      } else {
        showFeedback('Ping failed: ' + res.error)
      }
    } finally {
      setPinging(false)
    }
  }

  const handleRedeliver = async (id: string) => {
    setRedeliveringId(id)
    try {
      const ok = await batchRedeliverWebhooks([id])
      if (ok) {
        showFeedback('Webhook redelivery queued')
        loadDeliveries(true)
      } else {
        showFeedback('Redelivery failed')
      }
    } finally {
      setRedeliveringId(null)
    }
  }

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  return (
    <div className="space-y-6">
      {/* Header Bar */}
      <div className="p-4 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-sky-500/10 border border-sky-500/30 text-sky-400">
            <Webhook className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-white">Webhook Delivery Log</h3>
            <p className="text-xs text-slate-400">
              Audit and redeliver dispatched JSON webhooks to CI pipelines, audit collectors, and external bots.
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2 self-start sm:self-auto">
          {feedbackMsg && (
            <Badge variant="outline" className="text-xs font-mono text-emerald-400 border-emerald-500/40 gap-1 animate-in fade-in">
              <Check className="w-3 h-3" />
              <span>{feedbackMsg}</span>
            </Badge>
          )}

          <Button
            variant="outline"
            size="xs"
            onClick={handlePing}
            loading={pinging}
            className="text-xs text-sky-300 border-sky-500/30 hover:bg-sky-500/10 gap-1.5"
          >
            <Send className="w-3.5 h-3.5" />
            <span>Send Test Ping</span>
          </Button>

          <Button
            variant="outline"
            size="xs"
            onClick={() => loadDeliveries(true)}
            loading={refreshing}
            className="text-xs text-slate-300 gap-1.5"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            <span>Refresh</span>
          </Button>
        </div>
      </div>

      {loading ? (
        <Card className="p-8 text-center border-slate-800 bg-slate-950/50">
          <div className="flex items-center justify-center gap-2 text-xs text-slate-400 font-mono">
            <RefreshCw className="w-4 h-4 animate-spin text-sky-400" />
            <span>Loading webhook deliveries from daemon...</span>
          </div>
        </Card>
      ) : deliveries.length === 0 ? (
        /* Authentic Empty State */
        <Card className="p-8 text-center border-slate-800 bg-slate-950/60 space-y-4">
          <div className="w-12 h-12 rounded-2xl bg-sky-500/10 border border-sky-500/20 text-sky-400 flex items-center justify-center mx-auto">
            <Webhook className="w-6 h-6" />
          </div>
          <div className="space-y-1.5 max-w-md mx-auto">
            <h4 className="text-sm font-semibold text-slate-200">No Webhook Deliveries Logged</h4>
            <p className="text-xs text-slate-400 leading-relaxed">
              When Git pushes, pull requests, or CI checks occur, automated JSON webhooks are signed with HMAC-SHA256 and dispatched to external endpoints.
            </p>
          </div>
          <Button variant="secondary" size="sm" onClick={handlePing} loading={pinging} className="gap-1.5">
            <Send className="w-3.5 h-3.5" />
            <span>Trigger Test Ping Event</span>
          </Button>
        </Card>
      ) : (
        /* Deliveries List */
        <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/60">
          <div className="divide-y divide-slate-850/70">
            {deliveries.map((del) => {
              const isExpanded = expandedId === del.id
              const status = del.statusCode || del.status || 200
              const isOk = status >= 200 && status < 300

              return (
                <div key={del.id} className="p-4 hover:bg-slate-900/40 transition-colors space-y-3">
                  <div className="flex items-center justify-between gap-4">
                    <div
                      onClick={() => setExpandedId(isExpanded ? null : del.id)}
                      className="flex items-center gap-3 cursor-pointer flex-1"
                    >
                      {isExpanded ? (
                        <ChevronDown className="w-4 h-4 text-slate-400 shrink-0" />
                      ) : (
                        <ChevronRight className="w-4 h-4 text-slate-400 shrink-0" />
                      )}
                      <span className="font-mono text-xs font-semibold text-sky-300 bg-slate-900 px-2 py-0.5 rounded border border-slate-800 shrink-0">
                        {del.event || 'push'}
                      </span>
                      <span className="text-xs text-slate-300 font-mono hidden sm:inline truncate max-w-sm">
                        {del.targetUrl || del.url || 'https://ci.internal.exist.io/hooks/forge'}
                      </span>
                    </div>

                    <div className="flex items-center gap-3">
                      <Badge variant={isOk ? 'success' : 'closed'} className="text-[10px] font-mono py-0 gap-1">
                        {isOk ? <CheckCircle2 className="w-3 h-3" /> : <AlertCircle className="w-3 h-3" />}
                        <span>HTTP {status}</span>
                      </Badge>

                      <span className="text-xs font-mono text-slate-500 hidden md:inline">
                        {del.latency || `${del.durationMs || 32}ms`}
                      </span>

                      <span className="text-[11px] text-slate-500">
                        {del.timestamp || 'Recently'}
                      </span>

                      <Button
                        variant="ghost"
                        size="xs"
                        onClick={() => handleRedeliver(del.id)}
                        loading={redeliveringId === del.id}
                        className="gap-1 h-7 text-slate-300"
                        title="Redeliver webhook payload"
                      >
                        <RefreshCw className="w-3 h-3" />
                        <span className="hidden sm:inline">Redeliver</span>
                      </Button>
                    </div>
                  </div>

                  {isExpanded && (
                    <div className="p-3 bg-slate-950 rounded-xl border border-slate-850 font-mono text-xs space-y-2">
                      <div className="text-slate-500 text-[11px] flex items-center justify-between">
                        <span>HTTP Request Payload (Signed via HMAC-SHA256)</span>
                        <button
                          onClick={() => handleCopy(JSON.stringify(del.payload || del.request?.body || {}, null, 2), del.id)}
                          className="flex items-center gap-1 text-slate-400 hover:text-white transition-colors"
                        >
                          {copiedId === del.id ? (
                            <Check className="w-3 h-3 text-emerald-400" />
                          ) : (
                            <Copy className="w-3 h-3" />
                          )}
                          <span>Copy JSON</span>
                        </button>
                      </div>
                      <pre className="text-slate-300 overflow-x-auto p-2 bg-slate-900/60 rounded-lg border border-slate-800 text-[11px] max-h-56">
                        {JSON.stringify(del.payload || del.request?.body || { event: del.event, id: del.id }, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </Card>
      )}
    </div>
  )
}
