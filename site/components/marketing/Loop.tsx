import {
  IconLease,
  IconClone,
  IconPush,
  IconPullRequest,
  IconEye,
  IconMerge,
  IconArrowUpRight,
} from './icons';

const STEPS = [
  { id: '01', icon: IconLease, label: 'lease', title: 'Request a lease', detail: 'A short-lived HMAC capability token, scoped to a ref prefix.', meta: 'POST /api/v1/leases → 15m' },
  { id: '02', icon: IconClone, label: 'clone', title: 'Clone the repo', detail: 'The lease authenticates Git Smart-HTTP or SSH. Nothing else.', meta: 'git clone …/acme/demo' },
  { id: '03', icon: IconPush, label: 'push', title: 'Push a branch', detail: 'The pre-receive hook enforces the ref prefix and scans for secrets.', meta: 'git push origin agent/capture' },
  { id: '04', icon: IconPullRequest, label: 'open pr', title: 'Open a pull request', detail: 'The agent opens the PR through MCP — not a personal token.', meta: 'mcp create_pull_request' },
  { id: '05', icon: IconEye, label: 'review', title: 'Review the diff', detail: 'Checks, rulesets and the reviewer agent run before merge.', meta: 'checks: secret-scan ✓' },
  { id: '06', icon: IconMerge, label: 'merge', title: 'Merge through policy', detail: 'The merge gate passes, the lease expires, the audit log records it.', meta: 'merge gate: 3/3 ✓' },
] as const;

export function Loop() {
  return (
    <section id="agents" className="border-b border-line scroll-mt-20">
      <div className="container-site py-16 lg:py-24">
        <div className="flex flex-wrap items-end justify-between gap-6 border-b border-line pb-8">
          <div>
            <p className="eyebrow">01 — the agent loop</p>
            <h2 className="mt-4 text-4xl font-medium leading-[1.02] tracking-[-0.035em] text-ink sm:text-5xl">
              From lease to merge.
            </h2>
          </div>
          <a href="/docs/tutorials/first-agent/" className="group flex items-center gap-2 text-[15px] font-medium text-ember hover:text-ember-2">
            Walk through it
            <IconArrowUpRight width={16} height={16} strokeWidth={2} className="transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
          </a>
        </div>

        <ol className="grid grid-cols-1 gap-px border border-line bg-line sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
          {STEPS.map((step) => (
            <li key={step.id} className="flex flex-col bg-paper-3 p-5">
              <div className="flex items-center justify-between">
                <span className="mono text-[11px] font-semibold text-ember">{step.id}</span>
                <step.icon width={20} height={20} className="text-ink-3" />
              </div>
              <p className="mono mt-4 text-[10px] uppercase tracking-[0.2em] text-ink-3">{step.label}</p>
              <h3 className="mt-1.5 text-[17px] font-medium leading-snug tracking-tight text-ink">{step.title}</h3>
              <p className="mt-2 text-[13px] leading-relaxed text-ink-2">{step.detail}</p>
              <p className="mono mt-auto border-t border-line pt-3 text-[10px] leading-relaxed text-ink-3 [overflow-wrap:anywhere]">{step.meta}</p>
            </li>
          ))}
        </ol>

        <p className="mono mt-4 text-[11px] text-ink-3">
          Six steps. One short-lived credential. Zero standing privileges.
        </p>
      </div>
    </section>
  );
}
