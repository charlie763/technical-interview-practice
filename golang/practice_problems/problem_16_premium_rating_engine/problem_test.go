// Tests for Problem 16: Policy Premium Rating Engine
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_16_premium_rating_engine/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_16_premium_rating_engine.go \
//	  -c go test -v .
package premiumrating

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newEngine(t *testing.T) PremiumRatingEngine {
	t.Helper()
	return NewPremiumRatingEngine()
}

// seededEngine returns a PremiumRatingEngine pre-loaded with three submissions:
//
//	sub-epl   "Acme Corp"      epl,       50 employees, $5M revenue,  7 yrs, medium risk, $1M limit, $10k deductible
//	sub-do    "Bright Ventures" do,        10 employees, $10M revenue, 2 yrs, high risk,   $2M limit, $25k deductible
//	sub-fid   "Serene Health"  fiduciary, 200 employees, $3M revenue, 15 yrs, low risk,   $500k limit, $5k deductible
func seededEngine(t *testing.T) PremiumRatingEngine {
	t.Helper()
	e := NewPremiumRatingEngine()
	mustAddSub(t, e, "sub-epl", "epl", "Acme Corp", 50, 5_000_000, 7, "medium", 1_000_000, 10_000)
	mustAddSub(t, e, "sub-do", "do", "Bright Ventures", 10, 10_000_000, 2, "high", 2_000_000, 25_000)
	mustAddSub(t, e, "sub-fid", "fiduciary", "Serene Health", 200, 3_000_000, 15, "low", 500_000, 5_000)
	return e
}

func mustAddSub(t *testing.T, e PremiumRatingEngine, id, ct, company string, employees, revenue, yrs int, risk string, limit, ded int) {
	t.Helper()
	if _, err := e.AddSubmission(id, ct, company, employees, revenue, yrs, risk, limit, ded); err != nil {
		t.Fatalf("AddSubmission(%q): %v", id, err)
	}
}

func mustFinal(t *testing.T, e PremiumRatingEngine, id string, yr int) *FinalPremiumResult {
	t.Helper()
	r, err := e.CalculateFinalPremium(id, yr)
	if err != nil {
		t.Fatalf("CalculateFinalPremium(%q, %d): %v", id, yr, err)
	}
	return r
}

// ---------------------------------------------------------------------------
// PART 1 — Submission management and base premium calculation
// ---------------------------------------------------------------------------

