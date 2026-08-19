/**
 * =============================================================================
 * INTERVIEW PROBLEM 04: Platform Fee Calculator
 * Difficulty: Senior Software Engineer | Estimated time: 45–60 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the fee calculation module for a school fundraising platform.
 * The platform charges a fee on every transaction, and payment processors add
 * their own fees on top. A key donor experience feature is "donor covers fee":
 * the donor can opt to pay a higher gross amount so the school receives the
 * full intended net donation.
 *
 * Store all state in instance variables initialised in the constructor.
 * You choose the internal data structures; the public interface is what matters.
 *
 * FEE SCHEDULE (all fees calculated on the gross amount)
 * -------------------------------------------------------
 *   Platform fee:  3.5% of gross
 *   Processing fee by method:
 *     'card':  2.9% of gross + $0.30
 *     'ach':   0.8% of gross, capped at $5.00
 *     'check': $0.00
 *
 *   schoolReceives = gross − platformFee − processingFee
 *
 * Round all dollar values to the nearest cent: Math.round(x * 100) / 100.
 *
 * // Example — calculateFees
 * // calc.calculateFees(100, 'card')
 * // → { gross: 100, platformFee: 3.50, processingFee: 3.20,
 * //     totalFees: 6.70, schoolReceives: 93.30 }
 * //
 * // Example — grossForNet (donor covers fee so school gets exactly $100)
 * // Card formula: gross × (1 − 0.035 − 0.029) − 0.30 = net
 * //              gross = (net + 0.30) / 0.936
 * // calc.grossForNet(100, 'card')
 * // → { gross: 107.48, platformFee: 3.76, processingFee: 3.42,
 * //     totalFees: 7.18, schoolReceives: 100.30 }
 * //
 * // Tests allow ±$0.02 tolerance on schoolReceives for grossForNet
 * // because rounding the gross first may shift the result by a cent.
 *
 * =============================================================================
 * PART 1 — Fee calculation for a given gross amount
 * =============================================================================
 *
 * Implement calculateFees. The FeeCalculator class needs no stored state for
 * Part 1 (calculateFees is a pure calculation).
 */

// ── Types ──────────────────────────────────────────────────────────────────

export type PaymentMethod = 'card' | 'ach' | 'check'

export interface FeeBreakdown {
  gross: number          // dollar amount the donor pays
  platformFee: number    // 3.5% of gross
  processingFee: number  // depends on payment method
  totalFees: number      // platformFee + processingFee
  schoolReceives: number // gross − totalFees
}

export interface TransactionRecord {
  id: string             // auto-generated ("tx-1", "tx-2", …)
  campaignId: string
  gross: number
  schoolReceives: number
  totalFees: number
  method: PaymentMethod
  donorCoversFee: boolean
}

export interface MethodSummary {
  count: number
  totalGross: number
  totalSchoolReceives: number
  totalFees: number
}

export interface CampaignFeeSummary {
  campaignId: string
  transactionCount: number
  totalGross: number
  totalSchoolReceives: number
  totalFees: number
  feeCoverageRate: number  // fraction where donorCoversFee=true, 2 decimal places
  byMethod: Partial<Record<PaymentMethod, MethodSummary>>
}

// ── Class ──────────────────────────────────────────────────────────────────

export class FeeCalculator {
  static readonly PLATFORM_RATE  = 0.035
  static readonly CARD_PCT_RATE  = 0.029
  static readonly CARD_FIXED_FEE = 0.30
  static readonly ACH_PCT_RATE   = 0.008
  static readonly ACH_MAX_FEE    = 5.00

  constructor() {
    throw new Error('Not implemented')
  }

  // ── Part 1 ───────────────────────────────────────────────────────────────

  /**
   * Calculate the fee breakdown for a given gross donation amount.
   *
   * @throws {Error} if gross <= 0
   */
  calculateFees(gross: number, method: PaymentMethod): FeeBreakdown {
    throw new Error('Not implemented')
  }

  // ── Part 2 ───────────────────────────────────────────────────────────────

  /**
   * Calculate the gross amount a donor must pay so the school receives
   * at least netTarget (within ±$0.02 rounding tolerance).
   *
   * Use Math.ceil to the nearest cent when computing the gross, so the school
   * always receives >= netTarget. Then derive the FeeBreakdown from that gross
   * using calculateFees — do not duplicate the fee logic.
   *
   * Derived formulas (for reference):
   *   card:  gross = ceil((netTarget + CARD_FIXED_FEE) / (1 − PLATFORM_RATE − CARD_PCT_RATE), 2)
   *   ach (processingFee would be < ACH_MAX):
   *          gross = ceil(netTarget / (1 − PLATFORM_RATE − ACH_PCT_RATE), 2)
   *   ach (processingFee would reach ACH_MAX, i.e. gross × ACH_PCT_RATE >= ACH_MAX):
   *          gross = ceil((netTarget + ACH_MAX_FEE) / (1 − PLATFORM_RATE), 2)
   *   check: gross = ceil(netTarget / (1 − PLATFORM_RATE), 2)
   *
   * @throws {Error} if netTarget <= 0
   */
  grossForNet(netTarget: number, method: PaymentMethod): FeeBreakdown {
    throw new Error('Not implemented')
  }

  // ── Part 3 ───────────────────────────────────────────────────────────────

  /**
   * Record a transaction for a campaign.
   * Uses calculateFees(gross, method) to compute schoolReceives and totalFees.
   *
   * @throws {Error} if gross <= 0
   */
  recordTransaction(
    campaignId: string,
    gross: number,
    method: PaymentMethod,
    donorCoversFee: boolean,
  ): TransactionRecord {
    throw new Error('Not implemented')
  }

  /**
   * Return aggregated fee statistics for a campaign.
   * feeCoverageRate = coveredCount / totalCount, rounded to 2 decimal places.
   * Only includes methods that have at least one transaction in byMethod.
   *
   * @throws {Error} if campaignId has no recorded transactions
   */
  getCampaignFeeSummary(campaignId: string): CampaignFeeSummary {
    throw new Error('Not implemented')
  }
}
