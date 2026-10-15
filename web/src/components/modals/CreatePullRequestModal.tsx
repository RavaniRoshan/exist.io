import React, { useState, useEffect } from 'react'
import {
  GitPullRequest,
  GitBranch,
  ArrowRight,
  CheckCircle2,
  AlertCircle,
  Loader2,
} from 'lucide-react'
import { Dialog } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import {
  getBranches,
  createPullRequest,
  type RepoInfo,
  type BranchItem,
  type PullRequest,
} from '@/lib/api'

export interface CreatePullRequestModalProps {
  isOpen: boolean
  onClose: () => void
  repo?: RepoInfo
  onPullRequestCreated?: (newPr: PullRequest) => void
}

export const CreatePullRequestModal: React.FC<CreatePullRequestModalProps> = ({
  isOpen,
  onClose,
  repo,
  onPullRequestCreated,
}) => {
  const owner = repo?.ownerSlug || 'acme'
  const repoName = repo?.name || 'rocket'
  const defaultBranch = repo?.defaultBranch || 'main'

  const [branches, setBranches] = useState<BranchItem[]>([])
  const [baseBranch, setBaseBranch] = useState(defaultBranch)
  const [headBranch, setHeadBranch] = useState('')
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [isDraft, setIsDraft] = useState(false)
  const [tab, setTab] = useState<'write' | 'preview'>('write')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!isOpen) return
    setError(null)
    getBranches(owner, repoName)
      .then((data) => {
        setBranches(data)
        const nonDefault = data.find((b) => b.name !== defaultBranch)
        if (nonDefault) {
          setHeadBranch(nonDefault.name)
        } else if (data.length > 0) {
          setHeadBranch(data[0].name)
        }
      })
      .catch(() => {
        setBranches([{ name: defaultBranch, commitSha: '54926f8', isDefault: true }])
        setHeadBranch(defaultBranch)
      })
  }, [isOpen, owner, repoName, defaultBranch])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim() || !headBranch) {
      setError('Title and compare branch are required.')
      return
    }

    if (headBranch === baseBranch) {
      setError('Head and base branches must be different to open a pull request.')
      return
    }

    setSubmitting(true)
    setError(null)
    try {
      const created = await createPullRequest(owner, repoName, {
        title: title.trim(),
        head: headBranch,
        base: baseBranch,
        body: body.trim(),
        draft: isDraft,
      })

      if (onPullRequestCreated) {
        onPullRequestCreated(created)
      }
      onClose()
      setTitle('')
      setBody('')
    } catch (err: any) {
      setError(err.message || 'Failed to create pull request')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog isOpen={isOpen} onClose={onClose} title="Open a pull request" maxWidth="2xl">
      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Branch Comparison Pill Selector */}
        <div className="p-3 bg-secondary/40 border border-border rounded-md space-y-2">
          <div className="flex items-center gap-2 text-xs text-muted-foreground font-medium">
            <GitBranch className="w-3.5 h-3.5 text-primary" />
            <span>Comparing changes across branches:</span>
          </div>

          <div className="flex flex-wrap items-center gap-2 text-xs">
            {/* Base branch selector */}
            <div className="flex items-center gap-1.5 bg-background border border-border rounded px-2.5 py-1">
              <span className="text-muted-foreground">base:</span>
              <select
                value={baseBranch}
                onChange={(e) => setBaseBranch(e.target.value)}
                className="bg-transparent font-mono font-medium text-foreground outline-none cursor-pointer"
              >
                {branches.map((b) => (
                  <option key={`base-${b.name}`} value={b.name} className="bg-background text-foreground">
                    {b.name}
                  </option>
                ))}
              </select>
            </div>

            <ArrowRight className="w-3.5 h-3.5 text-muted-foreground" />

            {/* Compare branch selector */}
            <div className="flex items-center gap-1.5 bg-background border border-border rounded px-2.5 py-1">
              <span className="text-muted-foreground">compare:</span>
              {branches.length > 0 ? (
                <select
                  value={headBranch}
                  onChange={(e) => setHeadBranch(e.target.value)}
                  className="bg-transparent font-mono font-medium text-foreground outline-none cursor-pointer"
                >
                  {branches.map((b) => (
                    <option key={`head-${b.name}`} value={b.name} className="bg-background text-foreground">
                      {b.name}
                    </option>
                  ))}
                </select>
              ) : (
                <input
                  type="text"
                  placeholder="feature-branch"
                  value={headBranch}
                  onChange={(e) => setHeadBranch(e.target.value)}
                  className="bg-transparent font-mono font-medium text-foreground outline-none w-32"
                />
              )}
            </div>

            {/* Mergeability badge */}
            {headBranch && headBranch !== baseBranch && (
              <span className="inline-flex items-center gap-1 text-[11px] text-emerald-500 font-medium ml-auto">
                <CheckCircle2 className="w-3.5 h-3.5" />
                <span>Able to merge automatically</span>
              </span>
            )}
          </div>
        </div>

        {error && (
          <div className="p-3 bg-destructive/10 border border-destructive/30 rounded-md text-xs text-destructive flex items-center gap-2">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {/* PR Title */}
        <div className="space-y-1">
          <label className="text-xs font-semibold text-foreground">
            Title <span className="text-destructive">*</span>
          </label>
          <input
            type="text"
            required
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. feat: implement autonomous agent review gate"
            className="w-full h-9 px-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring font-medium"
          />
        </div>

        {/* PR Description Body Box with Write/Preview */}
        <div className="space-y-1">
          <div className="flex items-center justify-between">
            <label className="text-xs font-semibold text-foreground">Description</label>
            <div className="flex items-center gap-1 text-xs">
              <button
                type="button"
                onClick={() => setTab('write')}
                className={`px-2 py-0.5 rounded text-[11px] font-medium transition-colors ${
                  tab === 'write' ? 'bg-secondary text-foreground' : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                Write
              </button>
              <button
                type="button"
                onClick={() => setTab('preview')}
                className={`px-2 py-0.5 rounded text-[11px] font-medium transition-colors ${
                  tab === 'preview' ? 'bg-secondary text-foreground' : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                Preview
              </button>
            </div>
          </div>

          {tab === 'write' ? (
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="Explain the changes in this PR, motivation, and tests run..."
              rows={5}
              className="w-full p-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring resize-y leading-relaxed"
            />
          ) : (
            <div className="min-h-[110px] p-3 text-xs bg-secondary/20 border border-border rounded-md text-foreground whitespace-pre-wrap leading-relaxed">
              {body.trim() ? body : <span className="text-muted-foreground">Nothing to preview</span>}
            </div>
          )}
        </div>

        {/* Draft PR option */}
        <div className="flex items-center gap-2 pt-1">
          <input
            type="checkbox"
            id="draft-pr-checkbox"
            checked={isDraft}
            onChange={(e) => setIsDraft(e.target.checked)}
            className="rounded border-border text-primary focus:ring-ring"
          />
          <label htmlFor="draft-pr-checkbox" className="text-xs text-muted-foreground cursor-pointer select-none">
            Create as draft pull request (cannot be merged until marked ready for review)
          </label>
        </div>

        {/* Footer Actions */}
        <div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
          <Button type="button" variant="outline" size="sm" onClick={onClose} disabled={submitting}>
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            size="sm"
            disabled={submitting || !title.trim() || headBranch === baseBranch}
            className="gap-1.5"
          >
            {submitting ? (
              <>
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                <span>Creating...</span>
              </>
            ) : (
              <>
                <GitPullRequest className="w-3.5 h-3.5" />
                <span>Create pull request</span>
              </>
            )}
          </Button>
        </div>
      </form>
    </Dialog>
  )
}
