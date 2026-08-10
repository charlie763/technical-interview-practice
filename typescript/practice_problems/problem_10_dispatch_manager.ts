/**
 * =============================================================================
 * INTERVIEW PROBLEM 10: Responder Dispatch Manager
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the dispatch assignment layer for an emergency-response
 * platform. Incident alerts stream in and need to be routed to available field
 * responders. Responders specialize in certain incident types and have a
 * capacity limit — the maximum number of simultaneous open (unresolved)
 * incidents they can handle.
 *
 * For this problem you are building a DispatchManager class.
 * Store all state in instance properties initialized in the constructor.
 * Class-level (static) fields will bleed between tests and between
 * DispatchManager instances — avoid them.
 * You choose the internal data structures; the public interface is what matters.
 *
 * DATA MODEL
 * ----------
 * Responder:
 *   {
 *     responderId:      string,
 *     name:              string,
 *     subscribedTypes:   string[],  // incident types this responder handles
 *     capacity:          number,    // max simultaneous open assignments
 *   }
 *
 * Incident:
 *   {
 *     incidentId:    string,
 *     incidentType:  string,        // e.g. "shooting", "car-crash", "fire"
 *     severity:      number,        // 1 (low) – 5 (critical)
 *     ts:            string,        // ISO-8601 timestamp, when reported
 *     responderId:   string | undefined,  // undefined until assigned
 *     resolved:      boolean,       // false until resolveIncident is called
 *   }
 *
 * # Example
 * const dm = new DispatchManager();
 * dm.registerResponder("unit-12", "Alpha Team", ["shooting", "robbery"], 3);
 * dm.registerResponder("unit-14", "Beta Team", ["car-crash", "fire"], 2);
 * dm.addIncident("inc-001", "shooting", 5, "2024-01-01T10:00:00");
 * dm.addIncident("inc-002", "car-crash", 3, "2024-01-01T10:01:00");
 * dm.getIncidentsForResponder("unit-12");   // -> [inc-001]
 * dm.assignIncident("inc-001", "unit-12");
 * dm.getOpenAssignments("unit-12");          // -> [inc-001]
 * dm.autoAssign("inc-002");                  // -> "unit-14"
 * =============================================================================
 */

export type Responder = {
  responderId: string;
  name: string;
  subscribedTypes: string[];
  capacity: number;
};

export type Incident = {
  incidentId: string;
  incidentType: string;
  severity: number;
  ts: string;
  responderId: string | undefined;
  resolved: boolean;
};

export type DispatchSummaryEntry = {
  responderId: string;
  name: string;
  capacity: number;
  openCount: number;
  availableCapacity: number;
};

export class DispatchManager {
  constructor() {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 1 — Registration and basic queries  (~10 min)
  // ---------------------------------------------------------------------------

  /**
   * Register a new responder and return it.
   * Throws an Error if responderId already exists.
   */
  registerResponder(responderId: string, name: string, subscribedTypes: string[], capacity: number): Responder {
    throw new Error("Not implemented");
  }

  /**
   * Add a new incident (unassigned, unresolved) and return it.
   * Throws an Error if incidentId already exists.
   */
  addIncident(incidentId: string, incidentType: string, severity: number, ts: string): Incident {
    throw new Error("Not implemented");
  }

  /**
   * Return all incidents whose incidentType appears in the responder's
   * subscribedTypes list, regardless of whether the incident has been
   * assigned yet.
   *
   * Sort order: severity descending (5 first), then ts ascending (oldest
   * first within the same severity).
   *
   * Throws an Error if responderId does not exist.
   */
  getIncidentsForResponder(responderId: string): Incident[] {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 2 — Assignment and resolution  (~15 min)
  // ---------------------------------------------------------------------------

  /**
   * Assign an incident to a responder.
   * - Throws an Error if incidentId or responderId does not exist.
   * - Throws an Error if the incident already has a responder assigned.
   * - Throws an Error if the responder is at capacity. A responder is at
   *   capacity when their count of open assignments (assigned + not yet
   *   resolved) equals their capacity.
   * - On success, sets incident.responderId = responderId.
   */
  assignIncident(incidentId: string, responderId: string): void {
    throw new Error("Not implemented");
  }

  /**
   * Mark an incident as resolved (sets resolved = true), freeing the
   * assigned responder's capacity slot.
   * - Throws an Error if incidentId does not exist.
   * - Throws an Error if the incident is already resolved.
   */
  resolveIncident(incidentId: string): void {
    throw new Error("Not implemented");
  }

  /**
   * Return all incidents that are assigned to this responder and not yet
   * resolved.
   * Sort order: severity descending, then ts ascending.
   * Throws an Error if responderId does not exist.
   */
  getOpenAssignments(responderId: string): Incident[] {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 3 — Auto-assignment  (~20 min)
  // ---------------------------------------------------------------------------

  /**
   * Automatically assign an incident to the best available responder:
   *
   * Eligibility (both must hold):
   *   1. The responder's subscribedTypes includes the incident's
   *      incidentType.
   *   2. The responder's current open-assignment count is less than their
   *      capacity.
   *
   * Selection — among eligible responders, prefer:
   *   1. Fewest open assignments (least loaded).
   *   2. Tie-break: highest capacity (largest capacity value).
   *   3. Tie-break: responderId lexicographically ascending.
   *
   * - Throws an Error if incidentId does not exist.
   * - Throws an Error if the incident is already assigned.
   * - Throws an Error if no eligible responder is available.
   *
   * Call assignIncident to perform the assignment and return the
   * responderId of the chosen responder.
   */
  autoAssign(incidentId: string): string {
    throw new Error("Not implemented");
  }

  /**
   * Return an array with one entry per registered responder:
   *   { responderId, name, capacity, openCount, availableCapacity }
   * where availableCapacity = capacity - openCount.
   *
   * Sorted by responderId ascending.
   */
  getDispatchSummary(): DispatchSummaryEntry[] {
    throw new Error("Not implemented");
  }
}
