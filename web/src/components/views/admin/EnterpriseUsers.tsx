import React, { useState, useEffect } from 'react'
import {
  Users,
  UserPlus,
  Mail,
  RefreshCw,
  Search,
} from 'lucide-react'
import { Box, BoxHeader, BoxBody, BoxRow } from '@/components/ui/Card'
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
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        <div className="space-y-0.5">
          <div className="flex items-center gap-2">
            <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
              <Users className="w-4 h-4 text-primary" />
              <span>Enterprise User Management & RBAC</span>
            </h2>
            <Badge variant="outline" className="text-[10px]">
              {users.length} {users.length === 1 ? 'account' : 'accounts'}
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground">
            Manage developer memberships, access roles, and platform administrator privileges.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={loadUsers} disabled={loading} className="gap-1.5 text-xs">
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </Button>
          <Button variant="primary" size="sm" onClick={() => setInviteModalOpen(true)} className="gap-1.5 text-xs">
            <UserPlus className="w-3.5 h-3.5" />
            <span>Add User</span>
          </Button>
        </div>
      </div>

      {/* Filter Toolbar */}
      <div className="flex items-center justify-between gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search users by name, username, or email..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full h-8 pl-8 pr-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
          />
        </div>
      </div>

      {/* Users Table */}
      <Box>
        <BoxHeader className="bg-secondary/35 py-2.5">
          <span className="font-semibold text-foreground text-xs">Registered Members</span>
          <span className="text-muted-foreground font-mono text-[11px]">{filteredUsers.length} accounts</span>
        </BoxHeader>

        <BoxBody className="p-0">
          {loading ? (
            <div className="p-8 text-center text-xs text-muted-foreground font-mono flex items-center justify-center gap-2">
              <RefreshCw className="w-4 h-4 animate-spin text-primary" />
              <span>Loading user accounts...</span>
            </div>
          ) : filteredUsers.length === 0 ? (
            <div className="p-8 text-center space-y-3">
              <Users className="w-6 h-6 text-muted-foreground mx-auto" />
              <h4 className="text-xs font-semibold text-foreground">No users found</h4>
              <p className="text-xs text-muted-foreground max-w-sm mx-auto">
                {searchQuery ? 'No accounts matched your search query.' : 'Create an account or connect your SCIM/SSO provider to sync corporate users.'}
              </p>
              {!searchQuery && (
                <Button size="sm" variant="primary" onClick={() => setInviteModalOpen(true)} className="mt-2 text-xs">
                  Add First User
                </Button>
              )}
            </div>
          ) : (
            filteredUsers.map((u) => (
              <BoxRow
                key={u.id}
                className="flex flex-col sm:flex-row sm:items-center justify-between p-3.5 hover:bg-secondary/20 transition-colors gap-3"
              >
                <div className="flex items-center gap-3">
                  <LivingAvatar name={u.name} size="md" />
                  <div className="space-y-0.5">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="text-sm font-semibold text-foreground">{u.name}</span>
                      <span className="text-xs text-muted-foreground font-mono">@{u.username}</span>
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
                    <div className="flex items-center gap-2.5 text-xs text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <Mail className="w-3 h-3 text-muted-foreground" />
                        <span>{u.email}</span>
                      </span>
                      <span>•</span>
                      <span className="font-mono text-[11px]">{u.id}</span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2 self-end sm:self-auto">
                  <span className="text-[11px] text-muted-foreground">{u.lastActive}</span>
                </div>
              </BoxRow>
            ))
          )}
        </BoxBody>
      </Box>

      {/* Add User Modal */}
      <Dialog
        isOpen={inviteModalOpen}
        onClose={() => setInviteModalOpen(false)}
        title="Add Enterprise Member"
        description="Create an authenticated user account with role-based access controls."
      >
        <form onSubmit={handleCreateUser} className="space-y-3 pt-2">
          {errorMsg && (
            <div className="p-2.5 rounded-md bg-destructive/10 border border-destructive/30 text-destructive text-xs">
              {errorMsg}
            </div>
          )}

          <div className="space-y-1">
            <label className="text-xs font-medium text-foreground">Email Address *</label>
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
              className="text-xs"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label className="text-xs font-medium text-foreground">Username *</label>
              <Input
                placeholder="alice"
                value={login}
                onChange={(e) => setLogin(e.target.value)}
                required
                className="text-xs font-mono"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium text-foreground">Full Name</label>
              <Input
                placeholder="Alice Walker"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="text-xs"
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-foreground">Initial Password * (min. 8 chars)</label>
            <Input
              type="password"
              placeholder="••••••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              className="text-xs font-mono"
            />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-foreground">Platform Role</label>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => setRole('member')}
                className={`p-2.5 rounded-md border text-left transition-colors ${
                  role === 'member'
                    ? 'bg-secondary border-primary text-foreground'
                    : 'bg-card border-border text-muted-foreground hover:text-foreground'
                }`}
              >
                <div className="text-xs font-semibold">Member</div>
                <div className="text-[10px] text-muted-foreground mt-0.5">Read/write access to permitted repositories.</div>
              </button>

              <button
                type="button"
                onClick={() => setRole('admin')}
                className={`p-2.5 rounded-md border text-left transition-colors ${
                  role === 'admin'
                    ? 'bg-secondary border-primary text-foreground'
                    : 'bg-card border-border text-muted-foreground hover:text-foreground'
                }`}
              >
                <div className="text-xs font-semibold">Administrator</div>
                <div className="text-[10px] text-muted-foreground mt-0.5">Full control over users, cluster, SSO, and policies.</div>
              </button>
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-3 border-t border-border">
            <Button type="button" variant="outline" size="sm" onClick={() => setInviteModalOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" variant="primary" size="sm" disabled={submitting}>
              {submitting ? 'Creating...' : 'Create Account'}
            </Button>
          </div>
        </form>
      </Dialog>
    </div>
  )
}
