'use client';

import { useEffect, useState } from 'react';
import { ScrollProgress } from './ScrollProgress';
import { ThemeToggle } from './ThemeToggle';
import { Logomark, IconGitHub, IconArrowUpRight } from './icons';
import { asset, basePath } from '@/lib/asset';

const LINKS = [
  { label: 'Product', href: `${basePath}/#product` },
  { label: 'Agents', href: `${basePath}/#agents` },
  { label: 'Platform', href: `${basePath}/#platform` },
  { label: 'Docs', href: asset('/docs/') },
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
    <header className="sticky top-0 z-40 border-b border-line bg-canvas/85 backdrop-blur-md">
      <div className="container-site flex h-14 items-center justify-between gap-6">
        <a href={asset('/')} className="flex items-center gap-2.5" aria-label="exist.io — home">
          <Logomark size={20} />
          <span className="text-[19px] font-semibold tracking-[-0.03em] text-text-1">exist.io</span>
          <span className="mono ml-1 hidden rounded-sm border border-line bg-slab px-1.5 py-0.5 text-[9px] uppercase tracking-[0.12em] text-text-3 sm:inline">
            v0.6.0-beta
          </span>
        </a>

        <nav className="hidden items-center gap-7 lg:flex" aria-label="Primary">
          {LINKS.map((link) => (
            <a key={link.label} href={link.href} className="group relative text-[13.5px] font-medium text-text-2 transition-colors hover:text-text-1">
              {link.label}
              <span aria-hidden="true" className="absolute -bottom-1.5 left-0 h-px w-full origin-left scale-x-0 bg-patina transition-transform duration-200 group-hover:scale-x-100" />
            </a>
          ))}
        </nav>

        <div className="hidden items-center gap-2 lg:flex">
          <ThemeToggle />
          <a
            className="flex h-9 w-9 items-center justify-center rounded-md border border-transparent text-text-2 transition-colors hover:border-line hover:bg-slab hover:text-text-1"
            href="https://github.com/RavaniRoshan/exist.io"
            aria-label="GitHub repository"
            rel="noreferrer"
          >
            <IconGitHub width={20} height={20} />
          </a>
          <a className="btn btn-signal !h-9 px-4 text-[13px]" href={`${basePath}/#quickstart`}>
            Install
            <span className="btn-chip !h-5 !w-5" aria-hidden="true"><IconArrowUpRight width={11} height={11} strokeWidth={2} /></span>
          </a>
        </div>

        <div className="flex items-center gap-1 lg:hidden">
          <ThemeToggle />
          <button
            type="button"
            className="flex h-10 w-10 flex-col items-center justify-center gap-[6px]"
            aria-expanded={open}
            aria-controls="mobile-nav"
            aria-label={open ? 'Close menu' : 'Open menu'}
            onClick={() => setOpen((value) => !value)}
          >
            <span className={`block h-[1.5px] w-6 bg-text-1 transition-transform duration-300 ${open ? 'translate-y-[3.75px] rotate-45' : ''}`} />
            <span className={`block h-[1.5px] w-6 bg-text-1 transition-transform duration-300 ${open ? '-translate-y-[3.75px] -rotate-45' : ''}`} />
          </button>
        </div>
      </div>

      {open && (
        <div id="mobile-nav" className="border-t border-line bg-canvas lg:hidden">
          <nav className="container-site flex flex-col py-4" aria-label="Mobile">
            {LINKS.map((link) => (
              <a
                key={link.label}
                href={link.href}
                className="flex items-center justify-between border-b border-line py-4 text-[15px] font-medium text-text-1 last:border-b-0"
                onClick={() => setOpen(false)}
              >
                {link.label}
                <span aria-hidden="true" className="mono text-text-3">↗</span>
              </a>
            ))}
            <a
              href={`${basePath}/#quickstart`}
              className="btn btn-signal mt-4 justify-center"
              onClick={() => setOpen(false)}
            >
              Install exist.io
              <span className="btn-chip" aria-hidden="true"><IconArrowUpRight width={14} height={14} strokeWidth={2} /></span>
            </a>
          </nav>
        </div>
      )}
      <ScrollProgress />
    </header>
  );
}
