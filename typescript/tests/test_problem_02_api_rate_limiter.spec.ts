/**
 * Tests for Problem 2: Tiered API Rate Limiter
 *
 * Run from the typescript/ directory:
 *   npm run test:02
 */

import { describe, expect, it, beforeEach } from "vitest";
import {
  makeGateway,
  createKey,
  revokeKey,
  updatePlan,
  _countInWindow,
  isAllowed,
  recordRequest,
  handleRequest,
  getUsage,
  type GatewayState,
} from "@problems/problem_02_api_rate_limiter";

const BASE_TIME = 1_700_000_000.0; // arbitrary fixed "now" for deterministic tests

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** Gateway with a known set of plans. */
function makeGw(): GatewayState {
  const plans = {
    free: { rpm: 3, rpd: 10 },
    pro: { rpm: 100, rpd: 5_000 },
    unlimited: { rpm: undefined, rpd: undefined },
  };
  return makeGateway(plans);
}

/** Gateway with a known set of plans and one pre-created key. */
function makeGwWithKey(): GatewayState {
  const gw = makeGw();
  createKey(gw, "key_abc", "alice", "pro");
  return gw;
}

// ---------------------------------------------------------------------------
// PART 1 — Key management
// ---------------------------------------------------------------------------

describe("createKey", () => {
  let gw: GatewayState;

  beforeEach(() => {
    gw = makeGw();
  });

  it("creates a key", () => {
    const k = createKey(gw, "k1", "alice", "free");
    expect(gw.keys["k1"]).toBe(k);
    expect(k.owner).toBe("alice");
    expect(k.plan).toBe("free");
    expect(k.enabled).toBe(true);
    expect(k.requestLog).toEqual([]);
  });

  it("duplicate key throws", () => {
    createKey(gw, "k1", "alice", "free");
    expect(() => createKey(gw, "k1", "bob", "pro")).toThrow();
  });

  it("invalid plan throws", () => {
    expect(() => createKey(gw, "k1", "alice", "enterprise")).toThrow();
  });
});

describe("revokeKey", () => {
  let gw: GatewayState;

  beforeEach(() => {
    gw = makeGwWithKey();
  });

  it("disables key", () => {
    revokeKey(gw, "key_abc");
    expect(gw.keys["key_abc"].enabled).toBe(false);
  });

  it("missing key throws", () => {
    expect(() => revokeKey(makeGw(), "ghost")).toThrow();
  });
});

describe("updatePlan", () => {
  let gw: GatewayState;

  beforeEach(() => {
    gw = makeGwWithKey();
  });

  it("changes plan", () => {
    updatePlan(gw, "key_abc", "free");
    expect(gw.keys["key_abc"].plan).toBe("free");
  });

  it("preserves request log", () => {
    gw.keys["key_abc"].requestLog = [BASE_TIME - 5];
    updatePlan(gw, "key_abc", "free");
    expect(gw.keys["key_abc"].requestLog).toEqual([BASE_TIME - 5]);
  });

  it("invalid plan throws", () => {
    expect(() => updatePlan(gw, "key_abc", "nonexistent")).toThrow();
  });

  it("missing key throws", () => {
    expect(() => updatePlan(makeGw(), "ghost", "free")).toThrow();
  });
});

// ---------------------------------------------------------------------------
// PART 2 — _countInWindow
// ---------------------------------------------------------------------------

describe("_countInWindow", () => {
  it("empty log", () => {
    expect(_countInWindow([], BASE_TIME, 60)).toBe(0);
  });

  it("all within window", () => {
    const log = [BASE_TIME - 30, BASE_TIME - 10, BASE_TIME];
    expect(_countInWindow(log, BASE_TIME, 60)).toBe(3);
  });

  it("some outside window", () => {
    const log = [BASE_TIME - 120, BASE_TIME - 61, BASE_TIME - 30, BASE_TIME];
    expect(_countInWindow(log, BASE_TIME, 60)).toBe(2);
  });

  it("exactly on window edge excluded", () => {
    // window is (now - windowSeconds, now] — left side is exclusive
    const log = [BASE_TIME - 60];
    expect(_countInWindow(log, BASE_TIME, 60)).toBe(0);
  });

  it("one second inside", () => {
    const log = [BASE_TIME - 59.999];
    expect(_countInWindow(log, BASE_TIME, 60)).toBe(1);
  });
});

// ---------------------------------------------------------------------------
// PART 2 — isAllowed
// ---------------------------------------------------------------------------

