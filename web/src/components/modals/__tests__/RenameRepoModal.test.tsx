import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { RenameRepoModal } from '../RenameRepoModal'

// Mock updateRepoSettings API
vi.mock('@/lib/api', () => ({
  updateRepoSettings: vi.fn().mockImplementation(async (owner, repo, payload) => {
    if (payload.new_name === 'existing-repo') {
      throw new Error('Repository with this name already exists')
    }
    return { ok: true }
  }),
}))

describe('RenameRepoModal Component', () => {
  it('renders rename modal with inputs for new repository name and confirmation phrase', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <RenameRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        currentName="exist-core"
        onSuccess={onSuccess}
      />
    )

    expect(screen.getByRole('heading', { name: /rename repository/i })).toBeInTheDocument()
    expect(screen.getByText(/renaming this repository will update its bare git object store/i)).toBeInTheDocument()
    expect(screen.getByTestId('rename-new-name-input')).toBeInTheDocument()
    expect(screen.getByTestId('rename-confirm-name-input')).toBeInTheDocument()
    expect(screen.getByTestId('rename-submit-btn')).toBeDisabled()
  })

  it('enables submit button only when new name is valid and confirmation matches', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <RenameRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        currentName="exist-core"
        onSuccess={onSuccess}
      />
    )

    const newNameInput = screen.getByTestId('rename-new-name-input')
    const confirmInput = screen.getByTestId('rename-confirm-name-input')
    const submitBtn = screen.getByTestId('rename-submit-btn')

    // If new name is same as current name, stays disabled
    fireEvent.change(newNameInput, { target: { value: 'exist-core' } })
    fireEvent.change(confirmInput, { target: { value: 'shiva/exist-core' } })
    expect(submitBtn).toBeDisabled()

    // If new name is different, enables
    fireEvent.change(newNameInput, { target: { value: 'exist-platform' } })
    expect(submitBtn).not.toBeDisabled()
  })

  it('submits repository rename and invokes onSuccess with new name', async () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <RenameRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        currentName="exist-core"
        onSuccess={onSuccess}
      />
    )

    const newNameInput = screen.getByTestId('rename-new-name-input')
    const confirmInput = screen.getByTestId('rename-confirm-name-input')
    const submitBtn = screen.getByTestId('rename-submit-btn')

    fireEvent.change(newNameInput, { target: { value: 'exist-platform' } })
    fireEvent.change(confirmInput, { target: { value: 'shiva/exist-core' } })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledWith('exist-platform')
      expect(onClose).toHaveBeenCalled()
    })
  })

  it('displays error message when rename API fails', async () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <RenameRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        currentName="exist-core"
        onSuccess={onSuccess}
      />
    )

    const newNameInput = screen.getByTestId('rename-new-name-input')
    const confirmInput = screen.getByTestId('rename-confirm-name-input')
    const submitBtn = screen.getByTestId('rename-submit-btn')

    fireEvent.change(newNameInput, { target: { value: 'existing-repo' } })
    fireEvent.change(confirmInput, { target: { value: 'shiva/exist-core' } })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(screen.getByText('Repository with this name already exists')).toBeInTheDocument()
    })
    expect(onSuccess).not.toHaveBeenCalled()
  })
})
