/**
 * Tests for Problem 4: Biomarker Alert Monitor
 *
 * Run from the typescript/ directory:
 *   npm run test:04
 */

import { describe, expect, it, beforeEach } from "vitest";
import { BiomarkerMonitor, makeBiomarkerReading, type BiomarkerReading } from "@problems/problem_04_biomarker_alert";

// ---------------------------------------------------------------------------
// Pre-existing data — represents a snapshot of patient readings already in the
// system when the monitor is initialized. Tests below build on this dataset.
// ---------------------------------------------------------------------------

const ALICE_READINGS: BiomarkerReading[] = [
  // All in range — no outreach needed
  makeBiomarkerReading("alice", "glucose", 105.0, "2024-01-01"),
  makeBiomarkerReading("alice", "glucose", 98.0, "2024-01-02"),
  makeBiomarkerReading("alice", "glucose", 112.0, "2024-01-03"),
  makeBiomarkerReading("alice", "glucose", 91.0, "2024-01-04"),
];

const BOB_READINGS: BiomarkerReading[] = [
  // 4 consecutive high-glucose days → needs outreach
  makeBiomarkerReading("bob", "glucose", 195.0, "2024-01-01"),
  makeBiomarkerReading("bob", "glucose", 210.0, "2024-01-02"),
  makeBiomarkerReading("bob", "glucose", 188.0, "2024-01-03"),
  makeBiomarkerReading("bob", "glucose", 202.0, "2024-01-04"),
];

const CAROL_READINGS: BiomarkerReading[] = [
  // Streak of 1 (Jan 1), in-range on Jan 2, then streak of 2 (Jan 3–4) → max=2
  makeBiomarkerReading("carol", "glucose", 190.0, "2024-01-01"),
  makeBiomarkerReading("carol", "glucose", 150.0, "2024-01-02"), // in range
  makeBiomarkerReading("carol", "glucose", 185.0, "2024-01-03"),
  makeBiomarkerReading("carol", "glucose", 191.0, "2024-01-04"),
];

const DAVE_READINGS: BiomarkerReading[] = [
  // 2 consecutive high-glucose days — below default threshold of 3
  makeBiomarkerReading("dave", "glucose", 199.0, "2024-01-03"),
  makeBiomarkerReading("dave", "glucose", 205.0, "2024-01-04"),
];

const EVE_READINGS: BiomarkerReading[] = [
  // One dangerous low glucose (below 70) on Jan 1 — streak of 1 for glucose
  makeBiomarkerReading("eve", "glucose", 62.0, "2024-01-01"),
  // 3 consecutive days with ketones below target (< 0.5) → ketone outreach
  makeBiomarkerReading("eve", "ketone", 0.3, "2024-01-02"),
  makeBiomarkerReading("eve", "ketone", 0.2, "2024-01-03"),
  makeBiomarkerReading("eve", "ketone", 0.4, "2024-01-04"),
  // Weight readings — never out of range
  makeBiomarkerReading("eve", "weight", 165.0, "2024-01-01"),
  makeBiomarkerReading("eve", "weight", 164.5, "2024-01-02"),
];

const ALL_READINGS: BiomarkerReading[] = [...ALICE_READINGS, ...BOB_READINGS, ...CAROL_READINGS, ...DAVE_READINGS, ...EVE_READINGS];

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** BiomarkerMonitor seeded with the full pre-existing dataset. */
function makeMonitor(): BiomarkerMonitor {
  return new BiomarkerMonitor([...ALL_READINGS]);
}

/** Empty BiomarkerMonitor. */
function freshMonitor(): BiomarkerMonitor {
  return new BiomarkerMonitor([]);
}

// ---------------------------------------------------------------------------
// PART 1 — Single-reading classification
// ---------------------------------------------------------------------------

