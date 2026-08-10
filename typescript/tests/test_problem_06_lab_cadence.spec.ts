/**
 * Tests for Problem 6: Lab Cadence Compliance Monitor
 *
 * Run from the typescript/ directory:
 *   npm run test:06
 */

import { describe, expect, it, beforeEach } from "vitest";
import {
  makeMonitor,
  registerPatient,
  addRequiredLab,
  getRequiredLabs,
  setLabDeadline,
  recordSubmission,
  isOverdue,
  overdueLabs,
  complianceReport,
  submissionHistory,
  daysSinceLastSubmission,
  type LabCadenceMonitor,
} from "@problems/problem_06_lab_cadence";

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** Bare monitor with no patients. */
function makeM(): LabCadenceMonitor {
  return makeMonitor();
}

/** Monitor pre-loaded with a small realistic patient set. */
function makeSeeded(): LabCadenceMonitor {
  const m = makeM();
  registerPatient(m, "alice", ["hba1c", "bmp", "lipids"]);
  registerPatient(m, "bob", ["hba1c", "bmp"]);
  registerPatient(m, "carol", ["hba1c"]);

  // Alice: hba1c quarterly deadlines; submitted first one, missed second
  setLabDeadline(m, "alice", "hba1c", "2024-03-31");
  setLabDeadline(m, "alice", "hba1c", "2024-06-30");
  recordSubmission(m, "alice", "hba1c", "2024-03-28");

  // Alice: bmp due but no submission
  setLabDeadline(m, "alice", "bmp", "2024-04-15");

  // Bob: hba1c submitted on time; bmp overdue
  setLabDeadline(m, "bob", "hba1c", "2024-03-31");
  recordSubmission(m, "bob", "hba1c", "2024-03-25");
  setLabDeadline(m, "bob", "bmp", "2024-03-31");

  // Carol: all labs submitted on time
  setLabDeadline(m, "carol", "hba1c", "2024-03-31");
  recordSubmission(m, "carol", "hba1c", "2024-03-15");

  return m;
}

// ---------------------------------------------------------------------------
// PART 1 — Patient & lab registration
// ---------------------------------------------------------------------------

describe("registerPatient", () => {
  let m: LabCadenceMonitor;

  beforeEach(() => {
    m = makeM();
  });

  it("registers new patient", () => {
    registerPatient(m, "dave", ["hba1c"]);
    expect(getRequiredLabs(m, "dave")).toEqual(new Set(["hba1c"]));
  });

  it("empty required labs throws", () => {
    expect(() => registerPatient(m, "eve", [])).toThrow();
  });

  it("idempotent for existing lab", () => {
    registerPatient(m, "frank", ["hba1c"]);
    registerPatient(m, "frank", ["hba1c"]);
    expect(getRequiredLabs(m, "frank")).toEqual(new Set(["hba1c"]));
  });

  it("adds new labs to existing patient", () => {
    registerPatient(m, "grace", ["hba1c"]);
    registerPatient(m, "grace", ["bmp"]);
    expect(getRequiredLabs(m, "grace")).toEqual(new Set(["hba1c", "bmp"]));
  });

  it("does not remove existing labs", () => {
    registerPatient(m, "hank", ["hba1c", "bmp"]);
    registerPatient(m, "hank", ["lipids"]);
    const labs = getRequiredLabs(m, "hank");
    expect(labs.has("hba1c")).toBe(true);
    expect(labs.has("bmp")).toBe(true);
  });
});

describe("addRequiredLab", () => {
  let m: LabCadenceMonitor;

  beforeEach(() => {
    m = makeM();
  });

  it("adds lab to existing patient", () => {
    registerPatient(m, "iris", ["hba1c"]);
    addRequiredLab(m, "iris", "bmp");
    expect(getRequiredLabs(m, "iris").has("bmp")).toBe(true);
  });

  it("is idempotent", () => {
    registerPatient(m, "jack", ["hba1c"]);
    addRequiredLab(m, "jack", "hba1c"); // already required
    expect(getRequiredLabs(m, "jack")).toEqual(new Set(["hba1c"]));
  });

  it("unknown patient throws", () => {
    expect(() => addRequiredLab(m, "nobody", "hba1c")).toThrow();
  });
});

