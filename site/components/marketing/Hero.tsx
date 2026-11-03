import Image from 'next/image';
import { CommandChip } from './CommandChip';
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
      {/* blueprint dot field, right side, fading into paper */}
      <div
        aria-hidden="true"
        className="halftone pointer-events-none absolute inset-y-0 right-0 w-3/5 opacity-70 [mask-image:linear-gradient(to_right,transparent,black_45%)]"
      />
      {/* registration crosshairs */}
      <span aria-hidden="true" className="crosshair left-3 top-3 lg:left-6 lg:top-6" />
      <span aria-hidden="true" className="crosshair right-3 top-3 lg:right-6 lg:top-6" />

      <div className="container-site relative grid gap-12 pb-10 pt-14 lg:grid-cols-12 lg:items-center lg:gap-10 lg:pb-14 lg:pt-20">
        <div className="lg:col-span-7">
          <Reveal>
            <p className="eyebrow">self-hosted · apache-2.0 · pure go</p>
          </Reveal>

          <Reveal>
            <h1 className="mt-6 text-[clamp(3.1rem,7.4vw,5.3rem)] font-medium leading-[0.98] tracking-[-0.045em] text-ink">
              The self-hosted forge for <span className="marker">coding&nbsp;agents.</span>
            </h1>
          </Reveal>

          <Reveal>
            <p className="mt-6 max-w-[52ch] text-lg leading-relaxed text-ink-2">
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
              <a href="/docs/" className="btn">
                Read the docs
                <span className="btn-chip" aria-hidden="true"><IconArrowUpRight width={14} height={14} strokeWidth={2} /></span>
              </a>
            </div>
          </Reveal>

          <Reveal>
            <div className="mt-6">
              <CommandChip command={INSTALL} />
              <p className="mono mt-3 text-[11px] leading-relaxed text-ink-3">
                v0.6.0-beta · apache-2.0 · zero cgo · no external database
              </p>
            </div>
          </Reveal>
        </div>

        <figure className="lg:col-span-5">
          <Reveal>
            <div className="plate">
              <div className="plate-inner">
                <Image
                  src={asset('/product/forge-pr.png')}
                  alt="exist.io pull request view — checks passing, capability lease verified, merge gate ready"
                  width={1440}
                  height={900}
                  priority
                  className="h-full w-full object-cover object-top"
                />
              </div>
            </div>
            <figcaption className="mt-3 flex items-center justify-between gap-4">
              <span className="mono text-[11px] text-ink-3">fig. 1 — pr view · lease verified · checks 3/3</span>
              <span className="mono border border-signal-deep bg-signal px-2 py-1 text-[9.5px] font-semibold uppercase tracking-[0.18em] text-[#121210]">real ui</span>
            </figcaption>
          </Reveal>
        </figure>
      </div>

      <div className="ruler" aria-hidden="true" />
      <div className="border-b border-line bg-paper-3">
        <dl className="container-site grid grid-cols-2 divide-x divide-line md:grid-cols-4">
          {STATS.map(([value, label]) => (
            <div key={label} className="flex flex-col gap-1 px-4 py-4 first:pl-0 sm:px-6">
              <dt className="sr-only">{label}</dt>
              <dd className="mono text-lg font-semibold leading-none text-ink sm:text-xl">{value}</dd>
              <dd className="mono text-[9.5px] uppercase leading-snug tracking-[0.14em] text-ink-3">{label}</dd>
            </div>
          ))}
        </dl>
      </div>
    </section>
  );
}
