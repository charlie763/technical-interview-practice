/**
 * Playwright tests for Problem 03 — Alert Triage Console
 *
 * Run (from react/):
 *   npm run test:03
 *
 * These tests target a COMPLETED implementation of src/App.jsx.
 * Against the stub they will fail — that is expected behaviour.
 *
 * Seed data reference (from SEED_ALERTS in the problem file):
 *   alt-s1  radio-dispatch  sev=5  "Reports of shots fired at Central Station"
 *   alt-s2  sensor-alert    sev=4  "Collision detected on Highway 12"
 *   alt-s3  radio-dispatch  sev=3  "Structural fire reported at Warehouse District"
 *   alt-s4  sensor-alert    sev=2  "Flooding sensors triggered at Riverside Park"
 *
 * Seed alerts are emitted within ~300 ms of mount.
 *
 * Query strategy (in priority order):
 *   1. Seed data text — toContainText('Central Station') etc.
 *   2. Element type + text — locator('button').filter({ hasText: /assign/i })
 *      Buttons located this way: Assign, Ack All, tab buttons (Queued/Assigned/Acknowledged)
 *   3. data-testid — required for counting rows and the unit select dropdown
 *      (see problem file for which testids are required)
 */

import { test, expect } from '@playwright/test'

test.describe('Problem 03 — Alert Triage Console', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  // ── Part 1: alert queue ────────────────────────────────────────────────────

  test('renders the Alert Triage Console heading', async ({ page }) => {
    await expect(page.locator('h1, h2, h3, h4').filter({ hasText: /alert triage console/i })).toBeVisible()
  })

  test('seed alerts appear in the list', async ({ page }) => {
    // Seed data always includes 'Central Station' — use text rather than testid
    await expect(page.locator('body')).toContainText('Central Station', { timeout: 3_000 })
  })

  test('each alert row contains a severity badge', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Central Station', { timeout: 3_000 })
    await expect(page.locator('[data-testid="severity-badge"]').first()).toBeVisible()
  })

  test('seed alert message text appears in the list', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Central Station', { timeout: 3_000 })
    await expect(page.locator('body')).toContainText('Highway 12')
  })

  test('alerts are sorted highest severity first', async ({ page }) => {
    // Wait for at least 4 seed rows (sev 5, 4, 3, 2)
    await expect(page.locator('[data-testid="alert-row"]')).toHaveCount(4, { timeout: 3_000 })
    const firstBadgeText = await page.locator('[data-testid="severity-badge"]').first().textContent()
    // The first badge should represent the highest severity (sev=5)
    expect(firstBadgeText).toMatch(/5|P1|critical/i)
  })

  // ── Part 2: assign to unit ─────────────────────────────────────────────────

  test('unit select is present on each unassigned row', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Central Station', { timeout: 3_000 })
    await expect(page.locator('[data-testid="unit-select"]').first()).toBeVisible()
  })

  test('assign button is present on each unassigned row', async ({ page }) => {
    // Locate by button text — no data-testid required on the button itself
    await expect(page.locator('body')).toContainText('Central Station', { timeout: 3_000 })
    await expect(
      page.locator('[data-testid="assign-btn"], button:has-text("Assign")').first()
    ).toBeVisible()
  })

  test('unit select is populated with available units', async ({ page }) => {
    await expect(page.locator('[data-testid="unit-select"]').first()).toBeVisible({ timeout: 3_000 })
    const options = await page.locator('[data-testid="unit-select"]').first().locator('option').allTextContents()
    // Should include unit names from UNITS constant
    expect(options.join(' ')).toMatch(/Alpha Team|Beta Team|Fire Unit/i)
  })

  test('assign button and unit select are disabled while request is in-flight', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Central Station', { timeout: 3_000 })
    // Select a unit first (pick second option to avoid any placeholder)
    const sel = page.locator('[data-testid="unit-select"]').first()
    await sel.selectOption({ index: 1 })
    const btn = page.locator('[data-testid="assign-btn"], button:has-text("Assign")').first()
    await btn.click()
    await expect(btn).toBeDisabled()
    await expect(sel).toBeDisabled()
  })

  // ── Part 3: tab bar and bulk-ack ──────────────────────────────────────────

  test('Queued tab is present and visible', async ({ page }) => {
    // Located by tab text — data-testid="tab-queued" also accepted
    await expect(
      page.locator('[data-testid="tab-queued"], button:has-text("Queued")').first()
    ).toBeVisible({ timeout: 3_000 })
  })

  test('Assigned tab is present and visible', async ({ page }) => {
    await expect(
      page.locator('[data-testid="tab-assigned"], button:has-text("Assigned")').first()
    ).toBeVisible({ timeout: 3_000 })
  })

  test('Acknowledged tab is present and visible', async ({ page }) => {
    await expect(
      page.locator('[data-testid="tab-acknowledged"], button:has-text("Acknowledged")').first()
    ).toBeVisible({ timeout: 3_000 })
  })

  test('Ack All button is visible on the Queued tab', async ({ page }) => {
    // Ensure we're on the Queued tab (default) and rows are loaded
    await expect(page.locator('body')).toContainText('Central Station', { timeout: 3_000 })
    await expect(
      page.locator('[data-testid="ack-all-btn"], button:has-text("Ack All")').first()
    ).toBeVisible()
  })

  test('switching to Assigned tab hides unassigned rows', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Central Station', { timeout: 3_000 })
    await page.locator('[data-testid="tab-assigned"], button:has-text("Assigned")').first().click()
    // No rows should be visible yet (none have been assigned)
    const rowCount = await page.locator('[data-testid="alert-row"]').count()
    expect(rowCount).toBe(0)
  })

  test('Ack All acknowledges all visible queued alerts', async ({ page }) => {
    // Wait for seed rows on Queued tab
    await expect(page.locator('[data-testid="alert-row"]').first()).toBeVisible({ timeout: 3_000 })
    const queuedBefore = await page.locator('[data-testid="alert-row"]').count()
    expect(queuedBefore).toBeGreaterThan(0)

    await page.locator('[data-testid="ack-all-btn"], button:has-text("Ack All")').first().click()

    // After ack-all (300 ms per ack via mock), Queued tab should have 0 rows
    // Allow up to 3 s for all acks to complete
    await expect(page.locator('[data-testid="alert-row"]')).toHaveCount(0, { timeout: 3_000 })

    // Acknowledged tab should now show the acked rows
    await page.locator('[data-testid="tab-acknowledged"], button:has-text("Acknowledged")').first().click()
    await expect(page.locator('[data-testid="alert-row"]')).toHaveCount(queuedBefore, { timeout: 1_000 })
  })
})
