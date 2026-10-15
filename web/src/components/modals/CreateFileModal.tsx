import React, { useState } from 'react'
import { X, FileCode, GitCommit, Check, AlertCircle, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { commitFile } from '@/lib/api'

export interface CreateFileModalProps {
  isOpen: boolean
  onClose: () => void
  owner: string
  name: string
  branch?: string
  initialPath?: string
  onCommitSuccess: () => void
}

export const CreateFileModal: React.FC<CreateFileModalProps> = ({
  isOpen,
  onClose,
  owner,
  name,
  branch = 'main',
  initialPath = 'README.md',
  onCommitSuccess,
}) => {
  const [filePath, setFilePath] = useState(initialPath)
  const [content, setContent] = useState(
    initialPath.toLowerCase() === 'readme.md'
      ? `# ${name}\n\nA sovereign repository on exist.io.\n`
      : ''
  )
  const [message, setMessage] = useState(`Create ${initialPath}`)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (!isOpen) return null

  const handlePathChange = (val: string) => {
    setFilePath(val)
    if (!message || message.startsWith('Create ')) {
      setMessage(`Create ${val || 'file'}`)
    }
  }

  const handleCommit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!filePath.trim()) {
      setError('Please provide a file name and path.')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const res = await commitFile({
        owner,
        name,
        branch,
        path: filePath.trim(),
        content,
        message: message.trim() || `Create ${filePath.trim()}`,
      })

      if (!res.ok) {
        setError(res.error || 'Failed to commit file to repository')
        setLoading(false)
        return
      }

      onCommitSuccess()
      onClose()
    } catch (err: any) {
      setError(err.message || 'Network error committing file')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-xs animate-in fade-in duration-150"
      role="dialog"
      aria-modal="true"
      aria-labelledby="create-file-modal-title"
    >
      <div
        className="relative w-full max-w-2xl bg-card rounded-md border border-border shadow-2xl overflow-hidden flex flex-col max-h-[90vh]"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Modal Header */}
        <div className="flex items-center justify-between px-4 py-3 bg-secondary/40 border-b border-border">
          <div className="flex items-center gap-2">
            <FileCode className="w-4 h-4 text-primary" />
            <h2 id="create-file-modal-title" className="text-sm font-semibold text-foreground">
              Create new file
            </h2>
          </div>
          <button
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground p-1 rounded-sm transition-colors cursor-pointer"
            aria-label="Close dialog"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Modal Form */}
        <form onSubmit={handleCommit} className="p-4 sm:p-5 flex-1 overflow-y-auto space-y-4 text-xs">
          {error && (
            <div className="p-3 rounded-md bg-destructive/15 border border-destructive/30 text-destructive flex items-center gap-2">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Breadcrumb File Path Input */}
          <div className="space-y-1.5">
            <label htmlFor="file-path-input" className="font-semibold text-foreground">
              Name your file
            </label>
            <div className="flex items-center rounded-md border border-border bg-input px-3 py-1.5 focus-within:border-primary focus-within:ring-1 focus-within:ring-ring font-mono text-xs">
              <span className="text-muted-foreground select-none">
                {owner} / {name} /&nbsp;
              </span>
              <input
                id="file-path-input"
                type="text"
                value={filePath}
                onChange={(e) => handlePathChange(e.target.value)}
                placeholder="name your file, e.g. README.md, src/main.go..."
                className="flex-1 bg-transparent border-none outline-none text-foreground font-mono"
                required
                autoFocus
              />
            </div>
          </div>

          {/* File Content Area */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <label htmlFor="file-content-textarea" className="font-semibold text-foreground">
                Edit file content
              </label>
              <span className="text-[11px] text-muted-foreground font-mono">
                {content.length} characters • {content.split('\n').length} lines
              </span>
            </div>
            <textarea
              id="file-content-textarea"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Write your file content here..."
              rows={12}
              className="w-full bg-input text-foreground font-mono text-xs p-3 rounded-md border border-border focus:border-primary focus:ring-1 focus:ring-ring outline-none leading-relaxed resize-y"
            />
          </div>

          {/* Commit Changes Section */}
          <div className="p-4 rounded-md bg-secondary/30 border border-border space-y-3">
            <div className="flex items-center gap-2">
              <GitCommit className="w-4 h-4 text-primary" />
              <h3 className="font-semibold text-foreground">Commit changes</h3>
            </div>

            <div className="space-y-1">
              <label htmlFor="commit-msg-input" className="text-[11px] text-muted-foreground">
                Commit message
              </label>
              <input
                id="commit-msg-input"
                type="text"
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                className="w-full h-8 px-3 rounded-md bg-input border border-border text-foreground font-mono text-xs focus:border-primary focus:ring-1 focus:ring-ring outline-none"
                placeholder="Commit message"
                required
              />
            </div>

            <div className="flex items-center gap-2 text-muted-foreground pt-1">
              <div className="w-2 h-2 rounded-full bg-emerald-500" />
              <span>
                Commit directly to the <strong className="text-foreground font-mono">{branch}</strong> branch
              </span>
            </div>
          </div>

          {/* Modal Action Buttons */}
          <div className="flex items-center justify-end gap-2 pt-2 border-t border-border">
            <Button type="button" variant="outline" size="sm" onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button type="submit" variant="primary" size="sm" disabled={loading} className="gap-1.5">
              {loading ? (
                <>
                  <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                  <span>Committing...</span>
                </>
              ) : (
                <>
                  <Check className="w-3.5 h-3.5" />
                  <span>Commit new file</span>
                </>
              )}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
