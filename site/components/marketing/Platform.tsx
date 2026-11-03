const COLUMNS = [
  {
    index: '01',
    span: 'lg:col-span-5',
    title: 'Leases, not keys.',
    body: 'Agents don’t get passwords. They get short-lived HMAC-SHA256 capability leases scoped to refs/heads/agent/*. Every push, PR and merge re-presents the lease — then it expires. Production branches stay untouchable.',
    meta: 'pkg/agentlease/lease.go',
  },
  {
    index: '02',
    span: 'lg:col-span-4',
    title: 'A reviewer that never sleeps.',
    body: 'Every diff is scanned for leaked secrets and missing coverage before it can land. Results arrive as check runs; your rulesets decide who’s required, who approves, and who may bypass.',
    meta: 'internal/agentreviewer',
  },
  {
    index: '03',
    span: 'lg:col-span-3',
    title: 'One binary.',
    body: 'Smart-HTTP, embedded SSH, the Web UI, REST and MCP ship in a single static binary. SQLite WAL on disk. No database server, no queue, no control plane.',
    meta: 'existd · single process',
  },
] as const;

export function Platform() {
  return (
    <section id="platform" className="border-b border-line scroll-mt-20">
      <div className="container-site py-16 lg:py-24">
        <p className="eyebrow">03 — platform</p>
        <h2 className="mt-4 max-w-4xl text-4xl font-medium leading-[1.02] tracking-[-0.035em] text-ink sm:text-5xl">
          Built for agents.<br />Ready for people.
        </h2>

        <div className="mt-12 grid gap-10 lg:grid-cols-12 lg:gap-8">
          {COLUMNS.map((column) => (
            <div key={column.index} className={`border-t border-line-strong pt-6 ${column.span}`}>
              <div className="flex items-baseline justify-between gap-4">
                <span className="mono text-[12px] font-semibold text-ember">{column.index}</span>
                <span className="mono text-[10px] uppercase tracking-[0.18em] text-ink-3">{column.meta}</span>
              </div>
              <h3 className="mt-4 text-2xl font-medium leading-tight tracking-[-0.02em] text-ink sm:text-[28px]">{column.title}</h3>
              <p className="mt-3 max-w-[48ch] text-[15px] leading-relaxed text-ink-2">{column.body}</p>
            </div>
          ))}
        </div>

        <p className="mono mt-12 border-t border-line pt-5 text-[10.5px] uppercase tracking-[0.14em] text-ink-3">
          sso (oidc / saml) + scim 2.0 available · experimental · frozen for launch
        </p>
      </div>
    </section>
  );
}
