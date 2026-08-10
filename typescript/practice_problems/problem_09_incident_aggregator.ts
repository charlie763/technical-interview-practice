/**
 * =============================================================================
 * INTERVIEW PROBLEM 9: Multi-Source Incident Aggregator
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the incident aggregation layer for a public safety intelligence
 * platform. The platform ingests incident reports from multiple independent data
 * sources (radio dispatch transcriptions, field sensors, social media monitors).
 * Different sources frequently report the same real-world event, so the system
 * must deduplicate and group raw reports into unified incidents.
 *
 * For this problem you are building an IncidentAggregator class.
 * Store all state in instance properties initialized in the constructor.
 * Class-level (static) fields will bleed between tests and between
 * IncidentAggregator instances — avoid them.
 * You choose the internal data structures; the public interface is what matters.
 *
 * DATA MODEL
 * ----------
 * Report:
 *   {
 *     reportId:    string,
 *     sourceId:    string,
 *     eventType:   string,        // e.g. "shooting", "car-crash", "fire"
 *     locationKey: string,        // opaque string, e.g. "downtown", "sector-7"
 *     ts:          string,        // ISO-8601 timestamp (no timezone offset),
 *                                  // e.g. "2024-01-01T10:00:00"
 *     incidentId:  string | undefined  // undefined until assigned to an incident
 *   }
 *
 * Incident:
 *   {
 *     incidentId:   string,
 *     eventType:    string,        // set at creation time
 *     locationKey:  string,        // set at creation time
 *     reportIds:    string[],      // report IDs in ts-ascending order
 *     reportCount:  number,
 *     latestTs:     string | undefined  // ts of the most recently added report, or undefined
 *   }
 *
 * Timestamps are ISO-8601 strings without timezone offset. Use
 * `new Date(ts).getTime()` for arithmetic when comparing or computing durations
 * — plain string comparison also sorts them correctly since they're zero-padded.
 *
 * # Example
 * const agg = new IncidentAggregator();
 * agg.ingestReport("r1", "radio-north", "shooting", "downtown", "2024-01-01T10:00:00");
 * agg.ingestReport("r2", "radio-south", "shooting", "downtown", "2024-01-01T10:00:45");
 * agg.ingestReport("r3", "social-feed", "car-crash", "midtown",  "2024-01-01T10:01:00");
 * const inc = agg.createIncident("inc-001", "shooting", "downtown");
 * agg.addReportToIncident("inc-001", "r1");
 * agg.addReportToIncident("inc-001", "r2");
 * agg.getIncident("inc-001")!.reportCount;           // -> 2
 * agg.getUnassignedReports();                         // -> [r3 report]
 * agg.autoIngestReport("r4", "radio-east", "shooting",
 *                       "downtown", "2024-01-01T10:01:30",
 *                       120);
 * // -> "inc-001"  (within 120 s window, same type + location)
 * =============================================================================
 */

export type Report = {
  reportId: string;
  sourceId: string;
  eventType: string;
  locationKey: string;
  ts: string;
  incidentId: string | undefined;
};

export type Incident = {
  incidentId: string;
  eventType: string;
  locationKey: string;
  reportIds: string[];
  reportCount: number;
  latestTs: string | undefined;
};

export class IncidentAggregator {
  constructor() {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 1 — Report ingestion  (~10 min)
  // ---------------------------------------------------------------------------

  /**
   * Store a new raw report and return it.
   * The report's incidentId starts as undefined.
   * Throws an Error if reportId already exists.
   */
  ingestReport(reportId: string, sourceId: string, eventType: string, locationKey: string, ts: string): Report {
    throw new Error("Not implemented");
  }

  /** Return the report, or undefined if not found. */
  getReport(reportId: string): Report | undefined {
    throw new Error("Not implemented");
  }

  /**
   * Return all reports, optionally filtered by locationKey and/or
   * eventType (both filters applied when both are given).
   * Results are sorted by ts ascending.
   */
  getReports(filters?: { locationKey?: string; eventType?: string }): Report[] {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 2 — Manual incident grouping  (~15 min)
  // ---------------------------------------------------------------------------

  /**
   * Create and return a new, empty incident with the given eventType and
   * locationKey.
   * Throws an Error if incidentId already exists.
   */
  createIncident(incidentId: string, eventType: string, locationKey: string): Incident {
    throw new Error("Not implemented");
  }

  /**
   * Assign a report to an incident.
   * - Throws an Error if incidentId or reportId does not exist.
   * - Throws an Error if the report is already assigned to any incident.
   * - Sets report.incidentId = incidentId.
   * - Updates the incident's reportIds (kept in ts-ascending order),
   *   reportCount, and latestTs.
   */
  addReportToIncident(incidentId: string, reportId: string): void {
    throw new Error("Not implemented");
  }

  /**
   * Return the incident (including up-to-date reportIds, reportCount,
   * and latestTs), or undefined if not found.
   * reportIds must be ordered by the corresponding report's ts, ascending.
   */
  getIncident(incidentId: string): Incident | undefined {
    throw new Error("Not implemented");
  }

  /**
   * Return all reports whose incidentId is still undefined, sorted by ts
   * ascending.
   */
  getUnassignedReports(): Report[] {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 3 — Automatic deduplication  (~20 min)
  // ---------------------------------------------------------------------------

  /**
   * Ingest a new report and automatically assign it to an incident:
   *
   * 1. Call ingestReport to store the report.
   * 2. Find all *active* incidents whose eventType and locationKey match
   *    the incoming report's. An incident is "active" if its latestTs is
   *    within timeWindowSecs of the new report's ts:
   *        latestTs >= ts - timeWindowSecs
   *    Incidents with no reports (latestTs is undefined) are not active.
   * 3. If one or more matches exist, pick the one whose latestTs is closest
   *    to ts (i.e. most recently active). Break ties by incidentId
   *    lexicographically ascending.
   * 4. If no active match exists, create a new incident (auto-generate a
   *    unique incidentId; any scheme is fine as long as it doesn't clash
   *    with existing IDs).
   * 5. Call addReportToIncident to assign the report.
   * 6. Return the incidentId.
   */
  autoIngestReport(reportId: string, sourceId: string, eventType: string, locationKey: string, ts: string, timeWindowSecs: number): string {
    throw new Error("Not implemented");
  }

  /**
   * Return all incidents that have a latestTs within timeWindowSecs of
   * asOfTs:
   *     latestTs >= asOfTs - timeWindowSecs
   *
   * Sorted by latestTs descending (most-recently-active first).
   * Incidents with no reports (latestTs is undefined) are excluded.
   */
  getActiveIncidents(asOfTs: string, timeWindowSecs: number): Incident[] {
    throw new Error("Not implemented");
  }
}
