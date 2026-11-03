const LINKS = [
  { label: 'Claims ledger', note: 'every claim → a test or a file', href: 'https://github.com/RavaniRoshan/exist.io/blob/main/CLAIMS.md' },
  { label: 'Threat model', note: 'trust boundaries & attack surface', href: 'https://github.com/RavaniRoshan/exist.io/blob/main/THREAT_MODEL.md' },
  { label: 'Security policy', note: 'reporting & disclosure', href: 'https://github.com/RavaniRoshan/exist.io/blob/main/SECURITY.md' },
  { label: 'Release notes', note: 'v0.6.0-beta', href: 'https://github.com/RavaniRoshan/exist.io/releases/tag/v0.6.0-beta' },
] as const;

export function Evidence() {
  return (
    <section id="evidence" className="border-b border-line">
      <div className="container-site grid gap-10 py-16 lg:grid-cols-12 lg:py-20">
        <div className="lg:col-span-6">
          <p className="eyebrow">05 — proof</p>
          <h2 className="mt-4 text-4xl font-medium leading-[1.02] tracking-[-0.035em] text-ink sm:text-5xl">
            Evidence,<br />not adjectives.
          </h2>
          <p className="mt-5 max-w-md text-[15px] leading-relaxed text-ink-2">
            Every capability on this page is traced to a test or a file in the repo.
            If it isn’t verified, it’s labeled experimental — or it isn’t here.
          </p>
        </div>
        <div className="lg:col-span-6">
          <ul className="border-b border-line">
            {LINKS.map((link) => (
              <li key={link.label}>
                <a
                  href={link.href}
                  rel="noreferrer"
                  className="group flex items-baseline justify-between gap-6 border-t border-line py-4 transition-colors hover:bg-paper-3"
                >
                  <span className="min-w-0">
                    <span className="block text-lg font-medium tracking-tight text-ink group-hover:text-ember">{link.label}</span>
                    <span className="mono mt-1 block text-[10.5px] uppercase tracking-[0.14em] text-ink-3">{link.note}</span>
                  </span>
                  <span aria-hidden="true" className="mono flex-none text-xl text-ink-3 transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5 group-hover:text-ember">↗</span>
                </a>
              </li>
            ))}
          </ul>
          <p className="mono mt-4 text-[10.5px] leading-relaxed text-ink-3">
            claims.md — verified / partial / experimental, always current · updated per release
          </p>
        </div>
      </div>
    </section>
  );
}
