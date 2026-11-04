import { MergeGate } from './MergeGate';
import { ProductPreview } from './ProductPreview';

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
    <section id="product" className="relative border-b border-line scroll-mt-20">
      <span aria-hidden="true" className="crosshair left-0 top-0 hidden -translate-x-1/2 -translate-y-1/2 lg:block" />
      <span aria-hidden="true" className="crosshair right-0 top-0 hidden translate-x-1/2 -translate-y-1/2 lg:block" />

      <div className="container-site py-16 lg:py-24">
        <div className="grid gap-6 border-b border-line pb-8 lg:grid-cols-12 lg:items-end">
          <div className="lg:col-span-8">
            <p className="eyebrow">02 — under the hood</p>
            <h2 className="mt-4 text-4xl font-light leading-[1.02] tracking-[-0.035em] text-text-1 sm:text-5xl">
              <span className="text-patina">Policy</span> at every step,<br />not just at the edge.
            </h2>
          </div>
          <p className="max-w-md text-[15px] leading-relaxed text-text-2 lg:col-span-4 lg:justify-self-end">
            Each stage narrows what an agent can do next. The server enforces the
            loop — not convention, not vibes, not a codeowners file nobody reads.
          </p>
        </div>

        <div className="mt-8">
          <ProductPreview />
        </div>

        <div className="mt-5 grid grid-cols-12 gap-4 lg:gap-5">
          {/* Lease JSON */}
          <div className="col-span-12 flex flex-col lg:col-span-7">
            <div className="frame frame-strong group flex h-full flex-col transition-colors hover:border-patina/50">
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line px-5 py-3.5">
                <span className="mono text-[11px] uppercase tracking-[0.18em] text-text-2">capability lease</span>
                <span className="mono rounded-sm border border-line bg-raised px-2 py-1 text-[10px] text-text-2">HMAC-SHA256</span>
              </div>
              <pre className="overflow-x-auto px-5 py-4 text-[12px] leading-relaxed text-text-1">{LEASE_JSON}</pre>
              <div className="mt-auto flex flex-wrap items-center justify-between gap-3 border-t border-line px-5 py-3.5">
                <span className="mono text-[11px] text-text-3">ttl 15m · prefix-isolated · revocable</span>
                <a className="mono text-[11px] text-patina underline decoration-line underline-offset-4 hover:text-patina-hover" href="https://github.com/RavaniRoshan/exist.io/blob/main/pkg/agentlease/lease.go" rel="noreferrer">pkg/agentlease ↗</a>
              </div>
            </div>
          </div>

          {/* Ref-prefix isolation */}
          <div className="col-span-12 lg:col-span-5">
            <div className="frame flex h-full flex-col">
              <div className="border-b border-line px-5 py-3.5">
                <span className="mono text-[11px] uppercase tracking-[0.18em] text-text-2">ref-prefix isolation</span>
              </div>
              <ul className="flex flex-col divide-y divide-line">
                {REF_ROWS.map(([ref, verdict]) => (
                  <li key={ref} className="group flex items-center gap-3 px-5 py-3 transition-colors hover:bg-raised">
                    <span
                      aria-hidden="true"
                      className={`h-2 w-2 flex-none rounded-full ${verdict === 'allowed' ? 'bg-patina' : 'bg-brass'}`}
                    />
                    <code className="mono min-w-0 flex-1 truncate text-[11.5px] text-text-1">{ref}</code>
                    <span className={`mono flex-none text-[10px] uppercase tracking-[0.18em] ${verdict === 'allowed' ? 'text-patina' : 'text-brass'}`}>
                      {verdict}
                    </span>
                  </li>
                ))}
              </ul>
              <p className="mono mt-auto border-t border-line px-5 py-3.5 text-[10.5px] leading-relaxed text-text-3">
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
                <span className="mono text-[11px] uppercase tracking-[0.18em] text-text-2">one daemon, three layers</span>
              </div>
              <ol className="flex flex-col px-5 py-4">
                {LAYERS.map(([name, detail, sub], i) => (
                  <li key={name} className="group relative border-l border-line pb-5 pl-5 last:pb-0">
                    <span aria-hidden="true" className="absolute -left-[5px] top-1.5 h-[9px] w-[9px] rounded-full border border-line-strong bg-slab transition-colors group-hover:border-patina group-hover:bg-patina" />
                    <p className="text-[16px] font-medium tracking-tight text-text-1">{name}</p>
                    <p className="mono mt-1 text-[11px] leading-relaxed text-text-2">{detail}</p>
                    <p className="mono mt-0.5 text-[10px] leading-relaxed text-text-3">{sub}</p>
                    {i < LAYERS.length - 1 && <span aria-hidden="true" className="mono mt-3 block text-[10px] text-text-3">↓</span>}
                  </li>
                ))}
              </ol>
              <p className="mono mt-auto border-t border-line px-5 py-3.5 text-[10.5px] leading-relaxed text-text-3">
                CGO_ENABLED=0 · modernc.org/sqlite · go-git/v6
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
