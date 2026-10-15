import React, { useState } from 'react'
import { BookOpen, Lock, Globe, AlertCircle, Loader2 } from 'lucide-react'
import { Dialog, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { createRepository, type RepositoryRecord } from '@/lib/api'

export interface CreateRepoModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: (repo: RepositoryRecord) => void
  defaultOwner?: string
}

export const CreateRepoModal: React.FC<CreateRepoModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
  defaultOwner = 'shiva',
}) => {
  const [owner, setOwner] = useState(defaultOwner)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [visibility, setVisibility] = useState<'public' | 'private'>('public')
  const [defaultBranch, setDefaultBranch] = useState('main')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    const trimmedName = name.trim().toLowerCase().replace(/[^a-z0-9._-]/g, '-')
    if (!trimmedName) {
      setError('Please provide a valid repository name.')
      return
    }

    setIsSubmitting(true)
    const result = await createRepository({
      owner: owner.trim() || 'shiva',
      name: trimmedName,
      description: description.trim(),
      visibility,
      default_branch: defaultBranch.trim() || 'main',
    })
    setIsSubmitting(false)

    if (result.ok && result.repo) {
      onSuccess(result.repo)
      setName('')
      setDescription('')
      onClose()
    } else {
      setError(result.error || 'Failed to create repository')
    }
  }

  return (
    <Dialog isOpen={isOpen} onClose={onClose} maxWidth="lg">
      <DialogHeader>
        <DialogTitle className="flex items-center gap-2">
          <BookOpen className="w-4 h-4 text-primary" />
          <span>Create a new repository</span>
        </DialogTitle>
        <DialogDescription>
          A repository contains all project files, revision history, and agent policies hosted on your sovereign exist.io cluster.
        </DialogDescription>
      </DialogHeader>

      <form onSubmit={handleSubmit} className="space-y-4 pt-2">
        {error && (
          <div className="p-3 text-xs rounded-md bg-destructive/10 border border-destructive/20 text-destructive flex items-center gap-2">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {/* Owner & Repo Name */}
        <div className="grid grid-cols-1 sm:grid-cols-12 gap-3 items-end">
          <div className="sm:col-span-4 space-y-1">
            <label className="text-xs font-medium text-foreground">Owner</label>
            <input
              type="text"
              value={owner}
              onChange={(e) => setOwner(e.target.value)}
              className="w-full px-2.5 py-1.5 text-xs rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
              placeholder="shiva"
            />
          </div>
          <div className="sm:col-span-1 text-center pb-2 text-muted-foreground font-mono text-sm hidden sm:block">
            /
          </div>
          <div className="sm:col-span-7 space-y-1">
            <label className="text-xs font-medium text-foreground">
              Repository name <span className="text-destructive">*</span>
            </label>
            <input
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. backend-api or exist-docs"
              className="w-full px-2.5 py-1.5 text-xs rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
            />
          </div>
        </div>

        {/* Description */}
        <div className="space-y-1">
          <label className="text-xs font-medium text-foreground">
            Description <span className="text-[11px] text-muted-foreground font-normal">(optional)</span>
          </label>
          <input
            type="text"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Brief description of your repository"
            className="w-full px-2.5 py-1.5 text-xs rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
          />
        </div>

        {/* Visibility */}
        <div className="space-y-2 pt-2 border-t border-border">
          <label className="text-xs font-medium text-foreground block">Visibility</label>
          <div className="space-y-2">
            <label className="flex items-start gap-3 p-2.5 rounded-md border border-border hover:bg-secondary/40 cursor-pointer transition-colors">
              <input
                type="radio"
                name="visibility"
                checked={visibility === 'public'}
                onChange={() => setVisibility('public')}
                className="mt-0.5 text-primary focus:ring-primary"
              />
              <div className="space-y-0.5">
                <div className="flex items-center gap-1.5 text-xs font-medium text-foreground">
                  <Globe className="w-3.5 h-3.5 text-muted-foreground" />
                  <span>Public</span>
                </div>
                <p className="text-[11px] text-muted-foreground">
                  Anyone on your local network or cluster can view and clone this repository.
                </p>
              </div>
            </label>

            <label className="flex items-start gap-3 p-2.5 rounded-md border border-border hover:bg-secondary/40 cursor-pointer transition-colors">
              <input
                type="radio"
                name="visibility"
                checked={visibility === 'private'}
                onChange={() => setVisibility('private')}
                className="mt-0.5 text-primary focus:ring-primary"
              />
              <div className="space-y-0.5">
                <div className="flex items-center gap-1.5 text-xs font-medium text-foreground">
                  <Lock className="w-3.5 h-3.5 text-muted-foreground" />
                  <span>Private</span>
                </div>
                <p className="text-[11px] text-muted-foreground">
                  You choose who can see and commit to this repository. Authenticated via PAT or SSH key.
                </p>
              </div>
            </label>
          </div>
        </div>

        {/* Default Branch */}
        <div className="space-y-1 pt-1">
          <label className="text-xs font-medium text-foreground">Default branch</label>
          <input
            type="text"
            value={defaultBranch}
            onChange={(e) => setDefaultBranch(e.target.value)}
            className="w-32 px-2.5 py-1 text-xs rounded-md bg-input border border-border text-foreground font-mono"
            placeholder="main"
          />
        </div>

        <DialogFooter className="pt-2">
          <Button type="button" variant="secondary" size="sm" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" size="sm" disabled={isSubmitting || !name.trim()}>
            {isSubmitting ? (
              <>
                <Loader2 className="w-3.5 h-3.5 animate-spin mr-1.5" />
                <span>Creating...</span>
              </>
            ) : (
              <span>Create repository</span>
            )}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  )
}