describe("getRequiredLabs", () => {
  let m: LabCadenceMonitor;

  beforeEach(() => {
    m = makeM();
  });

  it("returns correct set", () => {
    registerPatient(m, "kate", ["hba1c", "lipids"]);
    expect(getRequiredLabs(m, "kate")).toEqual(new Set(["hba1c", "lipids"]));
  });

  it("unknown patient throws", () => {
    expect(() => getRequiredLabs(m, "nobody")).toThrow();
  });
});

// ---------------------------------------------------------------------------
// PART 2 — Deadlines and submissions
// ---------------------------------------------------------------------------

describe("setLabDeadline", () => {
  let seeded: LabCadenceMonitor;

  beforeEach(() => {
    seeded = makeSeeded();
  });

  it("sets deadline (no exception)", () => {
    expect(() => setLabDeadline(seeded, "alice", "lipids", "2024-05-01")).not.toThrow();
  });

  it("duplicate deadline ignored", () => {
    setLabDeadline(seeded, "alice", "lipids", "2024-05-01");
    setLabDeadline(seeded, "alice", "lipids", "2024-05-01"); // duplicate
    // isOverdue should work normally — no duplication side effects
    expect(isOverdue(seeded, "alice", "lipids", "2024-05-02")).toBe(true);
  });

  it("unknown patient throws", () => {
    const m = makeM();
    expect(() => setLabDeadline(m, "nobody", "hba1c", "2024-03-31")).toThrow();
  });

  it("unrequired lab throws", () => {
    const m = makeM();
    registerPatient(m, "leo", ["hba1c"]);
    expect(() => setLabDeadline(m, "leo", "bmp", "2024-03-31")).toThrow();
  });
});

describe("recordSubmission", () => {
  let seeded: LabCadenceMonitor;

  beforeEach(() => {
    seeded = makeSeeded();
  });

  it("submission clears deadline", () => {
    // carol submitted hba1c on Mar 15, deadline was Mar 31 → not overdue
    expect(isOverdue(seeded, "carol", "hba1c", "2024-04-01")).toBe(false);
  });

  it("submission clears earliest applicable deadline", () => {
    // alice: hba1c deadlines Mar 31 (cleared) and Jun 30 (not cleared)
    // as of Jul 1, the Jun 30 deadline should be overdue
    expect(isOverdue(seeded, "alice", "hba1c", "2024-07-01")).toBe(true);
  });

  it("unknown patient throws", () => {
    const m = makeM();
    expect(() => recordSubmission(m, "nobody", "hba1c", "2024-03-28")).toThrow();
  });

  it("unrequired lab throws", () => {
    const m = makeM();
    registerPatient(m, "mia", ["hba1c"]);
    expect(() => recordSubmission(m, "mia", "bmp", "2024-03-28")).toThrow();
  });

  it("late submission does not clear past deadline", () => {
    // A submission AFTER the due date should not clear it.
    const m = makeM();
    registerPatient(m, "noah", ["hba1c"]);
    setLabDeadline(m, "noah", "hba1c", "2024-03-31");
    recordSubmission(m, "noah", "hba1c", "2024-04-15"); // submitted late
    // deadline Mar 31 should still be considered overdue as of Apr 1
    expect(isOverdue(m, "noah", "hba1c", "2024-04-01")).toBe(true);
  });
});

