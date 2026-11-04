const LINKS = [
  { label: 'Claims ledger', note: 'every claim → a test or a file', href: 'https://github.com/RavaniRoshan/exist.io/blob/main/CLAIMS.md' },
  { label: 'Threat model', note: 'trust boundaries & attack surface', href: 'https://github.com/RavaniRoshan/exist.io/blob/main/THREAT_MODEL.md' },
  { label: 'Security policy', note: 'reporting & disclosure', href: 'https://github.com/RavaniRoshan/exist.io/blob/main/SECURITY.md' },
  { label: 'Release notes', note: 'v0.6.0-beta', href: 'https://github.com/RavaniRoshan/exist.io/releases/tag/v0.6.0-beta' },
] as const;

export function Evidence() {
  return (
    <section id="evidence" className="relative border-b border-line">
      <div className="container-site grid gap-10 py-16 lg:grid-cols-12 lg:py-20">
        <div className="lg:col-span-6">
          <p className="eyebrow">05 — proof</p>
          <h2 className="mt-4 text-4xl font-light leading-[1.02] tracking-[-0.035em] text-text-1 sm:text-5xl">
            <span className="text-patina">Evidence,</span><br />not adjectives.
          </h2>
          <p className="mt-5 max-w-md text-[15px] leading-relaxed text-text-2">
            Every capability on this page is traced to a test or a file in the repo.
            If it isn’t verified, it’s labeled experimental — or it isn’t here.
          </p>
          <p className="mono mt-8 inline-flex items-center gap-2 border border-line bg-slab px-3 py-1.5 text-[10px] uppercase tracking-[0.14em] text-text-2">
            <span className="status-dot !h-1.5 !w-1.5" aria-hidden="true" />
            gate: verified on every release
          </p>
        </div>

        <div className="lg:col-span-6">
          <ul className="border-y border-line">
            {LINKS.map((link, index) => (
              <li key={link.label}>
                <a
                  href={link.href}
                  rel="noreferrer"
                  className="group grid grid-cols-[32px_minmax(0,1fr)_auto] items-baseline gap-4 border-t border-line px-3 py-5 transition-colors first:border-t-0 hover:bg-patina hover:text-patina-ink focus-visible:bg-patina focus-visible:text-patina-ink"
                >
                  <span className="mono text-[11px] text-text-3 transition-colors group-hover:text-patina-ink/70 group-focus-visible:text-patina-ink/70">
                    E{index + 1}
                  </span>
                  <span className="min-w-0">
                    <span className="block text-lg font-medium tracking-tight text-text-1 transition-colors group-hover:text-patina-ink group-focus-visible:text-patina-ink">{link.label}</span>
                    <span className="mono mt-1 block text-[10.5px] uppercase tracking-[0.14em] text-text-3 transition-colors group-hover:text-patina-ink/70 group-focus-visible:text-patina-ink/70">{link.note}</span>
                  </span>
                  <span
                    aria-hidden="true"
                    className="mono flex-none text-xl text-text-3 transition-all group-hover:-translate-y-0.5 group-hover:translate-x-0.5 group-hover:text-patina-ink group-focus-visible:text-patina-ink"
                  >
                    ↗
                  </span>
                </a>
              </li>
            ))}
          </ul>
          <p className="mono mt-4 text-[10.5px] leading-relaxed text-text-3">
            claims.md — verified / partial / experimental, always current · updated per release
          </p>
        </div>
      </div>
    </section>
  );
}
