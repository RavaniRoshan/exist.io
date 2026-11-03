import { readFileSync } from 'node:fs';

// Machine gate for the marketing landing: every capability statement on the
// page must still trace to CLAIMS.md. Needles are matched case-insensitively
// against markup-stripped text so wording can drift without breaking the gate,
// but the underlying claims cannot silently vanish or weaken.

const raw = readFileSync(new URL('../out/index.html', import.meta.url), 'utf8');
const html = raw
  .replace(/<script[\s\S]*?<\/script>/gi, ' ')
  .replace(/<style[\s\S]*?<\/style>/gi, ' ')
  .replace(/<[^>]+>/g, ' ')
  .replace(/&nbsp;|&#160;|&amp;nbsp;/g, ' ')
  .replace(/&amp;/g, '&')
  .replace(/&#x27;|&#39;|&apos;|’/g, "'")
  .replace(/&quot;|&#34;/g, '"')
  .replace(/&lt;/g, '<')
  .replace(/&gt;/g, '>')
  .replace(/\s+/g, ' ')
  .toLowerCase();

const required = [
  ['the self-hosted forge for', 'hero headline'],
  ['coding agents.', 'hero accent'],
  ['install.sh | bash', 'install command'],
  ['existd bootstrap --email admin@example.com', 'bootstrap command'],
  ['refs/heads/agent/*', 'lease isolation'],
  ['v0.6.0-beta', 'current release'],
  ['apache-2.0', 'license'],
  [':8080', 'single-port web UI / api / mcp'],
  [':2222', 'embedded ssh'],
];

const forbidden = [
  ['35 mcp', 'stale tool count'],
  ['35 tools', 'stale tool count'],
  ['119 rest', 'stale operation count'],
  ['119 operations', 'stale operation count'],
  ['137 frontend', 'stale test count'],
  ['37 suites', 'stale test count'],
  ['25 go packages', 'stale package count'],
  ['primer', 'stale design language'],
  ['enterprise-ready', 'unverified enterprise claim'],
  ['enterprise ready', 'unverified enterprise claim'],
  ['wire-compatible', 'unverified compatibility claim'],
  ['full wire compatibility', 'unverified compatibility claim'],
  ['space ui', 'stale design language'],
  [':3000', 'stale web ui port'],
  ['production-ready terraform', 'demoted claim'],
  ['merge queue', 'remediated claim'],
  ['todo', 'placeholder'],
  ['tbd', 'placeholder'],
  ['lorem', 'placeholder'],
];

let failed = false;

for (const [needle, label] of required) {
  if (!html.includes(needle)) {
    console.error(`MISSING (${label}): "${needle}"`);
    failed = true;
  }
}
for (const [needle, label] of forbidden) {
  if (html.includes(needle)) {
    console.error(`FORBIDDEN (${label}): "${needle}"`);
    failed = true;
  }
}

if (failed) {
  console.error('\nclaims audit FAILED');
  process.exit(1);
}
console.log('claims audit passed —', required.length, 'claims verified,', forbidden.length, 'forbidden phrases absent');
