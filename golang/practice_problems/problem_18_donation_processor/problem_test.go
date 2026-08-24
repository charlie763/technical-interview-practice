// Tests for Problem 18: Donation Processor
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_18_donation_processor/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_18_donation_processor.go \
//	  -c go test -v .
package donations

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Shared timestamps
// ---------------------------------------------------------------------------

const (
	T1 = "2025-09-05T10:00:00"
	T2 = "2025-09-06T11:00:00"
	T3 = "2025-09-07T09:00:00"
	T4 = "2025-09-08T14:00:00"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newDP(t *testing.T) DonationProcessor {
	t.Helper()
	return NewDonationProcessor()
}

// seededDP returns a DonationProcessor with a campaign and several donations
// pre-loaded. alice total: 700, bob: 1000, carol: 750.
func seededDP(t *testing.T) (DonationProcessor, string, string) {
	t.Helper()
	dp := NewDonationProcessor()
	if _, err := dp.CreateCampaign("camp-seed", "Annual Fund", 100_000, "2025-09-01"); err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}
	d1, err := dp.AddDonation("camp-seed", "Alice Smith", "alice@x.com", 500, T1)
	if err != nil {
		t.Fatalf("AddDonation d1: %v", err)
	}
	d2, err := dp.AddDonation("camp-seed", "Bob Jones", "bob@x.com", 1_000, T2)
	if err != nil {
		t.Fatalf("AddDonation d2: %v", err)
	}
	if _, err := dp.AddDonation("camp-seed", "Alice Smith", "alice@x.com", 200, T3); err != nil {
		t.Fatalf("AddDonation alice2: %v", err)
	}
	if _, err := dp.AddDonation("camp-seed", "Carol Lee", "carol@x.com", 750, T4); err != nil {
		t.Fatalf("AddDonation carol: %v", err)
	}
	return dp, d1.ID, d2.ID
}

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PART 1 — Campaign setup, donation recording, and stats
// ---------------------------------------------------------------------------

