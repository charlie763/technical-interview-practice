// =============================================================================
// INTERVIEW PROBLEM 9: Multi-Source Incident Aggregator
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the incident aggregation layer for a public safety
// intelligence platform. The platform ingests incident reports from multiple
// independent data sources (radio dispatch transcriptions, field sensors,
// social media monitors). Different sources frequently report the same
// real-world event, so the system must deduplicate and group raw reports into
// unified incidents.
//
// You choose the internal data structures — the public interface
// (IncidentAggregator) is what matters. Store all state in struct fields set
// in your constructor. Package-level variables will bleed state between
// instances and test runs.
//
// DATA MODEL
// ----------
// Report:
//   ReportID    string
//   SourceID    string
//   EventType   string   // e.g. "shooting", "car-crash", "fire"
//   LocationKey string   // opaque string, e.g. "downtown", "sector-7"
//   Ts          string   // ISO-8601 timestamp, e.g. "2024-01-01T10:00:00"
//   IncidentID  *string  // nil until assigned to an incident
//
// Incident:
//   IncidentID  string
//   EventType   string
//   LocationKey string
//   ReportIDs   []string // report IDs in Ts-ascending order
//   ReportCount int
//   LatestTs    *string  // Ts of the most recently added report, or nil
//
// Timestamps are ISO-8601 strings without timezone offset. Use
// time.Parse("2006-01-02T15:04:05", ts) for arithmetic when comparing or
// computing durations.
//
// EXAMPLE
// -------
//   agg := NewIncidentAggregator()
//   agg.IngestReport("r1", "radio-north", "shooting", "downtown", "2024-01-01T10:00:00")
//   agg.IngestReport("r2", "radio-south", "shooting", "downtown", "2024-01-01T10:00:45")
//   agg.IngestReport("r3", "social-feed", "car-crash", "midtown",  "2024-01-01T10:01:00")
//   inc, _ := agg.CreateIncident("inc-001", "shooting", "downtown")
//   agg.AddReportToIncident("inc-001", "r1")
//   agg.AddReportToIncident("inc-001", "r2")
//   inc, _ = agg.GetIncident("inc-001")
//   // inc.ReportCount == 2
//   unassigned, _ := agg.GetUnassignedReports()
//   // -> [r3 Report]
//   incID, _ := agg.AutoIngestReport("r4", "radio-east", "shooting",
//                                    "downtown", "2024-01-01T10:01:30", 120)
//   // -> "inc-001"  (within 120 s window, same type + location)
// =============================================================================

package incidents

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// ErrAlreadyAssigned is returned when a report is already assigned to an
// incident.
var ErrAlreadyAssigned = errors.New("report already assigned to an incident")

// Report represents a raw incident report from a single source.
type Report struct {
	ReportID    string
	SourceID    string
	EventType   string
	LocationKey string
	Ts          string  // ISO-8601, e.g. "2024-01-01T10:00:00"
	IncidentID  *string // nil until assigned to an incident
}

// Incident represents a unified grouping of related reports.
type Incident struct {
	IncidentID  string
	EventType   string
	LocationKey string
	ReportIDs   []string // kept in Ts-ascending order
	ReportCount int
	LatestTs    *string // Ts of the most recently added report, or nil
}

// IncidentAggregator is the interface all implementations must satisfy.
//
// Implement it by defining your own struct and a NewIncidentAggregator constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myIA struct {
//	    reports   map[string]*Report
//	    incidents map[string]*Incident
//	    // ... add whatever fields you need
//	}
//
//	func NewIncidentAggregator() IncidentAggregator {
//	    return &myIA{
//	        reports:   make(map[string]*Report),
//	        incidents: make(map[string]*Incident),
//	    }
//	}
type IncidentAggregator interface {
	// -------------------------------------------------------------------------
	// PART 1 — Report ingestion  (~10 min)
	// -------------------------------------------------------------------------

	// IngestReport stores a new raw report and returns it.
	// The report's IncidentID starts as nil.
	// Returns ErrAlreadyExists if reportID already exists.
	IngestReport(reportID, sourceID, eventType, locationKey, ts string) (*Report, error)

	// GetReport returns the Report, or nil if not found.
	GetReport(reportID string) *Report

	// GetReports returns all reports, optionally filtered by locationKey and/or
	// eventType (both filters applied when both are given).
	// Results are sorted by Ts ascending.
	// Pass "" to skip a filter.
	GetReports(locationKey, eventType string) []*Report

	// -------------------------------------------------------------------------
	// PART 2 — Manual incident grouping  (~15 min)
	// -------------------------------------------------------------------------

	// CreateIncident creates and returns a new, empty Incident with the given
	// eventType and locationKey.
	// Returns ErrAlreadyExists if incidentID already exists.
	CreateIncident(incidentID, eventType, locationKey string) (*Incident, error)

	// AddReportToIncident assigns a report to an incident.
	//   - Returns ErrNotFound  if incidentID or reportID does not exist.
	//   - Returns ErrAlreadyAssigned if the report is already assigned to any incident.
	//   - Sets Report.IncidentID = &incidentID.
	//   - Updates the incident's ReportIDs (kept in Ts-ascending order),
	//     ReportCount, and LatestTs.
	AddReportToIncident(incidentID, reportID string) error

	// GetIncident returns the Incident (including up-to-date ReportIDs,
	// ReportCount, and LatestTs), or nil if not found.
	// ReportIDs must be ordered by the corresponding report's Ts, ascending.
	GetIncident(incidentID string) *Incident

	// GetUnassignedReports returns all reports whose IncidentID is still nil,
	// sorted by Ts ascending.
	GetUnassignedReports() []*Report

	// -------------------------------------------------------------------------
	// PART 3 — Automatic deduplication  (~20 min)
	// -------------------------------------------------------------------------

	// AutoIngestReport ingests a new report and automatically assigns it to an
	// incident:
	//
	//  1. Call IngestReport to store the report.
	//  2. Find all *active* incidents whose EventType and LocationKey match the
	//     incoming report's. An incident is "active" if its LatestTs is within
	//     timeWindowSecs of the new report's Ts:
	//         LatestTs >= Ts - timeWindowSecs
	//     Incidents with no reports (LatestTs is nil) are not active.
	//  3. If one or more matches exist, pick the one whose LatestTs is closest
	//     to Ts (i.e. most recently active). Break ties by IncidentID
	//     lexicographically ascending.
	//  4. If no active match exists, create a new incident (auto-generate a
	//     unique IncidentID; any scheme is fine as long as it doesn't clash
	//     with existing IDs).
	//  5. Call AddReportToIncident to assign the report.
	//  6. Return the IncidentID.
	AutoIngestReport(reportID, sourceID, eventType, locationKey, ts string, timeWindowSecs int) (string, error)

	// GetActiveIncidents returns all incidents that have a LatestTs within
	// timeWindowSecs of asOfTs:
	//     LatestTs >= asOfTs - timeWindowSecs
	//
	// Sorted by LatestTs descending (most-recently-active first).
	// Incidents with no reports (LatestTs is nil) are excluded.
	GetActiveIncidents(asOfTs string, timeWindowSecs int) []*Incident
}

// NewIncidentAggregator returns a new IncidentAggregator implementation.
//
// You must define your own struct type that implements the IncidentAggregator
// interface and return it from this constructor.
//
// All state must be stored in struct fields — do NOT use package-level variables.
func NewIncidentAggregator() IncidentAggregator {
	panic("not implemented")
}
