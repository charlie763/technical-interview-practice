/**
 * Tests for Problem 12: Contract Expiration Alert Scheduler
 *
 * Run from the typescript/ directory:
 *   npm run test:12
 */

import { describe, expect, it } from "vitest";
import { ContractAlertScheduler } from "@problems/problem_12_contract_alert_scheduler";

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** Empty ContractAlertScheduler. */
function freshSched(): ContractAlertScheduler {
  return new ContractAlertScheduler();
}

/**
 * Pre-seeded scheduler:
 *   c-seed-1  "Vendor MSA"        owner=legal@acme.com  expires=2025-06-30
 *   c-seed-2  "SaaS Subscription" owner=ops@acme.com    expires=2025-09-15
 *   c-seed-3  "NDA Agreement"     owner=legal@acme.com  expires=2025-12-31
 *
 * Alert configs:
 *   cfg-30   daysBefore=30  label="30-day notice"
 *   cfg-7    daysBefore=7   label="final warning"
 */
function makeSched(): ContractAlertScheduler {
  const s = new ContractAlertScheduler();
  s.addContract("c-seed-1", "Vendor MSA", "legal@acme.com", "2025-06-30");
  s.addContract("c-seed-2", "SaaS Subscription", "ops@acme.com", "2025-09-15");
  s.addContract("c-seed-3", "NDA Agreement", "legal@acme.com", "2025-12-31");
  s.addAlertConfig("cfg-30", 30, "30-day notice");
  s.addAlertConfig("cfg-7", 7, "final warning");
  return s;
}

// ---------------------------------------------------------------------------
// PART 1 — Contract and alert-config management
// ---------------------------------------------------------------------------

describe("addContract", () => {
  it("returns contract dict", () => {
    const fresh = freshSched();
    const c = fresh.addContract("c-add-1", "Test Contract", "a@b.com", "2025-06-01");
    expect(c.contractId).toBe("c-add-1");
    expect(c.title).toBe("Test Contract");
    expect(c.ownerEmail).toBe("a@b.com");
    expect(c.expiresOn).toBe("2025-06-01");
  });

  it("duplicate throws", () => {
    const fresh = freshSched();
    fresh.addContract("c-dup-1", "A", "a@b.com", "2025-06-01");
    expect(() => fresh.addContract("c-dup-1", "B", "b@c.com", "2025-07-01")).toThrow();
  });

  it("multiple contracts stored", () => {
    // sched has 3 contracts; retrieve their expiry dates via expiringBetween
    const sched = makeSched();
    const results = sched.getContractsExpiringBetween("2025-01-01", "2025-12-31");
    const ids = results.map((r) => r.contractId);
    expect(ids).toContain("c-seed-1");
    expect(ids).toContain("c-seed-2");
    expect(ids).toContain("c-seed-3");
  });
});

describe("addAlertConfig", () => {
  it("returns config dict", () => {
    const fresh = freshSched();
    const cfg = fresh.addAlertConfig("cfg-add-1", 14, "two-week notice");
    expect(cfg.configId).toBe("cfg-add-1");
    expect(cfg.daysBefore).toBe(14);
    expect(cfg.label).toBe("two-week notice");
  });

  it("duplicate throws", () => {
    const fresh = freshSched();
    fresh.addAlertConfig("cfg-dup-1", 30, "notice");
    expect(() => fresh.addAlertConfig("cfg-dup-1", 60, "other")).toThrow();
  });
});

