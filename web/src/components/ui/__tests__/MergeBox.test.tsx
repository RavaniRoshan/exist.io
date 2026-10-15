import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MergeBox } from '../MergeBox'

describe('Canonical GitHub PR MergeBox Component', () => {
  it('renders merge box in open state with checks passed', () => {
    render(<MergeBox onMerge={() => {}} checksPassed={true} isMerged={false} />)

    expect(screen.getByText('This branch has no conflicts with the base branch')).toBeInTheDocument()
    expect(screen.getByText(/merging can be performed automatically/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create a merge commit/i })).toBeInTheDocument()
  })

  it('renders merged state when PR is already merged', () => {
    render(<MergeBox onMerge={() => {}} isMerged={true} />)

    expect(screen.getByText('Pull request successfully merged and closed')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /merge/i })).not.toBeInTheDocument()
  })

  it('allows selecting merge strategy (squash, merge, rebase)', () => {
    render(<MergeBox onMerge={() => {}} defaultMethod="merge" />)

    const dropdownTrigger = screen.getByTitle('Select merge strategy')
    fireEvent.click(dropdownTrigger)

    const squashOption = screen.getByText('Squash and merge')
    fireEvent.click(squashOption)

    expect(screen.getByRole('button', { name: /squash and merge/i })).toBeInTheDocument()
  })

  it('triggers onMerge callback with selected method and commit title/message', () => {
    const handleMerge = vi.fn()
    render(
      <MergeBox
        onMerge={handleMerge}
        defaultMethod="squash"
        defaultTitle="feat: awesome feature (#42)"
        defaultMessage="Squashed commits into main."
      />
    )

    // Click confirm button
    const mergeBtn = screen.getByRole('button', { name: /squash and merge/i })
    fireEvent.click(mergeBtn)

    // Now in confirmation state
    expect(screen.getByText('Confirm squash and merge')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /confirm squash and merge/i }))

    expect(handleMerge).toHaveBeenCalledWith(
      'squash',
      'feat: awesome feature (#42)',
      'Squashed commits into main.'
    )
  })
})
