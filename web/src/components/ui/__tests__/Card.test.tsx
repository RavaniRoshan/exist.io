import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { Box, BoxHeader, BoxTitle, BoxBody, BoxRow, BoxFooter } from '../Card'

describe('Primer Box Component System', () => {
  it('renders Box container with 6px rounded-md border', () => {
    const { container } = render(<Box data-testid="test-box">Box Content</Box>)
    const box = screen.getByTestId('test-box')
    expect(box).toBeInTheDocument()
    expect(box).toHaveTextContent('Box Content')
    expect(container.firstChild).toHaveClass('rounded-md')
    expect(container.firstChild).toHaveClass('border')
  })

  it('renders BoxHeader, BoxTitle, BoxBody, BoxRow, and BoxFooter correctly', () => {
    render(
      <Box>
        <BoxHeader data-testid="header">
          <BoxTitle>Repository Overview</BoxTitle>
        </BoxHeader>
        <BoxBody data-testid="body">
          <BoxRow data-testid="row">Row Item 1</BoxRow>
        </BoxBody>
        <BoxFooter data-testid="footer">Footer Info</BoxFooter>
      </Box>
    )

    expect(screen.getByTestId('header')).toBeInTheDocument()
    expect(screen.getByText('Repository Overview')).toBeInTheDocument()
    expect(screen.getByTestId('body')).toBeInTheDocument()
    expect(screen.getByTestId('row')).toHaveTextContent('Row Item 1')
    expect(screen.getByTestId('footer')).toHaveTextContent('Footer Info')
  })

  it('supports flush mode without overflow-hidden', () => {
    const { container } = render(<Box flush>Flush Box</Box>)
    expect(container.firstChild).not.toHaveClass('overflow-hidden')
  })
})
