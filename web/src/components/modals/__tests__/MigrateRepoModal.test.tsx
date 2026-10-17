import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MigrateRepoModal } from '../MigrateRepoModal'
import * as api from '@/lib/api'

vi.mock('@/lib/api', () => ({
  startRepoMigration: vi.fn(),
  getMigrationStatus: vi.fn(),
}))

describe('MigrateRepoModal', () => {
  it('does not render when closed', () => {
    const { container } = render(
      <MigrateRepoModal isOpen={false} onClose={() => {}} onSuccess={() => {}} />
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders migration form when open', () => {
    render(
      <MigrateRepoModal isOpen={true} onClose={() => {}} onSuccess={() => {}} defaultOwner="shiva" />
    )
    expect(screen.getByText('Migrate Repository')).toBeDefined()
    expect(screen.getByPlaceholderText('https://github.com/owner/repository.git')).toBeDefined()
    expect(screen.getByText('Start Migration')).toBeDefined()
  })

  it('auto-fills repository name when typing a clone URL', () => {
    render(
      <MigrateRepoModal isOpen={true} onClose={() => {}} onSuccess={() => {}} defaultOwner="shiva" />
    )
    const urlInput = screen.getByPlaceholderText('https://github.com/owner/repository.git')
    fireEvent.change(urlInput, { target: { value: 'https://github.com/torvalds/linux.git' } })

    const repoInput = screen.getByPlaceholderText('my-repo') as HTMLInputElement
    expect(repoInput.value).toBe('linux')
  })

  it('submits migration and displays progress bar', async () => {
    vi.mocked(api.startRepoMigration).mockResolvedValueOnce({ ok: true, jobId: 'mig-123' })
    vi.mocked(api.getMigrationStatus).mockResolvedValueOnce({
      id: 'mig-123',
      phase: 'cloning',
      progress: 35,
    })

    render(
      <MigrateRepoModal isOpen={true} onClose={() => {}} onSuccess={() => {}} defaultOwner="shiva" />
    )

    const urlInput = screen.getByPlaceholderText('https://github.com/owner/repository.git')
    fireEvent.change(urlInput, { target: { value: 'https://github.com/expressjs/express.git' } })

    const startBtn = screen.getByText('Start Migration')
    fireEvent.click(startBtn)

    await waitFor(() => {
      expect(api.startRepoMigration).toHaveBeenCalledWith(
        expect.objectContaining({
          clone_url: 'https://github.com/expressjs/express.git',
          owner: 'shiva',
          repo_name: 'express',
        })
      )
    })
  })
})
