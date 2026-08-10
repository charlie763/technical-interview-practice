/**
 * =============================================================================
 * INTERVIEW PROBLEM 13: Contract Lifecycle State Machine
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the lifecycle management module for a contract platform.
 * Every contract moves through a defined set of states from creation to
 * completion or termination. Only specific transitions are allowed — illegal
 * transitions must be rejected. All changes are audit-logged.
 *
 * For this problem you are building a ContractLifecycle class.
 * Store all state in instance properties initialized in the constructor.
 * Class-level (static) fields will bleed between tests and between
 * ContractLifecycle instances — avoid them.
 * You choose the internal data structures; the public interface is what matters.
 *
 * DATA MODEL
 * ----------
 * Contract:
 *   {
 *     contractId:  string,
 *     title:       string,
 *     state:       string,               // current lifecycle state (see STATE MACHINE below)
 *     createdAt:   string,                // ISO-8601 datetime when the contract was created
 *     fields:      Record<string, unknown>,  // arbitrary key/value metadata
 *   }
 *
 * AuditEntry:
 *   {
 *     contractId:  string,
 *     fromState:   string | undefined,   // undefined for the initial "created" entry
 *     toState:     string,
 *     at:          string,               // ISO-8601 datetime of the transition
 *     actor:       string,               // user or system identifier
 *   }
 *
 * STATE MACHINE
 * -------------
 * Valid states and allowed forward transitions:
 *
 *   draft  →  in_review
 *   in_review  →  approved | draft     (can be sent back for revisions)
 *   approved  →  executed
 *   executed  →  active
 *   active  →  expiring_soon | terminated
 *   expiring_soon  →  expired | active | terminated
 *   expired  →  (terminal — no outgoing transitions)
 *   terminated  →  (terminal — no outgoing transitions)
 *
 * Timestamps are ISO-8601 strings without timezone offset, e.g.
 * "2025-01-01T09:00:00". Use `new Date(ts).getTime()` for arithmetic — plain
 * string comparison also sorts them correctly.
 *
 * # Example
 * const cl = new ContractLifecycle();
 * cl.createContract("c-001", "Vendor MSA", "2025-01-01T09:00:00", "alice");
 * cl.getContract("c-001").state;  // -> "draft"
 * cl.transition("c-001", "in_review", "2025-01-02T10:00:00", "alice");
 * cl.transition("c-001", "approved",  "2025-01-03T11:00:00", "bob");
 * cl.getContract("c-001").state;  // -> "approved"
 * cl.transition("c-001", "draft", "2025-01-04T09:00:00", "bob");
 * // throws — approved → draft is not a valid transition
 *
 * =============================================================================
 * PART 1 — Contract creation, field management, and transitions
 * =============================================================================
 *
 * Implement `createContract`, `setField`, `getContract`, and `transition`.
 */

// Valid forward transitions from each state.
export const VALID_TRANSITIONS: Record<string, string[]> = {
  draft: ["in_review"],
  in_review: ["approved", "draft"],
  approved: ["executed"],
  executed: ["active"],
  active: ["expiring_soon", "terminated"],
  expiring_soon: ["expired", "active", "terminated"],
  expired: [],
  terminated: [],
};

export type Contract = {
  contractId: string;
  title: string;
  state: string;
  createdAt: string;
  fields: Record<string, unknown>;
};

export type AuditEntry = {
  contractId: string;
  fromState: string | undefined;
  toState: string;
  at: string;
  actor: string;
};

export type BulkAdvanceResult = {
  succeeded: string[];
  failed: { contractId: string; reason: string }[];
};

export type LifecycleMetrics = {
  total: number;
  byState: Record<string, number>;
  terminalCount: number;
};

export type OverdueContract = {
  contractId: string;
  title: string;
  state: string;
  stuckSince: string;
  daysStuck: number;
};

/** Manages lifecycle state transitions and audit history for contracts. */
export class ContractLifecycle {
  constructor() {
    throw new Error("Not implemented");
  }

  // ── Part 1 ────────────────────────────────────────────────────────────────

  /**
   * Create a new contract in the "draft" state and record the initial
   * audit entry (fromState=undefined, toState="draft").
   *
   * Throws an Error if contractId already exists.
   */
  createContract(contractId: string, title: string, createdAt: string, actor: string): Contract {
    throw new Error("Not implemented");
  }

  /**
   * Set or update a field on the contract's `fields` object.
   * Throws an Error if contractId does not exist.
   */
  setField(contractId: string, key: string, value: unknown): Contract {
    throw new Error("Not implemented");
  }

  /** Return the contract. Throws an Error if contractId does not exist. */
  getContract(contractId: string): Contract {
    throw new Error("Not implemented");
  }

  /**
   * Move a contract to a new state if the transition is valid, and append
   * an AuditEntry.
   *
   * Throws an Error if contractId does not exist.
   * Throws an Error if the transition from the current state to toState is
   * not allowed (consult VALID_TRANSITIONS).
   */
  transition(contractId: string, toState: string, at: string, actor: string): Contract {
    throw new Error("Not implemented");
  }

  // ── Part 2 ────────────────────────────────────────────────────────────────

  /**
   * Return the full ordered audit trail for a contract (oldest first).
   * Throws an Error if contractId does not exist.
   */
  getAuditTrail(contractId: string): AuditEntry[] {
    throw new Error("Not implemented");
  }

  /**
   * Return all contracts currently in the given state, sorted by
   * contractId ascending.
   */
  getContractsByState(state: string): Contract[] {
    throw new Error("Not implemented");
  }

  /**
   * Attempt to transition each contract in contractIds to toState.
   * Calls `transition` internally — do not duplicate its logic.
   *
   * Continue processing remaining contracts even if one fails; collect all
   * failures.
   */
  bulkAdvance(contractIds: string[], toState: string, at: string, actor: string): BulkAdvanceResult {
    throw new Error("Not implemented");
  }

  // ── Part 3 ────────────────────────────────────────────────────────────────

  /**
   * Return aggregate counts across all contracts:
   *   {
   *     total:         number,
   *     byState:       { [state]: count },  // only states with count > 0
   *     terminalCount: number,   // contracts in "expired" or "terminated"
   *   }
   */
  getLifecycleMetrics(): LifecycleMetrics {
    throw new Error("Not implemented");
  }

  /**
   * Return contracts that have been stuck in the same non-terminal state
   * for more than 30 days without any transition.
   *
   * "Stuck since" is the `at` timestamp of the most recent AuditEntry for
   * the contract.
   *
   * Uses getAuditTrail internally — do not duplicate its logic.
   *
   * Each item:
   *   { contractId, title, state, stuckSince, daysStuck }
   * Sorted by daysStuck descending.
   */
  getOverdueContracts(asOf: string): OverdueContract[] {
    throw new Error("Not implemented");
  }
}
