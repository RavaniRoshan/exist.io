import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { DeleteRepoModal } from '../DeleteRepoModal'

// Mock deleteRepository API
vi.mock('@/lib/api', () => ({
  deleteRepository: vi.fn().mockImplementation(async (owner, repo, confirmName, sudoSecret) => {
    if (sudoSecret === 'wrong-secret') {
      throw new Error('Sudo mode authentication failed: invalid credentials')
    }
    return true
  }),
}))

describe('DeleteRepoModal Component', () => {
  it('renders modal with caution callout, consequences checklist, and sudo prompt', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <DeleteRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        onSuccess={onSuccess}
      />
    )

    expect(screen.getByText(/delete repository confirmation/i)).toBeInTheDocument()
    expect(screen.getByText(/this action cannot be undone/i)).toBeInTheDocument()
    expect(screen.getAllByText(/shiva\/exist-core/i).length).toBeGreaterThan(0)
    expect(screen.getByTestId('delete-acknowledge-checkbox')).toBeInTheDocument()
    expect(screen.getByTestId('delete-confirm-name-input')).toBeInTheDocument()
    expect(screen.getByTestId('delete-sudo-secret-input')).toBeInTheDocument()
    expect(screen.getByTestId('delete-submit-btn')).toBeDisabled()
  })

  it('keeps submit button disabled until acknowledgment, name confirmation, and sudo secret are valid', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <DeleteRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        onSuccess={onSuccess}
      />
    )

    const submitBtn = screen.getByTestId('delete-submit-btn')
    expect(submitBtn).toBeDisabled()

    // 1. Check acknowledge checkbox
    const checkbox = screen.getByTestId('delete-acknowledge-checkbox')
    fireEvent.click(checkbox)
    expect(submitBtn).toBeDisabled()

    // 2. Type wrong repo name
    const nameInput = screen.getByTestId('delete-confirm-name-input')
    fireEvent.change(nameInput, { target: { value: 'shiva/wrong-repo' } })
    expect(submitBtn).toBeDisabled()

    // 3. Type matching repo name
    fireEvent.change(nameInput, { target: { value: 'shiva/exist-core' } })
    expect(submitBtn).toBeDisabled()

    // 4. Enter sudo secret
    const secretInput = screen.getByTestId('delete-sudo-secret-input')
    fireEvent.change(secretInput, { target: { value: 'valid-sudo-token' } })
    expect(submitBtn).not.toBeDisabled()
  })

  it('submits deletion when form is valid and calls onSuccess and onClose', async () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <DeleteRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        onSuccess={onSuccess}
      />
    )

    const checkbox = screen.getByTestId('delete-acknowledge-checkbox')
    fireEvent.click(checkbox)

    const nameInput = screen.getByTestId('delete-confirm-name-input')
    fireEvent.change(nameInput, { target: { value: 'shiva/exist-core' } })

    const secretInput = screen.getByTestId('delete-sudo-secret-input')
    fireEvent.change(secretInput, { target: { value: 'super-secret' } })

    const submitBtn = screen.getByTestId('delete-submit-btn')
    expect(submitBtn).not.toBeDisabled()
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalled()
      expect(onClose).toHaveBeenCalled()
    })
  })

  it('displays error banner when API call fails', async () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <DeleteRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        onSuccess={onSuccess}
      />
    )

    fireEvent.click(screen.getByTestId('delete-acknowledge-checkbox'))
    fireEvent.change(screen.getByTestId('delete-confirm-name-input'), { target: { value: 'shiva/exist-core' } })
    fireEvent.change(screen.getByTestId('delete-sudo-secret-input'), { target: { value: 'wrong-secret' } })

    fireEvent.click(screen.getByTestId('delete-submit-btn'))

    await waitFor(() => {
      expect(screen.getByText(/sudo mode authentication failed: invalid credentials/i)).toBeInTheDocument()
    })
    expect(onSuccess).not.toHaveBeenCalled()
  })
})
