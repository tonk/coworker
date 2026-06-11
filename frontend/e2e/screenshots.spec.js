import { test, expect } from '@playwright/test'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const SS = path.resolve(__dirname, '../../screenshots')
const AUTH_FILE = path.resolve(__dirname, '.auth.json')

const BASE_URL = 'http://localhost:5173'

// Helpers ---------------------------------------------------------------
function ss(name) {
  return { path: path.join(SS, `${name}.png`), fullPage: true }
}

/** Cookie-based login — sets httpOnly session cookies on the browser context. */
async function loginViaApi(context, username = 'demo.admin', password = 'demo1234') {
  const res = await context.request.post(`${BASE_URL}/api/v1/auth/login`, {
    data: { login: username, password },
  })
  if (!res.ok()) {
    throw new Error(`API login failed: ${res.status()} ${await res.text()}`)
  }
}

async function loginAs(page, username = 'demo.admin', password = 'demo1234') {
  await page.goto(`${BASE_URL}/login`)
  await page.evaluate(() => localStorage.removeItem('mfa_trust_token'))
  await page.waitForSelector('.form-input')
  const inputs = page.locator('.form-input')
  await inputs.nth(0).fill(username)
  await inputs.nth(1).fill(password)
  await page.locator('button[type="submit"]').click()
  try {
    await page.waitForURL(url => !new URL(url).pathname.startsWith('/login'), { timeout: 15000 })
  } catch {
    // Fallback: API login (e.g. auth rate limit after repeated runs).
    await loginViaApi(page.context(), username, password)
    await page.goto(`${BASE_URL}/`)
    await page.waitForLoadState('networkidle')
  }
}

// Dismiss the welcome/news backdrop that appears on every fresh context
async function dismissWelcome(page) {
  const welcomeClose = page.locator('.welcome-close')
  if (await welcomeClose.isVisible({ timeout: 3000 }).catch(() => false)) {
    await welcomeClose.click()
    await page.waitForTimeout(400)
  }
}

