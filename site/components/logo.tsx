export function Logo() {
  return (
    <span className="flex items-center gap-2.5">
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path d="M6 5v9a4 4 0 0 0 4 4h6" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
        <circle cx="6" cy="4" r="2" fill="currentColor" />
        <circle cx="18" cy="18" r="2" fill="currentColor" />
        <path d="M18 5v4" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
      </svg>
      <span className="font-medium tracking-tight">exist.io</span>
      <span className="rounded-full border border-fd-border px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-[0.12em] text-fd-muted-foreground">
        docs
      </span>
    </span>
  );
}
