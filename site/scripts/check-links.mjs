import { readdirSync, readFileSync, existsSync, statSync } from 'node:fs';
import { extname, join, posix, relative, resolve } from 'node:path';

const outDir = resolve(import.meta.dirname, '..', 'out');
const basePath = '/exist.io';

if (!existsSync(outDir)) {
  console.error('check-links: out/ not found. Run `npm run build` first.');
  process.exit(1);
}

function* walk(dir) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) yield* walk(path);
    else yield path;
  }
}

const pages = [...walk(outDir)].filter((file) => extname(file) === '.html');
const failures = [];
const fileUrlPattern = /\b(?:href|src)\s*=\s*(['"])(file:\/\/\/[^'"]*)\1/gi;

function resolvesToFile(pathname) {
  const clean = decodeURIComponent(pathname.split('#')[0].split('?')[0]);
  const stripped = clean.startsWith(basePath) ? clean.slice(basePath.length) || '/' : clean;
  const target = stripped.startsWith('/') ? stripped.slice(1) : stripped;
  const candidates = [
    join(outDir, target),
    join(outDir, target, 'index.html'),
    join(outDir, `${target}.html`),
    join(outDir, target, 'index.txt'),
  ];

  return candidates.some((candidate) => existsSync(candidate));
}

let checked = 0;

for (const file of pages) {
  const html = readFileSync(file, 'utf8');
  const page = `/${relative(outDir, file).split(/[\\/]/).join('/')}`;

  for (const match of html.matchAll(fileUrlPattern)) {
    failures.push(`${page}: host-specific link ${match[2]}`);
  }

  if (html.includes('/home/')) {
    failures.push(`${page}: contains a developer home-directory path`);
  }

  for (const match of html.matchAll(/\b(?:href|src)="([^"]+)"/g)) {
    const value = match[1];
    if (!value.startsWith('/') || value.startsWith('//')) continue;

    checked += 1;
    if (!resolvesToFile(value)) {
      failures.push(`${page}: broken link ${value}`);
    }
  }
}

if (failures.length) {
  console.error(`check-links: ${failures.length} problem(s) found`);
  for (const failure of failures.slice(0, 50)) console.error(`  ${failure}`);
  process.exit(1);
}

console.log(`check-links: ${pages.length} pages, ${checked} internal links verified`);
