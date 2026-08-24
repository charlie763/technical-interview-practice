// Tests for Problem 8: Patient Enrollment Pipeline
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_08_enrollment_pipeline/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_08_enrollment_pipeline.go \
//	  -c go test -v .
package enrollment

import (
	"errors"
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func freshPipeline(t *testing.T) EnrollmentPipeline {
	t.Helper()
	return NewEnrollmentPipeline()
}

// seededPipeline returns a pipeline with four patients at various stages:
//
//	seed_active:      referred(0) → screened(100) → enrolled(200) → active(300)
//	seed_graduated:   referred(0) → screened(50)  → enrolled(150) → active(250) → graduated(1000)
//	seed_ineligible:  referred(0) → screened(10)  → ineligible(20)
//	seed_referred:    referred(0)  [no further transitions]
func seededPipeline(t *testing.T) EnrollmentPipeline {
	t.Helper()
	p := NewEnrollmentPipeline()

	mustAddPatient(t, p, "seed_active", 0.0)
	mustTransition(t, p, "seed_active", "screened", 100.0)
	mustTransition(t, p, "seed_active", "enrolled", 200.0)
	mustTransition(t, p, "seed_active", "active", 300.0)

	mustAddPatient(t, p, "seed_graduated", 0.0)
	mustTransition(t, p, "seed_graduated", "screened", 50.0)
	mustTransition(t, p, "seed_graduated", "enrolled", 150.0)
	mustTransition(t, p, "seed_graduated", "active", 250.0)
	mustTransition(t, p, "seed_graduated", "graduated", 1000.0)

	mustAddPatient(t, p, "seed_ineligible", 0.0)
	mustTransition(t, p, "seed_ineligible", "screened", 10.0)
	mustTransition(t, p, "seed_ineligible", "ineligible", 20.0)

	mustAddPatient(t, p, "seed_referred", 0.0)

	return p
}

func mustAddPatient(t *testing.T, p EnrollmentPipeline, id string, ts float64) {
	t.Helper()
	if err := p.AddPatient(id, ts); err != nil {
		t.Fatalf("AddPatient(%q): %v", id, err)
	}
}

func mustTransition(t *testing.T, p EnrollmentPipeline, id, state string, ts float64) {
	t.Helper()
	if err := p.Transition(id, state, ts); err != nil {
		t.Fatalf("Transition(%q → %q): %v", id, state, err)
	}
}

func approxEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

// ---------------------------------------------------------------------------
// PART 1 — State tracking
// ---------------------------------------------------------------------------

func TestAddPatient(t *testing.T) {
	t.Run("new_patient_starts_in_referred", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "add_p1", 0.0)
		state, err := p.GetState("add_p1")
		if err != nil {
			t.Fatalf("GetState: %v", err)
		}
		if state != "referred" {
			t.Errorf("state = %q, want 'referred'", state)
		}
	})

	t.Run("duplicate_patient_id_returns_error", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "dup_p", 0.0)
		if err := p.AddPatient("dup_p", 10.0); !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestTransition(t *testing.T) {
	t.Run("valid_transition_changes_state", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "trans_p1", 0.0)
		mustTransition(t, p, "trans_p1", "screened", 100.0)
		state, _ := p.GetState("trans_p1")
		if state != "screened" {
			t.Errorf("state = %q, want 'screened'", state)
		}
	})

	t.Run("invalid_transition_returns_error", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "trans_p2", 0.0)
		// Cannot jump from referred straight to graduated
		if err := p.Transition("trans_p2", "graduated", 100.0); !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("skipping_states_returns_error", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "trans_p3", 0.0)
		mustTransition(t, p, "trans_p3", "screened", 10.0)
		// Cannot skip enrolled → jump from screened to active
		if err := p.Transition("trans_p3", "active", 20.0); !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("unknown_patient_returns_not_found", func(t *testing.T) {
		p := freshPipeline(t)
		if err := p.Transition("ghost", "screened", 100.0); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("transition_from_terminal_state_returns_error", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "trans_p4", 0.0)
		mustTransition(t, p, "trans_p4", "screened", 10.0)
		mustTransition(t, p, "trans_p4", "ineligible", 20.0)
		if err := p.Transition("trans_p4", "enrolled", 30.0); !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("both_branches_from_screened_are_valid", func(t *testing.T) {
		p := freshPipeline(t)

		mustAddPatient(t, p, "branch_p1", 0.0)
		mustTransition(t, p, "branch_p1", "screened", 10.0)
		mustTransition(t, p, "branch_p1", "enrolled", 20.0)
		state, _ := p.GetState("branch_p1")
		if state != "enrolled" {
			t.Errorf("branch_p1 state = %q, want 'enrolled'", state)
		}

		mustAddPatient(t, p, "branch_p2", 0.0)
		mustTransition(t, p, "branch_p2", "screened", 10.0)
		mustTransition(t, p, "branch_p2", "ineligible", 20.0)
		state, _ = p.GetState("branch_p2")
		if state != "ineligible" {
			t.Errorf("branch_p2 state = %q, want 'ineligible'", state)
		}
	})
}

func TestGetState(t *testing.T) {
	t.Run("returns_current_state_for_each_patient", func(t *testing.T) {
		p := seededPipeline(t)
		cases := []struct {
			id    string
			state string
		}{
			{"seed_active", "active"},
			{"seed_graduated", "graduated"},
			{"seed_ineligible", "ineligible"},
			{"seed_referred", "referred"},
		}
		for _, tc := range cases {
			t.Run(tc.id, func(t *testing.T) {
				got, err := p.GetState(tc.id)
				if err != nil {
					t.Fatalf("GetState: %v", err)
				}
				if got != tc.state {
					t.Errorf("got %q, want %q", got, tc.state)
				}
			})
		}
	})

	t.Run("unknown_patient_returns_not_found", func(t *testing.T) {
		p := freshPipeline(t)
		if _, err := p.GetState("nobody"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetPatientsInState(t *testing.T) {
	t.Run("returns_patients_in_state", func(t *testing.T) {
		p := seededPipeline(t)
		active := p.GetPatientsInState("active")
		found := false
		for _, id := range active {
			if id == "seed_active" {
				found = true
			}
			if id == "seed_graduated" {
				t.Error("seed_graduated should not be in active state")
			}
		}
		if !found {
			t.Error("seed_active not found in active state")
		}
	})

	t.Run("result_is_sorted", func(t *testing.T) {
		p := seededPipeline(t)
		result := p.GetPatientsInState("referred")
		for i := 1; i < len(result); i++ {
			if result[i] < result[i-1] {
				t.Errorf("GetPatientsInState not sorted: %v", result)
				break
			}
		}
	})

	t.Run("returns_empty_for_unpopulated_state", func(t *testing.T) {
		p := seededPipeline(t)
		if got := p.GetPatientsInState("withdrawn"); len(got) != 0 {
			t.Errorf("expected empty slice, got %v", got)
		}
	})

	t.Run("patient_absent_after_leaving_state", func(t *testing.T) {
		p := seededPipeline(t)
		// seed_ineligible passed through screened but is no longer there
		screened := p.GetPatientsInState("screened")
		for _, id := range screened {
			if id == "seed_ineligible" {
				t.Error("seed_ineligible should not be in screened state")
			}
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Duration and conversion metrics
// ---------------------------------------------------------------------------

func TestTimeInState(t *testing.T) {
	t.Run("completed_state_returns_exact_duration", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "dur_p1", 0.0)
		mustTransition(t, p, "dur_p1", "screened", 1000.0)
		mustTransition(t, p, "dur_p1", "enrolled", 4000.0)
		// Spent exactly 3000 s in screened; as_of is ignored for completed states
		got := p.TimeInState("dur_p1", "screened", 99999.0)
		if !approxEqual(got, 3000.0, 0.001) {
			t.Errorf("TimeInState = %v, want 3000.0", got)
		}
	})

	t.Run("current_state_counts_up_to_as_of", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "dur_p2", 0.0)
		mustTransition(t, p, "dur_p2", "screened", 1000.0)
		got := p.TimeInState("dur_p2", "screened", 4000.0)
		if !approxEqual(got, 3000.0, 0.001) {
			t.Errorf("TimeInState = %v, want 3000.0", got)
		}
	})

	t.Run("returns_zero_for_state_never_visited", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "dur_p3", 0.0)
		got := p.TimeInState("dur_p3", "enrolled", 99999.0)
		if got != 0.0 {
			t.Errorf("TimeInState = %v, want 0.0", got)
		}
	})

	t.Run("initial_referred_state_timed_from_add_patient_timestamp", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "dur_p4", 500.0)
		mustTransition(t, p, "dur_p4", "screened", 1500.0)
		// Spent 1000 s in referred (1500 - 500)
		got := p.TimeInState("dur_p4", "referred", 99999.0)
		if !approxEqual(got, 1000.0, 0.001) {
			t.Errorf("TimeInState = %v, want 1000.0", got)
		}
	})
}

func TestConversionRate(t *testing.T) {
	t.Run("fifty_percent_conversion", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "conv_p1", 0.0)
		mustTransition(t, p, "conv_p1", "screened", 10.0)
		mustTransition(t, p, "conv_p1", "enrolled", 20.0)

		mustAddPatient(t, p, "conv_p2", 0.0)
		mustTransition(t, p, "conv_p2", "screened", 10.0)
		mustTransition(t, p, "conv_p2", "ineligible", 20.0)

		got := p.ConversionRate("screened", "enrolled")
		if !approxEqual(got, 0.5, 0.001) {
			t.Errorf("ConversionRate = %v, want 0.5", got)
		}
	})

	t.Run("excludes_patients_still_in_from_state", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "conv_p3", 0.0)
		mustTransition(t, p, "conv_p3", "screened", 10.0)
		// conv_p3 is still in screened — must not be counted

		mustAddPatient(t, p, "conv_p4", 0.0)
		mustTransition(t, p, "conv_p4", "screened", 10.0)
		mustTransition(t, p, "conv_p4", "enrolled", 20.0)

		// Only conv_p4 has exited; they enrolled → rate is 1.0
		got := p.ConversionRate("screened", "enrolled")
		if !approxEqual(got, 1.0, 0.001) {
			t.Errorf("ConversionRate = %v, want 1.0", got)
		}
	})

	t.Run("returns_zero_when_no_one_has_exited_from_state", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "conv_p5", 0.0)
		// conv_p5 is still in referred
		got := p.ConversionRate("referred", "screened")
		if got != 0.0 {
			t.Errorf("ConversionRate = %v, want 0.0", got)
		}
	})

	t.Run("hundred_percent_when_all_converted", func(t *testing.T) {
		p := freshPipeline(t)
		for _, pid := range []string{"conv_all_1", "conv_all_2"} {
			mustAddPatient(t, p, pid, 0.0)
			mustTransition(t, p, pid, "screened", 10.0)
			mustTransition(t, p, pid, "enrolled", 20.0)
		}
		got := p.ConversionRate("screened", "enrolled")
		if !approxEqual(got, 1.0, 0.001) {
			t.Errorf("ConversionRate = %v, want 1.0", got)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — SLA monitoring
// ---------------------------------------------------------------------------

func TestPatientsOverdue(t *testing.T) {
	t.Run("returns_patients_exceeding_threshold", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "over_p1", 0.0)
		mustTransition(t, p, "over_p1", "screened", 0.0)
		mustTransition(t, p, "over_p1", "enrolled", 0.0)
		mustTransition(t, p, "over_p1", "active", 0.0) // 10000 s in active

		mustAddPatient(t, p, "over_p2", 0.0)
		mustTransition(t, p, "over_p2", "screened", 0.0)
		mustTransition(t, p, "over_p2", "enrolled", 0.0)
		mustTransition(t, p, "over_p2", "active", 5000.0) // 5000 s in active

		overdue := p.PatientsOverdue("active", 6000.0, 10000.0)
		foundP1, foundP2 := false, false
		for _, id := range overdue {
			if id == "over_p1" {
				foundP1 = true
			}
			if id == "over_p2" {
				foundP2 = true
			}
		}
		if !foundP1 {
			t.Error("expected over_p1 to be overdue")
		}
		if foundP2 {
			t.Error("expected over_p2 to NOT be overdue (5000 s < 6000 s threshold)")
		}
	})

	t.Run("sorted_by_duration_descending", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "sort_p1", 0.0)
		mustTransition(t, p, "sort_p1", "screened", 0.0)
		mustTransition(t, p, "sort_p1", "enrolled", 0.0)
		mustTransition(t, p, "sort_p1", "active", 0.0) // 10000 s

		mustAddPatient(t, p, "sort_p2", 0.0)
		mustTransition(t, p, "sort_p2", "screened", 0.0)
		mustTransition(t, p, "sort_p2", "enrolled", 0.0)
		mustTransition(t, p, "sort_p2", "active", 3000.0) // 7000 s

		overdue := p.PatientsOverdue("active", 5000.0, 10000.0)
		if len(overdue) != 2 {
			t.Fatalf("expected 2 overdue patients, got %d: %v", len(overdue), overdue)
		}
		if overdue[0] != "sort_p1" || overdue[1] != "sort_p2" {
			t.Errorf("overdue = %v, want ['sort_p1', 'sort_p2']", overdue)
		}
	})

	t.Run("excludes_patients_not_currently_in_state", func(t *testing.T) {
		p := seededPipeline(t)
		// seed_graduated has left active — must not appear in active overdue list
		overdue := p.PatientsOverdue("active", 0.0, 2000.0)
		for _, id := range overdue {
			if id == "seed_graduated" {
				t.Error("seed_graduated should not appear (has already left active)")
			}
		}
	})

	t.Run("returns_empty_when_no_one_is_overdue", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "noover_p", 0.0)
		mustTransition(t, p, "noover_p", "screened", 0.0)
		mustTransition(t, p, "noover_p", "enrolled", 0.0)
		mustTransition(t, p, "noover_p", "active", 9900.0) // only 100 s in active
		if overdue := p.PatientsOverdue("active", 500.0, 10000.0); len(overdue) != 0 {
			t.Errorf("expected no overdue patients, got %v", overdue)
		}
	})
}

