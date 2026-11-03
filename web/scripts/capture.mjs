import { readFileSync } from 'node:fs'
import { chromium } from 'playwright'

const work = process.env.EXIST_CAPTURE_DIR || '/tmp/exist-capture'
const port = readFileSync(`${work}/port`, 'utf8').trim()
const token = readFileSync(`${work}/key`, 'utf8').trim()

const browser = await chromium.launch({
  executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium-browser',
  args: ['--no-sandbox'],
})
const page = await browser.newPage({ viewport: { width: 1280, height: 800 }, deviceScaleFactor: 1 })

await page.addInitScript((value) => {
  window.localStorage.setItem('exist_token', value)
  window.localStorage.setItem('exist_onboarding_tour_completed', 'true')
}, token)

await page.goto(`http://127.0.0.1:${port}/?tab=prs`, { waitUntil: 'networkidle' })
await page.waitForTimeout(1500)

const prRow = page.getByText('Add Forge capture note').first()
if (await prRow.isVisible().catch(() => false)) {
  await prRow.click()
  await page.waitForTimeout(1500)
}

await page.mouse.move(8, 8)
await page.waitForTimeout(400)

await page.screenshot({
  path: new URL('../../assets/launch/02-web-pr-view.png', import.meta.url).pathname,
  type: 'png',
})

await browser.close()
console.log('captured assets/launch/02-web-pr-view.png')
