/**
 * Tests for Problem 1: Geofence Alert Rule Engine
 *
 * Run from the typescript/ directory:
 *   npm run test:01
 */

import { describe, expect, it, beforeEach } from "vitest";
import {
  makeTracker,
  isInZone,
  getCurrentZoneId,
  processLocationUpdate,
  addZone,
  removeZone,
  addAsset,
  addAlertRule,
  type TrackerState,
  type Zone,
  type Asset,
} from "@problems/problem_01_geofence_alert_engine";

// ---------------------------------------------------------------------------
// Shared fixture
// ---------------------------------------------------------------------------

/** A tracker with two adjacent zones and two assets. */
function makeState(): TrackerState {
  const s = makeTracker();
  // Warehouse: lat [35.00, 35.10], lng [-106.70, -106.60]
  addZone(s, "warehouse", "Warehouse A", 35.0, 35.1, -106.7, -106.6);
  // Loading dock: lat [35.10, 35.20], lng [-106.70, -106.60]
  addZone(s, "loading_dock", "Loading Dock", 35.1, 35.2, -106.7, -106.6);
  addAsset(s, "forklift_1", "Forklift #1");
  addAsset(s, "drone_1", "Drone #1");
  return s;
}

// ---------------------------------------------------------------------------
// PART 1 — isInZone
// ---------------------------------------------------------------------------

