// Tests for Problem 17: Claims Processing Pipeline
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_17_claims_pipeline/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_17_claims_pipeline.go \
//	  -c go test -v .
package claims

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Shared timestamps
// ---------------------------------------------------------------------------

const (
	D0 = "2025-01-15"          // incident date
	T0 = "2025-02-01T09:00:00" // filed
	T1 = "2025-02-03T10:00:00" // investigating
	T2 = "2025-02-10T14:00:00" // evaluation
	T3 = "2025-02-20T11:00:00" // settled / denied
	T4 = "2025-03-01T09:00:00" // closed
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newPipeline(t *testing.T) ClaimsPipeline {
	t.Helper()
	return NewClaimsPipeline()
}

// seededPipeline returns a ClaimsPipeline pre-loaded with four claims:
//
//	clm-001  pol-101  epl   claimed=75_000  reserve=50_000  → settled (approved=60_000)
//	clm-002  pol-101  epl   claimed=30_000  reserve=30_000  → investigating
//	clm-003  pol-102  do    claimed=200_000 reserve=150_000 → evaluation
//	clm-004  pol-103  epl   claimed=10_000  reserve=10_000  → denied
func seededPipeline(t *testing.T) ClaimsPipeline {
	t.Helper()
	p := NewClaimsPipeline()

	// clm-001: filed → investigating → evaluation → settled
	mustFile(t, p, "clm-001", "pol-101", "epl", D0, T0, 75_000, 50_000, "adj-1")
	mustAdvance(t, p, "clm-001", "investigating", T1, "adj-1")
	mustAdvance(t, p, "clm-001", "evaluation", T2, "adj-1")
	mustSettle(t, p, "clm-001", 60_000, T3, "adj-1")

	// clm-002: filed → investigating
	mustFile(t, p, "clm-002", "pol-101", "epl", D0, T0, 30_000, 30_000, "adj-2")
	mustAdvance(t, p, "clm-002", "investigating", T1, "adj-2")

	// clm-003: filed → investigating → evaluation
	mustFile(t, p, "clm-003", "pol-102", "do", D0, T0, 200_000, 150_000, "adj-1")
	mustAdvance(t, p, "clm-003", "investigating", T1, "adj-1")
	mustAdvance(t, p, "clm-003", "evaluation", T2, "adj-1")

	// clm-004: filed → investigating → evaluation → denied
	mustFile(t, p, "clm-004", "pol-103", "epl", D0, T0, 10_000, 10_000, "adj-3")
	mustAdvance(t, p, "clm-004", "investigating", T1, "adj-3")
	mustAdvance(t, p, "clm-004", "evaluation", T2, "adj-3")
	mustDeny(t, p, "clm-004", "Coverage exclusion applies.", T3, "adj-3")

	return p
}

func mustFile(t *testing.T, p ClaimsPipeline, claimID, policyID, ct, incidentDate, filedAt string, claimed, reserve int, actor string) {
	t.Helper()
	if _, err := p.FileClaim(claimID, policyID, ct, incidentDate, filedAt, claimed, reserve, actor); err != nil {
		t.Fatalf("FileClaim(%q): %v", claimID, err)
	}
}

func mustAdvance(t *testing.T, p ClaimsPipeline, claimID, toStatus, at, actor string) {
	t.Helper()
	if _, err := p.AdvanceStatus(claimID, toStatus, at, actor); err != nil {
		t.Fatalf("AdvanceStatus(%q → %q): %v", claimID, toStatus, err)
	}
}

func mustSettle(t *testing.T, p ClaimsPipeline, claimID string, approved int, at, actor string) {
	t.Helper()
	if _, err := p.SettleClaim(claimID, approved, at, actor); err != nil {
		t.Fatalf("SettleClaim(%q): %v", claimID, err)
	}
}

func mustDeny(t *testing.T, p ClaimsPipeline, claimID, reason, at, actor string) {
	t.Helper()
	if _, err := p.DenyClaim(claimID, reason, at, actor); err != nil {
		t.Fatalf("DenyClaim(%q): %v", claimID, err)
	}
}

// ---------------------------------------------------------------------------
// PART 1 — Claim filing, status transitions, reserve updates
// ---------------------------------------------------------------------------

func TestFileClaim(t *testing.T) {
	t.Run("returns_claim_in_filed_state", func(t *testing.T) {
		p := newPipeline(t)
		c, err := p.FileClaim("clm-new", "pol-200", "do", D0, T0, 50_000, 40_000, "adj-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ClaimID != "clm-new" {
			t.Errorf("ClaimID = %q, want 'clm-new'", c.ClaimID)
		}
		if c.Status != "filed" {
			t.Errorf("Status = %q, want 'filed'", c.Status)
		}
		if c.ApprovedAmount != nil {
			t.Errorf("ApprovedAmount = %v, want nil", c.ApprovedAmount)
		}
		if c.ClaimedAmount != 50_000 {
			t.Errorf("ClaimedAmount = %d, want 50000", c.ClaimedAmount)
		}
		if c.ReserveAmount != 40_000 {
			t.Errorf("ReserveAmount = %d, want 40000", c.ReserveAmount)
		}
	})

	t.Run("initial_event_recorded", func(t *testing.T) {
		p := newPipeline(t)
		mustFile(t, p, "clm-evt", "pol-200", "epl", D0, T0, 20_000, 15_000, "adj-1")
		c, _ := p.GetClaim("clm-evt")
		if len(c.Events) != 1 {
			t.Fatalf("len(Events) = %d, want 1", len(c.Events))
		}
		if c.Events[0].Action != "filed" {
			t.Errorf("Events[0].Action = %q, want 'filed'", c.Events[0].Action)
		}
		if c.Events[0].Actor != "adj-1" {
			t.Errorf("Events[0].Actor = %q, want 'adj-1'", c.Events[0].Actor)
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		p := newPipeline(t)
		mustFile(t, p, "clm-dup", "pol-200", "do", D0, T0, 10_000, 10_000, "adj-1")
		_, err := p.FileClaim("clm-dup", "pol-201", "do", D0, T1, 5_000, 5_000, "adj-1")
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("zero_claimed_amount_returns_error", func(t *testing.T) {
		p := newPipeline(t)
		_, err := p.FileClaim("clm-zero", "pol-200", "epl", D0, T0, 0, 10_000, "adj-1")
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("zero_reserve_amount_returns_error", func(t *testing.T) {
		p := newPipeline(t)
		_, err := p.FileClaim("clm-zero-r", "pol-200", "epl", D0, T0, 10_000, 0, "adj-1")
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})
}

func TestGetClaim(t *testing.T) {
	t.Run("returns_existing_claim", func(t *testing.T) {
		p := seededPipeline(t)
		c, err := p.GetClaim("clm-001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ClaimID != "clm-001" {
			t.Errorf("ClaimID = %q, want 'clm-001'", c.ClaimID)
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		_, err := p.GetClaim("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestAdvanceStatus(t *testing.T) {
	t.Run("valid_transition_updates_status", func(t *testing.T) {
		p := newPipeline(t)
		mustFile(t, p, "clm-adv", "pol-200", "epl", D0, T0, 20_000, 15_000, "adj-1")
		mustAdvance(t, p, "clm-adv", "investigating", T1, "adj-1")
		c, _ := p.GetClaim("clm-adv")
		if c.Status != "investigating" {
			t.Errorf("Status = %q, want 'investigating'", c.Status)
		}
	})

	t.Run("invalid_transition_returns_error", func(t *testing.T) {
		p := newPipeline(t)
		mustFile(t, p, "clm-inv", "pol-200", "epl", D0, T0, 20_000, 15_000, "adj-1")
		_, err := p.AdvanceStatus("clm-inv", "evaluation", T1, "adj-1")
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("settling_via_advance_status_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		_, err := p.AdvanceStatus("clm-003", "settled", T3, "adj-1")
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition for 'settled', got %v", err)
		}
	})

	t.Run("denying_via_advance_status_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		_, err := p.AdvanceStatus("clm-003", "denied", T3, "adj-1")
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition for 'denied', got %v", err)
		}
	})

	t.Run("appends_status_change_event", func(t *testing.T) {
		p := newPipeline(t)
		mustFile(t, p, "clm-ev2", "pol-200", "epl", D0, T0, 20_000, 15_000, "adj-1")
		mustAdvance(t, p, "clm-ev2", "investigating", T1, "adj-2")
		c, _ := p.GetClaim("clm-ev2")
		last := c.Events[len(c.Events)-1]
		if last.Action != "status_change" {
			t.Errorf("Action = %q, want 'status_change'", last.Action)
		}
		if last.Payload["from_status"] != "filed" {
			t.Errorf("from_status = %v, want 'filed'", last.Payload["from_status"])
		}
		if last.Payload["to_status"] != "investigating" {
			t.Errorf("to_status = %v, want 'investigating'", last.Payload["to_status"])
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		_, err := p.AdvanceStatus("no-such", "investigating", T1, "adj-1")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestUpdateReserve(t *testing.T) {
	t.Run("updates_reserve_amount", func(t *testing.T) {
		p := seededPipeline(t)
		p.UpdateReserve("clm-002", 35_000, T2, "adj-2")
		c, _ := p.GetClaim("clm-002")
		if c.ReserveAmount != 35_000 {
			t.Errorf("ReserveAmount = %d, want 35000", c.ReserveAmount)
		}
	})

	t.Run("appends_reserve_update_event", func(t *testing.T) {
		p := seededPipeline(t)
		p.UpdateReserve("clm-002", 35_000, T2, "adj-2")
		c, _ := p.GetClaim("clm-002")
		last := c.Events[len(c.Events)-1]
		if last.Action != "reserve_update" {
			t.Errorf("Action = %q, want 'reserve_update'", last.Action)
		}
		if last.Payload["old_reserve"] != 30_000 && last.Payload["old_reserve"] != float64(30_000) {
			t.Errorf("old_reserve = %v, want 30000", last.Payload["old_reserve"])
		}
		if last.Payload["new_reserve"] != 35_000 && last.Payload["new_reserve"] != float64(35_000) {
			t.Errorf("new_reserve = %v, want 35000", last.Payload["new_reserve"])
		}
	})

	t.Run("zero_reserve_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		_, err := p.UpdateReserve("clm-002", 0, T2, "adj-2")
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("closed_claim_returns_error", func(t *testing.T) {
		p := newPipeline(t)
		mustFile(t, p, "clm-cls", "pol-200", "epl", D0, T0, 20_000, 15_000, "adj-1")
		mustAdvance(t, p, "clm-cls", "investigating", T1, "adj-1")
		mustAdvance(t, p, "clm-cls", "evaluation", T2, "adj-1")
		mustSettle(t, p, "clm-cls", 10_000, T3, "adj-1")
		mustAdvance(t, p, "clm-cls", "closed", T4, "adj-1")
		_, err := p.UpdateReserve("clm-cls", 5_000, T4, "adj-1")
		if !errors.Is(err, ErrTerminalState) {
			t.Errorf("expected ErrTerminalState, got %v", err)
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		_, err := p.UpdateReserve("no-such", 10_000, T2, "adj-1")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Settlement, denial, and query methods
// ---------------------------------------------------------------------------

func TestSettleClaim(t *testing.T) {
	t.Run("settles_claim_and_sets_approved_amount", func(t *testing.T) {
		p := seededPipeline(t)
		c, err := p.SettleClaim("clm-003", 150_000, T3, "adj-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Status != "settled" {
			t.Errorf("Status = %q, want 'settled'", c.Status)
		}
		if c.ApprovedAmount == nil || *c.ApprovedAmount != 150_000 {
			t.Errorf("ApprovedAmount = %v, want 150000", c.ApprovedAmount)
		}
	})

	t.Run("appends_settled_event", func(t *testing.T) {
		p := seededPipeline(t)
		p.SettleClaim("clm-003", 150_000, T3, "adj-1")
		c, _ := p.GetClaim("clm-003")
		last := c.Events[len(c.Events)-1]
		if last.Action != "settled" {
			t.Errorf("Action = %q, want 'settled'", last.Action)
		}
		if last.Payload["approved_amount"] != 150_000 && last.Payload["approved_amount"] != float64(150_000) {
			t.Errorf("approved_amount = %v, want 150000", last.Payload["approved_amount"])
		}
	})

	t.Run("wrong_state_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		// clm-002 is in investigating, not evaluation
		_, err := p.SettleClaim("clm-002", 20_000, T3, "adj-2")
		if !errors.Is(err, ErrWrongStatus) {
			t.Errorf("expected ErrWrongStatus, got %v", err)
		}
	})

	t.Run("approved_exceeds_claimed_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		_, err := p.SettleClaim("clm-003", 300_000, T3, "adj-1")
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("zero_approved_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		_, err := p.SettleClaim("clm-003", 0, T3, "adj-1")
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		_, err := p.SettleClaim("no-such", 10_000, T3, "adj-1")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestDenyClaim(t *testing.T) {
	t.Run("denies_claim", func(t *testing.T) {
		p := seededPipeline(t)
		c, err := p.DenyClaim("clm-003", "Outside coverage period.", T3, "adj-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Status != "denied" {
			t.Errorf("Status = %q, want 'denied'", c.Status)
		}
	})

	t.Run("appends_denied_event", func(t *testing.T) {
		p := seededPipeline(t)
		p.DenyClaim("clm-003", "Exclusion.", T3, "adj-1")
		c, _ := p.GetClaim("clm-003")
		last := c.Events[len(c.Events)-1]
		if last.Action != "denied" {
			t.Errorf("Action = %q, want 'denied'", last.Action)
		}
		if last.Payload["reason"] != "Exclusion." {
			t.Errorf("reason = %v, want 'Exclusion.'", last.Payload["reason"])
		}
	})

	t.Run("wrong_state_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		_, err := p.DenyClaim("clm-002", "Reason.", T3, "adj-2")
		if !errors.Is(err, ErrWrongStatus) {
			t.Errorf("expected ErrWrongStatus, got %v", err)
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		p := seededPipeline(t)
		_, err := p.DenyClaim("no-such", "Reason.", T3, "adj-1")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetClaimsByPolicy(t *testing.T) {
	t.Run("returns_claims_for_policy", func(t *testing.T) {
		p := seededPipeline(t)
		claims, err := p.GetClaimsByPolicy("pol-101")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ids := map[string]bool{}
		for _, c := range claims {
			ids[c.ClaimID] = true
		}
		if !ids["clm-001"] {
			t.Error("expected clm-001 in pol-101 claims")
		}
		if !ids["clm-002"] {
			t.Error("expected clm-002 in pol-101 claims")
		}
		if ids["clm-003"] {
			t.Error("clm-003 (pol-102) should not be in pol-101 claims")
		}
	})

	t.Run("sorted_by_filed_at", func(t *testing.T) {
		p := newPipeline(t)
		// file clm-b first with later timestamp, clm-a second with earlier timestamp
		mustFile(t, p, "clm-b", "pol-sort", "epl", D0, T1, 10_000, 8_000, "adj-1")
		mustFile(t, p, "clm-a", "pol-sort", "epl", D0, T0, 10_000, 8_000, "adj-1")
		claims, _ := p.GetClaimsByPolicy("pol-sort")
		for i := 1; i < len(claims); i++ {
			if claims[i-1].FiledAt > claims[i].FiledAt {
				t.Errorf("not sorted ascending: %q before %q", claims[i-1].FiledAt, claims[i].FiledAt)
			}
		}
	})

	t.Run("unknown_policy_returns_empty", func(t *testing.T) {
		p := seededPipeline(t)
		claims, err := p.GetClaimsByPolicy("pol-999")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(claims) != 0 {
			t.Errorf("expected empty slice, got %d claims", len(claims))
		}
	})
}

func TestGetOpenClaims(t *testing.T) {
	t.Run("excludes_denied_but_includes_settled_and_active", func(t *testing.T) {
		p := seededPipeline(t)
		openClaims, err := p.GetOpenClaims()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ids := map[string]bool{}
		for _, c := range openClaims {
			ids[c.ClaimID] = true
		}
		if !ids["clm-001"] {
			t.Error("clm-001 (settled, not closed) should be open")
		}
		if !ids["clm-002"] {
			t.Error("clm-002 (investigating) should be open")
		}
		if !ids["clm-003"] {
			t.Error("clm-003 (evaluation) should be open")
		}
		if ids["clm-004"] {
			t.Error("clm-004 (denied) should NOT be open")
		}
	})

	t.Run("sorted_by_filed_at", func(t *testing.T) {
		p := seededPipeline(t)
		openClaims, _ := p.GetOpenClaims()
		for i := 1; i < len(openClaims); i++ {
			if openClaims[i-1].FiledAt > openClaims[i].FiledAt {
				t.Errorf("not sorted ascending: %q before %q", openClaims[i-1].FiledAt, openClaims[i].FiledAt)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Reserve adequacy and metrics
// ---------------------------------------------------------------------------

func TestGetReserveAdequacy(t *testing.T) {
	t.Run("total_reserves_all_claims", func(t *testing.T) {
		// clm-001: 50_000  clm-002: 30_000  clm-003: 150_000  clm-004: 10_000
		p := seededPipeline(t)
		adequacy, err := p.GetReserveAdequacy()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if adequacy.TotalReserves != 240_000 {
			t.Errorf("TotalReserves = %d, want 240000", adequacy.TotalReserves)
		}
	})

	t.Run("total_approved_settled_only", func(t *testing.T) {
		// only clm-001 is settled, approved=60_000
		p := seededPipeline(t)
		adequacy, _ := p.GetReserveAdequacy()
		if adequacy.TotalApproved != 60_000 {
			t.Errorf("TotalApproved = %d, want 60000", adequacy.TotalApproved)
		}
	})

	t.Run("under_reserved_count_and_gap", func(t *testing.T) {
		// clm-001: reserve=50_000 vs approved=60_000 → under-reserved by 10_000
		p := seededPipeline(t)
		adequacy, _ := p.GetReserveAdequacy()
		if adequacy.UnderReservedCount != 1 {
			t.Errorf("UnderReservedCount = %d, want 1", adequacy.UnderReservedCount)
		}
		if adequacy.UnderReservedGap != 10_000 {
			t.Errorf("UnderReservedGap = %d, want 10000", adequacy.UnderReservedGap)
		}
	})

	t.Run("empty_pipeline", func(t *testing.T) {
		p := newPipeline(t)
		adequacy, err := p.GetReserveAdequacy()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if adequacy.TotalReserves != 0 {
			t.Errorf("TotalReserves = %d, want 0", adequacy.TotalReserves)
		}
		if adequacy.TotalApproved != 0 {
			t.Errorf("TotalApproved = %d, want 0", adequacy.TotalApproved)
		}
		if adequacy.UnderReservedCount != 0 {
			t.Errorf("UnderReservedCount = %d, want 0", adequacy.UnderReservedCount)
		}
		if adequacy.UnderReservedGap != 0 {
			t.Errorf("UnderReservedGap = %d, want 0", adequacy.UnderReservedGap)
		}
	})
}

func TestGetClaimsMetrics(t *testing.T) {
	t.Run("total_count", func(t *testing.T) {
		p := seededPipeline(t)
		metrics, err := p.GetClaimsMetrics()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if metrics.Total != 4 {
			t.Errorf("Total = %d, want 4", metrics.Total)
		}
	})

	t.Run("by_status", func(t *testing.T) {
		p := seededPipeline(t)
		metrics, _ := p.GetClaimsMetrics()
		if metrics.ByStatus["settled"] != 1 {
			t.Errorf("settled count = %d, want 1", metrics.ByStatus["settled"])
		}
		if metrics.ByStatus["investigating"] != 1 {
			t.Errorf("investigating count = %d, want 1", metrics.ByStatus["investigating"])
		}
		if metrics.ByStatus["evaluation"] != 1 {
			t.Errorf("evaluation count = %d, want 1", metrics.ByStatus["evaluation"])
		}
		if metrics.ByStatus["denied"] != 1 {
			t.Errorf("denied count = %d, want 1", metrics.ByStatus["denied"])
		}
	})

	t.Run("only_nonzero_statuses_in_by_status", func(t *testing.T) {
		p := seededPipeline(t)
		metrics, _ := p.GetClaimsMetrics()
		for status, count := range metrics.ByStatus {
			if count == 0 {
				t.Errorf("ByStatus[%q] = 0, should be omitted", status)
			}
		}
	})

	t.Run("total_claimed", func(t *testing.T) {
		p := seededPipeline(t)
		metrics, _ := p.GetClaimsMetrics()
		want := 75_000 + 30_000 + 200_000 + 10_000
		if metrics.TotalClaimed != want {
			t.Errorf("TotalClaimed = %d, want %d", metrics.TotalClaimed, want)
		}
	})

	t.Run("total_paid", func(t *testing.T) {
		// only clm-001 settled, approved=60_000
		p := seededPipeline(t)
		metrics, _ := p.GetClaimsMetrics()
		if metrics.TotalPaid != 60_000 {
			t.Errorf("TotalPaid = %d, want 60000", metrics.TotalPaid)
		}
	})

	t.Run("avg_settlement_ratio", func(t *testing.T) {
		// 1 settled claim: approved=60_000 / claimed=75_000 = 0.8
		p := seededPipeline(t)
		metrics, _ := p.GetClaimsMetrics()
		ratio := float64(60_000) / float64(75_000)
		want := float64(int(ratio*10000+0.5)) / 10000
		if metrics.AvgSettlementRatio != want {
			t.Errorf("AvgSettlementRatio = %v, want %v", metrics.AvgSettlementRatio, want)
		}
	})

	t.Run("no_settled_claims_avg_ratio_is_zero", func(t *testing.T) {
		p := newPipeline(t)
		mustFile(t, p, "clm-ns", "pol-200", "epl", D0, T0, 20_000, 15_000, "adj-1")
		metrics, _ := p.GetClaimsMetrics()
		if metrics.AvgSettlementRatio != 0.0 {
			t.Errorf("AvgSettlementRatio = %v, want 0.0", metrics.AvgSettlementRatio)
		}
	})
}

func TestGetPolicyLossHistory(t *testing.T) {
	t.Run("returns_correct_stats", func(t *testing.T) {
		p := seededPipeline(t)
		history, err := p.GetPolicyLossHistory("pol-101")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if history.PolicyID != "pol-101" {
			t.Errorf("PolicyID = %q, want 'pol-101'", history.PolicyID)
		}
		if history.ClaimCount != 2 {
			t.Errorf("ClaimCount = %d, want 2", history.ClaimCount)
		}
		// 75_000 + 30_000 = 105_000
		if history.TotalClaimed != 105_000 {
			t.Errorf("TotalClaimed = %d, want 105000", history.TotalClaimed)
		}
		// only clm-001 settled, approved=60_000
		if history.TotalPaid != 60_000 {
			t.Errorf("TotalPaid = %d, want 60000", history.TotalPaid)
		}
		lossRatioNumerator := float64(60_000) / float64(105_000)
		wantRatio := float64(int(lossRatioNumerator*10000+0.5)) / 10000
		if history.LossRatio != wantRatio {
			t.Errorf("LossRatio = %v, want %v", history.LossRatio, wantRatio)
		}
	})

	t.Run("unknown_policy_returns_zero_values", func(t *testing.T) {
		p := seededPipeline(t)
		history, err := p.GetPolicyLossHistory("pol-999")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if history.PolicyID != "pol-999" {
			t.Errorf("PolicyID = %q, want 'pol-999'", history.PolicyID)
		}
		if history.ClaimCount != 0 {
			t.Errorf("ClaimCount = %d, want 0", history.ClaimCount)
		}
		if history.TotalClaimed != 0 {
			t.Errorf("TotalClaimed = %d, want 0", history.TotalClaimed)
		}
		if history.TotalPaid != 0 {
			t.Errorf("TotalPaid = %d, want 0", history.TotalPaid)
		}
		if history.LossRatio != 0.0 {
			t.Errorf("LossRatio = %v, want 0.0", history.LossRatio)
		}
	})
}
