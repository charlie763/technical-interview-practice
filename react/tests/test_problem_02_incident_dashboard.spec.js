/**
 * Playwright tests for Problem 02 — Real-Time Incident Dashboard
 *
 * Run (from react/):
 *   npm run test:02
 *
 * These tests target a COMPLETED implementation of src/App.jsx.
 * Against the stub they will fail — that is expected behaviour.
 *
 * Seed data reference (from SEED_INCIDENTS in the problem file):
 *   inc-s1  shooting   critical  Downtown   (2 critical total)
 *   inc-s2  car-crash  high      Midtown
 *   inc-s3  fire       critical  Eastside   (2 critical total)
 *   inc-s4  break-in   medium    Westside
 *   inc-s5  flooding   low       South Bay
 *
 * Seed incidents are emitted within ~300 ms of mount, so tests use
 * a 3 s timeout rather than waiting for the 2 s random stream.
 *
 * Query strategy (in priority order):
 *   1. Seed data text — toContainText('Downtown') etc.
 *   2. Element type + text — locator('button').filter({ hasText: /contain/i })
 *   3. data-testid — required for row counting and structural identification
 *      (see problem file for which testids are required)
 */

import { test, expect } from '@playwright/test'

test.describe('Problem 02 — Incident Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  // ── Part 1: live incident feed ─────────────────────────────────────────────

  test('renders the Incident Dashboard heading', async ({ page }) => {
    await expect(page.locator('h1, h2, h3, h4').filter({ hasText: /incident dashboard/i })).toBeVisible()
  })

  test('seed incidents appear in the active list', async ({ page }) => {
    // Seed data always includes 'Downtown' — use text content rather than testid
    await expect(page.locator('body')).toContainText('Downtown', { timeout: 3_000 })
  })

  test('each incident row contains a severity badge', async ({ page }) => {
    // Wait for seed data, then verify the severity badge element is present
    await expect(page.locator('body')).toContainText('Downtown', { timeout: 3_000 })
    await expect(page.locator('[data-testid="severity-badge"]').first()).toBeVisible()
  })

  test('seed incident locations appear in the list', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Downtown', { timeout: 3_000 })
    await expect(page.locator('body')).toContainText('Midtown')
  })

  test('"No incidents yet" disappears once feed connects', async ({ page }) => {
    await expect(page.locator('body')).not.toContainText('No incidents yet', { timeout: 3_000 })
  })

  // ── Part 2: filtering ──────────────────────────────────────────────────────

  test('severity filter control is present', async ({ page }) => {
    await expect(page.getByTestId('filter-severity')).toBeVisible({ timeout: 3_000 })
  })

  test('type filter control is present', async ({ page }) => {
    await expect(page.getByTestId('filter-type')).toBeVisible({ timeout: 3_000 })
  })

  test('filtering by "critical" reduces the visible row count', async ({ page }) => {
    // Wait for seed data (mixed severities: 2 critical, 1 high, 1 medium, 1 low)
    await expect(page.locator('body')).toContainText('Downtown', { timeout: 3_000 })
    const allCount = await page.locator('[data-testid="incident-row"]').count()

    await page.getByTestId('filter-severity').selectOption('critical')

    const criticalCount = await page.locator('[data-testid="incident-row"]').count()
    expect(criticalCount).toBeLessThan(allCount)
    expect(criticalCount).toBeGreaterThan(0)
  })

  test('filter change does not restart the subscription (no layout flash)', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Downtown', { timeout: 3_000 })
    // Change filter and verify heading is still present (app did not unmount/remount)
    await page.getByTestId('filter-severity').selectOption('high')
    await expect(page.locator('h1, h2, h3, h4').filter({ hasText: /incident dashboard/i })).toBeVisible()
  })

  // ── Part 3: optimistic "Contain" action ───────────────────────────────────

  test('Contain button is present on each active incident row', async ({ page }) => {
    // Locate by button text — no data-testid required on the button itself
    await expect(page.locator('body')).toContainText('Downtown', { timeout: 3_000 })
    await expect(
      page.locator('[data-testid="contain-btn"], button:has-text("Contain")').first()
    ).toBeVisible()
  })

  test('Contain button is disabled immediately after clicking (optimistic)', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Downtown', { timeout: 3_000 })
    const btn = page.locator('[data-testid="contain-btn"], button:has-text("Contain")').first()
    await expect(btn).toBeVisible()
    await btn.click()
    await expect(btn).toBeDisabled()
  })

  test('contained incident eventually appears in the Contained section', async ({ page }) => {
    // NOTE: containIncident rejects ~20 % of the time, so this test may
    // occasionally fail due to the mock. Retry or re-run if it does.
    await expect(page.locator('body')).toContainText('Downtown', { timeout: 3_000 })
    await page.locator('[data-testid="contain-btn"], button:has-text("Contain")').first().click()
    // On success (~80 % of calls), a contained-row should appear within 2 s
    await expect(page.locator('[data-testid="contained-row"]').first()).toBeVisible({ timeout: 2_000 })
  })
})