// Tests ------------------------------------------------------------------
test.describe('screenshots', () => {

  // ── 01: Login page ─────────────────────────────────────────────────
  test('01-login', async ({ browser }) => {
    const context = await browser.newContext()
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/login`)
    await page.waitForSelector('.form-input')
    await page.screenshot(ss('01-login'))
    await loginAs(page)
    await context.storageState({ path: AUTH_FILE })
    await context.close()
  })

  // ── 02: Dashboard ──────────────────────────────────────────────────
  test('02-dashboard', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(BASE_URL)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    await page.screenshot(ss('02-dashboard'))
    await context.close()
  })

  // ── 03: Board (kanban) ─────────────────────────────────────────────
  test('03-board', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/projects/website-redesign`)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.board-column')
    await dismissWelcome(page)
    await page.screenshot(ss('03-board'))
    await context.close()
  })

  // ── 04: Card detail modal ──────────────────────────────────────────
  test('04-card-detail', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/projects/website-redesign`)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.board-card')
    await dismissWelcome(page)
    await page.locator('.board-card').first().click()
    await page.waitForSelector('.modal-backdrop')
    await page.screenshot(ss('04-card-detail'))
    await context.close()
  })

  // ── 05: Topics / discussions ───────────────────────────────────────
  test('05-topics', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/projects/website-redesign/topics`)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.topic-item')
    await dismissWelcome(page)
    await page.locator('.topic-item').first().click()
    await page.waitForSelector('.topic-detail')
    await page.screenshot(ss('05-topics'))
    await context.close()
  })

  // ── 06: Messages (direct messages) ─────────────────────────────────
  test('06-messages', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/chats`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    await page.screenshot(ss('06-messages'))
    await context.close()
  })

  // ── 07: Report ─────────────────────────────────────────────────────
  test('07-report', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/reports`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    await page.screenshot(ss('07-report'))
    await context.close()
  })

  // ── 08: Admin — Users tab ──────────────────────────────────────────
  test('08-admin-users', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/admin`)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.data-table')
    await dismissWelcome(page)
    await page.screenshot(ss('08-admin-users'))
    await context.close()
  })

  // ── 09: Admin — Settings tab ───────────────────────────────────────
  test('09-admin-settings', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/admin`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    await page.locator('.tab:has-text("Settings")').click()
    await page.waitForTimeout(500)
    await page.screenshot(ss('09-admin-settings'))
    await context.close()
  })

  // ── 10: User settings ──────────────────────────────────────────────
  test('10-user-settings', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/settings`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    await page.screenshot(ss('10-user-settings'))
    await context.close()
  })

  // ── 11–12: Chat reactions ─────────────────────────────────────────
  // NOTE: Reaction hover/selected screenshots require a chat with
  // existing messages that have reactions — capture manually.

  // ── 13: Gantt chart ────────────────────────────────────────────────
  test('13-gant', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/projects/website-redesign/gantt`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    await page.screenshot(ss('13-gant'))
    await context.close()
  })

  // ── 14: Cumulative flow diagram ────────────────────────────────────
  test('14-cumulative', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/projects/product-platform/charts`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    const cumTab = page.locator('.charts-tabs .tab', { hasText: /^🌊/ })
    if (await cumTab.isVisible()) {
      await cumTab.click()
      await page.waitForTimeout(1000)
    }
    await page.screenshot(ss('14-cumulative'))
    await context.close()
  })

  // ── 15: Scrum backlog ──────────────────────────────────────────────
  test('15-scrum-backlog', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/projects/product-platform/backlog`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    await page.screenshot(ss('15-scrum-backlog'))
    await context.close()
  })

  // ── 16: Throughput chart ───────────────────────────────────────────
  test('16-scrum-throughput', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/projects/product-platform/charts`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    const thruTab = page.locator('.charts-tabs .tab', { hasText: /^🔢/ })
    if (await thruTab.isVisible()) {
      await thruTab.click()
      await page.waitForTimeout(1000)
    }
    await page.screenshot(ss('16-scrum-throughput'))
    await context.close()
  })

  // ── 17: Burndown chart ─────────────────────────────────────────────
  test('17-scrum-burndown', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/projects/product-platform/charts`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    const bdTab = page.locator('.charts-tabs .tab', { hasText: /^📉/ })
    if (await bdTab.isVisible()) {
      await bdTab.click()
      await page.waitForTimeout(500)
      const sprintSelect = page.locator('.sprint-select')
      if (await sprintSelect.isVisible()) {
        const options = await sprintSelect.locator('option').all()
        if (options.length > 1) {
          await sprintSelect.selectOption({ index: 1 })
          await page.waitForTimeout(1500)
        }
      }
    }
    await page.screenshot(ss('17-scrum-burndown'))
    await context.close()
  })

  // ── 18: Burnup chart ──────────────────────────────────────────────
  test('18-scrum-burnup', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/projects/product-platform/charts`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    const buTab = page.locator('.charts-tabs .tab', { hasText: /^📈/ })
    if (await buTab.isVisible()) {
      await buTab.click()
      await page.waitForTimeout(500)
      const sprintSelect = page.locator('.sprint-select')
      if (await sprintSelect.isVisible()) {
        const options = await sprintSelect.locator('option').all()
        if (options.length > 1) {
          await sprintSelect.selectOption({ index: 1 })
          await page.waitForTimeout(1500)
        }
      }
    }
    await page.screenshot(ss('18-scrum-burnup'))
    await context.close()
  })

  // ── 19: Release chart ──────────────────────────────────────────────
  test('19-scrum-release', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/projects/product-platform/charts`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    const relTab = page.locator('.charts-tabs .tab', { hasText: /^🚀/ })
    if (await relTab.isVisible()) {
      await relTab.click()
      await page.waitForTimeout(1000)
    }
    await page.screenshot(ss('19-scrum-release'))
    await context.close()
  })

  // ── 20: Standby shift / Time tracking ─────────────────────────────
  // Shows the weekly grid with the Rate column and Mon–Sun standby entries
  // (dots = start/end time recorded, ✦ = special rate tier active).
  test('20-standby-shift', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/time-tracking`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    // Wait for the grid rows to render (time entries loaded)
    await page.waitForSelector('.tt-row', { timeout: 8000 }).catch(() => {})
    await page.screenshot(ss('20-standby-shift'))
    await context.close()
  })

  // ── 21: Ticket list ────────────────────────────────────────────────
  test('21-ticket-list', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(BASE_URL)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    // Navigate to first helpdesk customer tickets via sidebar
    const ticketLink = page.locator('[id^="section-body-helpdesk"] .sidebar-link').first()
    if (await ticketLink.isVisible()) {
      const href = await ticketLink.getAttribute('href')
      await page.goto(`${BASE_URL}${href}`)
    } else {
      await page.goto(`${BASE_URL}/customers`)
    }
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    await page.waitForSelector('.ticket-card', { timeout: 5000 }).catch(() => {})
    await page.screenshot(ss('21-ticket-list'))
    await context.close()
  })

  // ── 22: Ticket detail ──────────────────────────────────────────────
  test('22-ticket-detail', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(BASE_URL)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    const ticketLink = page.locator('[id^="section-body-helpdesk"] .sidebar-link').first()
    if (await ticketLink.isVisible()) {
      const href = await ticketLink.getAttribute('href')
      await page.goto(`${BASE_URL}${href}`)
      await page.waitForLoadState('networkidle')
      await dismissWelcome(page)
      await page.waitForSelector('.ticket-card', { timeout: 5000 }).catch(() => {})
      const firstCard = page.locator('.ticket-card').first()
      if (await firstCard.isVisible()) {
        await firstCard.click()
        await page.waitForLoadState('networkidle')
        // Wait for the full ticket detail to render — title, messages section, time section
        await page.waitForSelector('.ticket-detail-main', { timeout: 8000 }).catch(() => {})
        await page.waitForSelector('.ticket-messages', { timeout: 5000 }).catch(() => {})
        await page.waitForLoadState('networkidle')
        await page.waitForTimeout(800)
      }
    }
    await page.screenshot(ss('22-ticket-detail'))
    await context.close()
  })

  // ── 24: Time tracking — undeclarable alignment ─────────────────────
  test('24-time-tracking-undeclarable', async ({ browser }) => {
    const context = await browser.newContext()
    const page = await context.newPage()
    await loginAs(page, 'tonk', 'demo1234')
    await page.goto(`${BASE_URL}/time-tracking`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    await page.waitForSelector('.tt-table', { timeout: 10000 })
    await page.waitForSelector('.cell-undecl', { timeout: 10000 })
    const firstUndecl = page.locator('.cell-undecl').first()
    const undeclId = await firstUndecl.getAttribute('id')
    expect(undeclId).toBeTruthy()
    await expect(page.locator(`[aria-describedby="${undeclId}"]`)).toHaveCount(1)
    await expect(firstUndecl.locator('.sr-only')).toContainText('Undeclarable')
    await page.locator('.tt-table').screenshot({
      path: path.join(SS, '24-time-tracking-undeclarable.png'),
    })
    await context.close()
  })

  // ── 23: Ticket inbox ───────────────────────────────────────────────
  test('23-ticket-inbox', async ({ browser }) => {
    const context = await browser.newContext({ storageState: AUTH_FILE })
    const page = await context.newPage()
    await page.goto(`${BASE_URL}/tickets/inbox`)
    await page.waitForLoadState('networkidle')
    await dismissWelcome(page)
    await page.waitForSelector('.ticket-card, .empty-state', { timeout: 8000 }).catch(() => {})
    await page.screenshot(ss('23-ticket-inbox'))
    await context.close()
  })
})
