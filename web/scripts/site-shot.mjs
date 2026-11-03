import { chromium } from 'playwright'
const browser = await chromium.launch({ executablePath: '/usr/bin/chromium-browser', args: ['--no-sandbox'] })
const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 1 })
await page.goto('http://127.0.0.1:4321/exist.os/', { waitUntil: 'networkidle', timeout: 60000 })
await page.evaluate(async () => {
  for (let y = 0; y < document.body.scrollHeight; y += 700) {
    window.scrollTo(0, y)
    await new Promise((resolve) => setTimeout(resolve, 120))
  }
  window.scrollTo(0, 0)
})
await page.evaluate(() => {
  document.querySelectorAll('[data-reveal]').forEach((el) => el.classList.add('is-visible'))
})
await page.waitForTimeout(800)
await page.screenshot({ path: '/tmp/exist-os-fumadocs.png', fullPage: true })
const mobile = await browser.newPage({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 1 })
await mobile.goto('http://127.0.0.1:4321/exist.os/', { waitUntil: 'networkidle', timeout: 60000 })
await mobile.waitForTimeout(1200)
await mobile.screenshot({ path: '/tmp/exist-os-fumadocs-mobile.png', fullPage: true })
await browser.close()
console.log('captured')
