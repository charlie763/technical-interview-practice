/**
 * =============================================================================
 * INTERVIEW PROBLEM 05: Giving Day Challenge Engine
 * Difficulty: Senior Software Engineer | Estimated time: 45–60 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the challenge engine for a school giving-day fundraiser.
 * Sponsors set up challenges: "if the school raises $X by deadline Y, we'll
 * add a $Z bonus." As donations roll in, the engine checks whether any
 * challenges have been completed. At the end of the day, advancement staff
 * review an hourly breakdown of activity and which challenges fired when.
 *
 * Store all state in instance variables initialised in the constructor.
 * You choose the internal data structures; the public interface is what matters.
 *
 * ⚠️  TYPE CONTRACT
 * -----------------
 * This problem intentionally omits interface definitions for domain objects
 * and return values. Part of the challenge is designing well-typed TypeScript
 * interfaces that accurately model the domain. Define your types in your
 * answer file before implementing the methods.
 *
 * The only pre-defined type is BonusType (used as a parameter and cannot
 * change without breaking the test suite).
 *
 * CHALLENGE SEMANTICS
 * -------------------
 * A challenge is COMPLETE when:
 *   sum of all donations whose timestamp <= challenge.deadline >= challenge.thresholdAmount
 *
 * Assume donations are always recorded in chronological order.
 * Once a challenge is marked complete it stays complete.
 *
 * // Example
 * // const engine = new GivingDayEngine()
 * // engine.addChallenge('c1', 'Morning Boost', 1000, '2025-10-01T12:00:00', 'flat_bonus', 500)
 * // engine.processDonation(400, '2025-10-01T09:00:00')  // cum=400 < 1000, no completion
 * // engine.processDonation(300, '2025-10-01T10:00:00')  // cum=700 < 1000, no completion
 * // engine.processDonation(400, '2025-10-01T11:00:00')  // cum=1100 ≥ 1000, c1 completes!
 * // engine.getChallenge('c1').completed  // → true
 * // engine.getChallenge('c1').completedAt // → '2025-10-01T11:00:00'
 *
 * =============================================================================
 * PART 1 — Challenge setup and donation recording
 * =============================================================================
 *
 * Implement addChallenge, addDonation, getChallenge, and getTotalRaised.
 * Define your Challenge and DonationEntry types in this file.
 */

// ── Pre-defined type (do not change) ─────────────────────────────────────────

export type BonusType = 'flat_bonus' | 'match_pct'

// ── Define your own interfaces here ──────────────────────────────────────────
//
// You will need types for at least:
//
//   Challenge — stored per challenge; must include:
//     id, name, thresholdAmount, deadline (ISO-8601 datetime),
//     bonusType (BonusType), bonusValue (number),
//     completed (boolean), completedAt (string | null)
//
//   DonationEntry — a recorded donation; must include:
//     id (auto-generated), amount (number), timestamp (ISO-8601 datetime)
//
//   HourlyBucket (Part 3) — one bucket per calendar hour; must include:
//     hour (ISO-8601 string for the start of the hour, e.g. '2025-10-01T09:00:00'),
//     donationCount (number),
//     amountRaised (number — total donated this hour),
//     cumulativeTotal (number — running sum up to and including this hour),
//     challengesCompleted (string[] — IDs of challenges that first completed
//                          due to a donation in this hour)

// ── Class ──────────────────────────────────────────────────────────────────

export class GivingDayEngine {
  constructor() {
    throw new Error('Not implemented')
  }

  // ── Part 1 ───────────────────────────────────────────────────────────────

  /**
   * Register a challenge.
   *
   * Returns: the stored challenge object (completed=false, completedAt=null).
   * @throws {Error} if id already exists
   * @throws {Error} if thresholdAmount <= 0
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  addChallenge(
    id: string,
    name: string,
    thresholdAmount: number,
    deadline: string,
    bonusType: BonusType,
    bonusValue: number,
  ): any {
    throw new Error('Not implemented')
  }

  /**
   * Record a donation. Donation IDs are auto-generated ("don-1", "don-2", …).
   * Does NOT trigger challenge completion checks (use processDonation for that).
   *
   * Returns: the stored donation entry object.
   * @throws {Error} if amount <= 0
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  addDonation(amount: number, timestamp: string): any {
    throw new Error('Not implemented')
  }

  /**
   * Return a challenge by ID.
   * @throws {Error} if id does not exist
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getChallenge(id: string): any {
    throw new Error('Not implemented')
  }

  /**
   * Return the total amount raised from all donations whose timestamp <= asOf.
   * When asOf is undefined, return the total of ALL donations.
   */
  getTotalRaised(asOf?: string): number {
    throw new Error('Not implemented')
  }

  // ── Part 2 ───────────────────────────────────────────────────────────────

  /**
   * Record a donation AND check all incomplete challenges for completion.
   * Calls addDonation internally — do not duplicate donation recording logic.
   *
   * A challenge is "newly completed" if:
   *   - it was incomplete before this donation, AND
   *   - getTotalRaised(asOf=challenge.deadline) >= challenge.thresholdAmount after adding it
   *
   * Mark each newly-completed challenge: set completed=true and
   * completedAt=timestamp of this donation.
   *
   * Returns an object with:
   *   donation:        the stored donation entry (from addDonation)
   *   newlyCompleted:  array of challenge objects that just completed
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  processDonation(amount: number, timestamp: string): any {
    throw new Error('Not implemented')
  }

  // ── Part 3 ───────────────────────────────────────────────────────────────

  /**
   * Return an hourly breakdown of all donation activity, ordered by hour.
   *
   * One HourlyBucket per calendar hour that had at least one donation.
   * The hour string is the ISO-8601 start of that hour, e.g. '2025-10-01T09:00:00'.
   *
   * challengesCompleted lists challenge IDs whose completedAt falls within
   * this hour (i.e. completedAt starts with the same 'YYYY-MM-DDTHH' prefix).
   *
   * Uses stored donation and challenge data — do not duplicate
   * getTotalRaised logic.
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getHourlyBreakdown(): any[] {
    throw new Error('Not implemented')
  }
}
