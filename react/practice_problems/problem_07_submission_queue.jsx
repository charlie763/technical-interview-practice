/**
 * =============================================================================
 * INTERVIEW PROBLEM 07: Underwriting Submission Queue
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the underwriting queue for a commercial insurance platform.
 * Brokers submit coverage applications on behalf of their clients; underwriters
 * use this dashboard to review pending submissions, filter by coverage line,
 * and issue quotes without leaving the list view.
 *
 * =============================================================================
 * PROVIDED (do not modify)
 * =============================================================================
 *
 * SEED_SUBMISSIONS  — 6 deterministic submissions used in Playwright tests.
 * COVERAGE_COLORS   — hex colour map for coverage type badges.
 * STATUS_COLORS     — hex colour map for status badges.
 * issueQuote(submissionId, premium) → Promise
 *                   — mock API call; rejects ~15% of the time.
 *
 * =============================================================================
 * PART 1 — Submission table
 * =============================================================================
 *
 * Render a table (or list) of submissions from SEED_SUBMISSIONS.
 * Each row must include:
 *   • Company name
 *   • Coverage type badge (colour-coded using COVERAGE_COLORS)
 *   • Broker name
 *   • Submitted date (formatted as "Jan 5, 2025" or similar)
 *   • Status badge (colour-coded using STATUS_COLORS)
 *   • Requested limit (formatted as "$500,000" or similar)
 *
 * Required data-testid attributes:
 *   • data-testid="submission-row"   on every submission row
 *   • data-testid="status-badge"     on the status badge inside each row
 *   • data-testid="coverage-badge"   on the coverage type badge inside each row
 *
 * =============================================================================
 * PART 2 — Filter and sort
 * =============================================================================
 *
 * Add three controls above the table:
 *
 * 1. Coverage type filter — a <select> filtering by coverage_type.
 *    Options: "All", "epl", "do", "fiduciary"
 *    data-testid="filter-coverage"
 *
 * 2. Status filter — a <select> filtering by status.
 *    Options: "All", "pending", "quoted", "declined"
 *    data-testid="filter-status"
 *
 * 3. Sort by requested limit — a <button> toggling ascending / descending order.
 *    Tests accept either data-testid="sort-limit" OR a button whose visible
 *    text contains "limit" (case-insensitive).
 *
 * Use useMemo to derive the filtered + sorted list.
 * Show a row count: "Showing N submissions".
 *
 * =============================================================================
 * PART 3 — Inline "Issue Quote" with premium input
 * =============================================================================
 *
 * For every row whose status is "pending":
 *   • Show an "Issue Quote" button.
 *   • Clicking it reveals an inline input for the premium amount and a
 *     "Confirm" button (and a "Cancel" button to dismiss without saving).
 *   • Pressing "Confirm" calls issueQuote(submissionId, premiumValue) and shows
 *     a spinner (data-testid="quote-spinner") while the request is in-flight.
 *   • Optimistic update: immediately set status to "quoted" in the UI.
 *   • On success: keep the quoted state and hide the input form.
 *   • On failure: revert status to "pending" and show an inline error message.
 *   • While saving, the input and Confirm button are disabled.
 *
 * Required data-testid attributes:
 *   • data-testid="quote-spinner"    on the loading indicator during save
 *   • data-testid="premium-input"    on the premium amount input field
 *
 * TEST CONTRACT
 * -------------
 * Playwright tests use the following query strategy (in priority order):
 *
 * 1. Seed data text — toContainText('Meridian Staffing'), toContainText('500,000') etc.
 * 2. Element type + visible text — locator('button').filter({ hasText: /issue quote/i })
 * 3. data-testid — required for: submission rows, status badges, coverage badges,
 *    filter selects, sort button, quote spinner, premium input.
 *
 * =============================================================================
 */

// ── Seed data (do not modify) ─────────────────────────────────────────────────

export const SEED_SUBMISSIONS = [
  {
    submission_id: 'sub-001',
    company_name: 'Meridian Staffing Group',
    broker_name: 'Allied Risk Advisors',
    coverage_type: 'epl',
    submitted_on: '2025-06-01',
    status: 'pending',
    requested_limit: 1_000_000,
  },
  {
    submission_id: 'sub-002',
    company_name: 'Hartwell Capital Partners',
    broker_name: 'Summit Insurance Group',
    coverage_type: 'do',
    submitted_on: '2025-06-03',
    status: 'quoted',
    requested_limit: 5_000_000,
    quoted_premium: 42_500,
  },
  {
    submission_id: 'sub-003',
    company_name: 'Clearview Healthcare LLC',
    broker_name: 'Allied Risk Advisors',
    coverage_type: 'fiduciary',
    submitted_on: '2025-06-05',
    status: 'pending',
    requested_limit: 500_000,
  },
  {
    submission_id: 'sub-004',
    company_name: 'Pinnacle Tech Ventures',
    broker_name: 'Beacon Specialty Brokers',
    coverage_type: 'do',
    submitted_on: '2025-06-07',
    status: 'declined',
    requested_limit: 2_000_000,
  },
  {
    submission_id: 'sub-005',
    company_name: 'Blue Ridge Construction Co.',
    broker_name: 'Summit Insurance Group',
    coverage_type: 'epl',
    submitted_on: '2025-06-09',
    status: 'pending',
    requested_limit: 250_000,
  },
  {
    submission_id: 'sub-006',
    company_name: 'Solaris Property Management',
    broker_name: 'Beacon Specialty Brokers',
    coverage_type: 'fiduciary',
    submitted_on: '2025-06-10',
    status: 'quoted',
    requested_limit: 1_500_000,
    quoted_premium: 18_000,
  },
]

export const COVERAGE_COLORS = {
  epl:       '#3b82f6',   // blue
  do:        '#8b5cf6',   // purple
  fiduciary: '#f59e0b',   // amber
}

export const STATUS_COLORS = {
  pending:  '#94a3b8',   // slate
  quoted:   '#22c55e',   // green
  declined: '#ef4444',   // red
}

/**
 * Mock API call. Resolves after ~350 ms; rejects ~15% of the time.
 *
 * @param {string} submissionId
 * @param {number} premium  - quoted premium in dollars
 * @returns {Promise<{ submission_id: string, premium: number }>}
 */
export function issueQuote(submissionId, premium) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      if (Math.random() < 0.15) {
        reject(new Error('Quote failed — please try again.'))
      } else {
        resolve({ submission_id: submissionId, premium })
      }
    }, 350)
  })
}

// ── Your implementation goes below ───────────────────────────────────────────

export default function App() {
  throw new Error('Not implemented — replace this with your solution.')
}
