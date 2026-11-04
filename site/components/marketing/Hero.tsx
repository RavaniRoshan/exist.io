import { CommandChip } from './CommandChip';
import { ForgeDemo } from './ForgeDemo';
import { Reveal } from './Reveal';
import { IconArrowUpRight } from './icons';
import { asset } from '@/lib/asset';

const INSTALL = 'curl -fsSL https://raw.githubusercontent.com/RavaniRoshan/exist.io/main/scripts/install.sh | bash';

const STATS: Array<[string, string]> = [
  ['1', 'binary to deploy'],
  ['0', 'external services'],
  [':8080', 'ui · rest · smart-http · mcp'],
  [':2222', 'embedded ssh'],
];

export function Hero() {
  return (
    <section className="relative overflow-hidden border-b border-line">
      <div
        aria-hidden="true"
        className="grid-field pointer-events-none absolute inset-0 opacity-70 [mask-image:linear-gradient(to_bottom,black,transparent_88%)]"
      />
      <span aria-hidden="true" className="crosshair left-3 top-3 lg:left-6 lg:top-6" />
      <span aria-hidden="true" className="crosshair right-3 top-3 lg:right-6 lg:top-6" />

      <div className="container-site relative grid gap-12 pb-12 pt-14 lg:grid-cols-12 lg:items-center lg:gap-10 lg:pb-16 lg:pt-20">
        <div className="min-w-0 lg:col-span-7">
          <Reveal>
            <p className="eyebrow">self-hosted · apache-2.0 · pure go</p>
          </Reveal>

          <Reveal>
            <h1 className="mt-6 text-[clamp(3.05rem,7.2vw,5.15rem)] font-light leading-[0.98] tracking-[-0.045em] text-text-1">
              The self-hosted forge for <span className="marker text-patina">coding&nbsp;agents.</span>
            </h1>
          </Reveal>

          <Reveal>
            <p className="mt-6 max-w-[52ch] text-lg leading-relaxed text-text-2">
              exist.io gives Claude Code, Codex, Cline, and Cursor a real Git home —
              scoped leases, pull requests, and policy-gated merges — on your hardware,
              in one binary.
            </p>
          </Reveal>

          <Reveal>
            <div className="mt-8 flex flex-wrap items-center gap-3">
              <a href="#quickstart" className="btn btn-signal">
                Install exist.io
                <span className="btn-chip" aria-hidden="true"><IconArrowUpRight width={14} height={14} strokeWidth={2} /></span>
              </a>
              <a href={asset('/docs/')} className="btn">
                Read the docs
                <span className="btn-chip" aria-hidden="true"><IconArrowUpRight width={14} height={14} strokeWidth={2} /></span>
              </a>
            </div>
          </Reveal>

          <Reveal>
            <div className="mt-6">
              <CommandChip command={INSTALL} />
              <p className="mono mt-3 text-[11px] leading-relaxed text-text-3">
                v0.6.0-beta · apache-2.0 · zero cgo · no external database
              </p>
            </div>
          </Reveal>
        </div>

        <figure className="min-w-0 lg:col-span-5">
          <Reveal>
            <ForgeDemo />
            <figcaption className="mt-3 flex items-center justify-between gap-4">
              <span className="mono text-[10.5px] text-text-3">fig. 1 — live merge sequence · checks 3/3</span>
              <span className="mono rounded-sm border border-patina/40 bg-patina-soft px-2 py-1 text-[9.5px] font-semibold uppercase tracking-[0.16em] text-patina">
                interactive
              </span>
            </figcaption>
          </Reveal>
        </figure>
      </div>

      <div className="ruler" aria-hidden="true" />
      <div className="relative border-b border-line bg-slab">
        <dl className="container-site grid grid-cols-2 divide-x divide-line md:grid-cols-4">
          {STATS.map(([value, label]) => (
            <div key={label} className="flex flex-col gap-1 px-4 py-4 first:pl-0 sm:px-6">
              <dt className="sr-only">{label}</dt>
              <dd className="mono text-lg font-semibold leading-none text-text-1 sm:text-xl">{value}</dd>
              <dd className="mono text-[9.5px] uppercase leading-snug tracking-[0.14em] text-text-3">{label}</dd>
            </div>
          ))}
        </dl>
      </div>
    </section>
  );
}