describe("getContractsExpiringBetween", () => {
  it("exact range match", () => {
    const sched = makeSched();
    const results = sched.getContractsExpiringBetween("2025-06-30", "2025-06-30");
    expect(results).toHaveLength(1);
    expect(results[0].contractId).toBe("c-seed-1");
  });

  it("range spans multiple", () => {
    const sched = makeSched();
    const results = sched.getContractsExpiringBetween("2025-06-01", "2025-09-30");
    const ids = results.map((r) => r.contractId);
    expect(ids).toContain("c-seed-1");
    expect(ids).toContain("c-seed-2");
    expect(ids).not.toContain("c-seed-3");
  });

  it("sorted ascending", () => {
    const sched = makeSched();
    const results = sched.getContractsExpiringBetween("2025-01-01", "2025-12-31");
    const dates = results.map((r) => r.expiresOn);
    expect(dates).toEqual([...dates].sort());
  });

  it("empty when none in range", () => {
    const sched = makeSched();
    const results = sched.getContractsExpiringBetween("2024-01-01", "2024-12-31");
    expect(results).toEqual([]);
  });

  it("inclusive start boundary", () => {
    // c-seed-1 expires exactly on 2025-06-30; start = 2025-06-30 should include it
    const sched = makeSched();
    const results = sched.getContractsExpiringBetween("2025-06-30", "2025-12-31");
    const ids = results.map((r) => r.contractId);
    expect(ids).toContain("c-seed-1");
  });

  it("inclusive end boundary", () => {
    const sched = makeSched();
    const results = sched.getContractsExpiringBetween("2025-01-01", "2025-06-30");
    const ids = results.map((r) => r.contractId);
    expect(ids).toContain("c-seed-1");
  });
});

// ---------------------------------------------------------------------------
// PART 2 — Alert schedule computation
// ---------------------------------------------------------------------------

describe("computeAlertSchedule", () => {
  it("returns entry per config", () => {
    const sched = makeSched();
    const schedule = sched.computeAlertSchedule("c-seed-1");
    // 2 configs registered → 2 schedule entries
    expect(schedule).toHaveLength(2);
  });

  it("alertOn dates correct", () => {
    // c-seed-1 expires 2025-06-30
    // cfg-30: 2025-06-30 - 30d = 2025-05-31
    // cfg-7:  2025-06-30 - 7d  = 2025-06-23
    const sched = makeSched();
    const schedule = sched.computeAlertSchedule("c-seed-1");
    const byCfg = Object.fromEntries(schedule.map((e) => [e.configId, e]));
    expect(byCfg["cfg-30"].alertOn).toBe("2025-05-31");
    expect(byCfg["cfg-7"].alertOn).toBe("2025-06-23");
  });

  it("sorted by alertOn ascending", () => {
    const sched = makeSched();
    const schedule = sched.computeAlertSchedule("c-seed-1");
    const dates = schedule.map((e) => e.alertOn);
    expect(dates).toEqual([...dates].sort());
  });

  it("includes label", () => {
    const sched = makeSched();
    const schedule = sched.computeAlertSchedule("c-seed-1");
    const labels = new Set(schedule.map((e) => e.label));
    expect(labels.has("30-day notice")).toBe(true);
    expect(labels.has("final warning")).toBe(true);
  });

  it("unknown contract throws", () => {
    const sched = makeSched();
    expect(() => sched.computeAlertSchedule("no-such-contract")).toThrow();
  });

  it("no configs returns empty list", () => {
    const fresh = freshSched();
    fresh.addContract("c-no-cfg", "Bare Contract", "a@b.com", "2025-06-01");
    expect(fresh.computeAlertSchedule("c-no-cfg")).toEqual([]);
  });
});

describe("getDueAlerts", () => {
  it("returns alerts on or before date", () => {
    // cfg-30 for c-seed-1 fires on 2025-05-31
    const sched = makeSched();
    const due = sched.getDueAlerts("2025-05-31");
    const entries = due.map((e) => `${e.contractId}:${e.configId}`);
    expect(entries).toContain("c-seed-1:cfg-30");
  });

  it("excludes future alerts", () => {
    // 2025-01-01 → no alerts yet for any contract
    const sched = makeSched();
    const due = sched.getDueAlerts("2025-01-01");
    expect(due).toEqual([]);
  });

  it("includes ownerEmail and expiresOn", () => {
    const sched = makeSched();
    const due = sched.getDueAlerts("2025-05-31");
    const entry = due.find((e) => e.contractId === "c-seed-1" && e.configId === "cfg-30")!;
    expect(entry.ownerEmail).toBe("legal@acme.com");
    expect(entry.expiresOn).toBe("2025-06-30");
  });

  it("sorted by alertOn then contractId", () => {
    // Ask for a date far enough in the future to capture many alerts
    const sched = makeSched();
    const due = sched.getDueAlerts("2025-12-31");
    const dates = due.map((e) => e.alertOn);
    expect(dates).toEqual([...dates].sort());
  });
});

