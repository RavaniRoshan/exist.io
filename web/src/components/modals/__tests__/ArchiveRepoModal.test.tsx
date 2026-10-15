import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { ArchiveRepoModal } from '../ArchiveRepoModal'

// Mock updateRepoSettings API
vi.mock('@/lib/api', () => ({
  updateRepoSettings: vi.fn().mockImplementation(async (owner, repo, payload) => {
    if (payload.confirm_name !== `${owner}/${repo}` && payload.confirm_name !== repo) {
      throw new Error('Confirmation string does not match repository name')
    }
    return { ok: true }
  }),
}))

describe('ArchiveRepoModal Component', () => {
  it('renders archive modal with warnings and confirmation input', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <ArchiveRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        isCurrentlyArchived={false}
        onSuccess={onSuccess}
      />
    )

    expect(screen.getByText(/archive repository/i)).toBeInTheDocument()
    expect(screen.getByText(/mark repository as read-only and immutable/i)).toBeInTheDocument()
    expect(screen.getByPlaceholderText('shiva/exist-core')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /i understand, archive this repository/i })).toBeDisabled()
  })

  it('enables submit button only when repository name matches', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <ArchiveRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        isCurrentlyArchived={false}
        onSuccess={onSuccess}
      />
    )

    const input = screen.getByPlaceholderText('shiva/exist-core')
    const submitBtn = screen.getByRole('button', { name: /archive this repository/i })

    fireEvent.change(input, { target: { value: 'shiva/other-repo' } })
    expect(submitBtn).toBeDisabled()

    fireEvent.change(input, { target: { value: 'shiva/exist-core' } })
    expect(submitBtn).not.toBeDisabled()
  })

  it('submits archival and calls onSuccess and onClose', async () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <ArchiveRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        isCurrentlyArchived={false}
        onSuccess={onSuccess}
      />
    )

    const input = screen.getByPlaceholderText('shiva/exist-core')
    fireEvent.change(input, { target: { value: 'shiva/exist-core' } })

    const submitBtn = screen.getByRole('button', { name: /archive this repository/i })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalled()
      expect(onClose).toHaveBeenCalled()
    })
  })

  it('renders unarchive UI when isCurrentlyArchived is true', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <ArchiveRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        isCurrentlyArchived={true}
        onSuccess={onSuccess}
      />
    )

    expect(screen.getByText(/unarchive repository/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /unarchive this repository/i })).toBeInTheDocument()
  })
})
