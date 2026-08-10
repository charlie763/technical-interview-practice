/**
 * Tests for Problem 13: Contract Lifecycle State Machine
 *
 * Run from the typescript/ directory:
 *   npm run test:13
 */

import { describe, expect, it } from "vitest";
import { ContractLifecycle } from "@problems/problem_13_contract_lifecycle";

// ---------------------------------------------------------------------------
// Shared timestamps
// T0 = base datetime
// T1 = T0 + 1 day
// T2 = T0 + 5 days
// T3 = T0 + 10 days
// T4 = T0 + 40 days  (> 30 days for overdue tests)
// T5 = T0 + 45 days
// ---------------------------------------------------------------------------
const T0 = "2025-01-01T09:00:00";
const T1 = "2025-01-02T10:00:00";
const T2 = "2025-01-06T11:00:00";
const T3 = "2025-01-11T12:00:00";
const T4 = "2025-02-10T09:00:00"; // 40 days after T0
const T5 = "2025-02-15T09:00:00"; // 45 days after T0

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** Empty ContractLifecycle. */
function freshCl(): ContractLifecycle {
  return new ContractLifecycle();
}

/**
 * Pre-seeded lifecycle manager:
 *   c-seed-1  "Vendor MSA"        state=approved    (draft→in_review→approved)
 *   c-seed-2  "NDA Agreement"     state=draft
 *   c-seed-3  "SaaS License"      state=terminated  (draft→in_review→approved→executed→active→terminated)
 */
function makeCl(): ContractLifecycle {
  const c = new ContractLifecycle();
  // c-seed-1: draft → in_review → approved
  c.createContract("c-seed-1", "Vendor MSA", T0, "alice");
  c.transition("c-seed-1", "in_review", T1, "alice");
  c.transition("c-seed-1", "approved", T2, "bob");
  // c-seed-2: stays draft
  c.createContract("c-seed-2", "NDA Agreement", T0, "alice");
  // c-seed-3: full path to terminated
  c.createContract("c-seed-3", "SaaS License", T0, "carol");
  c.transition("c-seed-3", "in_review", T1, "carol");
  c.transition("c-seed-3", "approved", T2, "bob");
  c.transition("c-seed-3", "executed", T3, "carol");
  c.transition("c-seed-3", "active", T4, "carol");
  c.transition("c-seed-3", "terminated", T5, "carol");
  return c;
}

// ---------------------------------------------------------------------------
// PART 1 — Contract creation, field management, transitions
// ---------------------------------------------------------------------------

describe("createContract", () => {
  it("returns contract in draft", () => {
    const fresh = freshCl();
    const c = fresh.createContract("c-cr-1", "Title", T0, "alice");
    expect(c.contractId).toBe("c-cr-1");
    expect(c.state).toBe("draft");
    expect(c.title).toBe("Title");
    expect(c.fields).toEqual({});
  });

  it("stores createdAt", () => {
    const fresh = freshCl();
    const c = fresh.createContract("c-cr-ts", "Title", T0, "alice");
    expect(c.createdAt).toBe(T0);
  });

  it("duplicate throws", () => {
    const fresh = freshCl();
    fresh.createContract("c-dup-lc", "A", T0, "alice");
    expect(() => fresh.createContract("c-dup-lc", "B", T1, "alice")).toThrow();
  });

  it("initial audit entry recorded", () => {
    const fresh = freshCl();
    fresh.createContract("c-audit-init", "T", T0, "alice");
    const trail = fresh.getAuditTrail("c-audit-init");
    expect(trail).toHaveLength(1);
    expect(trail[0].fromState).toBeUndefined();
    expect(trail[0].toState).toBe("draft");
    expect(trail[0].actor).toBe("alice");
  });
});

describe("setField", () => {
  it("sets field", () => {
    const cl = makeCl();
    cl.setField("c-seed-1", "value", 100000);
    expect(cl.getContract("c-seed-1").fields["value"]).toBe(100000);
  });

  it("updates existing field", () => {
    const cl = makeCl();
    cl.setField("c-seed-1", "value", 50000);
    cl.setField("c-seed-1", "value", 75000);
    expect(cl.getContract("c-seed-1").fields["value"]).toBe(75000);
  });

  it("unknown contract throws", () => {
    const cl = makeCl();
    expect(() => cl.setField("no-such", "key", "val")).toThrow();
  });
});

describe("getContract", () => {
  it("returns contract", () => {
    const cl = makeCl();
    const c = cl.getContract("c-seed-1");
    expect(c.contractId).toBe("c-seed-1");
    expect(c.state).toBe("approved");
  });

  it("unknown throws", () => {
    const cl = makeCl();
    expect(() => cl.getContract("nonexistent")).toThrow();
  });
});

