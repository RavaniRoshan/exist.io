'use client';

import { useEffect, useState } from 'react';
import { Logomark, IconGitHub, IconArrowUpRight } from './icons';

const LINKS = [
  { label: 'Product', href: '/#product' },
  { label: 'Agents', href: '/#agents' },
  { label: 'Platform', href: '/#platform' },
  { label: 'Docs', href: '/docs/' },
] as const;

export function Nav() {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (!open) return;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open]);

  return (
    <header className="sticky top-0 z-40 border-b border-line bg-paper/85 backdrop-blur-md">
      <div className="container-site flex h-14 items-center justify-between gap-6">
        <a href="/" className="flex items-center gap-2.5" aria-label="exist.io — home">
          <Logomark size={20} />
          <span className="text-[19px] font-semibold tracking-[-0.03em] text-ink">exist.io</span>
        </a>

        <nav className="hidden items-center gap-7 lg:flex" aria-label="Primary">
          {LINKS.map((link) => (
            <a key={link.label} href={link.href} className="text-[13.5px] font-medium text-ink-2 transition-colors hover:text-ember">
              {link.label}
            </a>
          ))}
        </nav>

        <div className="hidden items-center gap-2 lg:flex">
          <a
            className="flex h-9 w-9 items-center justify-center border border-transparent text-ink-2 transition-colors hover:text-ink"
            href="https://github.com/RavaniRoshan/exist.io"
            aria-label="GitHub repository"
            rel="noreferrer"
          >
            <IconGitHub width={20} height={20} />
          </a>
          <a className="btn btn-signal !h-9 px-4 text-[13px]" href="/#quickstart">
            Install
            <span className="btn-chip !h-5 !w-5" aria-hidden="true"><IconArrowUpRight width={11} height={11} strokeWidth={2} /></span>
          </a>
        </div>

        <button
          type="button"
          className="flex h-10 w-10 flex-col items-center justify-center gap-[6px] lg:hidden"
          aria-expanded={open}
          aria-controls="mobile-nav"
          aria-label={open ? 'Close menu' : 'Open menu'}
          onClick={() => setOpen((value) => !value)}
        >
          <span className={`block h-[1.5px] w-6 bg-ink transition-transform duration-300 ${open ? 'translate-y-[3.75px] rotate-45' : ''}`} />
          <span className={`block h-[1.5px] w-6 bg-ink transition-transform duration-300 ${open ? '-translate-y-[3.75px] -rotate-45' : ''}`} />
        </button>
      </div>

      {open && (
        <div id="mobile-nav" className="border-t border-line bg-paper lg:hidden">
          <nav className="container-site flex flex-col py-4" aria-label="Mobile">
            {LINKS.map((link) => (
              <a
                key={link.label}
                href={link.href}
                className="border-b border-line py-4 text-[15px] font-medium text-ink last:border-b-0"
                onClick={() => setOpen(false)}
              >
                {link.label}
              </a>
            ))}
            <a
              href="/#quickstart"
              className="btn btn-signal mt-4 justify-center"
              onClick={() => setOpen(false)}
            >
              Install exist.io
              <span className="btn-chip" aria-hidden="true"><IconArrowUpRight width={14} height={14} strokeWidth={2} /></span>
            </a>
          </nav>
        </div>
      )}
    </header>
  );
}
