import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { SsoConfig } from '../admin/SsoConfig'
import * as api from '@/lib/api'

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api')>('@/lib/api')
  return {
    ...actual,
    fetchSSOConfig: vi.fn(),
    saveSSOConfig: vi.fn(),
  }
})

describe('SsoConfig Admin View', () => {
  const mockConfig = {
    configured: true,
    provider_type: 'oidc' as const,
    issuer_url: 'https://login.microsoftonline.com/tenant/v2.0',
    client_id: 'enterprise_client_id_123',
    enabled: true,
    enforce_sso: false,
  }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.fetchSSOConfig).mockResolvedValue(mockConfig)
    vi.mocked(api.saveSSOConfig).mockResolvedValue({ ok: true })
  })

  it('renders SSO configuration with loaded provider values', async () => {
    render(<SsoConfig />)

    await waitFor(() => {
      expect(screen.getByDisplayValue('https://login.microsoftonline.com/tenant/v2.0')).toBeInTheDocument()
    })

    expect(screen.getByDisplayValue('enterprise_client_id_123')).toBeInTheDocument()
    expect(screen.getByText(/Enterprise Single Sign-On/i)).toBeInTheDocument()
  })

  it('submits updated SSO configuration form', async () => {
    render(<SsoConfig />)

    await waitFor(() => {
      expect(screen.getByDisplayValue('https://login.microsoftonline.com/tenant/v2.0')).toBeInTheDocument()
    })

    const issuerInput = screen.getByDisplayValue('https://login.microsoftonline.com/tenant/v2.0')
    fireEvent.change(issuerInput, { target: { value: 'https://accounts.google.com' } })

    const saveBtn = screen.getByRole('button', { name: /Save Configuration/i })
    fireEvent.click(saveBtn)

    await waitFor(() => {
      expect(api.saveSSOConfig).toHaveBeenCalledWith(
        expect.objectContaining({
          provider_type: 'oidc',
          issuer_url: 'https://accounts.google.com',
          client_id: 'enterprise_client_id_123',
        })
      )
    })
  })

  it('switches between OIDC and SAML provider tabs', async () => {
    render(<SsoConfig />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /OpenID Connect \(OIDC\)/i })).toBeInTheDocument()
    })

    const samlTab = screen.getByRole('button', { name: /SAML 2\.0/i })
    fireEvent.click(samlTab)

    expect(screen.getByText(/SAML IdP Metadata XML/i)).toBeInTheDocument()
  })
})
