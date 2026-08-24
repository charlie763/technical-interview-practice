// =============================================================================
// INTERVIEW PROBLEM 10: Responder Dispatch Manager
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the dispatch assignment layer for an emergency-response
// platform. Incident alerts stream in and need to be routed to available field
// responders. Responders specialize in certain incident types and have a
// capacity limit — the maximum number of simultaneous open (unresolved)
// incidents they can handle.
//
// Store all state in struct fields set in your constructor.
// Package-level variables will bleed state between instances and test runs —
// avoid them. You choose the internal data structures; the public interface
// (DispatchManager) is what matters.
//
// DATA MODEL
// ----------
// Responder:
//   ID              string
//   Name            string
//   SubscribedTypes []string  // incident types this responder handles
//   Capacity        int       // max simultaneous open assignments
//
// Incident:
//   ID           string
//   IncidentType string   // e.g. "shooting", "car-crash", "fire"
//   Severity     int      // 1 (low) – 5 (critical)
//   Ts           string   // ISO-8601 timestamp, when reported
//   ResponderID  *string  // nil until assigned
//   Resolved     bool     // false until ResolveIncident is called
//
// DispatchSummary (returned by GetDispatchSummary):
//   ResponderID        string
//   Name               string
//   Capacity           int
//   OpenCount          int   // current open assignments
//   AvailableCapacity  int   // Capacity - OpenCount
//
// EXAMPLE
// -------
//   dm := NewDispatchManager()
//   dm.RegisterResponder("unit-12", "Alpha Team", []string{"shooting", "robbery"}, 3)
//   dm.RegisterResponder("unit-14", "Beta Team",  []string{"car-crash", "fire"},   2)
//   dm.AddIncident("inc-001", "shooting",  5, "2024-01-01T10:00:00")
//   dm.AddIncident("inc-002", "car-crash", 3, "2024-01-01T10:01:00")
//   dm.AssignIncident("inc-001", "unit-12")
//   dm.GetOpenAssignments("unit-12")  // -> [Incident{ID: "inc-001", ...}]
//   dm.AutoAssign("inc-002")          // -> "unit-14", nil
// =============================================================================

package dispatch

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// ErrAlreadyAssigned is returned when assigning an incident that already has a responder.
var ErrAlreadyAssigned = errors.New("already assigned")

// ErrAtCapacity is returned when the target responder has no remaining capacity.
var ErrAtCapacity = errors.New("responder at capacity")

// ErrAlreadyResolved is returned when resolving an incident that is already resolved.
var ErrAlreadyResolved = errors.New("already resolved")

// ErrNoEligibleResponder is returned by AutoAssign when no responder is available.
var ErrNoEligibleResponder = errors.New("no eligible responder available")

// Responder represents a field responder who can be assigned to incidents.
type Responder struct {
	ID              string
	Name            string
	SubscribedTypes []string
	Capacity        int
}

// Incident represents an emergency incident that needs to be handled.
type Incident struct {
	ID          string
	IncidentType string
	Severity    int
	Ts          string
	ResponderID *string // nil until assigned
	Resolved    bool
}

// DispatchSummary is returned by GetDispatchSummary for each registered responder.
type DispatchSummary struct {
	ResponderID       string
	Name              string
	Capacity          int
	OpenCount         int
	AvailableCapacity int
}

// DispatchManager is the interface candidates must implement.
//
// Implement it by defining your own struct and a NewDispatchManager constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myDM struct {
//	    responders map[string]*Responder
//	    incidents  map[string]*Incident
//	    // ... add whatever fields you need
//	}
//
//	func NewDispatchManager() DispatchManager {
//	    return &myDM{
//	        responders: make(map[string]*Responder),
//	        incidents:  make(map[string]*Incident),
//	    }
//	}
type DispatchManager interface {
	// -------------------------------------------------------------------------
	// PART 1 — Registration and basic queries  (~10 min)
	// -------------------------------------------------------------------------

	// RegisterResponder registers a new responder and returns it.
	// Returns ErrAlreadyExists if responderID already exists.
	RegisterResponder(responderID, name string, subscribedTypes []string, capacity int) (*Responder, error)

	// AddIncident adds a new incident (unassigned, unresolved) and returns it.
	// Returns ErrAlreadyExists if incidentID already exists.
	AddIncident(incidentID, incidentType string, severity int, ts string) (*Incident, error)

	// GetIncidentsForResponder returns all incidents whose incidentType appears in
	// the responder's SubscribedTypes list, regardless of whether the incident has
	// been assigned yet.
	//
	// Sort order: Severity descending (5 first), then Ts ascending (oldest first
	// within the same severity).
	//
	// Returns ErrNotFound if responderID does not exist.
	GetIncidentsForResponder(responderID string) ([]*Incident, error)

	// -------------------------------------------------------------------------
	// PART 2 — Assignment and resolution  (~15 min)
	// -------------------------------------------------------------------------

	// AssignIncident assigns an incident to a responder.
	// Returns ErrNotFound      if incidentID or responderID does not exist.
	// Returns ErrAlreadyAssigned if the incident already has a responder assigned.
	// Returns ErrAtCapacity    if the responder's open assignment count equals their capacity.
	// On success, sets Incident.ResponderID = &responderID.
	AssignIncident(incidentID, responderID string) error

	// ResolveIncident marks an incident as resolved (sets Resolved = true), freeing
	// the assigned responder's capacity slot.
	// Returns ErrNotFound      if incidentID does not exist.
	// Returns ErrAlreadyResolved if the incident is already resolved.
	ResolveIncident(incidentID string) error

	// GetOpenAssignments returns all incidents that are assigned to this responder
	// and not yet resolved.
	// Sort order: Severity descending, then Ts ascending.
	// Returns ErrNotFound if responderID does not exist.
	GetOpenAssignments(responderID string) ([]*Incident, error)

	// -------------------------------------------------------------------------
	// PART 3 — Auto-assignment  (~20 min)
	// -------------------------------------------------------------------------

	// AutoAssign automatically assigns an incident to the best available responder.
	//
	// Eligibility (both must hold):
	//   1. The responder's SubscribedTypes includes the incident's IncidentType.
	//   2. The responder's current open-assignment count is less than their Capacity.
	//
	// Selection — among eligible responders, prefer:
	//   1. Fewest open assignments (least loaded).
	//   2. Tie-break: highest Capacity.
	//   3. Tie-break: responderID lexicographically ascending.
	//
	// Returns ErrNotFound         if incidentID does not exist.
	// Returns ErrAlreadyAssigned  if the incident is already assigned.
	// Returns ErrNoEligibleResponder if no eligible responder is available.
	//
	// Calls AssignIncident internally to perform the assignment.
	// Returns the responderID of the chosen responder.
	AutoAssign(incidentID string) (string, error)

	// GetDispatchSummary returns a slice of DispatchSummary — one per registered
	// responder — sorted by ResponderID ascending.
	GetDispatchSummary() []*DispatchSummary
}

// NewDispatchManager returns a fresh, empty DispatchManager implementation.
// Candidates implement their own struct type in the answer file.
func NewDispatchManager() DispatchManager {
	panic("not implemented")
}
