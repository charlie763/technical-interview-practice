// Tests for Problem 1: Geofence Alert Rule Engine
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_01_geofence_alert_engine/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_01_geofence_alert_engine.go \
//	  -c go test -v .
package geofence

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func pf(f float64) *float64 { return &f }
func ps(s string) *string   { return &s }

func makeSeededState(t *testing.T) *TrackerState {
	t.Helper()
	s := MakeTracker()
	if _, err := AddZone(s, "warehouse", "Warehouse A", 35.00, 35.10, -106.70, -106.60); err != nil {
		t.Fatalf("AddZone warehouse: %v", err)
	}
	if _, err := AddZone(s, "loading_dock", "Loading Dock", 35.10, 35.20, -106.70, -106.60); err != nil {
		t.Fatalf("AddZone loading_dock: %v", err)
	}
	if _, err := AddAsset(s, "forklift_1", "Forklift #1"); err != nil {
		t.Fatalf("AddAsset forklift_1: %v", err)
	}
	if _, err := AddAsset(s, "drone_1", "Drone #1"); err != nil {
		t.Fatalf("AddAsset drone_1: %v", err)
	}
	return s
}

func makeZone(minLat, maxLat, minLng, maxLng float64) *Zone {
	return &Zone{ID: "z1", Name: "Z", MinLat: minLat, MaxLat: maxLat, MinLng: minLng, MaxLng: maxLng}
}

func makeAsset(lat, lng *float64) *Asset {
	return &Asset{ID: "a1", Name: "A", Lat: lat, Lng: lng}
}

// ---------------------------------------------------------------------------
// PART 1 — IsInZone
// ---------------------------------------------------------------------------

