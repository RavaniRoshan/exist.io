import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import React from 'react'
import { AuthModal } from '../AuthModal'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof api>('@/lib/api')
  return {
    ...actual,
    loginUser: vi.fn(),
    registerUser: vi.fn(),
  }
})

describe('AuthModal Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders sign in modal and submits credentials', async () => {
    const onSuccess = vi.fn()
    const onClose = vi.fn()
    vi.mocked(api.loginUser).mockResolvedValue({
      ok: true,
      user: {
        id: 'u-1',
        login: 'shiva',
        name: 'Roshan Ravani',
        email: 'roshan@exist.io',
        role: 'admin',
        status: 'active',
      },
    })

    render(<AuthModal isOpen={true} onClose={onClose} onSuccess={onSuccess} />)

    expect(screen.getByText('Sign In to exist.io')).toBeInTheDocument()

    const emailInput = screen.getByPlaceholderText(/developer@company.com/i)
    const passwordInput = screen.getByPlaceholderText(/••••••••••••/i)

    fireEvent.change(emailInput, { target: { value: 'roshan@exist.io' } })
    fireEvent.change(passwordInput, { target: { value: 'SecretPassword123' } })

    const submitBtn = screen.getByTestId('auth-submit-btn')
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(api.loginUser).toHaveBeenCalledWith('roshan@exist.io', 'SecretPassword123')
      expect(onSuccess).toHaveBeenCalledWith(expect.objectContaining({ login: 'shiva' }))
      expect(onClose).toHaveBeenCalled()
    })
  })

  it('switches to Create Account tab and submits registration', async () => {
    const onSuccess = vi.fn()
    const onClose = vi.fn()
    vi.mocked(api.registerUser).mockResolvedValue({
      ok: true,
      user: {
        id: 'u-2',
        login: 'alex',
        name: 'Alex Morgan',
        email: 'alex@exist.io',
        role: 'member',
        status: 'active',
      },
    })

    render(<AuthModal isOpen={true} onClose={onClose} onSuccess={onSuccess} />)

    const createTab = screen.getByTestId('auth-tab-register')
    fireEvent.click(createTab)

    expect(screen.getByText('Create Enterprise Account')).toBeInTheDocument()

    fireEvent.change(screen.getByPlaceholderText(/developer@company.com/i), {
      target: { value: 'alex@exist.io' },
    })
    fireEvent.change(screen.getByPlaceholderText(/e.g. alex/i), {
      target: { value: 'alex' },
    })
    fireEvent.change(screen.getByPlaceholderText(/Alex Morgan/i), {
      target: { value: 'Alex Morgan' },
    })
    fireEvent.change(screen.getByPlaceholderText(/••••••••••••/i), {
      target: { value: 'StrongPassword456' },
    })

    const submitBtn = screen.getByTestId('auth-submit-btn')
    fireEvent.click(submitBtn)

    await waitFor(() => {
      expect(api.registerUser).toHaveBeenCalledWith('alex@exist.io', 'alex', 'Alex Morgan', 'StrongPassword456')
      expect(onSuccess).toHaveBeenCalledWith(expect.objectContaining({ login: 'alex' }))
      expect(onClose).toHaveBeenCalled()
    })
  })

  it('displays error message when login fails', async () => {
    vi.mocked(api.loginUser).mockResolvedValue({
      ok: false,
      error: 'Invalid credentials provided',
    })

    render(<AuthModal isOpen={true} onClose={vi.fn()} onSuccess={vi.fn()} />)

    fireEvent.change(screen.getByPlaceholderText(/developer@company.com/i), {
      target: { value: 'wrong@exist.io' },
    })
    fireEvent.change(screen.getByPlaceholderText(/••••••••••••/i), {
      target: { value: 'WrongPass' },
    })

    fireEvent.click(screen.getByTestId('auth-submit-btn'))

    await waitFor(() => {
      expect(screen.getByText('Invalid credentials provided')).toBeInTheDocument()
    })
  })
})
