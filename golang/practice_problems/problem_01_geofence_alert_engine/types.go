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