// ---------------------------------------------------------------------------
// PART 3 — Sent records and upcoming alerts
// ---------------------------------------------------------------------------

describe("recordAlertSent", () => {
  it("returns sent record", () => {
    const sched = makeSched();
    const rec = sched.recordAlertSent("c-seed-1", "cfg-30", "2025-05-31");
    expect(rec.contractId).toBe("c-seed-1");
    expect(rec.configId).toBe("cfg-30");
    expect(rec.sentOn).toBe("2025-05-31");
  });

  it("unknown contract throws", () => {
    const sched = makeSched();
    expect(() => sched.recordAlertSent("no-contract", "cfg-30", "2025-05-31")).toThrow();
  });

  it("unknown config throws", () => {
    const sched = makeSched();
    expect(() => sched.recordAlertSent("c-seed-1", "no-cfg", "2025-05-31")).toThrow();
  });

  it("multiple sends stored", () => {
    const sched = makeSched();
    sched.recordAlertSent("c-seed-1", "cfg-30", "2025-05-31");
    sched.recordAlertSent("c-seed-1", "cfg-7", "2025-06-23");
    const upcoming = sched.getUpcomingAlerts("c-seed-1", "2025-05-01");
    const sentIds = new Set(upcoming.filter((e) => e.sent).map((e) => e.configId));
    expect(sentIds.has("cfg-30")).toBe(true);
    expect(sentIds.has("cfg-7")).toBe(true);
  });
});

describe("getUpcomingAlerts", () => {
  it("excludes past alerts", () => {
    // asOfDate = 2025-06-01; cfg-30 alertOn=2025-05-31 is in the past
    const sched = makeSched();
    const upcoming = sched.getUpcomingAlerts("c-seed-1", "2025-06-01");
    const configIds = upcoming.map((e) => e.configId);
    expect(configIds).not.toContain("cfg-30");
  });

  it("includes future alerts", () => {
    // cfg-7 alertOn=2025-06-23 is still upcoming from 2025-06-01
    const sched = makeSched();
    const upcoming = sched.getUpcomingAlerts("c-seed-1", "2025-06-01");
    const configIds = upcoming.map((e) => e.configId);
    expect(configIds).toContain("cfg-7");
  });

  it("sent flag false by default", () => {
    const sched = makeSched();
    const upcoming = sched.getUpcomingAlerts("c-seed-1", "2025-05-01");
    for (const e of upcoming) {
      expect(e.sent).toBe(false);
    }
  });

  it("sent flag true after record", () => {
    const sched = makeSched();
    sched.recordAlertSent("c-seed-1", "cfg-30", "2025-05-31");
    const upcoming = sched.getUpcomingAlerts("c-seed-1", "2025-05-01");
    const entry = upcoming.find((e) => e.configId === "cfg-30")!;
    expect(entry.sent).toBe(true);
  });

  it("sorted by alertOn ascending", () => {
    const sched = makeSched();
    const upcoming = sched.getUpcomingAlerts("c-seed-1", "2025-01-01");
    const dates = upcoming.map((e) => e.alertOn);
    expect(dates).toEqual([...dates].sort());
  });

  it("unknown contract throws", () => {
    const sched = makeSched();
    expect(() => sched.getUpcomingAlerts("no-such", "2025-01-01")).toThrow();
  });
});
