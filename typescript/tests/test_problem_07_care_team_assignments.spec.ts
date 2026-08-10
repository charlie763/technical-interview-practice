/**
 * Tests for Problem 7: Care Team Assignment Manager
 *
 * Run from the typescript/ directory:
 *   npm run test:07
 */

import { describe, expect, it, beforeEach } from "vitest";
import { CareTeamManager, CapacityError } from "@problems/problem_07_care_team_assignments";

// ── Fixtures ─────────────────────────────────────────────────────────────────

/** Empty CareTeamManager with no members or assignments. */
function freshMgr(): CareTeamManager {
  return new CareTeamManager();
}

/**
 * Pre-seeded manager:
 *   - coach_alpha  (coach,     max=3): assigned patient_seed_1, patient_seed_2
 *   - coach_beta   (coach,     max=1): no patients
 *   - dr_omega     (physician, max=100): assigned patient_seed_1
 */
function makeMgr(): CareTeamManager {
  const m = new CareTeamManager();
  m.addMember("coach_alpha", "coach", 3);
  m.addMember("coach_beta", "coach", 1);
  m.addMember("dr_omega", "physician", 100);
  m.assign("patient_seed_1", "coach_alpha", 1000.0);
  m.assign("patient_seed_1", "dr_omega", 1000.0);
  m.assign("patient_seed_2", "coach_alpha", 2000.0);
  return m;
}

// ── Part 1: Basic assignment and lookup ──────────────────────────────────────

describe("addMember", () => {
  let fresh: CareTeamManager;

  beforeEach(() => {
    fresh = freshMgr();
  });

  it("registered member can receive assignment", () => {
    fresh.addMember("member_reg_test", "coach", 5);
    // Should not throw — member is known
    expect(() => fresh.assign("patient_reg_test", "member_reg_test", 0.0)).not.toThrow();
  });

  it("unregistered member throws", () => {
    expect(() => fresh.assign("patient_unreg_test", "ghost_member", 0.0)).toThrow();
  });
});

describe("assign", () => {
  let fresh: CareTeamManager;

  beforeEach(() => {
    fresh = freshMgr();
  });

  it("basic assignment recorded", () => {
    fresh.addMember("coach_basic", "coach", 5);
    fresh.assign("patient_basic", "coach_basic", 100.0);
    expect(fresh.getAssignment("patient_basic", "coach")).toBe("coach_basic");
  });

  it("reassignment replaces current member", () => {
    fresh.addMember("coach_orig", "coach", 5);
    fresh.addMember("coach_new", "coach", 5);
    fresh.assign("patient_reassign", "coach_orig", 100.0);
    fresh.assign("patient_reassign", "coach_new", 200.0);
    expect(fresh.getAssignment("patient_reassign", "coach")).toBe("coach_new");
  });

  it("reassignment removes patient from old member", () => {
    fresh.addMember("coach_from", "coach", 5);
    fresh.addMember("coach_to", "coach", 5);
    fresh.assign("patient_move", "coach_from", 100.0);
    fresh.assign("patient_move", "coach_to", 200.0);
    expect(fresh.getPatients("coach_from")).not.toContain("patient_move");
  });

  it("assignments across roles are independent", () => {
    fresh.addMember("coach_ind", "coach", 5);
    fresh.addMember("dr_ind", "physician", 5);
    fresh.assign("patient_ind", "coach_ind", 100.0);
    fresh.assign("patient_ind", "dr_ind", 100.0);
    expect(fresh.getAssignment("patient_ind", "coach")).toBe("coach_ind");
    expect(fresh.getAssignment("patient_ind", "physician")).toBe("dr_ind");
  });
});

