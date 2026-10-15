import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { CreatePullRequestModal } from '../CreatePullRequestModal'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    getBranches: vi.fn(),
    createPullRequest: vi.fn(),
  }
})

describe('CreatePullRequestModal Component', () => {
  const mockBranches = [
    { name: 'main', commitSha: '54926f8', isDefault: true },
    { name: 'feature/telemetry', commitSha: 'a1b2c3d', isDefault: false },
  ]

  const mockRepo = {
    ownerSlug: 'acme',
    name: 'rocket',
    defaultBranch: 'main',
    description: 'Acme Rocket',
    isPrivate: false,
    starsCount: 10,
    forksCount: 2,
    updatedAt: 'Recently',
    cloneUrlHttp: 'http://localhost:8080/acme/rocket.git',
    cloneUrlSsh: 'git@localhost:acme/rocket.git',
  }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.getBranches).mockResolvedValue(mockBranches)
    vi.mocked(api.createPullRequest).mockResolvedValue({
      number: 1,
      title: 'feat: add telemetry streaming',
      body: 'Streams live engine metrics.',
      state: 'open',
      author: 'shiva',
      headRef: 'feature/telemetry',
      baseRef: 'main',
      createdAt: 'Just now',
      labels: [],
      diffStats: { additions: 25, deletions: 0, filesChanged: 1 },
      checks: [],
      comments: [],
      mergeable: true,
      diffHunks: [],
    })
  })

  it('renders modal with branch comparison bar and inputs', async () => {
    render(<CreatePullRequestModal isOpen={true} onClose={vi.fn()} repo={mockRepo} />)

    expect(screen.getByText('Open a pull request')).toBeInTheDocument()
    expect(screen.getByText(/Comparing changes across branches/i)).toBeInTheDocument()
    expect(screen.getByPlaceholderText(/e\.g\. feat: implement autonomous agent review gate/i)).toBeInTheDocument()

    // Base branch should show 'main'
    await waitFor(() => {
      expect(screen.getByDisplayValue('main')).toBeInTheDocument()
    })
  })

  it('validates that head and base branch cannot be identical', async () => {
    render(<CreatePullRequestModal isOpen={true} onClose={vi.fn()} repo={mockRepo} />)

    await waitFor(() => {
      expect(screen.getByDisplayValue('main')).toBeInTheDocument()
    })

    // Type a title
    const titleInput = screen.getByPlaceholderText(/e\.g\. feat: implement autonomous agent review gate/i)
    fireEvent.change(titleInput, { target: { value: 'Test PR' } })

    // Change compare branch to 'main'
    const selects = screen.getAllByRole('combobox')
    fireEvent.change(selects[1], { target: { value: 'main' } })

    // Submit button should be disabled when head === base
    const submitBtn = screen.getByRole('button', { name: /create pull request/i })
    expect(submitBtn).toBeDisabled()
  })

  it('submits valid pull request and triggers onPullRequestCreated callback', async () => {
    const handleCreated = vi.fn()
    const handleClose = vi.fn()

    render(
      <CreatePullRequestModal
        isOpen={true}
        onClose={handleClose}
        repo={mockRepo}
        onPullRequestCreated={handleCreated}
      />
    )

    await waitFor(() => {
      expect(screen.getByDisplayValue('main')).toBeInTheDocument()
    })

    const titleInput = screen.getByPlaceholderText(/e\.g\. feat: implement autonomous agent review gate/i)
    fireEvent.change(titleInput, { target: { value: 'feat: add telemetry streaming' } })

    const descInput = screen.getByPlaceholderText(/Explain the changes in this PR/i)
    fireEvent.change(descInput, { target: { value: 'Streams live engine metrics.' } })

    const submitBtn = screen.getByRole('button', { name: /create pull request/i })
    expect(submitBtn).not.toBeDisabled()
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(api.createPullRequest).toHaveBeenCalledWith('acme', 'rocket', {
        title: 'feat: add telemetry streaming',
        head: 'feature/telemetry',
        base: 'main',
        body: 'Streams live engine metrics.',
        draft: false,
      })
      expect(handleCreated).toHaveBeenCalled()
      expect(handleClose).toHaveBeenCalled()
    })
  })
})
