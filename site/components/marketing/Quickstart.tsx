import { QuickstartTerminal } from './QuickstartTerminal';
import { IconArrowUpRight } from './icons';
import { asset } from '@/lib/asset';

export function Quickstart() {
  return (
    <section id="quickstart" className="relative border-b border-line scroll-mt-20">
      <div className="container-site grid gap-12 py-16 lg:grid-cols-12 lg:py-24">
        <div className="lg:col-span-5">
          <p className="eyebrow">04 — start here</p>
          <h2 className="mt-4 text-4xl font-light leading-[1.02] tracking-[-0.035em] text-text-1 sm:text-5xl">
            A forge in <span className="text-patina">one command.</span>
          </h2>
          <p className="mt-5 max-w-md text-[15px] leading-relaxed text-text-2">
            A verified tarball, your own forge, and the admin key printed exactly once.
            The daemon embeds the database, the web UI, and both Git transports.
          </p>

          <div className="mt-8 border-y border-line">
            {[
              ['01', 'install', 'verified release tarball'],
              ['02', 'bootstrap', 'namespace · repo · admin key'],
              ['03', 'run', 'ui · rest · smart-http · mcp'],
            ].map(([index, label, note]) => (
              <div key={label} className="group flex items-center gap-4 border-b border-line py-3 last:border-b-0">
                <span className="mono w-6 flex-none text-[11px] font-semibold text-patina">{index}</span>
                <span className="text-[14px] font-medium text-text-1">{label}</span>
                <span className="mono ml-auto text-right text-[10px] uppercase tracking-[0.1em] text-text-3">{note}</span>
              </div>
            ))}
          </div>

          <div className="mt-6 flex flex-wrap items-center gap-5">
            <a href={asset('/docs/getting-started/')} className="group flex items-center gap-2 text-[15px] font-medium text-patina hover:text-patina-hover">
              Read the docs
              <IconArrowUpRight width={16} height={16} strokeWidth={2} className="transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
            </a>
            <a href="https://github.com/RavaniRoshan/exist.io/releases/tag/v0.6.0-beta" className="group flex items-center gap-2 text-[15px] font-medium text-text-2 hover:text-text-1" rel="noreferrer">
              v0.6.0-beta release notes
              <IconArrowUpRight width={16} height={16} strokeWidth={2} className="transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
            </a>
          </div>

          <p className="mono mt-6 max-w-sm text-[10.5px] leading-relaxed text-text-3">
            bootstrap prints the admin api key once — store it, then open http://localhost:8080.
            web ui, rest, smart-http and mcp all share the one port. ssh listens on 127.0.0.1:2222 by default.
          </p>
        </div>

        <div className="min-w-0 lg:col-span-7">
          <QuickstartTerminal />
          <p className="mono mt-3 flex flex-wrap items-center justify-between gap-2 text-[10.5px] text-text-3">
            <span>terminal replay · commands are copyable</span>
            <span className="text-text-3">install.sh verifies sha-256 before unpacking</span>
          </p>
        </div>
      </div>
    </section>
  );
}
