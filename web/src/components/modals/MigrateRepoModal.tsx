import React, { useState, useEffect } from 'react'
import {
  DownloadCloud,
  Lock,
  Globe,
  AlertCircle,
  Loader2,
  GitFork,
  GitBranch,
} from 'lucide-react'
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import {
  startRepoMigration,
  getMigrationStatus,
  type MigrateRepoPayload,
} from '@/lib/api'

export interface MigrateRepoModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: (repo: { owner: string; name: string }) => void
  defaultOwner?: string
}

export const MigrateRepoModal: React.FC<MigrateRepoModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
  defaultOwner = 'shiva',
}) => {
  const [platform, setPlatform] = useState<'github' | 'gitlab' | 'gitea' | 'generic'>('github')
  const [cloneUrl, setCloneUrl] = useState('')
  const [owner, setOwner] = useState(defaultOwner)
  const [repoName, setRepoName] = useState('')
  const [githubToken, setGithubToken] = useState('')
  const [importIssues, setImportIssues] = useState(true)
  const [importPRs, setImportPRs] = useState(true)
  const [visibility, setVisibility] = useState<'public' | 'private'>('public')

  const [jobId, setJobId] = useState<string | null>(null)
  const [phase, setPhase] = useState<string>('')
  const [progress, setProgress] = useState(0)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Auto-fill owner and repoName when cloneUrl changes
  const handleUrlChange = (url: string) => {
    setCloneUrl(url)
    const match = url.match(/[:/]([^/]+)\/([^/]+?)(?:\.git)?$/)
    if (match) {
      if (!repoName) {
        setRepoName(match[2].toLowerCase())
      }
    }
  }

  useEffect(() => {
    if (!jobId) return

    const interval = setInterval(async () => {
      try {
        const status = await getMigrationStatus(jobId)
        setPhase(status.phase)
        setProgress(status.progress)

        if (status.phase === 'complete') {
          clearInterval(interval)
          setIsSubmitting(false)
          onSuccess({ owner, name: repoName })
          handleReset()
        } else if (status.phase === 'failed') {
          clearInterval(interval)
          setIsSubmitting(false)
          setError(status.error || 'Migration failed')
        }
      } catch (err: any) {
        clearInterval(interval)
        setIsSubmitting(false)
        setError(err.message || 'Failed to poll migration status')
      }
    }, 1000)

    return () => clearInterval(interval)
  }, [jobId, owner, repoName])

  const handleReset = () => {
    setCloneUrl('')
    setRepoName('')
    setGithubToken('')
    setJobId(null)
    setPhase('')
    setProgress(0)
    setError(null)
    setIsSubmitting(false)
  }

  const handleClose = () => {
    if (!isSubmitting) {
      handleReset()
      onClose()
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    if (!cloneUrl.trim()) {
      setError('Please provide a clone URL.')
      return
    }
    if (!repoName.trim()) {
      setError('Please provide a repository name.')
      return
    }

    setIsSubmitting(true)
    setProgress(5)
    setPhase('cloning')

    const payload: MigrateRepoPayload = {
      clone_url: cloneUrl.trim(),
      owner: owner.trim() || 'shiva',
      repo_name: repoName.trim().toLowerCase().replace(/[^a-z0-9._-]/g, '-'),
      github_token: githubToken.trim() || undefined,
      import_issues: platform === 'github' && importIssues,
      import_prs: platform === 'github' && importPRs,
      private: visibility === 'private',
    }

    try {
      const resp = await startRepoMigration(payload)
      setJobId(resp.jobId)
    } catch (err: any) {
      setIsSubmitting(false)
      setError(err.message || 'Failed to start migration')
    }
  }

  const getPhaseLabel = (p: string) => {
    switch (p) {
      case 'cloning':
        return 'Cloning Git objects, branches, and commit tree...'
      case 'importing_issues':
        return 'Importing GitHub issues, labels, and discussion comments...'
      case 'importing_prs':
        return 'Importing pull requests, reviews, and diff annotations...'
      case 'complete':
        return 'Migration complete! Repository is ready.'
      default:
        return 'Initializing migration pipeline...'
    }
  }

  return (
    <Dialog isOpen={isOpen} onClose={handleClose} maxWidth="lg">
      <DialogHeader>
        <DialogTitle className="flex items-center gap-2">
          <DownloadCloud className="w-5 h-5 text-primary" />
          <span>Migrate Repository</span>
        </DialogTitle>
        <DialogDescription>
          Import code, revision history, issues, and PRs into your sovereign exist.io cluster.
        </DialogDescription>
      </DialogHeader>

      {isSubmitting ? (
        <div className="py-6 space-y-6">
          <div className="text-center space-y-2">
            <div className="inline-flex p-3 rounded-full bg-primary/10 text-primary animate-pulse">
              <DownloadCloud className="w-6 h-6 animate-bounce" />
            </div>
            <div className="text-sm font-semibold text-foreground">
              Migrating {owner}/{repoName}
            </div>
            <div className="text-xs text-muted-foreground">
              {getPhaseLabel(phase)}
            </div>
          </div>

          {/* Progress Bar */}
          <div className="space-y-1.5">
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Progress</span>
              <span className="font-semibold text-foreground">{progress}%</span>
            </div>
            <div className="w-full h-2 rounded-full bg-secondary overflow-hidden">
              <div
                className="h-full bg-primary transition-all duration-300 ease-out"
                style={{ width: `${progress}%` }}
              />
            </div>
          </div>

          <div className="p-3 rounded-md bg-secondary/50 border border-border text-xs text-muted-foreground flex items-center gap-2">
            <Loader2 className="w-4 h-4 animate-spin text-primary shrink-0" />
            <span>Zero downtime: git objects are streamed hermetically into bare storage.</span>
          </div>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-4 pt-2 text-xs font-sans">
          {error && (
            <div className="p-3 bg-destructive/10 border border-destructive/20 text-destructive rounded-md flex items-start gap-2">
              <AlertCircle className="w-4 h-4 mt-0.5 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Source Platform Selector */}
          <div>
            <label className="block text-xs font-medium text-foreground mb-1.5">
              Source Platform
            </label>
            <div className="grid grid-cols-4 gap-2">
              {[
                { id: 'github', label: 'GitHub', icon: <GitFork className="w-3.5 h-3.5" /> },
                { id: 'gitlab', label: 'GitLab', icon: <GitBranch className="w-3.5 h-3.5" /> },
                { id: 'gitea', label: 'Gitea', icon: <Globe className="w-3.5 h-3.5" /> },
                { id: 'generic', label: 'Git URL', icon: <DownloadCloud className="w-3.5 h-3.5" /> },
              ].map((p) => (
                <button
                  key={p.id}
                  type="button"
                  onClick={() => setPlatform(p.id as any)}
                  className={`flex items-center justify-center gap-1.5 py-1.5 px-2 rounded-md border text-xs font-medium transition-colors cursor-pointer ${
                    platform === p.id
                      ? 'bg-primary text-primary-foreground border-primary shadow-2xs'
                      : 'bg-secondary/60 text-foreground border-border hover:bg-secondary'
                  }`}
                >
                  {p.icon}
                  <span>{p.label}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Clone URL Input */}
          <div>
            <label className="block text-xs font-medium text-foreground mb-1">
              Clone URL <span className="text-destructive">*</span>
            </label>
            <input
              type="text"
              required
              placeholder="https://github.com/owner/repository.git"
              value={cloneUrl}
              onChange={(e) => handleUrlChange(e.target.value)}
              className="w-full px-3 py-1.5 rounded-md border border-border bg-input text-foreground focus:outline-hidden focus:ring-1 focus:ring-primary font-mono text-xs"
            />
          </div>

          {/* Target Owner & Repo Name */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-foreground mb-1">
                Target Namespace
              </label>
              <input
                type="text"
                value={owner}
                onChange={(e) => setOwner(e.target.value)}
                className="w-full px-3 py-1.5 rounded-md border border-border bg-input text-foreground focus:outline-hidden focus:ring-1 focus:ring-primary text-xs"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-foreground mb-1">
                Target Repository <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                required
                placeholder="my-repo"
                value={repoName}
                onChange={(e) => setRepoName(e.target.value)}
                className="w-full px-3 py-1.5 rounded-md border border-border bg-input text-foreground focus:outline-hidden focus:ring-1 focus:ring-primary text-xs font-mono"
              />
            </div>
          </div>

          {/* Optional GitHub Token */}
          {platform === 'github' && (
            <div>
              <label className="block text-xs font-medium text-foreground mb-1">
                GitHub Personal Access Token <span className="text-muted-foreground font-normal">(optional)</span>
              </label>
              <input
                type="password"
                placeholder="ghp_xxxxxxxxxxxxxxxxxxxx"
                value={githubToken}
                onChange={(e) => setGithubToken(e.target.value)}
                className="w-full px-3 py-1.5 rounded-md border border-border bg-input text-foreground focus:outline-hidden focus:ring-1 focus:ring-primary text-xs font-mono"
              />
              <div className="text-[10px] text-muted-foreground mt-1">
                Required for private repositories and importing issues / pull requests.
              </div>
            </div>
          )}

          {/* Metadata Import Toggles */}
          {platform === 'github' && (
            <div className="p-3 bg-secondary/40 rounded-md border border-border space-y-2">
              <div className="font-semibold text-foreground text-[11px]">Metadata Ingestion</div>
              <label className="flex items-center gap-2 cursor-pointer text-muted-foreground hover:text-foreground">
                <input
                  type="checkbox"
                  checked={importIssues}
                  onChange={(e) => setImportIssues(e.target.checked)}
                  className="rounded border-border text-primary focus:ring-primary/40"
                />
                <span>Import issues, labels, and comments</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer text-muted-foreground hover:text-foreground">
                <input
                  type="checkbox"
                  checked={importPRs}
                  onChange={(e) => setImportPRs(e.target.checked)}
                  className="rounded border-border text-primary focus:ring-primary/40"
                />
                <span>Import pull requests and reviews</span>
              </label>
            </div>
          )}

          {/* Visibility Selector */}
          <div>
            <label className="block text-xs font-medium text-foreground mb-1">Visibility</label>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => setVisibility('public')}
                className={`flex items-center gap-2 p-2 rounded-md border text-left cursor-pointer transition-colors ${
                  visibility === 'public'
                    ? 'border-primary bg-primary/5 text-foreground'
                    : 'border-border bg-secondary/30 text-muted-foreground hover:bg-secondary/60'
                }`}
              >
                <Globe className="w-4 h-4 text-primary shrink-0" />
                <div>
                  <div className="font-medium">Public</div>
                  <div className="text-[10px] text-muted-foreground">Anyone can clone</div>
                </div>
              </button>
              <button
                type="button"
                onClick={() => setVisibility('private')}
                className={`flex items-center gap-2 p-2 rounded-md border text-left cursor-pointer transition-colors ${
                  visibility === 'private'
                    ? 'border-primary bg-primary/5 text-foreground'
                    : 'border-border bg-secondary/30 text-muted-foreground hover:bg-secondary/60'
                }`}
              >
                <Lock className="w-4 h-4 text-amber-400 shrink-0" />
                <div>
                  <div className="font-medium">Private</div>
                  <div className="text-[10px] text-muted-foreground">Authorized only</div>
                </div>
              </button>
            </div>
          </div>

          <DialogFooter className="pt-2 border-t border-border flex justify-end gap-2">
            <Button type="button" variant="outline" size="sm" onClick={handleClose}>
              Cancel
            </Button>
            <Button type="submit" variant="primary" size="sm" disabled={!cloneUrl.trim() || !repoName.trim()}>
              <DownloadCloud className="w-3.5 h-3.5 mr-1.5" />
              <span>Start Migration</span>
            </Button>
          </DialogFooter>
        </form>
      )}
    </Dialog>
  )
}
