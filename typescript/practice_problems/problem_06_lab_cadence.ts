/**
 * =============================================================================
 * INTERVIEW PROBLEM 6: Lab Cadence Compliance Monitor
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * Company context: Health Tech
 * =============================================================================
 *
 * CONTEXT
 * -------
 * A health tech company requires patients to submit lab work at regular
 * intervals so clinicians can track metabolic health markers (HbA1c, fasting
 * glucose, lipids, kidney function, etc.). Patients who miss lab deadlines
 * need follow-up from their health coach.
 *
 * You are building the LabCadenceMonitor — a function-based module (no class)
 * that manages a registry of patients, their required lab types, submission
 * deadlines, and actual submissions.
 *
 * PRE-GIVEN (do not modify these)
 * --------------------------------
 * makeMonitor() creates and returns the data store you will work with.
 * All functions receive the monitor as their first argument.
 *
 * # Example
 * const m = makeMonitor();
 * registerPatient(m, "alice", ["hba1c", "bmp"]);
 * setLabDeadline(m, "alice", "hba1c", "2024-03-31");
 * recordSubmission(m, "alice", "hba1c", "2024-03-28");
 * isOverdue(m, "alice", "hba1c", "2024-04-01");  // -> false (submitted on time)
 * isOverdue(m, "alice", "bmp",   "2024-04-01");  // -> true  (no deadline set yet,
 *                                                //           but bmp is required)
 *
 * NOTES
 * -----
 *   - "overdue" means: a required lab has a deadline that has passed (asOf > dueDate)
 *     AND no submission exists on or before the due date.
 *   - If a required lab has no deadline set, it is NOT considered overdue.
 *   - A submission clears the specific deadline it satisfies (the earliest
 *     uncleared deadline on or after the submission date).
 *   - Patients can have multiple deadlines per lab type (e.g. quarterly HbA1c).
 *   - All dates are ISO calendar date strings ("YYYY-MM-DD"), which sort
 *     correctly lexicographically.
 *   - All state lives inside the object returned by makeMonitor(). No module-level state.
 * =============================================================================
 */

// ---------------------------------------------------------------------------
// PRE-GIVEN — do not modify
// ---------------------------------------------------------------------------

export type PatientRecord = {
  requiredLabs: Set<string>;
  deadlines: Record<string, string[]>; // lab_type -> sorted ascending ISO dates
  submissions: Record<string, string[]>; // lab_type -> sorted ascending ISO dates
};

export type LabCadenceMonitor = {
  patients: Record<string, PatientRecord>;
};

export type ComplianceEntry = {
  patientId: string;
  overdueLabs: string[];
  overdueCount: number;
};

/**
 * Return a fresh monitor data store.
 *
 * Schema (you may add keys as needed):
 *   {
 *     patients: {
 *       [patientId]: {
 *         requiredLabs: Set<string>,
 *         deadlines:    { [labType]: string[] },   // sorted ascending
 *         submissions:  { [labType]: string[] },   // sorted ascending
 *       }
 *     }
 *   }
 */
export function makeMonitor(): LabCadenceMonitor {
  return { patients: {} };
}

// ---------------------------------------------------------------------------
// YOUR IMPLEMENTATION
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// PART 1 — Patient & lab registration  (~10 min)
// ---------------------------------------------------------------------------

/**
 * Register a new patient with a list of required lab types.
 *
 * If the patient already exists, add any new lab types to their required set
 * (do not remove existing ones). Idempotent for labs already in the set.
 *
 * Throw an Error if requiredLabs is empty.
 */
export function registerPatient(monitor: LabCadenceMonitor, patientId: string, requiredLabs: string[]): void {
  throw new Error("Not implemented");
}

/**
 * Add a single required lab type to an existing patient's requirements.
 *
 * Throw an Error if the patient doesn't exist.
 * No-op if the lab is already required.
 */
export function addRequiredLab(monitor: LabCadenceMonitor, patientId: string, labType: string): void {
  throw new Error("Not implemented");
}

/**
 * Return the set of required lab types for the patient.
 * Throw an Error if the patient doesn't exist.
 */
export function getRequiredLabs(monitor: LabCadenceMonitor, patientId: string): Set<string> {
  throw new Error("Not implemented");
}

// ---------------------------------------------------------------------------
// PART 2 — Deadlines and submissions  (~15 min)
// ---------------------------------------------------------------------------

/**
 * Add a deadline for a specific lab type for the patient.
 *
 * A patient may have multiple deadlines for the same lab (e.g. quarterly).
 * Duplicate deadlines (same patient + lab + date) are ignored.
 *
 * Throw an Error if the patient doesn't exist.
 * Throw an Error if labType is not in the patient's requiredLabs.
 */
export function setLabDeadline(monitor: LabCadenceMonitor, patientId: string, labType: string, dueDate: string): void {
  throw new Error("Not implemented");
}

/**
 * Record that the patient submitted a lab result on submittedOn.
 *
 * Clears the earliest uncleared deadline for this lab type that is
 * >= submittedOn. If no such deadline exists, the submission is still
 * recorded (it may satisfy a future deadline or serve as history).
 *
 * Throw an Error if the patient doesn't exist.
 * Throw an Error if labType is not in the patient's requiredLabs.
 */
export function recordSubmission(monitor: LabCadenceMonitor, patientId: string, labType: string, submittedOn: string): void {
  throw new Error("Not implemented");
}

/**
 * Return true if the patient has at least one uncleared deadline for
 * labType that has passed as of `asOf` (i.e., dueDate < asOf).
 *
 * Return false if:
 *   - The patient doesn't exist.
 *   - labType is not required for the patient.
 *   - No deadline has been set for that lab.
 *   - All past deadlines have been cleared by a submission.
 */
export function isOverdue(monitor: LabCadenceMonitor, patientId: string, labType: string, asOf: string): boolean {
  throw new Error("Not implemented");
}

// ---------------------------------------------------------------------------
// PART 3 — Compliance reporting  (~15 min)
// ---------------------------------------------------------------------------

/**
 * Return a sorted list of lab type names that are currently overdue for
 * the patient as of `asOf`.
 *
 * Return an empty array if the patient doesn't exist or has no overdue labs.
 */
export function overdueLabs(monitor: LabCadenceMonitor, patientId: string, asOf: string): string[] {
  throw new Error("Not implemented");
}

/**
 * Return a report of all patients with at least one overdue lab as of `asOf`.
 *
 * Each entry in the array is:
 *   { patientId: string, overdueLabs: string[], overdueCount: number }
 *
 * Sort the array by overdueCount descending (most overdue first), then
 * alphabetically by patientId for ties.
 *
 * Return an empty array if no patient has overdue labs.
 */
export function complianceReport(monitor: LabCadenceMonitor, asOf: string): ComplianceEntry[] {
  throw new Error("Not implemented");
}

// ---------------------------------------------------------------------------
// PART 4 — Submission history  (~5 min)
// ---------------------------------------------------------------------------

/**
 * Return a chronologically sorted list of all submission dates for
 * (patientId, labType).
 *
 * Return an empty array if the patient doesn't exist or has no submissions
 * for that lab type.
 */
export function submissionHistory(monitor: LabCadenceMonitor, patientId: string, labType: string): string[] {
  throw new Error("Not implemented");
}

/**
 * Return the number of days between the patient's most recent submission
 * for labType and `asOf`.
 *
 * Return undefined if the patient has never submitted that lab type.
 */
export function daysSinceLastSubmission(monitor: LabCadenceMonitor, patientId: string, labType: string, asOf: string): number | undefined {
  throw new Error("Not implemented");
}
