/**
 * =============================================================================
 * INTERVIEW PROBLEM 2: Tiered API Rate Limiter
 * Difficulty: Senior Software Engineer | Estimated time: 40 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the rate-limiting layer for a developer-facing API platform
 * (think: Stripe, Vercel, Twilio — any product where external developers call
 * your API and pay for usage tiers).
 *
 * Each API key belongs to a plan with per-minute and per-day request caps.
 * Rate limiting uses a sliding window: a request at time T is within the
 * per-minute cap if fewer than `rpm` requests occurred in the window (T-60, T],
 * and within the per-day cap if fewer than `rpd` requests occurred in (T-86400, T].
 *
 * All timestamps are Unix time (seconds, as a number — fractional seconds allowed).
 *
 * DATA MODEL
 * ----------
 *
 * Plan limits are stored in state.plans:
 *   {
 *     free:       { rpm: 60,    rpd: 1_000 },
 *     starter:    { rpm: 300,   rpd: 25_000 },
 *     pro:        { rpm: 1_000, rpd: 200_000 },
 *     enterprise: { rpm: undefined, rpd: undefined },   // undefined = unlimited
 *   }
 *
 * Each API key entry in state.keys looks like:
 *   {
 *     id:          string,
 *     owner:       string,
 *     plan:        string,          // must be a key in state.plans
 *     enabled:     boolean,
 *     requestLog:  number[],        // sorted list of timestamps of recent requests
 *                                    // entries older than 25h may be pruned at any time
 *   }
 *
 * GatewayState:
 *   {
 *     plans: { [planName: string]: { rpm: number | undefined, rpd: number | undefined } },
 *     keys:  { [keyId: string]: ApiKey },
 *   }
 *
 * # Example
 * const gw = makeGateway();
 * createKey(gw, "key_abc", "alice", "free");
 * handleRequest(gw, "key_abc", 1_700_000_000);
 * // -> { allowed: true, keyId: "key_abc" }
 * getUsage(gw, "key_abc", 1_700_000_000);
 * // -> { keyId: "key_abc", plan: "free", rpmUsed: 1, rpmLimit: 60, rpdUsed: 1, rpdLimit: 1000 }
 * =============================================================================
 */

export type PlanLimits = { rpm: number | undefined; rpd: number | undefined };

export type ApiKey = {
  id: string;
  owner: string;
  plan: string;
  enabled: boolean;
  requestLog: number[];
};

export type GatewayState = {
  plans: Record<string, PlanLimits>;
  keys: Record<string, ApiKey>;
};

export type RequestResult =
  { allowed: true; keyId: string } | { allowed: false; reason: "key_not_found" | "key_disabled" | "rpm_exceeded" | "rpd_exceeded" };

export type UsageStats = {
  keyId: string;
  plan: string;
  rpmUsed: number;
  rpmLimit: number | undefined;
  rpdUsed: number;
  rpdLimit: number | undefined;
};

export const DEFAULT_PLANS: Record<string, PlanLimits> = {
  free: { rpm: 60, rpd: 1_000 },
  starter: { rpm: 300, rpd: 25_000 },
  pro: { rpm: 1_000, rpd: 200_000 },
  enterprise: { rpm: undefined, rpd: undefined },
};

/** Return a fresh gateway state. Uses DEFAULT_PLANS if none provided. */
export function makeGateway(plans?: Record<string, PlanLimits>): GatewayState {
  return {
    plans: plans !== undefined ? plans : DEFAULT_PLANS,
    keys: {},
  };
}

// ---------------------------------------------------------------------------
// PART 1 — Key management  (~10 min)
// ---------------------------------------------------------------------------

/**
 * Register a new API key. Return the key.
 * Throw an Error if keyId already exists or plan is not in state.plans.
 */
export function createKey(state: GatewayState, keyId: string, owner: string, plan: string): ApiKey {
  throw new Error("Not implemented");
}

/**
 * Set key.enabled = false. Throw an Error if keyId not found.
 * (Keys are disabled rather than deleted so historical logs are preserved.)
 */
export function revokeKey(state: GatewayState, keyId: string): void {
  throw new Error("Not implemented");
}

/**
 * Change a key's plan. Throw an Error if keyId not found.
 * Throw an Error if newPlan is not in state.plans.
 * The requestLog is preserved (no reset on plan change).
 */
export function updatePlan(state: GatewayState, keyId: string, newPlan: string): void {
  throw new Error("Not implemented");
}

// ---------------------------------------------------------------------------
// PART 2 — Sliding-window rate check  (~15 min)
// ---------------------------------------------------------------------------

/**
 * Return the number of entries in requestLog that fall within
 * the half-open window (now - windowSeconds, now].
 *
 * Helper — feel free to use this in Parts 2 and 3, or inline it.
 */
export function _countInWindow(requestLog: number[], now: number, windowSeconds: number): number {
  throw new Error("Not implemented");
}

/**
 * Return true if the key is allowed to make a request at time `now`.
 *
 * A request is NOT allowed if any of the following:
 *   - keyId doesn't exist in state.keys
 *   - key.enabled is false
 *   - the per-minute sliding window is at or above the plan's rpm cap
 *   - the per-day sliding window is at or above the plan's rpd cap
 *
 * An undefined limit means that dimension is unlimited.
 */
export function isAllowed(state: GatewayState, keyId: string, now: number): boolean {
  throw new Error("Not implemented");
}

// ---------------------------------------------------------------------------
// PART 3 — Recording requests + pruning  (~5 min)
// ---------------------------------------------------------------------------

/**
 * Append `now` to key.requestLog and prune any entries older than
 * 25 hours (90_000 seconds) to bound memory usage.
 *
 * Throw an Error if keyId not found.
 * Note: call this only AFTER confirming the request is allowed.
 */
export function recordRequest(state: GatewayState, keyId: string, now: number): void {
  throw new Error("Not implemented");
}

// ---------------------------------------------------------------------------
// PART 4 — Combined handler + usage stats  (~10 min)
// ---------------------------------------------------------------------------

/**
 * Attempt to process a request. Return a RequestResult.
 *
 * On success (request allowed):
 *   { allowed: true, keyId }
 *   Side effect: records the request.
 *
 * On failure:
 *   { allowed: false, reason: <one of the strings below> }
 *   No side effects.
 *
 * Reason strings:
 *   "key_not_found" — keyId not in state.keys
 *   "key_disabled"  — key exists but enabled=false
 *   "rpm_exceeded"  — per-minute cap hit
 *   "rpd_exceeded"  — per-day cap hit (only if rpm is OK)
 */
export function handleRequest(state: GatewayState, keyId: string, now: number): RequestResult {
  throw new Error("Not implemented");
}

/**
 * Return current usage stats for a key.
 * Throw an Error if keyId not found.
 */
export function getUsage(state: GatewayState, keyId: string, now: number): UsageStats {
  throw new Error("Not implemented");
}
