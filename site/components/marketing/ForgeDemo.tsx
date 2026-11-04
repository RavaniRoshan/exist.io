'use client';

import { useEffect, useState } from 'react';
import { IconCheck } from './icons';

const ROWS = [
  {
    tool: 'lease_request',
    detail: 'scope refs/heads/agent/* · ttl 15m',
    result: 'lease ok',
  },
  {
    tool: 'git push',
    detail: 'pre-receive · secret scan clean',
    result: 'pushed',
  },
  {
    tool: 'pull_request_create',
    detail: '#214 · agent/capture → main',
    result: 'opened',
  },
  {
    tool: 'checks_run',
    detail: 'policy · secret-scan · review',
    result: '3/3',
  },
  {
    tool: 'merge_gate',
    detail: 'lease revoked · audit entry written',
    result: 'merged',
  },
] as const;

export function ForgeDemo() {
  const [step, setStep] = useState(0);
  const [playing, setPlaying] = useState(true);

  useEffect(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      setStep(ROWS.length);
      setPlaying(false);
    }
  }, []);

  useEffect(() => {
    if (!playing) return;
    const delay = step === 0 ? 650 : step >= ROWS.length ? 3400 : 900;
    const timer = window.setTimeout(() => {
      setStep((value) => (value >= ROWS.length ? 0 : value + 1));
    }, delay);
    return () => window.clearTimeout(timer);
  }, [playing, step]);

  const complete = step >= ROWS.length;

  return (
    <div className="window relative" data-testid="forge-demo">
      <div className="window-bar">
        <span className="window-dots" aria-hidden="true">
          <span />
          <span />
          <span />
        </span>
        <span className="window-title">acme/demo — agent/capture</span>
        <span className="mono ml-auto flex items-center gap-2 text-[10px] uppercase tracking-[0.16em] text-text-3">
          <span className={`status-dot !h-1.5 !w-1.5 ${playing ? '' : 'opacity-40'}`} aria-hidden="true" />
          {playing ? 'live' : 'paused'}
        </span>
      </div>

      <div className="relative p-4 sm:p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="mono text-[10px] uppercase tracking-[0.2em] text-text-3">capability lease → merge</p>
            <p className="mt-1.5 truncate text-[15px] font-medium tracking-tight text-text-1">
              agent/capture <span className="text-text-3">→</span> main
            </p>
          </div>
          <div className="flex items-center gap-1.5">
            <button
              type="button"
              className="mono flex h-8 items-center gap-1.5 rounded-md border border-line bg-slab px-2.5 text-[10px] uppercase tracking-[0.12em] text-text-2 transition-colors hover:border-line-strong hover:text-text-1"
              onClick={() => setPlaying((value) => !value)}
              aria-pressed={!playing}
            >
              {playing ? <IconPause /> : <IconPlay />}
              {playing ? 'pause' : 'play'}
            </button>
            <button
              type="button"
              className="mono flex h-8 w-8 items-center justify-center rounded-md border border-line bg-slab text-text-2 transition-colors hover:border-line-strong hover:text-text-1"
              onClick={() => {
                setStep(0);
                setPlaying(true);
              }}
              aria-label="Replay the merge sequence"
            >
              <IconReplay />
            </button>
          </div>
        </div>

        <div className="mt-4 flex gap-1" aria-hidden="true">
          {ROWS.map((row, index) => (
            <button
              key={row.tool}
              type="button"
              tabIndex={-1}
              className="h-1 flex-1 overflow-hidden rounded-full bg-raised"
              onClick={() => {
                setStep(index + 1);
                setPlaying(false);
              }}
              aria-label={`Jump to ${row.tool}`}
            >
              <span
                className="block h-full rounded-full bg-patina transition-transform duration-500 ease-out"
                style={{ transform: `scaleX(${index < step ? 1 : 0})`, transformOrigin: 'left' }}
              />
            </button>
          ))}
        </div>

        <ol className="mt-4 flex flex-col" aria-label="Agent merge sequence">
          {ROWS.map((row, index) => {
            const done = index < step;
            const active = index === step;
            return (
              <li
                key={row.tool}
                className={`flex items-center gap-3 border-t border-line py-2.5 first:border-t-0 ${
                  done || active ? 'opacity-100' : 'opacity-45'
                } ${active ? 'rise-in' : ''}`}
              >
                <span
                  aria-hidden="true"
                  className={`flex h-5 w-5 flex-none items-center justify-center rounded-full border ${
                    done
                      ? 'border-patina bg-patina text-patina-ink'
                      : active
                        ? 'border-patina text-patina'
                        : 'border-line text-transparent'
                  }`}
                >
                  {done ? (
                    <IconCheck width={11} height={11} />
                  ) : active ? (
                    <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-patina" />
                  ) : (
                    <span className="h-1 w-1 rounded-full bg-line-strong" />
                  )}
                </span>
                <span className="min-w-0 flex-1">
                  <span className="mono block truncate text-[11.5px] text-text-1">{row.tool}</span>
                  <span className="mono block truncate text-[10px] text-text-3">{row.detail}</span>
                </span>
                <span
                  className={`mono flex-none text-[9.5px] uppercase tracking-[0.14em] ${
                    done ? 'text-patina' : active ? 'text-brass' : 'text-text-3'
                  }`}
                >
                  {done ? row.result : active ? 'running' : 'queued'}
                </span>
              </li>
            );
          })}
        </ol>

        <div className="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-line pt-3">
          <span className="mono text-[10px] text-text-3">lease 15m · hmac-sha256 · revocable</span>
          <span className="mono rounded border border-line bg-raised px-1.5 py-0.5 text-[9.5px] uppercase tracking-[0.14em] text-text-2">
            audit #a-1042
          </span>
        </div>

        <p className="sr-only" aria-live="polite">
          {complete
            ? 'Merge gate passed. Lease revoked, audit entry written, main updated.'
            : `Step ${Math.min(step, ROWS.length)} of ${ROWS.length}.`}
        </p>

        {complete && (
          <div className="pointer-events-none absolute right-5 top-16 rotate-[-7deg] border-2 border-patina px-3 py-1.5 text-center text-patina">
            <span className="mono block text-[13px] font-semibold uppercase tracking-[0.22em]">merged</span>
            <span className="mono block text-[8.5px] uppercase tracking-[0.14em] opacity-80">refs/heads/main</span>
          </div>
        )}
      </div>
    </div>
  );
}

function IconPlay() {
  return (
    <svg width="10" height="10" viewBox="0 0 10 10" fill="currentColor" aria-hidden="true">
      <path d="M2 1.2v7.6L8.6 5 2 1.2Z" />
    </svg>
  );
}

function IconPause() {
  return (
    <svg width="10" height="10" viewBox="0 0 10 10" fill="currentColor" aria-hidden="true">
      <rect x="2" y="1.4" width="2.1" height="7.2" rx="0.4" />
      <rect x="5.9" y="1.4" width="2.1" height="7.2" rx="0.4" />
    </svg>
  );
}

function IconReplay() {
  return (
    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M3 12a9 9 0 1 0 3-6.7L3 8" />
      <path d="M3 3v5h5" />
    </svg>
  );
}
