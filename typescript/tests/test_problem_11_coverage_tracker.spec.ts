/**
 * Tests for Problem 11: Sensor Coverage Tracker
 *
 * Run from the typescript/ directory:
 *   npm run test:11
 */

import { describe, expect, it } from "vitest";
import { CoverageTracker } from "@problems/problem_11_coverage_tracker";

// ---------------------------------------------------------------------------
// Shared timestamps
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

/** Empty CoverageTracker. */
function freshCt(): CoverageTracker {
  return new CoverageTracker();
}

/**
 * Pre-seeded tracker:
 *   sta-seed-1  North Tower  downtown  last hb: T0
 *   sta-seed-2  South Tower  downtown  last hb: T1
 *   sta-seed-3  East Hub     eastside  (never sent a heartbeat)
 */
function makeCt(): CoverageTracker {
  const c = new CoverageTracker();
  c.registerStation("sta-seed-1", "North Tower", "downtown");
  c.registerStation("sta-seed-2", "South Tower", "downtown");
  c.registerStation("sta-seed-3", "East Hub", "eastside");
  c.recordHeartbeat("sta-seed-1", T0);
  c.recordHeartbeat("sta-seed-2", T1);
  // sta-seed-3 intentionally has no heartbeat
  return c;
}

// ---------------------------------------------------------------------------
// PART 1 — Station registration and heartbeats
// ---------------------------------------------------------------------------

describe("registerStation", () => {
  it("stores and returns station", () => {
    const fresh = freshCt();
    const s = fresh.registerStation("sta_reg_test", "Tower", "north");
    expect(s.stationId).toBe("sta_reg_test");
    expect(s.name).toBe("Tower");
    expect(s.region).toBe("north");
  });

  it("duplicate throws", () => {
    const ct = makeCt();
    expect(() => ct.registerStation("sta-seed-1", "Dup", "downtown")).toThrow();
  });
});

describe("recordHeartbeat", () => {
  it("updates last heartbeat", () => {
    const ct = makeCt();
    ct.recordHeartbeat("sta-seed-1", T2);
    expect(ct.getLastHeartbeat("sta-seed-1")).toBe(T2);
  });

  it("missing station throws", () => {
    const ct = makeCt();
    expect(() => ct.recordHeartbeat("ghost", T0)).toThrow();
  });

  it("same ts throws", () => {
    // sta-seed-1 last hb is T0; equal ts should be rejected
    const ct = makeCt();
    expect(() => ct.recordHeartbeat("sta-seed-1", T0)).toThrow();
  });

  it("earlier ts throws", () => {
    // sta-seed-2 last hb is T1; T0 < T1 should be rejected
    const ct = makeCt();
    expect(() => ct.recordHeartbeat("sta-seed-2", T0)).toThrow();
  });

  it("first heartbeat accepted", () => {
    // sta-seed-3 has never sent one
    const ct = makeCt();
    ct.recordHeartbeat("sta-seed-3", T0);
    expect(ct.getLastHeartbeat("sta-seed-3")).toBe(T0);
  });
});

describe("getLastHeartbeat", () => {
  it("returns undefined if never sent", () => {
    const ct = makeCt();
    expect(ct.getLastHeartbeat("sta-seed-3")).toBeUndefined();
  });

  it("returns latest ts", () => {
    const ct = makeCt();
    expect(ct.getLastHeartbeat("sta-seed-1")).toBe(T0);
  });

  it("missing station throws", () => {
    const ct = makeCt();
    expect(() => ct.getLastHeartbeat("ghost")).toThrow();
  });
});

describe("getStations", () => {
  it("no filter returns all", () => {
    const ct = makeCt();
    expect(ct.getStations()).toHaveLength(3);
  });

  it("filter by region", () => {
    const ct = makeCt();
    const downtown = ct.getStations("downtown");
    expect(downtown).toHaveLength(2);
    expect(downtown.every((s) => s.region === "downtown")).toBe(true);
  });

  it("unknown region returns empty", () => {
    const ct = makeCt();
    expect(ct.getStations("nowhere")).toEqual([]);
  });

  it("sorted by stationId", () => {
    const ct = makeCt();
    const stations = ct.getStations();
    const ids = stations.map((s) => s.stationId);
    expect(ids).toEqual([...ids].sort());
  });
});

