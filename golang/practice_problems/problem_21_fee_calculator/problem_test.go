// Tests for Problem 21: Platform Fee Calculator
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_21_fee_calculator/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_21_fee_calculator.go \
//	  -c go test -v .
package fees

import (
	"errors"
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newCalc(t *testing.T) FeeCalculator {
	t.Helper()
	return NewFeeCalculator()
}

func mustCalcFees(t *testing.T, calc FeeCalculator, gross float64, method PaymentMethod) *FeeBreakdown {
	t.Helper()
	f, err := calc.CalculateFees(gross, method)
	if err != nil {
		t.Fatalf("CalculateFees(%v, %q): unexpected error: %v", gross, method, err)
	}
	return f
}

// approxEqual checks that two float64 values are within tolerance.
func approxEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

// seededCalc returns a FeeCalculator with four transactions pre-recorded.
//   - camp-a: card $100 (covered), check $200 (not covered), ach $1000 (covered)
//   - camp-b: card $500 (not covered)
func seededCalc(t *testing.T) FeeCalculator {
	t.Helper()
	calc := NewFeeCalculator()
	if _, err := calc.RecordTransaction("camp-a", 100, PaymentMethodCard, true); err != nil {
		t.Fatalf("seed camp-a card: %v", err)
	}
	if _, err := calc.RecordTransaction("camp-a", 200, PaymentMethodCheck, false); err != nil {
		t.Fatalf("seed camp-a check: %v", err)
	}
	if _, err := calc.RecordTransaction("camp-a", 1000, PaymentMethodACH, true); err != nil {
		t.Fatalf("seed camp-a ach: %v", err)
	}
	if _, err := calc.RecordTransaction("camp-b", 500, PaymentMethodCard, false); err != nil {
		t.Fatalf("seed camp-b card: %v", err)
	}
	return calc
}

// ---------------------------------------------------------------------------
// PART 1 — CalculateFees
// ---------------------------------------------------------------------------

func TestCalculateFees(t *testing.T) {
	t.Run("card_correct_breakdown", func(t *testing.T) {
		// gross=100, platform=3.50, processing=2.90+0.30=3.20, total=6.70, school=93.30
		calc := newCalc(t)
		f := mustCalcFees(t, calc, 100, PaymentMethodCard)
		if f.Gross != 100 {
			t.Errorf("Gross = %v, want 100", f.Gross)
		}
		if f.PlatformFee != 3.50 {
			t.Errorf("PlatformFee = %v, want 3.50", f.PlatformFee)
		}
		if f.ProcessingFee != 3.20 {
			t.Errorf("ProcessingFee = %v, want 3.20", f.ProcessingFee)
		}
		if f.TotalFees != 6.70 {
			t.Errorf("TotalFees = %v, want 6.70", f.TotalFees)
		}
		if f.SchoolReceives != 93.30 {
			t.Errorf("SchoolReceives = %v, want 93.30", f.SchoolReceives)
		}
	})

	t.Run("ach_below_cap", func(t *testing.T) {
		// gross=100, platform=3.50, processing=0.80 (100×0.008=0.80 < 5), total=4.30, school=95.70
		calc := newCalc(t)
		f := mustCalcFees(t, calc, 100, PaymentMethodACH)
		if f.PlatformFee != 3.50 {
			t.Errorf("PlatformFee = %v, want 3.50", f.PlatformFee)
		}
		if f.ProcessingFee != 0.80 {
			t.Errorf("ProcessingFee = %v, want 0.80", f.ProcessingFee)
		}
		if f.TotalFees != 4.30 {
			t.Errorf("TotalFees = %v, want 4.30", f.TotalFees)
		}
		if f.SchoolReceives != 95.70 {
			t.Errorf("SchoolReceives = %v, want 95.70", f.SchoolReceives)
		}
	})

	t.Run("ach_capped_at_five_dollars", func(t *testing.T) {
		// gross=1000, platform=35.00, processing=min(8.00,5.00)=5.00, total=40.00, school=960.00
		calc := newCalc(t)
		f := mustCalcFees(t, calc, 1000, PaymentMethodACH)
		if f.PlatformFee != 35.00 {
			t.Errorf("PlatformFee = %v, want 35.00", f.PlatformFee)
		}
		if f.ProcessingFee != 5.00 {
			t.Errorf("ProcessingFee = %v, want 5.00", f.ProcessingFee)
		}
		if f.TotalFees != 40.00 {
			t.Errorf("TotalFees = %v, want 40.00", f.TotalFees)
		}
		if f.SchoolReceives != 960.00 {
			t.Errorf("SchoolReceives = %v, want 960.00", f.SchoolReceives)
		}
	})

	t.Run("check_no_processing_fee", func(t *testing.T) {
		// gross=200, platform=7.00, processing=0, total=7.00, school=193.00
		calc := newCalc(t)
		f := mustCalcFees(t, calc, 200, PaymentMethodCheck)
		if f.Gross != 200 {
			t.Errorf("Gross = %v, want 200", f.Gross)
		}
		if f.PlatformFee != 7.00 {
			t.Errorf("PlatformFee = %v, want 7.00", f.PlatformFee)
		}
		if f.ProcessingFee != 0 {
			t.Errorf("ProcessingFee = %v, want 0", f.ProcessingFee)
		}
		if f.TotalFees != 7.00 {
			t.Errorf("TotalFees = %v, want 7.00", f.TotalFees)
		}
		if f.SchoolReceives != 193.00 {
			t.Errorf("SchoolReceives = %v, want 193.00", f.SchoolReceives)
		}
	})

	t.Run("rounds_to_nearest_cent", func(t *testing.T) {
		// gross=333, card: platform=11.655→11.66, processing=333×0.029+0.30=9.957→9.96
		calc := newCalc(t)
		f := mustCalcFees(t, calc, 333, PaymentMethodCard)
		if f.PlatformFee != 11.66 {
			t.Errorf("PlatformFee = %v, want 11.66", f.PlatformFee)
		}
		if f.ProcessingFee != 9.96 {
			t.Errorf("ProcessingFee = %v, want 9.96", f.ProcessingFee)
		}
		if f.TotalFees != 21.62 {
			t.Errorf("TotalFees = %v, want 21.62", f.TotalFees)
		}
		if f.SchoolReceives != 311.38 {
			t.Errorf("SchoolReceives = %v, want 311.38", f.SchoolReceives)
		}
	})

	t.Run("zero_gross_returns_error", func(t *testing.T) {
		calc := newCalc(t)
		_, err := calc.CalculateFees(0, PaymentMethodCard)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("negative_gross_returns_error", func(t *testing.T) {
		calc := newCalc(t)
		_, err := calc.CalculateFees(-10, PaymentMethodACH)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — GrossForNet
// ---------------------------------------------------------------------------

func TestGrossForNet(t *testing.T) {
	t.Run("card_school_receives_net_target", func(t *testing.T) {
		calc := newCalc(t)
		f, err := calc.GrossForNet(100, PaymentMethodCard)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !approxEqual(f.SchoolReceives, 100, 0.02) {
			t.Errorf("SchoolReceives = %v, want ~100 (±0.02)", f.SchoolReceives)
		}
	})

	t.Run("card_gross_is_reasonable", func(t *testing.T) {
		// gross = (100 + 0.30) / (1 - 0.035 - 0.029) = 100.30 / 0.936 ≈ 107.16
		calc := newCalc(t)
		f, err := calc.GrossForNet(100, PaymentMethodCard)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.Gross <= 100 {
			t.Errorf("Gross = %v, want > 100", f.Gross)
		}
		if f.Gross >= 115 {
			t.Errorf("Gross = %v, want < 115", f.Gross)
		}
	})

	t.Run("card_breakdown_derived_from_gross", func(t *testing.T) {
		// GrossForNet must call CalculateFees internally, not duplicate logic
		calc := newCalc(t)
		f, err := calc.GrossForNet(100, PaymentMethodCard)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		direct := mustCalcFees(t, calc, f.Gross, PaymentMethodCard)
		if f.PlatformFee != direct.PlatformFee {
			t.Errorf("PlatformFee = %v, want %v (from CalculateFees)", f.PlatformFee, direct.PlatformFee)
		}
		if f.ProcessingFee != direct.ProcessingFee {
			t.Errorf("ProcessingFee = %v, want %v (from CalculateFees)", f.ProcessingFee, direct.ProcessingFee)
		}
	})

	t.Run("ach_below_cap_school_receives_net_target", func(t *testing.T) {
		// net=100; gross=100/(1-0.035-0.008)=100/0.957≈104.49; processing=104.49×0.008≈0.84 < 5 → uncapped
		calc := newCalc(t)
		f, err := calc.GrossForNet(100, PaymentMethodACH)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !approxEqual(f.SchoolReceives, 100, 0.02) {
			t.Errorf("SchoolReceives = %v, want ~100 (±0.02)", f.SchoolReceives)
		}
	})

	t.Run("ach_capped_path_school_receives_net_target", func(t *testing.T) {
		// large net where 0.8% would exceed $5 cap
		// net=700; uncapped gross=700/0.957≈731.45; processing=731.45×0.008≈5.85 > 5 → use capped formula
		calc := newCalc(t)
		f, err := calc.GrossForNet(700, PaymentMethodACH)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.ProcessingFee != 5.00 {
			t.Errorf("ProcessingFee = %v, want 5.00 (capped)", f.ProcessingFee)
		}
		if !approxEqual(f.SchoolReceives, 700, 0.02) {
			t.Errorf("SchoolReceives = %v, want ~700 (±0.02)", f.SchoolReceives)
		}
	})

	t.Run("check_school_receives_net_target", func(t *testing.T) {
		calc := newCalc(t)
		f, err := calc.GrossForNet(500, PaymentMethodCheck)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.ProcessingFee != 0 {
			t.Errorf("ProcessingFee = %v, want 0 for check", f.ProcessingFee)
		}
		if !approxEqual(f.SchoolReceives, 500, 0.02) {
			t.Errorf("SchoolReceives = %v, want ~500 (±0.02)", f.SchoolReceives)
		}
	})

	t.Run("zero_net_target_returns_error", func(t *testing.T) {
		calc := newCalc(t)
		_, err := calc.GrossForNet(0, PaymentMethodCard)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("negative_net_target_returns_error", func(t *testing.T) {
		calc := newCalc(t)
		_, err := calc.GrossForNet(-50, PaymentMethodACH)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — RecordTransaction and GetCampaignFeeSummary
// ---------------------------------------------------------------------------

func TestRecordTransaction(t *testing.T) {
	t.Run("returns_correct_fields", func(t *testing.T) {
		calc := newCalc(t)
		tx, err := calc.RecordTransaction("camp-z", 100, PaymentMethodCard, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tx.CampaignID != "camp-z" {
			t.Errorf("CampaignID = %q, want 'camp-z'", tx.CampaignID)
		}
		if tx.Gross != 100 {
			t.Errorf("Gross = %v, want 100", tx.Gross)
		}
		if tx.Method != PaymentMethodCard {
			t.Errorf("Method = %q, want 'card'", tx.Method)
		}
		if !tx.DonorCoversFee {
			t.Error("DonorCoversFee = false, want true")
		}
		if tx.ID == "" {
			t.Error("ID is empty, want non-empty auto-generated ID")
		}
		if tx.SchoolReceives != 93.30 {
			t.Errorf("SchoolReceives = %v, want 93.30", tx.SchoolReceives)
		}
		if tx.TotalFees != 6.70 {
			t.Errorf("TotalFees = %v, want 6.70", tx.TotalFees)
		}
	})

	t.Run("auto_generates_unique_ids", func(t *testing.T) {
		calc := newCalc(t)
		a, err := calc.RecordTransaction("x", 100, PaymentMethodCard, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, err := calc.RecordTransaction("x", 200, PaymentMethodCard, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.ID == b.ID {
			t.Errorf("expected unique IDs, both got %q", a.ID)
		}
	})

	t.Run("zero_gross_returns_error", func(t *testing.T) {
		calc := newCalc(t)
		_, err := calc.RecordTransaction("camp-a", 0, PaymentMethodCard, false)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})
}

func TestGetCampaignFeeSummary(t *testing.T) {
	t.Run("correct_totals", func(t *testing.T) {
		// camp-a: card $100, check $200, ach $1000
		// school: 93.30 + 193.00 + 960.00 = 1246.30
		// fees:    6.70 +   7.00 +  40.00 =   53.70
		calc := seededCalc(t)
		s, err := calc.GetCampaignFeeSummary("camp-a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.CampaignID != "camp-a" {
			t.Errorf("CampaignID = %q, want 'camp-a'", s.CampaignID)
		}
		if s.TransactionCount != 3 {
			t.Errorf("TransactionCount = %d, want 3", s.TransactionCount)
		}
		if s.TotalGross != 1300 {
			t.Errorf("TotalGross = %v, want 1300", s.TotalGross)
		}
		if s.TotalSchoolReceives != 1246.30 {
			t.Errorf("TotalSchoolReceives = %v, want 1246.30", s.TotalSchoolReceives)
		}
		if s.TotalFees != 53.70 {
			t.Errorf("TotalFees = %v, want 53.70", s.TotalFees)
		}
	})

	t.Run("fee_coverage_rate_is_fraction_of_donor_covered", func(t *testing.T) {
		// camp-a: 2 covered out of 3 → 0.67
		calc := seededCalc(t)
		s, err := calc.GetCampaignFeeSummary("camp-a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.FeeCoverageRate != 0.67 {
			t.Errorf("FeeCoverageRate = %v, want 0.67", s.FeeCoverageRate)
		}
	})

	t.Run("by_method_only_includes_used_methods", func(t *testing.T) {
		calc := seededCalc(t)
		s, err := calc.GetCampaignFeeSummary("camp-a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.ByMethod[PaymentMethodCard] == nil {
			t.Error("expected card entry in ByMethod")
		}
		if s.ByMethod[PaymentMethodACH] == nil {
			t.Error("expected ach entry in ByMethod")
		}
		if s.ByMethod[PaymentMethodCheck] == nil {
			t.Error("expected check entry in ByMethod")
		}
	})

	t.Run("by_method_has_correct_aggregates", func(t *testing.T) {
		calc := seededCalc(t)
		s, err := calc.GetCampaignFeeSummary("camp-a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		cardSummary := s.ByMethod[PaymentMethodCard]
		if cardSummary == nil {
			t.Fatal("card summary is nil")
		}
		if cardSummary.Count != 1 {
			t.Errorf("card Count = %d, want 1", cardSummary.Count)
		}
		if cardSummary.TotalGross != 100 {
			t.Errorf("card TotalGross = %v, want 100", cardSummary.TotalGross)
		}
		achSummary := s.ByMethod[PaymentMethodACH]
		if achSummary == nil {
			t.Fatal("ach summary is nil")
		}
		if achSummary.Count != 1 {
			t.Errorf("ach Count = %d, want 1", achSummary.Count)
		}
		if achSummary.TotalGross != 1000 {
			t.Errorf("ach TotalGross = %v, want 1000", achSummary.TotalGross)
		}
	})

	t.Run("unknown_campaign_returns_error", func(t *testing.T) {
		calc := seededCalc(t)
		_, err := calc.GetCampaignFeeSummary("no-such")
		if !errors.Is(err, ErrUnknownCampaign) {
			t.Errorf("expected ErrUnknownCampaign, got %v", err)
		}
	})

	t.Run("isolated_per_campaign", func(t *testing.T) {
		calc := seededCalc(t)
		s, err := calc.GetCampaignFeeSummary("camp-b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.TransactionCount != 1 {
			t.Errorf("TransactionCount = %d, want 1", s.TransactionCount)
		}
		if s.TotalGross != 500 {
			t.Errorf("TotalGross = %v, want 500", s.TotalGross)
		}
	})
}
