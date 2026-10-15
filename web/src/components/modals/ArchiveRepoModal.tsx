import React, { useState } from 'react'
import { Archive, AlertTriangle, Loader2 } from 'lucide-react'
import { Dialog } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { updateRepoSettings } from '@/lib/api'

export interface ArchiveRepoModalProps {
  isOpen: boolean
  onClose: () => void
  owner: string
  repoName: string
  isCurrentlyArchived: boolean
  onSuccess: () => void
}

export const ArchiveRepoModal: React.FC<ArchiveRepoModalProps> = ({
  isOpen,
  onClose,
  owner,
  repoName,
  isCurrentlyArchived,
  onSuccess,
}) => {
  const fullRepoName = `${owner}/${repoName}`
  const [confirmInput, setConfirmInput] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const isMatched = confirmInput.trim() === fullRepoName || confirmInput.trim() === repoName

  const handleReset = () => {
    setConfirmInput('')
    setErrorMessage(null)
  }

  const handleModalClose = () => {
    handleReset()
    onClose()
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!isMatched) return

    setSubmitting(true)
    setErrorMessage(null)

    try {
      await updateRepoSettings(owner, repoName, {
        archived: !isCurrentlyArchived,
        confirm_name: fullRepoName,
      })
      handleReset()
      onSuccess()
      onClose()
    } catch (err: any) {
      setErrorMessage(err.message || 'Failed to update archival status')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog
      isOpen={isOpen}
      onClose={handleModalClose}
      title={isCurrentlyArchived ? 'Unarchive repository' : 'Archive repository'}
    >
      <form onSubmit={handleSubmit} className="space-y-4 pt-1">
        {errorMessage && (
          <div className="p-3 bg-destructive/10 border border-destructive/30 rounded-md text-xs text-destructive flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>{errorMessage}</span>
          </div>
        )}

        <div className="p-3.5 bg-amber-500/10 border border-amber-500/30 rounded-md text-xs space-y-2">
          <div className="font-semibold text-amber-500 flex items-center gap-2">
            <Archive className="w-4 h-4 shrink-0" />
            <span>
              {isCurrentlyArchived
                ? 'Restore repository to read/write mode'
                : 'Mark repository as read-only and immutable'}
            </span>
          </div>
          <p className="text-muted-foreground leading-relaxed">
            {isCurrentlyArchived ? (
              <>
                Unarchiving <strong className="text-foreground font-mono">{fullRepoName}</strong> will restore full read/write functionality. Collaborators will again be able to push commits, open pull requests, and file issues.
              </>
            ) : (
              <>
                Archiving <strong className="text-foreground font-mono">{fullRepoName}</strong> will lock the repository and mark it as read-only. Direct Git pushes, pull request merges, issues, and comments will be completely disabled.
              </>
            )}
          </p>
        </div>

        <div className="space-y-1.5 pt-1">
          <label className="text-xs text-muted-foreground">
            To confirm, please type <strong className="text-foreground font-mono">{fullRepoName}</strong> below:
          </label>
          <input
            type="text"
            data-testid="archive-confirm-name-input"
            value={confirmInput}
            onChange={(e) => setConfirmInput(e.target.value)}
            placeholder={fullRepoName}
            className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground font-mono outline-none focus:border-amber-500 focus:ring-1 focus:ring-amber-500"
          />
        </div>

        <div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
          <Button type="button" variant="outline" size="sm" onClick={handleModalClose}>
            Cancel
          </Button>
          <Button
            type="submit"
            variant="outline"
            size="sm"
            data-testid="archive-submit-btn"
            disabled={submitting || !isMatched}
            className="border-amber-500/40 text-amber-500 hover:bg-amber-500/10"
          >
            {submitting ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : isCurrentlyArchived ? (
              'I understand, unarchive this repository'
            ) : (
              'I understand, archive this repository'
            )}
          </Button>
        </div>
      </form>
    </Dialog>
  )
}
