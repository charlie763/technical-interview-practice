/**
 * Tests for Problem 14: Contract Amendment Manager
 *
 * Run from the typescript/ directory:
 *   npm run test:14
 */

import { describe, expect, it } from "vitest";
import { ContractAmendmentManager } from "@problems/problem_14_contract_amendment";

// ---------------------------------------------------------------------------
// Shared dates
// ---------------------------------------------------------------------------
const D_BASE = "2025-01-01"; // base / before any amendments
const D_AMD1 = "2025-03-01"; // amendment 1 effective date
const D_AMD2 = "2025-06-01"; // amendment 2 effective date
const D_AMD3 = "2025-09-01"; // amendment 3 effective date
const D_AFTER = "2025-12-31"; // well after all amendments

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** Empty ContractAmendmentManager. */
function freshMgr(): ContractAmendmentManager {
  return new ContractAmendmentManager();
}

/**
 * Pre-seeded manager:
 *   c-seed-1  "Vendor MSA"
 *             fields = { value: 50000, paymentTerms: "net-30", currency: "USD" }
 *             amd-s1 (2025-03-01): overrides = { paymentTerms: "net-45" }
 *             amd-s2 (2025-06-01): overrides = { value: 75000 }
 *
 *   c-seed-2  "NDA Agreement"
 *             fields = { termYears: 2, autoRenew: true }
 *             (no amendments)
 */
function makeMgr(): ContractAmendmentManager {
  const m = new ContractAmendmentManager();
  m.addContract("c-seed-1", "Vendor MSA", { value: 50000, paymentTerms: "net-30", currency: "USD" });
  m.addContract("c-seed-2", "NDA Agreement", { termYears: 2, autoRenew: true });
  m.addAmendment("amd-s1", "c-seed-1", D_AMD1, { paymentTerms: "net-45" }, "extended payment terms");
  m.addAmendment("amd-s2", "c-seed-1", D_AMD2, { value: 75000 }, "scope increase");
  return m;
}

// ---------------------------------------------------------------------------
// PART 1 — Base contract management
// ---------------------------------------------------------------------------

describe("addContract", () => {
  it("returns contract dict", () => {
    const fresh = freshMgr();
    const c = fresh.addContract("c-add-1", "Test", { x: 1 });
    expect(c.contractId).toBe("c-add-1");
    expect(c.title).toBe("Test");
    expect(c.fields["x"]).toBe(1);
  });

  it("stores copy of fields", () => {
    const fresh = freshMgr();
    const original: Record<string, unknown> = { x: 1 };
    fresh.addContract("c-copy-1", "Test", original);
    original["x"] = 999;
    expect(fresh.getBaseContract("c-copy-1").fields["x"]).toBe(1);
  });

  it("duplicate throws", () => {
    const fresh = freshMgr();
    fresh.addContract("c-dup-am", "A", {});
    expect(() => fresh.addContract("c-dup-am", "B", {})).toThrow();
  });
});

describe("getBaseContract", () => {
  it("returns original fields", () => {
    const mgr = makeMgr();
    const base = mgr.getBaseContract("c-seed-1");
    // Even though amendments exist, base fields are unchanged
    expect(base.fields["paymentTerms"]).toBe("net-30");
    expect(base.fields["value"]).toBe(50000);
  });

  it("unknown throws", () => {
    const mgr = makeMgr();
    expect(() => mgr.getBaseContract("no-such")).toThrow();
  });
});

// ---------------------------------------------------------------------------
// PART 2 — Amendments and effective contract
// ---------------------------------------------------------------------------

describe("addAmendment", () => {
  it("returns amendment dict", () => {
    const mgr = makeMgr();
    const amd = mgr.addAmendment("amd-add-1", "c-seed-1", D_AMD3, { currency: "EUR" }, "switch currency");
    expect(amd.amendmentId).toBe("amd-add-1");
    expect(amd.contractId).toBe("c-seed-1");
    expect(amd.effectiveOn).toBe(D_AMD3);
    expect(amd.overrides).toEqual({ currency: "EUR" });
    expect(amd.note).toBe("switch currency");
  });

  it("stores copy of overrides", () => {
    const fresh = freshMgr();
    fresh.addContract("c-amd-copy", "T", { x: 1 });
    const overrides: Record<string, unknown> = { x: 2 };
    fresh.addAmendment("amd-copy-1", "c-amd-copy", D_AMD1, overrides, "");
    overrides["x"] = 999;
    const amendments = fresh.getAmendments("c-amd-copy");
    expect(amendments[0].overrides["x"]).toBe(2);
  });

  it("duplicate amendment throws", () => {
    const mgr = makeMgr();
    expect(() => mgr.addAmendment("amd-s1", "c-seed-1", D_AMD3, {}, "")).toThrow();
  });

  it("unknown contract throws", () => {
    const mgr = makeMgr();
    expect(() => mgr.addAmendment("amd-new", "no-such", D_AMD1, {}, "")).toThrow();
  });
});

describe("getAmendments", () => {
  it("sorted by effectiveOn", () => {
    const mgr = makeMgr();
    const amendments = mgr.getAmendments("c-seed-1");
    const dates = amendments.map((a) => a.effectiveOn);
    expect(dates).toEqual([...dates].sort());
  });

  it("empty when none", () => {
    const mgr = makeMgr();
    expect(mgr.getAmendments("c-seed-2")).toEqual([]);
  });

  it("unknown throws", () => {
    const mgr = makeMgr();
    expect(() => mgr.getAmendments("no-such")).toThrow();
  });

  it("same date ordered by amendmentId", () => {
    const fresh = freshMgr();
    fresh.addContract("c-same-dt", "T", { x: 0 });
    fresh.addAmendment("amd-z", "c-same-dt", D_AMD1, { x: 2 }, "");
    fresh.addAmendment("amd-a", "c-same-dt", D_AMD1, { x: 1 }, "");
    const amendments = fresh.getAmendments("c-same-dt");
    const ids = amendments.map((a) => a.amendmentId);
    expect(ids).toEqual([...ids].sort());
  });
});

