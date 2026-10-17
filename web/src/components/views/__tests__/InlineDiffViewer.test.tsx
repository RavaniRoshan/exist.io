import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { InlineDiffViewer } from '../InlineDiffViewer'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    applyCodeSuggestion: vi.fn(),
    toggleThreadResolved: vi.fn(),
    addPullRequestComment: vi.fn(),
    submitPullRequestReview: vi.fn(),
  }
})

describe('InlineDiffViewer Component', () => {
  const mockHunks = [
    {
      oldPath: 'pkg/lease/lease.go',
      newPath: 'pkg/lease/lease.go',
      additions: 2,
      deletions: 1,
      hunk: '@@ -10,3 +10,4 @@\n-oldLine()\n+newLine()\n+anotherLine()',
    },
  ]

  const mockComments = [
    {
      id: 'c-sugg-1',
      author: 'alex',
      isAgent: false,
      createdAt: '10m ago',
      body: 'Consider optimizing this lease calculation:\n```suggestion\n+optimizedLine()\n```',
      path: 'pkg/lease/lease.go',
      line: 11,
      side: 'RIGHT',
      threadId: 'th-1',
      resolved: false,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.applyCodeSuggestion).mockResolvedValue({ ok: true, commitSha: 'abc1234' })
    vi.mocked(api.toggleThreadResolved).mockResolvedValue({ ok: true })
    vi.mocked(api.addPullRequestComment).mockResolvedValue({ id: 'c-new', body: 'Great change' })
    vi.mocked(api.submitPullRequestReview).mockResolvedValue({ ok: true })
  })

  it('renders diff lines and suggestion card with Apply button', () => {
    render(
      <InlineDiffViewer
        owner="shiva"
        repoName="exist.io"
        prNumber={101}
        diffHunks={mockHunks}
        comments={mockComments}
        onRefreshComments={() => {}}
      />
    )

    expect(screen.getByText('pkg/lease/lease.go')).toBeInTheDocument()
    expect(screen.getByText('Suggested change')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /apply suggestion/i })).toBeInTheDocument()
  })

  it('applies suggestion when clicking Apply suggestion button', async () => {
    const handleRefresh = vi.fn()
    render(
      <InlineDiffViewer
        owner="shiva"
        repoName="exist.io"
        prNumber={101}
        diffHunks={mockHunks}
        comments={mockComments}
        onRefreshComments={handleRefresh}
      />
    )

    const applyBtn = screen.getByRole('button', { name: /apply suggestion/i })
    fireEvent.click(applyBtn)

    await waitFor(() => {
      expect(api.applyCodeSuggestion).toHaveBeenCalledWith('shiva', 'exist.io', 101, 'c-sugg-1')
      expect(screen.getByText(/suggestion successfully committed/i)).toBeInTheDocument()
      expect(handleRefresh).toHaveBeenCalled()
    })
  })

  it('toggles thread resolution', async () => {
    const handleRefresh = vi.fn()
    render(
      <InlineDiffViewer
        owner="shiva"
        repoName="exist.io"
        prNumber={101}
        diffHunks={mockHunks}
        comments={mockComments}
        onRefreshComments={handleRefresh}
      />
    )

    const resolveBtn = screen.getByRole('button', { name: /resolve conversation/i })
    fireEvent.click(resolveBtn)

    await waitFor(() => {
      expect(api.toggleThreadResolved).toHaveBeenCalledWith('shiva', 'exist.io', 101, 'th-1', true)
      expect(handleRefresh).toHaveBeenCalled()
    })
  })

  it('submits a batch review with Approve verdict', async () => {
    const handleRefresh = vi.fn()
    render(
      <InlineDiffViewer
        owner="shiva"
        repoName="exist.io"
        prNumber={101}
        diffHunks={mockHunks}
        comments={mockComments}
        onRefreshComments={handleRefresh}
      />
    )

    const reviewChangesBtn = screen.getByRole('button', { name: /review changes/i })
    fireEvent.click(reviewChangesBtn)

    expect(screen.getByTestId('review-changes-modal')).toBeInTheDocument()

    // Select Approve radio
    const approveRadio = screen.getByRole('radio', { name: /approve/i })
    fireEvent.click(approveRadio)

    // Fill summary comment
    const summaryInput = screen.getByPlaceholderText(/summarizing your review/i)
    fireEvent.change(summaryInput, { target: { value: 'LGTM! Great work.' } })

    // Submit review
    const submitBtn = screen.getByRole('button', { name: /submit review/i })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(api.submitPullRequestReview).toHaveBeenCalledWith('shiva', 'exist.io', 101, {
        verdict: 'approve',
        body: 'LGTM! Great work.',
      })
      expect(screen.getByText(/review submitted with verdict: APPROVE/i)).toBeInTheDocument()
      expect(handleRefresh).toHaveBeenCalled()
    })
  })

  it('opens inline comment composer and inserts suggestion block', async () => {
    render(
      <InlineDiffViewer
        owner="shiva"
        repoName="exist.io"
        prNumber={101}
        diffHunks={mockHunks}
        comments={[]}
        onRefreshComments={() => {}}
      />
    )

    // Click hover plus button
    const plusButtons = screen.getAllByTitle('Add inline comment')
    fireEvent.click(plusButtons[0])

    expect(screen.getByTestId('inline-comment-composer')).toBeInTheDocument()

    // Click "Add a suggestion"
    const addSuggBtn = screen.getByRole('button', { name: /add a suggestion/i })
    fireEvent.click(addSuggBtn)

    const textarea = screen.getByPlaceholderText(/leave a comment or review suggestion/i) as HTMLTextAreaElement
    expect(textarea.value).toContain('```suggestion')
  })
})
