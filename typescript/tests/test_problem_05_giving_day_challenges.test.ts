/**
 * Tests for Problem 05 — Giving Day Challenge Engine
 *
 * Run (from typescript/):
 *   PRACTICE_ANSWER=cw_answer_05_giving_day_challenges npm run test:05
 */

import { GivingDayEngine } from '../practice_problems/problem_05_giving_day_challenges'

const DEADLINE = '2025-10-01T12:00:00'

// ── Part 1 ─────────────────────────────────────────────────────────────────

describe('Part 1 — Challenge setup and donation recording', () => {
  let engine: GivingDayEngine

  beforeEach(() => {
    engine = new GivingDayEngine()
  })

  it('addChallenge returns challenge with correct fields', () => {
    const c = engine.addChallenge('c1', 'Morning Boost', 1000, DEADLINE, 'flat_bonus', 500)
    expect(c.id).toBe('c1')
    expect(c.name).toBe('Morning Boost')
    expect(c.thresholdAmount).toBe(1000)
    expect(c.deadline).toBe(DEADLINE)
    expect(c.bonusType).toBe('flat_bonus')
    expect(c.bonusValue).toBe(500)
    expect(c.completed).toBe(false)
    expect(c.completedAt).toBeNull()
  })

  it('addChallenge throws on duplicate id', () => {
    engine.addChallenge('c1', 'A', 500, DEADLINE, 'flat_bonus', 100)
    expect(() => engine.addChallenge('c1', 'B', 500, DEADLINE, 'match_pct', 10)).toThrow()
  })

  it('addChallenge throws on thresholdAmount <= 0', () => {
    expect(() => engine.addChallenge('c2', 'Bad', 0, DEADLINE, 'flat_bonus', 100)).toThrow()
    expect(() => engine.addChallenge('c3', 'Neg', -100, DEADLINE, 'flat_bonus', 100)).toThrow()
  })

  it('getChallenge returns the correct challenge', () => {
    engine.addChallenge('c1', 'Morning Boost', 1000, DEADLINE, 'flat_bonus', 500)
    const c = engine.getChallenge('c1')
    expect(c.id).toBe('c1')
  })

  it('getChallenge throws on unknown id', () => {
    expect(() => engine.getChallenge('no-such')).toThrow()
  })

  it('addDonation returns donation with correct fields', () => {
    const d = engine.addDonation(400, '2025-10-01T09:00:00')
    expect(d.amount).toBe(400)
    expect(d.timestamp).toBe('2025-10-01T09:00:00')
    expect(d.id).toBeTruthy()
  })

  it('addDonation auto-generates sequential IDs', () => {
    const a = engine.addDonation(100, '2025-10-01T08:00:00')
    const b = engine.addDonation(200, '2025-10-01T09:00:00')
    expect(a.id).not.toBe(b.id)
  })

  it('addDonation throws on amount <= 0', () => {
    expect(() => engine.addDonation(0, '2025-10-01T09:00:00')).toThrow()
    expect(() => engine.addDonation(-50, '2025-10-01T09:00:00')).toThrow()
  })

  it('getTotalRaised without asOf returns sum of all donations', () => {
    engine.addDonation(400, '2025-10-01T09:00:00')
    engine.addDonation(300, '2025-10-01T10:00:00')
    expect(engine.getTotalRaised()).toBe(700)
  })

  it('getTotalRaised with asOf only counts donations at or before asOf', () => {
    engine.addDonation(400, '2025-10-01T09:00:00')
    engine.addDonation(300, '2025-10-01T10:00:00')
    engine.addDonation(500, '2025-10-01T13:00:00')
    expect(engine.getTotalRaised('2025-10-01T12:00:00')).toBe(700)
  })

  it('getTotalRaised returns 0 when no donations recorded', () => {
    expect(engine.getTotalRaised()).toBe(0)
  })
})

// ── Part 2 ─────────────────────────────────────────────────────────────────