describe("getAssignment", () => {
  let mgr: CareTeamManager;

  beforeEach(() => {
    mgr = makeMgr();
  });

  it("returns current member", () => {
    expect(mgr.getAssignment("patient_seed_1", "coach")).toBe("coach_alpha");
  });

  it("returns undefined for unassigned role", () => {
    expect(mgr.getAssignment("patient_seed_1", "dietitian")).toBeUndefined();
  });

  it("returns undefined for unknown patient", () => {
    expect(freshMgr().getAssignment("no_such_patient", "coach")).toBeUndefined();
  });
});

describe("getPatients", () => {
  let mgr: CareTeamManager;

  beforeEach(() => {
    mgr = makeMgr();
  });

  it("returns all assigned patients", () => {
    const patients = mgr.getPatients("coach_alpha");
    expect(patients).toContain("patient_seed_1");
    expect(patients).toContain("patient_seed_2");
  });

  it("result is sorted", () => {
    const patients = mgr.getPatients("coach_alpha");
    expect(patients).toEqual([...patients].sort());
  });

  it("returns empty list for member with no patients", () => {
    expect(mgr.getPatients("coach_beta")).toEqual([]);
  });

  it("unregistered member throws", () => {
    expect(() => freshMgr().getPatients("nobody")).toThrow();
  });
});

// ── Part 2: Capacity enforcement ─────────────────────────────────────────────

describe("capacity enforcement", () => {
  let fresh: CareTeamManager;

  beforeEach(() => {
    fresh = freshMgr();
  });

  it("throws CapacityError when member is full", () => {
    fresh.addMember("coach_full", "coach", 1);
    fresh.assign("patient_cap_1", "coach_full", 100.0);
    expect(() => fresh.assign("patient_cap_2", "coach_full", 200.0)).toThrow(CapacityError);
  });

  it("reassigning existing patient to same member does not throw", () => {
    fresh.addMember("coach_same", "coach", 1);
    fresh.assign("patient_same", "coach_same", 100.0);
    // Member is "full" but the patient is already theirs — must not throw
    expect(() => fresh.assign("patient_same", "coach_same", 200.0)).not.toThrow();
  });

  it("reassigning patient away frees capacity", () => {
    fresh.addMember("coach_donor", "coach", 1);
    fresh.addMember("coach_recv", "coach", 5);
    fresh.assign("patient_freed", "coach_donor", 100.0);
    // Move the patient away — coach_donor now has a free slot
    fresh.assign("patient_freed", "coach_recv", 200.0);
    // A new patient should now fit on coach_donor
    fresh.assign("patient_new", "coach_donor", 300.0);
    expect(fresh.getAssignment("patient_new", "coach")).toBe("coach_donor");
  });
});

describe("availableMembers", () => {
  let fresh: CareTeamManager;

  beforeEach(() => {
    fresh = freshMgr();
  });

  it("returns members with open capacity", () => {
    fresh.addMember("avail_coach_open", "coach", 2);
    fresh.addMember("avail_coach_full", "coach", 1);
    fresh.assign("avail_patient_1", "avail_coach_full", 100.0);
    const available = fresh.availableMembers("coach");
    expect(available).toContain("avail_coach_open");
    expect(available).not.toContain("avail_coach_full");
  });

  it("full member is excluded", () => {
    fresh.addMember("only_coach", "coach", 1);
    fresh.assign("only_patient", "only_coach", 100.0);
    expect(fresh.availableMembers("coach")).not.toContain("only_coach");
  });

  it("result is sorted", () => {
    fresh.addMember("sort_coach_z", "coach", 5);
    fresh.addMember("sort_coach_a", "coach", 5);
    const result = fresh.availableMembers("coach");
    expect(result).toEqual([...result].sort());
  });

  it("empty when no members with role", () => {
    fresh.addMember("solo_physician", "physician", 100);
    expect(fresh.availableMembers("coach")).toEqual([]);
  });
});

// ── Part 3: Assignment history ────────────────────────────────────────────────