describe("transition", () => {
  it("valid transition updates state", () => {
    const fresh = freshCl();
    fresh.createContract("c-tr-1", "T", T0, "alice");
    fresh.transition("c-tr-1", "in_review", T1, "alice");
    expect(fresh.getContract("c-tr-1").state).toBe("in_review");
  });

  it("invalid transition throws", () => {
    const fresh = freshCl();
    fresh.createContract("c-tr-inv", "T", T0, "alice");
    expect(() => fresh.transition("c-tr-inv", "approved", T1, "alice")).toThrow(); // draft→approved invalid
  });

  it("terminal state throws", () => {
    const cl = makeCl();
    expect(() => cl.transition("c-seed-3", "draft", T5, "alice")).toThrow(); // terminated is terminal
  });

  it("back transition draft to in_review to draft", () => {
    const fresh = freshCl();
    fresh.createContract("c-back", "T", T0, "alice");
    fresh.transition("c-back", "in_review", T1, "alice");
    fresh.transition("c-back", "draft", T2, "alice");
    expect(fresh.getContract("c-back").state).toBe("draft");
  });

  it("transition appends audit entry", () => {
    const fresh = freshCl();
    fresh.createContract("c-tr-audit", "T", T0, "alice");
    fresh.transition("c-tr-audit", "in_review", T1, "bob");
    const trail = fresh.getAuditTrail("c-tr-audit");
    expect(trail).toHaveLength(2);
    const last = trail[trail.length - 1];
    expect(last.fromState).toBe("draft");
    expect(last.toState).toBe("in_review");
    expect(last.actor).toBe("bob");
  });

  it("unknown contract throws", () => {
    const cl = makeCl();
    expect(() => cl.transition("no-such", "in_review", T1, "alice")).toThrow();
  });
});

// ---------------------------------------------------------------------------
// PART 2 — Audit trail, by-state query, bulk advance
// ---------------------------------------------------------------------------

describe("getAuditTrail", () => {
  it("returns all entries ordered", () => {
    const cl = makeCl();
    const trail = cl.getAuditTrail("c-seed-1");
    // create(draft) + in_review + approved = 3 entries
    expect(trail).toHaveLength(3);
    const toStates = trail.map((e) => e.toState);
    expect(toStates).toEqual(["draft", "in_review", "approved"]);
  });

  it("terminal contract full trail", () => {
    const cl = makeCl();
    const trail = cl.getAuditTrail("c-seed-3");
    const toStates = trail.map((e) => e.toState);
    expect(toStates).toEqual(["draft", "in_review", "approved", "executed", "active", "terminated"]);
  });

  it("unknown contract throws", () => {
    const cl = makeCl();
    expect(() => cl.getAuditTrail("no-such")).toThrow();
  });
});

describe("getContractsByState", () => {
  it("returns correct contracts", () => {
    const cl = makeCl();
    const approved = cl.getContractsByState("approved");
    const ids = approved.map((c) => c.contractId);
    expect(ids).toContain("c-seed-1");
    expect(ids).not.toContain("c-seed-2");
  });

  it("sorted by contractId", () => {
    const fresh = freshCl();
    fresh.createContract("c-z", "Z", T0, "a");
    fresh.createContract("c-a", "A", T0, "a");
    fresh.createContract("c-m", "M", T0, "a");
    const drafts = fresh.getContractsByState("draft");
    const ids = drafts.map((c) => c.contractId);
    expect(ids).toEqual([...ids].sort());
  });

  it("empty for unused state", () => {
    const cl = makeCl();
    expect(cl.getContractsByState("expired")).toEqual([]);
  });
});

describe("bulkAdvance", () => {
  it("all succeed", () => {
    const fresh = freshCl();
    for (let i = 0; i < 3; i++) {
      fresh.createContract(`c-bulk-${i}`, `Contract ${i}`, T0, "alice");
    }
    const result = fresh.bulkAdvance(["c-bulk-0", "c-bulk-1", "c-bulk-2"], "in_review", T1, "alice");
    expect(result.succeeded).toHaveLength(3);
    expect(result.failed).toHaveLength(0);
  });

  it("partial failure continues", () => {
    // c-ok is draft; c-bad is already in_review (can't go back to draft via bulk)
    const fresh = freshCl();
    fresh.createContract("c-ok", "OK", T0, "alice");
    fresh.createContract("c-bad", "Bad", T0, "alice");
    fresh.transition("c-bad", "in_review", T1, "alice");
    const result = fresh.bulkAdvance(["c-ok", "c-bad"], "in_review", T2, "alice");
    expect(result.succeeded).toContain("c-ok");
    expect(result.failed.some((f) => f.contractId === "c-bad")).toBe(true);
  });

  it("states updated for successes", () => {
    const fresh = freshCl();
    fresh.createContract("c-bs-1", "A", T0, "alice");
    fresh.createContract("c-bs-2", "B", T0, "alice");
    fresh.bulkAdvance(["c-bs-1", "c-bs-2"], "in_review", T1, "alice");
    expect(fresh.getContract("c-bs-1").state).toBe("in_review");
    expect(fresh.getContract("c-bs-2").state).toBe("in_review");
  });

  it("failed entry includes reason", () => {
    const fresh = freshCl();
    fresh.createContract("c-fail-r", "X", T0, "alice");
    const result = fresh.bulkAdvance(["c-fail-r"], "approved", T1, "alice");
    expect(result.failed).toHaveLength(1);
    expect(result.failed[0].reason).not.toBe("");
  });
});

