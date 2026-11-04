import { IconArrowUpRight } from './icons';
import { asset } from '@/lib/asset';

const COLUMNS = [
  {
    index: '01',
    span: 'lg:col-span-5',
    title: 'Leases, not keys.',
    body: 'Agents don’t get passwords. They get short-lived HMAC-SHA256 capability leases scoped to refs/heads/agent/*. Every push, PR and merge re-presents the lease — then it expires. Production branches stay untouchable.',
    meta: 'pkg/agentlease/lease.go',
    href: asset('/docs/how-to/mcp-setup/'),
  },
  {
    index: '02',
    span: 'lg:col-span-4',
    title: 'A reviewer that never sleeps.',
    body: 'Every diff is scanned for leaked secrets and missing coverage before it can land. Results arrive as check runs; your rulesets decide who’s required, who approves, and who may bypass.',
    meta: 'internal/agentreviewer',
    href: asset('/docs/explanation/merge-engine/'),
  },
  {
    index: '03',
    span: 'lg:col-span-3',
    title: 'One binary.',
    body: 'Smart-HTTP, embedded SSH, the Web UI, REST and MCP ship in a single static binary. SQLite WAL on disk. No database server, no queue, no control plane.',
    meta: 'existd · single process',
    href: asset('/docs/getting-started/'),
  },
] as const;

export function Platform() {
  return (
    <section id="platform" className="relative border-b border-line scroll-mt-20">
      <div className="container-site py-16 lg:py-24">
        <div className="flex flex-wrap items-end justify-between gap-6">
          <div>
            <p className="eyebrow">03 — platform</p>
            <h2 className="mt-4 max-w-4xl text-4xl font-light leading-[1.02] tracking-[-0.035em] text-text-1 sm:text-5xl">
              Built for <span className="text-patina">agents.</span><br />Ready for people.
            </h2>
          </div>
          <p className="mono max-w-xs text-[10.5px] leading-relaxed text-text-3">
            three guarantees, enforced in the daemon — not in a client library
          </p>
        </div>

        <div className="mt-12 grid gap-4 lg:grid-cols-12 lg:gap-5">
          {COLUMNS.map((column, index) => (
            <a
              key={column.index}
              href={column.href}
              className={`group relative flex min-h-[260px] flex-col rounded-lg border border-line-strong p-6 transition-colors duration-200 hover:border-patina hover:bg-patina hover:text-patina-ink focus-visible:border-patina focus-visible:bg-patina focus-visible:text-patina-ink ${column.span} ${
                index === 1 ? 'lg:mt-8' : index === 2 ? 'lg:mt-16' : ''
              }`}
            >
              <div className="flex items-baseline justify-between gap-4">
                <span className="mono text-[12px] font-semibold text-patina transition-colors group-hover:text-patina-ink/70 group-focus-visible:text-patina-ink/70">
                  {column.index}
                </span>
                <span className="mono text-[10px] uppercase tracking-[0.16em] text-text-3 transition-colors group-hover:text-patina-ink/70 group-focus-visible:text-patina-ink/70">
                  {column.meta}
                </span>
              </div>
              <h3 className="mt-5 text-2xl font-medium leading-tight tracking-[-0.02em] text-text-1 transition-colors group-hover:text-patina-ink group-focus-visible:text-patina-ink sm:text-[28px]">
                {column.title}
              </h3>
              <p className="mt-3 max-w-[48ch] text-[15px] leading-relaxed text-text-2 transition-colors group-hover:text-patina-ink/80 group-focus-visible:text-patina-ink/80">
                {column.body}
              </p>
              <span
                aria-hidden="true"
                className="mono mt-auto flex items-center gap-2 pt-6 text-[11px] uppercase tracking-[0.14em] text-text-3 transition-colors group-hover:text-patina-ink/80 group-focus-visible:text-patina-ink/80"
              >
                open the docs
                <IconArrowUpRight width={14} height={14} strokeWidth={2} className="transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
              </span>
            </a>
          ))}
        </div>

        <p className="mono mt-12 flex flex-wrap items-center justify-between gap-3 border-t border-line pt-5 text-[10.5px] uppercase tracking-[0.14em] text-text-3">
          <span>sso (oidc / saml) + scim 2.0 available · experimental · frozen for launch</span>
          <span className="flex items-center gap-2 normal-case tracking-normal">
            <span className="status-dot !h-1.5 !w-1.5" aria-hidden="true" />
            one process · zero external services
          </span>
        </p>
      </div>
    </section>
  );
}
