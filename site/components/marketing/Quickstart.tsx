import { CommandChip } from './CommandChip';
import { IconArrowUpRight } from './icons';

const STEPS = [
  { label: 'install', command: 'curl -fsSL https://raw.githubusercontent.com/RavaniRoshan/exist.io/main/scripts/install.sh | bash' },
  { label: 'bootstrap', command: 'existd bootstrap --email admin@example.com --namespace acme --data-dir ./data' },
  { label: 'run', command: 'existd' },
] as const;

export function Quickstart() {
  return (
    <section id="quickstart" className="border-b border-line scroll-mt-20">
      <div className="container-site grid gap-12 py-16 lg:grid-cols-12 lg:py-24">
        <div className="lg:col-span-5">
          <p className="eyebrow">04 — start here</p>
          <h2 className="mt-4 text-4xl font-medium leading-[1.02] tracking-[-0.035em] text-ink sm:text-5xl">
            A forge in one command.
          </h2>
          <p className="mt-5 max-w-md text-[15px] leading-relaxed text-ink-2">
            A verified tarball, your own forge, and the admin key printed exactly once.
            The daemon embeds the database, the web UI, and both Git transports.
          </p>
          <div className="mt-6 flex flex-wrap items-center gap-5">
            <a href="/docs/getting-started/" className="group flex items-center gap-2 text-[15px] font-medium text-ember hover:text-ember-2">
              Read the docs
              <IconArrowUpRight width={16} height={16} strokeWidth={2} className="transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
            </a>
            <a href="https://github.com/RavaniRoshan/exist.io/releases/tag/v0.6.0-beta" className="group flex items-center gap-2 text-[15px] font-medium text-ink-2 hover:text-ink" rel="noreferrer">
              v0.6.0-beta release notes
              <IconArrowUpRight width={16} height={16} strokeWidth={2} className="transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
            </a>
          </div>
        </div>

        <ol className="lg:col-span-7">
          {STEPS.map((step, i) => (
            <li key={step.label} className="flex items-start gap-5 border-t border-line py-5 last:border-b">
              <span className="mono mt-1 w-6 flex-none text-[12px] font-semibold text-ember">0{i + 1}</span>
              <div className="min-w-0">
                <p className="mono text-[10px] uppercase tracking-[0.2em] text-ink-3">{step.label}</p>
                <div className="mt-2"><CommandChip command={step.command} /></div>
              </div>
            </li>
          ))}
          <li className="border-t border-line pt-5">
            <p className="mono text-[10.5px] leading-relaxed text-ink-3">
              bootstrap prints the admin api key once — store it, then open http://localhost:8080.
              web ui, rest, smart-http and mcp all share the one port. ssh listens on 127.0.0.1:2222 by default.
            </p>
          </li>
        </ol>
      </div>
    </section>
  );
}
