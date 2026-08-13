/**
 * Playwright tests for Problem 06 — Fundraising Class Leaderboard
 *
 * Run (from react/):
 *   PRACTICE_ANSWER=practice_problem_answers/cw_answer_06_fundraising_leaderboard npm run test:06
 *
 * These tests target a COMPLETED implementation of the problem stub.
 * Against the stub they will fail — that is expected behaviour.
 *
 * Seed data reference (from SEED_COHORTS in the problem file):
 *   cls-1985  "Class of 1985"  $22,150  47 donors  goal $25,000  (88%)
 *   cls-1995  "Class of 1995"  $18,900  62 donors  goal $20,000  (94%)
 *   cls-2010  "Class of 2010"  $11,200  54 donors  goal $10,000  (112% — over goal!)
 *   cls-2005  "Class of 2005"   $8,400  31 donors  goal $15,000  (56%)
 *   cls-2015  "Class of 2015"   $5,600  28 donors  goal  $8,000  (70%)
 *   cls-2020  "Class of 2020"   $2,100  19 donors  goal  $5,000  (42%)
 *
 * Default sort (amountRaised desc): 1985, 1995, 2010, 2005, 2015, 2020
 * By donorCount desc:               1995 (62), 2010 (54), 1985 (47), 2005 (31) …
 *
 * Query strategy (in priority order):
 *   1. Seed data text — toContainText('Class of 1985'), /22[,.]?150/ etc.
 *   2. Element type + visible text — locator('button').filter({ hasText: /pledge/i })
 *   3. Input attributes — getByPlaceholder(/search/i)
 *   4. data-testid — required for: cohort rows, pledge-input, pledge-spinner.
 *      Sort buttons also accept data-testid (see problem file for full list).
 */

import { test, expect } from '@playwright/test'

test.describe('Problem 06 — Fundraising Class Leaderboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  // ── Part 1: leaderboard display ───────────────────────────────────────────

  test('renders 6 cohort rows', async ({ page }) => {
    await expect(page.locator('[data-testid="cohort-row"]')).toHaveCount(6)
  })

  test('seed class names are visible', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Class of 1985')
    await expect(page.locator('body')).toContainText('Class of 2010')
    await expect(page.locator('body')).toContainText('Class of 2020')
  })

  test('lead volunteer names are visible', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Patricia Nguyen')
    await expect(page.locator('body')).toContainText('Tyler Brooks')
  })

  test('amounts raised are visible', async ({ page }) => {
    await expect(page.locator('body')).toContainText(/22[,.]?150/)
    await expect(page.locator('body')).toContainText(/2[,.]?100/)
  })

  test('default sort puts Class of 1985 first (highest raised)', async ({ page }) => {
    const firstRow = page.locator('[data-testid="cohort-row"]').first()
    await expect(firstRow).toContainText('1985')
  })

  test('default sort puts Class of 2020 last (lowest raised)', async ({ page }) => {
    const lastRow = page.locator('[data-testid="cohort-row"]').last()
    await expect(lastRow).toContainText('2020')
  })

  // ── Part 2: sorting and search ────────────────────────────────────────────

  test('sort-by-raised control is present', async ({ page }) => {
    await expect(
      page.locator('[data-testid="sort-raised"]').or(
        page.locator('button').filter({ hasText: /raised/i })
      ).first()
    ).toBeVisible()
  })

  test('sort-by-donors control is present', async ({ page }) => {
    await expect(
      page.locator('[data-testid="sort-donors"]').or(
        page.locator('button').filter({ hasText: /donor/i })
      ).first()
    ).toBeVisible()
  })

  test('search input is present', async ({ page }) => {
    await expect(
      page.locator('[data-testid="search-input"], input[placeholder*="search" i]').first()
    ).toBeVisible()
  })

  test('sort by donors puts Class of 1995 first (62 donors)', async ({ page }) => {
    const sortBtn = page.locator('[data-testid="sort-donors"]').or(
      page.locator('button').filter({ hasText: /donor/i })
    ).first()
    await sortBtn.click()
    const firstRow = page.locator('[data-testid="cohort-row"]').first()
    await expect(firstRow).toContainText('1995')
  })

  test('sort-raised toggle reverses row order', async ({ page }) => {
    const sortBtn = page.locator('[data-testid="sort-raised"]').or(
      page.locator('button').filter({ hasText: /raised/i })
    ).first()
    await sortBtn.click()
    const rowsBefore = await page.locator('[data-testid="cohort-row"]').allTextContents()
    await sortBtn.click()
    const rowsAfter = await page.locator('[data-testid="cohort-row"]').allTextContents()
    const changed = rowsBefore.some((r, i) => r !== rowsAfter[i])
    expect(changed).toBeTruthy()
  })

  test('search filters by class name (case-insensitive)', async ({ page }) => {
    const searchInput = page.locator('[data-testid="search-input"], input[placeholder*="search" i]').first()
    await searchInput.fill('2020')
    await expect(page.locator('[data-testid="cohort-row"]')).toHaveCount(1)
    await expect(page.locator('body')).toContainText('Class of 2020')
  })

  test('shows "Showing N classes"', async ({ page }) => {
    await expect(page.locator('body')).toContainText(/showing 6/i)
    const searchInput = page.locator('[data-testid="search-input"], input[placeholder*="search" i]').first()
    await searchInput.fill('199')
    await expect(page.locator('body')).toContainText(/showing 1/i)
  })

  // ── Part 3: optimistic pledge entry ──────────────────────────────────────

  test('"Pledge" button is visible on each row', async ({ page }) => {
    const pledgeBtns = page.locator('button').filter({ hasText: /^pledge$/i })
    await expect(pledgeBtns).toHaveCount(6)
  })

  test('clicking "Pledge" reveals the pledge input', async ({ page }) => {
    const pledgeBtn = page.locator('button').filter({ hasText: /^pledge$/i }).first()
    await pledgeBtn.click()
    await expect(page.getByTestId('pledge-input').first()).toBeVisible()
  })

  test('pledge-spinner appears while submit is in-flight', async ({ page }) => {
    const pledgeBtn = page.locator('button').filter({ hasText: /^pledge$/i }).first()
    await pledgeBtn.click()
    const input = page.getByTestId('pledge-input').first()
    await input.fill('500')
    const submitBtn = page.locator('button').filter({ hasText: /submit/i }).first()
    await submitBtn.click()
    await expect(page.getByTestId('pledge-spinner').first()).toBeVisible({ timeout: 1_000 })
  })

  test('submit button is disabled while pledge is in-flight', async ({ page }) => {
    const pledgeBtn = page.locator('button').filter({ hasText: /^pledge$/i }).first()
    await pledgeBtn.click()
    const input = page.getByTestId('pledge-input').first()
    await input.fill('250')
    const submitBtn = page.locator('button').filter({ hasText: /submit/i }).first()
    await submitBtn.click()
    await expect(submitBtn).toBeDisabled()
  })
})