func TestIsInZone(t *testing.T) {
	zone := makeZone(35.0, 35.1, -106.7, -106.6)

	tests := []struct {
		name string
		lat  *float64
		lng  *float64
		want bool
	}{
		{"inside", pf(35.05), pf(-106.65), true},
		{"on_min_corner", pf(35.0), pf(-106.7), true},
		{"on_max_corner", pf(35.1), pf(-106.6), true},
		{"outside_lat", pf(35.15), pf(-106.65), false},
		{"outside_lng", pf(35.05), pf(-106.5), false},
		{"nil_lat", nil, pf(-106.65), false},
		{"nil_lng", pf(35.05), nil, false},
		{"both_nil", nil, nil, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsInZone(makeAsset(tc.lat, tc.lng), zone)
			if got != tc.want {
				t.Errorf("IsInZone = %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PART 2 — GetCurrentZoneID
// ---------------------------------------------------------------------------

func TestGetCurrentZoneID(t *testing.T) {
	t.Run("asset_in_warehouse", func(t *testing.T) {
		s := makeSeededState(t)
		s.Assets["forklift_1"].Lat = pf(35.05)
		s.Assets["forklift_1"].Lng = pf(-106.65)
		got := GetCurrentZoneID(s, "forklift_1")
		if got == nil || *got != "warehouse" {
			t.Errorf("got %v, want 'warehouse'", got)
		}
	})

	t.Run("asset_in_loading_dock", func(t *testing.T) {
		s := makeSeededState(t)
		s.Assets["forklift_1"].Lat = pf(35.15)
		s.Assets["forklift_1"].Lng = pf(-106.65)
		got := GetCurrentZoneID(s, "forklift_1")
		if got == nil || *got != "loading_dock" {
			t.Errorf("got %v, want 'loading_dock'", got)
		}
	})

	t.Run("asset_outside_all_zones", func(t *testing.T) {
		s := makeSeededState(t)
		s.Assets["forklift_1"].Lat = pf(36.0)
		s.Assets["forklift_1"].Lng = pf(-106.65)
		if got := GetCurrentZoneID(s, "forklift_1"); got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("asset_no_location", func(t *testing.T) {
		s := makeSeededState(t)
		if got := GetCurrentZoneID(s, "forklift_1"); got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("unknown_asset_returns_nil", func(t *testing.T) {
		s := makeSeededState(t)
		if got := GetCurrentZoneID(s, "ghost"); got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — ProcessLocationUpdate
// ---------------------------------------------------------------------------

func TestProcessLocationUpdate(t *testing.T) {
	t.Run("no_alert_on_first_update_outside_zone", func(t *testing.T) {
		s := makeSeededState(t)
		alerts, err := ProcessLocationUpdate(s, "forklift_1", 36.0, -106.65, "t1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(alerts) != 0 {
			t.Errorf("got %d alerts, want 0", len(alerts))
		}
		if s.Assets["forklift_1"].Lat == nil || *s.Assets["forklift_1"].Lat != 36.0 {
			t.Error("lat not updated")
		}
		if s.Assets["forklift_1"].ZoneID != nil {
			t.Errorf("zone_id = %v, want nil", s.Assets["forklift_1"].ZoneID)
		}
	})

	t.Run("zone_entry_triggers_matching_rule", func(t *testing.T) {
		s := makeSeededState(t)
		AddAlertRule(s, "rule_entry", nil, ps("warehouse"), nil)
		alerts, err := ProcessLocationUpdate(s, "forklift_1", 35.05, -106.65, "t1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(alerts) != 1 {
			t.Fatalf("got %d alerts, want 1", len(alerts))
		}
		a := alerts[0]
		if a.RuleID != "rule_entry" {
			t.Errorf("RuleID = %q, want 'rule_entry'", a.RuleID)
		}
		if a.AssetID != "forklift_1" {
			t.Errorf("AssetID = %q, want 'forklift_1'", a.AssetID)
		}
		if a.FromZoneID != nil {
			t.Errorf("FromZoneID = %v, want nil", a.FromZoneID)
		}
		if a.ToZoneID == nil || *a.ToZoneID != "warehouse" {
			t.Errorf("ToZoneID = %v, want 'warehouse'", a.ToZoneID)
		}
		if a.Timestamp != "t1" {
			t.Errorf("Timestamp = %q, want 't1'", a.Timestamp)
		}
	})

	t.Run("zone_exit_triggers_rule", func(t *testing.T) {
		s := makeSeededState(t)
		s.Assets["forklift_1"].Lat = pf(35.05)
		s.Assets["forklift_1"].Lng = pf(-106.65)
		s.Assets["forklift_1"].ZoneID = ps("warehouse")
		AddAlertRule(s, "rule_exit", ps("warehouse"), nil, nil)
		alerts, _ := ProcessLocationUpdate(s, "forklift_1", 36.0, -106.65, "t2")
		if len(alerts) != 1 {
			t.Fatalf("got %d alerts, want 1", len(alerts))
		}
		if alerts[0].FromZoneID == nil || *alerts[0].FromZoneID != "warehouse" {
			t.Errorf("FromZoneID = %v, want 'warehouse'", alerts[0].FromZoneID)
		}
		if alerts[0].ToZoneID != nil {
			t.Errorf("ToZoneID = %v, want nil", alerts[0].ToZoneID)
		}
	})

	t.Run("zone_to_zone_transition", func(t *testing.T) {
		s := makeSeededState(t)
		s.Assets["forklift_1"].ZoneID = ps("warehouse")
		AddAlertRule(s, "rule_wh_to_dock", ps("warehouse"), ps("loading_dock"), nil)
		alerts, _ := ProcessLocationUpdate(s, "forklift_1", 35.15, -106.65, "t3")
		if len(alerts) != 1 {
			t.Fatalf("got %d alerts, want 1", len(alerts))
		}
		if *alerts[0].FromZoneID != "warehouse" || *alerts[0].ToZoneID != "loading_dock" {
			t.Errorf("unexpected from/to: %v / %v", alerts[0].FromZoneID, alerts[0].ToZoneID)
		}
	})

	t.Run("no_alert_when_zone_unchanged", func(t *testing.T) {
		s := makeSeededState(t)
		s.Assets["forklift_1"].Lat = pf(35.05)
		s.Assets["forklift_1"].Lng = pf(-106.65)
		s.Assets["forklift_1"].ZoneID = ps("warehouse")
		AddAlertRule(s, "rule_any", nil, nil, nil)
		alerts, _ := ProcessLocationUpdate(s, "forklift_1", 35.06, -106.65, "t4")
		if len(alerts) != 0 {
			t.Errorf("got %d alerts, want 0", len(alerts))
		}
	})

	t.Run("asset_specific_rule_ignores_other_assets", func(t *testing.T) {
		s := makeSeededState(t)
		AddAlertRule(s, "rule_drone_only", nil, ps("warehouse"), ps("drone_1"))
		alerts, _ := ProcessLocationUpdate(s, "forklift_1", 35.05, -106.65, "t5")
		if len(alerts) != 0 {
			t.Errorf("got %d alerts, want 0", len(alerts))
		}
	})

	t.Run("asset_specific_rule_fires_for_correct_asset", func(t *testing.T) {
		s := makeSeededState(t)
		AddAlertRule(s, "rule_forklift", nil, ps("warehouse"), ps("forklift_1"))
		alerts, _ := ProcessLocationUpdate(s, "forklift_1", 35.05, -106.65, "t6")
		if len(alerts) != 1 {
			t.Errorf("got %d alerts, want 1", len(alerts))
		}
	})

	t.Run("multiple_matching_rules_all_fire", func(t *testing.T) {
		s := makeSeededState(t)
		AddAlertRule(s, "rule_a", nil, ps("warehouse"), nil)
		AddAlertRule(s, "rule_b", nil, nil, nil)
		alerts, _ := ProcessLocationUpdate(s, "forklift_1", 35.05, -106.65, "t7")
		if len(alerts) != 2 {
			t.Errorf("got %d alerts, want 2", len(alerts))
		}
	})

	t.Run("alerts_appended_to_log", func(t *testing.T) {
		s := makeSeededState(t)
		AddAlertRule(s, "rule_1", nil, ps("warehouse"), nil)
		ProcessLocationUpdate(s, "forklift_1", 35.05, -106.65, "t8")
		if len(s.AlertLog) != 1 {
			t.Errorf("AlertLog length = %d, want 1", len(s.AlertLog))
		}
	})

	t.Run("unknown_asset_returns_error", func(t *testing.T) {
		s := makeSeededState(t)
		_, err := ProcessLocationUpdate(s, "ghost_asset", 35.05, -106.65, "t9")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("lat_lng_updated_even_when_no_zone_change", func(t *testing.T) {
		s := makeSeededState(t)
		s.Assets["forklift_1"].Lat = pf(35.05)
		s.Assets["forklift_1"].Lng = pf(-106.65)
		s.Assets["forklift_1"].ZoneID = ps("warehouse")
		ProcessLocationUpdate(s, "forklift_1", 35.06, -106.64, "t10")
		if s.Assets["forklift_1"].Lat == nil || *s.Assets["forklift_1"].Lat != 35.06 {
			t.Error("lat not updated")
		}
		if s.Assets["forklift_1"].Lng == nil || *s.Assets["forklift_1"].Lng != -106.64 {
			t.Error("lng not updated")
		}
	})
}

// ---------------------------------------------------------------------------
// PART 4 — CRUD helpers
// ---------------------------------------------------------------------------

func TestAddZone(t *testing.T) {
	t.Run("adds_zone", func(t *testing.T) {
		s := makeSeededState(t)
		z, err := AddZone(s, "yard", "Yard", 35.3, 35.4, -106.7, -106.6)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Zones["yard"] != z {
			t.Error("zone not stored in state")
		}
		if z.Name != "Yard" {
			t.Errorf("Name = %q, want 'Yard'", z.Name)
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		s := makeSeededState(t)
		_, err := AddZone(s, "warehouse", "Duplicate", 0, 1, 0, 1)
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestRemoveZone(t *testing.T) {
	t.Run("removes_zone", func(t *testing.T) {
		s := makeSeededState(t)
		if err := RemoveZone(s, "warehouse"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := s.Zones["warehouse"]; ok {
			t.Error("zone still in state")
		}
	})

	t.Run("clears_zone_id_on_assets", func(t *testing.T) {
		s := makeSeededState(t)
		s.Assets["forklift_1"].ZoneID = ps("warehouse")
		RemoveZone(s, "warehouse")
		if s.Assets["forklift_1"].ZoneID != nil {
			t.Errorf("ZoneID = %v, want nil", s.Assets["forklift_1"].ZoneID)
		}
	})

	t.Run("does_not_affect_assets_in_other_zones", func(t *testing.T) {
		s := makeSeededState(t)
		s.Assets["forklift_1"].ZoneID = ps("loading_dock")
		RemoveZone(s, "warehouse")
		if s.Assets["forklift_1"].ZoneID == nil || *s.Assets["forklift_1"].ZoneID != "loading_dock" {
			t.Errorf("ZoneID = %v, want 'loading_dock'", s.Assets["forklift_1"].ZoneID)
		}
	})

	t.Run("missing_zone_returns_error", func(t *testing.T) {
		s := makeSeededState(t)
		if err := RemoveZone(s, "nonexistent"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestAddAsset(t *testing.T) {
	t.Run("adds_asset", func(t *testing.T) {
		s := makeSeededState(t)
		a, err := AddAsset(s, "scanner_1", "Scanner #1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Assets["scanner_1"] != a {
			t.Error("asset not stored in state")
		}
		if a.Lat != nil || a.Lng != nil || a.ZoneID != nil {
			t.Error("new asset should have nil Lat/Lng/ZoneID")
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		s := makeSeededState(t)
		_, err := AddAsset(s, "forklift_1", "Duplicate")
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestAddAlertRule(t *testing.T) {
	t.Run("adds_rule", func(t *testing.T) {
		s := makeSeededState(t)
		r, err := AddAlertRule(s, "r1", ps("warehouse"), ps("loading_dock"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		found := false
		for _, rule := range s.AlertRules {
			if rule == r {
				found = true
				break
			}
		}
		if !found {
			t.Error("rule not found in AlertRules")
		}
		if r.ID != "r1" {
			t.Errorf("ID = %q, want 'r1'", r.ID)
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		s := makeSeededState(t)
		AddAlertRule(s, "r1", nil, nil, nil)
		_, err := AddAlertRule(s, "r1", nil, nil, nil)
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}
