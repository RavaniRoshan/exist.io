'use client';

import { useEffect, useRef, useState } from 'react';
import { CopyButton } from './CopyButton';

const STEPS = [
  {
    label: 'install',
    command: 'curl -fsSL https://raw.githubusercontent.com/RavaniRoshan/exist.io/main/scripts/install.sh | bash',
    output: ['resolving v0.6.0-beta', 'sha-256 manifest verified', 'installed existd to ~/bin'],
  },
  {
    label: 'bootstrap',
    command: 'existd bootstrap --email admin@example.com --namespace acme --data-dir ./data',
    output: ['namespace acme created', 'repository acme/demo created', 'admin api key printed once'],
  },
  {
    label: 'run',
    command: 'existd',
    output: ['ui · rest · smart-http · mcp on :8080', 'embedded ssh on :2222', 'ready — open http://localhost:8080'],
  },
] as const;

const ALL = STEPS.map((step) => step.command).join('\n');

export function QuickstartTerminal() {
  const [step, setStep] = useState<number>(0);
  const [chars, setChars] = useState<number>(STEPS[0].command.length);
  const [out, setOut] = useState<number>(STEPS[0].output.length);
  const [started, setStarted] = useState(false);
  const [done, setDone] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const bodyRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const node = ref.current;
    if (!node) return;
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      setStarted(true);
      setStep(STEPS.length);
      setDone(true);
      return;
    }
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setStarted(true);
          observer.disconnect();
        }
      },
      { threshold: 0.25 },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (!started || done) return;
    if (step >= STEPS.length) {
      setDone(true);
      return;
    }
    const current = STEPS[step];
    if (chars < current.command.length) {
      const timer = window.setTimeout(() => setChars((value) => value + 1), 11);
      return () => window.clearTimeout(timer);
    }
    if (out < current.output.length) {
      const timer = window.setTimeout(() => setOut((value) => value + 1), 230);
      return () => window.clearTimeout(timer);
    }
    const timer = window.setTimeout(() => {
      setStep((value) => value + 1);
      setChars(0);
      setOut(0);
    }, 720);
    return () => window.clearTimeout(timer);
  }, [chars, done, out, started, step]);

  useEffect(() => {
    const body = bodyRef.current;
    if (body) body.scrollTop = body.scrollHeight;
  }, [chars, out, step]);

  const skipped = done || step >= STEPS.length;
  const status = !started ? 'install complete' : skipped ? 'complete' : `step ${step + 1} / ${STEPS.length} · typing`;

  return (
    <div ref={ref} className="window">
      <div className="sr-only">
        <p>Terminal commands to install, bootstrap, and run exist.io:</p>
        {STEPS.map((item) => (
          <p key={item.label}>
            {item.label}: <code>{item.command}</code>
          </p>
        ))}
      </div>
      <div className="window-bar">
        <span className="window-dots" aria-hidden="true">
          <span />
          <span />
          <span />
        </span>
        <span className="window-title">bash — existd</span>
        <span className="mono ml-auto text-[9.5px] uppercase tracking-[0.16em] text-text-3">{status}</span>
        <CopyButton text={ALL} label="Copy all install commands" />
      </div>

      <div ref={bodyRef} aria-hidden="true" className="min-h-[320px] overflow-y-auto bg-canvas p-5 sm:p-6">
        <div className="mono text-[11.5px] leading-[1.8]">
          {STEPS.map((item, index) => {
            const isCurrent = index === step;
            const complete = index < step || skipped;
            if (index > step && !skipped) return null;
            const typed = complete ? item.command : item.command.slice(0, chars);
            const shownOutput = complete ? item.output.length : isCurrent ? out : 0;
            return (
              <div key={item.label} className={`mb-4 last:mb-0 ${isCurrent && !skipped ? 'rise-in' : ''}`}>
                <p className="flex gap-2">
                  <span className="select-none text-text-3">$</span>
                  <span className="min-w-0 whitespace-pre-wrap break-all text-text-1">
                    {typed}
                    {isCurrent && started && !complete ? <span className="caret ml-0.5" aria-hidden="true" /> : null}
                  </span>
                </p>
                <div className="mt-1 flex flex-col gap-0.5 pl-4">
                  {item.output.slice(0, shownOutput).map((line) => (
                    <p key={line} className="rise-in flex gap-2 text-text-3">
                      <span className="select-none text-patina" aria-hidden="true">✓</span>
                      <span>{line}</span>
                    </p>
                  ))}
                </div>
                {isCurrent && started && !complete && out >= item.output.length ? (
                  <p className="mt-1 pl-4 text-[10px] uppercase tracking-[0.16em] text-text-3">running…</p>
                ) : null}
              </div>
            );
          })}
        </div>
      </div>

      <div className="flex items-center justify-between gap-3 border-t border-line bg-raised px-4 py-2.5">
        <span className="mono text-[10px] text-text-3">
          {skipped ? 'admin api key prints once — store it' : 'auto-typing · scroll to run'}
        </span>
        <button
          type="button"
          className="mono text-[10px] uppercase tracking-[0.14em] text-patina transition-colors hover:text-patina-hover"
          onClick={() => {
            if (!skipped) {
              setStep(STEPS.length);
              setDone(true);
              return;
            }
            setStep(0);
            setChars(0);
            setOut(0);
            setDone(false);
            setStarted(true);
          }}
        >
          {skipped ? 'run again' : 'skip'}
        </button>
      </div>
    </div>
  );
}
