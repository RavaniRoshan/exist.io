import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ClusterHealth } from '../admin/ClusterHealth'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api')>('@/lib/api')
  return {
    ...actual,
    getClusterMetrics: vi.fn(),
  }
})

describe('ClusterHealth Admin View', () => {
  const mockMetrics: api.ClusterMetrics = {
    serverStatus: 'healthy',
    uptime: '420s',
    version: 'existd v0.4.5-enterprise',
    activeGitSessions: 2,
    sqliteWalSizeBytes: 1048576,
    databasePath: 'data/meta.db',
    goroutines: 12,
    allocatedMemoryMb: 14.5,
    cpuUsagePercent: 0.5,
    litestreamSyncStatus: 'active',
    lastBackupTimestamp: '2026-10-15T11:00:00Z',
    mcpMountedToolsCount: 35,
  }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.getClusterMetrics).mockResolvedValue(mockMetrics)
  })

  it('renders loading state then displays cluster telemetry metrics', async () => {
    render(<ClusterHealth />)

    expect(screen.getByText(/Loading cluster health telemetry/i)).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText('Primary Forge Leader')).toBeInTheDocument()
    })

    expect(screen.getByText(/existd v0.4.5-enterprise/i)).toBeInTheDocument()
    expect(screen.getByText(/System Operational/i)).toBeInTheDocument()
    expect(screen.getByText('1.0 MB')).toBeInTheDocument() // formatBytes(1048576)
    expect(screen.getByText('Continuous')).toBeInTheDocument()
    expect(screen.getByText('14.5')).toBeInTheDocument()
    expect(screen.getByText('MB RAM')).toBeInTheDocument()
  })

  it('refreshes telemetry when clicking Refresh button', async () => {
    render(<ClusterHealth />)

    await waitFor(() => {
      expect(screen.getByText('Primary Forge Leader')).toBeInTheDocument()
    })

    const refreshBtn = screen.getByRole('button', { name: /Refresh Telemetry/i })
    fireEvent.click(refreshBtn)

    expect(api.getClusterMetrics).toHaveBeenCalledTimes(2)
  })

  it('copies command snippet to clipboard', async () => {
    const writeTextMock = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    })

    render(<ClusterHealth />)

    await waitFor(() => {
      expect(screen.getByText('Primary Forge Leader')).toBeInTheDocument()
    })

    const copyBtns = screen.getAllByRole('button', { name: /Copy/i })
    if (copyBtns.length > 0) {
      fireEvent.click(copyBtns[0])
      expect(writeTextMock).toHaveBeenCalled()
    }
  })
})
