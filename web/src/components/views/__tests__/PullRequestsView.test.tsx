import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { PullRequestsView } from '../PullRequestsView'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    getPullRequests: vi.fn(),
    mergePullRequest: vi.fn(),
    triggerAgentReview: vi.fn(),
  }
})

describe('PullRequestsView Component', () => {
  const mockPRs = [
    {
      number: 104,
      title: 'feat: add capability lease protocol',
      state: 'open' as const,
      author: 'roshan',
      headRef: 'feature/agent-leases',
      baseRef: 'main',
      createdAt: '2 hours ago',
      comments: [
        { id: 'c1', author: 'roshan', isAgent: false, createdAt: '1 hour ago', body: 'Initial implementation.' },
      ],
      checks: [
        { id: 'chk1', name: 'Go Build & Test Matrix', status: 'completed' as const, conclusion: 'success' as const, duration: '4.2s' },
      ],
      labels: [{ name: 'enhancement', color: '#0284c7' }],
      diffHunks: [
        { oldPath: 'pkg/agentlease/lease.go', newPath: 'pkg/agentlease/lease.go', additions: 14, deletions: 2, hunk: '@@ -1,4 +1,16 @@\n+func New()' },
      ],
    },
    {
      number: 103,
      title: 'fix: handle sqlite wal sync edge-case',
      state: 'merged' as const,
      author: 'alex',
      headRef: 'bugfix/sqlite-wal',
      baseRef: 'main',
      createdAt: '1 day ago',
      comments: [],
      checks: [],
      labels: [{ name: 'bug', color: '#e11d48' }],
      diffHunks: [],
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.getPullRequests).mockResolvedValue(mockPRs)
  })

  it('loads and renders list of pull requests with filter badges', async () => {
    render(<PullRequestsView />)

    expect(await screen.findByText('feat: add capability lease protocol')).toBeInTheDocument()
    expect(screen.getByText(/1 Open/i)).toBeInTheDocument()
    expect(screen.getByText(/1 Closed/i)).toBeInTheDocument()
  })

  it('filters pull requests when clicking Closed tab', async () => {
    render(<PullRequestsView />)

    expect(await screen.findByText('feat: add capability lease protocol')).toBeInTheDocument()

    // Click on Closed filter
    fireEvent.click(screen.getByText(/1 Closed/i))

    // Closed PR should appear in filtered list
    expect(await screen.findByText('fix: handle sqlite wal sync edge-case')).toBeInTheDocument()
  })

  it('displays PR details with canonical two-column layout and UnderlineNav', async () => {
    render(<PullRequestsView />)

    const prRow = await screen.findByText('feat: add capability lease protocol')
    fireEvent.click(prRow)

    // Details header assertions
    expect(await screen.findByRole('heading', { level: 1, name: /feat: add capability lease protocol/i })).toBeInTheDocument()
    expect(screen.getByText(/#104/i)).toBeInTheDocument()
    expect(screen.getByText(/wants to merge 1 commit into/i)).toBeInTheDocument()

    // UnderlineNav tabs inside PR
    expect(screen.getByRole('button', { name: /conversation/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /checks/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /files changed/i })).toBeInTheDocument()

    // Right sidebar assertions
    expect(screen.getByText('Reviewers')).toBeInTheDocument()
    expect(screen.getByText('Assignees')).toBeInTheDocument()
    expect(screen.getByText('Ruleset Protection')).toBeInTheDocument()
  })
})
