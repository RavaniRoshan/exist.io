import { describe, it, vi } from 'vitest'
import { render } from '@testing-library/react'
import React from 'react'
import { writeFileSync } from 'node:fs'
import { Command, CommandDialog, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/primitives/command'

describe('bisect10', () => {
  it('t19 with commanddialog wrapper', () => {
    const { container } = render(
      <CommandDialog open={true} onOpenChange={vi.fn()} title="t" description="d">
        <Command><CommandInput placeholder="Search…" /><CommandList><CommandGroup heading="Navigate"><CommandItem id="code" textValue="Code">Code</CommandItem></CommandGroup></CommandList></Command>
      </CommandDialog>
    )
    writeFileSync('/tmp/opencode/bisect10-t19.html', document.body.innerHTML)
    console.log('t19 menuitem=', (document.body.innerHTML.match(/role="menuitem"/g) ?? []).length)
  })
  it('t20 with commanddialog and two items', () => {
    render(
      <CommandDialog open={true} onOpenChange={vi.fn()} title="t" description="d">
        <Command><CommandInput placeholder="Search…" /><CommandList><CommandGroup heading="Navigate"><CommandItem id="code" textValue="Code">Code</CommandItem><CommandItem id="prs" textValue="Pull requests">Pull requests</CommandItem></CommandGroup></CommandList></Command>
      </CommandDialog>
    )
    console.log('t20 menuitem=', (document.body.innerHTML.match(/role="menuitem"/g) ?? []).length, 'Code=', document.body.innerHTML.includes('Code'))
  })
})
