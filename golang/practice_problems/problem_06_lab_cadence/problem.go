// =============================================================================
// INTERVIEW PROBLEM 6: Lab Cadence Compliance Monitor
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// A health-tech platform requires patients to submit lab work at regular
// intervals so clinicians can track metabolic health markers (HbA1c, fasting
// glucose, lipids, kidney function, etc.). Patients who miss lab deadlines
// need follow-up from their health coach.
//
// You are building the Lab Cadence Compliance Monitor — a function-based module
// (no class) that manages a registry of patients, their required lab types,
// submission deadlines, and actual submissions.
//
// DATA MODEL
// ----------
// MonitorState — holds all state for the monitor (passed to every function)
//   Patients map[string]*PatientRecord
//
// PatientRecord
//   RequiredLabs map[string]bool         // lab type → required?
//   Deadlines    map[string][]string      // lab type → sorted ISO dates (uncleared deadlines)
//   Submissions  map[string][]string      // lab type → sorted ISO dates
//
// OVERDUE SEMANTICS
// -----------------
//   A required lab is "overdue" if it has an uncleared deadline whose due_date < as_of.
//   A submission clears the earliest uncleared deadline for that lab that is >= submitted_on.
//   If a required lab has no deadline set, it is NOT considered overdue.
//   Late submissions (after the due date) do NOT clear the past deadline.
//
// EXAMPLE
// -------
//   state := MakeMonitor()
//   RegisterPatient(state, "alice", []string{"hba1c", "bmp"})
//   SetLabDeadline(state, "alice", "hba1c", "2024-03-31")
//   RecordSubmission(state, "alice", "hba1c", "2024-03-28")
//   IsOverdue(state, "alice", "hba1c", "2024-04-01")  // -> false (submitted on time)
//   IsOverdue(state, "alice", "bmp", "2024-04-01")    // -> false (no deadline set)
// =============================================================================

package labcadence

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// PatientRecord stores all lab compliance data for a single patient.
type PatientRecord struct {
	RequiredLabs map[string]bool    // lab type → required?
	Deadlines    map[string][]string // lab type → sorted ISO date strings (uncleared)
	Submissions  map[string][]string // lab type → sorted ISO date strings
}

// MonitorState is the top-level data store passed to every function.
type MonitorState struct {
	Patients map[string]*PatientRecord
}

// ComplianceEntry is one row in the report returned by ComplianceReport.
type ComplianceEntry struct {
	PatientID   string
	OverdueLabs []string // sorted lab names
	OverdueCount int
}

// MakeMonitor returns a fresh, empty MonitorState.
func MakeMonitor() *MonitorState {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 1 — Patient & lab registration  (~10 min)
// ---------------------------------------------------------------------------

// RegisterPatient registers a new patient with a list of required lab types.
//
// If the patient already exists, add any new lab types to their required set
// (do not remove existing ones). Idempotent for labs already in the set.
//
// Returns ErrAlreadyExists-style error if requiredLabs is empty (use a sentinel
// error: ErrEmptyLabList).
func RegisterPatient(state *MonitorState, patientID string, requiredLabs []string) error {
	panic("not implemented")
}

// AddRequiredLab adds a single required lab type to an existing patient's
// requirements.
//
// Returns ErrNotFound if the patient doesn't exist.
// No-op if the lab is already required.
func AddRequiredLab(state *MonitorState, patientID, labType string) error {
	panic("not implemented")
}

// GetRequiredLabs returns the set of required lab types for the patient.
// Returns ErrNotFound if the patient doesn't exist.
func GetRequiredLabs(state *MonitorState, patientID string) (map[string]bool, error) {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 2 — Deadlines and submissions  (~15 min)
// ---------------------------------------------------------------------------

// SetLabDeadline adds a deadline (ISO date string "2006-01-02") for a specific
// lab type for the patient.
//
// A patient may have multiple deadlines for the same lab (e.g. quarterly).
// Duplicate deadlines (same patient + lab + date) are ignored.
//
// Returns ErrNotFound if the patient doesn't exist.
// Returns ErrLabNotRequired if labType is not in the patient's required labs.
func SetLabDeadline(state *MonitorState, patientID, labType, dueDate string) error {
	panic("not implemented")
}

// RecordSubmission records that the patient submitted a lab result on submittedOn.
//
// Clears the earliest uncleared deadline for this lab type that is >= submittedOn.
// If no such deadline exists, the submission is still recorded (it may satisfy a
// future deadline or serve as history).
//
// Returns ErrNotFound if the patient doesn't exist.
// Returns ErrLabNotRequired if labType is not in the patient's required labs.
func RecordSubmission(state *MonitorState, patientID, labType, submittedOn string) error {
	panic("not implemented")
}

// IsOverdue returns true if the patient has at least one uncleared deadline for
// labType that has passed as of asOf (i.e., dueDate < asOf).
//
// Returns false if:
//   - The patient doesn't exist.
//   - labType is not required for the patient.
//   - No deadline has been set for that lab.
//   - All past deadlines have been cleared by a submission.
func IsOverdue(state *MonitorState, patientID, labType, asOf string) bool {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 3 — Compliance reporting  (~15 min)
// ---------------------------------------------------------------------------

// OverdueLabs returns a sorted list of lab type names that are currently
// overdue for the patient as of asOf.
//
// Returns an empty slice if the patient doesn't exist or has no overdue labs.
func OverdueLabs(state *MonitorState, patientID, asOf string) []string {
	panic("not implemented")
}

// ComplianceReport returns a report of all patients with at least one overdue
// lab as of asOf.
//
// Each ComplianceEntry contains:
//   - PatientID:    string
//   - OverdueLabs:  []string (sorted lab names)
//   - OverdueCount: int
//
// Sorted by OverdueCount descending (most overdue first), then alphabetically
// by PatientID for ties.
//
// Returns an empty slice if no patient has overdue labs.
func ComplianceReport(state *MonitorState, asOf string) []ComplianceEntry {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 4 — Submission history  (~5 min)
// ---------------------------------------------------------------------------

// SubmissionHistory returns a chronologically sorted list of all submission
// dates for (patientID, labType).
//
// Returns an empty slice if the patient doesn't exist or has no submissions
// for that lab type.
func SubmissionHistory(state *MonitorState, patientID, labType string) []string {
	panic("not implemented")
}

// DaysSinceLastSubmission returns the number of days between the patient's most
// recent submission for labType and asOf.
//
// Returns -1 (as a sentinel for "no data") if the patient has never submitted
// that lab type or the patient doesn't exist.
func DaysSinceLastSubmission(state *MonitorState, patientID, labType, asOf string) int {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// Sentinel errors for Part 1 and Part 2 validation
// ---------------------------------------------------------------------------

// ErrEmptyLabList is returned when RegisterPatient is called with an empty
// requiredLabs list.
var ErrEmptyLabList = errors.New("required labs list must not be empty")

// ErrLabNotRequired is returned when a deadline or submission is recorded for
// a lab type that is not in the patient's required labs.
var ErrLabNotRequired = errors.New("lab type not required for patient")
