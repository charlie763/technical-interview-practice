// Tests for Problem 19: Walkathon Pledge Tracker
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_19_pledge_tracker/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_19_pledge_tracker.go \
//	  -c go test -v .
package pledges

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newPT(t *testing.T) PledgeTracker {
	t.Helper()
	return NewPledgeTracker()
}

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// seededPart2 returns a PledgeTracker pre-loaded for Part 2 tests.
//
//	p1 (class-5a): $5/lap + $50 flat
//	p2 (class-5a): $10/lap
//	p3 (class-6b): $100 flat
func seededPart2(t *testing.T) PledgeTracker {
	t.Helper()
	pt := NewPledgeTracker()
	mustOK(t, func() error { _, err := pt.RegisterParticipant("p1", "Emma Torres", "class-5a"); return err }())
	mustOK(t, func() error { _, err := pt.AddPledge("p1", "Grandma Rose", "rose@example.com", PledgeTypePerLap, 5); return err }())
	mustOK(t, func() error { _, err := pt.AddPledge("p1", "Uncle Joe", "joe@example.com", PledgeTypeFlat, 50); return err }())
	mustOK(t, func() error { _, err := pt.RegisterParticipant("p2", "Liam Chen", "class-5a"); return err }())
	mustOK(t, func() error { _, err := pt.AddPledge("p2", "Dad", "dad@example.com", PledgeTypePerLap, 10); return err }())
	mustOK(t, func() error { _, err := pt.RegisterParticipant("p3", "Sofia Park", "class-6b"); return err }())
	mustOK(t, func() error { _, err := pt.AddPledge("p3", "Family Friend", "ff@example.com", PledgeTypeFlat, 100); return err }())
	return pt
}

// seededPart3 returns a PledgeTracker pre-loaded with laps recorded for Part 3.
//
//	p1 (class-5a): 12 laps × $5 + $50 flat = 110
//	p2 (class-5a): 8  laps × $10          =  80
//	p3 (class-6b): 1  lap  + $100 flat    = 100
//	class-5a total: 190, class-6b total: 100
func seededPart3(t *testing.T) PledgeTracker {
	t.Helper()
	pt := NewPledgeTracker()
	mustOK(t, func() error { _, err := pt.RegisterParticipant("p1", "Emma Torres", "class-5a"); return err }())
	mustOK(t, func() error { _, err := pt.AddPledge("p1", "Grandma", "g@x.com", PledgeTypePerLap, 5); return err }())
	mustOK(t, func() error { _, err := pt.AddPledge("p1", "Uncle", "u@x.com", PledgeTypeFlat, 50); return err }())
	mustOK(t, func() error { _, err := pt.RecordLaps("p1", 12); return err }()) // 60+50=110

	mustOK(t, func() error { _, err := pt.RegisterParticipant("p2", "Liam Chen", "class-5a"); return err }())
	mustOK(t, func() error { _, err := pt.AddPledge("p2", "Dad", "d@x.com", PledgeTypePerLap, 10); return err }())
	mustOK(t, func() error { _, err := pt.RecordLaps("p2", 8); return err }()) // 80

	mustOK(t, func() error { _, err := pt.RegisterParticipant("p3", "Sofia Park", "class-6b"); return err }())
	mustOK(t, func() error { _, err := pt.AddPledge("p3", "Friend", "f@x.com", PledgeTypeFlat, 100); return err }())
	mustOK(t, func() error { _, err := pt.RecordLaps("p3", 1); return err }()) // 100
	return pt
}

// ---------------------------------------------------------------------------
// PART 1 — Participant registration and pledge collection
// ---------------------------------------------------------------------------

