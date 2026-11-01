import { readdir, readFile, writeFile } from 'node:fs/promises'
import { join, relative, sep } from 'node:path'
import { fileURLToPath } from 'node:url'

const distDir = fileURLToPath(new URL('../dist/', import.meta.url))
const origin = 'https://ravaniroshan.github.io/exist.io'

async function htmlFiles(dir) {
  const entries = await readdir(dir, { withFileTypes: true })
  const files = []
  for (const entry of entries) {
    const full = join(dir, entry.name)
    if (entry.isDirectory()) files.push(...(await htmlFiles(full)))
    else if (entry.name.endsWith('.html')) files.push(full)
  }
  return files
}

function canonicalUrl(file) {
  const rel = relative(distDir, file).split(sep).join('/')
  if (rel === 'index.html') return `${origin}/`
  if (rel.endsWith('/index.html')) return `${origin}/${rel.slice(0, -'index.html'.length)}`
  return `${origin}/${rel}`
}

const files = await htmlFiles(distDir)
let injected = 0
let pages = 0

for (const file of files) {
  if (file.endsWith('404.html')) continue
  pages += 1
  const html = await readFile(file, 'utf8')
  if (html.includes('rel="canonical"')) continue
  if (!html.includes('</head>')) continue
  const tag = `<link rel="canonical" href="${canonicalUrl(file)}">`
  await writeFile(file, html.replace('</head>', `  ${tag}\n</head>`), 'utf8')
  injected += 1
}

console.log(`canonical: injected ${injected} tag(s) across ${pages} page(s)`)