func TestCreateCampaign(t *testing.T) {
	t.Run("returns_campaign_with_correct_fields", func(t *testing.T) {
		dp := newDP(t)
		c, err := dp.CreateCampaign("c1", "Annual Fund", 50_000, "2025-09-01")
		mustOK(t, err)
		if c.ID != "c1" {
			t.Errorf("ID = %q, want 'c1'", c.ID)
		}
		if c.Name != "Annual Fund" {
			t.Errorf("Name = %q, want 'Annual Fund'", c.Name)
		}
		if c.Goal != 50_000 {
			t.Errorf("Goal = %v, want 50000", c.Goal)
		}
		if c.StartDate != "2025-09-01" {
			t.Errorf("StartDate = %q, want '2025-09-01'", c.StartDate)
		}
	})

	t.Run("duplicate_id_returns_error", func(t *testing.T) {
		dp := newDP(t)
		mustOK(t, func() error { _, err := dp.CreateCampaign("c-dup", "A", 10_000, "2025-09-01"); return err }())
		_, err := dp.CreateCampaign("c-dup", "B", 20_000, "2025-09-01")
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestAddDonation(t *testing.T) {
	t.Run("returns_donation_with_correct_fields", func(t *testing.T) {
		dp := newDP(t)
		mustOK(t, func() error { _, err := dp.CreateCampaign("c1", "Fund", 10_000, "2025-09-01"); return err }())
		d, err := dp.AddDonation("c1", "Alice Smith", "alice@school.org", 500, T1)
		mustOK(t, err)
		if d.CampaignID != "c1" {
			t.Errorf("CampaignID = %q, want 'c1'", d.CampaignID)
		}
		if d.DonorName != "Alice Smith" {
			t.Errorf("DonorName = %q, want 'Alice Smith'", d.DonorName)
		}
		if d.DonorEmail != "alice@school.org" {
			t.Errorf("DonorEmail = %q, want 'alice@school.org'", d.DonorEmail)
		}
		if d.Amount != 500 {
			t.Errorf("Amount = %v, want 500", d.Amount)
		}
		if d.Timestamp != T1 {
			t.Errorf("Timestamp = %q, want %q", d.Timestamp, T1)
		}
		if d.Refunded {
			t.Error("Refunded should be false on new donation")
		}
	})

	t.Run("assigns_unique_auto_generated_ids", func(t *testing.T) {
		dp := newDP(t)
		mustOK(t, func() error { _, err := dp.CreateCampaign("c1", "Fund", 10_000, "2025-09-01"); return err }())
		d1, _ := dp.AddDonation("c1", "A", "a@x.com", 100, T1)
		d2, _ := dp.AddDonation("c1", "B", "b@x.com", 100, T2)
		if d1.ID == "" {
			t.Error("d1.ID should not be empty")
		}
		if d2.ID == "" {
			t.Error("d2.ID should not be empty")
		}
		if d1.ID == d2.ID {
			t.Errorf("IDs should be unique, both are %q", d1.ID)
		}
	})

	t.Run("unknown_campaign_returns_error", func(t *testing.T) {
		dp := newDP(t)
		_, err := dp.AddDonation("no-such", "A", "a@x.com", 100, T1)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("non_positive_amount_returns_error", func(t *testing.T) {
		dp := newDP(t)
		mustOK(t, func() error { _, err := dp.CreateCampaign("c1", "Fund", 10_000, "2025-09-01"); return err }())
		_, err := dp.AddDonation("c1", "A", "a@x.com", 0, T1)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount for amount=0, got %v", err)
		}
		_, err = dp.AddDonation("c1", "A", "a@x.com", -50, T1)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount for amount=-50, got %v", err)
		}
	})
}

func TestGetDonation(t *testing.T) {
	t.Run("returns_correct_donation", func(t *testing.T) {
		dp := newDP(t)
		mustOK(t, func() error { _, err := dp.CreateCampaign("c1", "Fund", 10_000, "2025-09-01"); return err }())
		d, _ := dp.AddDonation("c1", "Alice", "alice@x.com", 500, T1)
		got, err := dp.GetDonation(d.ID)
		mustOK(t, err)
		if got.ID != d.ID {
			t.Errorf("ID = %q, want %q", got.ID, d.ID)
		}
		if got.Amount != d.Amount {
			t.Errorf("Amount = %v, want %v", got.Amount, d.Amount)
		}
	})

	t.Run("unknown_id_returns_error", func(t *testing.T) {
		dp := newDP(t)
		_, err := dp.GetDonation("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetCampaignStats(t *testing.T) {
	t.Run("returns_correct_totals", func(t *testing.T) {
		dp := newDP(t)
		mustOK(t, func() error {
			_, err := dp.CreateCampaign("camp-seed", "Annual Fund", 100_000, "2025-09-01")
			return err
		}())
		dp.AddDonation("camp-seed", "Alice", "alice@x.com", 500, T1)
		dp.AddDonation("camp-seed", "Bob", "bob@x.com", 1_000, T2)
		dp.AddDonation("camp-seed", "Alice", "alice@x.com", 200, T3)
		dp.AddDonation("camp-seed", "Carol", "carol@x.com", 750, T4)
		stats, err := dp.GetCampaignStats("camp-seed")
		mustOK(t, err)
		if stats.TotalRaised != 2_450 {
			t.Errorf("TotalRaised = %v, want 2450", stats.TotalRaised)
		}
		if stats.DonationCount != 4 {
			t.Errorf("DonationCount = %v, want 4", stats.DonationCount)
		}
		if stats.DonorCount != 3 {
			t.Errorf("DonorCount = %v, want 3 (alice, bob, carol)", stats.DonorCount)
		}
		if stats.GoalPercent != 2.5 {
			t.Errorf("GoalPercent = %v, want 2.5", stats.GoalPercent)
		}
	})

	t.Run("empty_campaign_returns_zeros", func(t *testing.T) {
		dp := newDP(t)
		mustOK(t, func() error { _, err := dp.CreateCampaign("empty", "Empty", 10_000, "2025-09-01"); return err }())
		stats, err := dp.GetCampaignStats("empty")
		mustOK(t, err)
		if stats.TotalRaised != 0 || stats.DonorCount != 0 || stats.DonationCount != 0 || stats.GoalPercent != 0 {
			t.Errorf("expected all zeros, got %+v", stats)
		}
	})

	t.Run("unknown_campaign_returns_error", func(t *testing.T) {
		dp := newDP(t)
		_, err := dp.GetCampaignStats("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Top donors and date-range search
// ---------------------------------------------------------------------------

func TestGetTopDonors(t *testing.T) {
	t.Run("sorted_by_total_amount_descending", func(t *testing.T) {
		dp, _, _ := seededDP(t)
		top, err := dp.GetTopDonors("camp-seed", 3)
		mustOK(t, err)
		if len(top) != 3 {
			t.Fatalf("len = %d, want 3", len(top))
		}
		if top[0].DonorEmail != "bob@x.com" {
			t.Errorf("[0].DonorEmail = %q, want 'bob@x.com'", top[0].DonorEmail)
		}
		if top[0].TotalAmount != 1_000 {
			t.Errorf("[0].TotalAmount = %v, want 1000", top[0].TotalAmount)
		}
		if top[1].DonorEmail != "carol@x.com" {
			t.Errorf("[1].DonorEmail = %q, want 'carol@x.com'", top[1].DonorEmail)
		}
		if top[2].DonorEmail != "alice@x.com" {
			t.Errorf("[2].DonorEmail = %q, want 'alice@x.com'", top[2].DonorEmail)
		}
	})

	t.Run("aggregates_multiple_donations_from_same_email", func(t *testing.T) {
		dp, _, _ := seededDP(t)
		top, _ := dp.GetTopDonors("camp-seed", 5)
		var alice *DonorSummary
		for _, d := range top {
			if d.DonorEmail == "alice@x.com" {
				alice = d
				break
			}
		}
		if alice == nil {
			t.Fatal("alice not found in top donors")
		}
		if alice.TotalAmount != 700 {
			t.Errorf("alice.TotalAmount = %v, want 700", alice.TotalAmount)
		}
		if alice.DonationCount != 2 {
			t.Errorf("alice.DonationCount = %v, want 2", alice.DonationCount)
		}
	})

	t.Run("respects_limit", func(t *testing.T) {
		dp, _, _ := seededDP(t)
		top, _ := dp.GetTopDonors("camp-seed", 2)
		if len(top) != 2 {
			t.Errorf("len = %d, want 2", len(top))
		}
	})

	t.Run("unknown_campaign_returns_error", func(t *testing.T) {
		dp := newDP(t)
		_, err := dp.GetTopDonors("no-such", 5)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetDonationsInRange(t *testing.T) {
	t.Run("returns_donations_within_inclusive_range", func(t *testing.T) {
		dp, _, _ := seededDP(t)
		inRange, err := dp.GetDonationsInRange("camp-seed", T1, T3)
		mustOK(t, err)
		if len(inRange) != 3 {
			t.Fatalf("len = %d, want 3", len(inRange))
		}
	})

	t.Run("excludes_donations_outside_range", func(t *testing.T) {
		dp, _, _ := seededDP(t)
		inRange, _ := dp.GetDonationsInRange("camp-seed", T1, T2)
		for _, d := range inRange {
			if d.Timestamp > T2 || d.Timestamp < T1 {
				t.Errorf("donation %q timestamp %q outside range [%s, %s]", d.ID, d.Timestamp, T1, T2)
			}
		}
	})

	t.Run("sorted_by_timestamp_ascending", func(t *testing.T) {
		dp, _, _ := seededDP(t)
		inRange, _ := dp.GetDonationsInRange("camp-seed", T1, T4)
		for i := 1; i < len(inRange); i++ {
			if inRange[i].Timestamp < inRange[i-1].Timestamp {
				t.Errorf("not sorted: %q before %q", inRange[i-1].Timestamp, inRange[i].Timestamp)
			}
		}
	})

	t.Run("unknown_campaign_returns_error", func(t *testing.T) {
		dp := newDP(t)
		_, err := dp.GetDonationsInRange("no-such", T1, T4)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Refunds and matching gifts
// ---------------------------------------------------------------------------

func TestRefundDonation(t *testing.T) {
	t.Run("marks_donation_as_refunded", func(t *testing.T) {
		dp, d1ID, _ := seededDP(t)
		refunded, err := dp.RefundDonation(d1ID)
		mustOK(t, err)
		if !refunded.Refunded {
			t.Error("expected Refunded = true")
		}
	})

	t.Run("excludes_refunded_donation_from_stats", func(t *testing.T) {
		dp, d1ID, _ := seededDP(t)
		dp.RefundDonation(d1ID) // alice's 500 donation
		stats, _ := dp.GetCampaignStats("camp-seed")
		// remaining: bob 1000, alice 200, carol 750 = 1950
		if stats.TotalRaised != 1_950 {
			t.Errorf("TotalRaised = %v, want 1950", stats.TotalRaised)
		}
	})

	t.Run("unknown_id_returns_error", func(t *testing.T) {
		dp := newDP(t)
		_, err := dp.RefundDonation("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("already_refunded_returns_error", func(t *testing.T) {
		dp, d1ID, _ := seededDP(t)
		mustOK(t, func() error { _, err := dp.RefundDonation(d1ID); return err }())
		_, err := dp.RefundDonation(d1ID)
		if !errors.Is(err, ErrAlreadyRefunded) {
			t.Errorf("expected ErrAlreadyRefunded, got %v", err)
		}
	})
}

func TestAddMatchingGift(t *testing.T) {
	t.Run("creates_donation_with_min_amount_times_ratio_max", func(t *testing.T) {
		// source: 500, ratio: 0.5, max: 400 → match = min(250, 400) = 250
		dp, d1ID, _ := seededDP(t)
		match, err := dp.AddMatchingGift(d1ID, "Corp Fund", "corp@x.com", 0.5, 400)
		mustOK(t, err)
		if match.Amount != 250 {
			t.Errorf("Amount = %v, want 250", match.Amount)
		}
		if match.DonorName != "Corp Fund" {
			t.Errorf("DonorName = %q, want 'Corp Fund'", match.DonorName)
		}
		if match.DonorEmail != "corp@x.com" {
			t.Errorf("DonorEmail = %q, want 'corp@x.com'", match.DonorEmail)
		}
		if match.CampaignID != "camp-seed" {
			t.Errorf("CampaignID = %q, want 'camp-seed'", match.CampaignID)
		}
	})

	t.Run("respects_max_amount_cap", func(t *testing.T) {
		// source: 1000, ratio: 1.0, max: 300 → match = min(1000, 300) = 300
		dp, _, d2ID := seededDP(t)
		match, err := dp.AddMatchingGift(d2ID, "Matcher", "m@x.com", 1.0, 300)
		mustOK(t, err)
		if match.Amount != 300 {
			t.Errorf("Amount = %v, want 300", match.Amount)
		}
	})

	t.Run("matched_amount_appears_in_stats", func(t *testing.T) {
		dp, d1ID, _ := seededDP(t)
		dp.AddMatchingGift(d1ID, "Corp", "corp@x.com", 0.5, 400) // +250
		stats, _ := dp.GetCampaignStats("camp-seed")
		// 500 + 1000 + 200 + 750 + 250 = 2700
		if stats.TotalRaised != 2_700 {
			t.Errorf("TotalRaised = %v, want 2700", stats.TotalRaised)
		}
	})

	t.Run("source_refunded_returns_error", func(t *testing.T) {
		dp, d1ID, _ := seededDP(t)
		mustOK(t, func() error { _, err := dp.RefundDonation(d1ID); return err }())
		_, err := dp.AddMatchingGift(d1ID, "Corp", "c@x.com", 0.5, 400)
		if !errors.Is(err, ErrSourceRefunded) {
			t.Errorf("expected ErrSourceRefunded, got %v", err)
		}
	})

	t.Run("unknown_source_donation_returns_error", func(t *testing.T) {
		dp := newDP(t)
		_, err := dp.AddMatchingGift("no-such", "Corp", "c@x.com", 0.5, 400)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
