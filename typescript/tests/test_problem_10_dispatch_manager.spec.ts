/**
 * Tests for Problem 10: Responder Dispatch Manager
 *
 * Run from the typescript/ directory:
 *   npm run test:10
 */

import { describe, expect, it } from "vitest";
import { DispatchManager } from "@problems/problem_10_dispatch_manager";

// ---------------------------------------------------------------------------
// Shared timestamps
// ---------------------------------------------------------------------------
const T0 = "2024-06-01T10:00:00";
const T1 = "2024-06-01T10:01:00";
const T2 = "2024-06-01T10:02:00";

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** Empty DispatchManager. */
function freshDm(): DispatchManager {
  return new DispatchManager();
}

/**
 * Pre-seeded manager:
 *   Responders:
 *     unit-12  (shooting + robbery, capacity=3): has inc-seed-1 open
 *     unit-14  (car-crash + fire,   capacity=2): no assignments
 *   Incidents:
 *     inc-seed-1  shooting  sev=5  T0  → assigned to unit-12
 *     inc-seed-2  shooting  sev=3  T1  → unassigned
 *     inc-seed-3  car-crash sev=4  T2  → unassigned
 */
function makeDm(): DispatchManager {
  const d = new DispatchManager();
  d.registerResponder("unit-12", "Alpha Team", ["shooting", "robbery"], 3);
  d.registerResponder("unit-14", "Beta Team", ["car-crash", "fire"], 2);
  d.addIncident("inc-seed-1", "shooting", 5, T0);
  d.addIncident("inc-seed-2", "shooting", 3, T1);
  d.addIncident("inc-seed-3", "car-crash", 4, T2);
  d.assignIncident("inc-seed-1", "unit-12");
  return d;
}

// ---------------------------------------------------------------------------
// PART 1 — Registration and basic queries
// ---------------------------------------------------------------------------

describe("registerResponder", () => {
  it("stores and returns responder", () => {
    const fresh = freshDm();
    const r = fresh.registerResponder("unit_reg_test", "Gamma", ["fire"], 2);
    expect(r.responderId).toBe("unit_reg_test");
    expect(r.name).toBe("Gamma");
    expect(r.subscribedTypes).toEqual(["fire"]);
    expect(r.capacity).toBe(2);
  });

  it("duplicate throws", () => {
    const dm = makeDm();
    expect(() => dm.registerResponder("unit-12", "Duplicate", ["fire"], 1)).toThrow();
  });
});

describe("addIncident", () => {
  it("stores and returns incident", () => {
    const fresh = freshDm();
    const inc = fresh.addIncident("inc_add_test", "fire", 2, T0);
    expect(inc.incidentId).toBe("inc_add_test");
    expect(inc.incidentType).toBe("fire");
    expect(inc.severity).toBe(2);
    expect(inc.responderId).toBeUndefined();
    expect(inc.resolved).toBe(false);
  });

  it("duplicate throws", () => {
    const dm = makeDm();
    expect(() => dm.addIncident("inc-seed-1", "fire", 1, T0)).toThrow();
  });
});

describe("getIncidentsForResponder", () => {
  it("returns subscribed types only", () => {
    const dm = makeDm();
    const incidents = dm.getIncidentsForResponder("unit-12");
    expect(incidents.every((i) => ["shooting", "robbery"].includes(i.incidentType))).toBe(true);
  });

  it("sorted severity desc then ts asc", () => {
    // inc-seed-1 sev=5, inc-seed-2 sev=3 — both are shooting
    const dm = makeDm();
    const incidents = dm.getIncidentsForResponder("unit-12");
    const ids = incidents.map((i) => i.incidentId);
    expect(ids[0]).toBe("inc-seed-1"); // higher severity first
    expect(ids[1]).toBe("inc-seed-2");
  });

  it("same severity sorted by ts", () => {
    const fresh = freshDm();
    fresh.registerResponder("u_ts_test", "T", ["fire"], 5);
    fresh.addIncident("inc_ts_early", "fire", 3, T0);
    fresh.addIncident("inc_ts_late", "fire", 3, T1);
    const incidents = fresh.getIncidentsForResponder("u_ts_test");
    const ids = incidents.map((i) => i.incidentId);
    expect(ids).toEqual(["inc_ts_early", "inc_ts_late"]);
  });

  it("missing responder throws", () => {
    const dm = makeDm();
    expect(() => dm.getIncidentsForResponder("ghost")).toThrow();
  });
});

