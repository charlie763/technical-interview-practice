/**
 * Tests for Problem 8: Patient Enrollment Pipeline
 *
 * Run from the typescript/ directory:
 *   npm run test:08
 */

import { describe, expect, it, beforeEach } from "vitest";
import { EnrollmentPipeline } from "@problems/problem_08_enrollment_pipeline";

// ── Fixtures ─────────────────────────────────────────────────────────────────

/** Empty EnrollmentPipeline with no patients. */
function freshPipeline(): EnrollmentPipeline {
  return new EnrollmentPipeline();
}

/**
 * Pre-seeded pipeline with patients in various states (all timestamps in seconds):
 *   seed_active:      referred(0) → screened(100) → enrolled(200) → active(300)
 *   seed_graduated:   referred(0) → screened(50)  → enrolled(150) → active(250) → graduated(1000)
 *   seed_ineligible:  referred(0) → screened(10)  → ineligible(20)
 *   seed_referred:    referred(0)  [no further transitions]
 */
function makePipeline(): EnrollmentPipeline {
  const p = new EnrollmentPipeline();

  p.addPatient("seed_active", 0.0);
  p.transition("seed_active", "screened", 100.0);
  p.transition("seed_active", "enrolled", 200.0);
  p.transition("seed_active", "active", 300.0);

  p.addPatient("seed_graduated", 0.0);
  p.transition("seed_graduated", "screened", 50.0);
  p.transition("seed_graduated", "enrolled", 150.0);
  p.transition("seed_graduated", "active", 250.0);
  p.transition("seed_graduated", "graduated", 1000.0);

  p.addPatient("seed_ineligible", 0.0);
  p.transition("seed_ineligible", "screened", 10.0);
  p.transition("seed_ineligible", "ineligible", 20.0);

  p.addPatient("seed_referred", 0.0);
  return p;
}

// ── Part 1: State tracking ────────────────────────────────────────────────────

describe("addPatient", () => {
  let fresh: EnrollmentPipeline;

  beforeEach(() => {
    fresh = freshPipeline();
  });

  it("new patient starts in referred", () => {
    fresh.addPatient("add_p1", 0.0);
    expect(fresh.getState("add_p1")).toBe("referred");
  });

  it("duplicate patient id throws", () => {
    fresh.addPatient("dup_p", 0.0);
    expect(() => fresh.addPatient("dup_p", 10.0)).toThrow();
  });
});

describe("transition", () => {
  let fresh: EnrollmentPipeline;

  beforeEach(() => {
    fresh = freshPipeline();
  });

  it("valid transition changes state", () => {
    fresh.addPatient("trans_p1", 0.0);
    fresh.transition("trans_p1", "screened", 100.0);
    expect(fresh.getState("trans_p1")).toBe("screened");
  });

  it("invalid transition throws", () => {
    fresh.addPatient("trans_p2", 0.0);
    // Cannot jump from referred straight to graduated
    expect(() => fresh.transition("trans_p2", "graduated", 100.0)).toThrow();
  });

  it("skipping states throws", () => {
    fresh.addPatient("trans_p3", 0.0);
    fresh.transition("trans_p3", "screened", 10.0);
    // Cannot skip enrolled → jump from screened to active
    expect(() => fresh.transition("trans_p3", "active", 20.0)).toThrow();
  });

  it("unknown patient throws", () => {
    expect(() => fresh.transition("ghost", "screened", 100.0)).toThrow();
  });

  it("transition from terminal state throws", () => {
    fresh.addPatient("trans_p4", 0.0);
    fresh.transition("trans_p4", "screened", 10.0);
    fresh.transition("trans_p4", "ineligible", 20.0);
    expect(() => fresh.transition("trans_p4", "enrolled", 30.0)).toThrow();
  });

  it("both branches from screened are valid", () => {
    fresh.addPatient("branch_p1", 0.0);
    fresh.transition("branch_p1", "screened", 10.0);
    fresh.transition("branch_p1", "enrolled", 20.0);
    expect(fresh.getState("branch_p1")).toBe("enrolled");

    fresh.addPatient("branch_p2", 0.0);
    fresh.transition("branch_p2", "screened", 10.0);
    fresh.transition("branch_p2", "ineligible", 20.0);
    expect(fresh.getState("branch_p2")).toBe("ineligible");
  });
});