// ---------------------------------------------------------------------------
// PART 3 — Lifecycle metrics and overdue contracts
// ---------------------------------------------------------------------------

describe("getLifecycleMetrics", () => {
  it("total count", () => {
    const cl = makeCl();
    const metrics = cl.getLifecycleMetrics();
    expect(metrics.total).toBe(3);
  });

  it("byState counts", () => {
    const cl = makeCl();
    const metrics = cl.getLifecycleMetrics();
    expect(metrics.byState["approved"]).toBe(1);
    expect(metrics.byState["draft"]).toBe(1);
    expect(metrics.byState["terminated"]).toBe(1);
  });

  it("only nonzero states in byState", () => {
    const cl = makeCl();
    const metrics = cl.getLifecycleMetrics();
    for (const count of Object.values(metrics.byState)) {
      expect(count).toBeGreaterThan(0);
    }
  });

  it("terminalCount", () => {
    const cl = makeCl();
    const metrics = cl.getLifecycleMetrics();
    // c-seed-3 is terminated
    expect(metrics.terminalCount).toBe(1);
  });

  it("empty manager returns zeros", () => {
    const fresh = freshCl();
    const metrics = fresh.getLifecycleMetrics();
    expect(metrics.total).toBe(0);
    expect(metrics.byState).toEqual({});
    expect(metrics.terminalCount).toBe(0);
  });
});

describe("getOverdueContracts", () => {
  it("returns contracts stuck over 30 days", () => {
    // Create contract that transitions to in_review at T0, then nothing
    const fresh = freshCl();
    fresh.createContract("c-overdue", "Old Contract", T0, "alice");
    fresh.transition("c-overdue", "in_review", T0, "alice");
    // T4 = T0 + 40 days → should be overdue
    const overdue = fresh.getOverdueContracts(T4);
    const ids = overdue.map((o) => o.contractId);
    expect(ids).toContain("c-overdue");
  });

  it("recent transition not overdue", () => {
    const fresh = freshCl();
    fresh.createContract("c-recent", "New Contract", T0, "alice");
    fresh.transition("c-recent", "in_review", T3, "alice");
    // T4 = T3 + ~30 days; T4-T3 = T0+40 - T0+10 = 30 days exactly; use T4 + 1 extra day
    const overdue = fresh.getOverdueContracts(T4);
    const ids = overdue.map((o) => o.contractId);
    expect(ids).not.toContain("c-recent");
  });

  it("terminal contracts excluded", () => {
    const cl = makeCl();
    const overdue = cl.getOverdueContracts(T5);
    const ids = overdue.map((o) => o.contractId);
    expect(ids).not.toContain("c-seed-3"); // terminated → terminal
  });

  it("sorted by daysStuck descending", () => {
    const fresh = freshCl();
    fresh.createContract("c-od-a", "A", T0, "alice");
    fresh.createContract("c-od-b", "B", T0, "alice");
    fresh.transition("c-od-a", "in_review", T0, "alice");
    fresh.transition("c-od-b", "in_review", T1, "alice");
    const overdue = fresh.getOverdueContracts(T4);
    const days = overdue.map((o) => o.daysStuck);
    expect(days).toEqual([...days].sort((a, b) => b - a));
  });

  it("result includes required fields", () => {
    const fresh = freshCl();
    fresh.createContract("c-od-f", "Fields Test", T0, "alice");
    fresh.transition("c-od-f", "in_review", T0, "alice");
    const overdue = fresh.getOverdueContracts(T4);
    const entry = overdue.find((o) => o.contractId === "c-od-f")!;
    expect(entry.title).toBeDefined();
    expect(entry.state).toBeDefined();
    expect(entry.stuckSince).toBeDefined();
    expect(entry.daysStuck).toBeDefined();
    expect(entry.daysStuck).toBeGreaterThanOrEqual(31);
  });
});
