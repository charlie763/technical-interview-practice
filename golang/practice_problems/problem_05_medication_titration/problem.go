// =============================================================================
// INTERVIEW PROBLEM 5: Medication Titration Tracker
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// A health-tech platform's remote clinical care program often de-escalates
// (reduces or stops) diabetes medications as patients' blood sugar improves.
// Coaches and physicians need to track each patient's medication history —
// when doses were changed and why — so they can coordinate care and generate
// compliance reports.
//
// You are building the TitrationTracker — the type that ingests a stream of
// titration events and answers questions about each patient's medication history.
//
// DATA MODEL
// ----------
// TitrationEvent — PatientID, Medication, Direction, DoseMg, RecordedOn (string "2006-01-02")
//   Direction vocabulary:
//     "increase" — dose or frequency was raised
//     "decrease" — dose or frequency was lowered (de-escalation)
//     "stop"     — medication discontinued entirely
//     "start"    — new medication introduced
//
// Medication — Name, CurrentDose, LastChanged, TotalChanges
//   A medication is "active" if the most recent event for it is NOT "stop".
//
// NOTES
// -----
//   - Events are not guaranteed to arrive in chronological order — sort by date.
//   - A medication is "active" if the most recent event for it is NOT "stop".
//   - All mutable state must be stored in struct fields set in the constructor.
//     Do NOT use package-level variables — they bleed state between instances.
//   - You choose the internal data structures — the public interface is what matters.
//
// EXAMPLE
// -------
//   events := []TitrationEvent{
//       {"pt1", "metformin", "start",    500.0, "2024-01-01"},
//       {"pt1", "metformin", "increase", 1000.0, "2024-02-01"},
//       {"pt1", "metformin", "decrease",  500.0, "2024-03-01"},
//       {"pt1", "metformin", "stop",        0.0, "2024-04-01"},
//   }
//   t := NewTitrationTracker(events)
//   t.CurrentMedications("pt1")            // -> []
//   t.TitrationCount("pt1", "metformin", "decrease")  // -> 1
// =============================================================================

package titration

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// TitrationEvent is a single medication change recorded by a clinical coach
// or physician. RecordedOn is an ISO date string ("2006-01-02").
type TitrationEvent struct {
	PatientID  string
	Medication string // e.g. "metformin", "glipizide", "insulin_glargine"
	Direction  string // "start" | "increase" | "decrease" | "stop"
	DoseMg     float64 // dose in milligrams at the time of this event (0.0 for "stop")
	RecordedOn string  // ISO date: "2006-01-02"
}

// Medication is a summary of a patient's current relationship with a single
// medication.
type Medication struct {
	Name         string
	CurrentDose  float64
	LastChanged  string // ISO date: "2006-01-02"
	TotalChanges int    // total number of titration events (including start/stop)
}

// TitrationTracker is the interface candidates must implement.
//
// Implement it by defining your own struct type and a NewTitrationTracker
// constructor. Store all state in struct fields — do NOT use package-level
// variables, as they bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myTracker struct {
//	    events []TitrationEvent
//	    // ... add whatever fields you need
//	}
//
//	func NewTitrationTracker(events []TitrationEvent) TitrationTracker {
//	    return &myTracker{
//	        events: append([]TitrationEvent(nil), events...),
//	    }
//	}
type TitrationTracker interface {
	// -------------------------------------------------------------------------
	// PART 1 — Current medication snapshot  (~10 min)
	// -------------------------------------------------------------------------

	// CurrentMedications returns a list of Medication values for all currently
	// active medications for the given patient (i.e., medications whose latest
	// event is NOT "stop").
	//
	// Each Medication reflects:
	//   - Name:         the medication name
	//   - CurrentDose:  DoseMg from the most recent event for that medication
	//   - LastChanged:  date of the most recent event
	//   - TotalChanges: total number of TitrationEvents recorded for this medication
	//
	// Returns an empty slice if the patient has no events or all medications have
	// been stopped. The slice may be returned in any order.
	CurrentMedications(patientID string) []Medication

	// GetMedicationHistory returns all TitrationEvents for (patientID, medication),
	// sorted chronologically (earliest first).
	//
	// Returns an empty slice if no events exist for that combination.
	GetMedicationHistory(patientID, medication string) []TitrationEvent

	// -------------------------------------------------------------------------
	// PART 2 — Titration counts  (~10 min)
	// -------------------------------------------------------------------------

	// TitrationCount returns the number of titration events for
	// (patientID, medication).
	//
	// If direction is non-empty (one of "start", "increase", "decrease", "stop"),
	// return only events with that direction.
	//
	// Returns 0 if the patient or medication is unknown.
	TitrationCount(patientID, medication, direction string) int

	// DeEscalationSummary returns a map from medication name to the number of
	// "decrease" or "stop" events recorded for that patient.
	//
	// Only includes medications that have at least one decrease or stop event.
	// Returns an empty map if the patient has no such events.
	//
	// Example:
	//   {
	//     "metformin": 2,  // 1 decrease + 1 stop
	//     "glipizide": 1,  // 1 stop only
	//   }
	DeEscalationSummary(patientID string) map[string]int

	// -------------------------------------------------------------------------
	// PART 3 — Population-level queries  (~15 min)
	// -------------------------------------------------------------------------

	// PatientsOnMedication returns a sorted list of patientIDs who currently
	// have the given medication active (latest event is NOT "stop").
	PatientsOnMedication(medication string) []string

	// MostTitratedMedications returns the top N medications by total titration
	// event count across ALL patients as a []MedCount (see type below), sorted
	// descending by Count. If fewer than topN medications exist, returns all.
	// Ties may appear in any order.
	MostTitratedMedications(topN int) []MedCount

	// -------------------------------------------------------------------------
	// PART 4 — Live ingestion  (~10 min)
	// -------------------------------------------------------------------------

	// AddEvent adds a new TitrationEvent to the tracker.
	//
	// If an event with the same (PatientID, Medication, RecordedOn) already exists,
	// overwrite it with the new event (last-write wins).
	AddEvent(event TitrationEvent)
}

// MedCount is a (medication name, total event count) pair returned by
// MostTitratedMedications.
type MedCount struct {
	Name  string
	Count int
}

// NewTitrationTracker returns a fresh TitrationTracker seeded with the given
// events. Candidates implement their own struct type in the answer file.
func NewTitrationTracker(events []TitrationEvent) TitrationTracker {
	panic("not implemented")
}
