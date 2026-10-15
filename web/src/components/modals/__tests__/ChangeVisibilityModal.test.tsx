import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { ChangeVisibilityModal } from '../ChangeVisibilityModal'

// Mock updateRepoSettings API
vi.mock('@/lib/api', () => ({
  updateRepoSettings: vi.fn().mockImplementation(async (owner, repo, payload) => {
    if (payload.sudo_secret === 'wrong-secret') {
      throw new Error('Sudo mode authentication failed: invalid credentials')
    }
    return { ok: true }
  }),
}))

describe('ChangeVisibilityModal Component', () => {
  it('renders modal with caution callout, consequences checklist, and sudo prompt', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <ChangeVisibilityModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        isCurrentlyPrivate={false}
        onSuccess={onSuccess}
      />
    )

    expect(screen.getByText(/change visibility to private/i)).toBeInTheDocument()
    expect(screen.getByText(/caution: high-impact security and visibility change/i)).toBeInTheDocument()
    expect(screen.getAllByText(/shiva\/exist-core/i).length).toBeGreaterThan(0)
    expect(screen.getByTestId('visibility-acknowledge-checkbox')).toBeInTheDocument()
    expect(screen.getByTestId('visibility-confirm-name-input')).toBeInTheDocument()
    expect(screen.getByTestId('visibility-sudo-secret-input')).toBeInTheDocument()
    expect(screen.getByTestId('visibility-submit-btn')).toBeDisabled()
  })

  it('keeps submit button disabled until acknowledgment, name confirmation, and sudo secret are valid', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <ChangeVisibilityModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        isCurrentlyPrivate={true}
        onSuccess={onSuccess}
      />
    )

    const submitBtn = screen.getByTestId('visibility-submit-btn')
    expect(submitBtn).toBeDisabled()

    // 1. Check acknowledge checkbox
    const checkbox = screen.getByTestId('visibility-acknowledge-checkbox')
    fireEvent.click(checkbox)
    expect(submitBtn).toBeDisabled()

    // 2. Type wrong repo name
    const nameInput = screen.getByTestId('visibility-confirm-name-input')
    fireEvent.change(nameInput, { target: { value: 'shiva/wrong-repo' } })
    expect(submitBtn).toBeDisabled()

    // 3. Type matching repo name
    fireEvent.change(nameInput, { target: { value: 'shiva/exist-core' } })
    expect(submitBtn).toBeDisabled()

    // 4. Enter sudo secret (< 4 chars)
    const secretInput = screen.getByTestId('visibility-sudo-secret-input')
    fireEvent.change(secretInput, { target: { value: '123' } })
    expect(submitBtn).toBeDisabled()

    // 5. Enter valid sudo secret (>= 4 chars)
    fireEvent.change(secretInput, { target: { value: 'valid-sudo-token' } })
    expect(submitBtn).not.toBeDisabled()
  })

  it('submits visibility change and invokes onSuccess callback', async () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <ChangeVisibilityModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        isCurrentlyPrivate={true}
        onSuccess={onSuccess}
      />
    )

    fireEvent.click(screen.getByTestId('visibility-acknowledge-checkbox'))
    fireEvent.change(screen.getByTestId('visibility-confirm-name-input'), {
      target: { value: 'shiva/exist-core' },
    })
    fireEvent.change(screen.getByTestId('visibility-sudo-secret-input'), {
      target: { value: 'valid-sudo-token' },
    })

    const submitBtn = screen.getByTestId('visibility-submit-btn')
    expect(submitBtn).not.toBeDisabled()
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalled()
      expect(onClose).toHaveBeenCalled()
    })
  })

  it('displays error message when sudo authentication fails', async () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <ChangeVisibilityModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="exist-core"
        isCurrentlyPrivate={false}
        onSuccess={onSuccess}
      />
    )

    fireEvent.click(screen.getByTestId('visibility-acknowledge-checkbox'))
    fireEvent.change(screen.getByTestId('visibility-confirm-name-input'), {
      target: { value: 'shiva/exist-core' },
    })
    fireEvent.change(screen.getByTestId('visibility-sudo-secret-input'), {
      target: { value: 'wrong-secret' },
    })

    fireEvent.click(screen.getByTestId('visibility-submit-btn'))

    await waitFor(() => {
      expect(
        screen.getByText(/sudo mode authentication failed: invalid credentials/i)
      ).toBeInTheDocument()
    })
    expect(onSuccess).not.toHaveBeenCalled()
  })
})
