import React from 'react'
import { MergeBox, type MergeBoxProps } from './MergeBox'

export type { MergeBoxProps }

export interface SlideToConfirmProps {
  onConfirm: () => void
  label?: string
  confirmedLabel?: string
  disabled?: boolean
  className?: string
}

/**
 * SlideToConfirm backwards-compatibility wrapper that renders the authentic GitHub MergeBox.
 */
export const SlideToConfirm: React.FC<SlideToConfirmProps> = ({
  onConfirm,
  disabled = false,
  className,
}) => {
  return (
    <MergeBox
      onMerge={() => onConfirm()}
      disabled={disabled}
      className={className}
    />
  )
}
export { MergeBox }