describe("isInZone", () => {
  const zone = (minLat: number, maxLat: number, minLng: number, maxLng: number): Zone => ({
    id: "z1",
    name: "Z",
    bounds: { minLat, maxLat, minLng, maxLng },
  });

  const asset = (lat: number | undefined, lng: number | undefined): Asset => ({
    id: "a1",
    name: "A",
    lat,
    lng,
    zoneId: undefined,
  });

  it("inside", () => {
    expect(isInZone(asset(35.05, -106.65), zone(35.0, 35.1, -106.7, -106.6))).toBe(true);
  });

  it("on min corner", () => {
    expect(isInZone(asset(35.0, -106.7), zone(35.0, 35.1, -106.7, -106.6))).toBe(true);
  });

  it("on max corner", () => {
    expect(isInZone(asset(35.1, -106.6), zone(35.0, 35.1, -106.7, -106.6))).toBe(true);
  });

  it("outside lat", () => {
    expect(isInZone(asset(35.15, -106.65), zone(35.0, 35.1, -106.7, -106.6))).toBe(false);
  });

  it("outside lng", () => {
    expect(isInZone(asset(35.05, -106.5), zone(35.0, 35.1, -106.7, -106.6))).toBe(false);
  });

  it("no lat", () => {
    expect(isInZone(asset(undefined, -106.65), zone(35.0, 35.1, -106.7, -106.6))).toBe(false);
  });

  it("no lng", () => {
    expect(isInZone(asset(35.05, undefined), zone(35.0, 35.1, -106.7, -106.6))).toBe(false);
  });

  it("both undefined", () => {
    expect(isInZone(asset(undefined, undefined), zone(35.0, 35.1, -106.7, -106.6))).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// PART 2 — getCurrentZoneId
// ---------------------------------------------------------------------------

describe("getCurrentZoneId", () => {
  let state: TrackerState;

  beforeEach(() => {
    state = makeState();
  });

  it("asset in warehouse", () => {
    state.assets["forklift_1"].lat = 35.05;
    state.assets["forklift_1"].lng = -106.65;
    expect(getCurrentZoneId(state, "forklift_1")).toBe("warehouse");
  });

  it("asset in loading dock", () => {
    state.assets["forklift_1"].lat = 35.15;
    state.assets["forklift_1"].lng = -106.65;
    expect(getCurrentZoneId(state, "forklift_1")).toBe("loading_dock");
  });

  it("asset outside all zones", () => {
    state.assets["forklift_1"].lat = 36.0;
    state.assets["forklift_1"].lng = -106.65;
    expect(getCurrentZoneId(state, "forklift_1")).toBeUndefined();
  });

  it("asset has no location", () => {
    expect(getCurrentZoneId(state, "forklift_1")).toBeUndefined();
  });

  it("unknown asset id", () => {
    expect(getCurrentZoneId(state, "ghost")).toBeUndefined();
  });
});

// ---------------------------------------------------------------------------
// PART 3 — processLocationUpdate
// ---------------------------------------------------------------------------

describe("processLocationUpdate", () => {
  let state: TrackerState;

  beforeEach(() => {
    state = makeState();
  });

  it("no alert on first update outside zone", () => {
    const alerts = processLocationUpdate(state, "forklift_1", 36.0, -106.65, "t1");
    expect(alerts).toEqual([]);
    expect(state.assets["forklift_1"].lat).toBe(36.0);
    expect(state.assets["forklift_1"].zoneId).toBeUndefined();
  });

  it("zone entry triggers matching rule", () => {
    addAlertRule(state, "rule_entry", undefined, "warehouse", undefined);
    const alerts = processLocationUpdate(state, "forklift_1", 35.05, -106.65, "t1");
    expect(alerts).toHaveLength(1);
    expect(alerts[0].ruleId).toBe("rule_entry");
    expect(alerts[0].assetId).toBe("forklift_1");
    expect(alerts[0].fromZoneId).toBeUndefined();
    expect(alerts[0].toZoneId).toBe("warehouse");
    expect(alerts[0].timestamp).toBe("t1");
  });

  it("zone exit triggers rule", () => {
    // Pre-position asset in warehouse
    state.assets["forklift_1"].lat = 35.05;
    state.assets["forklift_1"].lng = -106.65;
    state.assets["forklift_1"].zoneId = "warehouse";
    addAlertRule(state, "rule_exit", "warehouse", undefined, undefined);
    const alerts = processLocationUpdate(state, "forklift_1", 36.0, -106.65, "t2");
    expect(alerts).toHaveLength(1);
    expect(alerts[0].fromZoneId).toBe("warehouse");
    expect(alerts[0].toZoneId).toBeUndefined();
  });

  it("zone to zone transition", () => {
    state.assets["forklift_1"].lat = 35.05;
    state.assets["forklift_1"].lng = -106.65;
    state.assets["forklift_1"].zoneId = "warehouse";
    addAlertRule(state, "rule_wh_to_dock", "warehouse", "loading_dock", undefined);
    const alerts = processLocationUpdate(state, "forklift_1", 35.15, -106.65, "t3");
    expect(alerts).toHaveLength(1);
    expect(alerts[0].fromZoneId).toBe("warehouse");
    expect(alerts[0].toZoneId).toBe("loading_dock");
  });

  it("no alert when zone unchanged", () => {
    state.assets["forklift_1"].lat = 35.05;
    state.assets["forklift_1"].lng = -106.65;
    state.assets["forklift_1"].zoneId = "warehouse";
    addAlertRule(state, "rule_any", undefined, undefined, undefined);
    const alerts = processLocationUpdate(state, "forklift_1", 35.06, -106.65, "t4");
    expect(alerts).toEqual([]);
  });

  it("asset-specific rule ignores other assets", () => {
    addAlertRule(state, "rule_drone_only", undefined, "warehouse", "drone_1");
    const alerts = processLocationUpdate(state, "forklift_1", 35.05, -106.65, "t5");
    expect(alerts).toEqual([]);
  });

  it("asset-specific rule fires for correct asset", () => {
    addAlertRule(state, "rule_forklift", undefined, "warehouse", "forklift_1");
    const alerts = processLocationUpdate(state, "forklift_1", 35.05, -106.65, "t6");
    expect(alerts).toHaveLength(1);
  });

  it("multiple matching rules all fire", () => {
    addAlertRule(state, "rule_a", undefined, "warehouse", undefined);
    addAlertRule(state, "rule_b", undefined, undefined, undefined);
    const alerts = processLocationUpdate(state, "forklift_1", 35.05, -106.65, "t7");
    expect(alerts).toHaveLength(2);
  });

  it("alerts appended to log", () => {
    addAlertRule(state, "rule_1", undefined, "warehouse", undefined);
    processLocationUpdate(state, "forklift_1", 35.05, -106.65, "t8");
    expect(state.alertLog).toHaveLength(1);
  });

  it("unknown asset throws", () => {
    expect(() => processLocationUpdate(state, "ghost_asset", 35.05, -106.65, "t9")).toThrow();
  });

  it("lat/lng updated even when no zone change", () => {
    state.assets["forklift_1"].lat = 35.05;
    state.assets["forklift_1"].lng = -106.65;
    state.assets["forklift_1"].zoneId = "warehouse";
    processLocationUpdate(state, "forklift_1", 35.06, -106.64, "t10");
    expect(state.assets["forklift_1"].lat).toBe(35.06);
    expect(state.assets["forklift_1"].lng).toBe(-106.64);
  });
});

// ---------------------------------------------------------------------------
// PART 4 — CRUD helpers
// ---------------------------------------------------------------------------

describe("addZone", () => {
  let state: TrackerState;

  beforeEach(() => {
    state = makeState();
  });

  it("adds a zone", () => {
    const z = addZone(state, "yard", "Yard", 35.3, 35.4, -106.7, -106.6);
    expect(state.zones["yard"]).toEqual(z);
    expect(z.name).toBe("Yard");
  });

  it("duplicate throws", () => {
    expect(() => addZone(state, "warehouse", "Duplicate", 0, 1, 0, 1)).toThrow();
  });
});

describe("removeZone", () => {
  let state: TrackerState;

  beforeEach(() => {
    state = makeState();
  });

  it("removes a zone", () => {
    removeZone(state, "warehouse");
    expect(state.zones["warehouse"]).toBeUndefined();
  });

  it("clears zoneId on assets", () => {
    state.assets["forklift_1"].zoneId = "warehouse";
    removeZone(state, "warehouse");
    expect(state.assets["forklift_1"].zoneId).toBeUndefined();
  });

  it("does not affect assets in other zones", () => {
    state.assets["forklift_1"].zoneId = "loading_dock";
    removeZone(state, "warehouse");
    expect(state.assets["forklift_1"].zoneId).toBe("loading_dock");
  });

  it("missing zone throws", () => {
    expect(() => removeZone(state, "nonexistent")).toThrow();
  });
});

describe("addAsset", () => {
  let state: TrackerState;

  beforeEach(() => {
    state = makeState();
  });

  it("adds an asset", () => {
    const a = addAsset(state, "scanner_1", "Scanner #1");
    expect(state.assets["scanner_1"]).toEqual(a);
    expect(a.lat).toBeUndefined();
    expect(a.lng).toBeUndefined();
    expect(a.zoneId).toBeUndefined();
  });

  it("duplicate throws", () => {
    expect(() => addAsset(state, "forklift_1", "Duplicate")).toThrow();
  });
});

describe("addAlertRule", () => {
  let state: TrackerState;

  beforeEach(() => {
    state = makeState();
  });

  it("adds a rule", () => {
    const r = addAlertRule(state, "r1", "warehouse", "loading_dock", undefined);
    expect(state.alertRules).toContainEqual(r);
    expect(r.id).toBe("r1");
  });

  it("duplicate throws", () => {
    addAlertRule(state, "r1", undefined, undefined, undefined);
    expect(() => addAlertRule(state, "r1", undefined, undefined, undefined)).toThrow();
  });
});
