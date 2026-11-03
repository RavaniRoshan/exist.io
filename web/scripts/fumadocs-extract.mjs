import { chromium } from 'playwright'
import { writeFileSync } from 'node:fs'

const browser = await chromium.launch({ executablePath: '/usr/bin/chromium-browser', args: ['--no-sandbox'] })
const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 1 })
await page.goto('https://www.fumadocs.dev/', { waitUntil: 'networkidle', timeout: 60000 })
await page.waitForTimeout(2000)
await page.screenshot({ path: '/tmp/fumadocs-desktop.png', fullPage: true })

const data = await page.evaluate(() => {
  const css = (sel, prop) => {
    const el = document.querySelector(sel)
    return el ? getComputedStyle(el).getPropertyValue(prop) : null
  }
  const root = getComputedStyle(document.documentElement)
  const vars = {}
  for (const sheet of Array.from(document.styleSheets)) {
    let rules = []
    try { rules = Array.from(sheet.cssRules) } catch { continue }
    for (const rule of rules) {
      if (rule.selectorText === ':root' || rule.selectorText === ':root, :host' || rule.selectorText === '.dark') {
        for (const prop of Array.from(rule.style)) {
          if (prop.startsWith('--')) vars[prop] = rule.style.getPropertyValue(prop).trim()
        }
      }
    }
  }
  const nav = document.querySelector('header, nav')
  const navLinks = Array.from(document.querySelectorAll('header a, nav a')).slice(0, 30).map((a) => ({
    text: a.textContent?.trim().replace(/\s+/g, ' ').slice(0, 40),
    href: a.getAttribute('href'),
  }))
  const headings = Array.from(document.querySelectorAll('h1, h2')).map((h) => ({
    tag: h.tagName,
    text: h.textContent?.trim().replace(/\s+/g, ' ').slice(0, 70),
    size: getComputedStyle(h).fontSize,
    weight: getComputedStyle(h).fontWeight,
    family: getComputedStyle(h).fontFamily,
  }))
  const sections = Array.from(document.querySelectorAll('main > *, body > div > main > *')).slice(0, 30).map((el) => ({
    tag: el.tagName,
    cls: (el.getAttribute('class') || '').slice(0, 100),
    text: el.textContent?.trim().replace(/\s+/g, ' ').slice(0, 60),
    height: Math.round(el.getBoundingClientRect().height),
  }))
  return {
    title: document.title,
    bodyFont: css('body', 'font-family'),
    bodyBg: css('body', 'background-color'),
    bodyColor: css('body', 'color'),
    navHeight: nav ? Math.round(nav.getBoundingClientRect().height) : null,
    navLinks,
    headings,
    sections,
    rootRadius: root.getPropertyValue('--radius'),
    vars: Object.fromEntries(Object.entries(vars).slice(0, 120)),
  }
})
writeFileSync('/tmp/fumadocs-design.json', JSON.stringify(data, null, 2))
console.log(JSON.stringify({ title: data.title, bodyFont: data.bodyFont, bodyBg: data.bodyBg, navHeight: data.navHeight, headings: data.headings.slice(0, 8), sections: data.sections.length, vars: Object.keys(data.vars).length }, null, 2))

const mobile = await browser.newPage({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 1 })
await mobile.goto('https://www.fumadocs.dev/', { waitUntil: 'networkidle', timeout: 60000 })
await mobile.waitForTimeout(1500)
await mobile.screenshot({ path: '/tmp/fumadocs-mobile.png', fullPage: true })
await browser.close()
