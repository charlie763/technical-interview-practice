/**
 * =============================================================================
 * INTERVIEW PROBLEM 12: Contract Expiration Alert Scheduler
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the alert-scheduling subsystem for a contract lifecycle
 * management (CLM) platform used by legal and operations teams. Contracts have
 * expiration dates, and stakeholders need to be notified days in advance so they
 * can act before expiry.
 *
 * For this problem you are building a ContractAlertScheduler class.
 * Store all state in instance properties initialized in the constructor.
 * Class-level (static) fields will bleed between tests and between
 * ContractAlertScheduler instances — avoid them.
 * You choose the internal data structures; the public interface is what matters.
 *
 * DATA MODEL
 * ----------
 * Contract:
 *   {
 *     contractId:  string,
 *     title:       string,
 *     ownerEmail:  string,
 *     expiresOn:   string,  // ISO-8601 date string, e.g. "2025-03-15"
 *   }
 *
 * AlertConfig:
 *   {
 *     configId:    string,
 *     daysBefore:  number,  // how many days before expiry to trigger the alert
 *     label:       string,  // e.g. "30-day notice", "final warning"
 *   }
 *
 * SentRecord:
 *   {
 *     contractId:  string,
 *     configId:    string,
 *     sentOn:      string,  // ISO-8601 date string when the alert was sent
 *   }
 *
 * Dates are ISO-8601 strings (date-only, no time component), which sort
 * correctly lexicographically. Use `addDaysToIsoDate` (below) for date
 * arithmetic — it avoids UTC/local timezone drift on plain date strings.
 *
 * # Example
 * const scheduler = new ContractAlertScheduler();
 * scheduler.addContract("c-001", "Vendor MSA", "legal@acme.com", "2025-06-30");
 * scheduler.addAlertConfig("cfg-30", 30, "30-day notice");
 * scheduler.addAlertConfig("cfg-7", 7, "final warning");
 *
 * scheduler.getContractsExpiringBetween("2025-06-01", "2025-06-30");
 * // -> [{ contractId: "c-001", title: "Vendor MSA", ... }]
 *
 * scheduler.computeAlertSchedule("c-001");
 * // -> [
 * //      { configId: "cfg-30", label: "30-day notice", alertOn: "2025-05-31" },
 * //      { configId: "cfg-7",  label: "final warning",  alertOn: "2025-06-23" },
 * //    ]
 *
 * scheduler.getDueAlerts("2025-06-01");
 * // -> [{ contractId: "c-001", configId: "cfg-30", alertOn: "2025-05-31", ... }]
 *
 * =============================================================================
 * PART 1 — Contract and alert-config management
 * =============================================================================
 *
 * Implement `addContract`, `addAlertConfig`, and `getContractsExpiringBetween`.
 */

/**
 * PRE-GIVEN — do not modify.
 * Return the ISO-8601 date string `days` days before/after `isoDate`.
 * (Negative `days` moves the date backward.)
 */
export function addDaysToIsoDate(isoDate: string, days: number): string {
  const [y, m, d] = isoDate.split("-").map(Number);
  const dt = new Date(Date.UTC(y, m - 1, d));
  dt.setUTCDate(dt.getUTCDate() + days);
  return dt.toISOString().slice(0, 10);
}

export type Contract = {
  contractId: string;
  title: string;
  ownerEmail: string;
  expiresOn: string;
};

export type AlertConfig = {
  configId: string;
  daysBefore: number;
  label: string;
};

export type SentRecord = {
  contractId: string;
  configId: string;
  sentOn: string;
};

export type ScheduleEntry = {
  configId: string;
  label: string;
  alertOn: string;
};

export type DueAlertEntry = {
  contractId: string;
  configId: string;
  label: string;
  alertOn: string;
  ownerEmail: string;
  expiresOn: string;
};

export type UpcomingAlertEntry = {
  configId: string;
  label: string;
  alertOn: string;
  sent: boolean;
};

/** Schedules and tracks expiration alerts for contracts. */
export class ContractAlertScheduler {
  constructor() {
    throw new Error("Not implemented");
  }

  // ── Part 1 ────────────────────────────────────────────────────────────────

  /**
   * Register a contract.
   * Throws an Error if contractId already exists.
   */
  addContract(contractId: string, title: string, ownerEmail: string, expiresOn: string): Contract {
    throw new Error("Not implemented");
  }

  /**
   * Register a global alert configuration.
   * Throws an Error if configId already exists.
   */
  addAlertConfig(configId: string, daysBefore: number, label: string): AlertConfig {
    throw new Error("Not implemented");
  }

  /**
   * Return all contracts whose expiration date falls within [startDate, endDate],
   * inclusive on both ends. Results are sorted by expiresOn ascending.
   */
  getContractsExpiringBetween(startDate: string, endDate: string): Contract[] {
    throw new Error("Not implemented");
  }

  // ── Part 2 ────────────────────────────────────────────────────────────────

  /**
   * Compute the full alert schedule for a contract by applying every
   * registered alert config.
   *
   * For each AlertConfig, the alert fires on:
   *     addDaysToIsoDate(expiresOn, -daysBefore)
   *
   * Results are sorted by alertOn date ascending.
   *
   * Throws an Error if contractId does not exist.
   */
  computeAlertSchedule(contractId: string): ScheduleEntry[] {
    throw new Error("Not implemented");
  }

  /**
   * Return all alert schedule entries whose alertOn date is on or before
   * asOfDate. Uses computeAlertSchedule internally — call it per contract;
   * do not duplicate its logic here.
   *
   * Results are sorted by alertOn ascending, then contractId ascending.
   */
  getDueAlerts(asOfDate: string): DueAlertEntry[] {
    throw new Error("Not implemented");
  }

  // ── Part 3 ────────────────────────────────────────────────────────────────

  /**
   * Record that an alert was sent for a specific contract/config pair.
   * Throws an Error if contractId or configId does not exist.
   */
  recordAlertSent(contractId: string, configId: string, sentOn: string): SentRecord {
    throw new Error("Not implemented");
  }

  /**
   * Return the alert schedule for a contract, enriched with a "sent" flag
   * indicating whether that alert has already been sent.
   *
   * Uses computeAlertSchedule and recordAlertSent state internally.
   *
   * Exclude alerts whose alertOn is strictly before asOfDate (they are in
   * the past).
   *
   * Throws an Error if contractId does not exist.
   */
  getUpcomingAlerts(contractId: string, asOfDate: string): UpcomingAlertEntry[] {
    throw new Error("Not implemented");
  }
}
