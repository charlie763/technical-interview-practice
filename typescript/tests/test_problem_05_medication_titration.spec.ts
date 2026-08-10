/**
 * Tests for Problem 5: Medication Titration Tracker
 *
 * Run from the typescript/ directory:
 *   npm run test:05
 */

import { describe, expect, it, beforeEach } from "vitest";
import { TitrationTracker, makeTitrationEvent, type TitrationEvent } from "@problems/problem_05_medication_titration";

// ---------------------------------------------------------------------------
// Pre-existing patient event histories
// ---------------------------------------------------------------------------

// Maria: successfully de-escalated off metformin, still on low-dose glipizide
const MARIA_EVENTS: TitrationEvent[] = [
  makeTitrationEvent("maria", "metformin", "start", 500.0, "2023-06-01"),
  makeTitrationEvent("maria", "metformin", "increase", 1000.0, "2023-08-01"),
  makeTitrationEvent("maria", "metformin", "decrease", 500.0, "2023-11-01"),
  makeTitrationEvent("maria", "metformin", "stop", 0.0, "2024-02-01"),
  makeTitrationEvent("maria", "glipizide", "start", 5.0, "2023-06-01"),
  makeTitrationEvent("maria", "glipizide", "decrease", 2.5, "2024-01-15"),
];

// James: still on two active medications, one recently increased
const JAMES_EVENTS: TitrationEvent[] = [
  makeTitrationEvent("james", "metformin", "start", 500.0, "2023-09-01"),
  makeTitrationEvent("james", "metformin", "increase", 1000.0, "2023-12-01"),
  makeTitrationEvent("james", "insulin_glargine", "start", 10.0, "2023-09-01"),
  makeTitrationEvent("james", "insulin_glargine", "increase", 15.0, "2024-01-01"),
  makeTitrationEvent("james", "insulin_glargine", "decrease", 10.0, "2024-03-01"),
];

// Susan: completely off all medications
const SUSAN_EVENTS: TitrationEvent[] = [
  makeTitrationEvent("susan", "metformin", "start", 500.0, "2023-03-01"),
  makeTitrationEvent("susan", "metformin", "increase", 750.0, "2023-05-01"),
  makeTitrationEvent("susan", "metformin", "stop", 0.0, "2023-10-01"),
  makeTitrationEvent("susan", "glipizide", "start", 5.0, "2023-03-01"),
  makeTitrationEvent("susan", "glipizide", "stop", 0.0, "2023-09-01"),
];

const ALL_EVENTS: TitrationEvent[] = [...MARIA_EVENTS, ...JAMES_EVENTS, ...SUSAN_EVENTS];

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** TitrationTracker seeded with the full pre-existing dataset. */
function makeTracker(): TitrationTracker {
  return new TitrationTracker([...ALL_EVENTS]);
}

/** Empty TitrationTracker. */
function freshTracker(): TitrationTracker {
  return new TitrationTracker([]);
}

// ---------------------------------------------------------------------------
// PART 1 — Current medication snapshot
// ---------------------------------------------------------------------------

