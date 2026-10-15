import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { CreateFileModal } from '../CreateFileModal'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<any>('@/lib/api')
  return {
    ...actual,
    commitFile: vi.fn(),
  }
})

describe('CreateFileModal Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders modal with default README.md filename and content when open', () => {
    render(
      <CreateFileModal
        isOpen={true}
        onClose={vi.fn()}
        owner="shiva"
        name="exist.io"
        branch="main"
        initialPath="README.md"
        onCommitSuccess={vi.fn()}
      />
    )

    expect(screen.getByRole('heading', { name: /create new file/i })).toBeDefined()
    expect(screen.getByLabelText(/name your file/i)).toHaveValue('README.md')
    expect(screen.getByLabelText(/commit message/i)).toHaveValue('Create README.md')
  })

  it('updates commit message when file path input changes', () => {
    render(
      <CreateFileModal
        isOpen={true}
        onClose={vi.fn()}
        owner="shiva"
        name="exist.io"
        branch="main"
        initialPath="README.md"
        onCommitSuccess={vi.fn()}
      />
    )

    const pathInput = screen.getByLabelText(/name your file/i)
    fireEvent.change(pathInput, { target: { value: 'main.go' } })

    expect(screen.getByLabelText(/commit message/i)).toHaveValue('Create main.go')
  })

  it('submits commitFile and invokes onCommitSuccess and onClose on success', async () => {
    const onCommitSuccess = vi.fn()
    const onClose = vi.fn()
    vi.mocked(api.commitFile).mockResolvedValueOnce({ ok: true, commit_sha: 'abc1234' })

    render(
      <CreateFileModal
        isOpen={true}
        onClose={onClose}
        owner="shiva"
        name="exist.io"
        branch="main"
        initialPath="README.md"
        onCommitSuccess={onCommitSuccess}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: /commit new file/i }))

    await waitFor(() => {
      expect(api.commitFile).toHaveBeenCalledWith(
        expect.objectContaining({
          owner: 'shiva',
          name: 'exist.io',
          path: 'README.md',
          branch: 'main',
        })
      )
      expect(onCommitSuccess).toHaveBeenCalled()
      expect(onClose).toHaveBeenCalled()
    })
  })
})
