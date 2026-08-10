/**
 * =============================================================================
 * Care Team Assignment Manager
 * =============================================================================
 *
 * A remote clinical platform supports patients through a care team. Each care
 * team member has a specific role ("coach", "physician", "dietitian", etc.) and a
 * maximum number of patients they can hold at one time. A patient may have at most
 * one assigned member per role at a time. When a patient is reassigned to a
 * different member, the full history of past assignments is preserved for audit
 * and care-continuity purposes.
 *
 * You choose the internal data structures — the public interface is what matters.
 *
 * Store all state in instance properties initialized in the constructor.
 * Class-level (static) fields will bleed between tests and between CareTeamManager
 * instances — avoid them.
 *
 * ────────────────────────────────────────────────────────────────────────────
 * Part 1 — Basic assignment and lookup
 *   addMember(memberId, role, maxPatients)
 *   assign(patientId, memberId, assignedAt)
 *   getAssignment(patientId, role)   -> string | undefined
 *   getPatients(memberId)            -> string[]
 *
 * Part 2 — Capacity enforcement
 *   assign() now throws CapacityError when the member is at maxPatients
 *   availableMembers(role)           -> string[]
 *
 * Part 3 — Assignment history
 *   getHistory(patientId, role)                    -> [string, number, number | undefined][]
 *   getAssignmentAt(patientId, role, timestamp)     -> string | undefined
 * ────────────────────────────────────────────────────────────────────────────
 *
 * # Example
 * const mgr = new CareTeamManager();
 * mgr.addMember("coach_a", "coach", 2);
 * mgr.addMember("dr_main", "physician", 100);
 * mgr.assign("patient_1", "coach_a", 1000.0);
 * mgr.assign("patient_1", "dr_main", 1000.0);
 * mgr.getAssignment("patient_1", "coach");      // -> "coach_a"
 * mgr.getAssignment("patient_1", "dietitian");  // -> undefined
 * mgr.getPatients("coach_a");                   // -> ["patient_1"]
 *
 * // Part 2
 * mgr.addMember("coach_b", "coach", 1);
 * mgr.assign("patient_2", "coach_b", 2000.0);
 * mgr.assign("patient_3", "coach_b", 3000.0);  // throws CapacityError
 * mgr.availableMembers("coach");                // -> ["coach_a"]
 *
 * // Part 3 — reassign patient_1 from coach_a to coach_b
 * mgr.assign("patient_1", "coach_b", 5000.0);
 * mgr.getHistory("patient_1", "coach");
 * // -> [["coach_a", 1000.0, 5000.0], ["coach_b", 5000.0, undefined]]
 * mgr.getAssignmentAt("patient_1", "coach",  500.0);   // -> undefined (before any assignment)
 * mgr.getAssignmentAt("patient_1", "coach", 3000.0);   // -> "coach_a"
 * mgr.getAssignmentAt("patient_1", "coach", 6000.0);   // -> "coach_b"
 * =============================================================================
 */

/** Thrown when assigning a patient to a member who is at their patient capacity. */
export class CapacityError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "CapacityError";
  }
}

export type AssignmentHistoryEntry = [memberId: string, assignedAt: number, unassignedAt: number | undefined];

/**
 * Manages patient-to-care-team-member assignments for a remote clinical platform.
 *
 * You choose the internal data structures — the public interface is what matters.
 *
 * Store all state in instance properties initialized in the constructor.
 * Class-level (static) fields will bleed between tests and between CareTeamManager
 * instances — avoid them.
 */
export class CareTeamManager {
  constructor() {
    throw new Error("Not implemented");
  }

  // ── Part 1: Basic assignment and lookup ───────────────────────────────────

  /** Register a care team member with the given role and patient capacity. */
  addMember(memberId: string, role: string, maxPatients: number): void {
    throw new Error("Not implemented");
  }

  /**
   * Assign a patient to a care team member (assignedAt is Unix seconds).
   *
   * A patient may have at most one assigned member per role at a time.
   * If the patient already has a member with the same role, that assignment
   * is replaced — the new assignment takes effect at assignedAt.
   *
   * Throws a plain Error if memberId has not been registered via addMember.
   *
   * Part 2 addition: throws CapacityError if the member is already at
   * maxPatients and the patient is not currently assigned to that exact member.
   * (Reassigning a patient who is already on this member does not count as
   * adding a new patient — it is a no-op for capacity purposes.)
   */
  assign(patientId: string, memberId: string, assignedAt: number): void {
    throw new Error("Not implemented");
  }

  /**
   * Return the memberId currently assigned to this patient for the given
   * role, or undefined if no member of that role is currently assigned.
   */
  getAssignment(patientId: string, role: string): string | undefined {
    throw new Error("Not implemented");
  }

  /**
   * Return a sorted list of patientIds currently assigned to this member.
   * Throws a plain Error if memberId has not been registered.
   */
  getPatients(memberId: string): string[] {
    throw new Error("Not implemented");
  }

  // ── Part 2: Capacity enforcement ──────────────────────────────────────────

  /**
   * Return a sorted list of memberIds with the given role that still have
   * open capacity (current patient count < maxPatients).
   */
  availableMembers(role: string): string[] {
    throw new Error("Not implemented");
  }

  // ── Part 3: Assignment history ────────────────────────────────────────────

  /**
   * Return the full assignment history for the patient's given role as a list
   * of [memberId, assignedAt, unassignedAt] tuples sorted by assignedAt.
   *
   * - unassignedAt is undefined for the current (still-active) assignment.
   * - unassignedAt equals the assignedAt of the subsequent assignment for
   *   past entries.
   * - Returns [] if the patient has never been assigned a member of this role.
   */
  getHistory(patientId: string, role: string): AssignmentHistoryEntry[] {
    throw new Error("Not implemented");
  }

  /**
   * Return the memberId assigned to the patient for the given role at the
   * given timestamp, or undefined if no assignment was active at that time.
   *
   * An assignment is active during the interval [assignedAt, unassignedAt).
   * The current assignment (unassignedAt is undefined) is active from assignedAt
   * onward.
   *
   * Implement this by calling getHistory() — do not duplicate the lookup logic.
   */
  getAssignmentAt(patientId: string, role: string, timestamp: number): string | undefined {
    throw new Error("Not implemented");
  }
}
