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
import { Box, BoxHeader, BoxBody, BoxRow } from '@/components/ui/Card'
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
    <div className="space-y-4">
      {/* Header Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        <div className="space-y-0.5">
          <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
            <Webhook className="w-4 h-4 text-primary" />
            <span>Webhook Delivery Log</span>
          </h2>
          <p className="text-xs text-muted-foreground">
            Audit and redeliver dispatched JSON webhooks to CI pipelines, audit collectors, and external bots.
          </p>
        </div>

        <div className="flex items-center gap-2 self-start sm:self-auto">
          {feedbackMsg && (
            <Badge variant="outline" className="text-xs font-mono text-emerald-500 border-emerald-500/40 gap-1">
              <Check className="w-3 h-3" />
              <span>{feedbackMsg}</span>
            </Badge>
          )}

          <Button
            variant="outline"
            size="sm"
            onClick={handlePing}
            disabled={pinging}
            className="text-xs gap-1.5"
          >
            <Send className="w-3.5 h-3.5" />
            <span>Send Test Ping</span>
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={() => loadDeliveries(true)}
            disabled={refreshing}
            className="text-xs gap-1.5"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${refreshing ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </Button>
        </div>
      </div>

      {loading ? (
        <Box className="p-8 text-center">
          <div className="flex items-center justify-center gap-2 text-xs text-muted-foreground font-mono">
            <RefreshCw className="w-4 h-4 animate-spin text-primary" />
            <span>Loading webhook deliveries from daemon...</span>
          </div>
        </Box>
      ) : deliveries.length === 0 ? (
        /* Empty State */
        <Box className="p-8 text-center space-y-4">
          <div className="w-10 h-10 rounded-md bg-secondary border border-border text-muted-foreground flex items-center justify-center mx-auto">
            <Webhook className="w-5 h-5" />
          </div>
          <div className="space-y-1 max-w-md mx-auto">
            <h4 className="text-sm font-semibold text-foreground">No Webhook Deliveries Logged</h4>
            <p className="text-xs text-muted-foreground leading-relaxed">
              When Git pushes, pull requests, or CI checks occur, automated JSON webhooks are signed with HMAC-SHA256 and dispatched to external endpoints.
            </p>
          </div>
          <Button variant="secondary" size="sm" onClick={handlePing} disabled={pinging} className="gap-1.5 text-xs">
            <Send className="w-3.5 h-3.5" />
            <span>Trigger Test Ping Event</span>
          </Button>
        </Box>
      ) : (
        /* Deliveries List Box */
        <Box>
          <BoxHeader className="bg-secondary/35 py-2.5">
            <span className="font-semibold text-foreground text-xs">Recent Deliveries</span>
            <span className="text-muted-foreground font-mono text-[11px]">{deliveries.length} records</span>
          </BoxHeader>

          <BoxBody className="p-0">
            {deliveries.map((del) => {
              const isExpanded = expandedId === del.id
              const status = del.statusCode || del.status || 200
              const isOk = status >= 200 && status < 300

              return (
                <div key={del.id} className="border-b border-border last:border-b-0">
                  <BoxRow className="p-3.5 flex items-center justify-between gap-4 hover:bg-secondary/20 transition-colors">
                    <div
                      onClick={() => setExpandedId(isExpanded ? null : del.id)}
                      className="flex items-center gap-3 cursor-pointer flex-1"
                    >
                      {isExpanded ? (
                        <ChevronDown className="w-4 h-4 text-muted-foreground shrink-0" />
                      ) : (
                        <ChevronRight className="w-4 h-4 text-muted-foreground shrink-0" />
                      )}
                      <span className="font-mono text-xs font-semibold text-foreground bg-secondary px-2 py-0.5 rounded border border-border shrink-0">
                        {del.event || 'push'}
                      </span>
                      <span className="text-xs text-foreground font-mono hidden sm:inline truncate max-w-sm">
                        {del.targetUrl || del.url || 'https://ci.internal.exist.io/hooks/forge'}
                      </span>
                    </div>

                    <div className="flex items-center gap-3">
                      <Badge variant={isOk ? 'success' : 'closed'} className="text-[10px] font-mono py-0 gap-1">
                        {isOk ? <CheckCircle2 className="w-3 h-3 text-emerald-500" /> : <AlertCircle className="w-3 h-3 text-destructive" />}
                        <span>HTTP {status}</span>
                      </Badge>

                      <span className="text-xs font-mono text-muted-foreground hidden md:inline">
                        {del.latency || `${del.durationMs || 32}ms`}
                      </span>

                      <span className="text-[11px] text-muted-foreground">
                        {del.timestamp || 'Recently'}
                      </span>

                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRedeliver(del.id)}
                        disabled={redeliveringId === del.id}
                        className="gap-1 h-7 px-2 text-xs"
                        title="Redeliver webhook payload"
                      >
                        <RefreshCw className={`w-3 h-3 ${redeliveringId === del.id ? 'animate-spin' : ''}`} />
                        <span className="hidden sm:inline">Redeliver</span>
                      </Button>
                    </div>
                  </BoxRow>

                  {isExpanded && (
                    <div className="p-3 bg-secondary/20 border-t border-border font-mono text-xs space-y-2">
                      <div className="text-muted-foreground text-[11px] flex items-center justify-between">
                        <span>HTTP Request Payload (Signed via HMAC-SHA256)</span>
                        <button
                          onClick={() => handleCopy(JSON.stringify(del.payload || del.request?.body || {}, null, 2), del.id)}
                          className="flex items-center gap-1 text-primary hover:underline transition-colors"
                        >
                          {copiedId === del.id ? (
                            <Check className="w-3 h-3 text-emerald-500" />
                          ) : (
                            <Copy className="w-3 h-3" />
                          )}
                          <span>Copy JSON</span>
                        </button>
                      </div>
                      <pre className="text-foreground overflow-x-auto p-2.5 bg-background rounded border border-border text-[11px] max-h-56">
                        {JSON.stringify(del.payload || del.request?.body || { event: del.event, id: del.id }, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>
              )
            })}
          </BoxBody>
        </Box>
      )}
    </div>
  )
}
