import { chromium } from 'playwright';
const base = process.env.SITE_URL ?? 'http://127.0.0.1:4173/exist.io/';
const browser = await chromium.launch({ executablePath: '/usr/bin/chromium-browser', args: ['--no-sandbox'] });
const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 1 });
await page.goto(base, { waitUntil: 'networkidle', timeout: 60000 });
await page.evaluate(async () => {
  for (let y = 0; y < document.body.scrollHeight; y += 700) {
    window.scrollTo(0, y)
    await new Promise((resolve) => setTimeout(resolve, 120))
  }
  window.scrollTo(0, 0)
})
await page.evaluate(() => {
  document.querySelectorAll('.reveal').forEach((el) => el.classList.add('is-visible'))
})
await page.waitForTimeout(800)
await page.screenshot({ path: '/tmp/exist-io-site.png', fullPage: true })
const mobile = await browser.newPage({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 1 })
await mobile.goto(base, { waitUntil: 'networkidle', timeout: 60000 })
await mobile.evaluate(() => {
  document.querySelectorAll('.reveal').forEach((el) => el.classList.add('is-visible'))
})
await mobile.waitForTimeout(800)
await mobile.screenshot({ path: '/tmp/exist-io-site-mobile.png', fullPage: true })
await browser.close()
console.log('captured', base)
