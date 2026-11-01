import React from 'react'
import { cn } from '@/lib/utils'
import {
  Dialog as PrimitiveDialog,
  DialogHeader as PrimitiveDialogHeader,
  DialogTitle as PrimitiveDialogTitle,
  DialogDescription as PrimitiveDialogDescription,
} from '@/components/primitives/dialog'

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

const widthMap = {
  sm: 'sm:max-w-sm',
  md: 'sm:max-w-md',
  lg: 'sm:max-w-lg',
  xl: 'sm:max-w-xl',
  '2xl': 'sm:max-w-2xl',
} as const

/** Forge dialog: React Aria focus management behind the existing API. */
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
  const isOpen = propIsOpen !== undefined ? propIsOpen : Boolean(propOpen)

  return (
    <PrimitiveDialog
      isOpen={isOpen}
      onOpenChange={(next) => {
        onOpenChange?.(next)
        if (!next) onClose?.()
      }}
      className={cn('gap-0 p-0 border border-hairline', widthMap[maxWidth], className)}
    >
      {title && (
        <PrimitiveDialogHeader className="border-b border-hairline px-4 py-3">
          <PrimitiveDialogTitle className="text-title-3">{title}</PrimitiveDialogTitle>
          {description && (
            <PrimitiveDialogDescription className="text-caption text-text-2">
              {description}
            </PrimitiveDialogDescription>
          )}
        </PrimitiveDialogHeader>
      )}
      <div className="p-4">{children}</div>
    </PrimitiveDialog>
  )
}

export const DialogHeader: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className }) => (
  <div className={cn('space-y-1 pb-3 mb-3 border-b border-hairline', className)}>{children}</div>
)

export const DialogTitle: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className }) => (
  <h3 className={cn('text-title-3 text-text-1', className)}>{children}</h3>
)

export const DialogDescription: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className }) => (
  <p className={cn('text-caption text-text-2', className)}>{children}</p>
)

export const DialogFooter: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className }) => (
  <div className={cn('flex items-center justify-end gap-2 pt-3 mt-3 border-t border-hairline', className)}>{children}</div>
)
