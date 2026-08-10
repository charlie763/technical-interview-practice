/**
 * Tests for Problem 9: Multi-Source Incident Aggregator
 *
 * Run from the typescript/ directory:
 *   npm run test:09
 */

import { describe, expect, it, beforeEach } from "vitest";
import { IncidentAggregator } from "@problems/problem_09_incident_aggregator";

// ---------------------------------------------------------------------------
// Shared timestamps  (all naive ISO-8601, lexicographically sortable)
// T0 = base
// T1 = T0 + 60s
// T2 = T0 + 120s
// T3 = T0 + 300s
// T4 = T0 + 600s
// ---------------------------------------------------------------------------
const T0 = "2024-06-01T10:00:00";
const T1 = "2024-06-01T10:01:00";
const T2 = "2024-06-01T10:02:00";
const T3 = "2024-06-01T10:05:00";
const T4 = "2024-06-01T10:10:00";

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** Empty IncidentAggregator. */
function freshAgg(): IncidentAggregator {
  return new IncidentAggregator();
}

/**
 * Pre-seeded aggregator:
 *   Reports: r1 (shooting/downtown/T0), r2 (shooting/downtown/T1),
 *            r3 (car-crash/midtown/T2)
 *   Incidents: inc-001 (shooting/downtown) containing r1 and r2
 *   r3 is unassigned.
 */
function makeAgg(): IncidentAggregator {
  const a = new IncidentAggregator();
  a.ingestReport("r1", "radio-north", "shooting", "downtown", T0);
  a.ingestReport("r2", "radio-south", "shooting", "downtown", T1);
  a.ingestReport("r3", "social-feed", "car-crash", "midtown", T2);
  a.createIncident("inc-001", "shooting", "downtown");
  a.addReportToIncident("inc-001", "r1");
  a.addReportToIncident("inc-001", "r2");
  return a;
}

// ---------------------------------------------------------------------------
// PART 1 — Report ingestion
// ---------------------------------------------------------------------------

describe("ingestReport", () => {
  let fresh: IncidentAggregator;

  beforeEach(() => {
    fresh = freshAgg();
  });

  it("stores and returns report", () => {
    const r = fresh.ingestReport("r_store", "radio-north", "shooting", "downtown", T0);
    expect(r.reportId).toBe("r_store");
    expect(r.sourceId).toBe("radio-north");
    expect(r.eventType).toBe("shooting");
    expect(r.locationKey).toBe("downtown");
    expect(r.ts).toBe(T0);
    expect(r.incidentId).toBeUndefined();
  });

  it("duplicate report id throws", () => {
    fresh.ingestReport("r_dup", "src-a", "fire", "east", T0);
    expect(() => fresh.ingestReport("r_dup", "src-b", "fire", "east", T1)).toThrow();
  });

  it("getReport returns undefined if missing", () => {
    expect(fresh.getReport("nonexistent")).toBeUndefined();
  });

  it("getReport returns stored", () => {
    const agg = makeAgg();
    const r = agg.getReport("r1");
    expect(r).toBeDefined();
    expect(r!.eventType).toBe("shooting");
  });
});

describe("getReports", () => {
  let agg: IncidentAggregator;

  beforeEach(() => {
    agg = makeAgg();
  });

  it("no filter returns all", () => {
    expect(agg.getReports()).toHaveLength(3);
  });

  it("filter by location", () => {
    const reports = agg.getReports({ locationKey: "downtown" });
    expect(reports).toHaveLength(2);
    expect(reports.every((r) => r.locationKey === "downtown")).toBe(true);
  });

  it("filter by event type", () => {
    const reports = agg.getReports({ eventType: "car-crash" });
    expect(reports).toHaveLength(1);
    expect(reports[0].reportId).toBe("r3");
  });

  it("filter both", () => {
    const reports = agg.getReports({ locationKey: "downtown", eventType: "shooting" });
    expect(reports).toHaveLength(2);
  });

  it("empty when no match", () => {
    expect(agg.getReports({ locationKey: "mars" })).toEqual([]);
  });

  it("sorted by ts ascending", () => {
    const fresh = freshAgg();
    fresh.ingestReport("r_sort_b", "src", "fire", "zone-1", T1);
    fresh.ingestReport("r_sort_a", "src", "fire", "zone-1", T0);
    const reports = fresh.getReports();
    const tss = reports.map((r) => r.ts);
    expect(tss).toEqual([...tss].sort());
  });
});

// ---------------------------------------------------------------------------
// PART 2 — Manual incident grouping
// ---------------------------------------------------------------------------

