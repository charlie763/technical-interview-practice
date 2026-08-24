// Tests for Problem 7: Care Team Assignment Manager
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_07_care_team_assignments/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_07_care_team_assignments.go \
//	  -c go test -v .
package careteam

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func pf64(f float64) *float64 { return &f }

// freshMgr returns a new empty CareTeamManager.
func freshMgr(t *testing.T) CareTeamManager {
	t.Helper()
	return NewCareTeamManager()
}

// seededMgr returns a pre-populated manager:
//
//	coach_alpha  (coach,     max=3): patients seed_1, seed_2
//	coach_beta   (coach,     max=1): no patients
//	dr_omega     (physician, max=100): patient seed_1
func seededMgr(t *testing.T) CareTeamManager {
	t.Helper()
	m := NewCareTeamManager()
	m.AddMember("coach_alpha", "coach", 3)
	m.AddMember("coach_beta", "coach", 1)
	m.AddMember("dr_omega", "physician", 100)
	if err := m.Assign("patient_seed_1", "coach_alpha", 1000.0); err != nil {
		t.Fatalf("seededMgr: Assign seed_1→coach_alpha: %v", err)
	}
	if err := m.Assign("patient_seed_1", "dr_omega", 1000.0); err != nil {
		t.Fatalf("seededMgr: Assign seed_1→dr_omega: %v", err)
	}
	if err := m.Assign("patient_seed_2", "coach_alpha", 2000.0); err != nil {
		t.Fatalf("seededMgr: Assign seed_2→coach_alpha: %v", err)
	}
	return m
}

// ---------------------------------------------------------------------------
// PART 1 — Basic assignment and lookup
// ---------------------------------------------------------------------------

func TestAddMember(t *testing.T) {
	t.Run("registered_member_can_receive_assignment", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("member_reg", "coach", 5)
		if err := m.Assign("patient_reg", "member_reg", 0.0); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("unregistered_member_returns_not_found", func(t *testing.T) {
		m := freshMgr(t)
		if err := m.Assign("patient_unreg", "ghost_member", 0.0); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestAssign(t *testing.T) {
	t.Run("basic_assignment_recorded", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("coach_basic", "coach", 5)
		if err := m.Assign("patient_basic", "coach_basic", 100.0); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := m.GetAssignment("patient_basic", "coach")
		if got != "coach_basic" {
			t.Errorf("GetAssignment = %q, want 'coach_basic'", got)
		}
	})

	t.Run("reassignment_replaces_current_member", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("coach_orig", "coach", 5)
		m.AddMember("coach_new", "coach", 5)
		m.Assign("patient_reassign", "coach_orig", 100.0)
		m.Assign("patient_reassign", "coach_new", 200.0)
		got := m.GetAssignment("patient_reassign", "coach")
		if got != "coach_new" {
			t.Errorf("GetAssignment = %q, want 'coach_new'", got)
		}
	})

	t.Run("reassignment_removes_patient_from_old_member", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("coach_from", "coach", 5)
		m.AddMember("coach_to", "coach", 5)
		m.Assign("patient_move", "coach_from", 100.0)
		m.Assign("patient_move", "coach_to", 200.0)
		patients, err := m.GetPatients("coach_from")
		if err != nil {
			t.Fatalf("GetPatients: %v", err)
		}
		for _, p := range patients {
			if p == "patient_move" {
				t.Error("patient_move should have been removed from coach_from")
			}
		}
	})

	t.Run("assignments_across_roles_are_independent", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("coach_ind", "coach", 5)
		m.AddMember("dr_ind", "physician", 5)
		m.Assign("patient_ind", "coach_ind", 100.0)
		m.Assign("patient_ind", "dr_ind", 100.0)
		if got := m.GetAssignment("patient_ind", "coach"); got != "coach_ind" {
			t.Errorf("coach assignment = %q, want 'coach_ind'", got)
		}
		if got := m.GetAssignment("patient_ind", "physician"); got != "dr_ind" {
			t.Errorf("physician assignment = %q, want 'dr_ind'", got)
		}
	})
}

