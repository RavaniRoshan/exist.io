import React, { useEffect } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface DialogProps {
  isOpen?: boolean
  open?: boolean
  onClose?: () => void
  onOpenChange?: (open: boolean) => void
  title?: React.ReactNode
  description?: React.ReactNode
  children: React.ReactNode
  className?: string
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl'
}

/**
 * GitHub Primer Dialog/Modal.
 * Features 8px (rounded-lg) corners, crisp 1px borders, and structured header/body.
 */
export const Dialog: React.FC<DialogProps> = ({
  isOpen: propIsOpen,
  open: propOpen,
  onClose,
  onOpenChange,
  title,
  description,
  children,
  className,
  maxWidth = 'md',
}) => {
  const isCurrentlyOpen = propIsOpen !== undefined ? propIsOpen : Boolean(propOpen)
  const handleClose = () => {
    onClose?.()
    onOpenChange?.(false)
  }

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isCurrentlyOpen) handleClose()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isCurrentlyOpen])

  const maxW = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-xl',
    '2xl': 'max-w-2xl',
  }[maxWidth]

  return (
    <AnimatePresence>
      {isCurrentlyOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.15 }}
            onClick={handleClose}
            className="fixed inset-0 bg-background/70 backdrop-blur-xs"
          />

            {/* Modal Container */}
            <motion.div
              role="dialog"
              aria-modal="true"
              aria-labelledby={title ? 'dialog-title' : undefined}
              initial={{ opacity: 0, scale: 0.98, y: 6 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.98, y: 6 }}
              transition={{ type: 'spring', damping: 30, stiffness: 450 }}
              className={cn(
                'relative w-full rounded-md border border-border bg-card text-card-foreground shadow-2xl z-10 overflow-hidden',
                maxW,
                className
              )}
            >
              {/* Modal Header */}
              {(title || description) && (
                <div className="px-4 py-3 border-b border-border bg-secondary/20 flex items-start justify-between gap-3">
                  <div className="space-y-0.5">
                    {title && <h3 id="dialog-title" className="text-sm font-semibold text-foreground leading-tight">{title}</h3>}
                    {description && <p className="text-xs text-muted-foreground leading-normal">{description}</p>}
                  </div>
                  <button
                    onClick={handleClose}
                    aria-label="Close dialog"
                    className="p-1 text-muted-foreground hover:text-foreground rounded-md hover:bg-secondary/60 transition-colors cursor-pointer"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              )}

            {/* Modal Body */}
            <div className="p-4">
              {children}
            </div>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  )
}

export const DialogHeader: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className }) => (
  <div className={cn('space-y-1 pb-3 mb-3 border-b border-border', className)}>{children}</div>
)

export const DialogTitle: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className }) => (
  <h3 className={cn('text-sm font-semibold text-foreground leading-tight', className)}>{children}</h3>
)

export const DialogDescription: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className }) => (
  <p className={cn('text-xs text-muted-foreground leading-normal', className)}>{children}</p>
)

export const DialogFooter: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className }) => (
  <div className={cn('flex items-center justify-end gap-2 pt-3 mt-3 border-t border-border', className)}>{children}</div>
)
