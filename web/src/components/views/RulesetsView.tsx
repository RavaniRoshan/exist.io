import React, { useState, useEffect } from 'react'
import {
  ShieldCheck,
  Lock,
  Globe,
  Key,
  Plus,
  CheckCircle2,
  AlertTriangle,
  Trash2,
  RefreshCw,
  GitBranch,
  Shield,
  Check,
} from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Input } from '@/components/ui/Input'
import {
  fetchRulesets,
  createRuleset,
  deleteRuleset,
  fetchIPAllowlist,
  setIPAllowlistEnabled,
  addIPAllowlistEntry,
  deleteIPAllowlistEntry,
  fetchSSHCAs,
  addSSHCA,
  deleteSSHCA,
  setSSHCARequirement,
  defaultRepo,
  type RulesetRecord,
  type IPAllowlistEntry,
  type SSHCARecord,
  type RepoInfo,
} from '@/lib/api'

interface RulesetsViewProps {
  repo?: RepoInfo
}

export const RulesetsView: React.FC<RulesetsViewProps> = ({ repo = defaultRepo }) => {
  const [rulesets, setRulesets] = useState<RulesetRecord[]>([])
  const [ipEnabled, setIpEnabled] = useState(false)
  const [ipEntries, setIpEntries] = useState<IPAllowlistEntry[]>([])
  const [sshRequired, setSshRequired] = useState(false)
  const [sshCas, setSshCas] = useState<SSHCARecord[]>([])

  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [newCidr, setNewCidr] = useState('')
  const [newCidrDesc, setNewCidrDesc] = useState('')
  const [newCaName, setNewCaName] = useState('')
  const [newCaKey, setNewCaKey] = useState('')
  const [showAddCa, setShowAddCa] = useState(false)
  const [actionMsg, setActionMsg] = useState<string | null>(null)

  const showFeedback = (msg: string) => {
    setActionMsg(msg)
    setTimeout(() => setActionMsg(null), 3000)
  }

  const loadData = async (isManual = false) => {
    if (isManual) setRefreshing(true)
    else setLoading(true)

    try {
      const [rs, ipData, caData] = await Promise.all([
        fetchRulesets(repo.ownerSlug, repo.name),
        fetchIPAllowlist(repo.ownerSlug),
        fetchSSHCAs(repo.ownerSlug),
      ])
      setRulesets(rs)
      setIpEnabled(ipData.enabled)
      setIpEntries(ipData.entries)
      setSshRequired(caData.required)
      setSshCas(caData.cas)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [repo.ownerSlug, repo.name])

  // Quick setup: Create standard production protection ruleset
  const handleCreateDefaultRuleset = async () => {
    setRefreshing(true)
    const res = await createRuleset(repo.ownerSlug, repo.name, {
      name: 'Mainline Production Policy',
      enforcement: 'active',
      kind: 'branch',
      includedRefNames: ['refs/heads/main'],
      rules: [
        { type: 'pull_request', parameters: { required_approving_review_count: 1 } },
        { type: 'required_status_checks', parameters: { checks: ['Go Build & Test Matrix'] } },
        { type: 'required_signatures', parameters: {} },
        { type: 'non_fast_forward', parameters: {} },
      ],
    })
    setRefreshing(false)
    if (res.ok) {
      showFeedback('Production ruleset activated on branch main')
      loadData(true)
    } else {
      showFeedback('Error: ' + res.error)
    }
  }

  const handleDeleteRuleset = async (id: string) => {
    const ok = await deleteRuleset(repo.ownerSlug, repo.name, id)
    if (ok) {
      showFeedback('Ruleset removed')
      loadData(true)
    }
  }

  const handleToggleIpAllowlist = async () => {
    const next = !ipEnabled
    setIpEnabled(next)
    await setIPAllowlistEnabled(repo.ownerSlug, next)
    showFeedback(next ? 'IP allowlist enforcement enabled' : 'IP allowlist enforcement disabled')
  }

  const handleAddCidr = async () => {
    if (!newCidr.trim()) return
    const res = await addIPAllowlistEntry(
      repo.ownerSlug,
      newCidr.trim(),
      newCidrDesc.trim() || 'Corporate allowed CIDR subnet'
    )
    if (res.ok) {
      setNewCidr('')
      setNewCidrDesc('')
      showFeedback('CIDR allowlist entry created')
      loadData(true)
    } else {
      showFeedback('Failed to add CIDR: ' + res.error)
    }
  }

  const handleDeleteCidr = async (id: string) => {
    const ok = await deleteIPAllowlistEntry(repo.ownerSlug, id)
    if (ok) {
      setIpEntries((prev) => prev.filter((e) => e.id !== id))
      showFeedback('CIDR entry deleted')
    }
  }

  const handleToggleSshRequired = async () => {
    const next = !sshRequired
    setSshRequired(next)
    await setSSHCARequirement(repo.ownerSlug, next)
    showFeedback(next ? 'Mandatory SSH certificates enforced' : 'SSH certificate requirement disabled')
  }

  const handleAddCa = async () => {
    if (!newCaName.trim() || !newCaKey.trim()) return
    const res = await addSSHCA(repo.ownerSlug, newCaName.trim(), newCaKey.trim())
    if (res.ok) {
      setNewCaName('')
      setNewCaKey('')
      setShowAddCa(false)
      showFeedback('Trusted SSH Certificate Authority enrolled')
      loadData(true)
    } else {
      showFeedback('Failed to add CA: ' + res.error)
    }
  }

  const handleDeleteCa = async (id: string) => {
    const ok = await deleteSSHCA(repo.ownerSlug, id)
    if (ok) {
      setSshCas((prev) => prev.filter((c) => c.id !== id))
      showFeedback('SSH CA removed')
    }
  }

  return (
    <div className="space-y-6">
      {/* Overview Header Banner */}
      <div className="p-4 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-purple-500/10 border border-purple-500/30 text-purple-400">
            <ShieldCheck className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-white">Repository Security & Merge Gates</h3>
            <p className="text-xs text-slate-400">
              Enforce cryptographic commit signatures, merge approvals, check suites, and corporate CIDR allowlists.
            </p>
          </div>
        </div>

        <div className="flex items-center gap-3 self-start sm:self-auto">
          {actionMsg && (
            <Badge variant="outline" className="text-xs font-mono text-emerald-400 border-emerald-500/40 gap-1 animate-in fade-in">
              <Check className="w-3 h-3" />
              <span>{actionMsg}</span>
            </Badge>
          )}

          <Button
            variant="outline"
            size="xs"
            onClick={() => loadData(true)}
            loading={refreshing}
            className="text-xs text-slate-300 gap-1.5"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            <span>Refresh</span>
          </Button>
        </div>
      </div>

      {loading ? (
        <Card className="p-8 text-center border-slate-800 bg-slate-950/50">
          <div className="flex items-center justify-center gap-2 text-xs text-slate-400 font-mono">
            <RefreshCw className="w-4 h-4 animate-spin text-purple-400" />
            <span>Querying security policies and cryptographic rulesets...</span>
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Branch Protection Rulesets */}
          <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
            <CardHeader className="p-0">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base text-white flex items-center gap-2">
                  <Lock className="w-4 h-4 text-sky-400" />
                  <span>Branch Protection & Merge Gates</span>
                </CardTitle>
                <Badge variant="outline" className="text-[10px] text-sky-300 border-sky-500/40 font-mono">
                  {rulesets.length} Active Ruleset{rulesets.length === 1 ? '' : 's'}
                </Badge>
              </div>
              <CardDescription>
                Conditions required before commits and pull requests can merge into protected branches.
              </CardDescription>
            </CardHeader>

            {rulesets.length === 0 ? (
              /* Authentic Empty State with 1-click CTA */
              <div className="p-6 rounded-xl bg-slate-900/40 border border-slate-800 text-center space-y-3">
                <Shield className="w-8 h-8 text-slate-600 mx-auto" />
                <div className="space-y-1">
                  <h4 className="text-sm font-semibold text-slate-300">No Branch Rulesets Configured</h4>
                  <p className="text-xs text-slate-500 max-w-sm mx-auto">
                    Branch merges are currently unguarded. Protect your production branch against unreviewed merges and unsigned commits.
                  </p>
                </div>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={handleCreateDefaultRuleset}
                  className="gap-1.5 shadow-[0_0_15px_rgba(56,189,248,0.25)]"
                >
                  <ShieldCheck className="w-4 h-4" />
                  <span>Activate Production Ruleset</span>
                </Button>
              </div>
            ) : (
              <div className="space-y-3 pt-1">
                {rulesets.map((rs) => (
                  <div key={rs.id} className="p-4 rounded-xl bg-slate-900/60 border border-slate-800 space-y-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-xs text-slate-200">{rs.name}</span>
                        <Badge
                          variant={rs.enforcement === 'active' ? 'success' : 'outline'}
                          className="text-[9px] py-0 uppercase"
                        >
                          {rs.enforcement}
                        </Badge>
                      </div>
                      <Button
                        variant="ghost"
                        size="xs"
                        onClick={() => handleDeleteRuleset(rs.id)}
                        className="text-slate-500 hover:text-rose-400 h-6 px-1.5"
                        title="Delete ruleset"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    </div>

                    <div className="flex items-center gap-1.5 text-[11px] font-mono text-slate-400">
                      <GitBranch className="w-3.5 h-3.5 text-sky-400" />
                      <span>Target Refs:</span>
                      <span className="text-sky-300 bg-slate-850 px-1.5 py-0.5 rounded border border-slate-800">
                        {rs.includedRefNames && rs.includedRefNames.length > 0
                          ? rs.includedRefNames.join(', ')
                          : 'refs/heads/main'}
                      </span>
                    </div>

                    {/* Enforced Rules List */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-1">
                      <div className="flex items-center gap-2 p-2 rounded-lg bg-slate-950/60 border border-slate-800 text-[11px] text-slate-300">
                        <CheckCircle2 className="w-3.5 h-3.5 text-teal-400 shrink-0" />
                        <span>Require Signed Commits</span>
                      </div>
                      <div className="flex items-center gap-2 p-2 rounded-lg bg-slate-950/60 border border-slate-800 text-[11px] text-slate-300">
                        <CheckCircle2 className="w-3.5 h-3.5 text-sky-400 shrink-0" />
                        <span>Require 1 PR Approval</span>
                      </div>
                      <div className="flex items-center gap-2 p-2 rounded-lg bg-slate-950/60 border border-slate-800 text-[11px] text-slate-300">
                        <CheckCircle2 className="w-3.5 h-3.5 text-purple-400 shrink-0" />
                        <span>Require CI Check Runs</span>
                      </div>
                      <div className="flex items-center gap-2 p-2 rounded-lg bg-slate-950/60 border border-slate-800 text-[11px] text-slate-300">
                        <CheckCircle2 className="w-3.5 h-3.5 text-amber-400 shrink-0" />
                        <span>Block Force Pushes</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Card>

          {/* Inbound IP Allowlists (CIDR Filtering) */}
          <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
            <CardHeader className="p-0">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base text-white flex items-center gap-2">
                  <Globe className="w-4 h-4 text-purple-400" />
                  <span>Inbound IP Allowlists (CIDR)</span>
                </CardTitle>
                <button
                  onClick={handleToggleIpAllowlist}
                  className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                    ipEnabled ? 'bg-purple-600' : 'bg-slate-800'
                  }`}
                  title={ipEnabled ? 'Disable IP Allowlist' : 'Enable IP Allowlist'}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white shadow transition duration-200 ease-in-out ${
                      ipEnabled ? 'translate-x-4' : 'translate-x-0'
                    }`}
                  />
                </button>
              </div>
              <CardDescription>
                Restrict Git Smart-HTTP, SSH, and MCP traffic to verified corporate CIDR subnets.
              </CardDescription>
            </CardHeader>

            <div className="space-y-2.5">
              {ipEntries.length === 0 ? (
                <div className="p-4 rounded-xl bg-slate-900/40 border border-slate-800 text-center text-xs text-slate-500">
                  No CIDR subnets configured. All inbound traffic is currently allowed.
                </div>
              ) : (
                ipEntries.map((item) => (
                  <div
                    key={item.id}
                    className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800"
                  >
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-xs font-semibold text-purple-300">{item.cidr}</span>
                        {ipEnabled && (
                          <Badge variant="success" className="text-[9px] py-0">
                            Enforced
                          </Badge>
                        )}
                      </div>
                      <span className="text-[11px] text-slate-500">{item.description}</span>
                    </div>

                    <Button
                      variant="ghost"
                      size="xs"
                      onClick={() => handleDeleteCidr(item.id)}
                      className="text-slate-500 hover:text-rose-400 h-6 px-1.5"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </div>
                ))
              )}

              {/* Add CIDR Form */}
              <div className="space-y-2 pt-2 border-t border-slate-850">
                <div className="flex items-center gap-2">
                  <Input
                    placeholder="e.g. 192.168.1.0/24"
                    value={newCidr}
                    onChange={(e) => setNewCidr(e.target.value)}
                    className="text-xs font-mono"
                  />
                  <Button variant="secondary" size="sm" onClick={handleAddCidr} className="shrink-0 gap-1">
                    <Plus className="w-3.5 h-3.5" />
                    <span>Add CIDR</span>
                  </Button>
                </div>
                <Input
                  placeholder="Description (e.g. Corporate VPN or Internal VPC)"
                  value={newCidrDesc}
                  onChange={(e) => setNewCidrDesc(e.target.value)}
                  className="text-xs"
                />
              </div>
            </div>
          </Card>

          {/* SSH Certificate Authorities (SSH CA) */}
          <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4 lg:col-span-2">
            <CardHeader className="p-0">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base text-white flex items-center gap-2">
                  <Key className="w-4 h-4 text-emerald-400" />
                  <span>SSH Certificate Authorities (SSH CA)</span>
                </CardTitle>
                <div className="flex items-center gap-3">
                  <span className="text-xs text-slate-400 font-mono">Require SSH Certs:</span>
                  <button
                    onClick={handleToggleSshRequired}
                    className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                      sshRequired ? 'bg-emerald-600' : 'bg-slate-800'
                    }`}
                    title={sshRequired ? 'Disable Requirement' : 'Enforce SSH Certs'}
                  >
                    <span
                      className={`inline-block h-4 w-4 transform rounded-full bg-white shadow transition duration-200 ease-in-out ${
                        sshRequired ? 'translate-x-4' : 'translate-x-0'
                      }`}
                    />
                  </button>
                </div>
              </div>
              <CardDescription>
                Permit developer authentication using short-lived cryptographically signed SSH certificates issued by HashiCorp Vault, Teleport, or corporate CA.
              </CardDescription>
            </CardHeader>

            <div className="space-y-3">
              {sshCas.length === 0 ? (
                <div className="p-4 rounded-xl bg-slate-900/40 border border-slate-800 text-center text-xs text-slate-500">
                  No SSH Certificate Authorities registered. Users authenticate via individual SSH public keys or PATs.
                </div>
              ) : (
                sshCas.map((ca) => (
                  <div
                    key={ca.id}
                    className="p-4 rounded-xl bg-slate-900/50 border border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-3"
                  >
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-semibold text-slate-200">{ca.name}</span>
                        <Badge variant="outline" className="text-[10px] text-emerald-400 border-emerald-500/40">
                          {ca.key_type || 'ssh-ed25519'}
                        </Badge>
                      </div>
                      <p className="font-mono text-[11px] text-slate-400 break-all">{ca.fingerprint}</p>
                    </div>

                    <Button
                      variant="ghost"
                      size="xs"
                      onClick={() => handleDeleteCa(ca.id)}
                      className="text-slate-500 hover:text-rose-400 self-end sm:self-auto h-7 px-2"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </div>
                ))
              )}

              {/* Add SSH CA Modal / Form Toggle */}
              {showAddCa ? (
                <div className="p-4 rounded-xl bg-slate-900/80 border border-emerald-500/30 space-y-3 animate-in fade-in">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-semibold text-slate-200">Enroll New SSH Certificate Authority</span>
                    <button onClick={() => setShowAddCa(false)} className="text-slate-500 hover:text-white text-xs">
                      Cancel
                    </button>
                  </div>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <Input
                      placeholder="CA Name (e.g. Vault Prod CA)"
                      value={newCaName}
                      onChange={(e) => setNewCaName(e.target.value)}
                      className="text-xs"
                    />
                    <Input
                      placeholder="Public Key (e.g. ssh-ed25519 AAAAC3...)"
                      value={newCaKey}
                      onChange={(e) => setNewCaKey(e.target.value)}
                      className="text-xs font-mono"
                    />
                  </div>
                  <Button variant="primary" size="sm" onClick={handleAddCa} className="gap-1.5">
                    <Key className="w-3.5 h-3.5" />
                    <span>Save Certificate Authority</span>
                  </Button>
                </div>
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowAddCa(true)}
                  className="gap-1.5 text-xs text-slate-300"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Enroll Trusted SSH CA</span>
                </Button>
              )}

              <div className="flex items-center gap-2 text-xs text-amber-400/90 bg-amber-950/20 border border-amber-900/40 p-2.5 rounded-lg">
                <AlertTriangle className="w-4 h-4 shrink-0" />
                <span>Certificates signed by registered CAs automatically inherit namespace role permissions on push/clone.</span>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}