func TestGetAssignment(t *testing.T) {
	t.Run("returns_current_member", func(t *testing.T) {
		m := seededMgr(t)
		if got := m.GetAssignment("patient_seed_1", "coach"); got != "coach_alpha" {
			t.Errorf("got %q, want 'coach_alpha'", got)
		}
	})

	t.Run("returns_empty_for_unassigned_role", func(t *testing.T) {
		m := seededMgr(t)
		if got := m.GetAssignment("patient_seed_1", "dietitian"); got != "" {
			t.Errorf("got %q, want ''", got)
		}
	})

	t.Run("returns_empty_for_unknown_patient", func(t *testing.T) {
		m := freshMgr(t)
		if got := m.GetAssignment("no_such_patient", "coach"); got != "" {
			t.Errorf("got %q, want ''", got)
		}
	})
}

func TestGetPatients(t *testing.T) {
	t.Run("returns_all_assigned_patients", func(t *testing.T) {
		m := seededMgr(t)
		patients, err := m.GetPatients("coach_alpha")
		if err != nil {
			t.Fatalf("GetPatients: %v", err)
		}
		found1, found2 := false, false
		for _, p := range patients {
			if p == "patient_seed_1" {
				found1 = true
			}
			if p == "patient_seed_2" {
				found2 = true
			}
		}
		if !found1 || !found2 {
			t.Errorf("expected both seed patients, got %v", patients)
		}
	})

	t.Run("result_is_sorted", func(t *testing.T) {
		m := seededMgr(t)
		patients, _ := m.GetPatients("coach_alpha")
		for i := 1; i < len(patients); i++ {
			if patients[i] < patients[i-1] {
				t.Errorf("patients not sorted: %v", patients)
				break
			}
		}
	})

	t.Run("returns_empty_for_member_with_no_patients", func(t *testing.T) {
		m := seededMgr(t)
		patients, err := m.GetPatients("coach_beta")
		if err != nil {
			t.Fatalf("GetPatients: %v", err)
		}
		if len(patients) != 0 {
			t.Errorf("expected empty slice, got %v", patients)
		}
	})

	t.Run("unregistered_member_returns_not_found", func(t *testing.T) {
		m := freshMgr(t)
		if _, err := m.GetPatients("nobody"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Capacity enforcement
// ---------------------------------------------------------------------------

func TestCapacityEnforcement(t *testing.T) {
	t.Run("raises_capacity_error_when_member_is_full", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("coach_full", "coach", 1)
		m.Assign("patient_cap_1", "coach_full", 100.0)
		if err := m.Assign("patient_cap_2", "coach_full", 200.0); !errors.Is(err, ErrCapacity) {
			t.Errorf("expected ErrCapacity, got %v", err)
		}
	})

	t.Run("reassigning_existing_patient_to_same_member_does_not_error", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("coach_same", "coach", 1)
		m.Assign("patient_same", "coach_same", 100.0)
		// Member is "full" but the patient is already theirs — must not raise
		if err := m.Assign("patient_same", "coach_same", 200.0); err != nil {
			t.Errorf("unexpected error on reassign to same member: %v", err)
		}
	})

	t.Run("reassigning_patient_away_frees_capacity", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("coach_donor", "coach", 1)
		m.AddMember("coach_recv", "coach", 5)
		m.Assign("patient_freed", "coach_donor", 100.0)
		// Move the patient away — coach_donor now has a free slot
		m.Assign("patient_freed", "coach_recv", 200.0)
		// A new patient should now fit on coach_donor
		if err := m.Assign("patient_new", "coach_donor", 300.0); err != nil {
			t.Errorf("unexpected error after freeing capacity: %v", err)
		}
		if got := m.GetAssignment("patient_new", "coach"); got != "coach_donor" {
			t.Errorf("GetAssignment = %q, want 'coach_donor'", got)
		}
	})
}

func TestAvailableMembers(t *testing.T) {
	t.Run("returns_members_with_open_capacity", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("avail_coach_open", "coach", 2)
		m.AddMember("avail_coach_full", "coach", 1)
		m.Assign("avail_patient_1", "avail_coach_full", 100.0)
		available := m.AvailableMembers("coach")
		foundOpen, foundFull := false, false
		for _, id := range available {
			if id == "avail_coach_open" {
				foundOpen = true
			}
			if id == "avail_coach_full" {
				foundFull = true
			}
		}
		if !foundOpen {
			t.Error("expected avail_coach_open in available members")
		}
		if foundFull {
			t.Error("expected avail_coach_full to be excluded (full)")
		}
	})

	t.Run("full_member_is_excluded", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("only_coach", "coach", 1)
		m.Assign("only_patient", "only_coach", 100.0)
		for _, id := range m.AvailableMembers("coach") {
			if id == "only_coach" {
				t.Error("full member should not appear in available members")
			}
		}
	})

	t.Run("result_is_sorted", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("sort_coach_z", "coach", 5)
		m.AddMember("sort_coach_a", "coach", 5)
		result := m.AvailableMembers("coach")
		for i := 1; i < len(result); i++ {
			if result[i] < result[i-1] {
				t.Errorf("AvailableMembers not sorted: %v", result)
				break
			}
		}
	})

	t.Run("empty_when_no_members_with_role", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("solo_physician", "physician", 100)
		if available := m.AvailableMembers("coach"); len(available) != 0 {
			t.Errorf("expected empty slice, got %v", available)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Assignment history
// ---------------------------------------------------------------------------

func TestGetHistory(t *testing.T) {
	t.Run("single_assignment_has_open_end", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("hist_coach_1", "coach", 5)
		m.Assign("hist_patient_1", "hist_coach_1", 1000.0)
		hist := m.GetHistory("hist_patient_1", "coach")
		if len(hist) != 1 {
			t.Fatalf("len(history) = %d, want 1", len(hist))
		}
		if hist[0].MemberID != "hist_coach_1" {
			t.Errorf("MemberID = %q, want 'hist_coach_1'", hist[0].MemberID)
		}
		if hist[0].AssignedAt != 1000.0 {
			t.Errorf("AssignedAt = %v, want 1000.0", hist[0].AssignedAt)
		}
		if hist[0].UnassignedAt != nil {
			t.Errorf("UnassignedAt = %v, want nil", hist[0].UnassignedAt)
		}
	})

	t.Run("reassignment_closes_previous_entry", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("hist_coach_a", "coach", 5)
		m.AddMember("hist_coach_b", "coach", 5)
		m.Assign("hist_patient_2", "hist_coach_a", 1000.0)
		m.Assign("hist_patient_2", "hist_coach_b", 3000.0)
		hist := m.GetHistory("hist_patient_2", "coach")
		if len(hist) != 2 {
			t.Fatalf("len(history) = %d, want 2", len(hist))
		}
		// First entry: coach_a, 1000→3000
		if hist[0].MemberID != "hist_coach_a" {
			t.Errorf("hist[0].MemberID = %q, want 'hist_coach_a'", hist[0].MemberID)
		}
		if hist[0].AssignedAt != 1000.0 {
			t.Errorf("hist[0].AssignedAt = %v, want 1000.0", hist[0].AssignedAt)
		}
		if hist[0].UnassignedAt == nil || *hist[0].UnassignedAt != 3000.0 {
			t.Errorf("hist[0].UnassignedAt = %v, want &3000.0", hist[0].UnassignedAt)
		}
		// Second entry: coach_b, 3000→nil
		if hist[1].MemberID != "hist_coach_b" {
			t.Errorf("hist[1].MemberID = %q, want 'hist_coach_b'", hist[1].MemberID)
		}
		if hist[1].UnassignedAt != nil {
			t.Errorf("hist[1].UnassignedAt = %v, want nil", hist[1].UnassignedAt)
		}
	})

	t.Run("multiple_reassignments_sorted_chronologically", func(t *testing.T) {
		m := freshMgr(t)
		for _, name := range []string{"hist_cx", "hist_cy", "hist_cz"} {
			m.AddMember(name, "coach", 5)
		}
		m.Assign("hist_patient_3", "hist_cx", 100.0)
		m.Assign("hist_patient_3", "hist_cy", 200.0)
		m.Assign("hist_patient_3", "hist_cz", 300.0)
		hist := m.GetHistory("hist_patient_3", "coach")
		if len(hist) != 3 {
			t.Fatalf("len(history) = %d, want 3", len(hist))
		}
		wantTimes := []float64{100.0, 200.0, 300.0}
		for i, r := range hist {
			if r.AssignedAt != wantTimes[i] {
				t.Errorf("hist[%d].AssignedAt = %v, want %v", i, r.AssignedAt, wantTimes[i])
			}
		}
		if hist[2].UnassignedAt != nil {
			t.Errorf("last entry UnassignedAt = %v, want nil", hist[2].UnassignedAt)
		}
	})

	t.Run("returns_empty_for_unassigned_role", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("hist_coach_only", "coach", 5)
		m.Assign("hist_patient_4", "hist_coach_only", 100.0)
		if hist := m.GetHistory("hist_patient_4", "physician"); len(hist) != 0 {
			t.Errorf("expected empty, got %v", hist)
		}
	})

	t.Run("returns_empty_for_unknown_patient", func(t *testing.T) {
		m := freshMgr(t)
		if hist := m.GetHistory("hist_nobody", "coach"); len(hist) != 0 {
			t.Errorf("expected empty, got %v", hist)
		}
	})
}