describe("getHistory", () => {
  let fresh: CareTeamManager;

  beforeEach(() => {
    fresh = freshMgr();
  });

  it("single assignment has open end", () => {
    fresh.addMember("hist_coach_1", "coach", 5);
    fresh.assign("hist_patient_1", "hist_coach_1", 1000.0);
    expect(fresh.getHistory("hist_patient_1", "coach")).toEqual([["hist_coach_1", 1000.0, undefined]]);
  });

  it("reassignment closes previous entry", () => {
    fresh.addMember("hist_coach_a", "coach", 5);
    fresh.addMember("hist_coach_b", "coach", 5);
    fresh.assign("hist_patient_2", "hist_coach_a", 1000.0);
    fresh.assign("hist_patient_2", "hist_coach_b", 3000.0);
    expect(fresh.getHistory("hist_patient_2", "coach")).toEqual([
      ["hist_coach_a", 1000.0, 3000.0],
      ["hist_coach_b", 3000.0, undefined],
    ]);
  });

  it("multiple reassignments sorted chronologically", () => {
    for (const name of ["hist_cx", "hist_cy", "hist_cz"]) {
      fresh.addMember(name, "coach", 5);
    }
    fresh.assign("hist_patient_3", "hist_cx", 100.0);
    fresh.assign("hist_patient_3", "hist_cy", 200.0);
    fresh.assign("hist_patient_3", "hist_cz", 300.0);
    const history = fresh.getHistory("hist_patient_3", "coach");
    expect(history.map((entry) => entry[1])).toEqual([100.0, 200.0, 300.0]);
    expect(history[history.length - 1][2]).toBeUndefined();
  });

  it("returns empty for unassigned role", () => {
    fresh.addMember("hist_coach_only", "coach", 5);
    fresh.assign("hist_patient_4", "hist_coach_only", 100.0);
    expect(fresh.getHistory("hist_patient_4", "physician")).toEqual([]);
  });

  it("returns empty for unknown patient", () => {
    expect(fresh.getHistory("hist_nobody", "coach")).toEqual([]);
  });
});

describe("getAssignmentAt", () => {
  let fresh: CareTeamManager;

  beforeEach(() => {
    fresh = freshMgr();
  });

  it("returns member during active window", () => {
    fresh.addMember("at_coach_1", "coach", 5);
    fresh.assign("at_patient_1", "at_coach_1", 1000.0);
    expect(fresh.getAssignmentAt("at_patient_1", "coach", 2000.0)).toBe("at_coach_1");
  });

  it("returns undefined before first assignment", () => {
    fresh.addMember("at_coach_2", "coach", 5);
    fresh.assign("at_patient_2", "at_coach_2", 1000.0);
    expect(fresh.getAssignmentAt("at_patient_2", "coach", 500.0)).toBeUndefined();
  });

  it("returns original member before reassignment", () => {
    fresh.addMember("at_coach_old", "coach", 5);
    fresh.addMember("at_coach_new", "coach", 5);
    fresh.assign("at_patient_3", "at_coach_old", 1000.0);
    fresh.assign("at_patient_3", "at_coach_new", 3000.0);
    expect(fresh.getAssignmentAt("at_patient_3", "coach", 2000.0)).toBe("at_coach_old");
  });

  it("returns new member after reassignment", () => {
    fresh.addMember("at_coach_prev", "coach", 5);
    fresh.addMember("at_coach_curr", "coach", 5);
    fresh.assign("at_patient_4", "at_coach_prev", 1000.0);
    fresh.assign("at_patient_4", "at_coach_curr", 3000.0);
    expect(fresh.getAssignmentAt("at_patient_4", "coach", 4000.0)).toBe("at_coach_curr");
  });

  it("returns undefined for unassigned role", () => {
    fresh.addMember("at_coach_role", "coach", 5);
    fresh.assign("at_patient_5", "at_coach_role", 1000.0);
    expect(fresh.getAssignmentAt("at_patient_5", "physician", 2000.0)).toBeUndefined();
  });
});