describe("createIncident", () => {
  it("creates empty incident", () => {
    const fresh = freshAgg();
    const inc = fresh.createIncident("inc_create_test", "fire", "east-side");
    expect(inc.incidentId).toBe("inc_create_test");
    expect(inc.eventType).toBe("fire");
    expect(inc.locationKey).toBe("east-side");
    expect(inc.reportCount).toBe(0);
    expect(inc.reportIds).toEqual([]);
    expect(inc.latestTs).toBeUndefined();
  });

  it("duplicate incident id throws", () => {
    const agg = makeAgg();
    expect(() => agg.createIncident("inc-001", "shooting", "downtown")).toThrow();
  });
});

describe("addReportToIncident", () => {
  it("assigns report", () => {
    const fresh = freshAgg();
    fresh.ingestReport("r_assign", "src", "fire", "east", T0);
    fresh.createIncident("inc_assign_test", "fire", "east");
    fresh.addReportToIncident("inc_assign_test", "r_assign");
    const inc = fresh.getIncident("inc_assign_test")!;
    expect(inc.reportIds).toContain("r_assign");
    expect(inc.reportCount).toBe(1);
    expect(inc.latestTs).toBe(T0);
  });

  it("updates report incidentId", () => {
    const fresh = freshAgg();
    fresh.ingestReport("r_update", "src", "fire", "east", T0);
    fresh.createIncident("inc_update_test", "fire", "east");
    fresh.addReportToIncident("inc_update_test", "r_update");
    expect(fresh.getReport("r_update")!.incidentId).toBe("inc_update_test");
  });

  it("already assigned throws", () => {
    // r1 is already in inc-001
    const agg = makeAgg();
    expect(() => agg.addReportToIncident("inc-001", "r1")).toThrow();
  });

  it("bad incident throws", () => {
    const fresh = freshAgg();
    fresh.ingestReport("r_bad_inc", "src", "fire", "east", T0);
    expect(() => fresh.addReportToIncident("nonexistent_inc", "r_bad_inc")).toThrow();
  });

  it("bad report throws", () => {
    const agg = makeAgg();
    expect(() => agg.addReportToIncident("inc-001", "nonexistent_report")).toThrow();
  });

  it("reportIds sorted by ts", () => {
    const fresh = freshAgg();
    fresh.ingestReport("r_ts_b", "src", "fire", "east", T1);
    fresh.ingestReport("r_ts_a", "src", "fire", "east", T0);
    fresh.createIncident("inc_ts_test", "fire", "east");
    fresh.addReportToIncident("inc_ts_test", "r_ts_b");
    fresh.addReportToIncident("inc_ts_test", "r_ts_a");
    const inc = fresh.getIncident("inc_ts_test")!;
    expect(inc.reportIds).toEqual(["r_ts_a", "r_ts_b"]);
  });

  it("latestTs updates to newest", () => {
    const fresh = freshAgg();
    fresh.ingestReport("r_lt_a", "src", "fire", "east", T0);
    fresh.ingestReport("r_lt_b", "src", "fire", "east", T2);
    fresh.createIncident("inc_lt_test", "fire", "east");
    fresh.addReportToIncident("inc_lt_test", "r_lt_a");
    fresh.addReportToIncident("inc_lt_test", "r_lt_b");
    expect(fresh.getIncident("inc_lt_test")!.latestTs).toBe(T2);
  });
});

describe("getIncident", () => {
  it("returns undefined if not found", () => {
    expect(freshAgg().getIncident("ghost")).toBeUndefined();
  });

  it("returns incident", () => {
    const agg = makeAgg();
    const inc = agg.getIncident("inc-001");
    expect(inc).toBeDefined();
    expect(inc!.reportCount).toBe(2);
  });
});

describe("getUnassignedReports", () => {
  it("returns unassigned", () => {
    const agg = makeAgg();
    const unassigned = agg.getUnassignedReports();
    expect(unassigned).toHaveLength(1);
    expect(unassigned[0].reportId).toBe("r3");
  });

  it("empty when all assigned", () => {
    const fresh = freshAgg();
    fresh.ingestReport("r_all", "src", "fire", "east", T0);
    fresh.createIncident("inc_all_test", "fire", "east");
    fresh.addReportToIncident("inc_all_test", "r_all");
    expect(fresh.getUnassignedReports()).toEqual([]);
  });

  it("sorted by ts ascending", () => {
    const fresh = freshAgg();
    fresh.ingestReport("r_ua_b", "src", "fire", "east", T1);
    fresh.ingestReport("r_ua_a", "src", "fire", "east", T0);
    const unassigned = fresh.getUnassignedReports();
    const tss = unassigned.map((r) => r.ts);
    expect(tss).toEqual([...tss].sort());
  });
});

// ---------------------------------------------------------------------------
// PART 3 — Automatic deduplication
// ---------------------------------------------------------------------------

