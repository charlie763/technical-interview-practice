// =============================================================================
// INTERVIEW PROBLEM 7: Care Team Assignment Manager
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// A remote clinical platform supports patients through a dedicated care team.
// Each care team member has a specific role ("coach", "physician", "dietitian",
// etc.) and a maximum number of patients they can hold at one time. A patient
// may have at most one assigned member per role at a time. When a patient is
// reassigned to a different member, the full history of past assignments is
// preserved for audit and care-continuity purposes.
//
// You choose the internal data structures — the public interface (CareTeamManager)
// is what matters. Store all state in struct fields set in your constructor.
// Package-level variables will bleed state between instances and test runs.
//
// DATA MODEL
// ----------
// Member   — ID, Role, MaxPatients (capacity ceiling)
// Assignment — MemberID, AssignedAt (Unix seconds), UnassignedAt (*float64,
//              nil = still active)
//
// EXAMPLE
// -------
//   mgr := NewCareTeamManager()
//   mgr.AddMember("coach_a", "coach", 2)
//   mgr.AddMember("dr_main", "physician", 100)
//   mgr.Assign("patient_1", "coach_a", 1000.0)
//   mgr.Assign("patient_1", "dr_main", 1000.0)
//   mgr.GetAssignment("patient_1", "coach")     // "coach_a", nil
//   mgr.GetAssignment("patient_1", "dietitian") // "", nil (not assigned)
//   mgr.GetPatients("coach_a")                  // ["patient_1"], nil
//
//   // Part 2
//   mgr.AddMember("coach_b", "coach", 1)
//   mgr.Assign("patient_2", "coach_b", 2000.0)
//   mgr.Assign("patient_3", "coach_b", 3000.0) // returns ErrCapacity
//   mgr.AvailableMembers("coach")              // ["coach_a"], nil
//
//   // Part 3 — reassign patient_1 from coach_a to coach_b
//   mgr.Assign("patient_1", "coach_b", 5000.0)
//   mgr.GetHistory("patient_1", "coach")
//   // -> [{"coach_a", 1000.0, &5000.0}, {"coach_b", 5000.0, nil}]
//   mgr.GetAssignmentAt("patient_1", "coach", 500.0)  // "", nil (before any assignment)
//   mgr.GetAssignmentAt("patient_1", "coach", 3000.0) // "coach_a", nil
//   mgr.GetAssignmentAt("patient_1", "coach", 6000.0) // "coach_b", nil
// =============================================================================

package careteam

import "errors"

// ErrNotFound is returned when accessing a member that has not been registered.
var ErrNotFound = errors.New("not found")

// ErrCapacity is returned when assigning a patient to a member who is already
// at their maximum patient capacity.
var ErrCapacity = errors.New("capacity exceeded")

// AssignmentRecord holds a single entry in a patient's assignment history for
// a given role.
type AssignmentRecord struct {
	MemberID     string
	AssignedAt   float64
	UnassignedAt *float64 // nil if this assignment is still active
}

// CareTeamManager is the interface all implementations must satisfy.
//
// Implement it by defining your own struct and a NewCareTeamManager constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myCTM struct {
//	    members map[string]*memberInfo
//	    // ... add whatever fields you need
//	}
//
//	func NewCareTeamManager() CareTeamManager {
//	    return &myCTM{
//	        members: make(map[string]*memberInfo),
//	    }
//	}
type CareTeamManager interface {
	// -------------------------------------------------------------------------
	// PART 1 — Basic assignment and lookup
	// -------------------------------------------------------------------------

	// AddMember registers a care team member with the given role and patient
	// capacity. Subsequent calls to Assign become valid for this memberID.
	AddMember(memberID, role string, maxPatients int)

	// Assign assigns a patient to a care team member (assignedAt is Unix seconds).
	//
	// A patient may have at most one assigned member per role at a time.
	// If the patient already has a member with the same role, that assignment
	// is replaced — the new assignment takes effect at assignedAt.
	//
	// Returns ErrNotFound if memberID has not been registered via AddMember.
	//
	// Part 2 addition: returns ErrCapacity if the member is already at
	// MaxPatients and the patient is not currently assigned to that exact member.
	// (Reassigning a patient who is already on this member does not count as
	// adding a new patient — it is a no-op for capacity purposes.)
	Assign(patientID, memberID string, assignedAt float64) error

	// GetAssignment returns the memberID currently assigned to this patient for
	// the given role, or "" if no member of that role is currently assigned.
	GetAssignment(patientID, role string) string

	// GetPatients returns a sorted list of patientIDs currently assigned to this
	// member. Returns ErrNotFound if memberID has not been registered.
	GetPatients(memberID string) ([]string, error)

	// -------------------------------------------------------------------------
	// PART 2 — Capacity enforcement
	// -------------------------------------------------------------------------

	// AvailableMembers returns a sorted list of memberIDs with the given role
	// that still have open capacity (current patient count < maxPatients).
	AvailableMembers(role string) []string

	// -------------------------------------------------------------------------
	// PART 3 — Assignment history
	// -------------------------------------------------------------------------

	// GetHistory returns the full assignment history for the patient's given role
	// as a slice of AssignmentRecords sorted by AssignedAt ascending.
	//
	//   - UnassignedAt is nil for the current (still-active) assignment.
	//   - UnassignedAt equals the AssignedAt of the subsequent assignment for
	//     past entries.
	//   - Returns an empty slice if the patient has never been assigned a member
	//     of this role.
	GetHistory(patientID, role string) []AssignmentRecord

	// GetAssignmentAt returns the memberID assigned to the patient for the given
	// role at the given timestamp, or "" if no assignment was active at that time.
	//
	// An assignment is active during the interval [AssignedAt, UnassignedAt).
	// The current assignment (UnassignedAt is nil) is active from AssignedAt onward.
	//
	// Implement this by calling GetHistory() — do not duplicate the lookup logic.
	GetAssignmentAt(patientID, role string, timestamp float64) string
}

// NewCareTeamManager returns a new CareTeamManager implementation.
//
// You must define your own struct type that implements the CareTeamManager
// interface and return it from this constructor.
//
// All state must be stored in struct fields — do NOT use package-level variables.
func NewCareTeamManager() CareTeamManager {
	panic("not implemented")
}