describe("getState", () => {
  let pipeline: EnrollmentPipeline;

  beforeEach(() => {
    pipeline = makePipeline();
  });

  it("returns current state for each patient", () => {
    expect(pipeline.getState("seed_active")).toBe("active");
    expect(pipeline.getState("seed_graduated")).toBe("graduated");
    expect(pipeline.getState("seed_ineligible")).toBe("ineligible");
    expect(pipeline.getState("seed_referred")).toBe("referred");
  });

  it("unknown patient throws", () => {
    expect(() => freshPipeline().getState("nobody")).toThrow();
  });
});

describe("getPatientsInState", () => {
  let pipeline: EnrollmentPipeline;

  beforeEach(() => {
    pipeline = makePipeline();
  });

  it("returns patients in state", () => {
    const activePatients = pipeline.getPatientsInState("active");
    expect(activePatients).toContain("seed_active");
    expect(activePatients).not.toContain("seed_graduated");
  });

  it("result is sorted", () => {
    const result = pipeline.getPatientsInState("referred");
    expect(result).toEqual([...result].sort());
  });

  it("returns empty for unpopulated state", () => {
    expect(pipeline.getPatientsInState("withdrawn")).toEqual([]);
  });

  it("patient absent after leaving state", () => {
    // seed_ineligible passed through screened but is no longer there
    expect(pipeline.getPatientsInState("screened")).not.toContain("seed_ineligible");
  });
});

// ── Part 2: Duration and conversion metrics ───────────────────────────────────

describe("timeInState", () => {
  let fresh: EnrollmentPipeline;

  beforeEach(() => {
    fresh = freshPipeline();
  });

  it("completed state returns exact duration", () => {
    fresh.addPatient("dur_p1", 0.0);
    fresh.transition("dur_p1", "screened", 1000.0);
    fresh.transition("dur_p1", "enrolled", 4000.0);
    // Spent exactly 3000 s in screened; asOf is ignored for completed states
    expect(fresh.timeInState("dur_p1", "screened", 99999.0)).toBe(3000.0);
  });

  it("current state counts up to asOf", () => {
    fresh.addPatient("dur_p2", 0.0);
    fresh.transition("dur_p2", "screened", 1000.0);
    expect(fresh.timeInState("dur_p2", "screened", 4000.0)).toBe(3000.0);
  });

  it("returns zero for state never visited", () => {
    fresh.addPatient("dur_p3", 0.0);
    expect(fresh.timeInState("dur_p3", "enrolled", 99999.0)).toBe(0.0);
  });

  it("initial referred state timed from addPatient timestamp", () => {
    fresh.addPatient("dur_p4", 500.0);
    fresh.transition("dur_p4", "screened", 1500.0);
    // Spent 1000 s in referred (1500 - 500)
    expect(fresh.timeInState("dur_p4", "referred", 99999.0)).toBe(1000.0);
  });
});

describe("conversionRate", () => {
  let fresh: EnrollmentPipeline;

  beforeEach(() => {
    fresh = freshPipeline();
  });

  it("fifty percent conversion", () => {
    fresh.addPatient("conv_p1", 0.0);
    fresh.transition("conv_p1", "screened", 10.0);
    fresh.transition("conv_p1", "enrolled", 20.0);

    fresh.addPatient("conv_p2", 0.0);
    fresh.transition("conv_p2", "screened", 10.0);
    fresh.transition("conv_p2", "ineligible", 20.0);

    expect(fresh.conversionRate("screened", "enrolled")).toBeCloseTo(0.5);
  });

  it("excludes patients still in from-state", () => {
    fresh.addPatient("conv_p3", 0.0);
    fresh.transition("conv_p3", "screened", 10.0);
    // conv_p3 is still in screened — must not be counted

    fresh.addPatient("conv_p4", 0.0);
    fresh.transition("conv_p4", "screened", 10.0);
    fresh.transition("conv_p4", "enrolled", 20.0);

    // Only conv_p4 has exited; they enrolled → rate is 1.0
    expect(fresh.conversionRate("screened", "enrolled")).toBeCloseTo(1.0);
  });

  it("returns zero when no one has exited from-state", () => {
    fresh.addPatient("conv_p5", 0.0);
    // conv_p5 is still in referred
    expect(fresh.conversionRate("referred", "screened")).toBeCloseTo(0.0);
  });

  it("hundred percent when all converted", () => {
    for (const pid of ["conv_all_1", "conv_all_2"]) {
      fresh.addPatient(pid, 0.0);
      fresh.transition(pid, "screened", 10.0);
      fresh.transition(pid, "enrolled", 20.0);
    }
    expect(fresh.conversionRate("screened", "enrolled")).toBeCloseTo(1.0);
  });
});

