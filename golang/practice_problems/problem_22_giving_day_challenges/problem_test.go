// Tests for Problem 22: Giving Day Challenge Engine
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_22_giving_day_challenges/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_22_giving_day_challenges.go \
//	  -c go test -v .
package challenges

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Constants used across tests
// ---------------------------------------------------------------------------

const morningDeadline = "2025-10-01T12:00:00"
const afternoonDeadline = "2025-10-01T18:00:00"

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newEngine(t *testing.T) GivingDayEngine {
	t.Helper()
	return NewGivingDayEngine()
}

func mustAddChallenge(t *testing.T, e GivingDayEngine, id, name string, threshold float64, deadline string, bt BonusType, bv float64) *Challenge {
	t.Helper()
	c, err := e.AddChallenge(id, name, threshold, deadline, bt, bv)
	if err != nil {
		t.Fatalf("AddChallenge(%q): unexpected error: %v", id, err)
	}
	return c
}

func mustAddDonation(t *testing.T, e GivingDayEngine, amount float64, timestamp string) *DonationEntry {
	t.Helper()
	d, err := e.AddDonation(amount, timestamp)
	if err != nil {
		t.Fatalf("AddDonation(%v, %q): unexpected error: %v", amount, timestamp, err)
	}
	return d
}

func mustProcess(t *testing.T, e GivingDayEngine, amount float64, timestamp string) *ProcessDonationResult {
	t.Helper()
	r, err := e.ProcessDonation(amount, timestamp)
	if err != nil {
		t.Fatalf("ProcessDonation(%v, %q): unexpected error: %v", amount, timestamp, err)
	}
	return r
}

// seededEngine returns an engine with two challenges and some donations processed.
// Challenges:
//   - c1: threshold=1000, deadline=12:00 (morning challenge)
//   - c2: threshold=2000, deadline=18:00 (afternoon challenge)
func seededEngine(t *testing.T) GivingDayEngine {
	t.Helper()
	e := NewGivingDayEngine()
	mustAddChallenge(t, e, "c1", "Morning Boost", 1000, morningDeadline, BonusTypeFlat, 500)
	mustAddChallenge(t, e, "c2", "Afternoon Push", 2000, afternoonDeadline, BonusTypeMatchPct, 10)
	return e
}

// hourlyEngine returns an engine seeded for Part 3 hourly-breakdown tests.
//
//	09:15 → $400, 09:45 → $300   (total 09:xx = 700; cumulative = 700)
//	11:30 → $400                 (total 11:xx = 400; cumulative = 1100; c1 completes)
//	14:00 → $500                 (total 14:xx = 500; cumulative = 1600)
//	17:00 → $500                 (total 17:xx = 500; cumulative = 2100; c2 completes)
func hourlyEngine(t *testing.T) GivingDayEngine {
	t.Helper()
	e := NewGivingDayEngine()
	mustAddChallenge(t, e, "c1", "Morning Boost", 1000, morningDeadline, BonusTypeFlat, 500)
	mustAddChallenge(t, e, "c2", "Afternoon Push", 1800, afternoonDeadline, BonusTypeMatchPct, 10)
	mustProcess(t, e, 400, "2025-10-01T09:15:00")
	mustProcess(t, e, 300, "2025-10-01T09:45:00")
	mustProcess(t, e, 400, "2025-10-01T11:30:00")
	mustProcess(t, e, 500, "2025-10-01T14:00:00")
	mustProcess(t, e, 500, "2025-10-01T17:00:00")
	return e
}

