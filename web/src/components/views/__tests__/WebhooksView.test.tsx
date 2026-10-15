import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { WebhooksView } from '../WebhooksView'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    fetchWebhookDeliveries: vi.fn().mockResolvedValue([
      {
        id: 'whd-1',
        event: 'push',
        targetUrl: 'https://ci.internal.exist.io/hooks/forge',
        statusCode: 200,
        status: 200,
        durationMs: 42,
        timestamp: 'Just now',
        payload: {
          ref: 'refs/heads/main',
          after: '5c92f55',
          repository: { name: 'rocket', owner: 'shiva' },
        },
      },
      {
        id: 'whd-2',
        event: 'pull_request',
        targetUrl: 'https://security-bot.exist.io/audit',
        statusCode: 200,
        status: 200,
        durationMs: 88,
        timestamp: '5m ago',
        payload: {
          action: 'opened',
          number: 1,
          pull_request: { title: 'Telemetry enhancement' },
        },
      },
    ]),
    pingWebhook: vi.fn().mockResolvedValue({
      ok: true,
      delivery_id: 'whd-ping-1',
    }),
    batchRedeliverWebhooks: vi.fn().mockResolvedValue(true),
  }
})

describe('WebhooksView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders webhook delivery log with event badges and URLs', async () => {
    render(<WebhooksView />)

    await waitFor(() => {
      expect(screen.getByText('Webhook Delivery Log')).toBeInTheDocument()
      expect(screen.getByText('push')).toBeInTheDocument()
      expect(screen.getByText('pull_request')).toBeInTheDocument()
      expect(screen.getByText('https://ci.internal.exist.io/hooks/forge')).toBeInTheDocument()
      expect(screen.getByText('https://security-bot.exist.io/audit')).toBeInTheDocument()
    })
    expect(screen.getAllByText('HTTP 200').length).toBe(2)
  })

  it('triggers test ping and shows success feedback badge', async () => {
    render(<WebhooksView />)

    await waitFor(() => {
      expect(screen.getByText('Send Test Ping')).toBeInTheDocument()
    })

    const pingBtn = screen.getByRole('button', { name: /send test ping/i })
    fireEvent.click(pingBtn)

    await waitFor(() => {
      expect(api.pingWebhook).toHaveBeenCalled()
      expect(screen.getByText('Test ping dispatched successfully')).toBeInTheDocument()
    })
  })

  it('redelivers webhook payload upon clicking Redeliver button', async () => {
    render(<WebhooksView />)

    await waitFor(() => {
      expect(screen.getByText('push')).toBeInTheDocument()
    })

    const redeliverBtns = screen.getAllByRole('button', { name: /redeliver/i })
    fireEvent.click(redeliverBtns[0])

    await waitFor(() => {
      expect(api.batchRedeliverWebhooks).toHaveBeenCalledWith(['whd-1'])
      expect(screen.getByText('Webhook redelivery queued')).toBeInTheDocument()
    })
  })
})