describe("getEffectiveContract", () => {
  it("before any amendments", () => {
    const mgr = makeMgr();
    const fields = mgr.getEffectiveContract("c-seed-1", D_BASE);
    expect(fields["paymentTerms"]).toBe("net-30");
    expect(fields["value"]).toBe(50000);
  });

  it("after first amendment", () => {
    const mgr = makeMgr();
    const fields = mgr.getEffectiveContract("c-seed-1", "2025-04-15");
    expect(fields["paymentTerms"]).toBe("net-45");
    expect(fields["value"]).toBe(50000); // amd-s2 not yet effective
  });

  it("after all amendments", () => {
    const mgr = makeMgr();
    const fields = mgr.getEffectiveContract("c-seed-1", D_AFTER);
    expect(fields["paymentTerms"]).toBe("net-45");
    expect(fields["value"]).toBe(75000);
  });

  it("original unamended fields preserved", () => {
    const mgr = makeMgr();
    const fields = mgr.getEffectiveContract("c-seed-1", D_AFTER);
    expect(fields["currency"]).toBe("USD");
  });

  it("no amendments returns base", () => {
    const mgr = makeMgr();
    const fields = mgr.getEffectiveContract("c-seed-2", D_AFTER);
    expect(fields["termYears"]).toBe(2);
    expect(fields["autoRenew"]).toBe(true);
  });

  it("exact effective date is inclusive", () => {
    // amd-s1 effectiveOn = D_AMD1; querying exactly D_AMD1 should apply it
    const mgr = makeMgr();
    const fields = mgr.getEffectiveContract("c-seed-1", D_AMD1);
    expect(fields["paymentTerms"]).toBe("net-45");
  });

  it("does not mutate base", () => {
    const mgr = makeMgr();
    mgr.getEffectiveContract("c-seed-1", D_AFTER);
    const base = mgr.getBaseContract("c-seed-1");
    expect(base.fields["paymentTerms"]).toBe("net-30");
  });

  it("unknown throws", () => {
    const mgr = makeMgr();
    expect(() => mgr.getEffectiveContract("no-such", D_AFTER)).toThrow();
  });
});

// ---------------------------------------------------------------------------
// PART 3 — Value history and amendment summary
// ---------------------------------------------------------------------------

describe("getValueHistory", () => {
  it("base value included first", () => {
    const mgr = makeMgr();
    const history = mgr.getValueHistory("c-seed-1", "value");
    expect(history[0].source).toBe("base");
    expect(history[0].value).toBe(50000);
  });

  it("amendment override included", () => {
    const mgr = makeMgr();
    const history = mgr.getValueHistory("c-seed-1", "paymentTerms");
    const sources = history.map((e) => e.source);
    expect(sources).toContain("amd-s1");
  });

  it("only amendments that touched field", () => {
    // amd-s2 changes "value", not "paymentTerms"
    const mgr = makeMgr();
    const history = mgr.getValueHistory("c-seed-1", "paymentTerms");
    const sources = history.map((e) => e.source);
    expect(sources).not.toContain("amd-s2");
  });

  it("sorted chronologically", () => {
    const mgr = makeMgr();
    const history = mgr.getValueHistory("c-seed-1", "value");
    // base first, then amendment entries in date order
    expect(history[0].source).toBe("base");
    const dates = history.filter((e) => e.source !== "base").map((e) => e.effectiveOn);
    expect(dates).toEqual([...dates].sort());
  });

  it("field not present throws", () => {
    const mgr = makeMgr();
    expect(() => mgr.getValueHistory("c-seed-1", "nonexistent_field")).toThrow();
  });

  it("unknown contract throws", () => {
    const mgr = makeMgr();
    expect(() => mgr.getValueHistory("no-such", "value")).toThrow();
  });
});

describe("getAmendmentSummary", () => {
  it("amendmentCount", () => {
    const mgr = makeMgr();
    const summary = mgr.getAmendmentSummary("c-seed-1");
    expect(summary.amendmentCount).toBe(2);
  });

  it("fieldsAmended sorted", () => {
    const mgr = makeMgr();
    const summary = mgr.getAmendmentSummary("c-seed-1");
    expect(summary.fieldsAmended).toEqual(["paymentTerms", "value"].sort());
  });

  it("latestAmendment date", () => {
    const mgr = makeMgr();
    const summary = mgr.getAmendmentSummary("c-seed-1");
    expect(summary.latestAmendment).toBe(D_AMD2);
  });

  it("currentFields reflect all amendments", () => {
    const mgr = makeMgr();
    const summary = mgr.getAmendmentSummary("c-seed-1");
    expect(summary.currentFields["paymentTerms"]).toBe("net-45");
    expect(summary.currentFields["value"]).toBe(75000);
  });

  it("no amendments returns undefined latest", () => {
    const mgr = makeMgr();
    const summary = mgr.getAmendmentSummary("c-seed-2");
    expect(summary.latestAmendment).toBeUndefined();
    expect(summary.amendmentCount).toBe(0);
    expect(summary.fieldsAmended).toEqual([]);
  });

  it("unknown throws", () => {
    const mgr = makeMgr();
    expect(() => mgr.getAmendmentSummary("no-such")).toThrow();
  });
});
