import React, { useState } from 'react'
import { motion } from 'motion/react'
import { Lock, Mail, User, ShieldCheck, KeyRound, AlertCircle, ArrowRight } from 'lucide-react'
import { Dialog } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { loginUser, registerUser, type UserProfile } from '@/lib/api'

interface AuthModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: (user: UserProfile) => void
}

export const AuthModal: React.FC<AuthModalProps> = ({ isOpen, onClose, onSuccess }) => {
  const [tab, setTab] = useState<'login' | 'register'>('login')
  const [email, setEmail] = useState('')
  const [login, setLogin] = useState('')
  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)

    try {
      if (tab === 'login') {
        const res = await loginUser(email, password)
        if (!res.ok || !res.user) {
          setError(res.error || 'Invalid email or password')
          setLoading(false)
          return
        }
        onSuccess(res.user)
        onClose()
      } else {
        const res = await registerUser(email, login || email.split('@')[0], name || login, password)
        if (!res.ok || !res.user) {
          setError(res.error || 'Registration failed')
          setLoading(false)
          return
        }
        onSuccess(res.user)
        onClose()
      }
    } catch (err: any) {
      setError(err.message || 'An unexpected authentication error occurred')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog
      isOpen={isOpen}
      onClose={onClose}
      title={tab === 'login' ? 'Sign In to exist.io' : 'Create Enterprise Account'}
      description={
        tab === 'login'
          ? 'Authenticate to access private repositories, submit reviews, and manage enterprise policies.'
          : 'Register a new developer account on your self-hosted forge instance.'
      }
    >
      <div className="space-y-4 pt-2">
        {/* Tab Switcher */}
        <div className="flex rounded-md bg-secondary p-0.5 border border-border text-xs font-medium">
          <button
            type="button"
            onClick={() => { setTab('login'); setError(null) }}
            className={`flex-1 py-1 rounded transition-colors ${
              tab === 'login' ? 'bg-background text-foreground font-semibold shadow-2xs' : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            Sign In
          </button>
          <button
            type="button"
            onClick={() => { setTab('register'); setError(null) }}
            className={`flex-1 py-1 rounded transition-colors ${
              tab === 'register' ? 'bg-background text-foreground font-semibold shadow-2xs' : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            Create Account
          </button>
        </div>

        {error && (
          <motion.div
            initial={{ opacity: 0, y: -4 }}
            animate={{ opacity: 1, y: 0 }}
            className="p-2.5 rounded-md bg-destructive/10 border border-destructive/30 flex items-start gap-2.5 text-xs text-destructive"
          >
            <AlertCircle className="w-4 h-4 text-destructive shrink-0 mt-0.5" />
            <span>{error}</span>
          </motion.div>
        )}

        <form onSubmit={handleSubmit} className="space-y-3.5">
          <div>
            <label className="block text-xs font-medium text-foreground mb-1">
              Email or Username
            </label>
            <div className="relative">
              <Mail className="w-4 h-4 text-muted-foreground absolute left-3 top-3" />
              <Input
                type="text"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="developer@company.com or username"
                className="pl-9"
              />
            </div>
          </div>

          {tab === 'register' && (
            <>
              <div>
                <label className="block text-xs font-medium text-foreground mb-1">
                  Login Username (Slug)
                </label>
                <div className="relative">
                  <User className="w-4 h-4 text-muted-foreground absolute left-3 top-3" />
                  <Input
                    type="text"
                    required
                    value={login}
                    onChange={(e) => setLogin(e.target.value)}
                    placeholder="e.g. alex"
                    className="pl-9"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-foreground mb-1">
                  Full Name
                </label>
                <div className="relative">
                  <ShieldCheck className="w-4 h-4 text-muted-foreground absolute left-3 top-3" />
                  <Input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Alex Morgan"
                    className="pl-9"
                  />
                </div>
              </div>
            </>
          )}

          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="block text-xs font-medium text-foreground">
                Password
              </label>
              {tab === 'register' && (
                <span className="text-[10px] text-muted-foreground">Min 8 characters</span>
              )}
            </div>
            <div className="relative">
              <Lock className="w-4 h-4 text-muted-foreground absolute left-3 top-3" />
              <Input
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••••••"
                className="pl-9"
              />
            </div>
          </div>

          <div className="pt-2">
            <Button
              type="submit"
              variant="primary"
              disabled={loading}
              className="w-full justify-center gap-2"
            >
              <KeyRound className="w-4 h-4" />
              <span>{loading ? 'Authenticating...' : tab === 'login' ? 'Sign In' : 'Create Account'}</span>
              <ArrowRight className="w-3.5 h-3.5 ml-1" />
            </Button>
          </div>
        </form>

        <div className="pt-2 border-t border-border text-center text-muted-foreground text-[11px]">
          <span>Protected by exist.io Argon2id credential hashing & session encryption</span>
        </div>
      </div>
    </Dialog>
  )
}
