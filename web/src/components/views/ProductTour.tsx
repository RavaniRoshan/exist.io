import React, { useState } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import {
  Server,
  GitBranch,
  Bot,
  ShieldCheck,
  Users,
  CheckCircle2,
  ArrowRight,
  ArrowLeft,
  X,
  Copy,
  Check,
  BookOpen,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { cn } from '@/lib/utils'

export interface ProductTourProps {
  isOpen: boolean
  onClose: () => void
  onOpenManual?: () => void
}

interface TourStep {
  id: string
  title: string
  subtitle: string
  icon: React.ReactNode
  badge: string
  color: string
  content: React.ReactNode
}

export const ProductTour: React.FC<ProductTourProps> = ({
  isOpen,
  onClose,
  onOpenManual,
}) => {
  const [currentStepIndex, setCurrentStepIndex] = useState(0)
  const [copiedText, setCopiedText] = useState<string | null>(null)

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text)
    setCopiedText(text)
    setTimeout(() => setCopiedText(null), 2000)
  }

  const handleFinish = () => {
    localStorage.setItem('exist_onboarding_tour_completed', 'true')
    onClose()
  }

  const steps: TourStep[] = [
    {
      id: 'cluster-architecture',
      title: 'Sovereign Cluster Architecture',
      subtitle: 'Single-binary Git forge with zero external dependencies and SQLite WAL storage',
      icon: <Server className="w-5 h-5" />,
      badge: 'Core Engine',
      color: 'text-primary',
      content: (
        <div className="space-y-4">
          <p className="text-xs text-muted-foreground leading-relaxed">
            exist.io runs as a self-contained daemon (<code className="px-1.5 py-0.5 rounded bg-secondary text-foreground font-mono">existd</code>) that unifies repository storage, Git wire protocols, and agent orchestration into a single sovereign service.
          </p>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
            <div className="p-3 rounded-md bg-secondary/30 border border-border space-y-1">
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-semibold text-foreground uppercase">Smart-HTTP</span>
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              </div>
              <div className="font-mono text-xs text-primary font-bold">:8080</div>
              <p className="text-[10px] text-muted-foreground">Standard Git push/pull & REST API</p>
            </div>

            <div className="p-3 rounded-md bg-secondary/30 border border-border space-y-1">
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-semibold text-foreground uppercase">Pure-Go SSH</span>
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              </div>
              <div className="font-mono text-xs text-primary font-bold">:2222</div>
              <p className="text-[10px] text-muted-foreground">Embedded SSH transport (zero-OpenSSH)</p>
            </div>

            <div className="p-3 rounded-md bg-secondary/30 border border-border space-y-1">
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-semibold text-foreground uppercase">Storage Engine</span>
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              </div>
              <div className="font-mono text-xs text-primary font-bold">SQLite WAL</div>
              <p className="text-[10px] text-muted-foreground">ACID metadata & Litestream replicas</p>
            </div>
          </div>

          <div className="p-3 rounded-md bg-secondary/20 border border-border text-xs space-y-1.5">
            <div className="font-semibold text-foreground flex items-center gap-1.5">
              <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
              <span>How Everything Connects</span>
            </div>
            <p className="text-[11px] text-muted-foreground">
              All Git refs, pull request reviews, and security policies are coordinated directly through the SQLite write-ahead log (<code className="font-mono text-foreground">data/meta.db</code>). Bare Git repositories are stored in <code className="font-mono text-foreground">data/git/</code> with zero CGO dependencies.
            </p>
          </div>
        </div>
      ),
    },
    {
      id: 'git-transports',
      title: 'Connecting Your Git Clients',
      subtitle: 'Clone and push code over Smart-HTTP or native SSH with Ed25519 key authentication',
      icon: <GitBranch className="w-5 h-5" />,
      badge: 'Git Wire Protocols',
      color: 'text-primary',
      content: (
        <div className="space-y-4">
          <p className="text-xs text-muted-foreground leading-relaxed">
            exist.io speaks standard Git protocol v2. You can use your existing Git CLI, IDEs (VS Code, Cursor, JetBrains), and CI/CD pipelines without any plugins.
          </p>

          <div className="space-y-2.5">
            <div className="space-y-1">
              <span className="text-[11px] font-medium text-foreground">1. Clone or Push over Smart-HTTP (:8080)</span>
              <div className="relative flex items-center rounded-md bg-secondary/50 border border-border px-3 py-2 font-mono text-xs text-foreground">
                <span className="truncate pr-8">git remote add origin http://localhost:8080/shiva/exist.io.git</span>
                <button
                  onClick={() => handleCopy('git remote add origin http://localhost:8080/shiva/exist.io.git')}
                  className="absolute right-2 p-1 text-muted-foreground hover:text-foreground cursor-pointer"
                >
                  {copiedText?.includes('http://localhost:8080') ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
                </button>
              </div>
            </div>

            <div className="space-y-1">
              <span className="text-[11px] font-medium text-foreground">2. Clone or Push over Pure-Go SSH (:2222)</span>
              <div className="relative flex items-center rounded-md bg-secondary/50 border border-border px-3 py-2 font-mono text-xs text-foreground">
                <span className="truncate pr-8">git remote add origin ssh://git@localhost:2222/shiva/exist.io.git</span>
                <button
                  onClick={() => handleCopy('git remote add origin ssh://git@localhost:2222/shiva/exist.io.git')}
                  className="absolute right-2 p-1 text-muted-foreground hover:text-foreground cursor-pointer"
                >
                  {copiedText?.includes('ssh://git@localhost:2222') ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
                </button>
              </div>
            </div>
          </div>

          <div className="p-2.5 rounded-md bg-secondary/20 border border-border text-[11px] text-muted-foreground flex items-center justify-between">
            <span>Enrolled SSH keys are managed directly under your <strong>User Profile &gt; SSH and GPG keys</strong>.</span>
          </div>
        </div>
      ),
    },
    {
      id: 'agent-mcp-bridge',
      title: 'Autonomous AI Agent Bridge (MCP)',
      subtitle: 'Native Model Context Protocol gateway exposing 35+ Git primitives to AI coding agents',
      icon: <Bot className="w-5 h-5" />,
      badge: 'Agent Gateway',
      color: 'text-primary',
      content: (
        <div className="space-y-4">
          <p className="text-xs text-muted-foreground leading-relaxed">
            exist.io is the first Git server built for autonomous AI agents. Any MCP-compatible agent (Claude Code, Cursor, Windsurf, Devin) connects to your cluster to review PRs, triage issues, and run checks.
          </p>

          <div className="p-3 rounded-md bg-secondary/40 border border-border space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="font-semibold text-foreground">MCP Server Endpoint</span>
              <Badge variant="outline" className="font-mono text-[10px]">JSON-RPC 2.0</Badge>
            </div>
            <div className="flex items-center justify-between bg-input p-2 rounded border border-border font-mono text-xs text-foreground">
              <span>http://localhost:8080/mcp</span>
              <button
                onClick={() => handleCopy('http://localhost:8080/mcp')}
                className="p-1 text-muted-foreground hover:text-foreground cursor-pointer"
              >
                {copiedText === 'http://localhost:8080/mcp' ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
              </button>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-2 text-[11px]">
            <div className="p-2 rounded bg-secondary/20 border border-border space-y-0.5">
              <span className="font-semibold text-foreground">35+ Mounted Tools</span>
              <p className="text-muted-foreground">Diff inspection, symbolic AST search, branch switching, merge execution.</p>
            </div>
            <div className="p-2 rounded bg-secondary/20 border border-border space-y-0.5">
              <span className="font-semibold text-foreground">Ephemeral Leases</span>
              <p className="text-muted-foreground">HMAC-signed short-lived tokens ensuring bounded agent blast radius.</p>
            </div>
          </div>
        </div>
      ),
    },
    {
      id: 'enterprise-rulesets',
      title: 'Rulesets & Cryptographic Enforcement',
      subtitle: 'Strict branch protection, Ed25519 commit signing verification, and automated merge queues',
      icon: <ShieldCheck className="w-5 h-5" />,
      badge: 'Enterprise Security',
      color: 'text-primary',
      content: (
        <div className="space-y-4">
          <p className="text-xs text-muted-foreground leading-relaxed">
            Protect your production mainline branches with cryptographic verification and enterprise-grade policy enforcement directly in SQLite WAL transactions.
          </p>

          <div className="space-y-2">
            <div className="p-2.5 rounded-md border border-border bg-secondary/20 flex items-start gap-2.5">
              <ShieldCheck className="w-4 h-4 text-primary shrink-0 mt-0.5" />
              <div className="space-y-0.5">
                <span className="text-xs font-semibold text-foreground">Cryptographic Commit Signing</span>
                <p className="text-[11px] text-muted-foreground">
                  Reject unsigned commits or unverified keys before they ever hit the repository refs.
                </p>
              </div>
            </div>

            <div className="p-2.5 rounded-md border border-border bg-secondary/20 flex items-start gap-2.5">
              <GitBranch className="w-4 h-4 text-primary shrink-0 mt-0.5" />
              <div className="space-y-0.5">
                <span className="text-xs font-semibold text-foreground">Linear History & Merge Queues</span>
                <p className="text-[11px] text-muted-foreground">
                  Prevent merge skew by running tests against queued branches before atomic fast-forward commits.
                </p>
              </div>
            </div>
          </div>
        </div>
      ),
    },
    {
      id: 'enterprise-identity',
      title: 'Enterprise Identity & Governance',
      subtitle: 'SAML 2.0 / OIDC SSO, SCIM 2.0 automated provisioning, and immutable audit logs',
      icon: <Users className="w-5 h-5" />,
      badge: 'Compliance & Scale',
      color: 'text-primary',
      content: (
        <div className="space-y-4">
          <p className="text-xs text-muted-foreground leading-relaxed">
            Federate your organization with existing corporate identity providers. Automatic user provisioning and deprovisioning via standard SCIM 2.0 endpoints.
          </p>

          <div className="grid grid-cols-2 gap-2 text-xs">
            <div className="p-2.5 rounded-md bg-secondary/30 border border-border space-y-1">
              <span className="font-semibold text-foreground">Identity Federation</span>
              <p className="text-[11px] text-muted-foreground">Okta, Azure AD, Google Workspace via SAML 2.0 and OpenID Connect.</p>
            </div>
            <div className="p-2.5 rounded-md bg-secondary/30 border border-border space-y-1">
              <span className="font-semibold text-foreground">SCIM 2.0 Sync</span>
              <p className="text-[11px] text-muted-foreground">Mounted at <code className="font-mono text-foreground">/scim/v2/Users</code> for instant lifecycle synchronization.</p>
            </div>
          </div>

          <div className="p-3 rounded-md bg-primary/10 border border-primary/20 text-xs space-y-1">
            <span className="font-semibold text-foreground">You are now ready to build!</span>
            <p className="text-[11px] text-muted-foreground">
              Create your first repository, configure your SSH keys, or explore the in-depth system manual anytime.
            </p>
          </div>
        </div>
      ),
    },
  ]

  const currentStep = steps[currentStepIndex]

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 bg-background/80 backdrop-blur-xs"
        onClick={onClose}
      />

      {/* Tour Dialog */}
      <motion.div
        initial={{ opacity: 0, scale: 0.98, y: 8 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.98, y: 8 }}
        className="relative w-full max-w-xl rounded-lg border border-border bg-card text-card-foreground shadow-2xl z-10 overflow-hidden"
      >
        {/* Header */}
        <div className="px-5 py-4 border-b border-border bg-secondary/30 flex items-start justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-md bg-primary/15 border border-primary/30 flex items-center justify-center text-primary shrink-0">
              {currentStep.icon}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-sm font-semibold text-foreground">{currentStep.title}</h3>
                <Badge variant="outline" className="text-[10px] uppercase tracking-wider">
                  {currentStep.badge}
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground">{currentStep.subtitle}</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 text-muted-foreground hover:text-foreground rounded-md hover:bg-secondary/60 transition-colors cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Content Body */}
        <div className="p-5 min-h-[220px]">
          <AnimatePresence mode="wait">
            <motion.div
              key={currentStep.id}
              initial={{ opacity: 0, x: 10 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -10 }}
              transition={{ duration: 0.15 }}
            >
              {currentStep.content}
            </motion.div>
          </AnimatePresence>
        </div>

        {/* Footer Navigation */}
        <div className="px-5 py-3 border-t border-border bg-secondary/20 flex items-center justify-between">
          {/* Progress Dots */}
          <div className="flex items-center gap-1.5">
            {steps.map((step, idx) => (
              <button
                key={step.id}
                onClick={() => setCurrentStepIndex(idx)}
                className={cn(
                  'h-1.5 rounded-full transition-all cursor-pointer',
                  idx === currentStepIndex
                    ? 'w-6 bg-primary'
                    : 'w-1.5 bg-border hover:bg-muted-foreground'
                )}
                aria-label={`Go to step ${idx + 1}`}
              />
            ))}
            <span className="text-[11px] text-muted-foreground ml-2">
              {currentStepIndex + 1} of {steps.length}
            </span>
          </div>

          <div className="flex items-center gap-2">
            {onOpenManual && (
              <Button
                variant="ghost"
                size="xs"
                onClick={() => {
                  onClose()
                  onOpenManual()
                }}
                className="gap-1 text-xs text-muted-foreground hover:text-foreground mr-1"
              >
                <BookOpen className="w-3.5 h-3.5" />
                <span>Open Manual</span>
              </Button>
            )}

            {currentStepIndex > 0 && (
              <Button
                variant="secondary"
                size="xs"
                onClick={() => setCurrentStepIndex((prev) => prev - 1)}
                className="gap-1 text-xs"
              >
                <ArrowLeft className="w-3 h-3" />
                <span>Back</span>
              </Button>
            )}

            {currentStepIndex < steps.length - 1 ? (
              <Button
                variant="primary"
                size="xs"
                onClick={() => setCurrentStepIndex((prev) => prev + 1)}
                className="gap-1 text-xs"
              >
                <span>Next</span>
                <ArrowRight className="w-3 h-3" />
              </Button>
            ) : (
              <Button
                variant="primary"
                size="xs"
                onClick={handleFinish}
                className="gap-1 text-xs"
              >
                <Check className="w-3 h-3" />
                <span>Get Started</span>
              </Button>
            )}
          </div>
        </div>
      </motion.div>
    </div>
  )
}
