import { LiveClock } from './LiveClock';
import { Logomark } from './icons';
import { asset } from '@/lib/asset';

const COLUMNS: Array<[title: string, links: Array<[label: string, href: string]>]> = [
  [
    'project',
    [
      ['GitHub', 'https://github.com/RavaniRoshan/exist.io'],
      ['Releases', 'https://github.com/RavaniRoshan/exist.io/releases'],
      ['Roadmap', 'https://github.com/RavaniRoshan/exist.io/blob/main/ROADMAP.md'],
      ['License', 'https://github.com/RavaniRoshan/exist.io/blob/main/LICENSE'],
    ],
  ],
  [
    'docs',
    [
      ['Documentation', asset('/docs/')],
      ['Getting started', asset('/docs/getting-started/')],
      ['MCP tools', asset('/docs/reference/mcp-tools/')],
      ['Claims ledger', 'https://github.com/RavaniRoshan/exist.io/blob/main/CLAIMS.md'],
    ],
  ],
  [
    'platform',
    [
      ['Capability leases', 'https://github.com/RavaniRoshan/exist.io/blob/main/pkg/agentlease/lease.go'],
      ['MCP server', 'https://github.com/RavaniRoshan/exist.io/tree/main/internal/mcp'],
      ['Threat model', 'https://github.com/RavaniRoshan/exist.io/blob/main/THREAT_MODEL.md'],
      ['Security policy', 'https://github.com/RavaniRoshan/exist.io/blob/main/SECURITY.md'],
    ],
  ],
  [
    'community',
    [
      ['Issues', 'https://github.com/RavaniRoshan/exist.io/issues'],
      ['Pull requests', 'https://github.com/RavaniRoshan/exist.io/pulls'],
      ['Contributing', 'https://github.com/RavaniRoshan/exist.io/blob/main/CONTRIBUTING.md'],
    ],
  ],
];

export function StatusFooter() {
  return (
    <footer className="bg-canvas">
      <div className="container-site grid gap-10 py-12 md:grid-cols-12">
        <div className="md:col-span-3">
          <a href={asset('/')} className="flex items-center gap-2.5" aria-label="exist.io — home">
            <Logomark size={20} />
            <span className="text-[19px] font-semibold tracking-[-0.03em] text-text-1">exist.io</span>
          </a>
          <p className="mt-4 max-w-[28ch] text-sm leading-relaxed text-text-2">
            The self-hosted Git forge for people and coding agents.
          </p>
          <p className="mono mt-5 flex items-center gap-2 text-[10px] uppercase tracking-[0.16em] text-text-3">
            <span className="status-dot !h-1.5 !w-1.5" aria-hidden="true" />
            system normal
          </p>
        </div>
        {COLUMNS.map(([title, links]) => (
          <nav key={title} className="md:col-span-2" aria-label={title}>
            <p className="mono text-[10px] uppercase tracking-[0.22em] text-text-3">{title}</p>
            <ul className="mt-3 flex flex-col gap-2">
              {links.map(([label, href]) => (
                <li key={label}>
                  <a
                    href={href}
                    className="group inline-flex items-center gap-1 text-[13.5px] leading-snug text-text-2 transition-colors hover:text-patina"
                    rel={href.startsWith('http') ? 'noreferrer' : undefined}
                  >
                    {label}
                    <span aria-hidden="true" className="opacity-0 transition-opacity group-hover:opacity-100">↗</span>
                  </a>
                </li>
              ))}
            </ul>
          </nav>
        ))}
      </div>

      <div className="statusbar">
        <div className="container-site flex h-11 items-center justify-between gap-6 overflow-x-auto whitespace-nowrap">
          <span className="flex items-center gap-2"><span aria-hidden="true" className="pixel" />v0.6.0-beta · pure go · zero cgo · apache-2.0</span>
          <span className="hidden items-center gap-2 md:flex" aria-hidden="true">:8080 ui · rest · smart-http · mcp — :2222 ssh</span>
          <span className="flex items-center gap-4">
            <LiveClock />
            <span>© 2026 exist.io contributors</span>
          </span>
        </div>
      </div>
    </footer>
  );
}
