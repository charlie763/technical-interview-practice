/**
 * =============================================================================
 * INTERVIEW PROBLEM 04: Contract Review Dashboard
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the contract review dashboard for a contract lifecycle
 * management (CLM) platform. Legal ops teams use this dashboard to monitor
 * contract statuses, find contracts expiring soon, and make quick field edits
 * without leaving the list view.
 *
 * =============================================================================
 * PROVIDED (do not modify)
 * =============================================================================
 *
 * SEED_CONTRACTS    — 6 deterministic contracts used in Playwright tests.
 * STATUS_COLORS     — hex colour map for status badges.
 * updateContractField(contractId, field, value) → Promise
 *                   — mock API call; rejects ~15% of the time.
 *
 * =============================================================================
 * PART 1 — Contract table
 * =============================================================================
 *
 * Render a table (or list) of contracts from SEED_CONTRACTS.
 * Each row must include:
 *   • Contract title
 *   • Status badge (colour-coded using STATUS_COLORS)
 *   • Owner email
 *   • Expiration date
 *
 * Required data-testid attributes:
 *   • data-testid="contract-row"    on every contract row
 *   • data-testid="status-badge"    on the status badge inside each row
 *
 * =============================================================================
 * PART 2 — Filter, search, and sort
 * =============================================================================
 *
 * Add three controls above the table:
 *
 * 1. Status filter — a <select> that filters by contract status.
 *    Options: "All", "draft", "in_review", "approved", "active", "expired"
 *    data-testid="filter-status"
 *
 * 2. Text search — an <input> that filters rows by title (case-insensitive
 *    substring match).
 *    data-testid="search-input"
 *
 * 3. Sort by expiration — a <button> that toggles between ascending and
 *    descending expiration date order.
 *    data-testid="sort-expiration"
 *
 * Use useMemo to derive the filtered + sorted list from the raw contracts state.
 * Show a row count: "Showing N contracts".
 *
 * =============================================================================
 * PART 3 — Inline field editing with optimistic updates
 * =============================================================================
 *
 * Make the contract title and owner email cells inline-editable:
 *
 * • Double-clicking a title or email cell enters edit mode (renders an <input>).
 * • Pressing Enter or clicking a "Save" button calls updateContractField and
 *   shows a spinner (data-testid="save-spinner") while the request is in-flight.
 * • On success: update the cell with the new value.
 * • On failure: revert to the previous value and show an inline error message.
 * • While saving, the input should be disabled.
 *
 * Required data-testid attributes:
 *   • data-testid="editable-field"   on every editable cell (view mode)
 *   • data-testid="save-spinner"     on the loading indicator during save
 *
 * =============================================================================
 */

// ── Seed data (do not modify) ─────────────────────────────────────────────────

export const SEED_CONTRACTS = [
  {
    contract_id: 'con-001',
    title: 'Vendor MSA',
    owner_email: 'legal@acme.com',
    status: 'active',
    expires_on: '2025-09-30',
  },
  {
    contract_id: 'con-002',
    title: 'SaaS Subscription Agreement',
    owner_email: 'ops@acme.com',
    status: 'in_review',
    expires_on: '2025-12-31',
  },
  {
    contract_id: 'con-003',
    title: 'NDA — Design Partner',
    owner_email: 'bizdev@acme.com',
    status: 'approved',
    expires_on: '2026-03-15',
  },
  {
    contract_id: 'con-004',
    title: 'Office Lease',
    owner_email: 'finance@acme.com',
    status: 'active',
    expires_on: '2027-06-01',
  },
  {
    contract_id: 'con-005',
    title: 'Legacy Reseller Agreement',
    owner_email: 'sales@acme.com',
    status: 'expired',
    expires_on: '2024-01-01',
  },
  {
    contract_id: 'con-006',
    title: 'Marketing Agency SOW',
    owner_email: 'marketing@acme.com',
    status: 'draft',
    expires_on: '2025-11-30',
  },
]

export const STATUS_COLORS = {
  draft:     '#94a3b8',
  in_review: '#f59e0b',
  approved:  '#3b82f6',
  active:    '#22c55e',
  expired:   '#ef4444',
}

/**
 * Mock API call. Resolves after ~300 ms; rejects ~15% of the time.
 *
 * @param {string} contractId
 * @param {string} field        - "title" | "owner_email"
 * @param {string} value        - new value
 * @returns {Promise<{ contract_id: string, field: string, value: string }>}
 */
export function updateContractField(contractId, field, value) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      if (Math.random() < 0.15) {
        reject(new Error('Save failed — please try again.'))
      } else {
        resolve({ contract_id: contractId, field, value })
      }
    }, 300)
  })
}

// ── Your implementation goes below ───────────────────────────────────────────

export default function App() {
  throw new Error('Not implemented — replace this with your solution.')
}
