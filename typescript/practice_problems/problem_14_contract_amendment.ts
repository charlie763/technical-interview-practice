/**
 * =============================================================================
 * INTERVIEW PROBLEM 14: Contract Amendment Manager
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the amendment-tracking module for a contract platform.
 * After a contract is signed, its terms can be modified through formal amendments.
 * Each amendment specifies a set of field overrides that take effect from a given
 * date. To know the effective terms on any given date, you apply the base contract
 * fields and then overlay amendments in chronological order up to that date.
 *
 * For this problem you are building a ContractAmendmentManager class.
 * Store all state in instance properties initialized in the constructor.
 * Class-level (static) fields will bleed between tests and between
 * ContractAmendmentManager instances — avoid them.
 * You choose the internal data structures; the public interface is what matters.
 *
 * DATA MODEL
 * ----------
 * Contract (base):
 *   {
 *     contractId:  string,
 *     title:       string,
 *     fields:      Record<string, unknown>,  // e.g. { value: 50000, paymentTerms: "net-30" }
 *   }
 *
 * Amendment:
 *   {
 *     amendmentId:  string,
 *     contractId:   string,
 *     effectiveOn:  string,  // ISO-8601 date string ("YYYY-MM-DD")
 *     overrides:    Record<string, unknown>,  // field key → new value (may be a subset of fields)
 *     note:         string,  // human-readable reason for the amendment
 *   }
 *
 * Dates are ISO-8601 date strings (date-only), which sort correctly
 * lexicographically.
 *
 * # Example
 * const mgr = new ContractAmendmentManager();
 * mgr.addContract("c-001", "Vendor MSA", { value: 50000, paymentTerms: "net-30" });
 *
 * mgr.addAmendment("amd-1", "c-001", "2025-03-01",
 *                   { paymentTerms: "net-45" }, "extended terms");
 * mgr.addAmendment("amd-2", "c-001", "2025-06-01",
 *                   { value: 75000 }, "scope increase");
 *
 * mgr.getEffectiveContract("c-001", "2025-01-01");
 * // -> { value: 50000, paymentTerms: "net-30" }   (no amendments yet)
 *
 * mgr.getEffectiveContract("c-001", "2025-04-15");
 * // -> { value: 50000, paymentTerms: "net-45" }   (amd-1 applied)
 *
 * mgr.getEffectiveContract("c-001", "2025-07-01");
 * // -> { value: 75000, paymentTerms: "net-45" }   (amd-1 + amd-2 applied)
 *
 * =============================================================================
 * PART 1 — Base contract management
 * =============================================================================
 *
 * Implement `addContract` and `getBaseContract`.
 */

export type Contract = {
  contractId: string;
  title: string;
  fields: Record<string, unknown>;
};

export type Amendment = {
  amendmentId: string;
  contractId: string;
  effectiveOn: string;
  overrides: Record<string, unknown>;
  note: string;
};

export type ValueHistoryEntry = {
  effectiveOn: string; // ISO-8601 date; "base" for the original value
  value: unknown;
  source: string; // "base" | amendmentId
};

export type AmendmentSummary = {
  contractId: string;
  amendmentCount: number;
  fieldsAmended: string[];
  latestAmendment: string | undefined;
  currentFields: Record<string, unknown>;
};

/**
 * Tracks contract base terms and their amendments, and resolves the effective
 * contract state as of any given date.
 */
export class ContractAmendmentManager {
  constructor() {
    throw new Error("Not implemented");
  }

  // ── Part 1 ────────────────────────────────────────────────────────────────

  /**
   * Register a base contract.
   * Store a copy of `fields` — do not hold a reference to the caller's object.
   *
   * Throws an Error if contractId already exists.
   */
  addContract(contractId: string, title: string, fields: Record<string, unknown>): Contract {
    throw new Error("Not implemented");
  }

  /**
   * Return the base contract (original fields, no amendments applied).
   * Throws an Error if contractId does not exist.
   */
  getBaseContract(contractId: string): Contract {
    throw new Error("Not implemented");
  }

  // ── Part 2 ────────────────────────────────────────────────────────────────

  /**
   * Register an amendment for a contract.
   * Store a copy of `overrides` — do not hold a reference to the caller's object.
   *
   * Throws an Error if amendmentId already exists.
   * Throws an Error if contractId does not exist.
   */
  addAmendment(amendmentId: string, contractId: string, effectiveOn: string, overrides: Record<string, unknown>, note: string): Amendment {
    throw new Error("Not implemented");
  }

  /**
   * Return all amendments for a contract, sorted by effectiveOn ascending,
   * then amendmentId ascending (for deterministic ordering when dates match).
   *
   * Throws an Error if contractId does not exist.
   */
  getAmendments(contractId: string): Amendment[] {
    throw new Error("Not implemented");
  }

  /**
   * Return the resolved field values for a contract as of asOfDate.
   *
   * Start with the base fields from `getBaseContract`, then apply
   * amendments in chronological order (earliest first) where
   * effectiveOn <= asOfDate, overlaying their overrides.
   *
   * Uses getBaseContract and getAmendments internally — do not
   * duplicate their logic.
   *
   * Throws an Error if contractId does not exist.
   */
  getEffectiveContract(contractId: string, asOfDate: string): Record<string, unknown> {
    throw new Error("Not implemented");
  }

  // ── Part 3 ────────────────────────────────────────────────────────────────

  /**
   * Return the full history of a specific field's value across base and
   * all amendments that touched it, in chronological order (base always first).
   *
   * Each item: { effectiveOn, value, source } where source is "base" or an
   * amendmentId.
   *
   * Throws an Error if contractId does not exist, or if the field is not
   * present in the base contract or any amendment for that contract.
   */
  getValueHistory(contractId: string, field: string): ValueHistoryEntry[] {
    throw new Error("Not implemented");
  }

  /**
   * Return a summary of amendment activity for a contract.
   * Uses getAmendments and getEffectiveContract internally — do not
   * duplicate their logic.
   *
   * {
   *   contractId,
   *   amendmentCount,
   *   fieldsAmended,     // unique field names ever overridden, sorted ascending
   *   latestAmendment,   // ISO-8601 date of most recent amendment, or undefined
   *                      // if no amendments
   *   currentFields,     // result of getEffectiveContract called with "today"
   * }
   *
   * For "today" use the most recent amendment's effectiveOn date if any
   * amendments exist, otherwise use "2099-12-31" as a far-future sentinel so
   * all amendments are included.
   *
   * Throws an Error if contractId does not exist.
   */
  getAmendmentSummary(contractId: string): AmendmentSummary {
    throw new Error("Not implemented");
  }
}
