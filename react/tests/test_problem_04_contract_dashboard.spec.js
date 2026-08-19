/**
 * Playwright tests for Problem 04 — Contract Review Dashboard
 *
 * Run (from react/):
 *   npm run test:04
 *
 * These tests target a COMPLETED implementation of src/App.jsx.
 * Against the stub they will fail — that is expected behaviour.
 *
 * Seed data reference (from SEED_CONTRACTS in the problem file):
 *   con-001  "Vendor MSA"                 active     expires 2025-09-30
 *   con-002  "SaaS Subscription Agreement" in_review  expires 2025-12-31
 *   con-003  "NDA — Design Partner"        approved   expires 2026-03-15
 *   con-004  "Office Lease"                active     expires 2027-06-01
 *   con-005  "Legacy Reseller Agreement"   expired    expires 2024-01-01
 *   con-006  "Marketing Agency SOW"        draft      expires 2025-11-30
 *
 * Query strategy (in priority order):
 *   1. Seed data text — toContainText('Vendor MSA') etc.
 *   2. Input attributes — getByPlaceholder() for search input
 *   3. Element type + text — locator('button').filter({ hasText: /expir/i })
 *      for the sort button (data-testid="sort-expiration" also accepted)
 *   4. data-testid — required for row counting, badges, filter selects,
 *      editable fields, and the save spinner
 *      (see problem file for which testids are required)
 */

import { test, expect } from '@playwright/test'

test.describe('Problem 04 — Contract Review Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  // ── Part 1: contract table ─────────────────────────────────────────────────

  test('renders 6 contract rows from seed data', async ({ page }) => {
    await expect(page.locator('[data-testid="contract-row"]')).toHaveCount(6)
  })

  test('each row has a status badge', async ({ page }) => {
    const badges = page.locator('[data-testid="status-badge"]')
    await expect(badges).toHaveCount(6)
  })

  test('seed contract titles are visible', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Vendor MSA')
    await expect(page.locator('body')).toContainText('Office Lease')
    await expect(page.locator('body')).toContainText('Legacy Reseller Agreement')
  })

  test('expiration dates are visible', async ({ page }) => {
    await expect(page.locator('body')).toContainText('2025-09-30')
  })

  // ── Part 2: filter, search, sort ──────────────────────────────────────────

  test('filter-status select is present', async ({ page }) => {
    await expect(page.getByTestId('filter-status')).toBeVisible()
  })

  test('search input is present', async ({ page }) => {
    // Accepts data-testid="search-input" OR an input with a placeholder matching /search/i
    await expect(
      page.locator('[data-testid="search-input"], input[placeholder*="search" i]').first()
    ).toBeVisible()
  })

  test('sort by expiration button is present', async ({ page }) => {
    // Accepts data-testid="sort-expiration" OR a button whose text contains "expir"
    await expect(
      page.locator('[data-testid="sort-expiration"]').or(
        page.locator('button').filter({ hasText: /expir/i })
      ).first()
    ).toBeVisible()
  })

  test('filtering by status "expired" shows only expired contracts', async ({ page }) => {
    await page.getByTestId('filter-status').selectOption('expired')
    await expect(page.locator('[data-testid="contract-row"]')).toHaveCount(1)
    await expect(page.locator('body')).toContainText('Legacy Reseller Agreement')
  })

  test('filtering by status "active" shows only active contracts', async ({ page }) => {
    await page.getByTestId('filter-status').selectOption('active')
    await expect(page.locator('[data-testid="contract-row"]')).toHaveCount(2)
  })

  test('text search filters by title (case-insensitive)', async ({ page }) => {
    const searchInput = page.locator('[data-testid="search-input"], input[placeholder*="search" i]').first()
    await searchInput.fill('nda')
    await expect(page.locator('[data-testid="contract-row"]')).toHaveCount(1)
    await expect(page.locator('body')).toContainText('NDA — Design Partner')
  })

  test('search combined with status filter narrows results', async ({ page }) => {
    await page.getByTestId('filter-status').selectOption('active')
    const searchInput = page.locator('[data-testid="search-input"], input[placeholder*="search" i]').first()
    await searchInput.fill('vendor')
    await expect(page.locator('[data-testid="contract-row"]')).toHaveCount(1)
  })

  test('row count shows "Showing N contracts"', async ({ page }) => {
    await expect(page.locator('body')).toContainText('Showing 6')
    await page.getByTestId('filter-status').selectOption('expired')
    await expect(page.locator('body')).toContainText('Showing 1')
  })

  test('sort button changes expiration order', async ({ page }) => {
    // Click once to sort — verify row order changes
    const rowsBefore = await page.locator('[data-testid="contract-row"]').allTextContents()
    const sortBtn = page.locator('[data-testid="sort-expiration"]').or(
      page.locator('button').filter({ hasText: /expir/i })
    ).first()
    await sortBtn.click()
    const rowsAfter = await page.locator('[data-testid="contract-row"]').allTextContents()
    // Order should differ (seed data has varied dates)
    const changed = rowsBefore.some((r, i) => r !== rowsAfter[i])
    expect(changed).toBeTruthy()
  })

  // ── Part 3: inline editing ────────────────────────────────────────────────

  test('editable-field elements are present on rows', async ({ page }) => {
    await expect(page.locator('[data-testid="editable-field"]').first()).toBeVisible()
  })

  test('double-clicking an editable field activates edit mode', async ({ page }) => {
    const field = page.locator('[data-testid="editable-field"]').first()
    await field.dblclick()
    // An input should now be visible within the row
    await expect(page.locator('input').first()).toBeVisible()
  })

  test('save-spinner appears while save is in-flight', async ({ page }) => {
    const field = page.locator('[data-testid="editable-field"]').first()
    await field.dblclick()
    const input = page.locator('input').first()
    await input.fill('Updated Title')
    await input.press('Enter')
    // Spinner should appear briefly
    await expect(page.getByTestId('save-spinner').first()).toBeVisible({ timeout: 1_000 })
  })

  test('input is disabled while save is in-flight', async ({ page }) => {
    const field = page.locator('[data-testid="editable-field"]').first()
    await field.dblclick()
    const input = page.locator('input').first()
    await input.fill('Saving Now')
    await input.press('Enter')
    await expect(input).toBeDisabled()
  })
})