// findBucket returns the bucket with the given hour, or nil.
func findBucket(buckets []*HourlyBucket, hour string) *HourlyBucket {
	for _, b := range buckets {
		if b.Hour == hour {
			return b
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// PART 1 — AddChallenge, AddDonation, GetChallenge, GetTotalRaised
// ---------------------------------------------------------------------------

func TestAddChallenge(t *testing.T) {
	t.Run("returns_challenge_with_correct_fields", func(t *testing.T) {
		e := newEngine(t)
		c := mustAddChallenge(t, e, "c1", "Morning Boost", 1000, morningDeadline, BonusTypeFlat, 500)
		if c.ID != "c1" {
			t.Errorf("ID = %q, want 'c1'", c.ID)
		}
		if c.Name != "Morning Boost" {
			t.Errorf("Name = %q, want 'Morning Boost'", c.Name)
		}
		if c.ThresholdAmount != 1000 {
			t.Errorf("ThresholdAmount = %v, want 1000", c.ThresholdAmount)
		}
		if c.Deadline != morningDeadline {
			t.Errorf("Deadline = %q, want %q", c.Deadline, morningDeadline)
		}
		if c.BonusType != BonusTypeFlat {
			t.Errorf("BonusType = %q, want %q", c.BonusType, BonusTypeFlat)
		}
		if c.BonusValue != 500 {
			t.Errorf("BonusValue = %v, want 500", c.BonusValue)
		}
		if c.Completed {
			t.Error("Completed = true, want false")
		}
		if c.CompletedAt != nil {
			t.Errorf("CompletedAt = %v, want nil", c.CompletedAt)
		}
	})

	t.Run("duplicate_id_returns_error", func(t *testing.T) {
		e := newEngine(t)
		mustAddChallenge(t, e, "c1", "A", 500, morningDeadline, BonusTypeFlat, 100)
		_, err := e.AddChallenge("c1", "B", 500, morningDeadline, BonusTypeMatchPct, 10)
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("zero_threshold_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.AddChallenge("c2", "Bad", 0, morningDeadline, BonusTypeFlat, 100)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("negative_threshold_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.AddChallenge("c3", "Neg", -100, morningDeadline, BonusTypeFlat, 100)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})
}

func TestGetChallenge(t *testing.T) {
	t.Run("returns_correct_challenge", func(t *testing.T) {
		e := newEngine(t)
		mustAddChallenge(t, e, "c1", "Morning Boost", 1000, morningDeadline, BonusTypeFlat, 500)
		c, err := e.GetChallenge("c1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID != "c1" {
			t.Errorf("ID = %q, want 'c1'", c.ID)
		}
	})

	t.Run("unknown_id_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.GetChallenge("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestAddDonation(t *testing.T) {
	t.Run("returns_donation_with_correct_fields", func(t *testing.T) {
		e := newEngine(t)
		d := mustAddDonation(t, e, 400, "2025-10-01T09:00:00")
		if d.Amount != 400 {
			t.Errorf("Amount = %v, want 400", d.Amount)
		}
		if d.Timestamp != "2025-10-01T09:00:00" {
			t.Errorf("Timestamp = %q, want '2025-10-01T09:00:00'", d.Timestamp)
		}
		if d.ID == "" {
			t.Error("ID is empty, want non-empty auto-generated ID")
		}
	})

	t.Run("auto_generates_sequential_unique_ids", func(t *testing.T) {
		e := newEngine(t)
		a := mustAddDonation(t, e, 100, "2025-10-01T08:00:00")
		b := mustAddDonation(t, e, 200, "2025-10-01T09:00:00")
		if a.ID == b.ID {
			t.Errorf("expected unique IDs, both got %q", a.ID)
		}
	})

	t.Run("zero_amount_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.AddDonation(0, "2025-10-01T09:00:00")
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("negative_amount_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.AddDonation(-50, "2025-10-01T09:00:00")
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})
}

func TestGetTotalRaised(t *testing.T) {
	t.Run("no_asof_returns_all_donations", func(t *testing.T) {
		e := newEngine(t)
		mustAddDonation(t, e, 400, "2025-10-01T09:00:00")
		mustAddDonation(t, e, 300, "2025-10-01T10:00:00")
		got := e.GetTotalRaised("")
		if got != 700 {
			t.Errorf("GetTotalRaised(\"\") = %v, want 700", got)
		}
	})

	t.Run("asof_filters_donations_at_or_before", func(t *testing.T) {
		e := newEngine(t)
		mustAddDonation(t, e, 400, "2025-10-01T09:00:00")
		mustAddDonation(t, e, 300, "2025-10-01T10:00:00")
		mustAddDonation(t, e, 500, "2025-10-01T13:00:00")
		got := e.GetTotalRaised("2025-10-01T12:00:00")
		if got != 700 {
			t.Errorf("GetTotalRaised(12:00) = %v, want 700", got)
		}
	})

	t.Run("returns_zero_when_no_donations", func(t *testing.T) {
		e := newEngine(t)
		got := e.GetTotalRaised("")
		if got != 0 {
			t.Errorf("GetTotalRaised(\"\") = %v, want 0", got)
		}
	})

	t.Run("asof_boundary_is_inclusive", func(t *testing.T) {
		e := newEngine(t)
		mustAddDonation(t, e, 400, "2025-10-01T09:00:00")
		mustAddDonation(t, e, 300, "2025-10-01T12:00:00") // exactly at asOf
		got := e.GetTotalRaised("2025-10-01T12:00:00")
		if got != 700 {
			t.Errorf("GetTotalRaised(12:00) = %v, want 700 (boundary inclusive)", got)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — ProcessDonation and challenge completion
// ---------------------------------------------------------------------------

func TestProcessDonation(t *testing.T) {
	t.Run("returns_the_donation_entry", func(t *testing.T) {
		e := seededEngine(t)
		r := mustProcess(t, e, 400, "2025-10-01T09:00:00")
		if r.Donation.Amount != 400 {
			t.Errorf("Donation.Amount = %v, want 400", r.Donation.Amount)
		}
		if r.Donation.Timestamp != "2025-10-01T09:00:00" {
			t.Errorf("Donation.Timestamp = %q, want '2025-10-01T09:00:00'", r.Donation.Timestamp)
		}
	})

	t.Run("newly_completed_empty_when_threshold_not_met", func(t *testing.T) {
		e := seededEngine(t)
		r := mustProcess(t, e, 400, "2025-10-01T09:00:00")
		if len(r.NewlyCompleted) != 0 {
			t.Errorf("NewlyCompleted length = %d, want 0", len(r.NewlyCompleted))
		}
	})

	t.Run("marks_challenge_complete_when_threshold_reached", func(t *testing.T) {
		e := seededEngine(t)
		mustProcess(t, e, 400, "2025-10-01T09:00:00")
		mustProcess(t, e, 300, "2025-10-01T10:00:00")
		r := mustProcess(t, e, 400, "2025-10-01T11:00:00") // total=1100 >= 1000
		if len(r.NewlyCompleted) != 1 {
			t.Fatalf("NewlyCompleted length = %d, want 1", len(r.NewlyCompleted))
		}
		if r.NewlyCompleted[0].ID != "c1" {
			t.Errorf("NewlyCompleted[0].ID = %q, want 'c1'", r.NewlyCompleted[0].ID)
		}
	})

	t.Run("sets_completed_at_to_donation_timestamp", func(t *testing.T) {
		e := seededEngine(t)
		mustProcess(t, e, 500, "2025-10-01T09:00:00")
		mustProcess(t, e, 600, "2025-10-01T10:00:00") // total=1100, c1 completes
		c1, err := e.GetChallenge("c1")
		if err != nil {
			t.Fatalf("GetChallenge: %v", err)
		}
		if !c1.Completed {
			t.Error("Completed = false, want true")
		}
		if c1.CompletedAt == nil || *c1.CompletedAt != "2025-10-01T10:00:00" {
			t.Errorf("CompletedAt = %v, want '2025-10-01T10:00:00'", c1.CompletedAt)
		}
	})

	t.Run("challenge_only_completes_once_idempotent", func(t *testing.T) {
		e := seededEngine(t)
		mustProcess(t, e, 600, "2025-10-01T09:00:00")
		r1 := mustProcess(t, e, 500, "2025-10-01T10:00:00") // total=1100, c1 completes
		r2 := mustProcess(t, e, 200, "2025-10-01T11:00:00") // c1 already done

		foundInR1 := false
		for _, c := range r1.NewlyCompleted {
			if c.ID == "c1" {
				foundInR1 = true
			}
		}
		if !foundInR1 {
			t.Error("c1 should appear in r1.NewlyCompleted")
		}

		foundInR2 := false
		for _, c := range r2.NewlyCompleted {
			if c.ID == "c1" {
				foundInR2 = true
			}
		}
		if foundInR2 {
			t.Error("c1 should NOT appear in r2.NewlyCompleted (already completed)")
		}
	})

	t.Run("donation_after_deadline_does_not_complete_challenge", func(t *testing.T) {
		// c1 deadline = 12:00; donation at 13:00 doesn't count for c1
		e := seededEngine(t)
		mustProcess(t, e, 900, "2025-10-01T11:00:00")
		r := mustProcess(t, e, 500, "2025-10-01T13:00:00") // after deadline
		for _, c := range r.NewlyCompleted {
			if c.ID == "c1" {
				t.Error("c1 should NOT complete from a donation after its deadline")
			}
		}
	})

	t.Run("multiple_challenges_can_complete_on_same_donation", func(t *testing.T) {
		e := newEngine(t)
		mustAddChallenge(t, e, "x1", "X1", 500, morningDeadline, BonusTypeFlat, 100)
		mustAddChallenge(t, e, "x2", "X2", 500, afternoonDeadline, BonusTypeFlat, 200)
		r := mustProcess(t, e, 600, "2025-10-01T10:00:00")
		if len(r.NewlyCompleted) != 2 {
			t.Errorf("NewlyCompleted length = %d, want 2", len(r.NewlyCompleted))
		}
	})

	t.Run("zero_amount_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.ProcessDonation(0, "2025-10-01T09:00:00")
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — GetHourlyBreakdown
// ---------------------------------------------------------------------------

func TestGetHourlyBreakdown(t *testing.T) {
	t.Run("returns_one_bucket_per_active_hour", func(t *testing.T) {
		e := hourlyEngine(t)
		buckets := e.GetHourlyBreakdown()
		hours := make(map[string]bool)
		for _, b := range buckets {
			hours[b.Hour] = true
		}
		for _, h := range []string{
			"2025-10-01T09:00:00",
			"2025-10-01T11:00:00",
			"2025-10-01T14:00:00",
			"2025-10-01T17:00:00",
		} {
			if !hours[h] {
				t.Errorf("expected bucket for hour %q", h)
			}
		}
		// hours with no donations should be absent
		if hours["2025-10-01T10:00:00"] {
			t.Error("unexpected bucket for hour 10:00")
		}
	})

	t.Run("ordered_by_hour_ascending", func(t *testing.T) {
		e := hourlyEngine(t)
		buckets := e.GetHourlyBreakdown()
		for i := 1; i < len(buckets); i++ {
			if buckets[i].Hour < buckets[i-1].Hour {
				t.Errorf("buckets not sorted: %q before %q", buckets[i-1].Hour, buckets[i].Hour)
			}
		}
	})

	t.Run("bucket_has_correct_donation_count_and_amount_raised", func(t *testing.T) {
		// 09:xx has two donations totalling 700
		e := hourlyEngine(t)
		buckets := e.GetHourlyBreakdown()
		nine := findBucket(buckets, "2025-10-01T09:00:00")
		if nine == nil {
			t.Fatal("no bucket for 09:00")
		}
		if nine.DonationCount != 2 {
			t.Errorf("DonationCount = %d, want 2", nine.DonationCount)
		}
		if nine.AmountRaised != 700 {
			t.Errorf("AmountRaised = %v, want 700", nine.AmountRaised)
		}
	})

	t.Run("cumulative_total_is_running_sum", func(t *testing.T) {
		e := hourlyEngine(t)
		buckets := e.GetHourlyBreakdown()

		nine := findBucket(buckets, "2025-10-01T09:00:00")
		eleven := findBucket(buckets, "2025-10-01T11:00:00")
		fourteen := findBucket(buckets, "2025-10-01T14:00:00")

		if nine == nil || eleven == nil || fourteen == nil {
			t.Fatal("missing expected bucket(s)")
		}
		if nine.CumulativeTotal != 700 {
			t.Errorf("09:00 CumulativeTotal = %v, want 700", nine.CumulativeTotal)
		}
		if eleven.CumulativeTotal != 1100 {
			t.Errorf("11:00 CumulativeTotal = %v, want 1100", eleven.CumulativeTotal)
		}
		if fourteen.CumulativeTotal != 1600 {
			t.Errorf("14:00 CumulativeTotal = %v, want 1600", fourteen.CumulativeTotal)
		}
	})

	t.Run("challenges_completed_lists_ids_that_completed_in_that_hour", func(t *testing.T) {
		e := hourlyEngine(t)
		buckets := e.GetHourlyBreakdown()

		eleven := findBucket(buckets, "2025-10-01T11:00:00")
		seventeen := findBucket(buckets, "2025-10-01T17:00:00")
		if eleven == nil || seventeen == nil {
			t.Fatal("missing expected bucket(s)")
		}

		foundC1 := false
		for _, id := range eleven.ChallengesCompleted {
			if id == "c1" {
				foundC1 = true
			}
		}
		if !foundC1 {
			t.Error("expected 'c1' in 11:00 ChallengesCompleted")
		}

		foundC2 := false
		for _, id := range seventeen.ChallengesCompleted {
			if id == "c2" {
				foundC2 = true
			}
		}
		if !foundC2 {
			t.Error("expected 'c2' in 17:00 ChallengesCompleted")
		}
	})

	t.Run("hours_with_no_completions_have_empty_slice", func(t *testing.T) {
		e := hourlyEngine(t)
		buckets := e.GetHourlyBreakdown()
		nine := findBucket(buckets, "2025-10-01T09:00:00")
		if nine == nil {
			t.Fatal("no bucket for 09:00")
		}
		if len(nine.ChallengesCompleted) != 0 {
			t.Errorf("ChallengesCompleted length = %d, want 0", len(nine.ChallengesCompleted))
		}
	})

	t.Run("returns_empty_slice_when_no_donations", func(t *testing.T) {
		e := newEngine(t)
		buckets := e.GetHourlyBreakdown()
		if len(buckets) != 0 {
			t.Errorf("GetHourlyBreakdown() length = %d, want 0", len(buckets))
		}
	})
}
