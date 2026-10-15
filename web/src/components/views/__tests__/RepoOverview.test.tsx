import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { RepoOverview } from '../RepoOverview'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<any>('@/lib/api')
  return {
    ...actual,
    getFiles: vi.fn(),
    getFileContent: vi.fn(),
    commitFile: vi.fn(),
    searchRepoSymbols: vi.fn(),
    isDemoModeActive: vi.fn(() => false),
  }
})

describe('RepoOverview Component', () => {
  const mockRepo = {
    ownerSlug: 'shiva',
    name: 'exist.io',
    description: 'Enterprise repository on exist.io',
    defaultBranch: 'main',
    isPrivate: false,
    starsCount: 12,
    forksCount: 3,
    updatedAt: 'Just now',
    cloneUrlHttp: 'http://127.0.0.1:8080/shiva/exist.io.git',
    cloneUrlSsh: 'git@127.0.0.1:2222/shiva/exist.io.git',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders canonical GitHub Quick Setup screen when repository is empty', async () => {
    vi.mocked(api.getFiles).mockResolvedValueOnce([])

    render(
      <RepoOverview
        repo={mockRepo}
        onSelectCommits={vi.fn()}
        onSelectAgentConsole={vi.fn()}
      />
    )

    await waitFor(() => {
      expect(
        screen.getByText(/quick setup — if you've done this kind of thing before/i)
      ).toBeDefined()
    })

    expect(screen.getByText(/…or create a new repository on the command line/i)).toBeDefined()
    expect(screen.getByText(/…or push an existing repository from the command line/i)).toBeDefined()
    expect(screen.getByText(/…or initialize with an autonomous ai coding agent/i)).toBeDefined()
  })

  it('renders files table and README when repository has files', async () => {
    vi.mocked(api.getFiles).mockResolvedValueOnce([
      {
        name: 'README.md',
        path: 'README.md',
        type: 'blob',
        size: 128,
        lastCommitMessage: 'Initial commit: create README.md',
        lastCommitDate: 'Just now',
      },
      {
        name: 'main.go',
        path: 'main.go',
        type: 'blob',
        size: 256,
        lastCommitMessage: 'feat: bootstrap server',
        lastCommitDate: 'Just now',
      },
    ])
    vi.mocked(api.getFileContent).mockResolvedValueOnce('# exist.io\n\nPure Go Git Forge')

    render(
      <RepoOverview
        repo={mockRepo}
        onSelectCommits={vi.fn()}
        onSelectAgentConsole={vi.fn()}
      />
    )

    await waitFor(() => {
      expect(screen.getAllByText('README.md').length).toBeGreaterThanOrEqual(1)
      expect(screen.getByText('main.go')).toBeDefined()
    })

    expect(screen.getAllByText('Initial commit: create README.md').length).toBeGreaterThanOrEqual(1)
  })

  it('opens CreateFileModal when clicking Add file button', async () => {
    vi.mocked(api.getFiles).mockResolvedValueOnce([])

    render(
      <RepoOverview
        repo={mockRepo}
        onSelectCommits={vi.fn()}
        onSelectAgentConsole={vi.fn()}
      />
    )

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /add file/i })).toBeDefined()
    })

    fireEvent.click(screen.getByRole('button', { name: /add file/i }))

    expect(screen.getByRole('heading', { name: /create new file/i })).toBeDefined()
  })
})
