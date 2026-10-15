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
import { Box, BoxHeader, BoxBody, BoxRow } from '@/components/ui/Card'
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
      ],
    })

    if (res.ok) {
      showFeedback('Production ruleset activated')
      loadData(true)
    } else {
      showFeedback('Failed to activate ruleset: ' + res.error)
      setRefreshing(false)
    }
  }

  const handleDeleteRuleset = async (id: string) => {
    const ok = await deleteRuleset(repo.ownerSlug, repo.name, id)
    if (ok) {
      setRulesets((prev) => prev.filter((r) => r.id !== id))
      showFeedback('Ruleset deleted')
    }
  }

  const handleToggleIpAllowlist = async () => {
    const next = !ipEnabled
    setIpEnabled(next)
    await setIPAllowlistEnabled(repo.ownerSlug, next)
    showFeedback(next ? 'IP allowlist enforced' : 'IP allowlist disabled')
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
    <div className="space-y-4">
      {/* Overview Header Banner */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        <div className="space-y-0.5">
          <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
            <ShieldCheck className="w-4 h-4 text-primary" />
            <span>Repository Security & Merge Gates</span>
          </h2>
          <p className="text-xs text-muted-foreground">
            Enforce cryptographic commit signatures, merge approvals, check suites, and corporate CIDR allowlists.
          </p>
        </div>

        <div className="flex items-center gap-2 self-start sm:self-auto">
          {actionMsg && (
            <Badge variant="outline" className="text-xs font-mono text-emerald-500 border-emerald-500/40 gap-1">
              <Check className="w-3 h-3" />
              <span>{actionMsg}</span>
            </Badge>
          )}

          <Button
            variant="outline"
            size="sm"
            onClick={() => loadData(true)}
            disabled={refreshing}
            className="text-xs gap-1.5"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${refreshing ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </Button>
        </div>
      </div>

      {loading ? (
        <Box className="p-8 text-center">
          <div className="flex items-center justify-center gap-2 text-xs text-muted-foreground font-mono">
            <RefreshCw className="w-4 h-4 animate-spin text-primary" />
            <span>Querying security policies and cryptographic rulesets...</span>
          </div>
        </Box>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {/* Branch Protection Rulesets */}
          <Box>
            <BoxHeader className="bg-secondary/35 py-2.5">
              <div className="flex items-center gap-2">
                <Lock className="w-4 h-4 text-primary" />
                <span className="font-semibold text-foreground text-xs">Branch Protection & Merge Gates</span>
              </div>
              <Badge variant="outline" className="text-[10px] font-mono">
                {rulesets.length} Active Ruleset{rulesets.length === 1 ? '' : 's'}
              </Badge>
            </BoxHeader>

            <BoxBody className="p-4 space-y-3">
              <p className="text-xs text-muted-foreground">
                Conditions required before commits and pull requests can merge into protected branches.
              </p>

              {rulesets.length === 0 ? (
                <div className="p-6 rounded-md bg-secondary/30 border border-border text-center space-y-3">
                  <Shield className="w-6 h-6 text-muted-foreground mx-auto" />
                  <div className="space-y-1">
                    <h4 className="text-xs font-semibold text-foreground">No Branch Rulesets Configured</h4>
                    <p className="text-xs text-muted-foreground max-w-sm mx-auto">
                      Branch merges are currently unguarded. Protect your production branch against unreviewed merges and unsigned commits.
                    </p>
                  </div>
                  <Button
                    variant="primary"
                    size="sm"
                    onClick={handleCreateDefaultRuleset}
                    className="gap-1.5 text-xs"
                  >
                    <ShieldCheck className="w-4 h-4" />
                    <span>Activate Production Ruleset</span>
                  </Button>
                </div>
              ) : (
                <div className="space-y-2.5">
                  {rulesets.map((rs) => (
                    <div key={rs.id} className="p-3 rounded-md bg-secondary/30 border border-border space-y-2.5">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-xs text-foreground">{rs.name}</span>
                          <Badge
                            variant={rs.enforcement === 'active' ? 'success' : 'outline'}
                            className="text-[9px] py-0 uppercase"
                          >
                            {rs.enforcement}
                          </Badge>
                        </div>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDeleteRuleset(rs.id)}
                          className="text-muted-foreground hover:text-destructive h-6 px-1.5"
                          title="Delete ruleset"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </Button>
                      </div>

                      <div className="flex items-center gap-1.5 text-[11px] font-mono text-muted-foreground">
                        <GitBranch className="w-3.5 h-3.5 text-primary" />
                        <span>Target Refs:</span>
                        <span className="text-foreground bg-secondary px-1.5 py-0.5 rounded border border-border">
                          {rs.includedRefNames && rs.includedRefNames.length > 0
                            ? rs.includedRefNames.join(', ')
                            : 'refs/heads/main'}
                        </span>
                      </div>

                      {/* Enforced Rules List */}
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-1.5 pt-1">
                        <div className="flex items-center gap-2 p-1.5 rounded bg-background border border-border text-[11px] text-foreground">
                          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                          <span>Require Signed Commits</span>
                        </div>
                        <div className="flex items-center gap-2 p-1.5 rounded bg-background border border-border text-[11px] text-foreground">
                          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                          <span>Require 1 PR Approval</span>
                        </div>
                        <div className="flex items-center gap-2 p-1.5 rounded bg-background border border-border text-[11px] text-foreground">
                          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                          <span>Require CI Check Runs</span>
                        </div>
                        <div className="flex items-center gap-2 p-1.5 rounded bg-background border border-border text-[11px] text-foreground">
                          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                          <span>Block Force Pushes</span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </BoxBody>
          </Box>

          {/* Inbound IP Allowlists (CIDR Filtering) */}
          <Box>
            <BoxHeader className="bg-secondary/35 py-2.5">
              <div className="flex items-center gap-2">
                <Globe className="w-4 h-4 text-primary" />
                <span className="font-semibold text-foreground text-xs">Inbound IP Allowlists (CIDR)</span>
              </div>
              <button
                onClick={handleToggleIpAllowlist}
                className={`relative inline-flex h-4 w-8 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                  ipEnabled ? 'bg-primary' : 'bg-secondary'
                }`}
                title={ipEnabled ? 'Disable IP Allowlist' : 'Enable IP Allowlist'}
              >
                <span
                  className={`inline-block h-3 w-3 transform rounded-full bg-white shadow transition duration-200 ease-in-out ${
                    ipEnabled ? 'translate-x-4' : 'translate-x-0'
                  }`}
                />
              </button>
            </BoxHeader>

            <BoxBody className="p-4 space-y-3">
              <p className="text-xs text-muted-foreground">
                Restrict Git Smart-HTTP, SSH, and MCP traffic to verified corporate CIDR subnets.
              </p>

              <div className="space-y-2">
                {ipEntries.length === 0 ? (
                  <div className="p-3 rounded-md bg-secondary/30 border border-border text-center text-xs text-muted-foreground">
                    No CIDR subnets configured. All inbound traffic is currently allowed.
                  </div>
                ) : (
                  ipEntries.map((item) => (
                    <div
                      key={item.id}
                      className="flex items-center justify-between p-2.5 rounded-md bg-secondary/30 border border-border"
                    >
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-xs font-semibold text-foreground">{item.cidr}</span>
                          {ipEnabled && (
                            <Badge variant="success" className="text-[9px] py-0">
                              Enforced
                            </Badge>
                          )}
                        </div>
                        <span className="text-[11px] text-muted-foreground">{item.description}</span>
                      </div>

                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDeleteCidr(item.id)}
                        className="text-muted-foreground hover:text-destructive h-6 px-1.5"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    </div>
                  ))
                )}

                {/* Add CIDR Form */}
                <div className="space-y-2 pt-2 border-t border-border">
                  <div className="flex items-center gap-2">
                    <Input
                      placeholder="e.g. 192.168.1.0/24"
                      value={newCidr}
                      onChange={(e) => setNewCidr(e.target.value)}
                      className="text-xs font-mono"
                    />
                    <Button variant="secondary" size="sm" onClick={handleAddCidr} className="shrink-0 gap-1 text-xs">
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
            </BoxBody>
          </Box>

          {/* SSH Certificate Authorities (SSH CA) */}
          <Box className="lg:col-span-2">
            <BoxHeader className="bg-secondary/35 py-2.5">
              <div className="flex items-center gap-2">
                <Key className="w-4 h-4 text-primary" />
                <span className="font-semibold text-foreground text-xs">SSH Certificate Authorities (SSH CA)</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs text-muted-foreground font-mono">Require SSH Certs:</span>
                <button
                  onClick={handleToggleSshRequired}
                  className={`relative inline-flex h-4 w-8 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                    sshRequired ? 'bg-primary' : 'bg-secondary'
                  }`}
                  title={sshRequired ? 'Disable Requirement' : 'Enforce SSH Certs'}
                >
                  <span
                    className={`inline-block h-3 w-3 transform rounded-full bg-white shadow transition duration-200 ease-in-out ${
                      sshRequired ? 'translate-x-4' : 'translate-x-0'
                    }`}
                  />
                </button>
              </div>
            </BoxHeader>

            <BoxBody className="p-4 space-y-3">
              <p className="text-xs text-muted-foreground">
                Permit developer authentication using short-lived cryptographically signed SSH certificates issued by HashiCorp Vault, Teleport, or corporate CA.
              </p>

              <div className="space-y-2.5">
                {sshCas.length === 0 ? (
                  <div className="p-3 rounded-md bg-secondary/30 border border-border text-center text-xs text-muted-foreground">
                    No SSH Certificate Authorities registered. Users authenticate via individual SSH public keys or PATs.
                  </div>
                ) : (
                  sshCas.map((ca) => (
                    <BoxRow
                      key={ca.id}
                      className="p-3 rounded-md bg-secondary/30 border border-border flex flex-col sm:flex-row sm:items-center justify-between gap-3"
                    >
                      <div className="space-y-0.5">
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-semibold text-foreground">{ca.name}</span>
                          <Badge variant="outline" className="text-[10px]">
                            {ca.key_type || 'ssh-ed25519'}
                          </Badge>
                        </div>
                        <p className="font-mono text-[11px] text-muted-foreground break-all">{ca.fingerprint}</p>
                      </div>

                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDeleteCa(ca.id)}
                        className="text-muted-foreground hover:text-destructive self-end sm:self-auto h-7 px-2"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    </BoxRow>
                  ))
                )}

                {/* Add SSH CA Form Toggle */}
                {showAddCa ? (
                  <div className="p-3 rounded-md bg-secondary/40 border border-border space-y-2.5 animate-in fade-in">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-semibold text-foreground">Enroll New SSH Certificate Authority</span>
                      <button onClick={() => setShowAddCa(false)} className="text-muted-foreground hover:text-foreground text-xs">
                        Cancel
                      </button>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
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
                    <Button variant="primary" size="sm" onClick={handleAddCa} className="gap-1.5 text-xs">
                      <Key className="w-3.5 h-3.5" />
                      <span>Save Certificate Authority</span>
                    </Button>
                  </div>
                ) : (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowAddCa(true)}
                    className="gap-1.5 text-xs"
                  >
                    <Plus className="w-3.5 h-3.5" />
                    <span>Enroll Trusted SSH CA</span>
                  </Button>
                )}

                <div className="flex items-center gap-2 text-xs text-muted-foreground bg-secondary/30 border border-border p-2 rounded-md">
                  <AlertTriangle className="w-4 h-4 text-amber-500 shrink-0" />
                  <span>Certificates signed by registered CAs automatically inherit namespace role permissions on push/clone.</span>
                </div>
              </div>
            </BoxBody>
          </Box>
        </div>
      )}
    </div>
  )
}
