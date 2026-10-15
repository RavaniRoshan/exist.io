import React, { useState, useEffect } from 'react'
import { Users, UserPlus, Key, ShieldCheck, Mail, CheckCircle2 } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Dialog } from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { getEnterpriseUsers, type EnterpriseUser } from '@/lib/api'

export const EnterpriseUsers: React.FC = () => {
  const [users, setUsers] = useState<EnterpriseUser[]>([])
  const [inviteModalOpen, setInviteModalOpen] = useState(false)
  const [newEmail, setNewEmail] = useState('')
  const [newRole, setNewRole] = useState<'admin' | 'maintainer' | 'contributor' | 'auditor'>('contributor')
  const [invitedSuccess, setInvitedSuccess] = useState(false)

  useEffect(() => {
    async function load() {
      const data = await getEnterpriseUsers()
      setUsers(data)
    }
    load()
  }, [])

  const handleInvite = () => {
    if (!newEmail.trim()) return
    const newUser: EnterpriseUser = {
      id: `u-${Date.now()}`,
      username: newEmail.split('@')[0],
      email: newEmail.trim(),
      name: newEmail.split('@')[0],
      role: newRole,
      isAgent: false,
      status: 'invited',
      twoFactorEnabled: false,
      lastActive: 'Invited just now',
      sshKeyCount: 0,
      patCount: 0,
    }
    setUsers((prev) => [...prev, newUser])
    setNewEmail('')
    setInvitedSuccess(true)
    setTimeout(() => {
      setInvitedSuccess(false)
      setInviteModalOpen(false)
    }, 1200)
  }

  const roleBadgeStyle = {
    owner: 'bg-purple-500/20 text-purple-300 border-purple-500/40',
    admin: 'bg-sky-500/20 text-sky-300 border-sky-500/40',
    maintainer: 'bg-teal-500/20 text-teal-300 border-teal-500/40',
    contributor: 'bg-slate-800 text-slate-300 border-slate-700',
    auditor: 'bg-amber-500/20 text-amber-300 border-amber-500/40',
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-5 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md">
        <div className="flex items-center gap-3">
          <div className="flex items-center justify-center w-11 h-11 rounded-xl bg-teal-500/10 border border-teal-500/30 text-teal-400">
            <Users className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-base font-semibold text-white">Enterprise User Roster & RBAC</h3>
            <p className="text-xs text-slate-400">
              Manage team memberships, role hierarchy, 2FA enforcement, and autonomous agent credentials.
            </p>
          </div>
        </div>

        <Button
          variant="glow"
          size="sm"
          onClick={() => setInviteModalOpen(true)}
          className="gap-1.5 self-start sm:self-auto"
        >
          <UserPlus className="w-3.5 h-3.5" />
          <span>Invite Member</span>
        </Button>
      </div>

      {/* Users Table Card */}
      <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/60">
        <div className="divide-y divide-slate-850/70">
          {users.map((u) => (
            <div
              key={u.id}
              className="flex flex-col sm:flex-row sm:items-center justify-between p-4 hover:bg-slate-900/40 transition-colors gap-3"
            >
              <div className="flex items-center gap-3.5">
                <LivingAvatar name={u.name} isAgent={u.isAgent} size="md" />
                <div className="space-y-0.5">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-sm font-semibold text-white">{u.name}</span>
                    <span className="text-xs text-slate-500 font-mono">@{u.username}</span>
                    <span className={`text-[10px] px-2 py-0.2 rounded-full font-medium border uppercase tracking-wider ${roleBadgeStyle[u.role]}`}>
                      {u.role}
                    </span>
                    {u.isAgent && (
                      <Badge variant="agent" className="text-[9px] py-0">
                        Agent Account
                      </Badge>
                    )}
                  </div>
                  <div className="text-xs text-slate-400 flex items-center gap-3">
                    <span>{u.email}</span>
                    <span>•</span>
                    <span className="text-slate-500">{u.lastActive}</span>
                  </div>
                </div>
              </div>

              {/* Status and Credentials stats */}
              <div className="flex items-center gap-3 self-end sm:self-auto text-xs font-mono text-slate-400">
                {u.twoFactorEnabled ? (
                  <Badge variant="success" className="text-[10px] py-0 gap-1">
                    <ShieldCheck className="w-3 h-3" /> 2FA On
                  </Badge>
                ) : (
                  <Badge variant="outline" className="text-[10px] py-0 text-slate-500">
                    2FA Off
                  </Badge>
                )}

                <div className="flex items-center gap-1.5 text-slate-400 text-xs">
                  <Key className="w-3.5 h-3.5 text-slate-500" />
                  <span>{u.sshKeyCount} keys</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </Card>

      {/* Invite Member Dialog */}
      <Dialog
        isOpen={inviteModalOpen}
        onClose={() => setInviteModalOpen(false)}
        title="Invite Enterprise Member or Agent"
        description="Send an invitation link with role-based permissions."
      >
        <div className="space-y-4 pt-2">
          {invitedSuccess ? (
            <div className="p-6 text-center space-y-2">
              <CheckCircle2 className="w-8 h-8 text-emerald-400 mx-auto" />
              <h4 className="text-sm font-semibold text-white">Invitation Dispatched</h4>
              <p className="text-xs text-slate-400">An invitation email has been sent.</p>
            </div>
          ) : (
            <>
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">Corporate Email Address</label>
                <Input
                  type="email"
                  placeholder="engineer@company.com"
                  value={newEmail}
                  onChange={(e) => setNewEmail(e.target.value)}
                  icon={<Mail className="w-4 h-4 text-slate-400" />}
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">Role Permissions</label>
                <select
                  value={newRole}
                  onChange={(e: any) => setNewRole(e.target.value)}
                  className="w-full bg-slate-900 text-slate-200 text-xs rounded-xl border border-slate-800 p-2.5 outline-none focus:border-sky-500"
                >
                  <option value="admin">Admin — Full organization control & user management</option>
                  <option value="maintainer">Maintainer — Branch protection rules & merge permissions</option>
                  <option value="contributor">Contributor — Clone, push feature branches, open PRs</option>
                  <option value="auditor">Auditor — Read-only compliance audit inspection</option>
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <Button variant="ghost" size="sm" onClick={() => setInviteModalOpen(false)}>
                  Cancel
                </Button>
                <Button
                  variant="glow"
                  size="sm"
                  onClick={handleInvite}
                  disabled={!newEmail.trim()}
                >
                  Send Invitation
                </Button>
              </div>
            </>
          )}
        </div>
      </Dialog>
    </div>
  )
}
