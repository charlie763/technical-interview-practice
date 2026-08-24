// Tests for Problem 10: Responder Dispatch Manager
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_10_dispatch_manager/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_10_dispatch_manager.go \
//	  -c go test -v .
package dispatch

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Shared timestamps
// ---------------------------------------------------------------------------

const (
	T0 = "2024-06-01T10:00:00"
	T1 = "2024-06-01T10:01:00"
	T2 = "2024-06-01T10:02:00"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newDM(t *testing.T) DispatchManager {
	t.Helper()
	return NewDispatchManager()
}

// seededDM returns a pre-seeded DispatchManager:
//
//	Responders:
//	  unit-12  (shooting + robbery, capacity=3): has inc-seed-1 open
//	  unit-14  (car-crash + fire,   capacity=2): no assignments
//	Incidents:
//	  inc-seed-1  shooting  sev=5  T0  -> assigned to unit-12
//	  inc-seed-2  shooting  sev=3  T1  -> unassigned
//	  inc-seed-3  car-crash sev=4  T2  -> unassigned
func seededDM(t *testing.T) DispatchManager {
	t.Helper()
	dm := NewDispatchManager()
	mustRegisterResponder(t, dm, "unit-12", "Alpha Team", []string{"shooting", "robbery"}, 3)
	mustRegisterResponder(t, dm, "unit-14", "Beta Team", []string{"car-crash", "fire"}, 2)
	mustAddIncident(t, dm, "inc-seed-1", "shooting", 5, T0)
	mustAddIncident(t, dm, "inc-seed-2", "shooting", 3, T1)
	mustAddIncident(t, dm, "inc-seed-3", "car-crash", 4, T2)
	mustOK(t, dm.AssignIncident("inc-seed-1", "unit-12"))
	return dm
}

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustRegisterResponder(t *testing.T, dm DispatchManager, id, name string, types []string, cap int) *Responder {
	t.Helper()
	r, err := dm.RegisterResponder(id, name, types, cap)
	if err != nil {
		t.Fatalf("RegisterResponder %q: %v", id, err)
	}
	return r
}

func mustAddIncident(t *testing.T, dm DispatchManager, id, incType string, sev int, ts string) *Incident {
	t.Helper()
	inc, err := dm.AddIncident(id, incType, sev, ts)
	if err != nil {
		t.Fatalf("AddIncident %q: %v", id, err)
	}
	return inc
}

// ---------------------------------------------------------------------------
// PART 1 — Registration and basic queries
// ---------------------------------------------------------------------------

func TestRegisterResponder(t *testing.T) {
	t.Run("stores_and_returns_responder", func(t *testing.T) {
		dm := newDM(t)
		r, err := dm.RegisterResponder("unit-reg", "Gamma", []string{"fire"}, 2)
		mustOK(t, err)
		if r.ID != "unit-reg" {
			t.Errorf("ID = %q, want 'unit-reg'", r.ID)
		}
		if r.Name != "Gamma" {
			t.Errorf("Name = %q, want 'Gamma'", r.Name)
		}
		if len(r.SubscribedTypes) != 1 || r.SubscribedTypes[0] != "fire" {
			t.Errorf("SubscribedTypes = %v, want ['fire']", r.SubscribedTypes)
		}
		if r.Capacity != 2 {
			t.Errorf("Capacity = %d, want 2", r.Capacity)
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		dm := seededDM(t)
		_, err := dm.RegisterResponder("unit-12", "Duplicate", []string{"fire"}, 1)
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestAddIncident(t *testing.T) {
	t.Run("stores_and_returns_incident", func(t *testing.T) {
		dm := newDM(t)
		inc, err := dm.AddIncident("inc-add", "fire", 2, T0)
		mustOK(t, err)
		if inc.ID != "inc-add" {
			t.Errorf("ID = %q, want 'inc-add'", inc.ID)
		}
		if inc.IncidentType != "fire" {
			t.Errorf("IncidentType = %q, want 'fire'", inc.IncidentType)
		}
		if inc.Severity != 2 {
			t.Errorf("Severity = %d, want 2", inc.Severity)
		}
		if inc.ResponderID != nil {
			t.Errorf("ResponderID = %v, want nil", inc.ResponderID)
		}
		if inc.Resolved {
			t.Error("Resolved should be false on creation")
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		dm := seededDM(t)
		_, err := dm.AddIncident("inc-seed-1", "fire", 1, T0)
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestGetIncidentsForResponder(t *testing.T) {
	t.Run("returns_subscribed_types_only", func(t *testing.T) {
		dm := seededDM(t)
		incidents, err := dm.GetIncidentsForResponder("unit-12")
		mustOK(t, err)
		for _, inc := range incidents {
			if inc.IncidentType != "shooting" && inc.IncidentType != "robbery" {
				t.Errorf("unexpected incident type %q for unit-12", inc.IncidentType)
			}
		}
	})

	t.Run("sorted_severity_desc_then_ts_asc", func(t *testing.T) {
		dm := seededDM(t)
		incidents, err := dm.GetIncidentsForResponder("unit-12")
		mustOK(t, err)
		if len(incidents) < 2 {
			t.Fatalf("expected at least 2 incidents, got %d", len(incidents))
		}
		if incidents[0].ID != "inc-seed-1" {
			t.Errorf("first incident = %q, want 'inc-seed-1' (higher severity)", incidents[0].ID)
		}
		if incidents[1].ID != "inc-seed-2" {
			t.Errorf("second incident = %q, want 'inc-seed-2'", incidents[1].ID)
		}
	})

	t.Run("same_severity_sorted_by_ts", func(t *testing.T) {
		dm := newDM(t)
		mustRegisterResponder(t, dm, "u-ts", "T", []string{"fire"}, 5)
		mustAddIncident(t, dm, "inc-ts-early", "fire", 3, T0)
		mustAddIncident(t, dm, "inc-ts-late", "fire", 3, T1)
		incidents, err := dm.GetIncidentsForResponder("u-ts")
		mustOK(t, err)
		if len(incidents) != 2 {
			t.Fatalf("expected 2 incidents, got %d", len(incidents))
		}
		if incidents[0].ID != "inc-ts-early" || incidents[1].ID != "inc-ts-late" {
			t.Errorf("sort order wrong: %v", []string{incidents[0].ID, incidents[1].ID})
		}
	})

	t.Run("missing_responder_returns_error", func(t *testing.T) {
		dm := seededDM(t)
		_, err := dm.GetIncidentsForResponder("ghost")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Assignment and resolution
// ---------------------------------------------------------------------------

func TestAssignIncident(t *testing.T) {
	t.Run("sets_responder_id", func(t *testing.T) {
		dm := seededDM(t)
		mustOK(t, dm.AssignIncident("inc-seed-2", "unit-12"))
		open, err := dm.GetOpenAssignments("unit-12")
		mustOK(t, err)
		found := false
		for _, inc := range open {
			if inc.ID == "inc-seed-2" {
				found = true
				break
			}
		}
		if !found {
			t.Error("inc-seed-2 not found in open assignments after assign")
		}
	})

	t.Run("missing_incident_returns_error", func(t *testing.T) {
		dm := seededDM(t)
		err := dm.AssignIncident("ghost-inc", "unit-12")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("missing_responder_returns_error", func(t *testing.T) {
		dm := seededDM(t)
		err := dm.AssignIncident("inc-seed-2", "ghost-unit")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("already_assigned_returns_error", func(t *testing.T) {
		dm := seededDM(t)
		// inc-seed-1 is already assigned to unit-12
		err := dm.AssignIncident("inc-seed-1", "unit-12")
		if !errors.Is(err, ErrAlreadyAssigned) {
			t.Errorf("expected ErrAlreadyAssigned, got %v", err)
		}
	})

	t.Run("at_capacity_returns_error", func(t *testing.T) {
		dm := newDM(t)
		mustRegisterResponder(t, dm, "cap-unit", "Cap", []string{"fire"}, 1)
		mustAddIncident(t, dm, "cap-inc-1", "fire", 1, T0)
		mustAddIncident(t, dm, "cap-inc-2", "fire", 1, T1)
		mustOK(t, dm.AssignIncident("cap-inc-1", "cap-unit"))
		err := dm.AssignIncident("cap-inc-2", "cap-unit")
		if !errors.Is(err, ErrAtCapacity) {
			t.Errorf("expected ErrAtCapacity, got %v", err)
		}
	})
}

func TestResolveIncident(t *testing.T) {
	t.Run("marks_resolved_and_removes_from_open", func(t *testing.T) {
		dm := seededDM(t)
		mustOK(t, dm.ResolveIncident("inc-seed-1"))
		open, err := dm.GetOpenAssignments("unit-12")
		mustOK(t, err)
		for _, inc := range open {
			if inc.ID == "inc-seed-1" {
				t.Error("inc-seed-1 still in open assignments after resolve")
			}
		}
	})

	t.Run("already_resolved_returns_error", func(t *testing.T) {
		dm := seededDM(t)
		mustOK(t, dm.ResolveIncident("inc-seed-1"))
		err := dm.ResolveIncident("inc-seed-1")
		if !errors.Is(err, ErrAlreadyResolved) {
			t.Errorf("expected ErrAlreadyResolved, got %v", err)
		}
	})

	t.Run("missing_incident_returns_error", func(t *testing.T) {
		dm := seededDM(t)
		err := dm.ResolveIncident("ghost")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("frees_capacity_for_next_assignment", func(t *testing.T) {
		dm := newDM(t)
		mustRegisterResponder(t, dm, "cap2-unit", "Cap2", []string{"fire"}, 1)
		mustAddIncident(t, dm, "cap2-inc-1", "fire", 1, T0)
		mustAddIncident(t, dm, "cap2-inc-2", "fire", 1, T1)
		mustOK(t, dm.AssignIncident("cap2-inc-1", "cap2-unit"))
		mustOK(t, dm.ResolveIncident("cap2-inc-1"))
		// Should no longer fail — capacity was freed
		mustOK(t, dm.AssignIncident("cap2-inc-2", "cap2-unit"))
	})
}

func TestGetOpenAssignments(t *testing.T) {
	t.Run("returns_open_assignments", func(t *testing.T) {
		dm := seededDM(t)
		open, err := dm.GetOpenAssignments("unit-12")
		mustOK(t, err)
		if len(open) != 1 {
			t.Fatalf("expected 1 open assignment, got %d", len(open))
		}
		if open[0].ID != "inc-seed-1" {
			t.Errorf("open assignment = %q, want 'inc-seed-1'", open[0].ID)
		}
	})

	t.Run("resolved_not_in_open", func(t *testing.T) {
		dm := seededDM(t)
		mustOK(t, dm.ResolveIncident("inc-seed-1"))
		open, err := dm.GetOpenAssignments("unit-12")
		mustOK(t, err)
		if len(open) != 0 {
			t.Errorf("expected 0 open assignments after resolve, got %d", len(open))
		}
	})

	t.Run("missing_responder_returns_error", func(t *testing.T) {
		dm := seededDM(t)
		_, err := dm.GetOpenAssignments("ghost")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("sorted_severity_desc_then_ts_asc", func(t *testing.T) {
		dm := newDM(t)
		mustRegisterResponder(t, dm, "u-sort", "S", []string{"fire"}, 5)
		mustAddIncident(t, dm, "inc-sort-low", "fire", 2, T0)
		mustAddIncident(t, dm, "inc-sort-high", "fire", 5, T1)
		mustOK(t, dm.AssignIncident("inc-sort-low", "u-sort"))
		mustOK(t, dm.AssignIncident("inc-sort-high", "u-sort"))
		open, err := dm.GetOpenAssignments("u-sort")
		mustOK(t, err)
		if len(open) < 1 || open[0].ID != "inc-sort-high" {
			t.Errorf("expected inc-sort-high first, got %q", open[0].ID)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Auto-assignment
// ---------------------------------------------------------------------------

func TestAutoAssign(t *testing.T) {
	t.Run("assigns_to_eligible_responder", func(t *testing.T) {
		dm := seededDM(t)
		// inc-seed-3 is car-crash; only unit-14 is subscribed
		responderID, err := dm.AutoAssign("inc-seed-3")
		mustOK(t, err)
		if responderID != "unit-14" {
			t.Errorf("AutoAssign = %q, want 'unit-14'", responderID)
		}
		open, err := dm.GetOpenAssignments("unit-14")
		mustOK(t, err)
		found := false
		for _, inc := range open {
			if inc.ID == "inc-seed-3" {
				found = true
				break
			}
		}
		if !found {
			t.Error("inc-seed-3 not in unit-14 open assignments after auto-assign")
		}
	})

	t.Run("missing_incident_returns_error", func(t *testing.T) {
		dm := seededDM(t)
		_, err := dm.AutoAssign("ghost")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("already_assigned_returns_error", func(t *testing.T) {
		dm := seededDM(t)
		// inc-seed-1 is already assigned to unit-12
		_, err := dm.AutoAssign("inc-seed-1")
		if !errors.Is(err, ErrAlreadyAssigned) {
			t.Errorf("expected ErrAlreadyAssigned, got %v", err)
		}
	})

	t.Run("no_eligible_responder_returns_error", func(t *testing.T) {
		dm := newDM(t)
		mustRegisterResponder(t, dm, "only-unit", "Only", []string{"shooting"}, 1)
		mustAddIncident(t, dm, "inc-no-sub", "fire", 1, T0)
		_, err := dm.AutoAssign("inc-no-sub")
		if !errors.Is(err, ErrNoEligibleResponder) {
			t.Errorf("expected ErrNoEligibleResponder, got %v", err)
		}
	})

	t.Run("full_capacity_responder_excluded", func(t *testing.T) {
		dm := newDM(t)
		mustRegisterResponder(t, dm, "full-unit", "Full", []string{"fire"}, 1)
		mustRegisterResponder(t, dm, "open-unit", "Open", []string{"fire"}, 2)
		mustAddIncident(t, dm, "inc-cap-fill", "fire", 1, T0)
		mustAddIncident(t, dm, "inc-cap-new", "fire", 1, T1)
		mustOK(t, dm.AssignIncident("inc-cap-fill", "full-unit"))
		responderID, err := dm.AutoAssign("inc-cap-new")
		mustOK(t, err)
		if responderID != "open-unit" {
			t.Errorf("AutoAssign = %q, want 'open-unit'", responderID)
		}
	})

	t.Run("picks_least_loaded_responder", func(t *testing.T) {
		dm := newDM(t)
		mustRegisterResponder(t, dm, "u-loaded", "Loaded", []string{"fire"}, 3)
		mustRegisterResponder(t, dm, "u-free", "Free", []string{"fire"}, 3)
		mustAddIncident(t, dm, "inc-load-seed", "fire", 1, T0)
		mustAddIncident(t, dm, "inc-load-new", "fire", 1, T1)
		mustOK(t, dm.AssignIncident("inc-load-seed", "u-loaded"))
		responderID, err := dm.AutoAssign("inc-load-new")
		mustOK(t, err)
		if responderID != "u-free" {
			t.Errorf("AutoAssign = %q, want 'u-free' (fewer open assignments)", responderID)
		}
	})

	t.Run("tiebreak_by_highest_capacity", func(t *testing.T) {
		dm := newDM(t)
		mustRegisterResponder(t, dm, "u-low-cap", "Low", []string{"fire"}, 1)
		mustRegisterResponder(t, dm, "u-high-cap", "High", []string{"fire"}, 5)
		mustAddIncident(t, dm, "inc-cap-tb", "fire", 1, T0)
		responderID, err := dm.AutoAssign("inc-cap-tb")
		mustOK(t, err)
		if responderID != "u-high-cap" {
			t.Errorf("AutoAssign = %q, want 'u-high-cap' (tiebreak by capacity)", responderID)
		}
	})

	t.Run("auto_assign_enforces_capacity_via_assign_incident", func(t *testing.T) {
		dm := newDM(t)
		mustRegisterResponder(t, dm, "u-delegate", "D", []string{"fire"}, 1)
		mustAddIncident(t, dm, "inc-delegate-1", "fire", 1, T0)
		mustAddIncident(t, dm, "inc-delegate-2", "fire", 1, T1)
		_, err := dm.AutoAssign("inc-delegate-1")
		mustOK(t, err)
		// Capacity now full; next auto-assign must fail
		_, err = dm.AutoAssign("inc-delegate-2")
		if !errors.Is(err, ErrNoEligibleResponder) {
			t.Errorf("expected ErrNoEligibleResponder after capacity full, got %v", err)
		}
	})
}

func TestGetDispatchSummary(t *testing.T) {
	t.Run("returns_all_responders", func(t *testing.T) {
		dm := seededDM(t)
		summary := dm.GetDispatchSummary()
		ids := make(map[string]bool)
		for _, s := range summary {
			ids[s.ResponderID] = true
		}
		if !ids["unit-12"] {
			t.Error("unit-12 not in summary")
		}
		if !ids["unit-14"] {
			t.Error("unit-14 not in summary")
		}
	})

	t.Run("sorted_by_responder_id", func(t *testing.T) {
		dm := seededDM(t)
		summary := dm.GetDispatchSummary()
		for i := 1; i < len(summary); i++ {
			if summary[i].ResponderID < summary[i-1].ResponderID {
				t.Errorf("summary not sorted: %q before %q", summary[i-1].ResponderID, summary[i].ResponderID)
			}
		}
	})

	t.Run("open_count_and_available_capacity", func(t *testing.T) {
		dm := seededDM(t)
		// unit-12 has 1 open assignment; capacity=3
		summary := dm.GetDispatchSummary()
		var u12 *DispatchSummary
		for _, s := range summary {
			if s.ResponderID == "unit-12" {
				u12 = s
				break
			}
		}
		if u12 == nil {
			t.Fatal("unit-12 not found in summary")
		}
		if u12.OpenCount != 1 {
			t.Errorf("OpenCount = %d, want 1", u12.OpenCount)
		}
		if u12.AvailableCapacity != 2 {
			t.Errorf("AvailableCapacity = %d, want 2", u12.AvailableCapacity)
		}
	})

	t.Run("zero_load_responder", func(t *testing.T) {
		dm := seededDM(t)
		// unit-14 has no assignments
		summary := dm.GetDispatchSummary()
		var u14 *DispatchSummary
		for _, s := range summary {
			if s.ResponderID == "unit-14" {
				u14 = s
				break
			}
		}
		if u14 == nil {
			t.Fatal("unit-14 not found in summary")
		}
		if u14.OpenCount != 0 {
			t.Errorf("OpenCount = %d, want 0", u14.OpenCount)
		}
		if u14.AvailableCapacity != 2 {
			t.Errorf("AvailableCapacity = %d, want 2", u14.AvailableCapacity)
		}
	})
}
