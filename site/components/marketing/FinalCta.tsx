import { IconArrowUpRight, IconStar } from './icons';

const TERMS = ['lease', 'push', 'review', 'merge', 'audit', 'revoke'];

export function FinalCta() {
  return (
    <section className="relative overflow-hidden bg-patina text-patina-ink">
      <div aria-hidden="true" className="hatch pointer-events-none absolute inset-0 opacity-[0.12]" />
      <div aria-hidden="true" className="dither pointer-events-none absolute inset-y-0 right-0 w-1/3 opacity-[0.10] [mask-image:linear-gradient(to_right,transparent,black)]" />

      <div className="marquee relative overflow-hidden border-b border-patina-ink/20">
        <div className="marquee-track">
          {[0, 1].map((copy) => (
            <div key={copy} className="flex shrink-0 items-center">
              {TERMS.map((term) => (
                <span key={`${copy}-${term}`} className="mono flex items-center gap-4 px-4 py-2 text-[10px] uppercase tracking-[0.22em] text-patina-ink/70">
                  {term}
                  <span aria-hidden="true" className="h-1 w-1 rounded-full bg-patina-ink/40" />
                </span>
              ))}
            </div>
          ))}
        </div>
      </div>

      <div className="container-site relative grid items-center gap-8 py-16 lg:grid-cols-12 lg:py-24">
        <div className="lg:col-span-8">
          <p className="mono text-[11px] uppercase tracking-[0.2em] text-patina-ink/70">exist.io — v0.6.0-beta</p>
          <h2 className="mt-4 text-5xl font-medium leading-[0.98] tracking-[-0.04em] sm:text-6xl lg:text-7xl">
            Run your own forge.
          </h2>
          <p className="mt-5 max-w-md text-lg leading-snug text-patina-ink/85">
            One binary. Your hardware. Your agents. No vendor account, no trial wall, no telemetry you didn’t ask for.
          </p>
        </div>
        <div className="flex flex-col items-start gap-3 lg:col-span-4 lg:items-end">
          <a href="#quickstart" className="btn w-full justify-center border-patina-ink/30 bg-canvas text-text-1 hover:bg-text-1 hover:text-canvas sm:w-auto">
            Install exist.io
            <span className="btn-chip" aria-hidden="true"><IconArrowUpRight width={14} height={14} strokeWidth={2} /></span>
          </a>
          <a
            href="https://github.com/RavaniRoshan/exist.io"
            rel="noreferrer"
            className="inline-flex items-center gap-2 rounded-md border border-patina-ink/40 px-5 py-3 text-[15px] font-medium text-patina-ink transition-colors hover:bg-patina-ink hover:text-patina"
          >
            <IconStar width={17} height={17} strokeWidth={2} />
            Star on GitHub
          </a>
          <p className="mono mt-1 text-[10px] uppercase tracking-[0.2em] text-patina-ink/60">
            apache-2.0 · no vendor account · no card
          </p>
        </div>
      </div>
    </section>
  );
}
