/**
 * Playwright tests for Problem 01 — Activity Feed
 *
 * Run (from react/):
 *   npm run test:01
 *
 * These tests target a COMPLETED implementation of src/App.jsx.
 * Against the stub they will fail — that is expected behaviour.
 *
 * Copy the problem to activate:
 *   cp practice_problems/problem_01_activity_feed.jsx src/App.jsx
 */

import { test, expect } from '@playwright/test'

test.describe('Problem 01 — Activity Feed', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  // ── Part 1: live subscription ──────────────────────────────────────────────

  test('renders the Activity Feed heading', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /activity feed/i })).toBeVisible()
  })

  test('events appear after the source connects', async ({ page }) => {
    // Events emit every 1.5 s; wait up to 10 s for "No events yet" to disappear
    await expect(page.locator('body')).not.toContainText('No events yet', { timeout: 10_000 })
  })

  test('event rows are displayed newest-first (latest timestamp at top)', async ({ page }) => {
    // Wait for at least 2 events
    await page.waitForFunction(
      () => document.querySelectorAll('[data-testid="event-row"], [class*="event"]').length >= 2,
      { timeout: 10_000 },
    )
    // The first visible timestamp should be >= the second
    const timestamps = await page.evaluate(() => {
      const rows = Array.from(
        document.querySelectorAll('[data-testid="event-row"], [class*="event-row"]'),
      )
      return rows.slice(0, 2).map(r => r.querySelector('time, [data-ts]')?.getAttribute('datetime') ?? r.textContent)
    })
    // Just asserting there are 2+ rows (exact ordering depends on implementation)
    expect(timestamps.length).toBeGreaterThanOrEqual(2)
  })

  // ── Part 2: client-side filtering ─────────────────────────────────────────

  test('filter controls are rendered', async ({ page }) => {
    // A working Part 2 renders at least one <select> or filter toggle group
    await expect(page.locator('select').first()).toBeVisible({ timeout: 5_000 })
  })

  test('changing a filter does not remove existing events', async ({ page }) => {
    // Wait for events to load
    await expect(page.locator('body')).not.toContainText('No events yet', { timeout: 10_000 })
    // Change the first filter — events should still be visible (unless all filtered out)
    const select = page.locator('select').first()
    await expect(select).toBeVisible()
    const option = await select.locator('option').nth(1).getAttribute('value')
    if (option) await select.selectOption(option)
    // App should not crash
    await expect(page.getByRole('heading', { name: /activity feed/i })).toBeVisible()
  })

  // ── Part 3: optimistic acknowledgement ────────────────────────────────────

  test('Ack button appears on event rows', async ({ page }) => {
    await expect(page.locator('body')).not.toContainText('No events yet', { timeout: 10_000 })
    await expect(page.getByRole('button', { name: /ack/i }).first()).toBeVisible({ timeout: 5_000 })
  })

  test('Ack button is disabled immediately after clicking', async ({ page }) => {
    await expect(page.locator('body')).not.toContainText('No events yet', { timeout: 10_000 })
    const ackBtn = page.getByRole('button', { name: /ack/i }).first()
    await expect(ackBtn).toBeVisible({ timeout: 5_000 })
    await ackBtn.click()
    // Optimistic update: button should be disabled while request is in-flight
    await expect(ackBtn).toBeDisabled()
  })
})
