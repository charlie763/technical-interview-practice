// Tests for Problem 4: Biomarker Alert Monitor
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_04_biomarker_alert_monitor/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_04_biomarker_alert_monitor.go \
//	  -c go test -v .
package biomarker

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Test data — mirrors the Python test fixtures
// ---------------------------------------------------------------------------

var aliceReadings = []BiomarkerReading{
	// All in range — no outreach needed
	{"alice", "glucose", 105.0, "2024-01-01"},
	{"alice", "glucose", 98.0, "2024-01-02"},
	{"alice", "glucose", 112.0, "2024-01-03"},
	{"alice", "glucose", 91.0, "2024-01-04"},
}

var bobReadings = []BiomarkerReading{
	// 4 consecutive high-glucose days → needs outreach
	{"bob", "glucose", 195.0, "2024-01-01"},
	{"bob", "glucose", 210.0, "2024-01-02"},
	{"bob", "glucose", 188.0, "2024-01-03"},
	{"bob", "glucose", 202.0, "2024-01-04"},
}

var carolReadings = []BiomarkerReading{
	// Streak of 1 (Jan 1), in-range on Jan 2, then streak of 2 (Jan 3–4) → max=2
	{"carol", "glucose", 190.0, "2024-01-01"},
	{"carol", "glucose", 150.0, "2024-01-02"}, // in range
	{"carol", "glucose", 185.0, "2024-01-03"},
	{"carol", "glucose", 191.0, "2024-01-04"},
}

var daveReadings = []BiomarkerReading{
	// 2 consecutive high-glucose days — below default threshold of 3
	{"dave", "glucose", 199.0, "2024-01-03"},
	{"dave", "glucose", 205.0, "2024-01-04"},
}

var eveReadings = []BiomarkerReading{
	// One dangerous low glucose (below 70) on Jan 1 — streak of 1 for glucose
	{"eve", "glucose", 62.0, "2024-01-01"},
	// 3 consecutive days with ketones below target (< 0.5) → ketone outreach
	{"eve", "ketone", 0.3, "2024-01-02"},
	{"eve", "ketone", 0.2, "2024-01-03"},
	{"eve", "ketone", 0.4, "2024-01-04"},
	// Weight readings — never out of range
	{"eve", "weight", 165.0, "2024-01-01"},
	{"eve", "weight", 164.5, "2024-01-02"},
}

