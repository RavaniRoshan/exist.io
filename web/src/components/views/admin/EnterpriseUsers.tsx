import React, { useState, useEffect } from 'react'
import {
  Users,
  UserPlus,
  Mail,
  RefreshCw,
  Search,
} from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Dialog } from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { getEnterpriseUsers, createEnterpriseUser, type EnterpriseUser } from '@/lib/api'

export const EnterpriseUsers: React.FC = () => {
  const [users, setUsers] = useState<EnterpriseUser[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [inviteModalOpen, setInviteModalOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  // Form fields
  const [email, setEmail] = useState('')
  const [login, setLogin] = useState('')
  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<'admin' | 'member'>('member')

  useEffect(() => {
    loadUsers()
  }, [])

  const loadUsers = async () => {
    setLoading(true)
    const list = await getEnterpriseUsers()
    setUsers(list)
    setLoading(false)
  }

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email.trim() || !password.trim()) {
      setErrorMsg('Email and password are required')
      return
    }
    if (password.length < 8) {
      setErrorMsg('Password must be at least 8 characters long')
      return
    }

    setSubmitting(true)
    setErrorMsg(null)

    const res = await createEnterpriseUser({
      email: email.trim(),
      login: (login || email.split('@')[0]).trim(),
      name: (name || login || email.split('@')[0]).trim(),
      password: password,
      role: role,
    })

    setSubmitting(false)
    if (res.ok) {
      setInviteModalOpen(false)
      setEmail('')
      setLogin('')
      setName('')
      setPassword('')
      setRole('member')
      loadUsers()
    } else {
      setErrorMsg(res.error || 'Failed to create user')
    }
  }

  const filteredUsers = users.filter((u) => {
    const q = searchQuery.toLowerCase()
    return (
      u.name.toLowerCase().includes(q) ||
      u.email.toLowerCase().includes(q) ||
      u.username.toLowerCase().includes(q)
    )
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-5 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md">
        <div className="flex items-center gap-3.5">
          <div className="flex items-center justify-center w-11 h-11 rounded-xl bg-teal-500/10 border border-teal-500/30 text-teal-400">
            <Users className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="text-base font-semibold text-white">Enterprise User Management & RBAC</h3>
              <Badge variant="outline" className="text-[10px] text-slate-300">
                {users.length} {users.length === 1 ? 'account' : 'accounts'}
              </Badge>
            </div>
            <p className="text-xs text-slate-400 mt-0.5">
              Manage developer memberships, access roles, and platform administrator privileges.
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2.5">
          <Button variant="outline" size="sm" onClick={loadUsers} disabled={loading} className="gap-1.5">
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </Button>
          <Button variant="glow" size="sm" onClick={() => setInviteModalOpen(true)} className="gap-1.5">
            <UserPlus className="w-3.5 h-3.5" />
            <span>Add User</span>
          </Button>
        </div>
      </div>

      {/* Filter Toolbar */}
      <div className="flex items-center justify-between gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder="Search users by name, username, or email..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-3 py-1.5 rounded-xl bg-slate-900 border border-slate-800 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-sky-500"
          />
        </div>
      </div>

      {/* Users Table */}
      <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/70">
        {loading ? (
          <div className="p-12 text-center text-xs text-slate-500 font-mono flex items-center justify-center gap-2">
            <RefreshCw className="w-4 h-4 animate-spin text-sky-400" />
            <span>Loading user accounts...</span>
          </div>
        ) : filteredUsers.length === 0 ? (
          <div className="p-12 text-center space-y-3">
            <Users className="w-8 h-8 text-slate-600 mx-auto" />
            <h4 className="text-sm font-medium text-slate-300">No users found</h4>
            <p className="text-xs text-slate-500 max-w-sm mx-auto">
              {searchQuery ? 'No accounts matched your search query.' : 'Create an account or connect your SCIM/SSO provider to sync corporate users.'}
            </p>
            {!searchQuery && (
              <Button size="sm" variant="glow" onClick={() => setInviteModalOpen(true)} className="mt-2">
                Add First User
              </Button>
            )}
          </div>
        ) : (
          <div className="divide-y divide-slate-850/80">
            {filteredUsers.map((u) => (
              <div
                key={u.id}
                className="flex flex-col sm:flex-row sm:items-center justify-between p-4 hover:bg-slate-900/40 transition-colors gap-3"
              >
                <div className="flex items-center gap-3.5">
                  <LivingAvatar name={u.name} size="md" />
                  <div className="space-y-0.5">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="text-sm font-semibold text-white">{u.name}</span>
                      <span className="text-xs text-slate-500 font-mono">@{u.username}</span>
                      <Badge
                        variant={u.role === 'admin' ? 'agent' : 'outline'}
                        className="text-[10px] py-0 capitalize"
                      >
                        {u.role}
                      </Badge>
                      <Badge
                        variant={u.status === 'active' ? 'success' : 'closed'}
                        className="text-[10px] py-0 capitalize"
                      >
                        {u.status}
                      </Badge>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-slate-400">
                      <span className="flex items-center gap-1">
                        <Mail className="w-3 h-3 text-slate-500" />
                        <span>{u.email}</span>
                      </span>
                      <span>•</span>
                      <span className="text-slate-500 font-mono text-[11px]">{u.id}</span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2 self-end sm:self-auto">
                  <span className="text-[11px] text-slate-500">{u.lastActive}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>

      {/* Add User Modal */}
      <Dialog
        isOpen={inviteModalOpen}
        onClose={() => setInviteModalOpen(false)}
        title="Add Enterprise Member"
        description="Create an authenticated user account with role-based access controls."
      >
        <form onSubmit={handleCreateUser} className="space-y-4 pt-2">
          {errorMsg && (
            <div className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs">
              {errorMsg}
            </div>
          )}

          <div className="space-y-1.5">
            <label className="text-xs font-medium text-slate-300">Email Address *</label>
            <Input
              type="email"
              placeholder="developer@enterprise.internal"
              value={email}
              onChange={(e) => {
                setEmail(e.target.value)
                if (!login && e.target.value.includes('@')) {
                  setLogin(e.target.value.split('@')[0])
                }
              }}
              required
              className="bg-slate-900 border-slate-700 text-xs"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-slate-300">Username *</label>
              <Input
                placeholder="alice"
                value={login}
                onChange={(e) => setLogin(e.target.value)}
                required
                className="bg-slate-900 border-slate-700 text-xs font-mono"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-slate-300">Full Name</label>
              <Input
                placeholder="Alice Walker"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="bg-slate-900 border-slate-700 text-xs"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-medium text-slate-300">Initial Password * (min. 8 chars)</label>
            <Input
              type="password"
              placeholder="••••••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              className="bg-slate-900 border-slate-700 text-xs font-mono"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-medium text-slate-300">Platform Role</label>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => setRole('member')}
                className={`p-3 rounded-xl border text-left transition-all ${
                  role === 'member'
                    ? 'bg-sky-500/10 border-sky-500 text-sky-300'
                    : 'bg-slate-900 border-slate-800 text-slate-400 hover:text-white'
                }`}
              >
                <div className="text-xs font-semibold">Member</div>
                <div className="text-[10px] text-slate-500 mt-0.5">Read/write access to permitted repositories.</div>
              </button>

              <button
                type="button"
                onClick={() => setRole('admin')}
                className={`p-3 rounded-xl border text-left transition-all ${
                  role === 'admin'
                    ? 'bg-purple-500/10 border-purple-500 text-purple-300'
                    : 'bg-slate-900 border-slate-800 text-slate-400 hover:text-white'
                }`}
              >
                <div className="text-xs font-semibold">Administrator</div>
                <div className="text-[10px] text-slate-500 mt-0.5">Full control over users, cluster, SSO, and policies.</div>
              </button>
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-3 border-t border-slate-800">
            <Button type="button" variant="ghost" size="sm" onClick={() => setInviteModalOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" variant="glow" size="sm" disabled={submitting}>
              {submitting ? 'Creating...' : 'Create Account'}
            </Button>
          </div>
        </form>
      </Dialog>
    </div>
  )
}
