/**
 * =============================================================================
 * INTERVIEW PROBLEM 4: Biomarker Alert Monitor
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * Company context: Health Tech
 * =============================================================================
 *
 * CONTEXT
 * -------
 * A health tech company delivers remote clinical care for type 2 diabetes
 * reversal. Patients use connected glucometers and ketone meters that sync
 * readings into the app several times per day. The care team dashboard needs
 * to surface patients whose numbers have been out of target range for
 * multiple consecutive days, so health coaches can prioritize outreach.
 *
 * You are building the BiomarkerMonitor — the class that processes a stream of
 * patient readings and answers questions about trends and outreach priority.
 *
 * PRE-GIVEN (do not modify these)
 * --------------------------------
 * BiomarkerReading and the makeBiomarkerReading factory are fully implemented.
 * You implement BiomarkerMonitor.
 *
 * Target ranges:
 *   Glucose: 70–180 mg/dL  (< 70 is a dangerous low, > 180 is hyperglycemia)
 *   Ketone:  0.5–3.0 mmol/L  (below = not in ketosis, above = monitor)
 *   Weight:  no absolute target — not flagged as out-of-range
 *
 * # Example
 * const readings = [
 *   makeBiomarkerReading("alice", "glucose", 195.0, "2024-01-01"),
 *   makeBiomarkerReading("alice", "glucose", 202.0, "2024-01-02"),
 *   makeBiomarkerReading("alice", "glucose", 188.0, "2024-01-03"),
 * ];
 * const m = new BiomarkerMonitor(readings);
 * m.maxConsecutiveOutOfRangeDays("alice", "glucose");  // -> 3
 * m.getOutreachList(3);
 * // -> [{ patientId: "alice", readingType: "glucose",
 * //       consecutiveDays: 3, latestValue: 188.0 }]
 *
 * NOTES
 * -----
 *   - Multiple readings on the same calendar day count as ONE day.
 *     A day is "out-of-range" if ANY reading that day is out of range.
 *   - recordedOn is an ISO calendar date string ("YYYY-MM-DD"), no time
 *     component — compare/sort it lexicographically or via `new Date(...)`.
 *   - Store all state in instance properties initialized in the constructor.
 * =============================================================================
 */

// ---------------------------------------------------------------------------
// PRE-GIVEN — do not modify
// ---------------------------------------------------------------------------

export const GLUCOSE_RANGE: [number, number] = [70.0, 180.0]; // mg/dL, inclusive
export const KETONE_RANGE: [number, number] = [0.5, 3.0]; // mmol/L, inclusive

export type ReadingType = "glucose" | "ketone" | "weight";

/** A single biomarker measurement from a patient device or manual entry. */
export type BiomarkerReading = {
  patientId: string;
  readingType: ReadingType;
  value: number;
  recordedOn: string; // ISO calendar date, "YYYY-MM-DD"
};

export function makeBiomarkerReading(patientId: string, readingType: ReadingType, value: number, recordedOn: string): BiomarkerReading {
  return { patientId, readingType, value, recordedOn };
}

export type OutreachEntry = {
  patientId: string;
  readingType: ReadingType;
  consecutiveDays: number;
  latestValue: number;
};

// ---------------------------------------------------------------------------
// YOUR IMPLEMENTATION
// ---------------------------------------------------------------------------

/**
 * Ingests a list of BiomarkerReadings and answers questions about
 * out-of-range trends across the patient population.
 */
export class BiomarkerMonitor {
  constructor(readings: BiomarkerReading[]) {
    // TODO: store and organize readings however makes the methods efficient.
    // All state must be in instance properties (not static/class-level fields).
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 1 — single-reading classification  (~5 min)
  // ---------------------------------------------------------------------------

  /**
   * Return true if the reading falls outside the target range for its type.
   * - Glucose: outside GLUCOSE_RANGE
   * - Ketone:  outside KETONE_RANGE
   * - Weight:  never out-of-range (return false)
   */
  isOutOfRange(reading: BiomarkerReading): boolean {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 2 — streak detection  (~15 min)
  // ---------------------------------------------------------------------------

  /**
   * Return the length of the longest streak of *consecutive calendar days*
   * on which the patient had at least one out-of-range reading of the given type.
   *
   * Return 0 if the patient has no out-of-range readings of that type.
   *
   * Consecutive means no gap: Jan 1, Jan 2, Jan 3 is a streak of 3.
   * Jan 1, Jan 3 (skipping Jan 2) is two separate streaks of 1.
   * Multiple readings on the same day collapse to one day.
   */
  maxConsecutiveOutOfRangeDays(patientId: string, readingType: ReadingType): number {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 3 — outreach list  (~10 min)
  // ---------------------------------------------------------------------------

  /**
   * Return a list of patients who need proactive coach outreach because
   * they have been out-of-range for at least minConsecutiveDays in a row.
   *
   * A patient can appear more than once if multiple reading types cross the
   * threshold (e.g., both glucose and ketone streaks).
   *
   * Sort the list by consecutiveDays descending (most urgent first).
   */
  getOutreachList(minConsecutiveDays: number = 3): OutreachEntry[] {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 4 — deduplication on ingestion  (~10 min)
  // ---------------------------------------------------------------------------

  /**
   * Add a new reading to the monitor's internal state.
   * Return true if the reading was added successfully.
   * Return false (without adding) if it is a duplicate.
   *
   * A duplicate is: same patientId, readingType, and recordedOn,
   * with a value within ±0.5 of an existing reading on that day.
   */
  addReading(reading: BiomarkerReading): boolean {
    throw new Error("Not implemented");
  }
}
