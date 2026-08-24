// Tests for Problem 6: Lab Cadence Compliance Monitor
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_06_lab_cadence/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_06_lab_cadence.go \
//	  -c go test -v .
package labcadence

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newState(t *testing.T) *MonitorState {
	t.Helper()
	return MakeMonitor()
}

// seededState returns a MonitorState pre-loaded with a small realistic patient
// set that mirrors the Python fixture.
func seededState(t *testing.T) *MonitorState {
	t.Helper()
	s := MakeMonitor()

	mustRegister(t, s, "alice", []string{"hba1c", "bmp", "lipids"})
	mustRegister(t, s, "bob", []string{"hba1c", "bmp"})
	mustRegister(t, s, "carol", []string{"hba1c"})

	// Alice: hba1c quarterly deadlines; submitted first one, missed second
	mustDeadline(t, s, "alice", "hba1c", "2024-03-31")
	mustDeadline(t, s, "alice", "hba1c", "2024-06-30")
	mustSubmission(t, s, "alice", "hba1c", "2024-03-28")

	// Alice: bmp due but no submission
	mustDeadline(t, s, "alice", "bmp", "2024-04-15")

	// Bob: hba1c submitted on time; bmp overdue
	mustDeadline(t, s, "bob", "hba1c", "2024-03-31")
	mustSubmission(t, s, "bob", "hba1c", "2024-03-25")
	mustDeadline(t, s, "bob", "bmp", "2024-03-31")

	// Carol: all labs submitted on time
	mustDeadline(t, s, "carol", "hba1c", "2024-03-31")
	mustSubmission(t, s, "carol", "hba1c", "2024-03-15")

	return s
}

func mustRegister(t *testing.T, s *MonitorState, id string, labs []string) {
	t.Helper()
	if err := RegisterPatient(s, id, labs); err != nil {
		t.Fatalf("RegisterPatient(%q): %v", id, err)
	}
}

func mustDeadline(t *testing.T, s *MonitorState, id, lab, date string) {
	t.Helper()
	if err := SetLabDeadline(s, id, lab, date); err != nil {
		t.Fatalf("SetLabDeadline(%q, %q, %q): %v", id, lab, date, err)
	}
}

func mustSubmission(t *testing.T, s *MonitorState, id, lab, date string) {
	t.Helper()
	if err := RecordSubmission(s, id, lab, date); err != nil {
		t.Fatalf("RecordSubmission(%q, %q, %q): %v", id, lab, date, err)
	}
}

// ---------------------------------------------------------------------------
// PART 1 — Patient & lab registration
// ---------------------------------------------------------------------------

func TestRegisterPatient(t *testing.T) {
	t.Run("registers_new_patient", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "dave", []string{"hba1c"})
		labs, err := GetRequiredLabs(s, "dave")
		if err != nil {
			t.Fatalf("GetRequiredLabs: %v", err)
		}
		if !labs["hba1c"] {
			t.Error("expected hba1c to be required for dave")
		}
	})

	t.Run("empty_required_labs_returns_error", func(t *testing.T) {
		s := newState(t)
		if err := RegisterPatient(s, "eve", []string{}); !errors.Is(err, ErrEmptyLabList) {
			t.Errorf("expected ErrEmptyLabList, got %v", err)
		}
	})

	t.Run("idempotent_for_existing_lab", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "frank", []string{"hba1c"})
		mustRegister(t, s, "frank", []string{"hba1c"}) // duplicate call
		labs, _ := GetRequiredLabs(s, "frank")
		if len(labs) != 1 {
			t.Errorf("expected 1 lab, got %d: %v", len(labs), labs)
		}
	})

	t.Run("adds_new_labs_to_existing_patient", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "grace", []string{"hba1c"})
		mustRegister(t, s, "grace", []string{"bmp"})
		labs, _ := GetRequiredLabs(s, "grace")
		if !labs["hba1c"] || !labs["bmp"] {
			t.Errorf("expected both hba1c and bmp, got %v", labs)
		}
	})

	t.Run("does_not_remove_existing_labs", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "hank", []string{"hba1c", "bmp"})
		mustRegister(t, s, "hank", []string{"lipids"})
		labs, _ := GetRequiredLabs(s, "hank")
		if !labs["hba1c"] || !labs["bmp"] {
			t.Error("original labs should still be required after second call")
		}
	})
}