func allReadings() []BiomarkerReading {
	var all []BiomarkerReading
	all = append(all, aliceReadings...)
	all = append(all, bobReadings...)
	all = append(all, carolReadings...)
	all = append(all, daveReadings...)
	all = append(all, eveReadings...)
	return all
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newMonitor(t *testing.T) BiomarkerMonitor {
	t.Helper()
	return NewBiomarkerMonitor(allReadings())
}

func freshMonitor(t *testing.T) BiomarkerMonitor {
	t.Helper()
	return NewBiomarkerMonitor(nil)
}

// ---------------------------------------------------------------------------
// PART 1 — IsOutOfRange
// ---------------------------------------------------------------------------

func TestIsOutOfRange(t *testing.T) {
	m := freshMonitor(t)

	tests := []struct {
		name    string
		reading BiomarkerReading
		want    bool
	}{
		{"glucose_above_range", BiomarkerReading{"p1", "glucose", 181.0, "2024-01-01"}, true},
		{"glucose_below_range", BiomarkerReading{"p1", "glucose", 69.9, "2024-01-01"}, true},
		{"glucose_at_upper_boundary", BiomarkerReading{"p1", "glucose", 180.0, "2024-01-01"}, false},
		{"glucose_at_lower_boundary", BiomarkerReading{"p1", "glucose", 70.0, "2024-01-01"}, false},
		{"glucose_in_range", BiomarkerReading{"p1", "glucose", 120.0, "2024-01-01"}, false},
		{"ketone_above_range", BiomarkerReading{"p1", "ketone", 3.1, "2024-01-01"}, true},
		{"ketone_below_range", BiomarkerReading{"p1", "ketone", 0.4, "2024-01-01"}, true},
		{"ketone_in_range", BiomarkerReading{"p1", "ketone", 1.5, "2024-01-01"}, false},
		{"ketone_at_lower_boundary", BiomarkerReading{"p1", "ketone", 0.5, "2024-01-01"}, false},
		{"ketone_at_upper_boundary", BiomarkerReading{"p1", "ketone", 3.0, "2024-01-01"}, false},
		{"weight_never_out_of_range", BiomarkerReading{"p1", "weight", 9999.0, "2024-01-01"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := m.IsOutOfRange(tc.reading)
			if got != tc.want {
				t.Errorf("IsOutOfRange(%v) = %v, want %v", tc.reading, got, tc.want)
			}
		})
	}

	t.Run("classification_uses_preloaded_data", func(t *testing.T) {
		// IsOutOfRange works on any BiomarkerReading regardless of monitor state
		seeded := newMonitor(t)
		inRange := BiomarkerReading{"alice", "glucose", 100.0, "2024-01-05"}
		outOfRange := BiomarkerReading{"bob", "glucose", 250.0, "2024-01-05"}
		if seeded.IsOutOfRange(inRange) {
			t.Error("expected in-range reading to return false")
		}
		if !seeded.IsOutOfRange(outOfRange) {
			t.Error("expected out-of-range reading to return true")
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — MaxConsecutiveOutOfRangeDays
// ---------------------------------------------------------------------------

func TestMaxConsecutiveOutOfRangeDays(t *testing.T) {
	m := newMonitor(t)

	t.Run("unknown_patient_returns_zero", func(t *testing.T) {
		if got := m.MaxConsecutiveOutOfRangeDays("unknown_patient", "glucose"); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("all_in_range_returns_zero", func(t *testing.T) {
		if got := m.MaxConsecutiveOutOfRangeDays("alice", "glucose"); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("consecutive_streak_of_four", func(t *testing.T) {
		if got := m.MaxConsecutiveOutOfRangeDays("bob", "glucose"); got != 4 {
			t.Errorf("got %d, want 4", got)
		}
	})

	t.Run("streak_broken_by_in_range_day", func(t *testing.T) {
		// carol: streak of 1, break, streak of 2 → max = 2
		if got := m.MaxConsecutiveOutOfRangeDays("carol", "glucose"); got != 2 {
			t.Errorf("got %d, want 2", got)
		}
	})

	t.Run("short_streak_of_two", func(t *testing.T) {
		if got := m.MaxConsecutiveOutOfRangeDays("dave", "glucose"); got != 2 {
			t.Errorf("got %d, want 2", got)
		}
	})

	t.Run("glucose_streak_of_one_for_low", func(t *testing.T) {
		if got := m.MaxConsecutiveOutOfRangeDays("eve", "glucose"); got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("ketone_streak_of_three", func(t *testing.T) {
		if got := m.MaxConsecutiveOutOfRangeDays("eve", "ketone"); got != 3 {
			t.Errorf("got %d, want 3", got)
		}
	})

	t.Run("weight_never_streaks", func(t *testing.T) {
		if got := m.MaxConsecutiveOutOfRangeDays("eve", "weight"); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("multiple_readings_same_day_count_as_one", func(t *testing.T) {
		readings := []BiomarkerReading{
			{"frank", "glucose", 200.0, "2024-02-01"},
			{"frank", "glucose", 210.0, "2024-02-01"}, // same day
			{"frank", "glucose", 195.0, "2024-02-02"},
		}
		fm := NewBiomarkerMonitor(readings)
		if got := fm.MaxConsecutiveOutOfRangeDays("frank", "glucose"); got != 2 {
			t.Errorf("got %d, want 2", got)
		}
	})

	t.Run("gap_breaks_streak", func(t *testing.T) {
		// Mar 1 and Mar 3 (no Mar 2) are two separate streaks of 1
		readings := []BiomarkerReading{
			{"grace", "glucose", 200.0, "2024-03-01"},
			{"grace", "glucose", 200.0, "2024-03-03"}, // skip Mar 2
		}
		gm := NewBiomarkerMonitor(readings)
		if got := gm.MaxConsecutiveOutOfRangeDays("grace", "glucose"); got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("wrong_reading_type_ignored", func(t *testing.T) {
		// Ketone readings don't count toward glucose streak and vice versa
		if got := m.MaxConsecutiveOutOfRangeDays("eve", "glucose"); got != 1 {
			t.Errorf("glucose streak = %d, want 1", got)
		}
		if got := m.MaxConsecutiveOutOfRangeDays("eve", "ketone"); got != 3 {
			t.Errorf("ketone streak = %d, want 3", got)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — GetOutreachList
// ---------------------------------------------------------------------------

func TestGetOutreachList(t *testing.T) {
	m := newMonitor(t)

	t.Run("default_threshold_filters_correctly", func(t *testing.T) {
		// With min=3, only bob and eve (ketone) qualify
		result := m.GetOutreachList(3)
		type key struct{ p, r string }
		found := make(map[key]bool)
		for _, e := range result {
			found[key{e.PatientID, e.ReadingType}] = true
		}
		if !found[key{"bob", "glucose"}] {
			t.Error("expected (bob, glucose) in outreach list")
		}
		if !found[key{"eve", "ketone"}] {
			t.Error("expected (eve, ketone) in outreach list")
		}
		if found[key{"alice", "glucose"}] {
			t.Error("alice should not be in outreach list (all in range)")
		}
		if found[key{"carol", "glucose"}] {
			t.Error("carol should not be in outreach list (max streak 2 < 3)")
		}
		if found[key{"dave", "glucose"}] {
			t.Error("dave should not be in outreach list (max streak 2 < 3)")
		}
	})

	t.Run("weight_never_in_outreach", func(t *testing.T) {
		result := m.GetOutreachList(1)
		for _, e := range result {
			if e.ReadingType == "weight" {
				t.Errorf("weight should never appear in outreach list, got %+v", e)
			}
		}
	})

	t.Run("sorted_by_consecutive_days_descending", func(t *testing.T) {
		result := m.GetOutreachList(1)
		for i := 1; i < len(result); i++ {
			if result[i].ConsecutiveDays > result[i-1].ConsecutiveDays {
				t.Errorf("result not sorted descending at index %d: %d > %d",
					i, result[i].ConsecutiveDays, result[i-1].ConsecutiveDays)
			}
		}
	})

	t.Run("entry_fields_bob", func(t *testing.T) {
		result := m.GetOutreachList(3)
		var bobEntry *OutreachEntry
		for i := range result {
			if result[i].PatientID == "bob" {
				bobEntry = &result[i]
				break
			}
		}
		if bobEntry == nil {
			t.Fatal("bob not found in outreach list")
		}
		if bobEntry.ConsecutiveDays != 4 {
			t.Errorf("ConsecutiveDays = %d, want 4", bobEntry.ConsecutiveDays)
		}
		if bobEntry.ReadingType != "glucose" {
			t.Errorf("ReadingType = %q, want 'glucose'", bobEntry.ReadingType)
		}
		// Bob's most recent OOR reading is Jan 4: 202.0
		if bobEntry.LatestValue != 202.0 {
			t.Errorf("LatestValue = %v, want 202.0", bobEntry.LatestValue)
		}
	})

	t.Run("patient_can_appear_twice_for_different_types", func(t *testing.T) {
		readings := []BiomarkerReading{
			// 3-day glucose streak
			{"hank", "glucose", 200.0, "2024-01-01"},
			{"hank", "glucose", 210.0, "2024-01-02"},
			{"hank", "glucose", 195.0, "2024-01-03"},
			// 3-day ketone streak
			{"hank", "ketone", 0.2, "2024-01-01"},
			{"hank", "ketone", 0.3, "2024-01-02"},
			{"hank", "ketone", 0.1, "2024-01-03"},
		}
		hm := NewBiomarkerMonitor(readings)
		result := hm.GetOutreachList(3)
		type key struct{ p, r string }
		found := make(map[key]bool)
		for _, e := range result {
			found[key{e.PatientID, e.ReadingType}] = true
		}
		if !found[key{"hank", "glucose"}] {
			t.Error("expected (hank, glucose) in outreach list")
		}
		if !found[key{"hank", "ketone"}] {
			t.Error("expected (hank, ketone) in outreach list")
		}
	})

	t.Run("custom_threshold_of_two", func(t *testing.T) {
		result := m.GetOutreachList(2)
		type key struct{ p, r string }
		found := make(map[key]bool)
		for _, e := range result {
			found[key{e.PatientID, e.ReadingType}] = true
		}
		if !found[key{"carol", "glucose"}] {
			t.Error("expected carol in outreach list at threshold=2")
		}
		if !found[key{"dave", "glucose"}] {
			t.Error("expected dave in outreach list at threshold=2")
		}
	})

	t.Run("empty_monitor_returns_empty", func(t *testing.T) {
		fm := freshMonitor(t)
		result := fm.GetOutreachList(3)
		if len(result) != 0 {
			t.Errorf("got %d entries, want 0", len(result))
		}
	})
}

// ---------------------------------------------------------------------------
// PART 4 — AddReading
// ---------------------------------------------------------------------------

func TestAddReading(t *testing.T) {
	t.Run("new_reading_returns_true", func(t *testing.T) {
		m := newMonitor(t)
		r := BiomarkerReading{"alice", "glucose", 100.0, "2024-01-10"}
		if got := m.AddReading(r); !got {
			t.Error("expected true for new reading")
		}
	})

	t.Run("duplicate_exact_same_returns_false", func(t *testing.T) {
		m := newMonitor(t)
		// alice Jan 1 = 105.0; exact duplicate
		r := BiomarkerReading{"alice", "glucose", 105.0, "2024-01-01"}
		if got := m.AddReading(r); got {
			t.Error("expected false for exact duplicate")
		}
	})

	t.Run("duplicate_within_tolerance_returns_false", func(t *testing.T) {
		m := newMonitor(t)
		// alice Jan 1 = 105.0; 105.4 is within ±0.5
		r := BiomarkerReading{"alice", "glucose", 105.4, "2024-01-01"}
		if got := m.AddReading(r); got {
			t.Error("expected false for reading within ±0.5 tolerance")
		}
	})

	t.Run("just_outside_tolerance_returns_true", func(t *testing.T) {
		m := newMonitor(t)
		// alice Jan 1 = 105.0; 105.6 is outside ±0.5
		r := BiomarkerReading{"alice", "glucose", 105.6, "2024-01-01"}
		if got := m.AddReading(r); !got {
			t.Error("expected true for reading just outside ±0.5 tolerance")
		}
	})

	t.Run("different_date_not_duplicate", func(t *testing.T) {
		m := newMonitor(t)
		r := BiomarkerReading{"alice", "glucose", 105.0, "2024-01-05"}
		if got := m.AddReading(r); !got {
			t.Error("expected true for different date")
		}
	})

	t.Run("different_type_not_duplicate", func(t *testing.T) {
		m := newMonitor(t)
		r := BiomarkerReading{"alice", "ketone", 1.0, "2024-01-01"}
		if got := m.AddReading(r); !got {
			t.Error("expected true for different reading type")
		}
	})

	t.Run("different_patient_not_duplicate", func(t *testing.T) {
		m := newMonitor(t)
		r := BiomarkerReading{"frank", "glucose", 105.0, "2024-01-01"}
		if got := m.AddReading(r); !got {
			t.Error("expected true for different patient")
		}
	})

	t.Run("added_reading_affects_streak", func(t *testing.T) {
		m := newMonitor(t)
		// bob's current streak is Jan 1–4 (4 days). Add Jan 5.
		r := BiomarkerReading{"bob", "glucose", 195.0, "2024-01-05"}
		m.AddReading(r)
		if got := m.MaxConsecutiveOutOfRangeDays("bob", "glucose"); got != 5 {
			t.Errorf("got %d, want 5 after adding Jan 5", got)
		}
	})

	t.Run("adding_duplicate_does_not_affect_streak", func(t *testing.T) {
		m := newMonitor(t)
		// bob Jan 1 = 195.0; duplicate, rejected
		r := BiomarkerReading{"bob", "glucose", 195.0, "2024-01-01"}
		m.AddReading(r)
		if got := m.MaxConsecutiveOutOfRangeDays("bob", "glucose"); got != 4 {
			t.Errorf("got %d, want 4 (duplicate should not inflate streak)", got)
		}
	})
}
