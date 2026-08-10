/**
 * =============================================================================
 * INTERVIEW PROBLEM 11: Sensor Coverage Tracker
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the health-monitoring subsystem for a platform that deploys
 * radio-receiver sensor stations across geographic regions. Each station
 * periodically sends a heartbeat. When a station falls silent, operators need
 * to know, and the platform's incident-detection coverage for that region may
 * be affected.
 *
 * For this problem you are building a CoverageTracker class.
 * Store all state in instance properties initialized in the constructor.
 * Class-level (static) fields will bleed between tests and between
 * CoverageTracker instances — avoid them.
 * You choose the internal data structures; the public interface is what matters.
 *
 * DATA MODEL
 * ----------
 * Station:
 *   {
 *     stationId: string,
 *     name:      string,
 *     region:    string,  // logical grouping, e.g. "downtown", "sector-7"
 *   }
 *
 * Outage:
 *   {
 *     stationId: string,
 *     startTs:   string,             // ISO-8601 when the outage began
 *     endTs:     string | undefined, // ISO-8601 when the outage ended; undefined = ongoing
 *   }
 *
 * Timestamps are ISO-8601 strings without timezone offset (e.g.
 * "2024-01-01T10:00:00"). Use `new Date(ts).getTime()` for arithmetic — plain
 * string comparison also sorts them correctly since they're zero-padded.
 *
 * # Example
 * const ct = new CoverageTracker();
 * ct.registerStation("sta-001", "North Tower", "downtown");
 * ct.registerStation("sta-002", "South Tower", "downtown");
 * ct.recordHeartbeat("sta-001", "2024-01-01T10:00:00");
 * ct.recordHeartbeat("sta-002", "2024-01-01T10:00:05");
 * ct.getLastHeartbeat("sta-001");              // -> "2024-01-01T10:00:00"
 * ct.getStaleStations("2024-01-01T10:05:00", 120);  // -> []
 * ct.recordOutageStart("sta-001", "2024-01-01T10:10:00");
 * ct.getRegionCoverage("downtown", "2024-01-01T10:10:30", 120);
 * // -> { region: "downtown", total: 2, healthy: 1, stale: 1, hasCoverage: true }
 * =============================================================================
 */

export type Station = {
  stationId: string;
  name: string;
  region: string;
};

export type Outage = {
  stationId: string;
  startTs: string;
  endTs: string | undefined;
};

export type RegionCoverage = {
  region: string;
  total: number;
  healthy: number;
  stale: number;
  hasCoverage: boolean;
};

export type OutageSummary = {
  stationId: string;
  totalOutages: number;
  openOutage: boolean;
  totalOutageSecs: number;
};

export class CoverageTracker {
  constructor() {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 1 — Station registration and heartbeats  (~10 min)
  // ---------------------------------------------------------------------------

  /**
   * Register a new station and return it.
   * Throws an Error if stationId already exists.
   */
  registerStation(stationId: string, name: string, region: string): Station {
    throw new Error("Not implemented");
  }

  /**
   * Record a heartbeat for the station.
   * - Throws an Error if stationId does not exist.
   * - Throws an Error if ts is earlier than or equal to the station's most
   *   recent heartbeat (out-of-order and duplicate heartbeats are rejected).
   */
  recordHeartbeat(stationId: string, ts: string): void {
    throw new Error("Not implemented");
  }

  /**
   * Return the timestamp of the most recent heartbeat, or undefined if the
   * station has never sent one.
   * Throws an Error if stationId does not exist.
   */
  getLastHeartbeat(stationId: string): string | undefined {
    throw new Error("Not implemented");
  }

  /**
   * Return all stations, optionally filtered to a specific region.
   * Sorted by stationId ascending.
   */
  getStations(region?: string): Station[] {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 2 — Staleness detection and outage tracking  (~15 min)
  // ---------------------------------------------------------------------------

  /**
   * Return stations that are stale as of asOfTs.
   * A station is stale if:
   *   - It has never sent a heartbeat, OR
   *   - Its last heartbeat was more than staleAfterSecs seconds before
   *     asOfTs (i.e. asOfTs - lastHeartbeat > staleAfterSecs).
   *
   * Results are sorted by stationId ascending.
   */
  getStaleStations(asOfTs: string, staleAfterSecs: number): Station[] {
    throw new Error("Not implemented");
  }

  /**
   * Open a new outage record for the station (endTs = undefined).
   * - Throws an Error if stationId does not exist.
   * - Throws an Error if the station already has an open outage
   *   (an outage with endTs = undefined).
   */
  recordOutageStart(stationId: string, ts: string): void {
    throw new Error("Not implemented");
  }

  /**
   * Close the most recent open outage for the station by setting its
   * endTs = ts.
   * - Throws an Error if stationId does not exist.
   * - Throws an Error if the station has no open outage.
   */
  recordOutageEnd(stationId: string, ts: string): void {
    throw new Error("Not implemented");
  }

  /**
   * Return all Outage records for the station, sorted by startTs ascending.
   * Throws an Error if stationId does not exist.
   */
  getOutages(stationId: string): Outage[] {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 3 — Coverage analysis  (~20 min)
  // ---------------------------------------------------------------------------

  /**
   * Return a coverage summary for the region:
   *   { region, total, healthy, stale, hasCoverage }
   * where hasCoverage is true if healthy >= 1.
   *
   * Use getStations (Part 1) and getStaleStations (Part 2) internally.
   */
  getRegionCoverage(region: string, asOfTs: string, staleAfterSecs: number): RegionCoverage {
    throw new Error("Not implemented");
  }

  /**
   * Return an outage summary for the station:
   *   { stationId, totalOutages, openOutage, totalOutageSecs }
   *
   * For an open outage (endTs is undefined), count duration from startTs up
   * to asOfTs.
   *
   * Use getOutages (Part 2) internally.
   * Throws an Error if stationId does not exist.
   */
  getOutageSummary(stationId: string, asOfTs: string): OutageSummary {
    throw new Error("Not implemented");
  }
}