func TestAddSubmission(t *testing.T) {
	t.Run("returns_submission_with_correct_fields", func(t *testing.T) {
		e := newEngine(t)
		s, err := e.AddSubmission("sub-cr-1", "epl", "TestCo", 10, 1_000_000, 5, "low", 500_000, 5_000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.SubmissionID != "sub-cr-1" {
			t.Errorf("SubmissionID = %q, want 'sub-cr-1'", s.SubmissionID)
		}
		if s.CoverageType != "epl" {
			t.Errorf("CoverageType = %q, want 'epl'", s.CoverageType)
		}
		if s.CompanyName != "TestCo" {
			t.Errorf("CompanyName = %q, want 'TestCo'", s.CompanyName)
		}
		if len(s.PriorClaims) != 0 {
			t.Errorf("PriorClaims = %v, want empty", s.PriorClaims)
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		e := newEngine(t)
		mustAddSub(t, e, "sub-dup", "do", "A", 5, 500_000, 3, "medium", 250_000, 2_500)
		_, err := e.AddSubmission("sub-dup", "do", "B", 5, 500_000, 3, "medium", 250_000, 2_500)
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("invalid_coverage_type_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.AddSubmission("sub-bad", "cyber", "BadCo", 10, 1_000_000, 5, "low", 500_000, 5_000)
		if !errors.Is(err, ErrInvalidCoverageType) {
			t.Errorf("expected ErrInvalidCoverageType, got %v", err)
		}
	})

	t.Run("invalid_industry_risk_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.AddSubmission("sub-bad2", "do", "BadCo2", 10, 1_000_000, 5, "extreme", 500_000, 5_000)
		if !errors.Is(err, ErrInvalidIndustryRisk) {
			t.Errorf("expected ErrInvalidIndustryRisk, got %v", err)
		}
	})
}

func TestGetSubmission(t *testing.T) {
	t.Run("returns_correct_submission", func(t *testing.T) {
		e := seededEngine(t)
		s, err := e.GetSubmission("sub-epl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.SubmissionID != "sub-epl" {
			t.Errorf("SubmissionID = %q, want 'sub-epl'", s.SubmissionID)
		}
		if s.CompanyName != "Acme Corp" {
			t.Errorf("CompanyName = %q, want 'Acme Corp'", s.CompanyName)
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		e := seededEngine(t)
		_, err := e.GetSubmission("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestCalculateBasePremium(t *testing.T) {
	tests := []struct {
		name           string
		ct             string
		employees      int
		revenue        int
		limit          int
		deductible     int
		wantPremium    int
	}{
		// EPL: 1200 + (15*50) + (5_000_000*0.0008) = 1200 + 750 + 4000 = 5950
		{"epl_formula", "epl", 50, 5_000_000, 1_000_000, 10_000, 5950},
		// D&O: 2500 + (10_000_000*0.0010) = 2500 + 10000 = 12500
		{"do_formula", "do", 10, 10_000_000, 5_000_000, 10_000, 12500},
		// Fiduciary: 800 + (3_000_000*0.0004) = 800 + 1200 = 2000
		{"fiduciary_formula", "fiduciary", 100, 3_000_000, 1_000_000, 5_000, 2000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEngine(t)
			mustAddSub(t, e, "sub-calc", tc.ct, "Co", tc.employees, tc.revenue, 5, "medium", tc.limit, tc.deductible)
			got, err := e.CalculateBasePremium("sub-calc")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantPremium {
				t.Errorf("CalculateBasePremium = %d, want %d", got, tc.wantPremium)
			}
		})
	}

	t.Run("cap_at_3pct_of_limit", func(t *testing.T) {
		// EPL: 1200 + (15*1000) + (100_000_000*0.0008) = 91200; cap = 3% of 100_000 = 3000
		e := newEngine(t)
		mustAddSub(t, e, "sub-cap", "epl", "BigCo", 1000, 100_000_000, 5, "medium", 100_000, 1_000)
		got, err := e.CalculateBasePremium("sub-cap")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 3000 {
			t.Errorf("CalculateBasePremium (cap) = %d, want 3000", got)
		}
	})

	t.Run("floor_at_500", func(t *testing.T) {
		// EPL with tiny revenue: cap = 3% of 10_000 = 300 < 500 → floor = 500
		e := newEngine(t)
		mustAddSub(t, e, "sub-floor", "epl", "MicroCo", 1, 1_000, 5, "medium", 10_000, 500)
		got, err := e.CalculateBasePremium("sub-floor")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 500 {
			t.Errorf("CalculateBasePremium (floor) = %d, want 500", got)
		}
	})

	t.Run("unknown_submission_returns_error", func(t *testing.T) {
		e := seededEngine(t)
		_, err := e.CalculateBasePremium("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Prior claims and final premium with modifiers
// ---------------------------------------------------------------------------

func TestRecordPriorClaim(t *testing.T) {
	t.Run("appends_claim", func(t *testing.T) {
		e := seededEngine(t)
		updated, err := e.RecordPriorClaim("sub-epl", 2023, 50_000, "epl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(updated.PriorClaims) != 1 {
			t.Fatalf("len(PriorClaims) = %d, want 1", len(updated.PriorClaims))
		}
		c := updated.PriorClaims[0]
		if c.Year != 2023 || c.Amount != 50_000 || c.ClaimType != "epl" {
			t.Errorf("PriorClaim = %+v, want {2023 50000 epl}", c)
		}
	})

	t.Run("multiple_claims_accumulate", func(t *testing.T) {
		e := seededEngine(t)
		e.RecordPriorClaim("sub-epl", 2023, 10_000, "epl")
		updated, _ := e.RecordPriorClaim("sub-epl", 2022, 20_000, "epl")
		if len(updated.PriorClaims) != 2 {
			t.Errorf("len(PriorClaims) = %d, want 2", len(updated.PriorClaims))
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		e := seededEngine(t)
		_, err := e.RecordPriorClaim("no-such", 2023, 10_000, "epl")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestCalculateFinalPremium(t *testing.T) {
	t.Run("medium_risk_no_claims_tenure_3_to_10", func(t *testing.T) {
		// sub-epl: medium×1.0, 7yrs×1.0, no claims×1.0; base=5950 → final=5950
		e := seededEngine(t)
		r := mustFinal(t, e, "sub-epl", 2025)
		if r.BasePremium != 5950 {
			t.Errorf("BasePremium = %d, want 5950", r.BasePremium)
		}
		if r.IndustryModifier != 1.0 {
			t.Errorf("IndustryModifier = %v, want 1.0", r.IndustryModifier)
		}
		if r.TenureModifier != 1.0 {
			t.Errorf("TenureModifier = %v, want 1.0", r.TenureModifier)
		}
		if r.ClaimsModifier != 1.0 {
			t.Errorf("ClaimsModifier = %v, want 1.0", r.ClaimsModifier)
		}
		if r.FinalPremium != 5950 {
			t.Errorf("FinalPremium = %d, want 5950", r.FinalPremium)
		}
	})

	t.Run("high_risk_young_company", func(t *testing.T) {
		// sub-do: high×1.35, 2yrs×1.25, no claims×1.0; base=12500
		// final = 12500 * 1.35 * 1.25 = 21093.75 → 21094
		e := seededEngine(t)
		r := mustFinal(t, e, "sub-do", 2025)
		if r.BasePremium != 12500 {
			t.Errorf("BasePremium = %d, want 12500", r.BasePremium)
		}
		if r.IndustryModifier != 1.35 {
			t.Errorf("IndustryModifier = %v, want 1.35", r.IndustryModifier)
		}
		if r.TenureModifier != 1.25 {
			t.Errorf("TenureModifier = %v, want 1.25", r.TenureModifier)
		}
		if r.FinalPremium != 21094 {
			t.Errorf("FinalPremium = %d, want 21094", r.FinalPremium)
		}
	})

	t.Run("low_risk_veteran_company", func(t *testing.T) {
		// sub-fid: low×0.85, 15yrs×0.90, no claims×1.0; base=2000
		// final = 2000 * 0.85 * 0.90 = 1530
		e := seededEngine(t)
		r := mustFinal(t, e, "sub-fid", 2025)
		if r.BasePremium != 2000 {
			t.Errorf("BasePremium = %d, want 2000", r.BasePremium)
		}
		if r.IndustryModifier != 0.85 {
			t.Errorf("IndustryModifier = %v, want 0.85", r.IndustryModifier)
		}
		if r.TenureModifier != 0.90 {
			t.Errorf("TenureModifier = %v, want 0.90", r.TenureModifier)
		}
		if r.FinalPremium != 1530 {
			t.Errorf("FinalPremium = %d, want 1530", r.FinalPremium)
		}
	})

	t.Run("one_recent_qualifying_claim", func(t *testing.T) {
		// sub-epl + 1 epl claim in 2024 → claims_modifier = 1.15
		// final = round(5950 * 1.0 * 1.0 * 1.15) = 6843
		e := seededEngine(t)
		e.RecordPriorClaim("sub-epl", 2024, 30_000, "epl")
		r := mustFinal(t, e, "sub-epl", 2025)
		if r.ClaimsModifier != 1.15 {
			t.Errorf("ClaimsModifier = %v, want 1.15", r.ClaimsModifier)
		}
		want := int(5950.0 * 1.0 * 1.0 * 1.15 + 0.5)
		if r.FinalPremium != want {
			t.Errorf("FinalPremium = %d, want %d", r.FinalPremium, want)
		}
	})

	t.Run("non_matching_claim_type_ignored", func(t *testing.T) {
		// sub-epl + "do" claim → not qualifying → modifier stays 1.0
		e := seededEngine(t)
		e.RecordPriorClaim("sub-epl", 2024, 30_000, "do")
		r := mustFinal(t, e, "sub-epl", 2025)
		if r.ClaimsModifier != 1.0 {
			t.Errorf("ClaimsModifier = %v, want 1.0 (non-matching type)", r.ClaimsModifier)
		}
	})

	t.Run("claim_type_any_qualifies", func(t *testing.T) {
		// "any" claim type should qualify for any coverage type
		e := seededEngine(t)
		e.RecordPriorClaim("sub-epl", 2024, 10_000, "any")
		r := mustFinal(t, e, "sub-epl", 2025)
		if r.ClaimsModifier != 1.15 {
			t.Errorf("ClaimsModifier = %v, want 1.15 for 'any' claim type", r.ClaimsModifier)
		}
	})

	t.Run("old_claim_outside_3_years_ignored", func(t *testing.T) {
		// claim from 2020, current_year=2025 → year < 2023 → not qualifying
		e := seededEngine(t)
		e.RecordPriorClaim("sub-epl", 2020, 30_000, "epl")
		r := mustFinal(t, e, "sub-epl", 2025)
		if r.ClaimsModifier != 1.0 {
			t.Errorf("ClaimsModifier = %v, want 1.0 for old claim", r.ClaimsModifier)
		}
	})

	t.Run("claims_modifier_capped_at_1_60", func(t *testing.T) {
		// 5 epl claims in last 3 years → cap at ×1.60
		e := seededEngine(t)
		for _, yr := range []int{2023, 2023, 2024, 2024, 2025} {
			e.RecordPriorClaim("sub-epl", yr, 10_000, "epl")
		}
		r := mustFinal(t, e, "sub-epl", 2025)
		if r.ClaimsModifier != 1.60 {
			t.Errorf("ClaimsModifier = %v, want 1.60 (cap)", r.ClaimsModifier)
		}
	})

	t.Run("final_premium_floor_from_deductible", func(t *testing.T) {
		// tiny revenue → base near floor; deductible/10 = 800 > computed → floor = 800
		e := newEngine(t)
		mustAddSub(t, e, "sub-flr2", "fiduciary", "MicroCo", 1, 1_000, 5, "medium", 10_000, 8_000)
		r := mustFinal(t, e, "sub-flr2", 2025)
		if r.FinalPremium != 800 {
			t.Errorf("FinalPremium = %d, want 800 (deductible floor)", r.FinalPremium)
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		e := seededEngine(t)
		_, err := e.CalculateFinalPremium("no-such", 2025)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Portfolio analytics
// ---------------------------------------------------------------------------

func TestGetSubmissionsByCoverageType(t *testing.T) {
	t.Run("groups_correctly", func(t *testing.T) {
		e := seededEngine(t)
		grouped := e.GetSubmissionsByCoverageType()
		if _, ok := grouped["epl"]; !ok {
			t.Error("expected 'epl' key in grouped result")
		}
		if _, ok := grouped["do"]; !ok {
			t.Error("expected 'do' key in grouped result")
		}
		if _, ok := grouped["fiduciary"]; !ok {
			t.Error("expected 'fiduciary' key in grouped result")
		}
		found := false
		for _, s := range grouped["epl"] {
			if s.SubmissionID == "sub-epl" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'sub-epl' in epl group")
		}
	})

	t.Run("sorted_by_submission_id", func(t *testing.T) {
		e := newEngine(t)
		for _, id := range []string{"sub-z", "sub-a", "sub-m"} {
			mustAddSub(t, e, id, "epl", "Co", 10, 500_000, 5, "medium", 250_000, 5_000)
		}
		grouped := e.GetSubmissionsByCoverageType()
		subs := grouped["epl"]
		for i := 1; i < len(subs); i++ {
			if subs[i-1].SubmissionID > subs[i].SubmissionID {
				t.Errorf("not sorted: %q before %q", subs[i-1].SubmissionID, subs[i].SubmissionID)
			}
		}
	})

	t.Run("empty_engine_returns_empty", func(t *testing.T) {
		e := newEngine(t)
		grouped := e.GetSubmissionsByCoverageType()
		if len(grouped) != 0 {
			t.Errorf("expected empty map, got %v", grouped)
		}
	})
}

func TestGetPortfolioMetrics(t *testing.T) {
	t.Run("total_submissions", func(t *testing.T) {
		e := seededEngine(t)
		metrics, err := e.GetPortfolioMetrics(2025)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if metrics.TotalSubmissions != 3 {
			t.Errorf("TotalSubmissions = %d, want 3", metrics.TotalSubmissions)
		}
	})

	t.Run("total_and_average_premium", func(t *testing.T) {
		// sub-epl final=5950, sub-do final=21094, sub-fid final=1530
		e := seededEngine(t)
		metrics, err := e.GetPortfolioMetrics(2025)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantTotal := 5950 + 21094 + 1530
		if metrics.TotalPremium != wantTotal {
			t.Errorf("TotalPremium = %d, want %d", metrics.TotalPremium, wantTotal)
		}
		wantAvg := int(float64(wantTotal)/3.0 + 0.5)
		if metrics.AveragePremium != wantAvg {
			t.Errorf("AveragePremium = %d, want %d", metrics.AveragePremium, wantAvg)
		}
	})

	t.Run("by_coverage_type_counts", func(t *testing.T) {
		e := seededEngine(t)
		metrics, err := e.GetPortfolioMetrics(2025)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if metrics.ByCoverageType["epl"].Count != 1 {
			t.Errorf("epl count = %d, want 1", metrics.ByCoverageType["epl"].Count)
		}
		if metrics.ByCoverageType["do"].Count != 1 {
			t.Errorf("do count = %d, want 1", metrics.ByCoverageType["do"].Count)
		}
		if metrics.ByCoverageType["fiduciary"].Count != 1 {
			t.Errorf("fiduciary count = %d, want 1", metrics.ByCoverageType["fiduciary"].Count)
		}
	})

	t.Run("empty_engine", func(t *testing.T) {
		e := newEngine(t)
		metrics, err := e.GetPortfolioMetrics(2025)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if metrics.TotalSubmissions != 0 {
			t.Errorf("TotalSubmissions = %d, want 0", metrics.TotalSubmissions)
		}
		if metrics.TotalPremium != 0 {
			t.Errorf("TotalPremium = %d, want 0", metrics.TotalPremium)
		}
		if len(metrics.ByCoverageType) != 0 {
			t.Errorf("ByCoverageType = %v, want empty map", metrics.ByCoverageType)
		}
	})
}

func TestGetHighRiskSubmissions(t *testing.T) {
	t.Run("returns_above_threshold", func(t *testing.T) {
		// sub-do has effective modifier 1.35*1.25 = 1.6875 > 1.30 threshold
		e := seededEngine(t)
		results, err := e.GetHighRiskSubmissions(2025, 1.30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		found := false
		for _, r := range results {
			if r.SubmissionID == "sub-do" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'sub-do' in high risk results above 1.30 threshold")
		}
	})

	t.Run("excludes_below_threshold", func(t *testing.T) {
		// sub-epl and sub-fid have effective modifiers at or below 1.0
		e := seededEngine(t)
		results, err := e.GetHighRiskSubmissions(2025, 1.30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, r := range results {
			if r.SubmissionID == "sub-epl" || r.SubmissionID == "sub-fid" {
				t.Errorf("unexpected submission %q in high risk results", r.SubmissionID)
			}
		}
	})

	t.Run("result_includes_required_fields", func(t *testing.T) {
		e := seededEngine(t)
		results, err := e.GetHighRiskSubmissions(2025, 0.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, r := range results {
			if r.SubmissionID == "" {
				t.Error("SubmissionID is empty")
			}
			if r.CompanyName == "" {
				t.Error("CompanyName is empty")
			}
			if r.CoverageType == "" {
				t.Error("CoverageType is empty")
			}
			if r.FinalPremium == 0 {
				t.Error("FinalPremium is 0")
			}
		}
	})

	t.Run("sorted_by_modifier_descending", func(t *testing.T) {
		e := seededEngine(t)
		results, err := e.GetHighRiskSubmissions(2025, 0.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for i := 1; i < len(results); i++ {
			if results[i-1].EffectiveModifier < results[i].EffectiveModifier {
				t.Errorf("not sorted descending: %v before %v", results[i-1].EffectiveModifier, results[i].EffectiveModifier)
			}
		}
	})

	t.Run("effective_modifier_rounded_to_4_places", func(t *testing.T) {
		e := seededEngine(t)
		results, err := e.GetHighRiskSubmissions(2025, 0.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, r := range results {
			// Check that rounding to 4 places gives the same value
			rounded := float64(int(r.EffectiveModifier*10000+0.5)) / 10000
			if r.EffectiveModifier != rounded {
				t.Errorf("EffectiveModifier %v is not rounded to 4 decimal places", r.EffectiveModifier)
			}
		}
	})
}