func TestAddRequiredLab(t *testing.T) {
	t.Run("adds_lab_to_existing_patient", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "iris", []string{"hba1c"})
		if err := AddRequiredLab(s, "iris", "bmp"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		labs, _ := GetRequiredLabs(s, "iris")
		if !labs["bmp"] {
			t.Error("expected bmp to be required for iris")
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "jack", []string{"hba1c"})
		if err := AddRequiredLab(s, "jack", "hba1c"); err != nil { // already required
			t.Errorf("expected no error, got %v", err)
		}
		labs, _ := GetRequiredLabs(s, "jack")
		if len(labs) != 1 {
			t.Errorf("expected 1 lab, got %d", len(labs))
		}
	})

	t.Run("unknown_patient_returns_error", func(t *testing.T) {
		s := newState(t)
		if err := AddRequiredLab(s, "nobody", "hba1c"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetRequiredLabs(t *testing.T) {
	t.Run("returns_correct_set", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "kate", []string{"hba1c", "lipids"})
		labs, err := GetRequiredLabs(s, "kate")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !labs["hba1c"] || !labs["lipids"] {
			t.Errorf("expected {hba1c, lipids}, got %v", labs)
		}
	})

	t.Run("unknown_patient_returns_error", func(t *testing.T) {
		s := newState(t)
		if _, err := GetRequiredLabs(s, "nobody"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Deadlines and submissions
// ---------------------------------------------------------------------------

func TestSetLabDeadline(t *testing.T) {
	t.Run("sets_deadline_no_error", func(t *testing.T) {
		s := seededState(t)
		if err := SetLabDeadline(s, "alice", "lipids", "2024-05-01"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("duplicate_deadline_ignored", func(t *testing.T) {
		s := seededState(t)
		mustDeadline(t, s, "alice", "lipids", "2024-05-01")
		mustDeadline(t, s, "alice", "lipids", "2024-05-01") // duplicate
		// should still register as overdue exactly once
		if !IsOverdue(s, "alice", "lipids", "2024-05-02") {
			t.Error("expected lipids to be overdue on 2024-05-02")
		}
	})

	t.Run("unknown_patient_returns_error", func(t *testing.T) {
		s := newState(t)
		if err := SetLabDeadline(s, "nobody", "hba1c", "2024-03-31"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("unrequired_lab_returns_error", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "leo", []string{"hba1c"})
		if err := SetLabDeadline(s, "leo", "bmp", "2024-03-31"); !errors.Is(err, ErrLabNotRequired) {
			t.Errorf("expected ErrLabNotRequired, got %v", err)
		}
	})
}

func TestRecordSubmission(t *testing.T) {
	t.Run("submission_clears_deadline", func(t *testing.T) {
		s := seededState(t)
		// carol submitted hba1c on Mar 15, deadline was Mar 31 → not overdue
		if IsOverdue(s, "carol", "hba1c", "2024-04-01") {
			t.Error("carol hba1c should not be overdue (submitted on time)")
		}
	})

	t.Run("submission_clears_earliest_applicable_deadline", func(t *testing.T) {
		s := seededState(t)
		// alice: hba1c deadlines Mar 31 (cleared by Mar 28 submission) and Jun 30 (not cleared)
		// as of Jul 1, the Jun 30 deadline should be overdue
		if !IsOverdue(s, "alice", "hba1c", "2024-07-01") {
			t.Error("alice hba1c should be overdue on 2024-07-01 (Jun 30 deadline missed)")
		}
	})

	t.Run("unknown_patient_returns_error", func(t *testing.T) {
		s := newState(t)
		if err := RecordSubmission(s, "nobody", "hba1c", "2024-03-28"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("unrequired_lab_returns_error", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "mia", []string{"hba1c"})
		if err := RecordSubmission(s, "mia", "bmp", "2024-03-28"); !errors.Is(err, ErrLabNotRequired) {
			t.Errorf("expected ErrLabNotRequired, got %v", err)
		}
	})

	t.Run("late_submission_does_not_clear_past_deadline", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "noah", []string{"hba1c"})
		mustDeadline(t, s, "noah", "hba1c", "2024-03-31")
		mustSubmission(t, s, "noah", "hba1c", "2024-04-15") // submitted late
		// deadline Mar 31 should still be overdue as of Apr 1
		if !IsOverdue(s, "noah", "hba1c", "2024-04-01") {
			t.Error("noah hba1c should still be overdue (submission was after deadline)")
		}
	})
}

func TestIsOverdue(t *testing.T) {
	t.Run("overdue_deadline_no_submission", func(t *testing.T) {
		s := seededState(t)
		// bob: bmp deadline Mar 31, no submission
		if !IsOverdue(s, "bob", "bmp", "2024-04-01") {
			t.Error("expected bob bmp to be overdue on 2024-04-01")
		}
	})

	t.Run("not_overdue_when_submitted", func(t *testing.T) {
		s := seededState(t)
		// bob: hba1c submitted on time
		if IsOverdue(s, "bob", "hba1c", "2024-04-01") {
			t.Error("bob hba1c should not be overdue (submitted on time)")
		}
	})

	t.Run("not_overdue_when_deadline_not_yet_passed", func(t *testing.T) {
		s := seededState(t)
		// alice: bmp deadline Apr 15; as of Apr 14 not yet overdue
		if IsOverdue(s, "alice", "bmp", "2024-04-14") {
			t.Error("alice bmp should not be overdue on 2024-04-14 (deadline is Apr 15)")
		}
	})

	t.Run("overdue_on_day_after_deadline", func(t *testing.T) {
		s := seededState(t)
		// alice: bmp deadline Apr 15; Apr 16 → overdue
		if !IsOverdue(s, "alice", "bmp", "2024-04-16") {
			t.Error("alice bmp should be overdue on 2024-04-16")
		}
	})

	t.Run("not_overdue_when_no_deadline_set", func(t *testing.T) {
		s := seededState(t)
		// alice: lipids has no deadline set
		if IsOverdue(s, "alice", "lipids", "2024-06-01") {
			t.Error("alice lipids should not be overdue (no deadline set)")
		}
	})

	t.Run("unknown_patient_returns_false", func(t *testing.T) {
		s := seededState(t)
		if IsOverdue(s, "nobody", "hba1c", "2024-04-01") {
			t.Error("expected false for unknown patient")
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Compliance reporting
// ---------------------------------------------------------------------------

func TestOverdueLabs(t *testing.T) {
	t.Run("returns_sorted_overdue_labs", func(t *testing.T) {
		s := seededState(t)
		// alice as of Jul 1: bmp (missed Apr 15) and hba1c (missed Jun 30)
		labs := OverdueLabs(s, "alice", "2024-07-01")
		for i := 1; i < len(labs); i++ {
			if labs[i] < labs[i-1] {
				t.Errorf("labs not sorted at index %d: %q < %q", i, labs[i], labs[i-1])
			}
		}
		found := make(map[string]bool)
		for _, l := range labs {
			found[l] = true
		}
		if !found["bmp"] {
			t.Error("expected bmp in overdue labs for alice")
		}
		if !found["hba1c"] {
			t.Error("expected hba1c in overdue labs for alice")
		}
	})

	t.Run("no_overdue_returns_empty", func(t *testing.T) {
		s := seededState(t)
		labs := OverdueLabs(s, "carol", "2024-04-01")
		if len(labs) != 0 {
			t.Errorf("expected empty, got %v", labs)
		}
	})

	t.Run("unknown_patient_returns_empty", func(t *testing.T) {
		s := seededState(t)
		labs := OverdueLabs(s, "nobody", "2024-04-01")
		if len(labs) != 0 {
			t.Errorf("expected empty, got %v", labs)
		}
	})
}

func TestComplianceReport(t *testing.T) {
	t.Run("report_contains_overdue_patients", func(t *testing.T) {
		s := seededState(t)
		report := ComplianceReport(s, "2024-04-16")
		found := make(map[string]bool)
		for _, e := range report {
			found[e.PatientID] = true
		}
		if !found["alice"] {
			t.Error("expected alice in compliance report (bmp overdue)")
		}
		if !found["bob"] {
			t.Error("expected bob in compliance report (bmp overdue)")
		}
	})

	t.Run("compliant_patients_excluded", func(t *testing.T) {
		s := seededState(t)
		report := ComplianceReport(s, "2024-04-16")
		for _, e := range report {
			if e.PatientID == "carol" {
				t.Error("carol should not be in compliance report (all submitted on time)")
			}
		}
	})

	t.Run("entry_fields_correct", func(t *testing.T) {
		s := seededState(t)
		report := ComplianceReport(s, "2024-04-16")
		var bobEntry *ComplianceEntry
		for i := range report {
			if report[i].PatientID == "bob" {
				bobEntry = &report[i]
				break
			}
		}
		if bobEntry == nil {
			t.Fatal("bob not found in report")
		}
		if bobEntry.OverdueCount != len(bobEntry.OverdueLabs) {
			t.Errorf("OverdueCount %d != len(OverdueLabs) %d",
				bobEntry.OverdueCount, len(bobEntry.OverdueLabs))
		}
	})

	t.Run("sorted_by_overdue_count_descending", func(t *testing.T) {
		s := seededState(t)
		report := ComplianceReport(s, "2024-07-01")
		for i := 1; i < len(report); i++ {
			if report[i].OverdueCount > report[i-1].OverdueCount {
				t.Errorf("report not sorted at index %d: %d > %d",
					i, report[i].OverdueCount, report[i-1].OverdueCount)
			}
		}
	})

	t.Run("empty_report_when_no_overdue", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "perfectly_compliant", []string{"hba1c"})
		mustDeadline(t, s, "perfectly_compliant", "hba1c", "2024-03-31")
		mustSubmission(t, s, "perfectly_compliant", "hba1c", "2024-03-20")
		report := ComplianceReport(s, "2024-04-01")
		if len(report) != 0 {
			t.Errorf("expected empty report, got %v", report)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 4 — Submission history
// ---------------------------------------------------------------------------

func TestSubmissionHistory(t *testing.T) {
	t.Run("returns_sorted_dates", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "quinn", []string{"hba1c"})
		mustDeadline(t, s, "quinn", "hba1c", "2024-03-31")
		mustDeadline(t, s, "quinn", "hba1c", "2024-06-30")
		mustSubmission(t, s, "quinn", "hba1c", "2024-03-20")
		mustSubmission(t, s, "quinn", "hba1c", "2024-06-15")
		history := SubmissionHistory(s, "quinn", "hba1c")
		for i := 1; i < len(history); i++ {
			if history[i] < history[i-1] {
				t.Errorf("history not sorted at index %d: %q < %q", i, history[i], history[i-1])
			}
		}
		if len(history) != 2 {
			t.Errorf("expected 2 submissions, got %d", len(history))
		}
	})

	t.Run("no_submissions_returns_empty", func(t *testing.T) {
		s := seededState(t)
		// alice has no lipids submissions
		history := SubmissionHistory(s, "alice", "lipids")
		if len(history) != 0 {
			t.Errorf("expected empty history, got %v", history)
		}
	})

	t.Run("unknown_patient_returns_empty", func(t *testing.T) {
		s := newState(t)
		history := SubmissionHistory(s, "nobody", "hba1c")
		if len(history) != 0 {
			t.Errorf("expected empty history, got %v", history)
		}
	})
}

func TestDaysSinceLastSubmission(t *testing.T) {
	t.Run("correct_days", func(t *testing.T) {
		s := newState(t)
		mustRegister(t, s, "rita", []string{"hba1c"})
		mustDeadline(t, s, "rita", "hba1c", "2024-03-31")
		mustSubmission(t, s, "rita", "hba1c", "2024-03-20")
		got := DaysSinceLastSubmission(s, "rita", "hba1c", "2024-04-20")
		if got != 31 {
			t.Errorf("DaysSinceLastSubmission = %d, want 31", got)
		}
	})

	t.Run("no_submission_returns_sentinel", func(t *testing.T) {
		s := seededState(t)
		// alice has no lipids submissions
		got := DaysSinceLastSubmission(s, "alice", "lipids", "2024-05-01")
		if got != -1 {
			t.Errorf("expected -1 (no data), got %d", got)
		}
	})

	t.Run("unknown_patient_returns_sentinel", func(t *testing.T) {
		s := newState(t)
		got := DaysSinceLastSubmission(s, "nobody", "hba1c", "2024-04-01")
		if got != -1 {
			t.Errorf("expected -1 (no data), got %d", got)
		}
	})
}