describe("isOverdue", () => {
  let seeded: LabCadenceMonitor;

  beforeEach(() => {
    seeded = makeSeeded();
  });

  it("overdue deadline no submission", () => {
    // bob: bmp deadline Mar 31, no submission
    expect(isOverdue(seeded, "bob", "bmp", "2024-04-01")).toBe(true);
  });

  it("not overdue when submitted", () => {
    // bob: hba1c submitted on time
    expect(isOverdue(seeded, "bob", "hba1c", "2024-04-01")).toBe(false);
  });

  it("not overdue when deadline not yet passed", () => {
    // alice: bmp deadline Apr 15; as of Apr 14 not yet overdue
    expect(isOverdue(seeded, "alice", "bmp", "2024-04-14")).toBe(false);
  });

  it("overdue on exact day after deadline", () => {
    // alice: bmp deadline Apr 15; Apr 16 → overdue
    expect(isOverdue(seeded, "alice", "bmp", "2024-04-16")).toBe(true);
  });

  it("not overdue when no deadline set", () => {
    // alice: lipids has no deadline set
    expect(isOverdue(seeded, "alice", "lipids", "2024-06-01")).toBe(false);
  });

  it("unknown patient returns false", () => {
    expect(isOverdue(seeded, "nobody", "hba1c", "2024-04-01")).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// PART 3 — Compliance reporting
// ---------------------------------------------------------------------------

describe("overdueLabs", () => {
  let seeded: LabCadenceMonitor;

  beforeEach(() => {
    seeded = makeSeeded();
  });

  it("returns sorted overdue labs", () => {
    // alice as of Jul 1: bmp (missed Apr 15 deadline) and hba1c (missed Jun 30 deadline)
    const labs = overdueLabs(seeded, "alice", "2024-07-01");
    expect([...labs].sort()).toEqual(labs);
    expect(labs).toContain("bmp");
    expect(labs).toContain("hba1c");
  });

  it("no overdue returns empty", () => {
    expect(overdueLabs(seeded, "carol", "2024-04-01")).toEqual([]);
  });

  it("unknown patient returns empty", () => {
    expect(overdueLabs(seeded, "nobody", "2024-04-01")).toEqual([]);
  });
});

describe("complianceReport", () => {
  let seeded: LabCadenceMonitor;

  beforeEach(() => {
    seeded = makeSeeded();
  });

  it("report contains overdue patients", () => {
    const report = complianceReport(seeded, "2024-04-16");
    const patientIds = report.map((e) => e.patientId);
    expect(patientIds).toContain("alice"); // bmp overdue
    expect(patientIds).toContain("bob"); // bmp overdue
  });

  it("compliant patients excluded", () => {
    const report = complianceReport(seeded, "2024-04-16");
    const patientIds = report.map((e) => e.patientId);
    expect(patientIds).not.toContain("carol");
  });

  it("entry fields", () => {
    const report = complianceReport(seeded, "2024-04-16");
    const bobEntry = report.find((e) => e.patientId === "bob")!;
    expect(Object.keys(bobEntry).sort()).toEqual(["overdueCount", "overdueLabs", "patientId"]);
    expect(bobEntry.overdueCount).toBe(bobEntry.overdueLabs.length);
  });

  it("sorted by overdueCount descending", () => {
    const report = complianceReport(seeded, "2024-07-01");
    const counts = report.map((e) => e.overdueCount);
    expect(counts).toEqual([...counts].sort((a, b) => b - a));
  });

  it("empty report when no overdue", () => {
    const m = makeM();
    registerPatient(m, "perfectly_compliant", ["hba1c"]);
    setLabDeadline(m, "perfectly_compliant", "hba1c", "2024-03-31");
    recordSubmission(m, "perfectly_compliant", "hba1c", "2024-03-20");
    const report = complianceReport(m, "2024-04-01");
    expect(report).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// PART 4 — Submission history
// ---------------------------------------------------------------------------

describe("submissionHistory", () => {
  it("returns sorted dates", () => {
    const m = makeM();
    registerPatient(m, "quinn", ["hba1c"]);
    setLabDeadline(m, "quinn", "hba1c", "2024-03-31");
    setLabDeadline(m, "quinn", "hba1c", "2024-06-30");
    recordSubmission(m, "quinn", "hba1c", "2024-03-20");
    recordSubmission(m, "quinn", "hba1c", "2024-06-15");
    const history = submissionHistory(m, "quinn", "hba1c");
    expect(history).toEqual([...history].sort());
    expect(history).toHaveLength(2);
  });

  it("no submissions returns empty", () => {
    // alice has no lipids submissions
    const seeded = makeSeeded();
    expect(submissionHistory(seeded, "alice", "lipids")).toEqual([]);
  });

  it("unknown patient returns empty", () => {
    const m = makeM();
    expect(submissionHistory(m, "nobody", "hba1c")).toEqual([]);
  });
});

describe("daysSinceLastSubmission", () => {
  it("correct days", () => {
    const m = makeM();
    registerPatient(m, "rita", ["hba1c"]);
    setLabDeadline(m, "rita", "hba1c", "2024-03-31");
    recordSubmission(m, "rita", "hba1c", "2024-03-20");
    expect(daysSinceLastSubmission(m, "rita", "hba1c", "2024-04-20")).toBe(31);
  });

  it("no submission returns undefined", () => {
    const seeded = makeSeeded();
    expect(daysSinceLastSubmission(seeded, "alice", "lipids", "2024-05-01")).toBeUndefined();
  });

  it("unknown patient returns undefined", () => {
    const m = makeM();
    expect(daysSinceLastSubmission(m, "nobody", "hba1c", "2024-04-01")).toBeUndefined();
  });
});