describe('Part 2 — processDonation and challenge completion', () => {
  let engine: GivingDayEngine

  beforeEach(() => {
    engine = new GivingDayEngine()
    // c1: threshold=1000, deadline=12:00 (morning challenge)
    engine.addChallenge('c1', 'Morning Boost', 1000, '2025-10-01T12:00:00', 'flat_bonus', 500)
    // c2: threshold=2000, deadline=18:00 (afternoon challenge)
    engine.addChallenge('c2', 'Afternoon Push', 2000, '2025-10-01T18:00:00', 'match_pct', 10)
  })

  it('processDonation returns the donation entry', () => {
    const result = engine.processDonation(400, '2025-10-01T09:00:00')
    expect(result.donation.amount).toBe(400)
    expect(result.donation.timestamp).toBe('2025-10-01T09:00:00')
  })

  it('processDonation returns empty newlyCompleted when threshold not met', () => {
    const result = engine.processDonation(400, '2025-10-01T09:00:00')
    expect(result.newlyCompleted).toHaveLength(0)
  })

  it('processDonation marks challenge complete when threshold reached', () => {
    engine.processDonation(400, '2025-10-01T09:00:00')
    engine.processDonation(300, '2025-10-01T10:00:00')
    const result = engine.processDonation(400, '2025-10-01T11:00:00')  // total=1100 >= 1000
    expect(result.newlyCompleted).toHaveLength(1)
    expect(result.newlyCompleted[0].id).toBe('c1')
  })

  it('processDonation sets completedAt to the donation timestamp', () => {
    engine.processDonation(500, '2025-10-01T09:00:00')
    engine.processDonation(600, '2025-10-01T10:00:00')  // total=1100, c1 completes
    const c1 = engine.getChallenge('c1')
    expect(c1.completed).toBe(true)
    expect(c1.completedAt).toBe('2025-10-01T10:00:00')
  })

  it('challenge only completes once (idempotent)', () => {
    engine.processDonation(600, '2025-10-01T09:00:00')
    const r1 = engine.processDonation(500, '2025-10-01T10:00:00')  // total=1100, c1 completes
    const r2 = engine.processDonation(200, '2025-10-01T11:00:00')  // c1 already done
    expect(r1.newlyCompleted.some((c: any) => c.id === 'c1')).toBe(true)
    expect(r2.newlyCompleted.some((c: any) => c.id === 'c1')).toBe(false)
  })

  it('donation after deadline does not count toward challenge', () => {
    // c1 deadline = 12:00; donation at 13:00 doesn't count for c1
    engine.processDonation(900, '2025-10-01T11:00:00')
    const result = engine.processDonation(500, '2025-10-01T13:00:00')  // after deadline
    // getTotalRaised(c1.deadline=12:00) = 900 < 1000 → c1 should NOT complete
    expect(result.newlyCompleted.some((c: any) => c.id === 'c1')).toBe(false)
  })

  it('multiple challenges can complete on the same donation', () => {
    // lower c2 threshold for this test by seeding engine fresh
    const e2 = new GivingDayEngine()
    e2.addChallenge('x1', 'X1', 500, '2025-10-01T12:00:00', 'flat_bonus', 100)
    e2.addChallenge('x2', 'X2', 500, '2025-10-01T18:00:00', 'flat_bonus', 200)
    const result = e2.processDonation(600, '2025-10-01T10:00:00')
    expect(result.newlyCompleted).toHaveLength(2)
  })
})

// ── Part 3 ─────────────────────────────────────────────────────────────────

describe('Part 3 — getHourlyBreakdown', () => {
  let engine: GivingDayEngine

  beforeEach(() => {
    engine = new GivingDayEngine()
    engine.addChallenge('c1', 'Morning Boost', 1000, '2025-10-01T12:00:00', 'flat_bonus', 500)
    engine.addChallenge('c2', 'Afternoon Push', 1800, '2025-10-01T18:00:00', 'match_pct', 10)

    // 09:xx — two donations totalling 700
    engine.processDonation(400, '2025-10-01T09:15:00')
    engine.processDonation(300, '2025-10-01T09:45:00')
    // 11:xx — one donation pushing c1 over the line (total=1100)
    engine.processDonation(400, '2025-10-01T11:30:00')
    // 14:xx — one donation (total=1500)
    engine.processDonation(500, '2025-10-01T14:00:00')
    // 17:xx — one donation pushing c2 over the line (total=2000)
    engine.processDonation(500, '2025-10-01T17:00:00')
  })

  it('getHourlyBreakdown returns one bucket per active hour', () => {
    const buckets = engine.getHourlyBreakdown()
    const hours = buckets.map((b: any) => b.hour)
    expect(hours).toContain('2025-10-01T09:00:00')
    expect(hours).toContain('2025-10-01T11:00:00')
    expect(hours).toContain('2025-10-01T14:00:00')
    expect(hours).toContain('2025-10-01T17:00:00')
    // no 10:xx, 12:xx, etc.
    expect(hours).not.toContain('2025-10-01T10:00:00')
  })

  it('getHourlyBreakdown is ordered by hour ascending', () => {
    const buckets = engine.getHourlyBreakdown()
    const hours = buckets.map((b: any) => b.hour)
    expect(hours).toEqual([...hours].sort())
  })

  it('getHourlyBreakdown bucket has correct donationCount and amountRaised', () => {
    const buckets = engine.getHourlyBreakdown()
    const nine = buckets.find((b: any) => b.hour === '2025-10-01T09:00:00')
    expect(nine.donationCount).toBe(2)
    expect(nine.amountRaised).toBe(700)
  })

  it('getHourlyBreakdown cumulativeTotal is running sum', () => {
    const buckets = engine.getHourlyBreakdown()
    const nine  = buckets.find((b: any) => b.hour === '2025-10-01T09:00:00')
    const eleven = buckets.find((b: any) => b.hour === '2025-10-01T11:00:00')
    const fourteen = buckets.find((b: any) => b.hour === '2025-10-01T14:00:00')
    expect(nine.cumulativeTotal).toBe(700)
    expect(eleven.cumulativeTotal).toBe(1100)
    expect(fourteen.cumulativeTotal).toBe(1600)
  })

  it('getHourlyBreakdown challengesCompleted lists IDs that completed in that hour', () => {
    const buckets = engine.getHourlyBreakdown()
    const eleven = buckets.find((b: any) => b.hour === '2025-10-01T11:00:00')
    const seventeen = buckets.find((b: any) => b.hour === '2025-10-01T17:00:00')
    expect(eleven.challengesCompleted).toContain('c1')
    expect(seventeen.challengesCompleted).toContain('c2')
  })

  it('getHourlyBreakdown hours with no completed challenges have empty array', () => {
    const buckets = engine.getHourlyBreakdown()
    const nine = buckets.find((b: any) => b.hour === '2025-10-01T09:00:00')
    expect(nine.challengesCompleted).toHaveLength(0)
  })

  it('getHourlyBreakdown returns empty array when no donations recorded', () => {
    const e2 = new GivingDayEngine()
    expect(e2.getHourlyBreakdown()).toHaveLength(0)
  })
})
