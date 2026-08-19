/**
 * =============================================================================
 * INTERVIEW PROBLEM 02: Walkathon Pledge Tracker
 * Difficulty: Senior Software Engineer | Estimated time: 45–60 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the pledge management module for a school walk-a-thon
 * fundraiser. Before the event, student participants collect pledges from
 * sponsors — either a flat donation or a per-lap amount. After the event, the
 * school records how many laps each student completed and the platform
 * calculates what each sponsor owes and produces class-level leaderboards.
 *
 * Store all state in instance variables initialised in the constructor.
 * You choose the internal data structures; the public interface is what matters.
 *
 * ⚠️  TYPE CONTRACT
 * -----------------
 * This problem intentionally omits interface definitions. Part of the challenge
 * is designing your own TypeScript types for the domain objects and return
 * values. Define your interfaces/types in your answer file before implementing
 * the methods. The docstrings below describe the required shape of each
 * return value.
 *
 * The only pre-defined type is PledgeType (it is used as a parameter type and
 * cannot change without breaking the test suite).
 *
 * PLEDGE CALCULATION
 * ------------------
 * participant.totalRaised = (sum of per_lap pledges × lapsCompleted)
 *                         + (sum of flat pledges)
 *
 * // Example
 * // const pt = new PledgeTracker()
 * // pt.registerParticipant('p1', 'Emma Torres', 'class-5a')
 * // pt.addPledge('p1', 'Grandma Rose', 'rose@example.com', 'per_lap', 5)
 * // pt.addPledge('p1', 'Uncle Joe', 'joe@example.com', 'flat', 50)
 * // pt.recordLaps('p1', 10)
 * // const total = pt.getParticipantTotal('p1')
 * // total.totalRaised  // → 100  (10 laps × $5 + $50 flat)
 * // total.perLapTotal  // → 50
 * // total.flatTotal    // → 50
 *
 * =============================================================================
 * PART 1 — Participant registration and pledge collection
 * =============================================================================
 *
 * Implement registerParticipant, addPledge, and getParticipant.
 * Define your Participant and Pledge interfaces in this file.
 */

// ── Pre-defined type (do not change) ─────────────────────────────────────────

export type PledgeType = 'per_lap' | 'flat'

// ── Define your own interfaces here ──────────────────────────────────────────
//
// You will need types for at least:
//   • Participant  — the domain object stored per participant
//                    must include: id, name, classId, lapsCompleted (starts 0)
//   • Pledge       — a single sponsor pledge
//                    must include: id (auto-generated), participantId,
//                    sponsorName, sponsorEmail, pledgeType, amount
//   • The return type(s) of getParticipantTotal and getClassLeaderboard
//     — define these as you design the methods below.

// ── Class ──────────────────────────────────────────────────────────────────

export class PledgeTracker {
  constructor() {
    throw new Error('Not implemented')
  }

  // ── Part 1 ───────────────────────────────────────────────────────────────

  /**
   * Register a new participant. lapsCompleted starts at 0.
   *
   * Returns: the stored participant object.
   * @throws {Error} if id already exists
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  registerParticipant(id: string, name: string, classId: string): any {
    throw new Error('Not implemented')
  }

  /**
   * Add a pledge for a participant.
   * Pledge IDs are auto-generated ("pledge-1", "pledge-2", …).
   *
   * Returns: the stored pledge object.
   * @throws {Error} if participantId does not exist
   * @throws {Error} if amount <= 0
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  addPledge(
    participantId: string,
    sponsorName: string,
    sponsorEmail: string,
    pledgeType: PledgeType,
    amount: number,
  ): any {
    throw new Error('Not implemented')
  }

  /**
   * Return the participant object for the given ID.
   * @throws {Error} if participantId does not exist
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getParticipant(participantId: string): any {
    throw new Error('Not implemented')
  }

  // ── Part 2 ───────────────────────────────────────────────────────────────

  /**
   * Set the number of laps a participant completed (not an increment).
   *
   * Returns: the updated participant object.
   * @throws {Error} if participantId does not exist
   * @throws {Error} if lapsCompleted < 0
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  recordLaps(participantId: string, lapsCompleted: number): any {
    throw new Error('Not implemented')
  }

  /**
   * Calculate total raised for a single participant.
   * Calls getParticipant internally — do not duplicate lookup logic.
   *
   * Returns an object with at minimum:
   *   participantId: string
   *   name:          string
   *   classId:       string
   *   lapsCompleted: number
   *   perLapTotal:   number   ← sum of per_lap pledges × lapsCompleted
   *   flatTotal:     number   ← sum of flat pledges
   *   totalRaised:   number   ← perLapTotal + flatTotal
   *
   * @throws {Error} if participantId does not exist
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getParticipantTotal(participantId: string): any {
    throw new Error('Not implemented')
  }

  /**
   * Return the sum of totalRaised across ALL participants.
   * Calls getParticipantTotal internally.
   */
  getCampaignTotal(): number {
    throw new Error('Not implemented')
  }

  // ── Part 3 ───────────────────────────────────────────────────────────────

  /**
   * Return a leaderboard of classes, sorted by totalRaised descending.
   * On a tie, sort by classId ascending.
   * Calls getParticipantTotal internally.
   *
   * Each entry must include at minimum:
   *   classId:          string
   *   totalRaised:      number
   *   participantCount: number
   *   topFundraiser:    { participantId, name, totalRaised }
   *     — the participant in the class with the highest totalRaised
   *     — ties broken by participantId ascending
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getClassLeaderboard(): any[] {
    throw new Error('Not implemented')
  }

  /**
   * Return participants ranked by totalRaised descending.
   * Ties broken by participantId ascending.
   * Returns all participants when limit is undefined.
   * Calls getParticipantTotal internally.
   *
   * Each entry must include all fields from getParticipantTotal PLUS:
   *   rank: number   ← 1-based position
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getParticipantLeaderboard(limit?: number): any[] {
    throw new Error('Not implemented')
  }
}
