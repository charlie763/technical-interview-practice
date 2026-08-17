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

export type PledgeType = "per_lap" | "flat";

// ── Define your own interfaces here ──────────────────────────────────────────
//
// You will need types for at least:
//   • Participant  — the domain object stored per participant
//                    must include: id, name, classId, lapsCompleted (starts 0)
//   • Pledge       — a single sponsor pledge
//                    must include: id (auto-generated), participantId,
//                    sponsorName, sponsorEmail, pledgeType, amount
//.  - Sponsor ?
//   - Class
//   • The return type(s) of getParticipantTotal and getClassLeaderboard
//     — define these as you design the methods below.

interface ParticipantBase {
  id: string;
  name: string;
  classId: string;
  lapsCompleted: number;
}

interface Participant extends ParticipantBase {
  pledges: Pledge[];
}

interface ParticipantTotal extends Omit<ParticipantBase, "id"> {
  participantId: string;
  perLapTotal: number;
  flatTotal: number;
  totalRaised: number;
}

interface Pledge {
  id: string; // auto-generated
  participantId: string;
  sponsorName: string;
  sponsorEmail: string;
  pledgeType: PledgeType;
  amount: number;
}

interface ClassLeaderBoardEntry {
  classId: string;
  totalRaised: number;
  participantCount: number;
  participantTotals: ParticipantTotal[];
  topFundraiser: ParticipantTotal;
}

// ── Class ──────────────────────────────────────────────────────────────────

export class PledgeTracker {
  participants: Record<string, Participant>;
  currentPledgeNum: number;
  constructor() {
    this.participants = {};
    this.currentPledgeNum = 1;
  }

  // ── Part 1 ───────────────────────────────────────────────────────────────

  /**
   * Register a new participant. lapsCompleted starts at 0.
   *
   * Returns: the stored participant object.
   * @throws {Error} if id already exists
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  registerParticipant(id: string, name: string, classId: string): Participant {
    const existingParticipant = this.participants[id];
    if (existingParticipant) {
      throw new Error(`participant with id ${id} already exists`);
    }
    const newParticipant: Participant = {
      id,
      name,
      classId,
      lapsCompleted: 0,
      pledges: [],
    };
    this.participants[id] = newParticipant;
    return newParticipant;
  }

  /**
   * Add a pledge for a participant.
   * Pledge IDs are auto-generated ("pledge-1", "pledge-2", …).
   *
   * Returns: the stored pledge object.
   * @throws {Error} if participantId does not exist
   * @throws {Error} if amount <= 0
   */
  addPledge(
    participantId: string,
    sponsorName: string,
    sponsorEmail: string,
    pledgeType: PledgeType,
    amount: number,
  ): Pledge {
    const participant = this.getParticipant(participantId);
    if (amount <= 0) {
      throw new Error("amount must be greater than zero");
    }
    const newPledge: Pledge = {
      id: `pledge-${this.currentPledgeNum}`,
      participantId,
      sponsorEmail,
      sponsorName,
      pledgeType,
      amount,
    };
    this.currentPledgeNum += 1;
    participant.pledges.push(newPledge);
    return newPledge;
  }

  /**
   * Return the participant object for the given ID.
   * @throws {Error} if participantId does not exist
   */
  getParticipant(participantId: string): Participant {
    const existingParticipant = this.participants[participantId];
    if (!existingParticipant) {
      throw new Error(`participant with id ${participantId} does not exist`);
    }
    return existingParticipant;
  }

  // ── Part 2 ───────────────────────────────────────────────────────────────

