'use client';

import { useEffect, useRef, useState, type KeyboardEvent } from 'react';
import {
  IconLease,
  IconClone,
  IconPush,
  IconPullRequest,
  IconEye,
  IconMerge,
  IconArrowUpRight,
} from './icons';
import { asset } from '@/lib/asset';

type Tone = 'cmd' | 'out' | 'ok' | 'dim';

const STEPS = [
  {
    id: '01',
    icon: IconLease,
    label: 'lease',
    title: 'Request a lease',
    detail: 'A short-lived HMAC capability token, scoped to a ref prefix.',
    meta: 'POST /api/v1/leases → 15m',
    lines: [
      ['$ curl -X POST /api/v1/leases', 'cmd'],
      ['  { scope: "refs/heads/agent/*", ttl: "15m" }', 'out'],
      ['→ 201 created · lease_7f3a · hmac-sha256', 'ok'],
      ['  revocable · prefix-isolated · no standing access', 'dim'],
    ],
  },
  {
    id: '02',
    icon: IconClone,
    label: 'clone',
    title: 'Clone the repo',
    detail: 'The lease authenticates Git Smart-HTTP or SSH. Nothing else.',
    meta: 'git clone …/acme/demo',
    lines: [
      ['$ git clone http://lease:***@127.0.0.1:8080/acme/demo.git', 'cmd'],
      ["Cloning into 'demo'...", 'out'],
      ['→ HEAD is now at 5e1c0aa · main', 'ok'],
      ['  transport: smart-http · no personal token', 'dim'],
    ],
  },
  {
    id: '03',
    icon: IconPush,
    label: 'push',
    title: 'Push a branch',
    detail: 'The pre-receive hook enforces the ref prefix and scans for secrets.',
    meta: 'git push origin agent/capture',
    lines: [
      ['$ git push origin agent/capture', 'cmd'],
      ['pre-receive: scope refs/heads/agent/* ✓', 'out'],
      ['pre-receive: secret scan clean ✓', 'out'],
      ['→ refs/heads/agent/capture · a91c2de', 'ok'],
    ],
  },
  {
    id: '04',
    icon: IconPullRequest,
    label: 'open pr',
    title: 'Open a pull request',
    detail: 'The agent opens the PR through MCP — not a personal token.',
    meta: 'mcp create_pull_request',
    lines: [
      ['$ mcp call create_pull_request', 'cmd'],
      ['  head: agent/capture · base: main', 'out'],
      ['→ #214 opened · lease re-presented', 'ok'],
      ['  actor: claude-code · scope: write:refs/heads/agent/*', 'dim'],
    ],
  },
  {
    id: '05',
    icon: IconEye,
    label: 'review',
    title: 'Review the diff',
    detail: 'Checks, rulesets and the reviewer agent run before merge.',
    meta: 'checks: secret-scan ✓',
    lines: [
      ['$ checks status --pr 214', 'cmd'],
      ['policy-guard    pass   0.2s', 'out'],
      ['secret-scan     pass   0.1s', 'out'],
      ['reviewer-agent  pass   0.8s', 'out'],
      ['→ 3/3 required checks · 1 approval', 'ok'],
    ],
  },
  {
    id: '06',
    icon: IconMerge,
    label: 'merge',
    title: 'Merge through policy',
    detail: 'The merge gate passes, the lease expires, the audit log records it.',
    meta: 'merge gate: 3/3 ✓',
    lines: [
      ['$ merge --pr 214', 'cmd'],
      ['gate: approvals ✓ checks ✓ secrets ✓', 'out'],
      ['→ refs/heads/main · 9b2f1c4', 'ok'],
      ['  lease revoked · audit entry a-1042 written', 'dim'],
    ],
  },
] as const;

const TONE: Record<Tone, string> = {
  cmd: 'text-patina',
  out: 'text-text-2',
  ok: 'text-text-1',
  dim: 'text-text-3',
};

