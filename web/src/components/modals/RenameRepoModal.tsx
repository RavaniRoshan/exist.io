import React, { useState } from 'react'
import { Edit3, AlertTriangle, Loader2 } from 'lucide-react'
import { Dialog } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { updateRepoSettings } from '@/lib/api'

export interface RenameRepoModalProps {
  isOpen: boolean
  onClose: () => void
  owner: string
  currentName: string
  onSuccess: (newName: string) => void
}

export const RenameRepoModal: React.FC<RenameRepoModalProps> = ({
  isOpen,
  onClose,
  owner,
  currentName,
  onSuccess,
}) => {
  const fullRepoName = `${owner}/${currentName}`
  const [newNameInput, setNewNameInput] = useState('')
  const [confirmInput, setConfirmInput] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const isMatched = confirmInput.trim() === fullRepoName || confirmInput.trim() === currentName
  const isNewNameValid = newNameInput.trim().length > 0 && newNameInput.trim() !== currentName
  const isFormValid = isMatched && isNewNameValid

  const handleReset = () => {
    setNewNameInput('')
    setConfirmInput('')
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
      const sanitized = newNameInput.trim()
      await updateRepoSettings(owner, currentName, {
        new_name: sanitized,
        confirm_name: fullRepoName,
      })
      handleReset()
      onSuccess(sanitized)
      onClose()
    } catch (err: any) {
      setErrorMessage(err.message || 'Failed to rename repository')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog isOpen={isOpen} onClose={handleModalClose} title="Rename repository">
      <form onSubmit={handleSubmit} className="space-y-4 pt-1">
        {errorMessage && (
          <div className="p-3 bg-destructive/10 border border-destructive/30 rounded-md text-xs text-destructive flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>{errorMessage}</span>
          </div>
        )}

        <div className="p-3.5 bg-secondary/50 border border-border rounded-md text-xs space-y-2">
          <div className="font-semibold text-foreground flex items-center gap-2">
            <Edit3 className="w-4 h-4 text-primary shrink-0" />
            <span>Rename {fullRepoName}</span>
          </div>
          <p className="text-muted-foreground leading-relaxed">
            Renaming this repository will update its bare Git object store and URL. Existing Git remotes will need to update their remote fetch and push URLs.
          </p>
        </div>

        <div className="space-y-1.5 pt-1">
          <label className="text-xs font-semibold text-foreground">New Repository Name</label>
          <input
            type="text"
            data-testid="rename-new-name-input"
            required
            value={newNameInput}
            onChange={(e) => setNewNameInput(e.target.value)}
            placeholder="e.g. quantum-v2"
            className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
          />
        </div>

        <div className="space-y-1.5 pt-1">
          <label className="text-xs text-muted-foreground">
            To confirm, please type current name <strong className="text-foreground font-mono">{fullRepoName}</strong>:
          </label>
          <input
            type="text"
            data-testid="rename-confirm-name-input"
            value={confirmInput}
            onChange={(e) => setConfirmInput(e.target.value)}
            placeholder={fullRepoName}
            className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
          />
        </div>

        <div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
          <Button type="button" variant="outline" size="sm" onClick={handleModalClose}>
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            size="sm"
            data-testid="rename-submit-btn"
            disabled={submitting || !isFormValid}
          >
            {submitting ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : 'I understand, rename repository'}
          </Button>
        </div>
      </form>
    </Dialog>
  )
}
