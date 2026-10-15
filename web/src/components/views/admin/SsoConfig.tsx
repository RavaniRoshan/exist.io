import React, { useState, useEffect } from 'react'
import {
  KeyRound,
  CheckCircle2,
  Save,
  Building2,
  Lock,
  AlertCircle,
  Copy,
  Check,
  RefreshCw,
  Zap,
} from 'lucide-react'
import { Box, BoxHeader, BoxBody } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Input } from '@/components/ui/Input'
import { fetchSSOConfig, saveSSOConfig } from '@/lib/api'
import type { SSOConfigData } from '@/lib/types'

export const SsoConfig: React.FC = () => {
  const [config, setConfig] = useState<SSOConfigData>({
    configured: false,
    provider_type: 'oidc',
    issuer_url: '',
    client_id: '',
    client_secret: '',
    metadata_xml: '',
    enabled: true,
    enforce_sso: false,
  })
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [statusMsg, setStatusMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [copiedSCIM, setCopiedSCIM] = useState(false)

  const scimEndpoint = `${window.location.origin}/scim/v2`

  useEffect(() => {
    loadConfig()
  }, [])

  const loadConfig = async () => {
    setLoading(true)
    const data = await fetchSSOConfig()
    if (data.configured) {
      setConfig({
        ...data,
        provider_type: data.provider_type || 'oidc',
        enabled: data.enabled ?? true,
        enforce_sso: data.enforce_sso ?? false,
      })
    }
    setLoading(false)
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!config.issuer_url?.trim() || !config.client_id?.trim()) {
      setStatusMsg({ type: 'error', text: 'Issuer URL and Client ID are required' })
      return
    }

    setSaving(true)
    setStatusMsg(null)

    const res = await saveSSOConfig({
      provider_type: config.provider_type || 'oidc',
      issuer_url: config.issuer_url.trim(),
      client_id: config.client_id.trim(),
      client_secret: config.client_secret?.trim() || '',
      metadata_xml: config.metadata_xml?.trim() || '',
      enabled: config.enabled ?? true,
      enforce_sso: config.enforce_sso ?? false,
    })

    setSaving(false)
    if (res.ok) {
      setConfig((prev) => ({ ...prev, configured: true }))
      setStatusMsg({ type: 'success', text: 'SSO configuration saved successfully' })
      setTimeout(() => setStatusMsg(null), 3500)
    } else {
      setStatusMsg({ type: 'error', text: res.error || 'Failed to update SSO configuration' })
    }
  }

  const handleCopySCIM = () => {
    navigator.clipboard.writeText(scimEndpoint)
    setCopiedSCIM(true)
    setTimeout(() => setCopiedSCIM(false), 2000)
  }

  if (loading) {
    return (
      <Box className="p-12 text-center text-muted-foreground font-mono text-xs">
        <RefreshCw className="w-4 h-4 animate-spin mr-2 inline text-primary" />
        <span>Loading enterprise federation configuration...</span>
      </Box>
    )
  }

  return (
    <div className="space-y-4 max-w-4xl">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        <div className="space-y-0.5">
          <div className="flex items-center gap-2">
            <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
              <KeyRound className="w-4 h-4 text-primary" />
              <span>Enterprise Single Sign-On (SSO)</span>
            </h2>
            <Badge variant={config.configured && config.enabled ? 'success' : 'default'} className="text-[10px]">
              {config.configured && config.enabled ? 'Active' : 'Unconfigured'}
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground">
            Federate developer and team accounts with corporate IdPs via SAML 2.0 or OpenID Connect (OIDC).
          </p>
        </div>
        <Badge variant="outline" className="text-xs font-mono">
          RFC 7643 / 7644 SCIM
        </Badge>
      </div>

      {statusMsg && (
        <div
          className={`p-3 rounded-md border text-xs flex items-center gap-2 ${
            statusMsg.type === 'success'
              ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-500'
              : 'bg-destructive/10 border-destructive/30 text-destructive'
          }`}
        >
          {statusMsg.type === 'success' ? <CheckCircle2 className="w-4 h-4 text-emerald-500" /> : <AlertCircle className="w-4 h-4 text-destructive" />}
          <span>{statusMsg.text}</span>
        </div>
      )}

      {/* Main Configuration Form */}
      <Box>
        <BoxHeader className="bg-secondary/35 py-2.5">
          <span className="font-semibold text-foreground text-xs">Identity Provider (IdP) Setup</span>
          <div className="flex items-center gap-1.5 p-0.5 bg-background border border-border rounded-md text-xs">
            <button
              type="button"
              onClick={() => setConfig((p) => ({ ...p, provider_type: 'oidc' }))}
              className={`px-2.5 py-1 rounded transition-colors ${
                config.provider_type === 'oidc'
                  ? 'bg-primary text-primary-foreground font-semibold shadow-2xs'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              OpenID Connect (OIDC)
            </button>
            <button
              type="button"
              onClick={() => setConfig((p) => ({ ...p, provider_type: 'saml' }))}
              className={`px-2.5 py-1 rounded transition-colors ${
                config.provider_type === 'saml'
                  ? 'bg-primary text-primary-foreground font-semibold shadow-2xs'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              SAML 2.0
            </button>
          </div>
        </BoxHeader>

        <BoxBody className="p-4 space-y-4">
          <form onSubmit={handleSave} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-1">
                <label className="text-xs font-medium text-foreground flex items-center gap-1.5">
                  <Building2 className="w-3.5 h-3.5 text-primary" />
                  <span>IdP Issuer URL</span>
                </label>
                <Input
                  placeholder="https://login.microsoftonline.com/{tenant-id}/v2.0"
                  value={config.issuer_url || ''}
                  onChange={(e) => setConfig((p) => ({ ...p, issuer_url: e.target.value }))}
                  required
                  className="text-xs font-mono"
                />
                <span className="text-[10px] text-muted-foreground">The canonical entity ID / discovery base URL of your identity provider.</span>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-medium text-foreground flex items-center gap-1.5">
                  <Lock className="w-3.5 h-3.5 text-primary" />
                  <span>Client ID (App ID)</span>
                </label>
                <Input
                  placeholder="00000000-0000-0000-0000-000000000000"
                  value={config.client_id || ''}
                  onChange={(e) => setConfig((p) => ({ ...p, client_id: e.target.value }))}
                  required
                  className="text-xs font-mono"
                />
                <span className="text-[10px] text-muted-foreground">Registered application client identifier.</span>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-medium text-foreground">Client Secret</label>
                <Input
                  type="password"
                  placeholder="••••••••••••••••••••••••"
                  value={config.client_secret || ''}
                  onChange={(e) => setConfig((p) => ({ ...p, client_secret: e.target.value }))}
                  className="text-xs font-mono"
                />
                <span className="text-[10px] text-muted-foreground">Encrypted and stored in SQLite WAL store.</span>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-medium text-foreground">Authorized Callback URL (Read-only)</label>
                <Input
                  value={`${window.location.origin}/api/v1/auth/sso/callback`}
                  readOnly
                  className="text-xs font-mono text-muted-foreground cursor-copy bg-secondary/30"
                  onClick={() => navigator.clipboard.writeText(`${window.location.origin}/api/v1/auth/sso/callback`)}
                />
                <span className="text-[10px] text-muted-foreground">Register this exact callback URL in your IdP application settings.</span>
              </div>
            </div>

            {config.provider_type === 'saml' && (
              <div className="space-y-1 pt-1">
                <label className="text-xs font-medium text-foreground">SAML IdP Metadata XML (Optional)</label>
                <textarea
                  rows={3}
                  placeholder="<EntityDescriptor xmlns='urn:oasis:names:tc:SAML:2.0:metadata'..."
                  value={config.metadata_xml || ''}
                  onChange={(e) => setConfig((p) => ({ ...p, metadata_xml: e.target.value }))}
                  className="w-full rounded-md bg-input border border-border p-2.5 text-xs font-mono text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                />
              </div>
            )}

            {/* Policy Toggles */}
            <div className="pt-3 border-t border-border space-y-2.5">
              <label className="flex items-center justify-between p-2.5 rounded-md bg-secondary/30 border border-border cursor-pointer">
                <div className="space-y-0.5">
                  <span className="text-xs font-medium text-foreground">Enable Single Sign-On</span>
                  <p className="text-[11px] text-muted-foreground">Allows users to sign in with corporate credentials.</p>
                </div>
                <input
                  type="checkbox"
                  checked={config.enabled ?? true}
                  onChange={(e) => setConfig((p) => ({ ...p, enabled: e.target.checked }))}
                  className="w-4 h-4 rounded text-primary accent-primary cursor-pointer"
                />
              </label>

              <label className="flex items-center justify-between p-2.5 rounded-md bg-secondary/30 border border-border cursor-pointer">
                <div className="space-y-0.5">
                  <span className="text-xs font-medium text-foreground">Enforce SSO (Strict Mode)</span>
                  <p className="text-[11px] text-muted-foreground">Disables local password login for non-admin accounts to ensure corporate compliance.</p>
                </div>
                <input
                  type="checkbox"
                  checked={config.enforce_sso ?? false}
                  onChange={(e) => setConfig((p) => ({ ...p, enforce_sso: e.target.checked }))}
                  className="w-4 h-4 rounded text-primary accent-primary cursor-pointer"
                />
              </label>
            </div>

            <div className="flex justify-end pt-2">
              <Button type="submit" variant="primary" size="sm" disabled={saving} className="gap-2 text-xs">
                <Save className="w-3.5 h-3.5" />
                <span>{saving ? 'Saving...' : 'Save Configuration'}</span>
              </Button>
            </div>
          </form>
        </BoxBody>
      </Box>

      {/* SCIM 2.0 User Provisioning Box */}
      <Box>
        <BoxHeader className="bg-secondary/35 py-2.5">
          <div className="flex items-center gap-2">
            <Zap className="w-4 h-4 text-amber-500" />
            <span className="text-xs font-semibold text-foreground">SCIM 2.0 Automatic User Provisioning</span>
          </div>
          <Badge variant="outline" className="text-[10px]">
            Active on Server
          </Badge>
        </BoxHeader>

        <BoxBody className="p-4 space-y-3">
          <p className="text-xs text-muted-foreground leading-relaxed">
            <code>exist.io</code> implements RFC 7643 and RFC 7644 SCIM 2.0. Users provisioned in Okta or Azure AD are automatically synchronized here, and departing employees have their repository access instantly revoked upon IdP suspension.
          </p>

          <div className="space-y-1 pt-1">
            <label className="text-xs font-medium text-foreground">SCIM 2.0 Base URL</label>
            <div className="flex items-center gap-2">
              <Input
                value={scimEndpoint}
                readOnly
                className="text-xs font-mono text-muted-foreground bg-secondary/30"
              />
              <Button variant="outline" size="sm" onClick={handleCopySCIM} className="shrink-0 gap-1.5 text-xs">
                {copiedSCIM ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
                <span>{copiedSCIM ? 'Copied' : 'Copy'}</span>
              </Button>
            </div>
          </div>
        </BoxBody>
      </Box>
    </div>
  )
}
