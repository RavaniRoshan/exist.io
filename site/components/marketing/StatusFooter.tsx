import { Logomark } from './icons';

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
      ['Documentation', '/docs/'],
      ['Getting started', '/docs/getting-started/'],
      ['MCP tools', '/docs/reference/mcp-tools/'],
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
    <footer className="bg-paper">
      <div className="container-site grid gap-10 py-12 md:grid-cols-12">
        <div className="md:col-span-3">
          <a href="/" className="flex items-center gap-2.5" aria-label="exist.io — home">
            <Logomark size={20} />
            <span className="text-[19px] font-semibold tracking-[-0.03em] text-ink">exist.io</span>
          </a>
          <p className="mt-4 max-w-[28ch] text-sm leading-relaxed text-ink-2">
            The self-hosted Git forge for people and coding agents.
          </p>
        </div>
        {COLUMNS.map(([title, links]) => (
          <nav key={title} className="md:col-span-2" aria-label={title}>
            <p className="mono text-[10px] uppercase tracking-[0.22em] text-ink-3">{title}</p>
            <ul className="mt-3 flex flex-col gap-2">
              {links.map(([label, href]) => (
                <li key={label}>
                  <a href={href} className="text-[13.5px] leading-snug text-ink-2 transition-colors hover:text-ember" rel={href.startsWith('http') ? 'noreferrer' : undefined}>
                    {label}
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
          <span className="hidden md:inline">:8080 ui · rest · smart-http · mcp — :2222 ssh</span>
          <span>© 2026 exist.io contributors</span>
        </div>
      </div>
    </footer>
  );
}
