import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { EnterpriseUsers } from '../admin/EnterpriseUsers'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api')>('@/lib/api')
  return {
    ...actual,
    getEnterpriseUsers: vi.fn(),
    createEnterpriseUser: vi.fn(),
  }
})

describe('EnterpriseUsers Admin View', () => {
  const mockUsers: api.EnterpriseUser[] = [
    {
      id: 'u-1',
      username: 'alice',
      email: 'alice@enterprise.internal',
      name: 'Alice Cooper',
      role: 'admin',
      status: 'active',
      twoFactorEnabled: true,
      lastActive: 'Active now',
      sshKeyCount: 2,
      patCount: 1,
    },
    {
      id: 'u-2',
      username: 'bob',
      email: 'bob@enterprise.internal',
      name: 'Bob Marley',
      role: 'contributor',
      status: 'active',
      twoFactorEnabled: false,
      lastActive: '2 hours ago',
      sshKeyCount: 1,
      patCount: 0,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.getEnterpriseUsers).mockResolvedValue(mockUsers)
    vi.mocked(api.createEnterpriseUser).mockResolvedValue({ ok: true })
  })

  it('renders enterprise users list and filters by search query', async () => {
    render(<EnterpriseUsers />)

    await waitFor(() => {
      expect(screen.getByText('Alice Cooper')).toBeInTheDocument()
      expect(screen.getByText('Bob Marley')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText(/Search users by name/i)
    fireEvent.change(searchInput, { target: { value: 'alice' } })

    expect(screen.getByText('Alice Cooper')).toBeInTheDocument()
    expect(screen.queryByText('Bob Marley')).not.toBeInTheDocument()
  })

  it('opens add user modal and validates required password length', async () => {
    render(<EnterpriseUsers />)

    await waitFor(() => {
      expect(screen.getByText('Alice Cooper')).toBeInTheDocument()
    })

    const addBtn = screen.getByRole('button', { name: /Add User/i })
    fireEvent.click(addBtn)

    expect(screen.getByText('Add Enterprise Member')).toBeInTheDocument()

    const emailInput = screen.getByPlaceholderText(/developer@enterprise.internal/i)
    const passInput = screen.getByPlaceholderText(/••••••••/i)

    fireEvent.change(emailInput, { target: { value: 'carol@enterprise.internal' } })
    fireEvent.change(passInput, { target: { value: 'short' } })

    const submitBtn = screen.getByRole('button', { name: /Create Account/i })
    fireEvent.click(submitBtn)

    expect(screen.getByText(/Password must be at least 8 characters long/i)).toBeInTheDocument()
    expect(api.createEnterpriseUser).not.toHaveBeenCalled()
  })

  it('submits valid new enterprise user and reloads user list', async () => {
    render(<EnterpriseUsers />)

    await waitFor(() => {
      expect(screen.getByText('Alice Cooper')).toBeInTheDocument()
    })

    const addBtn = screen.getByRole('button', { name: /Add User/i })
    fireEvent.click(addBtn)

    const emailInput = screen.getByPlaceholderText(/developer@enterprise.internal/i)
    const passInput = screen.getByPlaceholderText(/••••••••/i)

    fireEvent.change(emailInput, { target: { value: 'carol@enterprise.internal' } })
    fireEvent.change(passInput, { target: { value: 'strongpassword123' } })

    const submitBtn = screen.getByRole('button', { name: /Create Account/i })
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(api.createEnterpriseUser).toHaveBeenCalledWith(
        expect.objectContaining({
          email: 'carol@enterprise.internal',
          login: 'carol',
          password: 'strongpassword123',
        })
      )
    })
  })
})
