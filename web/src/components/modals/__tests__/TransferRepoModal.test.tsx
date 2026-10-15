import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { TransferRepoModal } from '../TransferRepoModal'

// Mock transferRepository API
vi.mock('@/lib/api', () => ({
  transferRepository: vi.fn().mockImplementation(async (owner, repo, payload) => {
    if (payload.sudo_secret === 'wrong-secret') {
      throw new Error('Invalid security credentials for sudo mode')
    }
    return { ok: true, repo: { fullName: `${payload.new_owner}/${repo}` } }
  }),
}))

describe('TransferRepoModal Component', () => {
  it('renders modal with warning callout, consequences checklist, owner input, and sudo prompt', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <TransferRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="satellite"
        onSuccess={onSuccess}
      />
    )

    expect(screen.getByText(/transfer repository ownership/i)).toBeInTheDocument()
    expect(screen.getByText(/transferring ownership to another account or organization/i)).toBeInTheDocument()
    expect(screen.getAllByText(/shiva\/satellite/i).length).toBeGreaterThan(0)
    expect(screen.getByTestId('transfer-owner-input')).toBeInTheDocument()
    expect(screen.getByTestId('transfer-ack-checkbox')).toBeInTheDocument()
    expect(screen.getByTestId('transfer-confirm-input')).toBeInTheDocument()
    expect(screen.getByTestId('transfer-sudo-input')).toBeInTheDocument()
    expect(screen.getByTestId('transfer-submit-btn')).toBeDisabled()
  })

  it('keeps submit button disabled until all criteria are fulfilled', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <TransferRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="satellite"
        onSuccess={onSuccess}
      />
    )

    const submitBtn = screen.getByTestId('transfer-submit-btn')
    expect(submitBtn).toBeDisabled()

    // 1. Enter same owner -> should show error notice and remain disabled
    const ownerInput = screen.getByTestId('transfer-owner-input')
    fireEvent.change(ownerInput, { target: { value: 'shiva' } })
    expect(screen.getByText(/new owner cannot be the current repository owner/i)).toBeInTheDocument()
    expect(submitBtn).toBeDisabled()

    // 2. Enter valid new owner
    fireEvent.change(ownerInput, { target: { value: 'globex' } })
    expect(submitBtn).toBeDisabled()

    // 3. Check acknowledge checkbox
    const ackCheckbox = screen.getByTestId('transfer-ack-checkbox')
    fireEvent.click(ackCheckbox)
    expect(submitBtn).toBeDisabled()

    // 4. Enter wrong confirmation name
    const confirmInput = screen.getByTestId('transfer-confirm-input')
    fireEvent.change(confirmInput, { target: { value: 'shiva/wrong' } })
    expect(submitBtn).toBeDisabled()

    // 5. Enter correct confirmation name
    fireEvent.change(confirmInput, { target: { value: 'shiva/satellite' } })
    expect(submitBtn).toBeDisabled()

    // 6. Enter sudo secret (less than 4 chars)
    const sudoInput = screen.getByTestId('transfer-sudo-input')
    fireEvent.change(sudoInput, { target: { value: '123' } })
    expect(submitBtn).toBeDisabled()

    // 7. Enter valid sudo secret (>= 4 chars) -> submit enabled!
    fireEvent.change(sudoInput, { target: { value: 'secret123' } })
    expect(submitBtn).toBeEnabled()
  })

  it('submits repository transfer and triggers onSuccess callback', async () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <TransferRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="satellite"
        onSuccess={onSuccess}
      />
    )

    fireEvent.change(screen.getByTestId('transfer-owner-input'), { target: { value: 'globex' } })
    fireEvent.click(screen.getByTestId('transfer-ack-checkbox'))
    fireEvent.change(screen.getByTestId('transfer-confirm-input'), { target: { value: 'shiva/satellite' } })
    fireEvent.change(screen.getByTestId('transfer-sudo-input'), { target: { value: 'valid-secret' } })

    const submitBtn = screen.getByTestId('transfer-submit-btn')
    expect(submitBtn).toBeEnabled()
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledWith('globex', 'satellite')
      expect(onClose).toHaveBeenCalled()
    })
  })

  it('displays error message when sudo authentication fails', async () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <TransferRepoModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        repoName="satellite"
        onSuccess={onSuccess}
      />
    )

    fireEvent.change(screen.getByTestId('transfer-owner-input'), { target: { value: 'globex' } })
    fireEvent.click(screen.getByTestId('transfer-ack-checkbox'))
    fireEvent.change(screen.getByTestId('transfer-confirm-input'), { target: { value: 'shiva/satellite' } })
    fireEvent.change(screen.getByTestId('transfer-sudo-input'), { target: { value: 'wrong-secret' } })

    fireEvent.click(screen.getByTestId('transfer-submit-btn'))

    await waitFor(() => {
      expect(screen.getByText(/invalid security credentials for sudo mode/i)).toBeInTheDocument()
      expect(onSuccess).not.toHaveBeenCalled()
    })
  })
})
