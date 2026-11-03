import Image from 'next/image';
import { MergeGate } from './MergeGate';
import { asset } from '@/lib/asset';

const LEASE_JSON = `{
  "agent": "claude-code",
  "repo": "acme/demo",
  "refs": ["refs/heads/agent/*"],
  "ttl": "15m",
  "signature": "HMAC-SHA256"
}`;

const REF_ROWS: Array<[ref: string, verdict: 'allowed' | 'denied']> = [
  ['refs/heads/agent/fix-lint', 'allowed'],
  ['refs/heads/agent/capture', 'allowed'],
  ['refs/heads/main', 'denied'],
  ['refs/tags/v0.6.0', 'denied'],
];

const LAYERS: Array<[name: string, detail: string, sub: string]> = [
  ['Clients & agents', 'HTTP · SSH · MCP', 'git, curl, mcp clients, ci'],
  ['existd', 'policy · transports · embedded ui', 'web ui on :8080 · ssh on :2222'],
  ['Local disk', 'sqlite wal · bare repositories', 'no external database · no host sshd'],
];

export function Product() {
  return (
    <section id="product" className="border-b border-line scroll-mt-20">
      <div className="container-site py-16 lg:py-24">
        <div className="grid gap-6 border-b border-line pb-8 lg:grid-cols-12 lg:items-end">
          <div className="lg:col-span-8">
            <p className="eyebrow">02 — under the hood</p>
            <h2 className="mt-4 text-4xl font-medium leading-[1.02] tracking-[-0.035em] text-ink sm:text-5xl">
              Policy at every step,<br />not just at the edge.
            </h2>
          </div>
          <p className="max-w-md text-[15px] leading-relaxed text-ink-2 lg:col-span-4 lg:justify-self-end">
            Each stage narrows what an agent can do next. The server enforces the
            loop — not convention, not vibes, not a codeowners file nobody reads.
          </p>
        </div>

        <div className="mt-8 grid grid-cols-12 gap-4 lg:gap-5">
          {/* Lease JSON */}
          <div className="col-span-12 flex flex-col lg:col-span-7">
            <div className="frame frame-strong flex h-full flex-col">
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line px-5 py-3.5">
                <span className="mono text-[11px] uppercase tracking-[0.18em] text-ink-2">capability lease</span>
                <span className="mono border border-line bg-paper-2 px-2 py-1 text-[10px] text-ink-2">HMAC-SHA256</span>
              </div>
              <pre className="overflow-x-auto px-5 py-4 text-[12px] leading-relaxed text-ink">{LEASE_JSON}</pre>
              <div className="mt-auto flex flex-wrap items-center justify-between gap-3 border-t border-line px-5 py-3.5">
                <span className="mono text-[11px] text-ink-3">ttl 15m · prefix-isolated · revocable</span>
                <a className="mono text-[11px] text-ember underline decoration-line underline-offset-4 hover:text-ember-2" href="https://github.com/RavaniRoshan/exist.io/blob/main/pkg/agentlease/lease.go" rel="noreferrer">pkg/agentlease ↗</a>
              </div>
            </div>
          </div>

          {/* Ref-prefix isolation */}
          <div className="col-span-12 lg:col-span-5">
            <div className="frame flex h-full flex-col">
              <div className="border-b border-line px-5 py-3.5">
                <span className="mono text-[11px] uppercase tracking-[0.18em] text-ink-2">ref-prefix isolation</span>
              </div>
              <ul className="flex flex-col divide-y divide-line">
                {REF_ROWS.map(([ref, verdict]) => (
                  <li key={ref} className="flex items-center gap-3 px-5 py-3">
                    <span
                      aria-hidden="true"
                      className={`h-2 w-2 flex-none rounded-full ${verdict === 'allowed' ? 'bg-pass' : 'bg-ember'}`}
                    />
                    <code className="mono min-w-0 flex-1 truncate text-[11.5px] text-ink">{ref}</code>
                    <span className={`mono flex-none text-[10px] uppercase tracking-[0.18em] ${verdict === 'allowed' ? 'text-pass' : 'text-ember'}`}>
                      {verdict}
                    </span>
                  </li>
                ))}
              </ul>
              <p className="mono mt-auto border-t border-line px-5 py-3.5 text-[10.5px] leading-relaxed text-ink-3">
                enforced by the pre-receive hook + ssh policy — not by the client
              </p>
            </div>
          </div>

          {/* Interactive merge gate */}
          <div className="col-span-12 lg:col-span-7">
            <MergeGate />
          </div>

          {/* One daemon, three layers */}
          <div className="col-span-12 lg:col-span-5">
            <div className="frame flex h-full flex-col">
              <div className="border-b border-line px-5 py-3.5">
                <span className="mono text-[11px] uppercase tracking-[0.18em] text-ink-2">one daemon, three layers</span>
              </div>
              <ol className="flex flex-col px-5 py-4">
                {LAYERS.map(([name, detail, sub], i) => (
                  <li key={name} className="relative border-l border-line pl-5 pb-5 last:pb-0">
                    <span aria-hidden="true" className="absolute -left-[5px] top-1.5 h-[9px] w-[9px] rounded-full border border-line-strong bg-paper-3" />
                    <p className="text-[16px] font-medium tracking-tight text-ink">{name}</p>
                    <p className="mono mt-1 text-[11px] leading-relaxed text-ink-2">{detail}</p>
                    <p className="mono mt-0.5 text-[10px] leading-relaxed text-ink-3">{sub}</p>
                    {i < LAYERS.length - 1 && <span aria-hidden="true" className="mono mt-3 block text-[10px] text-ink-3">↓</span>}
                  </li>
                ))}
              </ol>
              <p className="mono mt-auto border-t border-line px-5 py-3.5 text-[10.5px] leading-relaxed text-ink-3">
                CGO_ENABLED=0 · modernc.org/sqlite · go-git/v6
              </p>
            </div>
          </div>

          {/* Real UI plates */}
          <figure className="col-span-12 md:col-span-7">
            <div className="plate">
              <div className="plate-inner aspect-[16/10]">
                <Image
                  src={asset('/product/forge-home.png')}
                  alt="exist.io web UI — repository overview with agent activity"
                  fill
                  sizes="(min-width: 768px) 58vw, 100vw"
                  className="object-cover object-top"
                />
              </div>
            </div>
            <figcaption className="mono mt-3 text-[11px] text-ink-3">fig. 2 — web ui · repository overview</figcaption>
          </figure>
          <figure className="col-span-12 md:col-span-5">
            <div className="plate">
              <div className="plate-inner aspect-[16/10]">
                <Image
                  src={asset('/product/forge-agents.png')}
                  alt="exist.io agent activity — leases, scopes and merge eligibility"
                  fill
                  sizes="(min-width: 768px) 42vw, 100vw"
                  className="object-cover object-top"
                />
              </div>
            </div>
            <figcaption className="mono mt-3 text-[11px] text-ink-3">fig. 3 — agent activity · lease scopes</figcaption>
          </figure>
        </div>
      </div>
    </section>
  );
}
