// =============================================================================
// INTERVIEW PROBLEM 1: Geofence Alert Rule Engine
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building a backend service for an IoT asset-tracking platform. Physical
// assets (forklifts, shipping containers, field equipment) carry GPS sensors that
// periodically report coordinates. The platform tracks which geographic "zone"
// (geofence) each asset is currently inside, and fires configured alert rules
// whenever an asset transitions between zones.
//
// For this problem, zones are axis-aligned bounding boxes — no geospatial
// libraries needed.
//
// DATA MODEL
// ----------
// Zone     — ID, Name, MinLat/MaxLat/MinLng/MaxLng bounding box (inclusive)
// Asset    — ID, Name, Lat/Lng (*float64, nil until first GPS ping), ZoneID (*string)
// AlertRule — ID, FromZoneID/ToZoneID/AssetID (all *string, nil = wildcard)
// TriggeredAlert — RuleID, AssetID, FromZoneID, ToZoneID, Timestamp
// TrackerState   — Zones map, ZoneOrder []string, Assets map, AlertRules/AlertLog slices
//
// ALERT-RULE MATCHING
// -------------------
// A rule matches a transition when ALL three conditions hold:
//   rule.AssetID     == nil  OR  *rule.AssetID     == assetID
//   rule.FromZoneID  == nil  OR  *rule.FromZoneID  == oldZoneID (or both nil)
//   rule.ToZoneID    == nil  OR  *rule.ToZoneID    == newZoneID (or both nil)
//
// EXAMPLE
// -------
//   state := MakeTracker()
//   AddZone(state, "wh", "Warehouse", 35.0, 35.1, -106.7, -106.6)
//   AddAsset(state, "fl1", "Forklift #1")
//   ruleID := "entry_alert"
//   AddAlertRule(state, ruleID, nil, strPtr("wh"), nil)
//   alerts, _ := ProcessLocationUpdate(state, "fl1", 35.05, -106.65, "2024-01-01T00:00:00Z")
//   // len(alerts) == 1, alerts[0].RuleID == "entry_alert"
// =============================================================================

package geofence

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// TrackerState holds all geofence tracking state.
type TrackerState struct {
	Zones      map[string]*Zone
	ZoneOrder  []string // insertion order for deterministic zone iteration
	Assets     map[string]*Asset
	AlertRules []*AlertRule
	AlertLog   []*TriggeredAlert
}

// Zone represents a geographic axis-aligned bounding box.
type Zone struct {
	ID     string
	Name   string
	MinLat float64
	MaxLat float64
	MinLng float64
	MaxLng float64
}

// Asset represents a tracked physical asset with an optional GPS location.
type Asset struct {
	ID     string
	Name   string
	Lat    *float64 // nil until the first GPS update arrives
	Lng    *float64 // nil until the first GPS update arrives
	ZoneID *string  // nil if not currently inside any zone
}

// AlertRule defines the conditions for a zone-transition alert.
type AlertRule struct {
	ID         string
	FromZoneID *string // nil = any previous zone (including "no zone")
	ToZoneID   *string // nil = any new zone (including "no zone")
	AssetID    *string // nil = any asset
}

// TriggeredAlert is appended to TrackerState.AlertLog when a rule fires.
type TriggeredAlert struct {
	RuleID     string
	AssetID    string
	FromZoneID *string
	ToZoneID   *string
	Timestamp  string
}

// MakeTracker returns a fresh, empty TrackerState.
func MakeTracker() *TrackerState {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 1 — Zone membership  (warm-up, ~5 min)
// ---------------------------------------------------------------------------

// IsInZone returns true if the asset's Lat/Lng falls inside the zone's bounding box.
//   - Bounds are inclusive on all four edges.
//   - Returns false if the asset has no location (Lat or Lng is nil).
func IsInZone(asset *Asset, zone *Zone) bool {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 2 — Locate an asset  (~5 min)
// ---------------------------------------------------------------------------

// GetCurrentZoneID returns a pointer to the ID of the first zone (in insertion
// order via TrackerState.ZoneOrder) that contains the asset, or nil.
// Returns nil if the asset doesn't exist or has no location.
func GetCurrentZoneID(state *TrackerState, assetID string) *string {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 3 — Process a location update  (core logic, ~15 min)
// ---------------------------------------------------------------------------

// ProcessLocationUpdate handles a new GPS reading for an asset:
//  1. Updates the asset's Lat and Lng.
//  2. Recomputes ZoneID via GetCurrentZoneID and stores it on the asset.
//  3. If ZoneID changed (including nil→zone or zone→nil), evaluates all
//     AlertRules and collects any that match.
//  4. Appends each matched rule as a TriggeredAlert to state.AlertLog.
//  5. Returns the newly triggered alerts (empty slice if no zone change
//     or no matching rules).
//
// Returns ErrNotFound if assetID is not in state.Assets.
func ProcessLocationUpdate(state *TrackerState, assetID string, lat, lng float64, timestamp string) ([]*TriggeredAlert, error) {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 4 — CRUD helpers  (~15 min)
// ---------------------------------------------------------------------------

// AddZone creates a zone, appends it to state, and returns it.
// Returns ErrAlreadyExists if zoneID is already in state.Zones.
func AddZone(state *TrackerState, zoneID, name string, minLat, maxLat, minLng, maxLng float64) (*Zone, error) {
	panic("not implemented")
}

// RemoveZone removes a zone from state.
// Any asset currently assigned to the removed zone has its ZoneID set to nil.
// Alert rules are NOT fired for this forced change.
// Returns ErrNotFound if zoneID is not in state.Zones.
func RemoveZone(state *TrackerState, zoneID string) error {
	panic("not implemented")
}

// AddAsset creates an asset with no initial location (Lat/Lng/ZoneID all nil),
// stores it in state, and returns it.
// Returns ErrAlreadyExists if assetID is already in state.Assets.
func AddAsset(state *TrackerState, assetID, name string) (*Asset, error) {
	panic("not implemented")
}

// AddAlertRule adds an alert rule to state and returns it.
// Returns ErrAlreadyExists if ruleID already exists in state.AlertRules.
func AddAlertRule(state *TrackerState, ruleID string, fromZoneID, toZoneID, assetID *string) (*AlertRule, error) {
	panic("not implemented")
}
