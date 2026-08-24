// =============================================================================
// INTERVIEW PROBLEM 4: Biomarker Alert Monitor
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// A health-tech platform delivers remote clinical care for type 2 diabetes
// reversal. Patients use connected glucometers and ketone meters that sync
// readings into the app several times per day. The care team dashboard needs
// to surface patients whose numbers have been out of target range for multiple
// consecutive days, so health coaches can prioritize outreach.
//
// You are building the BiomarkerMonitor — the type that processes a stream of
// patient readings and answers questions about trends and outreach priority.
//
// DATA MODEL
// ----------
// BiomarkerReading — PatientID, ReadingType ("glucose"|"ketone"|"weight"),
//                    Value (float64), RecordedOn (string "2006-01-02")
//
// Target ranges (inclusive):
//   Glucose: 70–180 mg/dL   (< 70 = dangerous low, > 180 = hyperglycemia)
//   Ketone:  0.5–3.0 mmol/L (below = not in ketosis, above = monitor)
//   Weight:  no absolute target — never flagged as out-of-range
//
// OutreachEntry — PatientID, ReadingType, ConsecutiveDays, LatestValue
//
// NOTES
// -----
//   - Multiple readings on the same calendar day count as ONE day.
//     A day is "out-of-range" if ANY reading that day is out of range.
//   - All mutable state must be stored in struct fields set in the constructor.
//     Do NOT use package-level variables — they bleed state between instances.
//
// EXAMPLE
// -------
//   readings := []BiomarkerReading{
//       {"alice", "glucose", 195.0, "2024-01-01"},
//       {"alice", "glucose", 202.0, "2024-01-02"},
//       {"alice", "glucose", 188.0, "2024-01-03"},
//   }
//   m := NewBiomarkerMonitor(readings)
//   m.MaxConsecutiveOutOfRangeDays("alice", "glucose")  // -> 3
//   m.GetOutreachList(3)
//   // -> [{PatientID:"alice", ReadingType:"glucose", ConsecutiveDays:3, LatestValue:188.0}]
// =============================================================================

package biomarker

import "errors"

// GlucoseMin and GlucoseMax define the inclusive target range for glucose (mg/dL).
const GlucoseMin = 70.0
const GlucoseMax = 180.0

// KetoneMin and KetoneMax define the inclusive target range for ketones (mmol/L).
const KetoneMin = 0.5
const KetoneMax = 3.0

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// BiomarkerReading is a single biomarker measurement from a patient device or
// manual entry. RecordedOn is an ISO date string ("2006-01-02").
type BiomarkerReading struct {
	PatientID   string
	ReadingType string // "glucose" | "ketone" | "weight"
	Value       float64
	RecordedOn  string // ISO date: "2006-01-02"
}

// OutreachEntry describes a patient who needs proactive coach outreach.
type OutreachEntry struct {
	PatientID      string
	ReadingType    string
	ConsecutiveDays int
	LatestValue    float64
}

// BiomarkerMonitor is the interface candidates must implement.
//
// Implement it by defining your own struct type and a NewBiomarkerMonitor
// constructor. Store all state in struct fields — do NOT use package-level
// variables, as they bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myMonitor struct {
//	    readings []BiomarkerReading
//	    // ... add whatever fields you need
//	}
//
//	func NewBiomarkerMonitor(readings []BiomarkerReading) BiomarkerMonitor {
//	    return &myMonitor{
//	        readings: append([]BiomarkerReading(nil), readings...),
//	    }
//	}
type BiomarkerMonitor interface {
	// -------------------------------------------------------------------------
	// PART 1 — Single-reading classification  (~5 min)
	// -------------------------------------------------------------------------

	// IsOutOfRange returns true if the reading falls outside the target range
	// for its reading type.
	//   - Glucose: outside [70, 180] mg/dL (inclusive)
	//   - Ketone:  outside [0.5, 3.0] mmol/L (inclusive)
	//   - Weight:  never out-of-range (always returns false)
	IsOutOfRange(reading BiomarkerReading) bool

	// -------------------------------------------------------------------------
	// PART 2 — Streak detection  (~15 min)
	// -------------------------------------------------------------------------

	// MaxConsecutiveOutOfRangeDays returns the length of the longest streak of
	// consecutive calendar days on which the patient had at least one out-of-range
	// reading of the given type.
	//
	// Returns 0 if the patient has no out-of-range readings of that type.
	//
	// "Consecutive" means no gap: Jan 1, Jan 2, Jan 3 is a streak of 3.
	// Jan 1 and Jan 3 (skipping Jan 2) are two separate streaks of 1.
	// Multiple readings on the same day collapse to one day.
	MaxConsecutiveOutOfRangeDays(patientID, readingType string) int

	// -------------------------------------------------------------------------
	// PART 3 — Outreach list  (~10 min)
	// -------------------------------------------------------------------------

	// GetOutreachList returns a list of patients who need proactive coach
	// outreach because they have been out-of-range for at least
	// minConsecutiveDays in a row.
	//
	// A patient can appear more than once if multiple reading types cross the
	// threshold (e.g., both glucose and ketone streaks).
	//
	// The list is sorted by ConsecutiveDays descending (most urgent first).
	GetOutreachList(minConsecutiveDays int) []OutreachEntry

	// -------------------------------------------------------------------------
	// PART 4 — Deduplication on ingestion  (~10 min)
	// -------------------------------------------------------------------------

	// AddReading adds a new reading to the monitor's internal state.
	// Returns true if the reading was added successfully.
	// Returns false (without adding) if it is a duplicate.
	//
	// A duplicate is: same PatientID, ReadingType, and RecordedOn,
	// with a Value within ±0.5 of an existing reading on that day.
	AddReading(reading BiomarkerReading) bool
}

// NewBiomarkerMonitor returns a fresh BiomarkerMonitor seeded with the given
// readings. Candidates implement their own struct type in the answer file.
func NewBiomarkerMonitor(readings []BiomarkerReading) BiomarkerMonitor {
	panic("not implemented")
}
