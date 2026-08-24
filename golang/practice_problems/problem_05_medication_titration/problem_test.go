// Tests for Problem 5: Medication Titration Tracker
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_05_medication_titration/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_05_medication_titration.go \
//	  -c go test -v .
package titration

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Test data — mirrors the Python test fixtures
// ---------------------------------------------------------------------------

// Maria: successfully de-escalated off metformin, still on low-dose glipizide
var mariaEvents = []TitrationEvent{
	{"maria", "metformin", "start", 500.0, "2023-06-01"},
	{"maria", "metformin", "increase", 1000.0, "2023-08-01"},
	{"maria", "metformin", "decrease", 500.0, "2023-11-01"},
	{"maria", "metformin", "stop", 0.0, "2024-02-01"},
	{"maria", "glipizide", "start", 5.0, "2023-06-01"},
	{"maria", "glipizide", "decrease", 2.5, "2024-01-15"},
}

// James: still on two active medications, one recently decreased
var jamesEvents = []TitrationEvent{
	{"james", "metformin", "start", 500.0, "2023-09-01"},
	{"james", "metformin", "increase", 1000.0, "2023-12-01"},
	{"james", "insulin_glargine", "start", 10.0, "2023-09-01"},
	{"james", "insulin_glargine", "increase", 15.0, "2024-01-01"},
	{"james", "insulin_glargine", "decrease", 10.0, "2024-03-01"},
}

// Susan: completely off all medications
var susanEvents = []TitrationEvent{
	{"susan", "metformin", "start", 500.0, "2023-03-01"},
	{"susan", "metformin", "increase", 750.0, "2023-05-01"},
	{"susan", "metformin", "stop", 0.0, "2023-10-01"},
	{"susan", "glipizide", "start", 5.0, "2023-03-01"},
	{"susan", "glipizide", "stop", 0.0, "2023-09-01"},
}

