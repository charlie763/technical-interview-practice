/**
 * Tests for Problem 04 — Platform Fee Calculator
 *
 * Run (from typescript/):
 *   PRACTICE_ANSWER=cw_answer_04_fee_calculator npm run test:04
 */

import { FeeCalculator } from '../practice_problems/problem_04_fee_calculator'

// ── Part 1 ─────────────────────────────────────────────────────────────────

describe('Part 1 — calculateFees', () => {
  let calc: FeeCalculator

  beforeEach(() => {
    calc = new FeeCalculator()
  })

  it('calculateFees returns correct breakdown for card', () => {
    // gross=100, platform=3.50, processing=2.90+0.30=3.20, total=6.70, school=93.30
    const f = calc.calculateFees(100, 'card')
    expect(f.gross).toBe(100)
    expect(f.platformFee).toBe(3.50)
    expect(f.processingFee).toBe(3.20)
    expect(f.totalFees).toBe(6.70)
    expect(f.schoolReceives).toBe(93.30)
  })

  it('calculateFees returns correct breakdown for ach (below cap)', () => {
    // gross=100, platform=3.50, processing=0.80 (100×0.008=0.80 < 5), total=4.30, school=95.70
    const f = calc.calculateFees(100, 'ach')
    expect(f.gross).toBe(100)
    expect(f.platformFee).toBe(3.50)
    expect(f.processingFee).toBe(0.80)
    expect(f.totalFees).toBe(4.30)
    expect(f.schoolReceives).toBe(95.70)
  })

  it('calculateFees caps ach processingFee at $5.00', () => {
    // gross=1000, platform=35.00, processing=min(8.00,5.00)=5.00, total=40.00, school=960.00
    const f = calc.calculateFees(1000, 'ach')
    expect(f.platformFee).toBe(35.00)
    expect(f.processingFee).toBe(5.00)
    expect(f.totalFees).toBe(40.00)
    expect(f.schoolReceives).toBe(960.00)
  })

  it('calculateFees returns correct breakdown for check', () => {
    // gross=200, platform=7.00, processing=0, total=7.00, school=193.00
    const f = calc.calculateFees(200, 'check')
    expect(f.gross).toBe(200)
    expect(f.platformFee).toBe(7.00)
    expect(f.processingFee).toBe(0)
    expect(f.totalFees).toBe(7.00)
    expect(f.schoolReceives).toBe(193.00)
  })

  it('calculateFees rounds to nearest cent', () => {
    // gross=333, card: platform=11.655→11.66, processing=333×0.029+0.30=9.957→9.96
    const f = calc.calculateFees(333, 'card')
    expect(f.platformFee).toBe(11.66)
    expect(f.processingFee).toBe(9.96)
    expect(f.totalFees).toBe(21.62)
    expect(f.schoolReceives).toBe(311.38)
  })

  it('calculateFees throws on zero gross', () => {
    expect(() => calc.calculateFees(0, 'card')).toThrow()
  })

  it('calculateFees throws on negative gross', () => {
    expect(() => calc.calculateFees(-10, 'ach')).toThrow()
  })
})

// ── Part 2 ─────────────────────────────────────────────────────────────────

describe('Part 2 — grossForNet', () => {
  let calc: FeeCalculator

  beforeEach(() => {
    calc = new FeeCalculator()
  })

  it('grossForNet card: school receives at least netTarget (±$0.02)', () => {
    const f = calc.grossForNet(100, 'card')
    expect(f.schoolReceives).toBeGreaterThanOrEqual(99.98)
    expect(f.schoolReceives).toBeLessThanOrEqual(100.02)
  })

  it('grossForNet card: gross is ~107.48 for netTarget=100', () => {
    const f = calc.grossForNet(100, 'card')
    // (100 + 0.30) / (1 - 0.035 - 0.029) = 100.30 / 0.936 ≈ 107.158 → ceil to 107.16
    expect(f.gross).toBeGreaterThan(100)
    expect(f.gross).toBeLessThan(115)
  })

  it('grossForNet card: breakdown is derived from calculated gross', () => {
    const f = calc.grossForNet(100, 'card')
    const direct = calc.calculateFees(f.gross, 'card')
    expect(f.platformFee).toBe(direct.platformFee)
    expect(f.processingFee).toBe(direct.processingFee)
  })

  it('grossForNet ach (below cap): school receives at least netTarget (±$0.02)', () => {
    // net=100; gross=100/(1-0.035-0.008)=100/0.957≈104.49; processing=104.49×0.008≈0.84 < 5 → uncapped
    const f = calc.grossForNet(100, 'ach')
    expect(f.schoolReceives).toBeGreaterThanOrEqual(99.98)
    expect(f.schoolReceives).toBeLessThanOrEqual(100.02)
  })

  it('grossForNet ach (capped path): school receives at least netTarget (±$0.02)', () => {
    // large net where 0.8% would exceed $5 cap
    // net=700; uncapped gross=700/0.957≈731.45; processing=731.45×0.008≈5.85 > 5 → use capped formula
    // capped gross=(700+5)/(1-0.035)=705/0.965≈730.57
    const f = calc.grossForNet(700, 'ach')
    expect(f.processingFee).toBe(5.00)
    expect(f.schoolReceives).toBeGreaterThanOrEqual(699.98)
    expect(f.schoolReceives).toBeLessThanOrEqual(700.02)
  })

  it('grossForNet check: school receives at least netTarget (±$0.02)', () => {
    const f = calc.grossForNet(500, 'check')
    expect(f.processingFee).toBe(0)
    expect(f.schoolReceives).toBeGreaterThanOrEqual(499.98)
    expect(f.schoolReceives).toBeLessThanOrEqual(500.02)
  })

  it('grossForNet throws on zero netTarget', () => {
    expect(() => calc.grossForNet(0, 'card')).toThrow()
  })

  it('grossForNet throws on negative netTarget', () => {
    expect(() => calc.grossForNet(-50, 'ach')).toThrow()
  })
})