func TestGetAssignmentAt(t *testing.T) {
	t.Run("returns_member_during_active_window", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("at_coach_1", "coach", 5)
		m.Assign("at_patient_1", "at_coach_1", 1000.0)
		if got := m.GetAssignmentAt("at_patient_1", "coach", 2000.0); got != "at_coach_1" {
			t.Errorf("got %q, want 'at_coach_1'", got)
		}
	})

	t.Run("returns_empty_before_first_assignment", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("at_coach_2", "coach", 5)
		m.Assign("at_patient_2", "at_coach_2", 1000.0)
		if got := m.GetAssignmentAt("at_patient_2", "coach", 500.0); got != "" {
			t.Errorf("got %q, want '' (before any assignment)", got)
		}
	})

	t.Run("returns_original_member_before_reassignment", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("at_coach_old", "coach", 5)
		m.AddMember("at_coach_new", "coach", 5)
		m.Assign("at_patient_3", "at_coach_old", 1000.0)
		m.Assign("at_patient_3", "at_coach_new", 3000.0)
		if got := m.GetAssignmentAt("at_patient_3", "coach", 2000.0); got != "at_coach_old" {
			t.Errorf("got %q, want 'at_coach_old'", got)
		}
	})

	t.Run("returns_new_member_after_reassignment", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("at_coach_prev", "coach", 5)
		m.AddMember("at_coach_curr", "coach", 5)
		m.Assign("at_patient_4", "at_coach_prev", 1000.0)
		m.Assign("at_patient_4", "at_coach_curr", 3000.0)
		if got := m.GetAssignmentAt("at_patient_4", "coach", 4000.0); got != "at_coach_curr" {
			t.Errorf("got %q, want 'at_coach_curr'", got)
		}
	})

	t.Run("returns_empty_for_unassigned_role", func(t *testing.T) {
		m := freshMgr(t)
		m.AddMember("at_coach_role", "coach", 5)
		m.Assign("at_patient_5", "at_coach_role", 1000.0)
		if got := m.GetAssignmentAt("at_patient_5", "physician", 2000.0); got != "" {
			t.Errorf("got %q, want '' (role never assigned)", got)
		}
	})
}