func allEvents() []TitrationEvent {
	var all []TitrationEvent
	all = append(all, mariaEvents...)
	all = append(all, jamesEvents...)
	all = append(all, susanEvents...)
	return all
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newTracker(t *testing.T) TitrationTracker {
	t.Helper()
	return NewTitrationTracker(allEvents())
}

func freshTracker(t *testing.T) TitrationTracker {
	t.Helper()
	return NewTitrationTracker(nil)
}

// findMed returns the Medication with the given name from the slice, or nil.
func findMed(meds []Medication, name string) *Medication {
	for i := range meds {
		if meds[i].Name == name {
			return &meds[i]
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// PART 1 — CurrentMedications + GetMedicationHistory
// ---------------------------------------------------------------------------

func TestCurrentMedications(t *testing.T) {
	tr := newTracker(t)

	t.Run("active_medications_returned", func(t *testing.T) {
		meds := tr.CurrentMedications("maria")
		if findMed(meds, "glipizide") == nil {
			t.Error("expected glipizide to be active for maria")
		}
	})

	t.Run("stopped_medication_excluded", func(t *testing.T) {
		meds := tr.CurrentMedications("maria")
		if findMed(meds, "metformin") != nil {
			t.Error("metformin should not be active for maria (was stopped)")
		}
	})

	t.Run("all_stopped_returns_empty", func(t *testing.T) {
		meds := tr.CurrentMedications("susan")
		if len(meds) != 0 {
			t.Errorf("expected empty slice for susan, got %v", meds)
		}
	})

	t.Run("unknown_patient_returns_empty", func(t *testing.T) {
		meds := tr.CurrentMedications("nobody")
		if len(meds) != 0 {
			t.Errorf("expected empty slice for unknown patient, got %v", meds)
		}
	})

	t.Run("medication_fields_correct", func(t *testing.T) {
		meds := tr.CurrentMedications("maria")
		glip := findMed(meds, "glipizide")
		if glip == nil {
			t.Fatal("glipizide not found for maria")
		}
		if glip.CurrentDose != 2.5 {
			t.Errorf("CurrentDose = %v, want 2.5", glip.CurrentDose)
		}
		if glip.LastChanged != "2024-01-15" {
			t.Errorf("LastChanged = %q, want '2024-01-15'", glip.LastChanged)
		}
		if glip.TotalChanges != 2 { // start + decrease
			t.Errorf("TotalChanges = %d, want 2", glip.TotalChanges)
		}
	})

	t.Run("multiple_active_medications", func(t *testing.T) {
		meds := tr.CurrentMedications("james")
		if findMed(meds, "metformin") == nil {
			t.Error("expected metformin to be active for james")
		}
		if findMed(meds, "insulin_glargine") == nil {
			t.Error("expected insulin_glargine to be active for james")
		}
	})

	t.Run("events_out_of_order_handled", func(t *testing.T) {
		// Events inserted in non-chronological order must be sorted internally
		events := []TitrationEvent{
			{"pt_x", "metformin", "increase", 1000.0, "2024-03-01"},
			{"pt_x", "metformin", "start", 500.0, "2024-01-01"},
			{"pt_x", "metformin", "decrease", 750.0, "2024-02-01"},
		}
		xt := NewTitrationTracker(events)
		meds := xt.CurrentMedications("pt_x")
		if len(meds) != 1 {
			t.Fatalf("expected 1 med, got %d", len(meds))
		}
		if meds[0].CurrentDose != 1000.0 {
			t.Errorf("CurrentDose = %v, want 1000.0 (most recent event)", meds[0].CurrentDose)
		}
	})
}

func TestGetMedicationHistory(t *testing.T) {
	tr := newTracker(t)

	t.Run("sorted_chronologically", func(t *testing.T) {
		history := tr.GetMedicationHistory("maria", "metformin")
		for i := 1; i < len(history); i++ {
			if history[i].RecordedOn < history[i-1].RecordedOn {
				t.Errorf("history not sorted at index %d: %q < %q",
					i, history[i].RecordedOn, history[i-1].RecordedOn)
			}
		}
	})

	t.Run("correct_events_in_order", func(t *testing.T) {
		history := tr.GetMedicationHistory("maria", "metformin")
		wantDirections := []string{"start", "increase", "decrease", "stop"}
		if len(history) != len(wantDirections) {
			t.Fatalf("expected %d events, got %d", len(wantDirections), len(history))
		}
		for i, want := range wantDirections {
			if history[i].Direction != want {
				t.Errorf("event[%d].Direction = %q, want %q", i, history[i].Direction, want)
			}
		}
	})

	t.Run("unknown_returns_empty", func(t *testing.T) {
		history := tr.GetMedicationHistory("nobody", "metformin")
		if len(history) != 0 {
			t.Errorf("expected empty slice, got %v", history)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — TitrationCount + DeEscalationSummary
// ---------------------------------------------------------------------------

func TestTitrationCount(t *testing.T) {
	tr := newTracker(t)

	tests := []struct {
		name      string
		patient   string
		med       string
		direction string
		want      int
	}{
		{"total_count_for_metformin", "maria", "metformin", "", 4},
		{"directional_decrease", "maria", "metformin", "decrease", 1},
		{"directional_stop", "susan", "metformin", "stop", 1},
		{"directional_zero_no_such_direction", "james", "metformin", "stop", 0},
		{"unknown_patient_returns_zero", "nobody", "metformin", "", 0},
		{"unknown_medication_returns_zero", "maria", "insulin_glargine", "", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tr.TitrationCount(tc.patient, tc.med, tc.direction)
			if got != tc.want {
				t.Errorf("TitrationCount(%q, %q, %q) = %d, want %d",
					tc.patient, tc.med, tc.direction, got, tc.want)
			}
		})
	}
}

func TestDeEscalationSummary(t *testing.T) {
	tr := newTracker(t)

	t.Run("summary_includes_decrease_and_stop", func(t *testing.T) {
		summary := tr.DeEscalationSummary("maria")
		// metformin: 1 decrease + 1 stop = 2; glipizide: 1 decrease = 1
		if summary["metformin"] != 2 {
			t.Errorf("metformin count = %d, want 2", summary["metformin"])
		}
		if summary["glipizide"] != 1 {
			t.Errorf("glipizide count = %d, want 1", summary["glipizide"])
		}
	})

	t.Run("no_de_escalations_returns_empty", func(t *testing.T) {
		events := []TitrationEvent{
			{"pt_y", "metformin", "start", 500.0, "2024-01-01"},
			{"pt_y", "metformin", "increase", 1000.0, "2024-02-01"},
		}
		yt := NewTitrationTracker(events)
		summary := yt.DeEscalationSummary("pt_y")
		if len(summary) != 0 {
			t.Errorf("expected empty map, got %v", summary)
		}
	})

	t.Run("only_de_escalated_meds_included", func(t *testing.T) {
		// james has increases but also one decrease on insulin_glargine
		summary := tr.DeEscalationSummary("james")
		if summary["insulin_glargine"] != 1 {
			t.Errorf("insulin_glargine count = %d, want 1", summary["insulin_glargine"])
		}
		// metformin only has start+increase — not included
		if _, ok := summary["metformin"]; ok {
			t.Error("metformin should not be in de-escalation summary for james")
		}
	})

	t.Run("unknown_patient_returns_empty", func(t *testing.T) {
		summary := tr.DeEscalationSummary("nobody")
		if len(summary) != 0 {
			t.Errorf("expected empty map for unknown patient, got %v", summary)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — PatientsOnMedication + MostTitratedMedications
// ---------------------------------------------------------------------------

func TestPatientsOnMedication(t *testing.T) {
	tr := newTracker(t)

	t.Run("active_only_returned", func(t *testing.T) {
		// maria stopped metformin, susan stopped metformin, james still active
		patients := tr.PatientsOnMedication("metformin")
		if len(patients) != 1 || patients[0] != "james" {
			t.Errorf("got %v, want [james]", patients)
		}
	})

	t.Run("result_is_sorted", func(t *testing.T) {
		// add extra patients on glipizide to verify sorting
		extra := []TitrationEvent{
			{"zoe", "glipizide", "start", 5.0, "2024-01-01"},
			{"anna", "glipizide", "start", 5.0, "2024-01-01"},
		}
		et := NewTitrationTracker(append(allEvents(), extra...))
		patients := et.PatientsOnMedication("glipizide")
		for i := 1; i < len(patients); i++ {
			if patients[i] < patients[i-1] {
				t.Errorf("result not sorted at index %d: %q < %q", i, patients[i], patients[i-1])
			}
		}
		// verify membership
		found := make(map[string]bool)
		for _, p := range patients {
			found[p] = true
		}
		if !found["maria"] {
			t.Error("expected maria in patients_on_medication glipizide")
		}
		if !found["anna"] {
			t.Error("expected anna in patients_on_medication glipizide")
		}
		if !found["zoe"] {
			t.Error("expected zoe in patients_on_medication glipizide")
		}
		if found["susan"] {
			t.Error("susan should NOT be in patients_on_medication glipizide (stopped)")
		}
	})

	t.Run("unknown_medication_returns_empty", func(t *testing.T) {
		patients := tr.PatientsOnMedication("glipizide_xr")
		if len(patients) != 0 {
			t.Errorf("expected empty slice, got %v", patients)
		}
	})
}

func TestMostTitratedMedications(t *testing.T) {
	tr := newTracker(t)

	t.Run("metformin_is_most_titrated", func(t *testing.T) {
		top := tr.MostTitratedMedications(3)
		if len(top) == 0 {
			t.Fatal("expected at least 1 result")
		}
		if top[0].Name != "metformin" {
			t.Errorf("top medication = %q, want 'metformin'", top[0].Name)
		}
	})

	t.Run("sorted_descending_by_count", func(t *testing.T) {
		top := tr.MostTitratedMedications(3)
		for i := 1; i < len(top); i++ {
			if top[i].Count > top[i-1].Count {
				t.Errorf("result not sorted at index %d: %d > %d",
					i, top[i].Count, top[i-1].Count)
			}
		}
	})

	t.Run("correct_metformin_count", func(t *testing.T) {
		// metformin: maria(4) + james(2) + susan(3) = 9 events total
		top := tr.MostTitratedMedications(10)
		var metCount int
		for _, mc := range top {
			if mc.Name == "metformin" {
				metCount = mc.Count
				break
			}
		}
		if metCount != 9 {
			t.Errorf("metformin count = %d, want 9", metCount)
		}
	})

	t.Run("fewer_than_n_returns_all", func(t *testing.T) {
		events := []TitrationEvent{
			{"pt_z", "drug_a", "start", 10.0, "2024-01-01"},
		}
		zt := NewTitrationTracker(events)
		top := zt.MostTitratedMedications(5)
		if len(top) != 1 {
			t.Errorf("expected 1 result, got %d", len(top))
		}
	})
}

// ---------------------------------------------------------------------------
// PART 4 — AddEvent
// ---------------------------------------------------------------------------

func TestAddEvent(t *testing.T) {
	t.Run("new_event_updates_current_medications", func(t *testing.T) {
		tr := newTracker(t)
		// Adding a stop event removes the medication from CurrentMedications
		event := TitrationEvent{"james", "metformin", "stop", 0.0, "2024-06-01"}
		tr.AddEvent(event)
		meds := tr.CurrentMedications("james")
		if findMed(meds, "metformin") != nil {
			t.Error("metformin should not be active for james after stop event")
		}
	})

	t.Run("new_event_updates_history", func(t *testing.T) {
		tr := newTracker(t)
		event := TitrationEvent{"maria", "glipizide", "stop", 0.0, "2024-06-01"}
		tr.AddEvent(event)
		history := tr.GetMedicationHistory("maria", "glipizide")
		found := false
		for _, e := range history {
			if e.Direction == "stop" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'stop' event in maria's glipizide history")
		}
	})

	t.Run("overwrite_same_date", func(t *testing.T) {
		tr := newTracker(t)
		// An event on the same (patient, medication, date) overwrites the existing one
		// james metformin 2023-12-01 was "increase" — overwrite with "stop"
		event := TitrationEvent{"james", "metformin", "stop", 0.0, "2023-12-01"}
		tr.AddEvent(event)
		history := tr.GetMedicationHistory("james", "metformin")
		var dec *TitrationEvent
		for i := range history {
			if history[i].RecordedOn == "2023-12-01" {
				dec = &history[i]
				break
			}
		}
		if dec == nil {
			t.Fatal("event on 2023-12-01 not found in history")
		}
		if dec.Direction != "stop" {
			t.Errorf("Direction = %q, want 'stop' (overwrite)", dec.Direction)
		}
	})

	t.Run("add_event_new_patient", func(t *testing.T) {
		tr := newTracker(t)
		event := TitrationEvent{"new_pt", "metformin", "start", 500.0, "2024-05-01"}
		tr.AddEvent(event)
		meds := tr.CurrentMedications("new_pt")
		if len(meds) != 1 || meds[0].Name != "metformin" {
			t.Errorf("expected [metformin], got %v", meds)
		}
	})

	t.Run("add_event_affects_population_query", func(t *testing.T) {
		tr := newTracker(t)
		// After adding a new patient on metformin, they appear in PatientsOnMedication
		event := TitrationEvent{"new_pt2", "metformin", "start", 500.0, "2024-05-01"}
		tr.AddEvent(event)
		patients := tr.PatientsOnMedication("metformin")
		found := false
		for _, p := range patients {
			if p == "new_pt2" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected new_pt2 in PatientsOnMedication after AddEvent")
		}
	})
}