describe("isAllowed", () => {
  let gw: GatewayState;

  beforeEach(() => {
    gw = makeGwWithKey();
  });

  it("allowed when under limits", () => {
    expect(isAllowed(gw, "key_abc", BASE_TIME)).toBe(true);
  });

  it("denied when key not found", () => {
    expect(isAllowed(makeGw(), "ghost", BASE_TIME)).toBe(false);
  });

  it("denied when key disabled", () => {
    revokeKey(gw, "key_abc");
    expect(isAllowed(gw, "key_abc", BASE_TIME)).toBe(false);
  });

  it("denied when rpm exceeded", () => {
    const g = makeGw();
    createKey(g, "k1", "alice", "free"); // rpm=3
    // Seed 3 requests in the last minute
    g.keys["k1"].requestLog = [BASE_TIME - 30, BASE_TIME - 20, BASE_TIME - 10];
    expect(isAllowed(g, "k1", BASE_TIME)).toBe(false);
  });

  it("allowed when rpm window has rolled off", () => {
    const g = makeGw();
    createKey(g, "k1", "alice", "free"); // rpm=3
    // 3 requests but all > 60s ago — they're outside the window
    g.keys["k1"].requestLog = [BASE_TIME - 90, BASE_TIME - 80, BASE_TIME - 70];
    expect(isAllowed(g, "k1", BASE_TIME)).toBe(true);
  });

  it("denied when rpd exceeded", () => {
    const g = makeGw();
    createKey(g, "k1", "alice", "free"); // rpd=10
    g.keys["k1"].requestLog = Array.from({ length: 10 }, (_, i) => BASE_TIME - i * 100);
    expect(isAllowed(g, "k1", BASE_TIME)).toBe(false);
  });

  it("unlimited plan always allowed", () => {
    const g = makeGw();
    createKey(g, "k1", "alice", "unlimited");
    // Even with a huge log, unlimited plan is never blocked
    g.keys["k1"].requestLog = Array.from({ length: 1000 }, (_, i) => BASE_TIME - i);
    expect(isAllowed(g, "k1", BASE_TIME)).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// PART 3 — recordRequest
// ---------------------------------------------------------------------------

describe("recordRequest", () => {
  let gw: GatewayState;

  beforeEach(() => {
    gw = makeGwWithKey();
  });

  it("appends timestamp", () => {
    recordRequest(gw, "key_abc", BASE_TIME);
    expect(gw.keys["key_abc"].requestLog).toContain(BASE_TIME);
  });

  it("prunes old entries", () => {
    const old = BASE_TIME - 90_001; // older than 25h
    gw.keys["key_abc"].requestLog = [old];
    recordRequest(gw, "key_abc", BASE_TIME);
    expect(gw.keys["key_abc"].requestLog).not.toContain(old);
  });

  it("keeps recent entries", () => {
    const recent = BASE_TIME - 3600;
    gw.keys["key_abc"].requestLog = [recent];
    recordRequest(gw, "key_abc", BASE_TIME);
    expect(gw.keys["key_abc"].requestLog).toContain(recent);
  });

  it("missing key throws", () => {
    expect(() => recordRequest(makeGw(), "ghost", BASE_TIME)).toThrow();
  });
});

// ---------------------------------------------------------------------------
// PART 4 — handleRequest
// ---------------------------------------------------------------------------

describe("handleRequest", () => {
  let gw: GatewayState;

  beforeEach(() => {
    gw = makeGwWithKey();
  });

  it("success records request", () => {
    const result = handleRequest(gw, "key_abc", BASE_TIME);
    expect(result.allowed).toBe(true);
    expect(gw.keys["key_abc"].requestLog).toContain(BASE_TIME);
  });

  it("key not found", () => {
    const result = handleRequest(makeGw(), "ghost", BASE_TIME);
    expect(result.allowed).toBe(false);
    expect((result as { reason: string }).reason).toBe("key_not_found");
  });

  it("key disabled", () => {
    revokeKey(gw, "key_abc");
    const result = handleRequest(gw, "key_abc", BASE_TIME);
    expect(result.allowed).toBe(false);
    expect((result as { reason: string }).reason).toBe("key_disabled");
  });

  it("rpm exceeded", () => {
    const g = makeGw();
    createKey(g, "k1", "alice", "free"); // rpm=3
    g.keys["k1"].requestLog = [BASE_TIME - 10, BASE_TIME - 5, BASE_TIME - 1];
    const result = handleRequest(g, "k1", BASE_TIME);
    expect(result.allowed).toBe(false);
    expect((result as { reason: string }).reason).toBe("rpm_exceeded");
  });

  it("rpd exceeded not recorded", () => {
    const g = makeGw();
    createKey(g, "k1", "alice", "free"); // rpd=10
    g.keys["k1"].requestLog = Array.from({ length: 10 }, (_, i) => BASE_TIME - i * 100);
    const logBefore = [...g.keys["k1"].requestLog];
    const result = handleRequest(g, "k1", BASE_TIME);
    expect(result.allowed).toBe(false);
    expect((result as { reason: string }).reason).toBe("rpd_exceeded");
    // log must not be modified on failure
    expect(g.keys["k1"].requestLog).toEqual(logBefore);
  });

  it("rpd checked after rpm (rpd_exceeded should only appear when rpm is within limits)", () => {
    const g = makeGw();
    createKey(g, "k1", "alice", "free"); // rpm=3, rpd=10
    // rpm is OK (0 in last minute), but rpd is blown
    g.keys["k1"].requestLog = Array.from({ length: 10 }, (_, i) => BASE_TIME - 3600 * (i + 1));
    const result = handleRequest(g, "k1", BASE_TIME);
    expect((result as { reason: string }).reason).toBe("rpd_exceeded");
  });
});

// ---------------------------------------------------------------------------
// PART 4 — getUsage
// ---------------------------------------------------------------------------

describe("getUsage", () => {
  let gw: GatewayState;

  beforeEach(() => {
    gw = makeGwWithKey();
  });

  it("returns correct counts", () => {
    gw.keys["key_abc"].requestLog = [
      BASE_TIME - 30, // within both minute and day windows
      BASE_TIME - 3600, // within day window only
    ];
    const stats = getUsage(gw, "key_abc", BASE_TIME);
    expect(stats.rpmUsed).toBe(1);
    expect(stats.rpdUsed).toBe(2);
    expect(stats.plan).toBe("pro");
    expect(stats.rpmLimit).toBe(100);
    expect(stats.rpdLimit).toBe(5_000);
  });

  it("unlimited plan shows undefined limits", () => {
    const g = makeGw();
    createKey(g, "k1", "alice", "unlimited");
    const stats = getUsage(g, "k1", BASE_TIME);
    expect(stats.rpmLimit).toBeUndefined();
    expect(stats.rpdLimit).toBeUndefined();
  });

  it("missing key throws", () => {
    expect(() => getUsage(makeGw(), "ghost", BASE_TIME)).toThrow();
  });
});
