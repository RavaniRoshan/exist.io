'use client';

import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { IconCheck, IconPullRequest } from './icons';

type Phase = 'idle' | 'merging' | 'merged';

const CHECKS = [
  { id: 'rulesets', label: 'Rulesets & policy guard', detail: 'approvals · checks · bypass actors' },
  { id: 'secrets', label: 'Secret scan', detail: 'keys · tokens · private material' },
  { id: 'review', label: 'Review approval', detail: 'reviewer agent + human sign-off' },
] as const;

export function MergeGate() {
  const [checked, setChecked] = useState<Record<string, boolean>>({});
  const [phase, setPhase] = useState<Phase>('idle');

  const passed = CHECKS.filter((check) => checked[check.id]).length;
  const all = passed === CHECKS.length;

  useEffect(() => {
    if (phase !== 'merging') return;
    const id = window.setTimeout(() => setPhase('merged'), 1100);
    return () => window.clearTimeout(id);
  }, [phase]);

  if (phase === 'merged') {
    return (
      <div className="frame frame-strong rise-in flex h-full min-h-[280px] flex-col justify-between bg-text-1 p-6 text-canvas sm:p-8" data-testid="merge-gate-merged">
        <div>
          <div className="flex items-center justify-between gap-4">
            <span className="mono text-[11px] uppercase tracking-[0.18em] text-canvas/70">pr-#214 · agent/capture</span>
            <span className="mono rounded-full border border-canvas/25 px-2.5 py-1 text-[10px] uppercase tracking-[0.18em] text-patina">merged</span>
          </div>
          <div className="mt-8 flex items-center gap-4">
            <span className="pulse-ring relative flex h-12 w-12 items-center justify-center rounded-full border border-patina/60 bg-patina text-patina-ink">
              <IconCheck width={26} height={26} />
            </span>
            <p className="text-2xl font-medium leading-tight tracking-tight">Merge gate passed.</p>
          </div>
        </div>
        <div className="mt-8 border-t border-canvas/15 pt-5">
          <p className="mono text-[12px] leading-relaxed text-canvas/75">lease revoked · audit entry written · refs/heads/main updated</p>
          <button type="button" className="mono mt-4 text-[11px] uppercase tracking-[0.18em] text-canvas/60 underline underline-offset-4 hover:text-patina" onClick={() => { setChecked({}); setPhase('idle'); }}>
            Reset demo
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="frame frame-strong flex h-full flex-col p-5 sm:p-6" data-testid="merge-gate">
      <div className="flex items-center justify-between gap-4 border-b border-line pb-4">
        <span className="mono flex items-center gap-2 text-[11px] uppercase tracking-[0.18em] text-text-2">
          <IconPullRequest width={14} height={14} /> pr-#214 · agent/capture
        </span>
        <span className="mono text-[11px] text-text-3">{passed}/3 checks</span>
      </div>

      <fieldset className="mt-4 flex flex-col gap-2.5">
        <legend className="sr-only">Required merge checks</legend>
        {CHECKS.map((check) => {
          const isChecked = Boolean(checked[check.id]);
          return (
            <div
              key={check.id}
              className="gate-item"
              data-checked={isChecked}
              onClick={() => setChecked((state) => ({ ...state, [check.id]: !state[check.id] }))}
            >
              <span className="contents" onClick={(event) => event.stopPropagation()}>
                <Checkbox
                  unstyled
                  aria-label={check.label}
                  className="gate-box"
                  isSelected={isChecked}
                  onChange={(selected) =>
                    setChecked((state) => ({ ...state, [check.id]: selected }))
                  }
                />
              </span>
              <span className="min-w-0">
                <span className="block text-[15px] font-medium leading-tight text-text-1">{check.label}</span>
                <span className="mono mt-1 block text-[11px] leading-tight text-text-3">{check.detail}</span>
              </span>
              <span className={`mono ml-auto text-[10px] uppercase tracking-[0.18em] ${isChecked ? 'text-pass' : 'text-text-3'}`}>
                {isChecked ? 'pass' : 'wait'}
              </span>
            </div>
          );
        })}
      </fieldset>

      <div className="mt-5 border-t border-line pt-4">
        <Button
          unstyled
          type="button"
          className="btn btn-signal w-full justify-center disabled:cursor-not-allowed disabled:border-line disabled:bg-raised disabled:text-text-3"
          isDisabled={!all || phase === 'merging'}
          onPress={() => setPhase('merging')}
        >
          {phase === 'merging' ? (
            <>
              <span className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" aria-hidden="true" />
              Merging…
            </>
          ) : (
            'Merge pull request'
          )}
        </Button>
        <p className="mono mt-3 text-center text-[11px] leading-relaxed text-text-3" role="status">
          {all ? 'all checks passed — merge allowed by policy' : 'merge blocked by policy'}
        </p>
      </div>
    </div>
  );
}
