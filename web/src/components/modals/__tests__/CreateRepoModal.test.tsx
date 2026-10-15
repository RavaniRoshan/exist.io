import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { CreateRepoModal } from '../CreateRepoModal'

// Mock createRepository API
vi.mock('@/lib/api', () => ({
  createRepository: vi.fn().mockImplementation(async (params) => {
    if (params.name === 'error-repo') {
      return { ok: false, error: 'Repository already exists' }
    }
    return {
      ok: true,
      repo: {
        id: 'repo-123',
        name: params.name,
        full_name: `${params.owner || 'shiva'}/${params.name}`,
        default_branch: params.default_branch || 'main',
        visibility: params.visibility || 'public',
        is_private: params.visibility === 'private',
        clone_url: `http://localhost:8080/${params.owner}/${params.name}.git`,
        web_url: `http://localhost:8080/${params.owner}/${params.name}`,
        updated_at: new Date().toISOString(),
      },
    }
  }),
}))

describe('CreateRepoModal Component', () => {
  it('renders modal with owner, repository name, and visibility options', () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <CreateRepoModal
        isOpen={true}
        onClose={onClose}
        onSuccess={onSuccess}
        defaultOwner="shiva"
      />
    )

    expect(screen.getByText('Create a new repository')).toBeInTheDocument()
    expect(screen.getByDisplayValue('shiva')).toBeInTheDocument()
    expect(screen.getByPlaceholderText(/e\.g\. backend-api or exist-docs/i)).toBeInTheDocument()
    expect(screen.getByText('Public')).toBeInTheDocument()
    expect(screen.getByText('Private')).toBeInTheDocument()
  })

  it('submits repository creation and calls onSuccess', async () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <CreateRepoModal
        isOpen={true}
        onClose={onClose}
        onSuccess={onSuccess}
        defaultOwner="shiva"
      />
    )

    const nameInput = screen.getByPlaceholderText(/e\.g\. backend-api or exist-docs/i)
    fireEvent.change(nameInput, { target: { value: 'microservice-orchestrator' } })

    const submitBtn = screen.getByRole('button', { name: /create repository/i })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'microservice-orchestrator',
          full_name: 'shiva/microservice-orchestrator',
        })
      )
      expect(onClose).toHaveBeenCalled()
    })
  })

  it('displays error alert when repository creation fails', async () => {
    const onClose = vi.fn()
    const onSuccess = vi.fn()

    render(
      <CreateRepoModal
        isOpen={true}
        onClose={onClose}
        onSuccess={onSuccess}
      />
    )

    const nameInput = screen.getByPlaceholderText(/e\.g\. backend-api or exist-docs/i)
    fireEvent.change(nameInput, { target: { value: 'error-repo' } })

    const submitBtn = screen.getByRole('button', { name: /create repository/i })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(screen.getByText('Repository already exists')).toBeInTheDocument()
    })
    expect(onSuccess).not.toHaveBeenCalled()
  })
})