// ── Part 3: SLA monitoring ────────────────────────────────────────────────────

describe("patientsOverdue", () => {
  let fresh: EnrollmentPipeline;

  beforeEach(() => {
    fresh = freshPipeline();
  });

  it("returns patients exceeding threshold", () => {
    fresh.addPatient("over_p1", 0.0);
    fresh.transition("over_p1", "screened", 0.0);
    fresh.transition("over_p1", "enrolled", 0.0);
    fresh.transition("over_p1", "active", 0.0); // 10000 s in active

    fresh.addPatient("over_p2", 0.0);
    fresh.transition("over_p2", "screened", 0.0);
    fresh.transition("over_p2", "enrolled", 0.0);
    fresh.transition("over_p2", "active", 5000.0); // 5000 s in active

    const overdue = fresh.patientsOverdue("active", 6000.0, 10000.0);
    expect(overdue).toContain("over_p1");
    expect(overdue).not.toContain("over_p2");
  });

  it("sorted by duration descending", () => {
    fresh.addPatient("sort_p1", 0.0);
    fresh.transition("sort_p1", "screened", 0.0);
    fresh.transition("sort_p1", "enrolled", 0.0);
    fresh.transition("sort_p1", "active", 0.0); // 10000 s

    fresh.addPatient("sort_p2", 0.0);
    fresh.transition("sort_p2", "screened", 0.0);
    fresh.transition("sort_p2", "enrolled", 0.0);
    fresh.transition("sort_p2", "active", 3000.0); // 7000 s

    const overdue = fresh.patientsOverdue("active", 5000.0, 10000.0);
    expect(overdue).toEqual(["sort_p1", "sort_p2"]);
  });

  it("excludes patients not currently in state", () => {
    // seed_graduated has left active — must not appear in active overdue list
    const pipeline = makePipeline();
    const overdue = pipeline.patientsOverdue("active", 0.0, 2000.0);
    expect(overdue).not.toContain("seed_graduated");
  });

  it("returns empty when no one is overdue", () => {
    fresh.addPatient("noover_p", 0.0);
    fresh.transition("noover_p", "screened", 0.0);
    fresh.transition("noover_p", "enrolled", 0.0);
    fresh.transition("noover_p", "active", 9900.0); // only 100 s in active
    expect(fresh.patientsOverdue("active", 500.0, 10000.0)).toEqual([]);
  });
});

describe("averageTimeInState", () => {
  let fresh: EnrollmentPipeline;

  beforeEach(() => {
    fresh = freshPipeline();
  });

  it("average over all exited patients", () => {
    // avg_pa: 1000 s in screened; avg_pb: 3000 s in screened → average = 2000
    fresh.addPatient("avg_pa", 0.0);
    fresh.transition("avg_pa", "screened", 0.0);
    fresh.transition("avg_pa", "enrolled", 1000.0);

    fresh.addPatient("avg_pb", 0.0);
    fresh.transition("avg_pb", "screened", 0.0);
    fresh.transition("avg_pb", "enrolled", 3000.0);

    expect(fresh.averageTimeInState("screened", 99999.0)).toBeCloseTo(2000.0);
  });

  it("excludes patients still in state", () => {
    fresh.addPatient("avg_pc", 0.0);
    fresh.transition("avg_pc", "screened", 0.0);
    fresh.transition("avg_pc", "enrolled", 1000.0); // exited: 1000 s

    fresh.addPatient("avg_pd", 0.0);
    fresh.transition("avg_pd", "screened", 0.0);
    // avg_pd is still in screened at asOf=5000 — must be excluded

    expect(fresh.averageTimeInState("screened", 5000.0)).toBeCloseTo(1000.0);
  });

  it("returns zero when no one has exited", () => {
    fresh.addPatient("avg_pe", 0.0);
    fresh.transition("avg_pe", "screened", 0.0);
    // avg_pe is still in screened
    expect(fresh.averageTimeInState("screened", 5000.0)).toBeCloseTo(0.0);
  });
});