export function Loop() {
  const [active, setActive] = useState(0);
  const [engaged, setEngaged] = useState(false);
  const tabRefs = useRef<Array<HTMLButtonElement | null>>([]);

  useEffect(() => {
    if (engaged) return;
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    const timer = window.setTimeout(() => {
      setActive((value) => (value + 1) % STEPS.length);
    }, 4600);
    return () => window.clearTimeout(timer);
  }, [active, engaged]);

  const engage = () => setEngaged(true);

  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    const last = STEPS.length - 1;
    let next = active;
    if (event.key === 'ArrowRight') next = active === last ? 0 : active + 1;
    else if (event.key === 'ArrowLeft') next = active === 0 ? last : active - 1;
    else if (event.key === 'Home') next = 0;
    else if (event.key === 'End') next = last;
    else return;
    event.preventDefault();
    engage();
    setActive(next);
    tabRefs.current[next]?.focus();
  };

  const step = STEPS[active];

  return (
    <section id="agents" className="relative border-b border-line scroll-mt-20">
      <span aria-hidden="true" className="crosshair left-0 top-0 hidden -translate-x-1/2 -translate-y-1/2 lg:block" />
      <span aria-hidden="true" className="crosshair right-0 top-0 hidden translate-x-1/2 -translate-y-1/2 lg:block" />
      <div className="container-site py-16 lg:py-24">
        <div className="flex flex-wrap items-end justify-between gap-6 border-b border-line pb-8">
          <div>
            <p className="eyebrow">01 — the agent loop</p>
            <h2 className="mt-4 text-4xl font-light leading-[1.02] tracking-[-0.035em] text-text-1 sm:text-5xl">
              From lease to <span className="text-patina">merge.</span>
            </h2>
          </div>
          <a href={asset('/docs/tutorials/first-agent/')} className="group flex items-center gap-2 text-[15px] font-medium text-patina hover:text-patina-hover">
            Walk through it
            <IconArrowUpRight width={16} height={16} strokeWidth={2} className="transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
          </a>
        </div>

        <div
          className="mt-8 overflow-hidden rounded-lg border border-line-strong bg-slab"
          onPointerEnter={engage}
          onFocusCapture={engage}
        >
          <div
            role="tablist"
            aria-label="The agent loop"
            className="flex overflow-x-auto border-b border-line bg-raised"
            onKeyDown={onKeyDown}
          >
            {STEPS.map((item, index) => (
              <button
                key={item.id}
                ref={(node) => {
                  tabRefs.current[index] = node;
                }}
                type="button"
                role="tab"
                id={`loop-tab-${item.id}`}
                aria-selected={active === index}
                aria-controls="loop-panel"
                tabIndex={active === index ? 0 : -1}
                className="tab min-w-max"
                onClick={() => {
                  engage();
                  setActive(index);
                }}
              >
                <span className={`mono text-[9.5px] ${active === index ? 'text-patina' : 'text-text-3'}`}>{item.id}</span>
                {item.label}
              </button>
            ))}
          </div>

          <div
            id="loop-panel"
            role="tabpanel"
            aria-labelledby={`loop-tab-${step.id}`}
            className="grid lg:grid-cols-12"
          >
            <div key={`copy-${active}`} className="rise-in border-b border-line p-6 lg:col-span-5 lg:border-b-0 lg:border-r lg:p-8">
              <div className="flex items-center justify-between gap-4">
                <span className="flex h-11 w-11 items-center justify-center rounded-md border border-line bg-raised text-patina">
                  <step.icon width={22} height={22} />
                </span>
                <span className="mono text-[10px] uppercase tracking-[0.18em] text-text-3">
                  {step.id} / 06
                </span>
              </div>
              <h3 className="mt-6 text-2xl font-medium leading-tight tracking-[-0.02em] text-text-1 sm:text-[28px]">
                {step.title}
              </h3>
              <p className="mt-3 max-w-[42ch] text-[15px] leading-relaxed text-text-2">{step.detail}</p>
              <p className="mono mt-6 inline-flex border border-line bg-raised px-2 py-1 text-[10px] text-text-2">
                {step.meta}
              </p>

              <div className="mt-8 flex items-center gap-2" aria-hidden="true">
                {STEPS.map((item, index) => (
                  <span
                    key={item.id}
                    className={`h-1 rounded-full transition-all duration-300 ${
                      index === active ? 'w-8 bg-patina' : index < active ? 'w-3 bg-patina/40' : 'w-3 bg-line'
                    }`}
                  />
                ))}
              </div>
            </div>

            <div key={`code-${active}`} className="relative min-w-0 lg:col-span-7">
              <div className="window h-full rounded-none border-0 bg-canvas shadow-none">
                <div className="window-bar bg-raised">
                  <span className="window-dots" aria-hidden="true">
                    <span />
                    <span />
                    <span />
                  </span>
                  <span className="window-title">forge — {step.label}</span>
                  <span className="mono ml-auto text-[9.5px] uppercase tracking-[0.16em] text-text-3">agent@existd</span>
                </div>
                <div className="overflow-x-auto p-5 sm:p-6">
                  <pre className="mono text-[12px] leading-[1.9]">
                    {step.lines.map(([line, tone], index) => (
                      <span
                        key={line}
                        className={`rise-in block whitespace-pre ${TONE[tone as Tone]}`}
                        style={{ animationDelay: `${index * 70}ms` }}
                      >
                        {line}
                      </span>
                    ))}
                    <span className="rise-in mt-1 block" style={{ animationDelay: `${step.lines.length * 70}ms` }}>
                      <span className="caret" aria-hidden="true" />
                    </span>
                  </pre>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
          <p className="mono text-[11px] text-text-3">
            Six steps. One short-lived credential. Zero standing privileges.
          </p>
          <p className="mono text-[10px] uppercase tracking-[0.14em] text-text-3">
            {engaged ? 'manual · select a step' : 'auto-advancing · interact to pause'}
          </p>
        </div>
      </div>
    </section>
  );
}
