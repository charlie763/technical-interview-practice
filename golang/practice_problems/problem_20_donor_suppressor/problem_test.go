// Tests for Problem 20: Donor Communication Suppressor
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_20_donor_suppressor/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_20_donor_suppressor.go \
//	  -c go test -v .
package suppressor

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Shared test constants
// ---------------------------------------------------------------------------

const (
	asOf     = "2025-08-20T00:00:00"
	campaign = "camp-fall-2025"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newDS(t *testing.T) DonorSuppressor {
	t.Helper()
	return NewDonorSuppressor()
}

// seededPart2 returns a DonorSuppressor with:
//
//	alice@example.com — opted out
//	bob@example.com   — donated to `campaign` on 2025-08-01 (already_donated)
//	carol@example.com — donated to camp-spring-2025 on 2025-07-25 (26 days before asOf)
//	dave@example.com  — donated to camp-spring-2025 on 2025-06-01 (80 days before asOf)
func seededPart2(t *testing.T) DonorSuppressor {
	t.Helper()
	s := NewDonorSuppressor()
	s.OptOut("alice@example.com")
	s.RecordDonation("bob@example.com", campaign, "2025-08-01T10:00:00")
	s.RecordDonation("carol@example.com", "camp-spring-2025", "2025-07-25T09:00:00")
	s.RecordDonation("dave@example.com", "camp-spring-2025", "2025-06-01T09:00:00")
	return s
}

// seededPart3 returns a DonorSuppressor pre-loaded for bulk filter tests.
//
//	alice@example.com — opted out
//	bob@example.com   — donated to `campaign` on 2025-08-01
//	carol@example.com — donated to camp-other on 2025-08-10 (10 days before asOf, in 30-day window)
//	dave@example.com  — donated to camp-other on 2025-06-01 (80 days before asOf, outside window)
//	eve@example.com   — no history
func seededPart3(t *testing.T) DonorSuppressor {
	t.Helper()
	s := NewDonorSuppressor()
	s.OptOut("alice@example.com")
	s.RecordDonation("bob@example.com", campaign, "2025-08-01T00:00:00")
	s.RecordDonation("carol@example.com", "camp-other", "2025-08-10T00:00:00")
	s.RecordDonation("dave@example.com", "camp-other", "2025-06-01T00:00:00")
	return s
}

func containsEmail(emails []string, target string) bool {
	for _, e := range emails {
		if e == target {
			return true
		}
	}
	return false
}

func suppressedContains(results []*SuppressResult, email string) bool {
	for _, r := range results {
		if r.Email == email {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// PART 1 — Opt-out management and basic suppression check
// ---------------------------------------------------------------------------

func TestOptOut(t *testing.T) {
	t.Run("adds_email_to_opt_out_list", func(t *testing.T) {
		s := newDS(t)
		s.OptOut("alice@example.com")
		if !s.IsOptedOut("alice@example.com") {
			t.Error("expected alice to be opted out")
		}
	})

	t.Run("is_idempotent", func(t *testing.T) {
		s := newDS(t)
		s.OptOut("alice@example.com")
		s.OptOut("alice@example.com") // second call should not error or change state
		if !s.IsOptedOut("alice@example.com") {
			t.Error("expected alice to remain opted out")
		}
	})
}

func TestOptIn(t *testing.T) {
	t.Run("removes_email_from_opt_out_list", func(t *testing.T) {
		s := newDS(t)
		s.OptOut("alice@example.com")
		s.OptIn("alice@example.com")
		if s.IsOptedOut("alice@example.com") {
			t.Error("expected alice to not be opted out after opt-in")
		}
	})

	t.Run("is_idempotent_for_emails_not_opted_out", func(t *testing.T) {
		s := newDS(t)
		// should not panic
		s.OptIn("nobody@example.com")
		if s.IsOptedOut("nobody@example.com") {
			t.Error("expected nobody to not be opted out")
		}
	})
}

func TestIsOptedOut(t *testing.T) {
	t.Run("returns_false_for_unknown_email", func(t *testing.T) {
		s := newDS(t)
		if s.IsOptedOut("unknown@example.com") {
			t.Error("expected false for unknown email")
		}
	})
}

func TestCheckSuppressionPart1(t *testing.T) {
	t.Run("suppresses_opted_out_donors", func(t *testing.T) {
		s := newDS(t)
		s.OptOut("alice@example.com")
		result := s.CheckSuppression("alice@example.com", campaign, asOf, SuppressOptions{})
		if result.Email != "alice@example.com" {
			t.Errorf("Email = %q, want 'alice@example.com'", result.Email)
		}
		if !result.Suppressed {
			t.Error("expected Suppressed = true")
		}
		if result.Reason != SuppressReasonOptedOut {
			t.Errorf("Reason = %q, want %q", result.Reason, SuppressReasonOptedOut)
		}
	})

	t.Run("not_suppressed_for_eligible_donors", func(t *testing.T) {
		s := newDS(t)
		result := s.CheckSuppression("carol@example.com", campaign, asOf, SuppressOptions{})
		if result.Suppressed {
			t.Error("expected Suppressed = false")
		}
		if result.Reason != "" {
			t.Errorf("Reason = %q, want empty", result.Reason)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Donation-based suppression rules
// ---------------------------------------------------------------------------

func TestCheckSuppressionPart2(t *testing.T) {
	t.Run("already_donated_suppresses_donor_who_gave_to_this_campaign", func(t *testing.T) {
		s := seededPart2(t)
		result := s.CheckSuppression("bob@example.com", campaign, asOf, SuppressOptions{})
		if !result.Suppressed {
			t.Error("expected Suppressed = true")
		}
		if result.Reason != SuppressReasonAlreadyDonated {
			t.Errorf("Reason = %q, want %q", result.Reason, SuppressReasonAlreadyDonated)
		}
	})

	t.Run("already_donated_does_not_apply_to_different_campaign", func(t *testing.T) {
		s := seededPart2(t)
		result := s.CheckSuppression("bob@example.com", "camp-other", asOf, SuppressOptions{})
		if result.Suppressed {
			t.Error("expected Suppressed = false for different campaign")
		}
	})

	t.Run("recency_cooldown_suppresses_within_window", func(t *testing.T) {
		// carol gave to camp-spring on 2025-07-25; asOf=2025-08-20; 26 days ago → within 30-day window
		s := seededPart2(t)
		result := s.CheckSuppression("carol@example.com", campaign, asOf, SuppressOptions{CooldownDays: 30})
		if !result.Suppressed {
			t.Error("expected Suppressed = true (carol within 30-day cooldown)")
		}
		if result.Reason != SuppressReasonRecencyCooldown {
			t.Errorf("Reason = %q, want %q", result.Reason, SuppressReasonRecencyCooldown)
		}
	})

	t.Run("recency_cooldown_does_not_suppress_outside_window", func(t *testing.T) {
		// dave gave on 2025-06-01; asOf=2025-08-20; 80 days ago → outside 30-day window
		s := seededPart2(t)
		result := s.CheckSuppression("dave@example.com", campaign, asOf, SuppressOptions{CooldownDays: 30})
		if result.Suppressed {
			t.Error("expected Suppressed = false (dave outside 30-day cooldown)")
		}
	})

	t.Run("opted_out_takes_priority_over_already_donated", func(t *testing.T) {
		s := seededPart2(t)
		s.RecordDonation("alice@example.com", campaign, "2025-08-01T00:00:00")
		result := s.CheckSuppression("alice@example.com", campaign, asOf, SuppressOptions{CooldownDays: 30})
		if result.Reason != SuppressReasonOptedOut {
			t.Errorf("Reason = %q, want %q (opted_out has highest priority)", result.Reason, SuppressReasonOptedOut)
		}
	})

	t.Run("already_donated_takes_priority_over_recency_cooldown", func(t *testing.T) {
		// bob already donated to campaign and also is within cooldown window
		s := seededPart2(t)
		result := s.CheckSuppression("bob@example.com", campaign, asOf, SuppressOptions{CooldownDays: 30})
		if result.Reason != SuppressReasonAlreadyDonated {
			t.Errorf("Reason = %q, want %q (already_donated > recency_cooldown)", result.Reason, SuppressReasonAlreadyDonated)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Bulk campaign filtering and summary
// ---------------------------------------------------------------------------

var bulkEmails = []string{
	"alice@example.com",  // opted out
	"bob@example.com",    // donated to this campaign
	"carol@example.com",  // recent cooldown
	"dave@example.com",   // old donation → eligible
	"eve@example.com",    // no history → eligible
}

func TestFilterCampaignList(t *testing.T) {
	t.Run("separates_eligible_from_suppressed", func(t *testing.T) {
		s := seededPart3(t)
		result := s.FilterCampaignList(bulkEmails, campaign, asOf, SuppressOptions{CooldownDays: 30})
		if !containsEmail(result.Eligible, "dave@example.com") {
			t.Error("dave should be eligible")
		}
		if !containsEmail(result.Eligible, "eve@example.com") {
			t.Error("eve should be eligible")
		}
		if !suppressedContains(result.Suppressed, "alice@example.com") {
			t.Error("alice should be suppressed")
		}
		if !suppressedContains(result.Suppressed, "bob@example.com") {
			t.Error("bob should be suppressed")
		}
		if !suppressedContains(result.Suppressed, "carol@example.com") {
			t.Error("carol should be suppressed")
		}
	})

	t.Run("eligible_plus_suppressed_covers_all_emails", func(t *testing.T) {
		s := seededPart3(t)
		result := s.FilterCampaignList(bulkEmails, campaign, asOf, SuppressOptions{CooldownDays: 30})
		total := len(result.Eligible) + len(result.Suppressed)
		if total != len(bulkEmails) {
			t.Errorf("total coverage = %d, want %d", total, len(bulkEmails))
		}
	})
}

func TestGetSuppressedSummary(t *testing.T) {
	t.Run("returns_correct_counts", func(t *testing.T) {
		s := seededPart3(t)
		summary := s.GetSuppressedSummary(bulkEmails, campaign, asOf, SuppressOptions{CooldownDays: 30})
		if summary.Total != 5 {
			t.Errorf("Total = %d, want 5", summary.Total)
		}
		if summary.EligibleCount != 2 {
			t.Errorf("EligibleCount = %d, want 2 (dave + eve)", summary.EligibleCount)
		}
		if summary.SuppressedCount != 3 {
			t.Errorf("SuppressedCount = %d, want 3 (alice + bob + carol)", summary.SuppressedCount)
		}
	})

	t.Run("breaks_down_by_reason", func(t *testing.T) {
		s := seededPart3(t)
		summary := s.GetSuppressedSummary(bulkEmails, campaign, asOf, SuppressOptions{CooldownDays: 30})
		if summary.ByReason[SuppressReasonOptedOut] != 1 {
			t.Errorf("ByReason[opted_out] = %d, want 1", summary.ByReason[SuppressReasonOptedOut])
		}
		if summary.ByReason[SuppressReasonAlreadyDonated] != 1 {
			t.Errorf("ByReason[already_donated] = %d, want 1", summary.ByReason[SuppressReasonAlreadyDonated])
		}
		if summary.ByReason[SuppressReasonRecencyCooldown] != 1 {
			t.Errorf("ByReason[recency_cooldown] = %d, want 1", summary.ByReason[SuppressReasonRecencyCooldown])
		}
	})
}
