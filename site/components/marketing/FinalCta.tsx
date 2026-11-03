import { IconArrowUpRight, IconStar } from './icons';

export function FinalCta() {
  return (
    <section className="bg-signal text-[#121210]">
      <div className="container-site grid items-center gap-8 py-16 lg:grid-cols-12 lg:py-24">
        <div className="lg:col-span-8">
          <p className="mono text-[11px] uppercase tracking-[0.2em] text-[#121210]/70">exist.io — v0.6.0-beta</p>
          <h2 className="mt-4 text-5xl font-medium leading-[0.98] tracking-[-0.04em] sm:text-6xl lg:text-7xl">
            Run your own forge.
          </h2>
          <p className="mt-5 max-w-md text-lg leading-snug text-[#121210]/85">
            One binary. Your hardware. Your agents. No vendor account, no trial wall, no telemetry you didn’t ask for.
          </p>
        </div>
        <div className="flex flex-col items-start gap-3 lg:col-span-4 lg:items-end">
          <a href="#quickstart" className="btn btn-ink w-full justify-center sm:w-auto">
            Install exist.io
            <span className="btn-chip" aria-hidden="true"><IconArrowUpRight width={14} height={14} strokeWidth={2} /></span>
          </a>
          <a
            href="https://github.com/RavaniRoshan/exist.io"
            rel="noreferrer"
            className="inline-flex items-center gap-2 border border-[#121210]/40 px-5 py-3 text-[15px] font-medium text-[#121210] transition-colors hover:bg-[#121210] hover:text-signal"
          >
            <IconStar width={17} height={17} strokeWidth={2} />
            Star on GitHub
          </a>
          <p className="mono mt-1 text-[10px] uppercase tracking-[0.2em] text-[#121210]/60">
            apache-2.0 · no vendor account · no card
          </p>
        </div>
      </div>
    </section>
  );
}