describe("currentMedications / getMedicationHistory", () => {
  let tracker: TitrationTracker;

  beforeEach(() => {
    tracker = makeTracker();
  });

  it("active medications returned", () => {
    const meds = tracker.currentMedications("maria");
    const names = new Set(meds.map((m) => m.name));
    expect(names.has("glipizide")).toBe(true);
  });

  it("stopped medication excluded", () => {
    const meds = tracker.currentMedications("maria");
    const names = new Set(meds.map((m) => m.name));
    expect(names.has("metformin")).toBe(false);
  });

  it("all stopped returns empty", () => {
    expect(tracker.currentMedications("susan")).toEqual([]);
  });

  it("unknown patient returns empty", () => {
    expect(tracker.currentMedications("nobody")).toEqual([]);
  });

  it("medication fields", () => {
    const meds = tracker.currentMedications("maria");
    const glip = meds.find((m) => m.name === "glipizide")!;
    expect(glip.currentDose).toBe(2.5);
    expect(glip.lastChanged).toBe("2024-01-15");
    expect(glip.totalChanges).toBe(2); // start + decrease
  });

  it("multiple active medications", () => {
    const meds = tracker.currentMedications("james");
    const names = new Set(meds.map((m) => m.name));
    expect(names.has("metformin")).toBe(true);
    expect(names.has("insulin_glargine")).toBe(true);
  });

  it("events out of order handled (sorted internally, insertion order shouldn't matter)", () => {
    const events = [
      makeTitrationEvent("pt_x", "metformin", "increase", 1000.0, "2024-03-01"),
      makeTitrationEvent("pt_x", "metformin", "start", 500.0, "2024-01-01"),
      makeTitrationEvent("pt_x", "metformin", "decrease", 750.0, "2024-02-01"),
    ];
    const t = new TitrationTracker(events);
    const meds = t.currentMedications("pt_x");
    expect(meds).toHaveLength(1);
    expect(meds[0].currentDose).toBe(1000.0);
  });

  it("get medication history sorted", () => {
    const history = tracker.getMedicationHistory("maria", "metformin");
    const dates = history.map((e) => e.recordedOn);
    expect(dates).toEqual([...dates].sort());
  });

  it("get medication history correct events", () => {
    const history = tracker.getMedicationHistory("maria", "metformin");
    const directions = history.map((e) => e.direction);
    expect(directions).toEqual(["start", "increase", "decrease", "stop"]);
  });

  it("get medication history unknown returns empty", () => {
    expect(tracker.getMedicationHistory("nobody", "metformin")).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// PART 2 — Titration counts
// ---------------------------------------------------------------------------

describe("titrationCount", () => {
  let tracker: TitrationTracker;

  beforeEach(() => {
    tracker = makeTracker();
  });

  it("total count", () => {
    expect(tracker.titrationCount("maria", "metformin")).toBe(4);
  });

  it("directional count decrease", () => {
    expect(tracker.titrationCount("maria", "metformin", "decrease")).toBe(1);
  });

  it("directional count stop", () => {
    expect(tracker.titrationCount("susan", "metformin", "stop")).toBe(1);
  });

  it("directional count zero", () => {
    expect(tracker.titrationCount("james", "metformin", "stop")).toBe(0);
  });

  it("unknown patient returns zero", () => {
    expect(tracker.titrationCount("nobody", "metformin")).toBe(0);
  });

  it("unknown medication returns zero", () => {
    expect(tracker.titrationCount("maria", "insulin_glargine")).toBe(0);
  });
});

describe("deEscalationSummary", () => {
  let tracker: TitrationTracker;

  beforeEach(() => {
    tracker = makeTracker();
  });

  it("summary includes decrease and stop", () => {
    const summary = tracker.deEscalationSummary("maria");
    // metformin: 1 decrease + 1 stop = 2; glipizide: 1 decrease = 1
    expect(summary["metformin"]).toBe(2);
    expect(summary["glipizide"]).toBe(1);
  });

  it("no de-escalations returns empty", () => {
    const events = [
      makeTitrationEvent("pt_y", "metformin", "start", 500.0, "2024-01-01"),
      makeTitrationEvent("pt_y", "metformin", "increase", 1000.0, "2024-02-01"),
    ];
    const t = new TitrationTracker(events);
    expect(t.deEscalationSummary("pt_y")).toEqual({});
  });

  it("only de-escalated meds included (james has increases but also one decrease)", () => {
    const summary = tracker.deEscalationSummary("james");
    expect("insulin_glargine" in summary).toBe(true);
    expect(summary["insulin_glargine"]).toBe(1);
    // metformin only has start+increase — not included
    expect("metformin" in summary).toBe(false);
  });

  it("unknown patient returns empty", () => {
    expect(tracker.deEscalationSummary("nobody")).toEqual({});
  });
});

// ---------------------------------------------------------------------------
// PART 3 — Population-level queries
// ---------------------------------------------------------------------------

describe("population queries", () => {
  let tracker: TitrationTracker;

  beforeEach(() => {
    tracker = makeTracker();
  });

  it("patients on medication active only", () => {
    const patients = tracker.patientsOnMedication("metformin");
    // maria stopped metformin, susan stopped metformin, james still active
    expect(patients).toEqual(["james"]);
  });

  it("patients on medication sorted", () => {
    // add two patients both on glipizide to verify sorting
    const extra = [
      makeTitrationEvent("zoe", "glipizide", "start", 5.0, "2024-01-01"),
      makeTitrationEvent("anna", "glipizide", "start", 5.0, "2024-01-01"),
    ];
    const t = new TitrationTracker([...ALL_EVENTS, ...extra]);
    const patients = t.patientsOnMedication("glipizide");
    expect(patients).toEqual([...patients].sort());
    expect(patients).toContain("maria");
    expect(patients).toContain("anna");
    expect(patients).toContain("zoe");
    expect(patients).not.toContain("susan"); // susan stopped glipizide
  });

  it("patients on medication none active", () => {
    const patients = tracker.patientsOnMedication("glipizide_xr"); // unknown med
    expect(patients).toEqual([]);
  });

  it("most titrated medications", () => {
    const top = tracker.mostTitratedMedications(3);
    const names = top.map(([name]) => name);
    const counts = top.map(([, count]) => count);
    // metformin: maria(4) + james(2) + susan(3) = 9 events total
    expect(names).toContain("metformin");
    expect(counts).toEqual([...counts].sort((a, b) => b - a));
  });

  it("most titrated fewer than n", () => {
    const events = [makeTitrationEvent("pt_z", "drug_a", "start", 10.0, "2024-01-01")];
    const t = new TitrationTracker(events);
    const top = t.mostTitratedMedications(5);
    expect(top).toHaveLength(1);
  });

  it("most titrated returns correct count", () => {
    const topDict = Object.fromEntries(tracker.mostTitratedMedications(10));
    expect(topDict["metformin"]).toBe(9); // 4 + 2 + 3
  });
});

// ---------------------------------------------------------------------------
// PART 4 — Live ingestion
// ---------------------------------------------------------------------------

describe("addEvent", () => {
  let tracker: TitrationTracker;

  beforeEach(() => {
    tracker = makeTracker();
  });

  it("new event updates current medications (a stop event removes it)", () => {
    const event = makeTitrationEvent("james", "metformin", "stop", 0.0, "2024-06-01");
    tracker.addEvent(event);
    const names = new Set(tracker.currentMedications("james").map((m) => m.name));
    expect(names.has("metformin")).toBe(false);
  });

  it("new event updates history", () => {
    const event = makeTitrationEvent("maria", "glipizide", "stop", 0.0, "2024-06-01");
    tracker.addEvent(event);
    const history = tracker.getMedicationHistory("maria", "glipizide");
    const directions = history.map((e) => e.direction);
    expect(directions).toContain("stop");
  });

  it("overwrite same date (event on same patient/medication/date overwrites)", () => {
    const event = makeTitrationEvent("james", "metformin", "stop", 0.0, "2023-12-01");
    tracker.addEvent(event);
    const history = tracker.getMedicationHistory("james", "metformin");
    const decEvent = history.find((e) => e.recordedOn === "2023-12-01")!;
    expect(decEvent.direction).toBe("stop");
  });

  it("add event new patient", () => {
    const event = makeTitrationEvent("new_pt", "metformin", "start", 500.0, "2024-05-01");
    tracker.addEvent(event);
    const meds = tracker.currentMedications("new_pt");
    expect(meds).toHaveLength(1);
    expect(meds[0].name).toBe("metformin");
  });

  it("add event affects population query", () => {
    const event = makeTitrationEvent("new_pt2", "metformin", "start", 500.0, "2024-05-01");
    tracker.addEvent(event);
    expect(tracker.patientsOnMedication("metformin")).toContain("new_pt2");
  });
});