describe("autoIngestReport", () => {
  it("creates new incident when no match", () => {
    const fresh = freshAgg();
    const incidentId = fresh.autoIngestReport("r_auto_new", "src", "fire", "east-side", T0, 120);
    expect(incidentId).toBeDefined();
    const inc = fresh.getIncident(incidentId)!;
    expect(inc.reportCount).toBe(1);
    expect(inc.reportIds).toContain("r_auto_new");
  });

  it("merges into existing when within window", () => {
    // inc-001 latestTs = T1; new ts = T2; T2 - T1 = 60s <= 120s
    const agg = makeAgg();
    const result = agg.autoIngestReport("r_merge", "radio-east", "shooting", "downtown", T2, 120);
    expect(result).toBe("inc-001");
    expect(agg.getIncident("inc-001")!.reportCount).toBe(3);
  });

  it("no merge on different event type", () => {
    const agg = makeAgg();
    const result = agg.autoIngestReport("r_diff_type", "src", "car-crash", "downtown", T2, 120);
    expect(result).not.toBe("inc-001");
  });

  it("no merge on different location", () => {
    const agg = makeAgg();
    const result = agg.autoIngestReport("r_diff_loc", "src", "shooting", "uptown", T2, 120);
    expect(result).not.toBe("inc-001");
  });

  it("no merge outside window", () => {
    // inc-001 latestTs = T1; new ts = T4; T4 - T1 = 540s > 120s
    const agg = makeAgg();
    const result = agg.autoIngestReport("r_expired", "src", "shooting", "downtown", T4, 120);
    expect(result).not.toBe("inc-001");
  });

  it("new report stored", () => {
    const fresh = freshAgg();
    const incidentId = fresh.autoIngestReport("r_stored", "src", "fire", "east", T0, 60);
    expect(fresh.getReport("r_stored")).toBeDefined();
    void incidentId;
  });

  it("picks most recently active match", () => {
    // Two incidents for the same type+location, both within the window
    const fresh = freshAgg();
    fresh.ingestReport("ra1", "src", "shooting", "downtown", T0);
    fresh.ingestReport("ra2", "src", "shooting", "downtown", T1);
    fresh.createIncident("inc_older", "shooting", "downtown");
    fresh.addReportToIncident("inc_older", "ra1"); // latestTs = T0
    fresh.createIncident("inc_newer", "shooting", "downtown");
    fresh.addReportToIncident("inc_newer", "ra2"); // latestTs = T1

    // T2 - T0 = 120s <= 300s and T2 - T1 = 60s <= 300s → both active
    const result = fresh.autoIngestReport("r_pick", "src", "shooting", "downtown", T2, 300);
    // inc_newer is more recently active (T1 > T0)
    expect(result).toBe("inc_newer");
  });

  it("auto-generated id does not clash", () => {
    const fresh = freshAgg();
    const id1 = fresh.autoIngestReport("r_id1", "src", "fire", "west", T0, 0);
    const id2 = fresh.autoIngestReport("r_id2", "src", "fire", "east", T1, 0);
    expect(id1).not.toBe(id2);
  });
});

describe("getActiveIncidents", () => {
  it("returns active incidents", () => {
    // inc-001 latestTs=T1; T2-T1=60s <= 120s → active
    const agg = makeAgg();
    const active = agg.getActiveIncidents(T2, 120);
    expect(active.some((i) => i.incidentId === "inc-001")).toBe(true);
  });

  it("excludes stale incidents", () => {
    // inc-001 latestTs=T1; T4-T1=540s > 120s → not active
    const agg = makeAgg();
    const active = agg.getActiveIncidents(T4, 120);
    expect(active.some((i) => i.incidentId === "inc-001")).toBe(false);
  });

  it("excludes empty incidents", () => {
    const fresh = freshAgg();
    fresh.createIncident("inc_empty_active", "fire", "north");
    const active = fresh.getActiveIncidents(T0, 300);
    expect(active.some((i) => i.incidentId === "inc_empty_active")).toBe(false);
  });

  it("sorted by latestTs descending", () => {
    const fresh = freshAgg();
    fresh.ingestReport("ri1", "src", "fire", "east", T0);
    fresh.ingestReport("ri2", "src", "fire", "west", T2);
    fresh.createIncident("inc_sort_a", "fire", "east");
    fresh.addReportToIncident("inc_sort_a", "ri1"); // latestTs=T0
    fresh.createIncident("inc_sort_b", "fire", "west");
    fresh.addReportToIncident("inc_sort_b", "ri2"); // latestTs=T2

    const active = fresh.getActiveIncidents(T3, 600);
    const ids = active.map((i) => i.incidentId);
    expect(ids.indexOf("inc_sort_b")).toBeLessThan(ids.indexOf("inc_sort_a"));
  });
});