func TestAverageTimeInState(t *testing.T) {
	t.Run("average_over_all_exited_patients", func(t *testing.T) {
		// avg_pa: 1000 s in screened; avg_pb: 3000 s in screened → average = 2000
		p := freshPipeline(t)
		mustAddPatient(t, p, "avg_pa", 0.0)
		mustTransition(t, p, "avg_pa", "screened", 0.0)
		mustTransition(t, p, "avg_pa", "enrolled", 1000.0)

		mustAddPatient(t, p, "avg_pb", 0.0)
		mustTransition(t, p, "avg_pb", "screened", 0.0)
		mustTransition(t, p, "avg_pb", "enrolled", 3000.0)

		got := p.AverageTimeInState("screened", 99999.0)
		if !approxEqual(got, 2000.0, 0.001) {
			t.Errorf("AverageTimeInState = %v, want 2000.0", got)
		}
	})

	t.Run("excludes_patients_still_in_state", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "avg_pc", 0.0)
		mustTransition(t, p, "avg_pc", "screened", 0.0)
		mustTransition(t, p, "avg_pc", "enrolled", 1000.0) // exited: 1000 s

		mustAddPatient(t, p, "avg_pd", 0.0)
		mustTransition(t, p, "avg_pd", "screened", 0.0)
		// avg_pd is still in screened at as_of=5000 — must be excluded

		got := p.AverageTimeInState("screened", 5000.0)
		if !approxEqual(got, 1000.0, 0.001) {
			t.Errorf("AverageTimeInState = %v, want 1000.0", got)
		}
	})

	t.Run("returns_zero_when_no_one_has_exited", func(t *testing.T) {
		p := freshPipeline(t)
		mustAddPatient(t, p, "avg_pe", 0.0)
		mustTransition(t, p, "avg_pe", "screened", 0.0)
		// avg_pe is still in screened
		got := p.AverageTimeInState("screened", 5000.0)
		if got != 0.0 {
			t.Errorf("AverageTimeInState = %v, want 0.0", got)
		}
	})
}