func TestRegisterParticipant(t *testing.T) {
	t.Run("returns_object_with_correct_fields", func(t *testing.T) {
		pt := newPT(t)
		p, err := pt.RegisterParticipant("p1", "Emma Torres", "class-5a")
		mustOK(t, err)
		if p.ID != "p1" {
			t.Errorf("ID = %q, want 'p1'", p.ID)
		}
		if p.Name != "Emma Torres" {
			t.Errorf("Name = %q, want 'Emma Torres'", p.Name)
		}
		if p.ClassID != "class-5a" {
			t.Errorf("ClassID = %q, want 'class-5a'", p.ClassID)
		}
		if p.LapsCompleted != 0 {
			t.Errorf("LapsCompleted = %d, want 0", p.LapsCompleted)
		}
	})

	t.Run("duplicate_id_returns_error", func(t *testing.T) {
		pt := newPT(t)
		mustOK(t, func() error { _, err := pt.RegisterParticipant("p-dup", "A", "class-1"); return err }())
		_, err := pt.RegisterParticipant("p-dup", "B", "class-1")
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestAddPledge(t *testing.T) {
	t.Run("returns_object_with_correct_fields", func(t *testing.T) {
		pt := newPT(t)
		mustOK(t, func() error { _, err := pt.RegisterParticipant("p1", "Emma", "class-5a"); return err }())
		pledge, err := pt.AddPledge("p1", "Grandma Rose", "rose@example.com", PledgeTypePerLap, 5)
		mustOK(t, err)
		if pledge.ParticipantID != "p1" {
			t.Errorf("ParticipantID = %q, want 'p1'", pledge.ParticipantID)
		}
		if pledge.SponsorName != "Grandma Rose" {
			t.Errorf("SponsorName = %q, want 'Grandma Rose'", pledge.SponsorName)
		}
		if pledge.SponsorEmail != "rose@example.com" {
			t.Errorf("SponsorEmail = %q, want 'rose@example.com'", pledge.SponsorEmail)
		}
		if pledge.PledgeType != PledgeTypePerLap {
			t.Errorf("PledgeType = %q, want 'per_lap'", pledge.PledgeType)
		}
		if pledge.Amount != 5 {
			t.Errorf("Amount = %v, want 5", pledge.Amount)
		}
		if pledge.ID == "" {
			t.Error("ID should not be empty")
		}
	})

	t.Run("assigns_unique_auto_generated_ids", func(t *testing.T) {
		pt := newPT(t)
		mustOK(t, func() error { _, err := pt.RegisterParticipant("p1", "Emma", "class-5a"); return err }())
		a, _ := pt.AddPledge("p1", "A", "a@x.com", PledgeTypeFlat, 10)
		b, _ := pt.AddPledge("p1", "B", "b@x.com", PledgeTypeFlat, 20)
		if a.ID == b.ID {
			t.Errorf("IDs should be unique, both are %q", a.ID)
		}
	})

	t.Run("unknown_participant_returns_error", func(t *testing.T) {
		pt := newPT(t)
		_, err := pt.AddPledge("no-such", "A", "a@x.com", PledgeTypeFlat, 10)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("non_positive_amount_returns_error", func(t *testing.T) {
		pt := newPT(t)
		mustOK(t, func() error { _, err := pt.RegisterParticipant("p1", "Emma", "class-5a"); return err }())
		_, err := pt.AddPledge("p1", "A", "a@x.com", PledgeTypeFlat, 0)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount for amount=0, got %v", err)
		}
		_, err = pt.AddPledge("p1", "A", "a@x.com", PledgeTypeFlat, -5)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount for amount=-5, got %v", err)
		}
	})
}

func TestGetParticipant(t *testing.T) {
	t.Run("returns_correct_participant", func(t *testing.T) {
		pt := newPT(t)
		mustOK(t, func() error { _, err := pt.RegisterParticipant("p1", "Emma", "class-5a"); return err }())
		p, err := pt.GetParticipant("p1")
		mustOK(t, err)
		if p.ID != "p1" {
			t.Errorf("ID = %q, want 'p1'", p.ID)
		}
	})

	t.Run("unknown_id_returns_error", func(t *testing.T) {
		pt := newPT(t)
		_, err := pt.GetParticipant("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Lap recording and totals
// ---------------------------------------------------------------------------

func TestRecordLaps(t *testing.T) {
	t.Run("updates_laps_completed", func(t *testing.T) {
		pt := seededPart2(t)
		p, err := pt.RecordLaps("p1", 10)
		mustOK(t, err)
		if p.LapsCompleted != 10 {
			t.Errorf("LapsCompleted = %d, want 10", p.LapsCompleted)
		}
	})

	t.Run("unknown_participant_returns_error", func(t *testing.T) {
		pt := newPT(t)
		_, err := pt.RecordLaps("no-such", 5)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("negative_laps_returns_error", func(t *testing.T) {
		pt := seededPart2(t)
		_, err := pt.RecordLaps("p1", -1)
		if !errors.Is(err, ErrInvalidLaps) {
			t.Errorf("expected ErrInvalidLaps, got %v", err)
		}
	})
}

func TestGetParticipantTotal(t *testing.T) {
	t.Run("computes_per_lap_flat_and_total", func(t *testing.T) {
		pt := seededPart2(t)
		mustOK(t, func() error { _, err := pt.RecordLaps("p1", 10); return err }())
		total, err := pt.GetParticipantTotal("p1")
		mustOK(t, err)
		if total.PerLapTotal != 50 {
			t.Errorf("PerLapTotal = %v, want 50 (10 × $5)", total.PerLapTotal)
		}
		if total.FlatTotal != 50 {
			t.Errorf("FlatTotal = %v, want 50", total.FlatTotal)
		}
		if total.TotalRaised != 100 {
			t.Errorf("TotalRaised = %v, want 100", total.TotalRaised)
		}
		if total.LapsCompleted != 10 {
			t.Errorf("LapsCompleted = %v, want 10", total.LapsCompleted)
		}
		if total.ParticipantID != "p1" {
			t.Errorf("ParticipantID = %q, want 'p1'", total.ParticipantID)
		}
		if total.Name != "Emma Torres" {
			t.Errorf("Name = %q, want 'Emma Torres'", total.Name)
		}
	})

	t.Run("per_lap_is_zero_when_laps_zero", func(t *testing.T) {
		pt := seededPart2(t)
		total, _ := pt.GetParticipantTotal("p1")
		if total.PerLapTotal != 0 {
			t.Errorf("PerLapTotal = %v, want 0 (0 laps)", total.PerLapTotal)
		}
		if total.TotalRaised != 50 {
			t.Errorf("TotalRaised = %v, want 50 (flat only)", total.TotalRaised)
		}
	})

	t.Run("handles_only_flat_pledges", func(t *testing.T) {
		pt := seededPart2(t)
		mustOK(t, func() error { _, err := pt.RecordLaps("p3", 15); return err }())
		total, err := pt.GetParticipantTotal("p3")
		mustOK(t, err)
		if total.PerLapTotal != 0 {
			t.Errorf("PerLapTotal = %v, want 0", total.PerLapTotal)
		}
		if total.FlatTotal != 100 {
			t.Errorf("FlatTotal = %v, want 100", total.FlatTotal)
		}
		if total.TotalRaised != 100 {
			t.Errorf("TotalRaised = %v, want 100", total.TotalRaised)
		}
	})

	t.Run("unknown_participant_returns_error", func(t *testing.T) {
		pt := newPT(t)
		_, err := pt.GetParticipantTotal("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetCampaignTotal(t *testing.T) {
	t.Run("returns_sum_across_all_participants", func(t *testing.T) {
		pt := seededPart2(t)
		pt.RecordLaps("p1", 12) // 12×5 + 50 = 110
		pt.RecordLaps("p2", 8)  // 8×10 = 80
		pt.RecordLaps("p3", 15) // 100 flat
		got := pt.GetCampaignTotal()
		if got != 290 {
			t.Errorf("GetCampaignTotal = %v, want 290", got)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Class leaderboard and participant ranking
// ---------------------------------------------------------------------------

func TestGetClassLeaderboard(t *testing.T) {
	t.Run("sorts_classes_by_total_raised_descending", func(t *testing.T) {
		pt := seededPart3(t)
		lb, err := pt.GetClassLeaderboard()
		mustOK(t, err)
		if len(lb) != 2 {
			t.Fatalf("len = %d, want 2", len(lb))
		}
		if lb[0].ClassID != "class-5a" {
			t.Errorf("[0].ClassID = %q, want 'class-5a'", lb[0].ClassID)
		}
		if lb[0].TotalRaised != 190 {
			t.Errorf("[0].TotalRaised = %v, want 190", lb[0].TotalRaised)
		}
		if lb[1].ClassID != "class-6b" {
			t.Errorf("[1].ClassID = %q, want 'class-6b'", lb[1].ClassID)
		}
		if lb[1].TotalRaised != 100 {
			t.Errorf("[1].TotalRaised = %v, want 100", lb[1].TotalRaised)
		}
	})

	t.Run("includes_participant_count", func(t *testing.T) {
		pt := seededPart3(t)
		lb, _ := pt.GetClassLeaderboard()
		var a *ClassEntry
		for _, c := range lb {
			if c.ClassID == "class-5a" {
				a = c
				break
			}
		}
		if a == nil {
			t.Fatal("class-5a not found in leaderboard")
		}
		if a.ParticipantCount != 2 {
			t.Errorf("ParticipantCount = %d, want 2", a.ParticipantCount)
		}
	})

	t.Run("includes_top_fundraiser", func(t *testing.T) {
		pt := seededPart3(t)
		lb, _ := pt.GetClassLeaderboard()
		var a *ClassEntry
		for _, c := range lb {
			if c.ClassID == "class-5a" {
				a = c
				break
			}
		}
		if a == nil {
			t.Fatal("class-5a not found in leaderboard")
		}
		if a.TopFundraiser.ParticipantID != "p1" {
			t.Errorf("TopFundraiser.ParticipantID = %q, want 'p1'", a.TopFundraiser.ParticipantID)
		}
		if a.TopFundraiser.TotalRaised != 110 {
			t.Errorf("TopFundraiser.TotalRaised = %v, want 110", a.TopFundraiser.TotalRaised)
		}
	})
}

func TestGetParticipantLeaderboard(t *testing.T) {
	t.Run("ranks_by_total_raised_descending", func(t *testing.T) {
		pt := seededPart3(t)
		lb, err := pt.GetParticipantLeaderboard(0) // 0 = no limit
		mustOK(t, err)
		if lb[0].ParticipantID != "p1" {
			t.Errorf("[0].ParticipantID = %q, want 'p1'", lb[0].ParticipantID)
		}
		if lb[0].Rank != 1 {
			t.Errorf("[0].Rank = %d, want 1", lb[0].Rank)
		}
		if lb[0].TotalRaised != 110 {
			t.Errorf("[0].TotalRaised = %v, want 110", lb[0].TotalRaised)
		}
		if lb[1].Rank != 2 {
			t.Errorf("[1].Rank = %d, want 2", lb[1].Rank)
		}
		if lb[2].Rank != 3 {
			t.Errorf("[2].Rank = %d, want 3", lb[2].Rank)
		}
	})

	t.Run("respects_limit", func(t *testing.T) {
		pt := seededPart3(t)
		lb, _ := pt.GetParticipantLeaderboard(2)
		if len(lb) != 2 {
			t.Errorf("len = %d, want 2", len(lb))
		}
	})

	t.Run("returns_all_when_limit_is_zero", func(t *testing.T) {
		pt := seededPart3(t)
		lb, _ := pt.GetParticipantLeaderboard(0)
		if len(lb) != 3 {
			t.Errorf("len = %d, want 3", len(lb))
		}
	})
}
