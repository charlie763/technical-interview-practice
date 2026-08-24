// =============================================================================
// INTERVIEW PROBLEM 8: Patient Enrollment Pipeline
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// A clinical care program moves patients through a structured enrollment
// pipeline. Each patient begins in the "referred" state and advances through a
// predefined set of transitions until reaching a terminal state.
//
// You choose the internal data structures — the public interface
// (EnrollmentPipeline) is what matters. Store all state in struct fields set
// in your constructor. Package-level variables will bleed state between
// instances and test runs.
//
// DATA MODEL
// ----------
// Patient — ID, current state, timestamps per state visited
//
// ALLOWED TRANSITIONS
// -------------------
//   referred   →  screened
//   screened   →  enrolled  |  ineligible
//   enrolled   →  active    |  withdrawn
//   active     →  graduated |  churned  |  withdrawn
//
// Terminal states: ineligible, withdrawn, graduated, churned
//
// EXAMPLE
// -------
//   pipeline := NewEnrollmentPipeline()
//   pipeline.AddPatient("p_001", 0.0)
//   pipeline.Transition("p_001", "screened", 86400.0)    // 1 day later
//   pipeline.Transition("p_001", "enrolled", 172800.0)   // 2 days later
//   pipeline.GetState("p_001")                   // "enrolled", nil
//   pipeline.GetPatientsInState("enrolled")      // ["p_001"]
//
//   pipeline.AddPatient("p_002", 0.0)
//   pipeline.Transition("p_002", "screened",   43200.0)
//   pipeline.Transition("p_002", "ineligible", 86400.0)
//   pipeline.GetPatientsInState("screened")      // [] (both have moved on)
//
//   // Part 2
//   pipeline.TimeInState("p_001", "screened", 999999.0)
//   // -> 86400.0  (172800 - 86400; already exited, as_of ignored)
//   pipeline.ConversionRate("screened", "enrolled")
//   // -> 0.5  (p_001 enrolled, p_002 ineligible; one of two converted)
//
//   // Part 3
//   pipeline.Transition("p_001", "active", 259200.0)
//   pipeline.PatientsOverdue("active", 3600.0, 270000.0)
//   // -> ["p_001"]  (has been active 10800 s > 3600 s threshold)
//   pipeline.AverageTimeInState("screened", 999999.0)
//   // -> 64800.0  ((86400 + 43200) / 2; both patients have exited screened)
// =============================================================================

package enrollment

import "errors"

// ValidTransitions defines which states a patient may advance to from each
// non-terminal state.
var ValidTransitions = map[string][]string{
	"referred": {"screened"},
	"screened": {"enrolled", "ineligible"},
	"enrolled": {"active", "withdrawn"},
	"active":   {"graduated", "churned", "withdrawn"},
}

// TerminalStates lists the states from which no further transitions are allowed.
var TerminalStates = map[string]bool{
	"ineligible": true,
	"withdrawn":  true,
	"graduated":  true,
	"churned":    true,
}

// ErrNotFound is returned when accessing a patient that has not been registered.
var ErrNotFound = errors.New("not found")

// ErrAlreadyExists is returned when adding a patient that is already registered.
var ErrAlreadyExists = errors.New("already exists")

// ErrInvalidTransition is returned when attempting an invalid state transition.
var ErrInvalidTransition = errors.New("invalid transition")

// EnrollmentPipeline is the interface all implementations must satisfy.
//
// Implement it by defining your own struct and a NewEnrollmentPipeline constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myEP struct {
//	    patients map[string]*patientRecord
//	    // ... add whatever fields you need
//	}
//
//	func NewEnrollmentPipeline() EnrollmentPipeline {
//	    return &myEP{
//	        patients: make(map[string]*patientRecord),
//	    }
//	}
type EnrollmentPipeline interface {
	// -------------------------------------------------------------------------
	// PART 1 — State tracking
	// -------------------------------------------------------------------------

	// AddPatient registers a patient in the pipeline at the "referred" state.
	// timestamp is when they entered the "referred" state (Unix seconds).
	// Returns ErrAlreadyExists if patientID is already registered.
	AddPatient(patientID string, timestamp float64) error

	// Transition advances a patient to newState at the given timestamp.
	//
	// Returns ErrNotFound if patientID is not registered.
	// Returns ErrInvalidTransition if:
	//   - newState is not a valid next state from the patient's current state
	//     (consult ValidTransitions).
	//   - the patient is already in a terminal state.
	Transition(patientID, newState string, timestamp float64) error

	// GetState returns the patient's current state.
	// Returns ErrNotFound if patientID is not registered.
	GetState(patientID string) (string, error)

	// GetPatientsInState returns a sorted list of patientIDs currently in the
	// given state.
	GetPatientsInState(state string) []string

	// -------------------------------------------------------------------------
	// PART 2 — Duration and conversion metrics
	// -------------------------------------------------------------------------

	// TimeInState returns the total seconds the patient has spent in the given
	// state.
	//
	//   - If the patient is currently in that state, count time from state entry
	//     up to asOf.
	//   - If the patient has already left that state, return the exact duration
	//     spent there (asOf is ignored).
	//   - Returns 0.0 if the patient has never been in that state.
	//
	// With the allowed transitions above, each state is visited at most once,
	// so there is no ambiguity about multiple visits.
	TimeInState(patientID, state string, asOf float64) float64

	// ConversionRate returns the fraction of patients who exited fromState and
	// transitioned directly to toState.
	//
	//   - Only patients who have already left fromState are counted; patients
	//     currently sitting in fromState are excluded (still undecided).
	//   - Returns 0.0 if no patients have exited fromState yet.
	//
	// Example: ConversionRate("screened", "enrolled") returns the share of
	// screened patients who went on to enroll (vs. being marked ineligible).
	ConversionRate(fromState, toState string) float64

	// -------------------------------------------------------------------------
	// PART 3 — SLA monitoring
	// -------------------------------------------------------------------------

	// PatientsOverdue returns patientIDs currently in state who have spent more
	// than maxSeconds there, sorted by time spent descending (longest-waiting
	// first).
	//
	// Call TimeInState() for each patient's duration — do not re-implement the
	// duration logic here.
	PatientsOverdue(state string, maxSeconds, asOf float64) []string

	// AverageTimeInState returns the mean seconds spent in state across all
	// patients who have fully exited that state (their time is complete and will
	// not grow further).
	//
	//   - Patients currently in state are excluded from the average.
	//   - Returns 0.0 if no patients have fully exited state yet.
	//
	// Call TimeInState() for each patient's duration.
	AverageTimeInState(state string, asOf float64) float64
}

// NewEnrollmentPipeline returns a new EnrollmentPipeline implementation.
//
// You must define your own struct type that implements the EnrollmentPipeline
// interface and return it from this constructor.
//
// All state must be stored in struct fields — do NOT use package-level variables.
func NewEnrollmentPipeline() EnrollmentPipeline {
	panic("not implemented")
}
