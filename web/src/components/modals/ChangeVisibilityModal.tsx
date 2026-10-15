import React, { useState } from 'react'
import { AlertTriangle, Lock, Globe, KeyRound, Loader2 } from 'lucide-react'
import { Dialog } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { updateRepoSettings } from '@/lib/api'

export interface ChangeVisibilityModalProps {
  isOpen: boolean
  onClose: () => void
  owner: string
  repoName: string
  isCurrentlyPrivate: boolean
  onSuccess: () => void
}

export const ChangeVisibilityModal: React.FC<ChangeVisibilityModalProps> = ({
  isOpen,
  onClose,
  owner,
  repoName,
  isCurrentlyPrivate,
  onSuccess,
}) => {
  const targetVisibility = isCurrentlyPrivate ? 'public' : 'private'
  const fullRepoName = `${owner}/${repoName}`

  const [hasAcknowledged, setHasAcknowledged] = useState(false)
  const [confirmNameInput, setConfirmNameInput] = useState('')
  const [sudoSecretInput, setSudoSecretInput] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const isNameMatched = confirmNameInput.trim() === fullRepoName || confirmNameInput.trim() === repoName
  const isFormValid = hasAcknowledged && isNameMatched && sudoSecretInput.trim().length >= 4

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
      await updateRepoSettings(owner, repoName, {
        visibility: targetVisibility,
        confirm_name: fullRepoName,
        sudo_secret: sudoSecretInput.trim(),
      })
      handleReset()
      onSuccess()
      onClose()
    } catch (err: any) {
      setErrorMessage(err.message || 'Failed to change repository visibility')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog
      isOpen={isOpen}
      onClose={handleModalClose}
      title={`Change visibility to ${targetVisibility}`}
    >
      <form onSubmit={handleSubmit} className="space-y-4 pt-1">
        {errorMessage && (
          <div className="p-3 bg-destructive/10 border border-destructive/30 rounded-md text-xs text-destructive flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>{errorMessage}</span>
          </div>
        )}

        {/* Warning Callout Box */}
        <div className="p-3.5 bg-amber-500/10 border border-amber-500/30 rounded-md text-xs space-y-2">
          <div className="font-semibold text-amber-500 flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>Caution: High-impact security and visibility change</span>
          </div>
          <p className="text-muted-foreground leading-relaxed">
            You are changing the visibility of <strong className="text-foreground font-mono">{fullRepoName}</strong> from{' '}
            <strong className="text-foreground">{isCurrentlyPrivate ? 'Private' : 'Public'}</strong> to{' '}
            <strong className="text-foreground uppercase">{targetVisibility}</strong>.
          </p>

          <ul className="list-disc list-inside space-y-1 text-muted-foreground text-[11px] pt-1">
            {isCurrentlyPrivate ? (
              <>
                <li>All source code, commit history, and branches will become world-readable.</li>
                <li>Anyone with network access to this cluster will be able to clone and inspect code.</li>
                <li>Issues, pull requests, releases, and discussions will be publicly accessible.</li>
              </>
            ) : (
              <>
                <li>Access will be restricted strictly to direct collaborators and authorized organization members.</li>
                <li>Existing public forks will be detached from this repository network.</li>
                <li>Repository metadata and code will be excluded from global search indexing.</li>
              </>
            )}
          </ul>
        </div>

        {/* Step 1: Acknowledgment Checkbox */}
        <label className="flex items-start gap-2.5 cursor-pointer text-xs text-foreground select-none pt-1">
          <input
            type="checkbox"
            data-testid="visibility-acknowledge-checkbox"
            checked={hasAcknowledged}
            onChange={(e) => setHasAcknowledged(e.target.checked)}
            className="mt-0.5 rounded border-border text-primary focus:ring-ring"
          />
          <span className="leading-tight text-muted-foreground">
            I have read and understand the effects and security consequences of changing this repository to{' '}
            <strong className="text-foreground">{targetVisibility}</strong>.
          </span>
        </label>

        {/* Step 2: Name Confirmation Input */}
        <div className="space-y-1.5 pt-1">
          <label className="text-xs text-muted-foreground">
            To confirm, please type <strong className="text-foreground font-mono">{fullRepoName}</strong> below:
          </label>
          <input
            type="text"
            data-testid="visibility-confirm-name-input"
            value={confirmNameInput}
            onChange={(e) => setConfirmNameInput(e.target.value)}
            placeholder={fullRepoName}
            className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
          />
        </div>

        {/* Step 3: Sudo Security Authentication */}
        <div className="space-y-1.5 pt-1 border-t border-border/60">
          <div className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
            <KeyRound className="w-3.5 h-3.5 text-primary" />
            <span>Sudo Mode Security Challenge</span>
          </div>
          <p className="text-[11px] text-muted-foreground">
            Enter your account password or personal access key to authorize this action:
          </p>
          <input
            type="password"
            data-testid="visibility-sudo-secret-input"
            required
            value={sudoSecretInput}
            onChange={(e) => setSudoSecretInput(e.target.value)}
            placeholder="Account password or personal access key"
            className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring font-mono"
          />
        </div>

        {/* Modal Actions */}
        <div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
          <Button type="button" variant="outline" size="sm" onClick={handleModalClose}>
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            size="sm"
            data-testid="visibility-submit-btn"
            disabled={submitting || !isFormValid}
            className={
              isCurrentlyPrivate
                ? 'bg-amber-600 hover:bg-amber-500 text-white'
                : 'bg-primary hover:bg-primary/90 text-primary-foreground'
            }
          >
            {submitting ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : isCurrentlyPrivate ? (
              <span className="flex items-center gap-1.5">
                <Globe className="w-3.5 h-3.5" />
                <span>Make this repository public</span>
              </span>
            ) : (
              <span className="flex items-center gap-1.5">
                <Lock className="w-3.5 h-3.5" />
                <span>Make this repository private</span>
              </span>
            )}
          </Button>
        </div>
      </form>
    </Dialog>
  )
}