  /**
   * Set the number of laps a participant completed (not an increment).
   *
   * Returns: the updated participant object.
   * @throws {Error} if participantId does not exist
   * @throws {Error} if lapsCompleted < 0
   */
  recordLaps(participantId: string, lapsCompleted: number): Participant {
    const participant = this.getParticipant(participantId);
    if (lapsCompleted < 0) {
      throw new Error("lapsCompleted must be greater than 0");
    }
    participant.lapsCompleted = lapsCompleted;
    return participant;
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
  getParticipantTotal(participantId: string): ParticipantTotal {
    const participant = this.getParticipant(participantId);
    let sumOfPerLapPledges = 0;
    let sumOfFlatPledges = 0;
    participant.pledges.forEach((pledge) => {
      if (pledge.pledgeType === "flat") {
        sumOfFlatPledges += pledge.amount;
      } else if (pledge.pledgeType === "per_lap") {
        sumOfPerLapPledges += pledge.amount;
      }
    });
    const perLapTotal = sumOfPerLapPledges * participant.lapsCompleted;
    const participantTotal: ParticipantTotal = {
      participantId: participant.id,
      name: participant.name,
      classId: participant.classId,
      lapsCompleted: participant.lapsCompleted,
      perLapTotal,
      flatTotal: sumOfFlatPledges,
      totalRaised: perLapTotal + sumOfFlatPledges,
    };
    return participantTotal;
  }

  /**
   * Return the sum of totalRaised across ALL participants.
   * Calls getParticipantTotal internally.
   */
  getCampaignTotal(): number {
    let sumTotal = 0;
    Object.values(this.participants).forEach((participant) => {
      const particpantTotal = this.getParticipantTotal(participant.id);
      sumTotal += particpantTotal.totalRaised;
    });
    return sumTotal;
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

  getClassLeaderboard(): ClassLeaderBoardEntry[] {
    const classParticipantTotaltMap: Record<string, ClassLeaderBoardEntry> = {}; //index is classId
    Object.values(this.participants).forEach((participant) => {
      const participantTotal = this.getParticipantTotal(participant.id);
      const existingClassParticipantsTotals =
        classParticipantTotaltMap[participant.classId];
      if (existingClassParticipantsTotals) {
        existingClassParticipantsTotals.participantTotals.push(
          participantTotal,
        );
        existingClassParticipantsTotals.participantCount += 1;
        existingClassParticipantsTotals.totalRaised +=
          participantTotal.totalRaised;
        if (
          participantTotal.totalRaised >
          existingClassParticipantsTotals.topFundraiser.totalRaised
        ) {
          existingClassParticipantsTotals.topFundraiser = participantTotal;
        }
      } else {
        classParticipantTotaltMap[participant.classId] = {
          classId: participant.classId,
          totalRaised: participantTotal.totalRaised,
          participantCount: 1,
          participantTotals: [participantTotal],
          topFundraiser: participantTotal,
        };
      }
    });
    const leaderboardSortFn = (
      leaderboardEntryA: ClassLeaderBoardEntry,
      leaderboardEntryB: ClassLeaderBoardEntry,
    ) => {
      if (leaderboardEntryB.totalRaised > leaderboardEntryA.totalRaised) {
        return 1;
      } else if (
        leaderboardEntryB.totalRaised < leaderboardEntryA.totalRaised
      ) {
        return -1;
      } else {
        if (leaderboardEntryA.classId < leaderboardEntryB.classId) {
          return 1;
        } else if (leaderboardEntryA.classId > leaderboardEntryB.classId) {
          return -1;
        } else {
          return 1;
        }
      }
    };
    return Object.values(classParticipantTotaltMap).sort(leaderboardSortFn);
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
  getParticipantLeaderboard(limit?: number): ParticipantTotal[] {
    const participantTotals = Object.values(this.participants).map(
      (participant) => this.getParticipantTotal(participant.id),
    );
    const participantToalSortFn = (
      ptA: ParticipantTotal,
      ptB: ParticipantTotal,
    ) => {
      if (ptB.totalRaised > ptA.totalRaised) {
        return 1;
      } else if (ptB.totalRaised < ptA.totalRaised) {
        return -1;
      } else {
        if (ptA.participantId < ptB.participantId) {
          return 1;
        } else if (ptA.participantId > ptB.participantId) {
          return -1;
        } else {
          return 1;
        }
      }
    };
    const sortedParticipantTotals = [...participantTotals].sort(
      participantToalSortFn,
    );
    let rank = 0;
    return sortedParticipantTotals.slice(0, limit).map((participantTotal) => {
      rank += 1;
      return { ...participantTotal, rank };
    });
  }
}
