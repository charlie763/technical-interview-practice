/**
 * Playwright tests for Problem 07 — Underwriting Submission Queue
 *
 * Run (from react/):
 *   PRACTICE_ANSWER=practice_problem_answers/cw_answer_07_submission_queue npm run test:07
 *
 * These tests target a COMPLETED implementation of the problem stub.
 * Against the stub they will fail — that is expected behaviour.
 *
 * Seed data reference (from SEED_SUBMISSIONS in the problem file):
 *   sub-001  Meridian Staffing Group     epl        pending   $1,000,000
 *   sub-002  Hartwell Capital Partners   do         quoted    $5,000,000
 *   sub-003  Clearview Healthcare LLC    fiduciary  pending   $500,000
 *   sub-004  Pinnacle Tech Ventures      do         declined  $2,000,000
 *   sub-005  Blue Ridge Construction Co. epl        pending   $250,000
 *   sub-006  Solaris Property Management fiduciary  quoted    $1,500,000
 *
 * Query strategy (in priority order):
 *   1. Seed data text — toContainText('Meridian Staffing'), /500,000/ etc.
 *   2. Element type + visible text — locator('button').filter({ hasText: /issue quote/i })
 *   3. data-testid — submission rows, status badges, coverage badges,
 *      filter selects, sort button, quote spinner, premium input
 */

import { test, expect } from '@playwright/test'

