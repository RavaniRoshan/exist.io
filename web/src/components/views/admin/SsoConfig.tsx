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
import { Card } from '@/components/ui/Card'
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
      <div className="flex items-center justify-center p-12 text-slate-500 font-mono text-xs">
        <RefreshCw className="w-4 h-4 animate-spin mr-2 text-sky-400" />
        <span>Loading enterprise federation configuration...</span>
      </div>
    )
  }

  return (
    <div className="space-y-6 max-w-4xl">
      {/* Banner */}
      <div className="p-5 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3.5">
          <div className="flex items-center justify-center w-11 h-11 rounded-xl bg-sky-500/10 border border-sky-500/30 text-sky-400">
            <KeyRound className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="text-base font-semibold text-white">Enterprise Single Sign-On (SSO)</h3>
              <Badge variant={config.configured && config.enabled ? 'success' : 'default'} className="text-[10px]">
                {config.configured && config.enabled ? 'Active' : 'Unconfigured'}
              </Badge>
            </div>
            <p className="text-xs text-slate-400 mt-0.5">
              Federate developer and team accounts with corporate IdPs via SAML 2.0 or OpenID Connect (OIDC).
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="text-xs text-slate-300 font-mono">
            RFC 7643 / 7644 SCIM
          </Badge>
        </div>
      </div>

      {statusMsg && (
        <div
          className={`p-3.5 rounded-xl border text-xs flex items-center gap-2 ${
            statusMsg.type === 'success'
              ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300'
              : 'bg-rose-500/10 border-rose-500/30 text-rose-300'
          }`}
        >
          {statusMsg.type === 'success' ? <CheckCircle2 className="w-4 h-4 text-emerald-400" /> : <AlertCircle className="w-4 h-4 text-rose-400" />}
          <span>{statusMsg.text}</span>
        </div>
      )}

      {/* Main Configuration Form */}
      <Card className="border-slate-800 bg-slate-950/70 p-6 space-y-6">
        <form onSubmit={handleSave} className="space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 border-b border-slate-800 gap-3">
            <div>
              <h4 className="text-sm font-semibold text-white">Identity Provider (IdP) Setup</h4>
              <p className="text-xs text-slate-400">Works with Okta, Microsoft Entra ID / Azure AD, Keycloak, Ping, and Google Workspace.</p>
            </div>
            <div className="flex items-center gap-2 p-1 bg-slate-900 border border-slate-800 rounded-lg text-xs">
              <button
                type="button"
                onClick={() => setConfig((p) => ({ ...p, provider_type: 'oidc' }))}
                className={`px-3 py-1 rounded-md font-medium transition-colors ${
                  config.provider_type === 'oidc'
                    ? 'bg-sky-500 text-white shadow-sm'
                    : 'text-slate-400 hover:text-white'
                }`}
              >
                OpenID Connect (OIDC)
              </button>
              <button
                type="button"
                onClick={() => setConfig((p) => ({ ...p, provider_type: 'saml' }))}
                className={`px-3 py-1 rounded-md font-medium transition-colors ${
                  config.provider_type === 'saml'
                    ? 'bg-sky-500 text-white shadow-sm'
                    : 'text-slate-400 hover:text-white'
                }`}
              >
                SAML 2.0
              </button>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-slate-300 flex items-center gap-1.5">
                <Building2 className="w-3.5 h-3.5 text-sky-400" />
                <span>IdP Issuer URL</span>
              </label>
              <Input
                placeholder="https://login.microsoftonline.com/{tenant-id}/v2.0"
                value={config.issuer_url || ''}
                onChange={(e) => setConfig((p) => ({ ...p, issuer_url: e.target.value }))}
                required
                className="bg-slate-900 border-slate-700 text-xs font-mono"
              />
              <span className="text-[10px] text-slate-500">The canonical entity ID / discovery base URL of your identity provider.</span>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-slate-300 flex items-center gap-1.5">
                <Lock className="w-3.5 h-3.5 text-teal-400" />
                <span>Client ID (App ID)</span>
              </label>
              <Input
                placeholder="00000000-0000-0000-0000-000000000000"
                value={config.client_id || ''}
                onChange={(e) => setConfig((p) => ({ ...p, client_id: e.target.value }))}
                required
                className="bg-slate-900 border-slate-700 text-xs font-mono"
              />
              <span className="text-[10px] text-slate-500">Registered application client identifier.</span>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-slate-300">Client Secret</label>
              <Input
                type="password"
                placeholder="••••••••••••••••••••••••"
                value={config.client_secret || ''}
                onChange={(e) => setConfig((p) => ({ ...p, client_secret: e.target.value }))}
                className="bg-slate-900 border-slate-700 text-xs font-mono"
              />
              <span className="text-[10px] text-slate-500">Encrypted and stored in SQLite WAL store.</span>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-slate-300">Authorized Callback URL (Read-only)</label>
              <Input
                value={`${window.location.origin}/api/v1/auth/sso/callback`}
                readOnly
                className="bg-slate-950 border-slate-800 text-xs font-mono text-slate-400 cursor-copy"
                onClick={() => navigator.clipboard.writeText(`${window.location.origin}/api/v1/auth/sso/callback`)}
              />
              <span className="text-[10px] text-slate-500">Register this exact callback URL in your IdP application settings.</span>
            </div>
          </div>

          {config.provider_type === 'saml' && (
            <div className="space-y-1.5 pt-2">
              <label className="text-xs font-medium text-slate-300">SAML IdP Metadata XML (Optional)</label>
              <textarea
                rows={3}
                placeholder="<EntityDescriptor xmlns='urn:oasis:names:tc:SAML:2.0:metadata'..."
                value={config.metadata_xml || ''}
                onChange={(e) => setConfig((p) => ({ ...p, metadata_xml: e.target.value }))}
                className="w-full rounded-xl bg-slate-900 border border-slate-700 p-2.5 text-xs font-mono text-slate-300 focus:outline-none focus:ring-1 focus:ring-sky-500"
              />
            </div>
          )}

          {/* Policy Toggles */}
          <div className="pt-4 border-t border-slate-800/80 space-y-3">
            <label className="flex items-center justify-between p-3 rounded-xl bg-slate-900/50 border border-slate-800 cursor-pointer">
              <div className="space-y-0.5">
                <span className="text-xs font-medium text-slate-200">Enable Single Sign-On</span>
                <p className="text-[11px] text-slate-400">Allows users to sign in with corporate credentials.</p>
              </div>
              <input
                type="checkbox"
                checked={config.enabled ?? true}
                onChange={(e) => setConfig((p) => ({ ...p, enabled: e.target.checked }))}
                className="w-4 h-4 rounded text-sky-500 accent-sky-500 cursor-pointer"
              />
            </label>

            <label className="flex items-center justify-between p-3 rounded-xl bg-slate-900/50 border border-slate-800 cursor-pointer">
              <div className="space-y-0.5">
                <span className="text-xs font-medium text-slate-200">Enforce SSO (Strict Mode)</span>
                <p className="text-[11px] text-slate-400">Disables local password login for non-admin accounts to ensure corporate compliance.</p>
              </div>
              <input
                type="checkbox"
                checked={config.enforce_sso ?? false}
                onChange={(e) => setConfig((p) => ({ ...p, enforce_sso: e.target.checked }))}
                className="w-4 h-4 rounded text-sky-500 accent-sky-500 cursor-pointer"
              />
            </label>
          </div>

          <div className="flex justify-end pt-2">
            <Button type="submit" variant="glow" disabled={saving} className="gap-2">
              <Save className="w-4 h-4" />
              <span>{saving ? 'Saving...' : 'Save Configuration'}</span>
            </Button>
          </div>
        </form>
      </Card>

      {/* SCIM 2.0 User Provisioning Card */}
      <Card className="border-slate-800 bg-slate-950/70 p-6 space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <Zap className="w-4 h-4 text-amber-400" />
            <h4 className="text-sm font-semibold text-white">SCIM 2.0 Automatic User Provisioning</h4>
          </div>
          <Badge variant="outline" className="text-[10px] text-emerald-400 border-emerald-500/30">
            Active on Server
          </Badge>
        </div>
        <p className="text-xs text-slate-400 leading-relaxed">
          `exist.io` implements RFC 7643 and RFC 7644 SCIM 2.0. Users provisioned in Okta or Azure AD are automatically synchronized here, and departing employees have their repository access instantly revoked upon IdP suspension.
        </p>

        <div className="space-y-1.5 pt-2">
          <label className="text-xs font-medium text-slate-300">SCIM 2.0 Base URL</label>
          <div className="flex items-center gap-2">
            <Input
              value={scimEndpoint}
              readOnly
              className="bg-slate-900 border-slate-700 text-xs font-mono text-slate-200"
            />
            <Button variant="outline" size="sm" onClick={handleCopySCIM} className="shrink-0 gap-1.5">
              {copiedSCIM ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              <span>{copiedSCIM ? 'Copied' : 'Copy'}</span>
            </Button>
          </div>
        </div>
      </Card>
    </div>
  )
}
