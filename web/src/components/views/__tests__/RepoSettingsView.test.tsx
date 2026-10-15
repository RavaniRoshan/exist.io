import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { RepoSettingsView } from '../RepoSettingsView'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    getBranches: vi.fn().mockResolvedValue([
      { name: 'main', commitSha: '5c92f55', isDefault: true, protected: true },
      { name: 'feat/test', commitSha: '1234567', isDefault: false, protected: false },
    ]),
    getCollaborators: vi.fn().mockResolvedValue([
      {
        id: 'collab-1',
        login: 'shiva',
        name: 'Shiva Admin',
        email: 'shiva@exist.io',
        role: 'ADMIN',
        permission: 'admin',
        createdAt: '1 day ago',
      },
      {
        id: 'collab-2',
        login: 'alex-engineer',
        name: 'Alex Engineer',
        email: 'alex@exist.io',
        role: 'WRITE',
        permission: 'write',
        createdAt: '2 hours ago',
      },
    ]),
    updateRepoSettings: vi.fn().mockResolvedValue({
      name: 'rocket',
      defaultBranch: 'main',
      isPrivate: false,
      allowSquashMerge: true,
      allowMergeCommit: true,
      deleteBranchOnMerge: false,
    }),
    addCollaborator: vi.fn().mockResolvedValue({
      id: 'collab-3',
      login: 'new-dev',
      name: 'new-dev',
      role: 'WRITE',
      permission: 'write',
      createdAt: 'Just now',
    }),
    removeCollaborator: vi.fn().mockResolvedValue(undefined),
    deleteRepository: vi.fn().mockResolvedValue(undefined),
  }
})

describe('RepoSettingsView', () => {
  const mockRepo = {
    ownerSlug: 'shiva',
    name: 'rocket',
    description: 'Enterprise aerospace system',
    defaultBranch: 'main',
    isPrivate: false,
    starsCount: 42,
    forksCount: 3,
    updatedAt: 'Just now',
    cloneUrlHttp: 'http://localhost:8080/shiva/rocket.git',
    cloneUrlSsh: 'git@localhost:2222/shiva/rocket.git',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders General settings section with branch options and merge toggles', async () => {
    render(<RepoSettingsView repo={mockRepo} />)

    expect(screen.getByText('General Repository Configuration')).toBeInTheDocument()
    expect(screen.getByText('Default branch')).toBeInTheDocument()
    expect(screen.getByText('Pull Request Merge Buttons')).toBeInTheDocument()
    expect(screen.getByText('Allow squash merging')).toBeInTheDocument()
    expect(screen.getByText('Allow merge commits')).toBeInTheDocument()

    // Test save changes
    const saveButton = screen.getByRole('button', { name: /save changes/i })
    fireEvent.click(saveButton)

    await waitFor(() => {
      expect(api.updateRepoSettings).toHaveBeenCalledWith(
        'shiva',
        'rocket',
        expect.objectContaining({
          default_branch: 'main',
          allow_squash_merge: true,
          allow_merge_commit: true,
        })
      )
    })
  })

  it('switches to Collaborators & Access section, displays members and adds new member', async () => {
    render(<RepoSettingsView repo={mockRepo} />)

    // Switch tab
    const collabTab = screen.getByRole('button', { name: /collaborators & access/i })
    fireEvent.click(collabTab)

    await waitFor(() => {
      expect(screen.getByText('Manage Access & Collaborators')).toBeInTheDocument()
      expect(screen.getByText('Shiva Admin')).toBeInTheDocument()
      expect(screen.getByText('@alex-engineer')).toBeInTheDocument()
    })

    // Open Add Collaborator modal
    const openAddBtn = screen.getByRole('button', { name: /add collaborator/i })
    fireEvent.click(openAddBtn)

    // Modal input
    const input = screen.getByPlaceholderText('e.g. alex')
    fireEvent.change(input, { target: { value: 'new-dev' } })

    const allAddButtons = screen.getAllByRole('button', { name: /add collaborator/i })
    // The modal submit button is the second one
    fireEvent.click(allAddButtons[1])

    await waitFor(() => {
      expect(api.addCollaborator).toHaveBeenCalledWith('shiva', 'rocket', 'new-dev', 'write')
    })
  })

  it('switches to Danger Zone section and opens repository deletion modal', async () => {
    render(<RepoSettingsView repo={mockRepo} />)

    // Switch to danger zone
    const dangerTab = screen.getByRole('button', { name: /danger zone/i })
    fireEvent.click(dangerTab)

    expect(screen.getByText('Irreversible actions that affect repository visibility or permanently destroy data.')).toBeInTheDocument()
    const deleteRepoBtn = screen.getByRole('button', { name: /delete this repository/i })
    fireEvent.click(deleteRepoBtn)

    // Confirmation modal should be open
    expect(screen.getByText('Delete repository confirmation')).toBeInTheDocument()
    expect(screen.getByText('This action cannot be undone.')).toBeInTheDocument()

    const confirmInput = screen.getByPlaceholderText('shiva/rocket')
    fireEvent.change(confirmInput, { target: { value: 'shiva/rocket' } })

    const confirmDeleteBtn = screen.getByRole('button', { name: /i understand, delete this repository/i })
    expect(confirmDeleteBtn).not.toBeDisabled()
  })
})
