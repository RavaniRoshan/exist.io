import { Logomark } from '@/components/marketing/icons';

export function Logo() {
  return (
    <span className="flex items-center gap-2.5">
      <Logomark size={22} />
      <span className="text-[17px] font-semibold tracking-[-0.03em]">exist.io</span>
      <span className="mono rounded-sm border border-fd-border px-1.5 py-0.5 text-[9.5px] uppercase tracking-[0.18em] text-fd-muted-foreground">
        docs
      </span>
    </span>
  );
}
