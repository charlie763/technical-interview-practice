/**
 * Playwright tests for Problem 08 — Claims Tracker
 *
 * Run (from react/):
 *   PRACTICE_ANSWER=practice_problem_answers/cw_answer_08_claims_tracker npm run test:08
 *
 * These tests target a COMPLETED implementation of the problem stub.
 * Against the stub they will fail — that is expected behaviour.
 *
 * Seed data reference (from SEED_CLAIMS in the problem file):
 *   CLM-2001  POL-101  epl        investigating  $75,000   1 note
 *   CLM-2002  POL-102  do         evaluation     $150,000  0 notes
 *   CLM-2003  POL-101  epl        filed          $30,000   0 notes
 *   CLM-2004  POL-103  fiduciary  settled        $40,000   1 note
 *   CLM-2005  POL-104  do         denied         $20,000   0 notes
 *
 * Total claimed from seed: $315,000 across 5 claims.
 *
 * Query strategy (in priority order):
 *   1. Seed data text — toContainText('CLM-2001'), /315,000/ etc.
 *   2. Element type + visible text — locator('button').filter({ hasText: /add note/i })
 *   3. data-testid — claim rows, tabs, detail panel, note spinner, note textarea
 */

import { test, expect } from '@playwright/test'

test.describe('Problem 08 — Claims Tracker', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  // ── Part 1: claims list ───────────────────────────────────────────────────

  test('renders 5 claim rows from seed data', async ({ page }) => {
    await expect(page.locator('[data-testid="claim-row"]')).toHaveCount(5)
  })

  test('seed claim IDs are visible', async ({ page }) => {
    await expect(page.locator('body')).toContainText('CLM-2001')
    await expect(page.locator('body')).toContainText('CLM-2004')
    await expect(page.locator('body')).toContainText('CLM-2005')
  })

  test('total claimed amount is visible in summary', async ({ page }) => {
    // $315,000 total — any format matching 315,000 or 315000
    await expect(page.locator('body')).toContainText(/315[,.]?000/)
  })

  test('claim count is visible in summary', async ({ page }) => {
    await expect(page.locator('body')).toContainText(/5.{0,10}claim/i)
  })

  // ── Part 2: polling and status tabs ──────────────────────────────────────

  test('"All" tab is present', async ({ page }) => {
    await expect(page.getByTestId('tab-all')).toBeVisible()
  })

  test('"investigating" tab is present', async ({ page }) => {
    await expect(page.getByTestId('tab-investigating')).toBeVisible()
  })

  test('"settled" tab is present', async ({ page }) => {
    await expect(page.getByTestId('tab-settled')).toBeVisible()
  })

  test('filtering by "investigating" tab shows only investigating claims', async ({ page }) => {
    await page.getByTestId('tab-investigating').click()
    // CLM-2001 is investigating; CLM-2003 is filed (not yet upgraded by poll)
    await expect(page.locator('[data-testid="claim-row"]')).toHaveCount(1)
    await expect(page.locator('body')).toContainText('CLM-2001')
  })

  test('filtering by "settled" shows only settled claims', async ({ page }) => {
    await page.getByTestId('tab-settled').click()
    await expect(page.locator('[data-testid="claim-row"]')).toHaveCount(1)
    await expect(page.locator('body')).toContainText('CLM-2004')
  })

  test('shows "Showing N claims" count', async ({ page }) => {
    await expect(page.locator('body')).toContainText(/showing 5/i)
    await page.getByTestId('tab-settled').click()
    await expect(page.locator('body')).toContainText(/showing 1/i)
  })

  test('poll updates CLM-2003 status from "filed" to "investigating"', async ({ page }) => {
    // fetchClaims advances CLM-2003 on the 3rd+ call (≥ 10 s)
    await page.getByTestId('tab-investigating').click()
    // Initially only CLM-2001 (1 row)
    await expect(page.locator('[data-testid="claim-row"]')).toHaveCount(1)
    // Wait up to 15 s for the poll to fire twice more and surface the update
    await expect(page.locator('[data-testid="claim-row"]')).toHaveCount(2, { timeout: 15_000 })
    await expect(page.locator('body')).toContainText('CLM-2003')
  })

  // ── Part 3: detail panel and note submission ──────────────────────────────

  test('clicking a claim row (or View button) opens a detail panel', async ({ page }) => {
    await page.locator('[data-testid="claim-row"]').first().click()
    await expect(page.getByTestId('claim-detail').first()).toBeVisible()
  })

  test('detail panel shows policy ID', async ({ page }) => {
    await page.locator('[data-testid="claim-row"]').first().click()
    await expect(page.getByTestId('claim-detail').first()).toContainText('POL-101')
  })

  test('detail panel shows existing note text', async ({ page }) => {
    // CLM-2001 has a note: "Initial review complete. Requesting HR records."
    const clm2001Row = page.locator('[data-testid="claim-row"]').filter({ hasText: 'CLM-2001' })
    await clm2001Row.click()
    await expect(page.getByTestId('claim-detail').first()).toContainText('Initial review complete')
  })

  test('detail panel shows "No notes yet." when claim has no notes', async ({ page }) => {
    const clm2002Row = page.locator('[data-testid="claim-row"]').filter({ hasText: 'CLM-2002' })
    await clm2002Row.click()
    await expect(page.getByTestId('claim-detail').first()).toContainText(/no notes/i)
  })

  test('note textarea is present in the detail panel', async ({ page }) => {
    await page.locator('[data-testid="claim-row"]').first().click()
    await expect(page.getByTestId('note-textarea').first()).toBeVisible()
  })

  test('"Add Note" button is present in the detail panel', async ({ page }) => {
    await page.locator('[data-testid="claim-row"]').first().click()
    await expect(
      page.locator('button').filter({ hasText: /add note/i }).first()
    ).toBeVisible()
  })

  test('note-spinner appears while save is in-flight', async ({ page }) => {
    await page.locator('[data-testid="claim-row"]').first().click()
    await page.getByTestId('note-textarea').first().fill('Following up with legal.')
    await page.locator('button').filter({ hasText: /add note/i }).first().click()
    await expect(page.getByTestId('note-spinner').first()).toBeVisible({ timeout: 1_000 })
  })

  test('note textarea is disabled while save is in-flight', async ({ page }) => {
    await page.locator('[data-testid="claim-row"]').first().click()
    const textarea = page.getByTestId('note-textarea').first()
    await textarea.fill('Testing disable state.')
    await page.locator('button').filter({ hasText: /add note/i }).first().click()
    await expect(textarea).toBeDisabled()
  })

  test('clicking the same row again collapses the detail panel', async ({ page }) => {
    const row = page.locator('[data-testid="claim-row"]').first()
    await row.click()
    await expect(page.getByTestId('claim-detail').first()).toBeVisible()
    await row.click()
    await expect(page.getByTestId('claim-detail')).toHaveCount(0)
  })

  test('only one detail panel is open at a time', async ({ page }) => {
    const rows = page.locator('[data-testid="claim-row"]')
    await rows.nth(0).click()
    await rows.nth(1).click()
    await expect(page.getByTestId('claim-detail')).toHaveCount(1)
  })
})
