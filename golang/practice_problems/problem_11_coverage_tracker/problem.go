// =============================================================================
// INTERVIEW PROBLEM 11: Sensor Coverage Tracker
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the health-monitoring subsystem for a platform that deploys
// radio-receiver sensor stations across geographic regions. Each station
// periodically sends a heartbeat. When a station falls silent, operators need
// to know, and the platform's incident-detection coverage for that region may
// be affected.
//
// Store all state in struct fields set in your constructor.
// Package-level variables will bleed state between instances and test runs —
// avoid them. You choose the internal data structures; the public interface
// (CoverageTracker) is what matters.
//
// DATA MODEL
// ----------
// Station:
//   ID     string
//   Name   string
//   Region string  // logical grouping, e.g. "downtown", "sector-7"
//
// Outage:
//   StationID string
//   StartTs   string  // ISO-8601 when the outage began, e.g. "2024-01-01T10:00:00"
//   EndTs     *string // ISO-8601 when the outage ended; nil = ongoing
//
// RegionCoverage (returned by GetRegionCoverage):
//   Region      string
//   Total       int   // stations in this region
//   Healthy     int   // stations NOT stale
//   Stale       int   // stations that ARE stale
//   HasCoverage bool  // true if Healthy >= 1
//
// OutageSummary (returned by GetOutageSummary):
//   StationID       string
//   TotalOutages    int
//   OpenOutage      bool
//   TotalOutageSecs int
//
// TIMESTAMP FORMAT
// ----------------
// All timestamps are ISO-8601 strings without timezone offset,
// e.g. "2024-01-01T10:00:00". Use time.Parse with layout "2006-01-02T15:04:05".
//
// EXAMPLE
// -------
//   ct := NewCoverageTracker()
//   ct.RegisterStation("sta-001", "North Tower", "downtown")
//   ct.RegisterStation("sta-002", "South Tower", "downtown")
//   ct.RecordHeartbeat("sta-001", "2024-01-01T10:00:00")
//   ct.RecordHeartbeat("sta-002", "2024-01-01T10:00:05")
//   ct.GetLastHeartbeat("sta-001")  // -> "2024-01-01T10:00:00", nil
//   ct.GetStaleStations("2024-01-01T10:05:00", 120)  // -> [], nil
//   ct.RecordOutageStart("sta-001", "2024-01-01T10:10:00")
//   ct.GetRegionCoverage("downtown", "2024-01-01T10:10:30", 120)
//   // -> {Region:"downtown", Total:2, Healthy:1, Stale:1, HasCoverage:true}, nil
// =============================================================================

package coverage

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// ErrOutOfOrder is returned when a heartbeat timestamp is not strictly after
// the station's most recent heartbeat.
var ErrOutOfOrder = errors.New("heartbeat timestamp out of order")

// ErrOpenOutageExists is returned when opening a new outage while one is already open.
var ErrOpenOutageExists = errors.New("station already has an open outage")

// ErrNoOpenOutage is returned when closing an outage but none is open.
var ErrNoOpenOutage = errors.New("station has no open outage")

// Station represents a sensor station in a geographic region.
type Station struct {
	ID     string
	Name   string
	Region string
}

// Outage represents a period when a station was unavailable.
type Outage struct {
	StationID string
	StartTs   string
	EndTs     *string // nil if the outage is still ongoing
}

// RegionCoverage summarizes health coverage for a region at a point in time.
type RegionCoverage struct {
	Region      string
	Total       int
	Healthy     int
	Stale       int
	HasCoverage bool
}

// OutageSummary summarizes the outage history for a station.
type OutageSummary struct {
	StationID       string
	TotalOutages    int
	OpenOutage      bool
	TotalOutageSecs int
}

// CoverageTracker is the interface candidates must implement.
//
// Implement it by defining your own struct and a NewCoverageTracker constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myCT struct {
//	    stations map[string]*Station
//	    // ... add whatever fields you need
//	}
//
//	func NewCoverageTracker() CoverageTracker {
//	    return &myCT{
//	        stations: make(map[string]*Station),
//	    }
//	}
type CoverageTracker interface {
	// -------------------------------------------------------------------------
	// PART 1 — Station registration and heartbeats  (~10 min)
	// -------------------------------------------------------------------------

	// RegisterStation registers a new station and returns it.
	// Returns ErrAlreadyExists if stationID already exists.
	RegisterStation(stationID, name, region string) (*Station, error)

	// RecordHeartbeat records a heartbeat for the station.
	// Returns ErrNotFound  if stationID does not exist.
	// Returns ErrOutOfOrder if ts is earlier than or equal to the station's most
	// recent heartbeat (out-of-order and duplicate heartbeats are rejected).
	RecordHeartbeat(stationID, ts string) error

	// GetLastHeartbeat returns the timestamp of the most recent heartbeat, or ""
	// if the station has never sent one. The second return value is always nil
	// when the station exists; returns ErrNotFound if stationID does not exist.
	GetLastHeartbeat(stationID string) (string, error)

	// GetStations returns all stations, optionally filtered to a specific region.
	// Pass region="" to return all stations. Sorted by stationID ascending.
	GetStations(region string) ([]*Station, error)

	// -------------------------------------------------------------------------
	// PART 2 — Staleness detection and outage tracking  (~15 min)
	// -------------------------------------------------------------------------

	// GetStaleStations returns station structs for all stations that are stale as
	// of asOfTs. A station is stale if:
	//   - It has never sent a heartbeat, OR
	//   - Its last heartbeat was more than staleAfterSecs seconds before asOfTs
	//     (i.e. asOfTs - lastHeartbeat > staleAfterSecs).
	//
	// Results are sorted by stationID ascending.
	GetStaleStations(asOfTs string, staleAfterSecs int) ([]*Station, error)

	// RecordOutageStart opens a new outage record for the station (EndTs = nil).
	// Returns ErrNotFound       if stationID does not exist.
	// Returns ErrOpenOutageExists if the station already has an open outage.
	RecordOutageStart(stationID, ts string) error

	// RecordOutageEnd closes the most recent open outage for the station by
	// setting its EndTs = ts.
	// Returns ErrNotFound    if stationID does not exist.
	// Returns ErrNoOpenOutage if the station has no open outage.
	RecordOutageEnd(stationID, ts string) error

	// GetOutages returns all Outage structs for the station, sorted by StartTs
	// ascending.
	// Returns ErrNotFound if stationID does not exist.
	GetOutages(stationID string) ([]*Outage, error)

	// -------------------------------------------------------------------------
	// PART 3 — Coverage analysis  (~20 min)
	// -------------------------------------------------------------------------

	// GetRegionCoverage returns a coverage summary for the region at asOfTs.
	// Uses GetStations and GetStaleStations internally.
	GetRegionCoverage(region, asOfTs string, staleAfterSecs int) (*RegionCoverage, error)

	// GetOutageSummary returns an outage summary for the station.
	// For an open outage (EndTs is nil), count duration from StartTs up to asOfTs.
	// Uses GetOutages internally.
	// Returns ErrNotFound if stationID does not exist.
	GetOutageSummary(stationID, asOfTs string) (*OutageSummary, error)
}

// NewCoverageTracker returns a fresh, empty CoverageTracker implementation.
// Candidates implement their own struct type in the answer file.
func NewCoverageTracker() CoverageTracker {
	panic("not implemented")
}
