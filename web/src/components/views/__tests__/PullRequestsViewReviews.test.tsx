import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { PullRequestsView } from '../PullRequestsView'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    getPullRequests: vi.fn(),
    getPullRequestComments: vi.fn(),
    getAutoMergeStatus: vi.fn(),
    setAutoMerge: vi.fn(),
    applyCodeSuggestion: vi.fn(),
    submitPullRequestReview: vi.fn(),
    toggleThreadResolved: vi.fn(),
  }
})

describe('PullRequestsView Code Review & AutoMerge Integration', () => {
  const mockPR = {
    number: 105,
    title: 'feat: add distributed capability lease validation',
    state: 'open' as const,
    author: 'shiva',
    headRef: 'feature/agent-leases',
    baseRef: 'main',
    createdAt: '1 hour ago',
    diffStats: { additions: 15, deletions: 2, filesChanged: 1 },
    checks: [
      { id: 'chk-1', name: 'Go Build & Test Matrix', status: 'completed' as const, conclusion: 'success' as const, duration: '4.2s' },
    ],
    comments: [
      {
        id: 'c-rev-1',
        author: 'alex',
        isAgent: false,
        createdAt: '30m ago',
        body: 'Consider optimizing this lease calculation:\n```suggestion\n+return calculateFastLease()\n```',
        path: 'pkg/agentlease/lease.go',
        line: 21,
        side: 'RIGHT',
        threadId: 'th-pr-1',
        resolved: false,
      },
    ],
    labels: [{ name: 'enhancement', color: '#0284c7' }],
    diffHunks: [
      {
        oldPath: 'pkg/agentlease/lease.go',
        newPath: 'pkg/agentlease/lease.go',
        additions: 15,
        deletions: 2,
        hunk: '@@ -20,5 +20,6 @@\n func New() {\n-return calculateSlowLease()\n+return calculateFastLease()\n }',
      },
    ],
  }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.getPullRequests).mockResolvedValue([mockPR])
    vi.mocked(api.getPullRequestComments).mockResolvedValue(mockPR.comments)
    vi.mocked(api.getAutoMergeStatus).mockResolvedValue({ ok: true, enabled: false })
    vi.mocked(api.setAutoMerge).mockResolvedValue({ ok: true, enabled: true })
    vi.mocked(api.applyCodeSuggestion).mockResolvedValue({ ok: true, commitSha: 'beefcafe1234' })
  })

  it('navigates to Files Changed tab, displays inline diff and applies suggestion', async () => {
    render(<PullRequestsView />)

    // Click PR from list
    const prItem = await screen.findByText('feat: add distributed capability lease validation')
    fireEvent.click(prItem)

    // Verify detail header
    expect(await screen.findByRole('heading', { level: 1, name: /feat: add distributed capability lease validation/i })).toBeInTheDocument()

    // Switch to Files changed tab
    const filesTab = screen.getByRole('button', { name: /files changed/i })
    fireEvent.click(filesTab)

    // Check diff hunk rendered
    expect(await screen.findByText('pkg/agentlease/lease.go')).toBeInTheDocument()
    expect(screen.getByText('Suggested change')).toBeInTheDocument()

    // Click Apply suggestion
    const applyBtn = screen.getByRole('button', { name: /apply suggestion/i })
    fireEvent.click(applyBtn)

    await waitFor(() => {
      expect(api.applyCodeSuggestion).toHaveBeenCalledWith('acme', 'rocket', 105, 'c-rev-1')
    })
  })

  it('toggles auto-merge in MergeBox', async () => {
    render(<PullRequestsView />)

    // Click PR from list
    const prItem = await screen.findByText('feat: add distributed capability lease validation')
    fireEvent.click(prItem)

    // Find and click Enable auto-merge
    const autoMergeBtn = await screen.findByRole('button', { name: /enable auto-merge/i })
    fireEvent.click(autoMergeBtn)

    await waitFor(() => {
      expect(api.setAutoMerge).toHaveBeenCalledWith('acme', 'rocket', 105, true, 'merge')
    })
  })
})