// ── Part 3 ─────────────────────────────────────────────────────────────────

describe('Part 3 — recordTransaction and getCampaignFeeSummary', () => {
  let calc: FeeCalculator

  beforeEach(() => {
    calc = new FeeCalculator()
    calc.recordTransaction('camp-a', 100,   'card',  true)   // school: 93.30
    calc.recordTransaction('camp-a', 200,   'check', false)  // school: 193.00
    calc.recordTransaction('camp-a', 1000,  'ach',   true)   // school: 960.00 (capped)
    calc.recordTransaction('camp-b', 500,   'card',  false)  // school: 464.20
  })

  it('recordTransaction returns correct TransactionRecord fields', () => {
    calc2: {
      const c2 = new FeeCalculator()
      const tx = c2.recordTransaction('camp-z', 100, 'card', true)
      expect(tx.campaignId).toBe('camp-z')
      expect(tx.gross).toBe(100)
      expect(tx.method).toBe('card')
      expect(tx.donorCoversFee).toBe(true)
      expect(tx.id).toBeTruthy()
      expect(tx.schoolReceives).toBe(93.30)
      expect(tx.totalFees).toBe(6.70)
    }
  })

  it('recordTransaction auto-generates unique IDs', () => {
    const c2 = new FeeCalculator()
    const a = c2.recordTransaction('x', 100, 'card', false)
    const b = c2.recordTransaction('x', 200, 'card', false)
    expect(a.id).not.toBe(b.id)
  })

  it('recordTransaction throws on gross <= 0', () => {
    expect(() => calc.recordTransaction('camp-a', 0, 'card', false)).toThrow()
  })

  it('getCampaignFeeSummary returns correct totals', () => {
    const s = calc.getCampaignFeeSummary('camp-a')
    expect(s.campaignId).toBe('camp-a')
    expect(s.transactionCount).toBe(3)
    expect(s.totalGross).toBe(1300)
    // school: 93.30 + 193.00 + 960.00 = 1246.30
    expect(s.totalSchoolReceives).toBe(1246.30)
    // fees: 6.70 + 7.00 + 40.00 = 53.70
    expect(s.totalFees).toBe(53.70)
  })

  it('getCampaignFeeSummary feeCoverageRate is fraction of donorCoversFee=true', () => {
    // camp-a: 2 covered out of 3 → 0.67
    const s = calc.getCampaignFeeSummary('camp-a')
    expect(s.feeCoverageRate).toBe(0.67)
  })

  it('getCampaignFeeSummary byMethod only includes used methods', () => {
    const s = calc.getCampaignFeeSummary('camp-a')
    expect(s.byMethod['card']).toBeDefined()
    expect(s.byMethod['ach']).toBeDefined()
    expect(s.byMethod['check']).toBeDefined()
  })

  it('getCampaignFeeSummary byMethod has correct aggregates', () => {
    const s = calc.getCampaignFeeSummary('camp-a')
    expect(s.byMethod['card']?.count).toBe(1)
    expect(s.byMethod['card']?.totalGross).toBe(100)
    expect(s.byMethod['ach']?.count).toBe(1)
    expect(s.byMethod['ach']?.totalGross).toBe(1000)
  })

  it('getCampaignFeeSummary throws on unknown campaign', () => {
    expect(() => calc.getCampaignFeeSummary('no-such')).toThrow()
  })

  it('getCampaignFeeSummary is isolated per campaign', () => {
    const s = calc.getCampaignFeeSummary('camp-b')
    expect(s.transactionCount).toBe(1)
    expect(s.totalGross).toBe(500)
  })
})
