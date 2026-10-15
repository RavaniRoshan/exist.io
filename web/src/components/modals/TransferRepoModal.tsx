import React, { useState } from 'react'
import { AlertTriangle, ArrowRightLeft, KeyRound, Loader2 } from 'lucide-react'
import { Dialog } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { transferRepository } from '@/lib/api'

export interface TransferRepoModalProps {
  isOpen: boolean
  onClose: () => void
  owner: string
  repoName: string
  onSuccess: (newOwner: string, repoName: string) => void
}

export const TransferRepoModal: React.FC<TransferRepoModalProps> = ({
  isOpen,
  onClose,
  owner,
  repoName,
  onSuccess,
}) => {
  const fullRepoName = `${owner}/${repoName}`

  const [newOwnerInput, setNewOwnerInput] = useState('')
  const [confirmNameInput, setConfirmNameInput] = useState('')
  const [sudoSecretInput, setSudoSecretInput] = useState('')
  const [hasAcknowledged, setHasAcknowledged] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const isNameMatched =
    confirmNameInput.trim() === fullRepoName || confirmNameInput.trim() === repoName
  const isOwnerValid =
    newOwnerInput.trim().length > 0 &&
    newOwnerInput.trim().toLowerCase() !== owner.toLowerCase()
  const isFormValid =
    hasAcknowledged &&
    isOwnerValid &&
    isNameMatched &&
    sudoSecretInput.trim().length >= 4

  const handleReset = () => {
    setNewOwnerInput('')
    setConfirmNameInput('')
    setSudoSecretInput('')
    setHasAcknowledged(false)
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
      await transferRepository(owner, repoName, {
        new_owner: newOwnerInput.trim(),
        confirm_name: fullRepoName,
        sudo_secret: sudoSecretInput.trim(),
      })
      const targetOwner = newOwnerInput.trim()
      handleReset()
      onSuccess(targetOwner, repoName)
      onClose()
    } catch (err: any) {
      setErrorMessage(err.message || 'Failed to transfer repository')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog
      isOpen={isOpen}
      onClose={handleModalClose}
      title="Transfer repository ownership"
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
            <ArrowRightLeft className="w-4 h-4 shrink-0" />
            <span>Transferring ownership to another account or organization</span>
          </div>
          <p className="text-muted-foreground leading-relaxed">
            Transferring <strong className="text-foreground font-mono">{fullRepoName}</strong> will immediately reassign ownership to another user or organization namespace.
          </p>

          <ul className="list-disc list-inside space-y-1 text-muted-foreground text-[11px] pt-1">
            <li>Git remote clone URLs and Web URLs will redirect to the new namespace.</li>
            <li>All commit history, branches, tags, releases, and Git LFS objects will be preserved.</li>
            <li>All open and closed issues, pull requests, comments, and reviews will move to the new repository.</li>
            <li>Configured webhooks and deploy keys will remain attached.</li>
          </ul>
        </div>

        {/* New Owner Input */}
        <div className="space-y-1.5">
          <label className="text-xs font-semibold text-foreground block">
            New owner&apos;s username or organization name
          </label>
          <input
            type="text"
            required
            data-testid="transfer-owner-input"
            value={newOwnerInput}
            onChange={(e) => setNewOwnerInput(e.target.value)}
            placeholder="e.g. acme-org or alex"
            className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring font-mono"
          />
          {newOwnerInput.trim().toLowerCase() === owner.toLowerCase() && (
            <p className="text-[11px] text-destructive">
              New owner cannot be the current repository owner.
            </p>
          )}
        </div>

        {/* Consequence Acknowledgement */}
        <label className="flex items-start gap-2.5 cursor-pointer text-xs text-foreground select-none pt-1">
          <input
            type="checkbox"
            data-testid="transfer-ack-checkbox"
            checked={hasAcknowledged}
            onChange={(e) => setHasAcknowledged(e.target.checked)}
            className="mt-0.5 rounded border-border text-primary focus:ring-ring"
          />
          <span className="leading-tight">
            I understand that existing direct collaborators may lose access if the new owner does not grant permissions, and that repository administrative control will transfer immediately.
          </span>
        </label>

        {/* Repository Name Confirmation Input */}
        <div className="space-y-1.5 pt-1 border-t border-border">
          <label className="text-xs font-semibold text-foreground block">
            To confirm, type <strong className="font-mono text-foreground select-all">{fullRepoName}</strong> in the box below
          </label>
          <input
            type="text"
            required
            data-testid="transfer-confirm-input"
            value={confirmNameInput}
            onChange={(e) => setConfirmNameInput(e.target.value)}
            placeholder={fullRepoName}
            className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring font-mono"
          />
        </div>

        {/* Sudo Mode Password Challenge */}
        <div className="space-y-1.5 pt-1 border-t border-border">
          <div className="flex items-center gap-1.5">
            <KeyRound className="w-3.5 h-3.5 text-muted-foreground" />
            <label className="text-xs font-semibold text-foreground">
              Confirm your password or personal access key
            </label>
          </div>
          <input
            type="password"
            required
            data-testid="transfer-sudo-input"
            value={sudoSecretInput}
            onChange={(e) => setSudoSecretInput(e.target.value)}
            placeholder="Enter password or PAT token"
            className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring font-mono"
          />
          <p className="text-[11px] text-muted-foreground">
            Sudo mode authentication ensures critical repository ownership transfers are verified.
          </p>
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end gap-2 pt-2 border-t border-border">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleModalClose}
            disabled={submitting}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            size="sm"
            data-testid="transfer-submit-btn"
            disabled={!isFormValid || submitting}
            className="gap-1.5 bg-amber-600 hover:bg-amber-700 text-white border-amber-600"
          >
            {submitting ? (
              <>
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                <span>Transferring...</span>
              </>
            ) : (
              <span>I understand the consequences, transfer this repository</span>
            )}
          </Button>
        </div>
      </form>
    </Dialog>
  )
}