describe("isOutOfRange", () => {
  let monitor: BiomarkerMonitor;

  beforeEach(() => {
    monitor = freshMonitor();
  });

  it("glucose above range", () => {
    const r = makeBiomarkerReading("p1", "glucose", 181.0, "2024-01-01");
    expect(monitor.isOutOfRange(r)).toBe(true);
  });

  it("glucose below range", () => {
    const r = makeBiomarkerReading("p1", "glucose", 69.9, "2024-01-01");
    expect(monitor.isOutOfRange(r)).toBe(true);
  });

  it("glucose at upper boundary", () => {
    const r = makeBiomarkerReading("p1", "glucose", 180.0, "2024-01-01");
    expect(monitor.isOutOfRange(r)).toBe(false);
  });

  it("glucose at lower boundary", () => {
    const r = makeBiomarkerReading("p1", "glucose", 70.0, "2024-01-01");
    expect(monitor.isOutOfRange(r)).toBe(false);
  });

  it("glucose in range", () => {
    const r = makeBiomarkerReading("p1", "glucose", 120.0, "2024-01-01");
    expect(monitor.isOutOfRange(r)).toBe(false);
  });

  it("ketone above range", () => {
    const r = makeBiomarkerReading("p1", "ketone", 3.1, "2024-01-01");
    expect(monitor.isOutOfRange(r)).toBe(true);
  });

  it("ketone below range", () => {
    const r = makeBiomarkerReading("p1", "ketone", 0.4, "2024-01-01");
    expect(monitor.isOutOfRange(r)).toBe(true);
  });

  it("ketone in range", () => {
    const r = makeBiomarkerReading("p1", "ketone", 1.5, "2024-01-01");
    expect(monitor.isOutOfRange(r)).toBe(false);
  });

  it("ketone at boundaries", () => {
    const lo = makeBiomarkerReading("p1", "ketone", 0.5, "2024-01-01");
    const hi = makeBiomarkerReading("p1", "ketone", 3.0, "2024-01-01");
    expect(monitor.isOutOfRange(lo)).toBe(false);
    expect(monitor.isOutOfRange(hi)).toBe(false);
  });

  it("weight never out of range", () => {
    const r = makeBiomarkerReading("p1", "weight", 9999.0, "2024-01-01");
    expect(monitor.isOutOfRange(r)).toBe(false);
  });

  it("classification uses preloaded data (works regardless of monitor state)", () => {
    const m = makeMonitor();
    const inRange = makeBiomarkerReading("alice", "glucose", 100.0, "2024-01-05");
    const outOfRange = makeBiomarkerReading("bob", "glucose", 250.0, "2024-01-05");
    expect(m.isOutOfRange(inRange)).toBe(false);
    expect(m.isOutOfRange(outOfRange)).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// PART 2 — Streak detection
// ---------------------------------------------------------------------------

describe("maxConsecutiveOutOfRangeDays", () => {
  let monitor: BiomarkerMonitor;

  beforeEach(() => {
    monitor = makeMonitor();
  });

  it("no readings returns zero", () => {
    expect(monitor.maxConsecutiveOutOfRangeDays("unknown_patient", "glucose")).toBe(0);
  });

  it("all in range returns zero", () => {
    expect(monitor.maxConsecutiveOutOfRangeDays("alice", "glucose")).toBe(0);
  });

  it("consecutive streak", () => {
    expect(monitor.maxConsecutiveOutOfRangeDays("bob", "glucose")).toBe(4);
  });

  it("streak broken by in-range day", () => {
    // carol: streak of 1, break, streak of 2 → max = 2
    expect(monitor.maxConsecutiveOutOfRangeDays("carol", "glucose")).toBe(2);
  });

  it("short streak", () => {
    expect(monitor.maxConsecutiveOutOfRangeDays("dave", "glucose")).toBe(2);
  });

  it("glucose streak of one for low", () => {
    expect(monitor.maxConsecutiveOutOfRangeDays("eve", "glucose")).toBe(1);
  });

  it("ketone streak", () => {
    expect(monitor.maxConsecutiveOutOfRangeDays("eve", "ketone")).toBe(3);
  });

  it("weight never streaks", () => {
    expect(monitor.maxConsecutiveOutOfRangeDays("eve", "weight")).toBe(0);
  });

  it("multiple readings same day count as one", () => {
    const readings = [
      makeBiomarkerReading("frank", "glucose", 200.0, "2024-02-01"),
      makeBiomarkerReading("frank", "glucose", 210.0, "2024-02-01"), // same day
      makeBiomarkerReading("frank", "glucose", 195.0, "2024-02-02"),
    ];
    const m = new BiomarkerMonitor(readings);
    expect(m.maxConsecutiveOutOfRangeDays("frank", "glucose")).toBe(2);
  });

  it("gap breaks streak", () => {
    const readings = [
      makeBiomarkerReading("grace", "glucose", 200.0, "2024-03-01"),
      makeBiomarkerReading("grace", "glucose", 200.0, "2024-03-03"), // skip Mar 2
    ];
    const m = new BiomarkerMonitor(readings);
    expect(m.maxConsecutiveOutOfRangeDays("grace", "glucose")).toBe(1);
  });

  it("wrong reading type ignored", () => {
    expect(monitor.maxConsecutiveOutOfRangeDays("eve", "glucose")).toBe(1);
    expect(monitor.maxConsecutiveOutOfRangeDays("eve", "ketone")).toBe(3);
  });
});

// ---------------------------------------------------------------------------
// PART 3 — Outreach list
// ---------------------------------------------------------------------------

describe("getOutreachList", () => {
  let monitor: BiomarkerMonitor;

  beforeEach(() => {
    monitor = makeMonitor();
  });

  it("default threshold filters correctly (only bob and eve/ketone qualify)", () => {
    const result = monitor.getOutreachList();
    const patientTypes = new Set(result.map((e) => `${e.patientId}:${e.readingType}`));
    expect(patientTypes.has("bob:glucose")).toBe(true);
    expect(patientTypes.has("eve:ketone")).toBe(true);
    // alice, carol, dave don't qualify at threshold 3
    expect(patientTypes.has("alice:glucose")).toBe(false);
    expect(patientTypes.has("carol:glucose")).toBe(false);
    expect(patientTypes.has("dave:glucose")).toBe(false);
  });

  it("weight never in outreach", () => {
    const result = monitor.getOutreachList(1);
    for (const entry of result) {
      expect(entry.readingType).not.toBe("weight");
    }
  });

  it("result sorted by consecutiveDays descending", () => {
    const result = monitor.getOutreachList(1);
    const days = result.map((e) => e.consecutiveDays);
    const sorted = [...days].sort((a, b) => b - a);
    expect(days).toEqual(sorted);
  });

  it("entry fields", () => {
    const result = monitor.getOutreachList();
    const bobEntry = result.find((e) => e.patientId === "bob")!;
    expect(Object.keys(bobEntry).sort()).toEqual(["patientId", "readingType", "consecutiveDays", "latestValue"].sort());
    expect(bobEntry.consecutiveDays).toBe(4);
    expect(bobEntry.readingType).toBe("glucose");
    expect(typeof bobEntry.latestValue).toBe("number");
  });

  it("latest value is most recent out of range", () => {
    const result = monitor.getOutreachList();
    const bobEntry = result.find((e) => e.patientId === "bob")!;
    // Bob's most recent OOR reading is Jan 4: 202.0
    expect(bobEntry.latestValue).toBe(202.0);
  });

  it("patient can appear twice for different types", () => {
    const readings = [
      // 3-day glucose streak
      makeBiomarkerReading("hank", "glucose", 200.0, "2024-01-01"),
      makeBiomarkerReading("hank", "glucose", 210.0, "2024-01-02"),
      makeBiomarkerReading("hank", "glucose", 195.0, "2024-01-03"),
      // 3-day ketone streak
      makeBiomarkerReading("hank", "ketone", 0.2, "2024-01-01"),
      makeBiomarkerReading("hank", "ketone", 0.3, "2024-01-02"),
      makeBiomarkerReading("hank", "ketone", 0.1, "2024-01-03"),
    ];
    const m = new BiomarkerMonitor(readings);
    const result = m.getOutreachList(3);
    const patientTypes = result.map((e) => `${e.patientId}:${e.readingType}`);
    expect(patientTypes).toContain("hank:glucose");
    expect(patientTypes).toContain("hank:ketone");
  });

  it("custom threshold", () => {
    const result2 = monitor.getOutreachList(2);
    const patientTypes2 = new Set(result2.map((e) => `${e.patientId}:${e.readingType}`));
    expect(patientTypes2.has("carol:glucose")).toBe(true);
    expect(patientTypes2.has("dave:glucose")).toBe(true);
  });

  it("empty monitor returns empty", () => {
    expect(freshMonitor().getOutreachList()).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// PART 4 — Deduplication on ingestion
// ---------------------------------------------------------------------------

describe("addReading", () => {
  let monitor: BiomarkerMonitor;

  beforeEach(() => {
    monitor = makeMonitor();
  });

  it("new reading returns true", () => {
    const r = makeBiomarkerReading("alice", "glucose", 100.0, "2024-01-10");
    expect(monitor.addReading(r)).toBe(true);
  });

  it("duplicate exact same returns false", () => {
    const r = makeBiomarkerReading("alice", "glucose", 105.0, "2024-01-01");
    expect(monitor.addReading(r)).toBe(false);
  });

  it("duplicate within tolerance returns false", () => {
    // alice jan 1 = 105.0; 105.4 is within ±0.5
    const r = makeBiomarkerReading("alice", "glucose", 105.4, "2024-01-01");
    expect(monitor.addReading(r)).toBe(false);
  });

  it("just outside tolerance returns true", () => {
    // alice jan 1 = 105.0; 105.6 is outside ±0.5
    const r = makeBiomarkerReading("alice", "glucose", 105.6, "2024-01-01");
    expect(monitor.addReading(r)).toBe(true);
  });

  it("different date not duplicate", () => {
    const r = makeBiomarkerReading("alice", "glucose", 105.0, "2024-01-05");
    expect(monitor.addReading(r)).toBe(true);
  });

  it("different type not duplicate", () => {
    const r = makeBiomarkerReading("alice", "ketone", 1.0, "2024-01-01");
    expect(monitor.addReading(r)).toBe(true);
  });

  it("different patient not duplicate", () => {
    const r = makeBiomarkerReading("frank", "glucose", 105.0, "2024-01-01");
    expect(monitor.addReading(r)).toBe(true);
  });

  it("added reading affects streak", () => {
    // bob's current streak is Jan 1-4 (4 days). Add Jan 5.
    const r = makeBiomarkerReading("bob", "glucose", 195.0, "2024-01-05");
    monitor.addReading(r);
    expect(monitor.maxConsecutiveOutOfRangeDays("bob", "glucose")).toBe(5);
  });

  it("adding duplicate does not affect streak", () => {
    const r = makeBiomarkerReading("bob", "glucose", 195.0, "2024-01-01");
    monitor.addReading(r); // duplicate, rejected
    expect(monitor.maxConsecutiveOutOfRangeDays("bob", "glucose")).toBe(4);
  });
});
