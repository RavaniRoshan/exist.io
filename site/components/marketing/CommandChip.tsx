import { CopyButton } from './CopyButton';

export function CommandChip({ command }: { command: string }) {
  return (
    <div className="cmd" role="group" aria-label="Shell command">
      <span className="select-none text-ink-3">$</span>
      <code className="cmd-scroll">{command}</code>
      <CopyButton text={command} label="Copy install command" />
    </div>
  );
}