// ---------------------------------------------------------------------------
// PART 2 — Assignment and resolution
// ---------------------------------------------------------------------------

describe("assignIncident", () => {
  it("sets responderId", () => {
    const dm = makeDm();
    dm.assignIncident("inc-seed-2", "unit-12");
    const openIds = new Set(dm.getOpenAssignments("unit-12").map((i) => i.incidentId));
    expect(openIds.has("inc-seed-2")).toBe(true);
  });

  it("missing incident throws", () => {
    const dm = makeDm();
    expect(() => dm.assignIncident("ghost-inc", "unit-12")).toThrow();
  });

  it("missing responder throws", () => {
    const dm = makeDm();
    expect(() => dm.assignIncident("inc-seed-2", "ghost-unit")).toThrow();
  });

  it("already assigned throws", () => {
    // inc-seed-1 is already assigned to unit-12
    const dm = makeDm();
    expect(() => dm.assignIncident("inc-seed-1", "unit-12")).toThrow();
  });

  it("at capacity throws", () => {
    const fresh = freshDm();
    fresh.registerResponder("cap_unit", "Cap", ["fire"], 1);
    fresh.addIncident("cap_inc_1", "fire", 1, T0);
    fresh.addIncident("cap_inc_2", "fire", 1, T1);
    fresh.assignIncident("cap_inc_1", "cap_unit");
    expect(() => fresh.assignIncident("cap_inc_2", "cap_unit")).toThrow();
  });
});

describe("resolveIncident", () => {
  it("marks resolved and removes from open", () => {
    const dm = makeDm();
    dm.resolveIncident("inc-seed-1");
    const openIds = new Set(dm.getOpenAssignments("unit-12").map((i) => i.incidentId));
    expect(openIds.has("inc-seed-1")).toBe(false);
  });

  it("already resolved throws", () => {
    const dm = makeDm();
    dm.resolveIncident("inc-seed-1");
    expect(() => dm.resolveIncident("inc-seed-1")).toThrow();
  });

  it("missing throws", () => {
    const dm = makeDm();
    expect(() => dm.resolveIncident("ghost")).toThrow();
  });

  it("frees capacity for next assignment", () => {
    const fresh = freshDm();
    fresh.registerResponder("cap2_unit", "Cap2", ["fire"], 1);
    fresh.addIncident("cap2_inc_1", "fire", 1, T0);
    fresh.addIncident("cap2_inc_2", "fire", 1, T1);
    fresh.assignIncident("cap2_inc_1", "cap2_unit");
    fresh.resolveIncident("cap2_inc_1");
    // Should no longer throw — capacity was freed
    expect(() => fresh.assignIncident("cap2_inc_2", "cap2_unit")).not.toThrow();
  });
});

describe("getOpenAssignments", () => {
  it("returns open assignments", () => {
    const dm = makeDm();
    const openList = dm.getOpenAssignments("unit-12");
    expect(openList).toHaveLength(1);
    expect(openList[0].incidentId).toBe("inc-seed-1");
  });

  it("resolved not in open", () => {
    const dm = makeDm();
    dm.resolveIncident("inc-seed-1");
    expect(dm.getOpenAssignments("unit-12")).toEqual([]);
  });

  it("missing responder throws", () => {
    const dm = makeDm();
    expect(() => dm.getOpenAssignments("ghost")).toThrow();
  });

  it("sorted severity desc then ts asc", () => {
    const fresh = freshDm();
    fresh.registerResponder("u_sort", "S", ["fire"], 5);
    fresh.addIncident("inc_sort_low", "fire", 2, T0);
    fresh.addIncident("inc_sort_high", "fire", 5, T1);
    fresh.assignIncident("inc_sort_low", "u_sort");
    fresh.assignIncident("inc_sort_high", "u_sort");
    const openList = fresh.getOpenAssignments("u_sort");
    expect(openList[0].incidentId).toBe("inc_sort_high");
  });
});

// ---------------------------------------------------------------------------
// PART 3 — Auto-assignment
// ---------------------------------------------------------------------------

