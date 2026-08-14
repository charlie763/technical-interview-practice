/**
 * Tests for Problem 03 — Donor Communication Suppressor
 *
 * Run (from typescript/):
 *   PRACTICE_ANSWER=cw_answer_03_donor_suppressor npm run test:03
 */

import { DonorSuppressor } from '../practice_problems/problem_03_donor_suppressor'

const AS_OF = '2025-08-20T00:00:00'
const CAMPAIGN = 'camp-fall-2025'

// ── Part 1 ─────────────────────────────────────────────────────────────────

describe('Part 1 — Opt-out management and basic suppression', () => {
  let s: DonorSuppressor

  beforeEach(() => {
    s = new DonorSuppressor()
  })

  it('optOut adds email to opt-out list', () => {
    s.optOut('alice@example.com')
    expect(s.isOptedOut('alice@example.com')).toBe(true)
  })

  it('optOut is idempotent', () => {
    s.optOut('alice@example.com')
    s.optOut('alice@example.com')
    expect(s.isOptedOut('alice@example.com')).toBe(true)
  })

  it('optIn removes email from opt-out list', () => {
    s.optOut('alice@example.com')
    s.optIn('alice@example.com')
    expect(s.isOptedOut('alice@example.com')).toBe(false)
  })

  it('optIn is idempotent for emails not opted out', () => {
    expect(() => s.optIn('nobody@example.com')).not.toThrow()
    expect(s.isOptedOut('nobody@example.com')).toBe(false)
  })

  it('isOptedOut returns false for unknown email', () => {
    expect(s.isOptedOut('unknown@example.com')).toBe(false)
  })

  it('checkSuppression suppresses opted-out donors', () => {
    s.optOut('alice@example.com')
    const result = s.checkSuppression('alice@example.com', CAMPAIGN, AS_OF)
    expect(result.email).toBe('alice@example.com')
    expect(result.suppressed).toBe(true)
    expect(result.reason).toBe('opted_out')
  })

  it('checkSuppression returns not-suppressed for eligible donors', () => {
    const result = s.checkSuppression('carol@example.com', CAMPAIGN, AS_OF)
    expect(result.suppressed).toBe(false)
    expect(result.reason).toBeNull()
  })
})

// ── Part 2 ─────────────────────────────────────────────────────────────────

describe('Part 2 — Donation-based suppression rules', () => {
  let s: DonorSuppressor

  beforeEach(() => {
    s = new DonorSuppressor()
    s.optOut('alice@example.com')
    // bob donated to the target campaign recently
    s.recordDonation('bob@example.com', CAMPAIGN, '2025-08-01T10:00:00')
    // carol donated to a DIFFERENT campaign recently
    s.recordDonation('carol@example.com', 'camp-spring-2025', '2025-07-25T09:00:00')
    // dave donated to a different campaign but a long time ago
    s.recordDonation('dave@example.com', 'camp-spring-2025', '2025-06-01T09:00:00')
  })

  // already_donated
  it('checkSuppression suppresses donor who already gave to this campaign', () => {
    const result = s.checkSuppression('bob@example.com', CAMPAIGN, AS_OF)
    expect(result.suppressed).toBe(true)
    expect(result.reason).toBe('already_donated')
  })

  it('already_donated does not apply to a different campaign', () => {
    const result = s.checkSuppression('bob@example.com', 'camp-other', AS_OF)
    // no cooldown set, no opt-out → should be eligible
    expect(result.suppressed).toBe(false)
  })

  // recency_cooldown
  it('checkSuppression suppresses by recency_cooldown when cooldownDays is set', () => {
    // carol gave to camp-spring on 2025-07-25; asOf=2025-08-20; 26 days ago → within 30-day window
    const result = s.checkSuppression('carol@example.com', CAMPAIGN, AS_OF, { cooldownDays: 30 })
    expect(result.suppressed).toBe(true)
    expect(result.reason).toBe('recency_cooldown')
  })

  it('recency_cooldown does not suppress donations older than cooldownDays', () => {
    // dave gave on 2025-06-01; asOf=2025-08-20; 80 days ago → outside 30-day window
    const result = s.checkSuppression('dave@example.com', CAMPAIGN, AS_OF, { cooldownDays: 30 })
    expect(result.suppressed).toBe(false)
  })

  // priority: opted_out > already_donated > recency_cooldown
  it('opted_out takes priority over already_donated', () => {
    s.recordDonation('alice@example.com', CAMPAIGN, '2025-08-01T00:00:00')
    const result = s.checkSuppression('alice@example.com', CAMPAIGN, AS_OF, { cooldownDays: 30 })
    expect(result.reason).toBe('opted_out')
  })

  it('already_donated takes priority over recency_cooldown', () => {
    // bob already donated to CAMPAIGN and also within cooldown window
    const result = s.checkSuppression('bob@example.com', CAMPAIGN, AS_OF, { cooldownDays: 30 })
    expect(result.reason).toBe('already_donated')
  })
})

// ── Part 3 ─────────────────────────────────────────────────────────────────

describe('Part 3 — Bulk campaign filtering and summary', () => {
  let s: DonorSuppressor
  const emails = [
    'alice@example.com',   // opted out
    'bob@example.com',     // donated to this campaign
    'carol@example.com',   // recent cooldown
    'dave@example.com',    // old donation → eligible
    'eve@example.com',     // no history → eligible
  ]

  beforeEach(() => {
    s = new DonorSuppressor()
    s.optOut('alice@example.com')
    s.recordDonation('bob@example.com', CAMPAIGN, '2025-08-01T00:00:00')
    s.recordDonation('carol@example.com', 'camp-other', '2025-08-10T00:00:00')
    s.recordDonation('dave@example.com', 'camp-other', '2025-06-01T00:00:00')
  })

  it('filterCampaignList separates eligible from suppressed', () => {
    const result = s.filterCampaignList(emails, CAMPAIGN, AS_OF, { cooldownDays: 30 })
    expect(result.eligible).toContain('dave@example.com')
    expect(result.eligible).toContain('eve@example.com')
    expect(result.suppressed.some(r => r.email === 'alice@example.com')).toBe(true)
    expect(result.suppressed.some(r => r.email === 'bob@example.com')).toBe(true)
    expect(result.suppressed.some(r => r.email === 'carol@example.com')).toBe(true)
  })

  it('filterCampaignList eligible + suppressed covers all emails', () => {
    const result = s.filterCampaignList(emails, CAMPAIGN, AS_OF, { cooldownDays: 30 })
    const allEmails = [...result.eligible, ...result.suppressed.map(r => r.email)]
    expect(allEmails.sort()).toEqual([...emails].sort())
  })

  it('getSuppressedSummary returns correct counts', () => {
    const summary = s.getSuppressedSummary(emails, CAMPAIGN, AS_OF, { cooldownDays: 30 })
    expect(summary.total).toBe(5)
    expect(summary.eligibleCount).toBe(2)   // dave + eve
    expect(summary.suppressedCount).toBe(3) // alice + bob + carol
  })

  it('getSuppressedSummary breaks down by reason', () => {
    const summary = s.getSuppressedSummary(emails, CAMPAIGN, AS_OF, { cooldownDays: 30 })
    expect(summary.byReason['opted_out']).toBe(1)
    expect(summary.byReason['already_donated']).toBe(1)
    expect(summary.byReason['recency_cooldown']).toBe(1)
  })
})
