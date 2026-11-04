'use client';

import { useEffect, useState } from 'react';
import { useTheme } from 'fumadocs-ui/provider/base';
import { IconMoon, IconSun } from './icons';

/**
 * One theme switch for the whole property. Fumadocs' provider persists the
 * choice, so flipping the landing also flips the docs (and back).
 */
export function ThemeToggle({ className = '' }: { className?: string }) {
  const { resolvedTheme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => setMounted(true), []);

  const dark = mounted && resolvedTheme === 'dark';

  return (
    <button
      type="button"
      className={`flex h-9 w-9 items-center justify-center rounded-md border border-transparent text-text-2 transition-colors hover:border-line hover:bg-slab hover:text-text-1 ${className}`}
      onClick={() => setTheme(dark ? 'light' : 'dark')}
      aria-label={dark ? 'Switch to light theme' : 'Switch to dark theme'}
      title={dark ? 'Switch to light theme' : 'Switch to dark theme'}
    >
      {dark ? <IconSun width={18} height={18} /> : <IconMoon width={18} height={18} />}
    </button>
  );
}
