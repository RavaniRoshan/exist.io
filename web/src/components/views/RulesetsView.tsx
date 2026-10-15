import React, { useState } from 'react'
import { ShieldCheck, Lock, Globe, Key, Plus, CheckCircle2, AlertTriangle } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Input } from '@/components/ui/Input'

export const RulesetsView: React.FC = () => {
  const [requireReviews, setRequireReviews] = useState(true)
  const [requireChecks, setRequireChecks] = useState(true)
  const [blockForcePush, setBlockForcePush] = useState(true)
  const [newCidr, setNewCidr] = useState('')
  const [cidrs, setCidrs] = useState([
    { id: '1', cidr: '10.0.0.0/8', description: 'Internal VPC / Container Mesh', active: true },
    { id: '2', cidr: '192.168.1.0/24', description: 'On-premise Engineering Subnet', active: true },
    { id: '3', cidr: '172.16.0.0/12', description: 'Agent Sandbox Isolated Bridge', active: true },
  ])

  const handleAddCidr = () => {
    if (!newCidr.trim()) return
    setCidrs([
      ...cidrs,
      {
        id: String(Date.now()),
        cidr: newCidr.trim(),
        description: 'Custom allowed subnet',
        active: true,
      },
    ])
    setNewCidr('')
  }

  return (
    <div className="space-y-6">
      {/* Overview Header */}
      <div className="p-4 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-xl bg-purple-500/10 border border-purple-500/30 text-purple-400">
            <ShieldCheck className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-white">Repository Security & Merge Gates</h3>
            <p className="text-xs text-slate-400">
              Enforce deterministic merge policies, branch protection, and network allowlists.
            </p>
          </div>
        </div>
        <Badge variant="success" pulse>
          Enforcing Policy
        </Badge>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Branch Protection Ruleset */}
        <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
          <CardHeader className="p-0">
            <div className="flex items-center justify-between">
              <CardTitle className="text-base text-white flex items-center gap-2">
                <Lock className="w-4 h-4 text-sky-400" />
                <span>Branch Protection: main</span>
              </CardTitle>
              <Badge variant="outline" className="text-[10px] text-sky-300 border-sky-500/40">
                Default Branch
              </Badge>
            </div>
            <CardDescription>
              Rules that must be met before branches can be merged into production.
            </CardDescription>
          </CardHeader>

          <div className="space-y-3 pt-2">
            {/* Rule 1 */}
            <div className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800">
              <div className="space-y-0.5">
                <span className="text-xs font-semibold text-slate-200">Require Pull Request Reviews</span>
                <p className="text-[11px] text-slate-500">Require at least 1 approved review before merging</p>
              </div>
              <input
                type="checkbox"
                checked={requireReviews}
                onChange={(e) => setRequireReviews(e.target.checked)}
                className="w-4 h-4 rounded text-sky-500 accent-sky-500 cursor-pointer"
              />
            </div>

            {/* Rule 2 */}
            <div className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800">
              <div className="space-y-0.5">
                <span className="text-xs font-semibold text-slate-200">Require Passing Checks</span>
                <p className="text-[11px] text-slate-500">Block merge if CI/CD check suites fail</p>
              </div>
              <input
                type="checkbox"
                checked={requireChecks}
                onChange={(e) => setRequireChecks(e.target.checked)}
                className="w-4 h-4 rounded text-sky-500 accent-sky-500 cursor-pointer"
              />
            </div>

            {/* Rule 3 */}
            <div className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800">
              <div className="space-y-0.5">
                <span className="text-xs font-semibold text-slate-200">Block Force Pushes</span>
                <p className="text-[11px] text-slate-500">Prevent destructive git push --force on main</p>
              </div>
              <input
                type="checkbox"
                checked={blockForcePush}
                onChange={(e) => setBlockForcePush(e.target.checked)}
                className="w-4 h-4 rounded text-sky-500 accent-sky-500 cursor-pointer"
              />
            </div>
          </div>
        </Card>

        {/* Inbound IP Allowlists (CIDR Filtering) */}
        <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
          <CardHeader className="p-0">
            <CardTitle className="text-base text-white flex items-center gap-2">
              <Globe className="w-4 h-4 text-purple-400" />
              <span>Inbound IP Allowlists (CIDR)</span>
            </CardTitle>
            <CardDescription>
              Restrict Git Smart-HTTP, SSH, and MCP traffic to verified corporate CIDR subnets.
            </CardDescription>
          </CardHeader>

          <div className="space-y-2.5">
            {cidrs.map((item) => (
              <div
                key={item.id}
                className="flex items-center justify-between p-3 rounded-xl bg-slate-900/60 border border-slate-800"
              >
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-xs font-semibold text-sky-300">{item.cidr}</span>
                    <Badge variant="success" className="text-[9px] py-0">Enforced</Badge>
                  </div>
                  <span className="text-[11px] text-slate-500">{item.description}</span>
                </div>
                <CheckCircle2 className="w-4 h-4 text-emerald-400" />
              </div>
            ))}

            {/* Add CIDR Bar */}
            <div className="flex items-center gap-2 pt-2">
              <Input
                placeholder="e.g. 192.168.10.0/24"
                value={newCidr}
                onChange={(e) => setNewCidr(e.target.value)}
                className="text-xs font-mono"
              />
              <Button variant="secondary" size="sm" onClick={handleAddCidr} className="shrink-0 gap-1">
                <Plus className="w-3.5 h-3.5" />
                <span>Add CIDR</span>
              </Button>
            </div>
          </div>
        </Card>

        {/* SSH Certificate Authorities */}
        <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4 lg:col-span-2">
          <CardHeader className="p-0">
            <CardTitle className="text-base text-white flex items-center gap-2">
              <Key className="w-4 h-4 text-emerald-400" />
              <span>SSH Certificate Authorities (SSH CA)</span>
            </CardTitle>
            <CardDescription>
              Permit developer authentication using short-lived cryptographically signed SSH certificates.
            </CardDescription>
          </CardHeader>

          <div className="p-4 rounded-xl bg-slate-900/50 border border-slate-800 space-y-3">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-slate-200">HashiCorp Vault Root CA</span>
                  <Badge variant="outline" className="text-[10px] text-emerald-400 border-emerald-500/40">
                    Trusted CA
                  </Badge>
                </div>
                <p className="font-mono text-[11px] text-slate-400">
                  ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIH6... (vault-prod-ca)
                </p>
              </div>
              <span className="text-xs text-slate-500 font-mono">Expires in 342 days</span>
            </div>
            <div className="flex items-center gap-2 text-xs text-amber-400/90 bg-amber-950/20 border border-amber-900/40 p-2.5 rounded-lg">
              <AlertTriangle className="w-4 h-4 shrink-0" />
              <span>Certificates signed by this CA automatically map to namespace role permissions.</span>
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}