test.describe('Problem 07 — Underwriting Submission Queue', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  // ── Part 1: submission table ──────────────────────────────────────────────

  test('renders 6 submission rows from seed data', async ({ page }) => {
    await expect(page.locator('[data-testid="submission-row"]')).toHaveCount(6)
  })

  test('seed company names are visible', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Meridian Staffing Group')
    await expect(page.locator('body')).toContainText('Hartwell Capital Partners')
    await expect(page.locator('body')).toContainText('Pinnacle Tech Ventures')
  })

  test('status badges are present on each row', async ({ page }) => {
    await expect(page.locator('[data-testid="status-badge"]')).toHaveCount(6)
  })

  test('coverage badges are present on each row', async ({ page }) => {
    await expect(page.locator('[data-testid="coverage-badge"]')).toHaveCount(6)
  })

  test('requested limits are visible', async ({ page }) => {
    // sub-001: $1,000,000  sub-003: $500,000  sub-005: $250,000
    await expect(page.locator('body')).toContainText(/1[,.]?000[,.]?000/)
    await expect(page.locator('body')).toContainText(/500[,.]?000/)
    await expect(page.locator('body')).toContainText(/250[,.]?000/)
  })

  // ── Part 2: filter and sort ───────────────────────────────────────────────

  test('coverage filter select is present', async ({ page }) => {
    await expect(page.getByTestId('filter-coverage')).toBeVisible()
  })

  test('status filter select is present', async ({ page }) => {
    await expect(page.getByTestId('filter-status')).toBeVisible()
  })

  test('filtering by coverage "epl" shows only EPL rows', async ({ page }) => {
    await page.getByTestId('filter-coverage').selectOption('epl')
    await expect(page.locator('[data-testid="submission-row"]')).toHaveCount(2)
    await expect(page.locator('body')).toContainText('Meridian Staffing Group')
    await expect(page.locator('body')).toContainText('Blue Ridge Construction')
  })

  test('filtering by coverage "do" shows only D&O rows', async ({ page }) => {
    await page.getByTestId('filter-coverage').selectOption('do')
    await expect(page.locator('[data-testid="submission-row"]')).toHaveCount(2)
  })

  test('filtering by coverage "fiduciary" shows only fiduciary rows', async ({ page }) => {
    await page.getByTestId('filter-coverage').selectOption('fiduciary')
    await expect(page.locator('[data-testid="submission-row"]')).toHaveCount(2)
  })

  test('filtering by status "pending" shows only pending rows', async ({ page }) => {
    await page.getByTestId('filter-status').selectOption('pending')
    await expect(page.locator('[data-testid="submission-row"]')).toHaveCount(3)
  })

  test('filtering by status "quoted" shows only quoted rows', async ({ page }) => {
    await page.getByTestId('filter-status').selectOption('quoted')
    await expect(page.locator('[data-testid="submission-row"]')).toHaveCount(2)
  })

  test('filtering by status "declined" shows only declined rows', async ({ page }) => {
    await page.getByTestId('filter-status').selectOption('declined')
    await expect(page.locator('[data-testid="submission-row"]')).toHaveCount(1)
    await expect(page.locator('body')).toContainText('Pinnacle Tech Ventures')
  })

  test('combining coverage and status filters works', async ({ page }) => {
    await page.getByTestId('filter-coverage').selectOption('epl')
    await page.getByTestId('filter-status').selectOption('pending')
    await expect(page.locator('[data-testid="submission-row"]')).toHaveCount(2)
  })

  test('shows "Showing N submissions" count', async ({ page }) => {
    await expect(page.locator('body')).toContainText(/showing 6/i)
    await page.getByTestId('filter-status').selectOption('pending')
    await expect(page.locator('body')).toContainText(/showing 3/i)
  })

  test('sort button is present', async ({ page }) => {
    const sortBtn = page.locator(
      '[data-testid="sort-limit"], button:has-text("limit")'
    ).first()
    await expect(sortBtn).toBeVisible()
  })

  test('sort by limit toggles order', async ({ page }) => {
    const sortBtn = page.locator(
      '[data-testid="sort-limit"], button:has-text("limit")'
    ).first()

    // First click: one direction
    await sortBtn.click()
    const rows1 = page.locator('[data-testid="submission-row"]')
    const first1 = await rows1.first().textContent()

    // Second click: reversed
    await sortBtn.click()
    const first2 = await rows1.first().textContent()

    expect(first1).not.toBe(first2)
  })

  // ── Part 3: issue quote inline form ──────────────────────────────────────

  test('"Issue Quote" button present on pending rows only', async ({ page }) => {
    // 3 pending rows
    const issueButtons = page.locator('button').filter({ hasText: /issue quote/i })
    await expect(issueButtons).toHaveCount(3)
  })

  test('no "Issue Quote" button on quoted or declined rows', async ({ page }) => {
    const quotedRow = page.locator('[data-testid="submission-row"]').filter({
      hasText: 'Hartwell Capital Partners',
    })
    await expect(quotedRow.locator('button').filter({ hasText: /issue quote/i })).toHaveCount(0)
  })

  test('clicking "Issue Quote" reveals premium input and Confirm button', async ({ page }) => {
    const btn = page.locator('button').filter({ hasText: /issue quote/i }).first()
    await btn.click()
    await expect(page.getByTestId('premium-input').first()).toBeVisible()
    await expect(page.locator('button').filter({ hasText: /confirm/i }).first()).toBeVisible()
  })

  test('"Cancel" button hides the form', async ({ page }) => {
    const btn = page.locator('button').filter({ hasText: /issue quote/i }).first()
    await btn.click()
    await page.locator('button').filter({ hasText: /cancel/i }).first().click()
    await expect(page.getByTestId('premium-input')).toHaveCount(0)
  })

  test('quote-spinner appears while save is in-flight', async ({ page }) => {
    const btn = page.locator('button').filter({ hasText: /issue quote/i }).first()
    await btn.click()
    await page.getByTestId('premium-input').first().fill('12500')
    await page.locator('button').filter({ hasText: /confirm/i }).first().click()
    await expect(page.getByTestId('quote-spinner').first()).toBeVisible({ timeout: 1_000 })
  })

  test('"Confirm" button is disabled while in-flight', async ({ page }) => {
    const btn = page.locator('button').filter({ hasText: /issue quote/i }).first()
    await btn.click()
    await page.getByTestId('premium-input').first().fill('9800')
    const confirmBtn = page.locator('button').filter({ hasText: /confirm/i }).first()
    await confirmBtn.click()
    await expect(confirmBtn).toBeDisabled()
  })
})
