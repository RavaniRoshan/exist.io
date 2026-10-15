import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { IssuesView } from '../IssuesView'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    getIssues: vi.fn().mockResolvedValue([
      {
        id: 'issue-1',
        number: 1,
        title: 'Support WebAuthn credentials',
        body: 'Implement FIDO2 challenge-response.',
        state: 'open',
        author: 'Roshan Ravani',
        assignees: ['shiva'],
        labels: [{ name: 'security', color: '#b60205' }],
        createdAt: '2 days ago',
        comments: [
          {
            id: 'ic-1',
            author: 'Exist Agent',
            isAgent: true,
            body: 'Inspected auth handlers. Ready to implement.',
            createdAt: '1 day ago',
          },
        ],
      },
      {
        id: 'issue-2',
        number: 2,
        title: 'Optimize git-upload-pack packfile negotiation',
        body: 'Improve large monorepo fetch times.',
        state: 'closed',
        author: 'dev-lead',
        assignees: [],
        labels: [{ name: 'performance', color: '#d4c5f9' }],
        createdAt: '3 days ago',
        closedAt: '1 day ago',
        comments: [],
      },
    ]),
    createIssue: vi.fn().mockImplementation((owner, repo, title, body, labels) =>
      Promise.resolve({
        id: 'issue-3',
        number: 3,
        title,
        body,
        state: 'open',
        author: 'Roshan Ravani',
        assignees: [],
        labels: labels.map((l: string) => ({ name: l, color: '#1d76db' })),
        createdAt: 'Just now',
        comments: [],
      })
    ),
    addIssueComment: vi.fn().mockImplementation((owner, repo, num, body) =>
      Promise.resolve({
        id: 'ic-new',
        author: 'Roshan Ravani',
        isAgent: false,
        body,
        createdAt: 'Just now',
      })
    ),
    toggleIssueState: vi.fn().mockImplementation((owner, repo, num, currentState) =>
      Promise.resolve(currentState === 'open' ? 'closed' : 'open')
    ),
    delegateIssueToAgent: vi.fn().mockImplementation((owner, repo, num) =>
      Promise.resolve({
        ok: true,
        comment: {
          id: 'ic-agent-123',
          author: 'Exist Agent (Autonomous)',
          isAgent: true,
          body: '🤖 Autonomous Triage: Draft pull request created.',
          createdAt: 'Just now',
        },
        prNumber: num + 3,
      })
    ),
  }
})

describe('IssuesView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders issues list and allows filtering by open and closed status', async () => {
    render(<IssuesView />)

    await waitFor(() => {
      expect(screen.getByText('Support WebAuthn credentials')).toBeInTheDocument()
    })
    expect(screen.getByText('1 Open')).toBeInTheDocument()
    expect(screen.getByText('1 Closed')).toBeInTheDocument()

    // Switch to Closed filter
    const closedTab = screen.getByRole('button', { name: /1 closed/i })
    fireEvent.click(closedTab)

    expect(screen.getByText('Optimize git-upload-pack packfile negotiation')).toBeInTheDocument()
    expect(screen.queryByText('Support WebAuthn credentials')).not.toBeInTheDocument()
  })

  it('selects an issue, views details and thread, and adds a comment', async () => {
    render(<IssuesView />)

    await waitFor(() => {
      expect(screen.getByText('Support WebAuthn credentials')).toBeInTheDocument()
    })

    // Click on the issue row to open detail view
    fireEvent.click(screen.getByText('Support WebAuthn credentials'))

    expect(screen.getByText('Implement FIDO2 challenge-response.')).toBeInTheDocument()
    expect(screen.getByText('Inspected auth handlers. Ready to implement.')).toBeInTheDocument()

    // Add a comment
    const commentInput = screen.getByPlaceholderText(/leave a comment/i)
    fireEvent.change(commentInput, { target: { value: 'Testing issue reply' } })

    const commentBtn = screen.getByRole('button', { name: /comment/i })
    fireEvent.click(commentBtn)

    await waitFor(() => {
      expect(api.addIssueComment).toHaveBeenCalledWith('shiva', 'exist.io', 1, 'Testing issue reply')
      expect(screen.getByText('Testing issue reply')).toBeInTheDocument()
    })
  })

  it('delegates issue to Exist Agent and displays autonomous agent response', async () => {
    render(<IssuesView />)

    await waitFor(() => {
      expect(screen.getByText('Support WebAuthn credentials')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Support WebAuthn credentials'))

    // Click "Delegate to Agent"
    const delegateBtn = screen.getByRole('button', { name: /delegate to agent/i })
    fireEvent.click(delegateBtn)

    await waitFor(() => {
      expect(api.delegateIssueToAgent).toHaveBeenCalledWith('shiva', 'exist.io', 1)
      expect(screen.getByText(/Draft pull request created/i)).toBeInTheDocument()
    })
  })

  it('toggles issue state between open and closed', async () => {
    render(<IssuesView />)

    await waitFor(() => {
      expect(screen.getByText('Support WebAuthn credentials')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Support WebAuthn credentials'))

    const closeBtn = screen.getByRole('button', { name: /close issue/i })
    fireEvent.click(closeBtn)

    await waitFor(() => {
      expect(api.toggleIssueState).toHaveBeenCalledWith('shiva', 'exist.io', 1, 'open')
    })
  })

  it('opens New Issue modal and submits a new issue', async () => {
    render(<IssuesView />)

    await waitFor(() => {
      expect(screen.getByText('Support WebAuthn credentials')).toBeInTheDocument()
    })

    const newIssueBtn = screen.getByRole('button', { name: /new issue/i })
    fireEvent.click(newIssueBtn)

    expect(screen.getByText('Create new issue')).toBeInTheDocument()
    const titleInput = screen.getByPlaceholderText('Title')
    const descInput = screen.getByPlaceholderText(/describe the steps/i)

    fireEvent.change(titleInput, { target: { value: 'New Test Issue' } })
    fireEvent.change(descInput, { target: { value: 'Description of test issue' } })

    const submitBtn = screen.getByRole('button', { name: /submit new issue/i })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(api.createIssue).toHaveBeenCalled()
      expect(screen.getByText('New Test Issue')).toBeInTheDocument()
    })
  })
})