describe("autoAssign", () => {
  it("assigns to eligible responder", () => {
    // inc-seed-3 is car-crash → only unit-14 is subscribed
    const dm = makeDm();
    const result = dm.autoAssign("inc-seed-3");
    expect(result).toBe("unit-14");
    const openIds = new Set(dm.getOpenAssignments("unit-14").map((i) => i.incidentId));
    expect(openIds.has("inc-seed-3")).toBe(true);
  });

  it("missing incident throws", () => {
    const dm = makeDm();
    expect(() => dm.autoAssign("ghost")).toThrow();
  });

  it("already assigned throws", () => {
    const dm = makeDm();
    expect(() => dm.autoAssign("inc-seed-1")).toThrow(); // already assigned to unit-12
  });

  it("no eligible responder throws", () => {
    const fresh = freshDm();
    fresh.registerResponder("only_unit", "Only", ["shooting"], 1);
    fresh.addIncident("inc_no_sub", "fire", 1, T0);
    expect(() => fresh.autoAssign("inc_no_sub")).toThrow();
  });

  it("full capacity responder excluded", () => {
    const fresh = freshDm();
    fresh.registerResponder("full_unit", "Full", ["fire"], 1);
    fresh.registerResponder("open_unit", "Open", ["fire"], 2);
    fresh.addIncident("inc_cap_fill", "fire", 1, T0);
    fresh.addIncident("inc_cap_new", "fire", 1, T1);
    fresh.assignIncident("inc_cap_fill", "full_unit");
    const result = fresh.autoAssign("inc_cap_new");
    expect(result).toBe("open_unit");
  });

  it("picks least loaded responder", () => {
    const fresh = freshDm();
    fresh.registerResponder("u_loaded", "Loaded", ["fire"], 3);
    fresh.registerResponder("u_free", "Free", ["fire"], 3);
    fresh.addIncident("inc_load_seed", "fire", 1, T0);
    fresh.addIncident("inc_load_new", "fire", 1, T1);
    // Give u_loaded one open incident
    fresh.assignIncident("inc_load_seed", "u_loaded");
    const result = fresh.autoAssign("inc_load_new");
    expect(result).toBe("u_free"); // fewer open assignments
  });

  it("tiebreak by highest capacity", () => {
    // Equal open assignments (0 each); tiebreak → higher capacity wins
    const fresh = freshDm();
    fresh.registerResponder("u_low_cap", "Low", ["fire"], 1);
    fresh.registerResponder("u_high_cap", "High", ["fire"], 5);
    fresh.addIncident("inc_cap_tb", "fire", 1, T0);
    const result = fresh.autoAssign("inc_cap_tb");
    expect(result).toBe("u_high_cap");
  });

  it("calls assignIncident internally (must delegate, not duplicate logic)", () => {
    const fresh = freshDm();
    fresh.registerResponder("u_delegate", "D", ["fire"], 1);
    fresh.addIncident("inc_delegate_1", "fire", 1, T0);
    fresh.addIncident("inc_delegate_2", "fire", 1, T1);
    fresh.autoAssign("inc_delegate_1");
    // Capacity should now be full; assigning again must throw via assignIncident
    expect(() => fresh.autoAssign("inc_delegate_2")).toThrow();
  });
});

describe("getDispatchSummary", () => {
  it("returns all responders", () => {
    const dm = makeDm();
    const summary = dm.getDispatchSummary();
    const ids = new Set(summary.map((s) => s.responderId));
    expect(ids.has("unit-12")).toBe(true);
    expect(ids.has("unit-14")).toBe(true);
  });

  it("sorted by responderId", () => {
    const dm = makeDm();
    const summary = dm.getDispatchSummary();
    const ids = summary.map((s) => s.responderId);
    expect(ids).toEqual([...ids].sort());
  });

  it("openCount and availableCapacity", () => {
    // unit-12 has 1 open assignment; capacity=3
    const dm = makeDm();
    const summary = dm.getDispatchSummary();
    const u12 = summary.find((s) => s.responderId === "unit-12")!;
    expect(u12.openCount).toBe(1);
    expect(u12.availableCapacity).toBe(2);
  });

  it("zero load responder", () => {
    // unit-14 has no assignments
    const dm = makeDm();
    const summary = dm.getDispatchSummary();
    const u14 = summary.find((s) => s.responderId === "unit-14")!;
    expect(u14.openCount).toBe(0);
    expect(u14.availableCapacity).toBe(2);
  });
});