// ---------------------------------------------------------------------------
// PART 2 — Staleness detection and outage tracking
// ---------------------------------------------------------------------------

describe("getStaleStations", () => {
  it("never heartbeat is stale", () => {
    const ct = makeCt();
    const staleIds = new Set(ct.getStaleStations(T2, 120).map((s) => s.stationId));
    expect(staleIds.has("sta-seed-3")).toBe(true);
  });

  it("recent heartbeat not stale", () => {
    // sta-seed-2 hb=T1; T2-T1=60s <= 120s → not stale
    const ct = makeCt();
    const staleIds = new Set(ct.getStaleStations(T2, 120).map((s) => s.stationId));
    expect(staleIds.has("sta-seed-2")).toBe(false);
  });

  it("old heartbeat is stale", () => {
    // sta-seed-1 hb=T0; T3-T0=300s > 120s → stale
    const ct = makeCt();
    const staleIds = new Set(ct.getStaleStations(T3, 120).map((s) => s.stationId));
    expect(staleIds.has("sta-seed-1")).toBe(true);
  });

  it("all stale when threshold is tiny", () => {
    // 5s threshold: all stations stale at T4
    const ct = makeCt();
    const stale = ct.getStaleStations(T4, 5);
    expect(stale).toHaveLength(3);
  });

  it("sorted by stationId", () => {
    const ct = makeCt();
    const stale = ct.getStaleStations(T4, 5);
    const ids = stale.map((s) => s.stationId);
    expect(ids).toEqual([...ids].sort());
  });
});

describe("recordOutageStart / recordOutageEnd", () => {
  it("open outage recorded", () => {
    const fresh = freshCt();
    fresh.registerStation("sta_out_test", "T", "north");
    fresh.recordOutageStart("sta_out_test", T0);
    const outages = fresh.getOutages("sta_out_test");
    expect(outages).toHaveLength(1);
    expect(outages[0].startTs).toBe(T0);
    expect(outages[0].endTs).toBeUndefined();
  });

  it("duplicate open outage throws", () => {
    const fresh = freshCt();
    fresh.registerStation("sta_dup_out", "T", "north");
    fresh.recordOutageStart("sta_dup_out", T0);
    expect(() => fresh.recordOutageStart("sta_dup_out", T1)).toThrow();
  });

  it("end closes outage", () => {
    const fresh = freshCt();
    fresh.registerStation("sta_end_out", "T", "north");
    fresh.recordOutageStart("sta_end_out", T0);
    fresh.recordOutageEnd("sta_end_out", T1);
    const outages = fresh.getOutages("sta_end_out");
    expect(outages[0].endTs).toBe(T1);
  });

  it("end with no open outage throws", () => {
    // sta-seed-1 has no open outage
    const ct = makeCt();
    expect(() => ct.recordOutageEnd("sta-seed-1", T2)).toThrow();
  });

  it("start on missing station throws", () => {
    const ct = makeCt();
    expect(() => ct.recordOutageStart("ghost", T0)).toThrow();
  });

  it("end on missing station throws", () => {
    const ct = makeCt();
    expect(() => ct.recordOutageEnd("ghost", T0)).toThrow();
  });

  it("second outage allowed after first closed", () => {
    const fresh = freshCt();
    fresh.registerStation("sta_2nd_out", "T", "north");
    fresh.recordOutageStart("sta_2nd_out", T0);
    fresh.recordOutageEnd("sta_2nd_out", T1);
    fresh.recordOutageStart("sta_2nd_out", T2); // should not throw
    const outages = fresh.getOutages("sta_2nd_out");
    expect(outages).toHaveLength(2);
  });

  it("outages sorted by startTs", () => {
    const fresh = freshCt();
    fresh.registerStation("sta_sort_out", "T", "north");
    fresh.recordOutageStart("sta_sort_out", T0);
    fresh.recordOutageEnd("sta_sort_out", T1);
    fresh.recordOutageStart("sta_sort_out", T2);
    fresh.recordOutageEnd("sta_sort_out", T3);
    const outages = fresh.getOutages("sta_sort_out");
    expect(outages[0].startTs).toBe(T0);
    expect(outages[1].startTs).toBe(T2);
  });

  it("getOutages missing station throws", () => {
    const ct = makeCt();
    expect(() => ct.getOutages("ghost")).toThrow();
  });
});

