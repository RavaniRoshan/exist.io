'use client';

import { useEffect, useState } from 'react';
import { IconCopy, IconCheck } from './icons';

export function CopyButton({ text, label = 'Copy to clipboard' }: { text: string; label?: string }) {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!copied) return;
    const id = window.setTimeout(() => setCopied(false), 1600);
    return () => window.clearTimeout(id);
  }, [copied]);

  return (
    <button
      type="button"
      className="cmd-copy"
      aria-label={copied ? 'Copied' : label}
      onClick={async () => {
        try {
          await navigator.clipboard.writeText(text);
          setCopied(true);
        } catch {
          /* clipboard unavailable — leave state unchanged */
        }
      }}
    >
      {copied ? <IconCheck width={16} height={16} /> : <IconCopy width={16} height={16} />}
    </button>
  );
}
