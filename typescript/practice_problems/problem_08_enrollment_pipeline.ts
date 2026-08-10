/**
 * =============================================================================
 * Patient Enrollment Pipeline
 * =============================================================================
 *
 * A clinical care program moves patients through a structured enrollment pipeline.
 * Each patient begins in the "referred" state and advances through a predefined set
 * of transitions until reaching a terminal state.
 *
 * Allowed transitions (defined in ALLOWED_TRANSITIONS below):
 *   referred   →  screened
 *   screened   →  enrolled  |  ineligible
 *   enrolled   →  active    |  withdrawn
 *   active     →  graduated |  churned  |  withdrawn
 *
 * Terminal states: ineligible, withdrawn, graduated, churned
 *
 * You choose the internal data structures — the public interface is what matters.
 *
 * Store all state in instance properties initialized in the constructor.
 * Class-level (static) fields will bleed between tests and between EnrollmentPipeline
 * instances — avoid them.
 *
 * ────────────────────────────────────────────────────────────────────────────
 * Part 1 — State tracking
 *   addPatient(patientId, timestamp)
 *   transition(patientId, newState, timestamp)
 *   getState(patientId)               -> string
 *   getPatientsInState(state)         -> string[]
 *
 * Part 2 — Duration and conversion metrics
 *   timeInState(patientId, state, asOf) -> number
 *   conversionRate(fromState, toState)  -> number
 *
 * Part 3 — SLA monitoring
 *   patientsOverdue(state, maxSeconds, asOf) -> string[]
 *   averageTimeInState(state, asOf)          -> number
 * ────────────────────────────────────────────────────────────────────────────
 *
 * # Example
 * const pipeline = new EnrollmentPipeline();
 * pipeline.addPatient("p_001", 0.0);
 * pipeline.transition("p_001", "screened", 86400.0);   // 1 day later
 * pipeline.transition("p_001", "enrolled", 172800.0);  // 2 days later
 * pipeline.getState("p_001");                   // -> "enrolled"
 * pipeline.getPatientsInState("enrolled");      // -> ["p_001"]
 *
 * pipeline.addPatient("p_002", 0.0);
 * pipeline.transition("p_002", "screened",   43200.0);
 * pipeline.transition("p_002", "ineligible", 86400.0);
 * pipeline.getPatientsInState("screened");      // -> []  (both have moved on)
 *
 * // Part 2
 * pipeline.timeInState("p_001", "screened", 999999.0);
 * // -> 86400.0  (172800 - 86400; already exited, asOf ignored)
 * pipeline.conversionRate("screened", "enrolled");
 * // -> 0.5  (p_001 enrolled, p_002 ineligible; one of two converted)
 *
 * // Part 3
 * pipeline.transition("p_001", "active", 259200.0);
 * pipeline.patientsOverdue("active", 3600.0, 270000.0);
 * // -> ["p_001"]  (has been active 10800 s > 3600 s threshold)
 * pipeline.averageTimeInState("screened", 999999.0);
 * // -> 64800.0  ((86400 + 43200) / 2; both p_001 and p_002 have exited screened)
 * =============================================================================
 */

export const ALLOWED_TRANSITIONS: Record<string, Set<string>> = {
  referred: new Set(["screened"]),
  screened: new Set(["enrolled", "ineligible"]),
  enrolled: new Set(["active", "withdrawn"]),
  active: new Set(["graduated", "churned", "withdrawn"]),
};

export const TERMINAL_STATES: Set<string> = new Set(["ineligible", "withdrawn", "graduated", "churned"]);

/**
 * Tracks patients moving through a structured clinical enrollment pipeline.
 *
 * You choose the internal data structures — the public interface is what matters.
 *
 * Store all state in instance properties initialized in the constructor.
 * Class-level (static) fields will bleed between tests and between EnrollmentPipeline
 * instances — avoid them.
 */
export class EnrollmentPipeline {
  constructor() {
    throw new Error("Not implemented");
  }

  // ── Part 1: State tracking ────────────────────────────────────────────────

  /**
   * Register a patient in the pipeline at the "referred" state.
   * timestamp is when they entered the "referred" state (Unix seconds).
   * Throws an Error if patientId is already registered.
   */
  addPatient(patientId: string, timestamp: number = 0.0): void {
    throw new Error("Not implemented");
  }

  /**
   * Advance a patient to newState at the given timestamp.
   *
   * Throws an Error if:
   *   - patientId is not registered.
   *   - newState is not a valid next state from the patient's current state
   *     (consult ALLOWED_TRANSITIONS).
   *   - the patient is already in a terminal state.
   */
  transition(patientId: string, newState: string, timestamp: number): void {
    throw new Error("Not implemented");
  }

  /** Return the patient's current state. Throws an Error if not registered. */
  getState(patientId: string): string {
    throw new Error("Not implemented");
  }

  /** Return a sorted list of patientIds currently in the given state. */
  getPatientsInState(state: string): string[] {
    throw new Error("Not implemented");
  }

  // ── Part 2: Duration and conversion metrics ───────────────────────────────

  /**
   * Return the total seconds the patient has spent in the given state.
   *
   * - If the patient is currently in that state, count time from state entry
   *   up to asOf.
   * - If the patient has already left that state, return the exact duration
   *   spent there (asOf is ignored).
   * - Returns 0.0 if the patient has never been in that state.
   *
   * With the allowed transitions above, each state is visited at most once,
   * so there is no ambiguity about multiple visits.
   */
  timeInState(patientId: string, state: string, asOf: number): number {
    throw new Error("Not implemented");
  }

  /**
   * Of all patients who have exited fromState, return the fraction that
   * transitioned directly to toState.
   *
   * - Only patients who have already left fromState are counted; patients
   *   currently sitting in fromState are excluded (still undecided).
   * - Returns 0.0 if no patients have exited fromState yet.
   *
   * Example: conversionRate("screened", "enrolled") returns the share of
   * screened patients who went on to enroll (vs. being marked ineligible).
   */
  conversionRate(fromState: string, toState: string): number {
    throw new Error("Not implemented");
  }

  // ── Part 3: SLA monitoring ────────────────────────────────────────────────

  /**
   * Return patientIds currently in state who have spent more than
   * maxSeconds there, sorted by time spent descending (longest-waiting first).
   *
   * Call timeInState() for each patient's duration — do not re-implement
   * the duration logic here.
   */
  patientsOverdue(state: string, maxSeconds: number, asOf: number): string[] {
    throw new Error("Not implemented");
  }

  /**
   * Return the mean seconds spent in state across all patients who have fully
   * exited that state (their time is complete and will not grow further).
   *
   * - Patients currently in state are excluded from the average.
   * - Returns 0.0 if no patients have fully exited state yet.
   *
   * Call timeInState() for each patient's duration.
   */
  averageTimeInState(state: string, asOf: number): number {
    throw new Error("Not implemented");
  }
}
