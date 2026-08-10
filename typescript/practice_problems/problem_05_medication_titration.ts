/**
 * =============================================================================
 * INTERVIEW PROBLEM 5: Medication Titration Tracker
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * Company context: Health Tech
 * =============================================================================
 *
 * CONTEXT
 * -------
 * A health tech company's remote clinical care program often de-escalates
 * (reduces or stops) diabetes medications as patients' blood sugar improves.
 * Coaches and physicians need to track each patient's medication history —
 * when doses were changed and why — so they can coordinate care and generate
 * compliance reports.
 *
 * You are building the TitrationTracker — the class that ingests a stream of
 * titration events and answers questions about each patient's medication history.
 *
 * PRE-GIVEN (do not modify these)
 * --------------------------------
 * TitrationEvent and Medication types, and their factory functions, are fully
 * implemented. You implement TitrationTracker.
 *
 * Titration direction vocabulary:
 *   "increase" — dose or frequency was raised
 *   "decrease" — dose or frequency was lowered (de-escalation)
 *   "stop"     — medication discontinued entirely
 *   "start"    — new medication introduced
 *
 * # Example
 * const events = [
 *   makeTitrationEvent("pt1", "metformin", "start",    500.0, "2024-01-01"),
 *   makeTitrationEvent("pt1", "metformin", "increase", 1000.0, "2024-02-01"),
 *   makeTitrationEvent("pt1", "metformin", "decrease", 500.0, "2024-03-01"),
 *   makeTitrationEvent("pt1", "metformin", "stop",     0.0, "2024-04-01"),
 * ];
 * const t = new TitrationTracker(events);
 * t.currentMedications("pt1");
 * // -> []
 * t.titrationCount("pt1", "metformin", "decrease");
 * // -> 1
 *
 * NOTES
 * -----
 *   - Events are not guaranteed to arrive in chronological order — sort by date.
 *   - A medication is "active" if the most recent event for it is NOT "stop".
 *   - recordedOn is an ISO calendar date string ("YYYY-MM-DD"); it sorts
 *     correctly lexicographically.
 *   - Store all state in instance properties initialized in the constructor.
 *   - You choose the internal data structures — the public interface is what matters.
 * =============================================================================
 */

// ---------------------------------------------------------------------------
// PRE-GIVEN — do not modify
// ---------------------------------------------------------------------------

export type TitrationDirection = "start" | "increase" | "decrease" | "stop";

/** A single medication change recorded by a clinical coach or physician. */
export type TitrationEvent = {
  patientId: string;
  medication: string; // e.g. "metformin", "glipizide", "insulin_glargine"
  direction: TitrationDirection;
  doseMg: number; // dose in milligrams at the time of this event (0.0 for "stop")
  recordedOn: string; // ISO calendar date, "YYYY-MM-DD"
};

export function makeTitrationEvent(
  patientId: string,
  medication: string,
  direction: TitrationDirection,
  doseMg: number,
  recordedOn: string,
): TitrationEvent {
  return { patientId, medication, direction, doseMg, recordedOn };
}

/** Summary of a patient's current relationship with a single medication. */
export type Medication = {
  name: string;
  currentDose: number;
  lastChanged: string; // ISO calendar date, "YYYY-MM-DD"
  totalChanges: number; // total number of titration events (including start/stop)
};

// ---------------------------------------------------------------------------
// YOUR IMPLEMENTATION
// ---------------------------------------------------------------------------

/**
 * Ingests a list of TitrationEvents and answers questions about
 * patient medication histories.
 */
export class TitrationTracker {
  constructor(events: TitrationEvent[]) {
    // TODO: store and organize events however makes the methods efficient.
    // All state must be in instance properties (not static/class-level fields).
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 1 — Current medication snapshot  (~10 min)
  // ---------------------------------------------------------------------------

  /**
   * Return a list of Medication objects for all currently active medications
   * for the given patient (i.e., medications whose latest event is NOT "stop").
   *
   * Each Medication reflects:
   *   - name:         the medication name
   *   - currentDose:  doseMg from the most recent event for that medication
   *   - lastChanged:  date of the most recent event
   *   - totalChanges: total number of TitrationEvents recorded for this medication
   *
   * Return an empty array if the patient has no events or all medications have
   * been stopped.
   *
   * The array may be returned in any order.
   */
  currentMedications(patientId: string): Medication[] {
    throw new Error("Not implemented");
  }

  /**
   * Return all TitrationEvents for (patientId, medication), sorted
   * chronologically (earliest first).
   *
   * Return an empty array if no events exist for that combination.
   */
  getMedicationHistory(patientId: string, medication: string): TitrationEvent[] {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 2 — Titration counts  (~10 min)
  // ---------------------------------------------------------------------------

  /**
   * Return the number of titration events for (patientId, medication).
   *
   * If direction is provided (one of "start", "increase", "decrease", "stop"),
   * return only events with that direction.
   *
   * Return 0 if the patient or medication is unknown.
   */
  titrationCount(patientId: string, medication: string, direction?: TitrationDirection): number {
    throw new Error("Not implemented");
  }

  /**
   * Return a map of each medication name to the number of "decrease" or
   * "stop" events recorded for that patient.
   *
   * Only include medications that have at least one decrease or stop event.
   * Return an empty map if the patient has no such events.
   *
   * Example:
   *   { metformin: 2, glipizide: 1 }   // metformin: 1 decrease + 1 stop
   */
  deEscalationSummary(patientId: string): Record<string, number> {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 3 — Population-level queries  (~15 min)
  // ---------------------------------------------------------------------------

  /**
   * Return a sorted list of patientIds who currently have the given
   * medication active (latest event is NOT "stop").
   */
  patientsOnMedication(medication: string): string[] {
    throw new Error("Not implemented");
  }

  /**
   * Return the topN medications (by total titration event count across ALL
   * patients) as a list of [medicationName, totalCount] tuples,
   * sorted descending by count.
   *
   * If fewer than topN medications exist, return all of them.
   * Ties may appear in any order.
   */
  mostTitratedMedications(topN: number = 3): [string, number][] {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 4 — Live ingestion  (~10 min)
  // ---------------------------------------------------------------------------

  /**
   * Add a new TitrationEvent to the tracker.
   *
   * If an event with the same (patientId, medication, recordedOn) already
   * exists in the tracker, overwrite it with the new event (last-write wins).
   */
  addEvent(event: TitrationEvent): void {
    throw new Error("Not implemented");
  }
}
