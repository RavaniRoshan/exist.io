import React, { useState } from 'react'
import { GitMerge, CheckCircle2, ChevronDown, Check } from 'lucide-react'
import { Button } from './Button'
import { cn } from '@/lib/utils'

export interface MergeBoxProps {
  onMerge: (method: 'merge' | 'squash' | 'rebase', commitTitle?: string, commitMessage?: string) => void
  disabled?: boolean
  defaultMethod?: 'merge' | 'squash' | 'rebase'
  defaultTitle?: string
  defaultMessage?: string
  className?: string
  checksPassed?: boolean
  isMerged?: boolean
}

/**
 * GitHub Canonical Pull Request Merge Box.
 * Replaces gimmicky touch-sliders with the authentic developer workflow:
 * Status summary, merge strategy selection, commit title/message editing, and confirmation.
 */
export const MergeBox: React.FC<MergeBoxProps> = ({
  onMerge,
  disabled = false,
  defaultMethod = 'merge',
  defaultTitle = 'Merge pull request',
  defaultMessage = '',
  className,
  checksPassed = true,
  isMerged = false,
}) => {
  const [method, setMethod] = useState<'merge' | 'squash' | 'rebase'>(defaultMethod)
  const [isConfirming, setIsConfirming] = useState(false)
  const [showDropdown, setShowDropdown] = useState(false)
  const [commitTitle, setCommitTitle] = useState(defaultTitle)
  const [commitMessage, setCommitMessage] = useState(defaultMessage)

  const methodLabels = {
    merge: 'Create a merge commit',
    squash: 'Squash and merge',
    rebase: 'Rebase and merge',
  }

  const methodDescriptions = {
    merge: 'All commits from this branch will be added to the base branch via a merge commit.',
    squash: 'The commits will be combined into one commit in the base branch.',
    rebase: 'The commits will be rebased and added to the base branch.',
  }

  if (isMerged) {
    return (
      <div className={cn('rounded-md border border-purple-500/40 bg-purple-950/15 p-4 flex items-center gap-3', className)}>
        <div className="w-8 h-8 rounded-full bg-purple-600 text-white flex items-center justify-center shrink-0">
          <GitMerge className="w-4 h-4" />
        </div>
        <div className="space-y-0.5 text-xs">
          <span className="font-semibold text-foreground">Pull request successfully merged and closed</span>
          <p className="text-muted-foreground">Changes have been incorporated into the target mainline branch.</p>
        </div>
      </div>
    )
  }

  return (
    <div className={cn('rounded-md border border-border bg-card overflow-hidden text-xs', className)}>
      {/* Status Header */}
      <div className="px-4 py-3 bg-secondary/30 border-b border-border flex items-start gap-3">
        <div className={cn('w-6 h-6 rounded-full flex items-center justify-center shrink-0 mt-0.5', checksPassed ? 'text-emerald-500 bg-emerald-500/10' : 'text-amber-500 bg-amber-500/10')}>
          <CheckCircle2 className="w-5 h-5" />
        </div>
        <div className="space-y-0.5 flex-1">
          <span className="font-semibold text-foreground block">
            {checksPassed ? 'This branch has no conflicts with the base branch' : 'Checks or policy gates in progress'}
          </span>
          <p className="text-muted-foreground">
            {checksPassed
              ? 'Merging can be performed automatically. All ruleset conditions satisfied.'
              : 'Review checks and required approval policies before proceeding.'}
          </p>
        </div>
      </div>

      {/* Main Action Area */}
      <div className="p-4 space-y-3">
        {!isConfirming ? (
          <div className="relative inline-flex items-center">
            {/* Primary Action Button */}
            <Button
              variant="primary"
              disabled={disabled}
              onClick={() => setIsConfirming(true)}
              className="rounded-r-none border-r-0 font-semibold gap-1.5"
            >
              <GitMerge className="w-3.5 h-3.5" />
              <span>{methodLabels[method]}</span>
            </Button>

            {/* Strategy Selector Dropdown Button */}
            <Button
              variant="primary"
              disabled={disabled}
              onClick={() => setShowDropdown(!showDropdown)}
              className="rounded-l-none px-2 border-l border-primary-foreground/20"
              title="Select merge strategy"
            >
              <ChevronDown className="w-3.5 h-3.5" />
            </Button>

            {/* Dropdown Menu */}
            {showDropdown && (
              <div className="absolute top-full left-0 mt-1 z-30 w-72 rounded-md border border-border bg-card shadow-xl p-1 space-y-1">
                {(['merge', 'squash', 'rebase'] as const).map((m) => (
                  <button
                    key={m}
                    type="button"
                    onClick={() => {
                      setMethod(m)
                      setShowDropdown(false)
                    }}
                    className={cn(
                      'w-full text-left p-2 rounded-md transition-colors flex items-start gap-2 cursor-pointer',
                      method === m ? 'bg-secondary text-foreground' : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground'
                    )}
                  >
                    <span className="mt-0.5 w-3.5 h-3.5 shrink-0 flex items-center justify-center">
                      {method === m && <Check className="w-3.5 h-3.5 text-primary" />}
                    </span>
                    <div className="space-y-0.5">
                      <span className="font-semibold block text-xs">{methodLabels[m]}</span>
                      <p className="text-[11px] leading-tight text-muted-foreground">{methodDescriptions[m]}</p>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>
        ) : (
          /* Confirmation Box with Commit Message inputs */
          <div className="space-y-3 p-3 bg-secondary/20 rounded-md border border-border">
            <div className="space-y-1">
              <label className="font-semibold text-foreground block text-[11px]">Commit Title</label>
              <input
                type="text"
                value={commitTitle}
                onChange={(e) => setCommitTitle(e.target.value)}
                className="w-full h-8 bg-input text-foreground border border-border rounded-md px-2.5 text-xs font-mono outline-none focus:border-primary"
              />
            </div>

            <div className="space-y-1">
              <label className="font-semibold text-foreground block text-[11px]">Commit Description (Optional)</label>
              <textarea
                rows={3}
                value={commitMessage}
                onChange={(e) => setCommitMessage(e.target.value)}
                placeholder="Add an optional extended description..."
                className="w-full bg-input text-foreground border border-border rounded-md p-2 text-xs font-mono outline-none focus:border-primary resize-none"
              />
            </div>

            <div className="flex items-center gap-2 pt-1">
              <Button
                variant="primary"
                onClick={() => {
                  onMerge(method, commitTitle, commitMessage)
                  setIsConfirming(false)
                }}
                className="gap-1.5"
              >
                <GitMerge className="w-3.5 h-3.5" />
                <span>Confirm {methodLabels[method].toLowerCase()}</span>
              </Button>
              <Button
                variant="ghost"
                onClick={() => setIsConfirming(false)}
              >
                Cancel
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
