import React, { useState, useEffect } from 'react'
import { KeyRound, ShieldCheck, CheckCircle2, Save } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Input } from '@/components/ui/Input'
import { getSsoProviders, type SsoProvider } from '@/lib/api'

export const SsoConfig: React.FC = () => {
  const [providers, setProviders] = useState<SsoProvider[]>([])
  const [saving, setSaving] = useState(false)
  const [savedSuccess, setSavedSuccess] = useState(false)

  useEffect(() => {
    async function load() {
      const data = await getSsoProviders()
      setProviders(data)
    }
    load()
  }, [])

  const handleToggle = (id: string) => {
    setProviders((prev) =>
      prev.map((p) => (p.id === id ? { ...p, enabled: !p.enabled } : p))
    )
  }

  const handleSave = () => {
    setSaving(true)
    setTimeout(() => {
      setSaving(false)
      setSavedSuccess(true)
      setTimeout(() => setSavedSuccess(false), 2500)
    }, 600)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="p-5 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex items-center justify-center w-11 h-11 rounded-xl bg-sky-500/10 border border-sky-500/30 text-sky-400">
            <KeyRound className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-base font-semibold text-white">Enterprise Single Sign-On (SSO) & SCIM</h3>
            <p className="text-xs text-slate-400">
              Integrate with corporate identity providers like Okta, Keycloak, Azure AD/Microsoft Entra, or Google Workspace.
            </p>
          </div>
        </div>
        <Badge variant="outline" className="text-xs text-slate-300">
          SAML 2.0 & OIDC
        </Badge>
      </div>

      <div className="space-y-4">
        {providers.map((p) => (
          <Card key={p.id} className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
            <CardHeader className="p-0">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2.5">
                  <CardTitle className="text-base text-white">{p.name}</CardTitle>
                  <Badge variant={p.enabled ? 'success' : 'default'} className="text-[10px] py-0">
                    {p.enabled ? 'Enabled' : 'Disabled'}
                  </Badge>
                  <Badge variant="outline" className="text-[10px] py-0 uppercase">
                    {p.type}
                  </Badge>
                </div>

                <div className="flex items-center gap-2">
                  <label className="text-xs text-slate-400">Enable Provider</label>
                  <input
                    type="checkbox"
                    checked={p.enabled}
                    onChange={() => handleToggle(p.id)}
                    className="w-4 h-4 rounded text-sky-500 accent-sky-500 cursor-pointer"
                  />
                </div>
              </div>
              <CardDescription>
                Federate corporate employee and agent identities securely.
              </CardDescription>
            </CardHeader>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-2">
              <div className="space-y-1">
                <label className="text-xs font-medium text-slate-400">Identity Provider Issuer URL</label>
                <Input
                  value={p.issuerUrl}
                  readOnly
                  className="font-mono text-xs text-slate-300 bg-slate-950"
                />
              </div>

              <div className="space-y-1">
                <label className="text-xs font-medium text-slate-400">Single Sign-On (SSO) Target URL</label>
                <Input
                  value={p.ssoUrl}
                  readOnly
                  className="font-mono text-xs text-slate-300 bg-slate-950"
                />
              </div>

              {p.certificateFingerprint && (
                <div className="space-y-1 md:col-span-2">
                  <label className="text-xs font-medium text-slate-400">X.509 Certificate Fingerprint</label>
                  <Input
                    value={p.certificateFingerprint}
                    readOnly
                    className="font-mono text-xs text-slate-400 bg-slate-950"
                  />
                </div>
              )}
            </div>

            {/* SCIM Provisioning Toggle */}
            <div className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800 text-xs">
              <div className="space-y-0.5">
                <span className="font-semibold text-slate-200">SCIM 2.0 User Lifecycle Provisioning</span>
                <p className="text-[11px] text-slate-500">
                  Automatically sync employee on-boarding, department changes, and off-boarding deactivation.
                </p>
              </div>
              <Badge variant={p.scimEnabled ? 'success' : 'outline'} className="text-[10px] py-0">
                {p.scimEnabled ? 'SCIM Active' : 'SCIM Inactive'}
              </Badge>
            </div>
          </Card>
        ))}

        {/* Action button */}
        <div className="flex items-center justify-between pt-2">
          <div className="text-xs text-slate-500 flex items-center gap-1.5">
            <ShieldCheck className="w-4 h-4 text-emerald-400" />
            <span>SAML metadata endpoint: <code className="text-slate-400 font-mono">/auth/saml/metadata</code></span>
          </div>

          <Button
            variant="glow"
            size="sm"
            onClick={handleSave}
            loading={saving}
            className="gap-1.5"
          >
            {savedSuccess ? (
              <>
                <CheckCircle2 className="w-3.5 h-3.5 text-emerald-300" />
                <span>Configuration Saved</span>
              </>
            ) : (
              <>
                <Save className="w-3.5 h-3.5" />
                <span>Save SSO Settings</span>
              </>
            )}
          </Button>
        </div>
      </div>
    </div>
  )
}
