import React, { useState } from 'react'
import { AlertTriangle, Trash2, KeyRound, Loader2 } from 'lucide-react'
import { Dialog } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { deleteRepository } from '@/lib/api'

export interface DeleteRepoModalProps {
  isOpen: boolean
  onClose: () => void
  owner: string
  repoName: string
  onSuccess: () => void
}

export const DeleteRepoModal: React.FC<DeleteRepoModalProps> = ({
  isOpen,
  onClose,
  owner,
  repoName,
  onSuccess,
}) => {
  const fullRepoName = `${owner}/${repoName}`

  const [hasAcknowledged, setHasAcknowledged] = useState(false)
  const [confirmNameInput, setConfirmNameInput] = useState('')
  const [sudoSecretInput, setSudoSecretInput] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const isNameMatched =
    confirmNameInput.trim() === fullRepoName || confirmNameInput.trim() === repoName
  const isFormValid = hasAcknowledged && isNameMatched && sudoSecretInput.trim().length >= 1

  const handleReset = () => {
    setHasAcknowledged(false)
    setConfirmNameInput('')
    setSudoSecretInput('')
    setErrorMessage(null)
  }

  const handleModalClose = () => {
    handleReset()
    onClose()
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!isFormValid) return

    setSubmitting(true)
    setErrorMessage(null)

    try {
      await deleteRepository(owner, repoName, fullRepoName, sudoSecretInput.trim())
      handleReset()
      onSuccess()
      onClose()
    } catch (err: any) {
      setErrorMessage(err.message || 'Failed to delete repository')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog
      isOpen={isOpen}
      onClose={handleModalClose}
      title="Delete repository confirmation"
    >
      <form onSubmit={handleSubmit} className="space-y-4 pt-1">
        {errorMessage && (
          <div className="p-3 bg-destructive/10 border border-destructive/30 rounded-md text-xs text-destructive flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>{errorMessage}</span>
          </div>
        )}

        {/* Warning Callout Box */}
        <div className="p-3 bg-destructive/10 border border-destructive/30 rounded-md text-xs text-destructive space-y-1.5">
          <p className="font-semibold flex items-center gap-1.5">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>This action cannot be undone.</span>
          </p>
          <p className="text-muted-foreground leading-relaxed">
            This will permanently delete the <strong className="text-foreground font-mono">{fullRepoName}</strong> repository, Git bare directories on disk, all commit histories, issues, and branches.
          </p>
        </div>

        {/* Step 1: Consequence Acknowledgment Checkbox */}
        <label className="flex items-start gap-2.5 cursor-pointer text-xs text-foreground select-none pt-1">
          <input
            type="checkbox"
            data-testid="delete-acknowledge-checkbox"
            checked={hasAcknowledged}
            onChange={(e) => setHasAcknowledged(e.target.checked)}
            className="mt-0.5 rounded border-border text-destructive focus:ring-destructive cursor-pointer"
          />
          <span className="leading-tight text-muted-foreground">
            I understand that all repository data, commits, issues, and branches will be permanently destroyed and cannot be restored.
          </span>
        </label>

        {/* Step 2: Name Confirmation Input */}
        <div className="space-y-1.5 pt-1">
          <label className="text-xs text-muted-foreground">
            To confirm deletion, please type <strong className="text-foreground font-mono">{fullRepoName}</strong>:
          </label>
          <input
            type="text"
            data-testid="delete-confirm-name-input"
            value={confirmNameInput}
            onChange={(e) => setConfirmNameInput(e.target.value)}
            placeholder={fullRepoName}
            className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground font-mono outline-none focus:border-destructive focus:ring-1 focus:ring-destructive"
          />
        </div>

        {/* Step 3: Sudo Security Authentication */}
        <div className="space-y-1.5 pt-1 border-t border-border/60">
          <div className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
            <KeyRound className="w-3.5 h-3.5 text-destructive" />
            <span>Sudo Mode Security Challenge</span>
          </div>
          <p className="text-[11px] text-muted-foreground">
            Enter your account password or personal access key to authorize permanent deletion:
          </p>
          <input
            type="password"
            data-testid="delete-sudo-secret-input"
            required
            value={sudoSecretInput}
            onChange={(e) => setSudoSecretInput(e.target.value)}
            placeholder="Account password or personal access key"
            className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground outline-none focus:border-destructive focus:ring-1 focus:ring-destructive font-mono"
          />
        </div>

        {/* Modal Actions */}
        <div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
          <Button type="button" variant="outline" size="sm" onClick={handleModalClose}>
            Cancel
          </Button>
          <Button
            type="submit"
            variant="outline"
            size="sm"
            data-testid="delete-submit-btn"
            disabled={submitting || !isFormValid}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90 border-transparent text-xs"
          >
            {submitting ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <span className="flex items-center gap-1.5">
                <Trash2 className="w-3.5 h-3.5" />
                <span>I understand, delete this repository</span>
              </span>
            )}
          </Button>
        </div>
      </form>
    </Dialog>
  )
}