// ---------------------------------------------------------------------------
// PART 3 — Coverage analysis
// ---------------------------------------------------------------------------

describe("getRegionCoverage", () => {
  it("basic partial coverage", () => {
    // asOf=T2, threshold=90s
    // sta-seed-1: T2-T0=120s > 90s → stale
    // sta-seed-2: T2-T1=60s  ≤ 90s → healthy
    const ct = makeCt();
    const result = ct.getRegionCoverage("downtown", T2, 90);
    expect(result.region).toBe("downtown");
    expect(result.total).toBe(2);
    expect(result.healthy).toBe(1);
    expect(result.stale).toBe(1);
    expect(result.hasCoverage).toBe(true);
  });

  it("full coverage", () => {
    // asOf=T2, threshold=300s → both hbs are within window
    const ct = makeCt();
    const result = ct.getRegionCoverage("downtown", T2, 300);
    expect(result.healthy).toBe(2);
    expect(result.stale).toBe(0);
    expect(result.hasCoverage).toBe(true);
  });

  it("no coverage all stale", () => {
    // asOf=T4, threshold=30s → both hbs are way older than 30s
    const ct = makeCt();
    const result = ct.getRegionCoverage("downtown", T4, 30);
    expect(result.healthy).toBe(0);
    expect(result.hasCoverage).toBe(false);
  });

  it("empty region returns zeros", () => {
    const fresh = freshCt();
    const result = fresh.getRegionCoverage("ghost-region", T0, 60);
    expect(result.total).toBe(0);
    expect(result.healthy).toBe(0);
    expect(result.hasCoverage).toBe(false);
  });

  it("correct total for region", () => {
    const ct = makeCt();
    const result = ct.getRegionCoverage("eastside", T0, 60);
    expect(result.total).toBe(1);
  });
});

describe("getOutageSummary", () => {
  it("no outages", () => {
    const ct = makeCt();
    const summary = ct.getOutageSummary("sta-seed-1", T4);
    expect(summary.stationId).toBe("sta-seed-1");
    expect(summary.totalOutages).toBe(0);
    expect(summary.openOutage).toBe(false);
    expect(summary.totalOutageSecs).toBe(0);
  });

  it("closed outage duration", () => {
    const fresh = freshCt();
    fresh.registerStation("sta_dur_test", "T", "north");
    fresh.recordOutageStart("sta_dur_test", T0);
    fresh.recordOutageEnd("sta_dur_test", T2); // T2 - T0 = 120s
    const summary = fresh.getOutageSummary("sta_dur_test", T3);
    expect(summary.totalOutages).toBe(1);
    expect(summary.openOutage).toBe(false);
    expect(summary.totalOutageSecs).toBe(120);
  });

  it("open outage counts to asOf", () => {
    const fresh = freshCt();
    fresh.registerStation("sta_open_test", "T", "north");
    fresh.recordOutageStart("sta_open_test", T0);
    // T3 - T0 = 300s
    const summary = fresh.getOutageSummary("sta_open_test", T3);
    expect(summary.openOutage).toBe(true);
    expect(summary.totalOutageSecs).toBe(300);
  });

  it("multiple closed outages cumulative", () => {
    const fresh = freshCt();
    fresh.registerStation("sta_cumul_test", "T", "north");
    // Outage 1: T0 → T1 = 60s
    fresh.recordOutageStart("sta_cumul_test", T0);
    fresh.recordOutageEnd("sta_cumul_test", T1);
    // Outage 2: T2 → T3 = 180s
    fresh.recordOutageStart("sta_cumul_test", T2);
    fresh.recordOutageEnd("sta_cumul_test", T3);
    const summary = fresh.getOutageSummary("sta_cumul_test", T4);
    expect(summary.totalOutages).toBe(2);
    expect(summary.openOutage).toBe(false);
    expect(summary.totalOutageSecs).toBe(60 + 180); // 240s total
  });

  it("missing station throws", () => {
    const ct = makeCt();
    expect(() => ct.getOutageSummary("ghost", T0)).toThrow();
  });
});
