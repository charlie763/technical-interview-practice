/**
 * Playwright tests for Problem 05 — Campaign Fundraising Dashboard
 *
 * Run (from react/):
 *   PRACTICE_ANSWER=practice_problem_answers/cw_answer_05_campaign_dashboard npm run test:05
 *
 * These tests target a COMPLETED implementation of the problem stub.
 * Against the stub they will fail — that is expected behaviour.
 *
 * Seed data reference (from SEED_DONATIONS in the problem file):
 *   don-s1  Margaret Chen        $1,000  alumni  thanksSent: false
 *   don-s2  The Patel Family       $500  family  thanksSent: false
 *   don-s3  Robert Kim             $250  staff   thanksSent: false
 *   don-s4  Westfield School Board $5,000 board  thanksSent: true
 *   don-s5  Sarah Williams         $150  family  thanksSent: false
 *   don-s6  James Park           $2,500  alumni  thanksSent: false
 *
 * Total raised from seed: $9,400  |  Campaign goal: $50,000
 *
 * Query strategy (in priority order):
 *   1. Seed data text — toContainText('Margaret Chen'), /9[,.]?400/ etc.
 *   2. Element type + visible text — locator('button').filter({ hasText: /send thanks/i })
 *   3. Input attributes — getByPlaceholder(/search/i)
 *   4. data-testid — required for: donation rows, progress bar, filter-type
 *      select, and thanks-spinner (see problem file for full list)
 */

import { test, expect } from '@playwright/test'

test.describe('Problem 05 — Campaign Fundraising Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  // ── Part 1: campaign header & donation feed ───────────────────────────────

  test('renders campaign name', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Westfield Academy Annual Fund 2026')
  })

  test('renders 6 donation rows from seed data', async ({ page }) => {
    await expect(page.locator('[data-testid="donation-row"]')).toHaveCount(6)
  })

  test('seed donor names are visible', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Margaret Chen')
    await expect(page.locator('body')).toContainText('James Park')
    await expect(page.locator('body')).toContainText('Westfield School Board')
  })

  test('progress bar is present', async ({ page }) => {
    await expect(page.getByTestId('progress-bar')).toBeVisible()
  })

  test('total raised from seed data is visible', async ({ page }) => {
    // Seed total $9,400 — any format containing "9,400" or "9400" is acceptable
    await expect(page.locator('body')).toContainText(/9[,.]?400/)
  })

  test('donor count is visible', async ({ page }) => {
    // 6 donors from seed data
    await expect(page.locator('body')).toContainText(/6.{0,12}donor/i)
  })

  // ── Part 2: filter and search ─────────────────────────────────────────────

  test('filter-type select is present', async ({ page }) => {
    await expect(page.getByTestId('filter-type')).toBeVisible()
  })

  test('search input is present', async ({ page }) => {
    await expect(
      page.locator('[data-testid="search-input"], input[placeholder*="search" i]').first()
    ).toBeVisible()
  })

  test('filtering by "alumni" shows only alumni donors', async ({ page }) => {
    await page.getByTestId('filter-type').selectOption('alumni')
    await expect(page.locator('[data-testid="donation-row"]')).toHaveCount(2)
    await expect(page.locator('body')).toContainText('Margaret Chen')
    await expect(page.locator('body')).toContainText('James Park')
  })

  test('filtering by "family" shows only family donors', async ({ page }) => {
    await page.getByTestId('filter-type').selectOption('family')
    await expect(page.locator('[data-testid="donation-row"]')).toHaveCount(2)
  })

  test('filtering by "staff" shows only staff donors', async ({ page }) => {
    await page.getByTestId('filter-type').selectOption('staff')
    await expect(page.locator('[data-testid="donation-row"]')).toHaveCount(1)
    await expect(page.locator('body')).toContainText('Robert Kim')
  })

  test('search filters by donor name (case-insensitive)', async ({ page }) => {
    const searchInput = page.locator('[data-testid="search-input"], input[placeholder*="search" i]').first()
    await searchInput.fill('patel')
    await expect(page.locator('[data-testid="donation-row"]')).toHaveCount(1)
    await expect(page.locator('body')).toContainText('The Patel Family')
  })

  test('filter and search can be combined', async ({ page }) => {
    await page.getByTestId('filter-type').selectOption('alumni')
    const searchInput = page.locator('[data-testid="search-input"], input[placeholder*="search" i]').first()
    await searchInput.fill('james')
    await expect(page.locator('[data-testid="donation-row"]')).toHaveCount(1)
  })

  test('shows "Showing N donations" count', async ({ page }) => {
    await expect(page.locator('body')).toContainText(/showing 6/i)
    await page.getByTestId('filter-type').selectOption('staff')
    await expect(page.locator('body')).toContainText(/showing 1/i)
  })

  // ── Part 3: send thanks ───────────────────────────────────────────────────

  test('"Send Thanks" button present on rows where thanksSent is false', async ({ page }) => {
    // 5 of the 6 seed donations have thanksSent: false
    const thanksButtons = page.locator('button').filter({ hasText: /send thanks/i })
    await expect(thanksButtons).toHaveCount(5)
  })

  test('no "Send Thanks" button on the already-thanked donation', async ({ page }) => {
    // don-s4 Westfield School Board has thanksSent: true — no button on that row
    const boardRow = page.locator('[data-testid="donation-row"]').filter({
      hasText: 'Westfield School Board',
    })
    await expect(boardRow.locator('button').filter({ hasText: /send thanks/i })).toHaveCount(0)
  })

  test('thanks-spinner appears while send is in-flight', async ({ page }) => {
    const btn = page.locator('button').filter({ hasText: /send thanks/i }).first()
    await btn.click()
    await expect(page.getByTestId('thanks-spinner').first()).toBeVisible({ timeout: 1_000 })
  })

  test('"Send Thanks" button is disabled while in-flight', async ({ page }) => {
    const btn = page.locator('button').filter({ hasText: /send thanks/i }).first()
    await btn.click()
    await expect(btn).toBeDisabled()
  })
})
