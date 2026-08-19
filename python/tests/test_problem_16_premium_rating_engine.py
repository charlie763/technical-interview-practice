"""Tests for Problem 16: Policy Premium Rating Engine

Run from the python/ directory:
    pytest tests/test_problem_16_premium_rating_engine.py -v
"""

import pytest
from practice_problems.problem_16_premium_rating_engine import PremiumRatingEngine


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def fresh_engine():
    """Empty PremiumRatingEngine."""
    return PremiumRatingEngine()


@pytest.fixture
def engine():
    """
    Pre-seeded engine with three submissions:
      sub-epl   "Acme Corp"      epl,       50 employees,  $5M revenue,  7 yrs, medium risk, $1M limit, $10k deductible
      sub-do    "Bright Ventures" do,        N/A,           $10M revenue, 2 yrs, high risk,   $2M limit, $25k deductible
      sub-fid   "Serene Health"  fiduciary, N/A,           $3M revenue,  15 yrs, low risk,   $500k limit, $5k deductible
    """
    e = PremiumRatingEngine()
    e.add_submission(
        "sub-epl", "epl", "Acme Corp",
        employee_count=50, annual_revenue=5_000_000,
        years_in_business=7, industry_risk="medium",
        requested_limit=1_000_000, deductible=10_000,
    )
    e.add_submission(
        "sub-do", "do", "Bright Ventures",
        employee_count=10, annual_revenue=10_000_000,
        years_in_business=2, industry_risk="high",
        requested_limit=2_000_000, deductible=25_000,
    )
    e.add_submission(
        "sub-fid", "fiduciary", "Serene Health",
        employee_count=200, annual_revenue=3_000_000,
        years_in_business=15, industry_risk="low",
        requested_limit=500_000, deductible=5_000,
    )
    return e


# ---------------------------------------------------------------------------
# PART 1 — Submission management and base premium calculation
# ---------------------------------------------------------------------------

class TestAddSubmission:
    def test_returns_submission_dict(self, fresh_engine):
        s = fresh_engine.add_submission(
            "sub-cr-1", "epl", "TestCo",
            employee_count=10, annual_revenue=1_000_000,
            years_in_business=5, industry_risk="low",
            requested_limit=500_000, deductible=5_000,
        )
        assert s["submission_id"] == "sub-cr-1"
        assert s["coverage_type"] == "epl"
        assert s["company_name"] == "TestCo"
        assert s["prior_claims"] == []

    def test_duplicate_raises(self, fresh_engine):
        fresh_engine.add_submission(
            "sub-dup", "do", "A",
            employee_count=5, annual_revenue=500_000,
            years_in_business=3, industry_risk="medium",
            requested_limit=250_000, deductible=2_500,
        )
        with pytest.raises(ValueError):
            fresh_engine.add_submission(
                "sub-dup", "do", "B",
                employee_count=5, annual_revenue=500_000,
                years_in_business=3, industry_risk="medium",
                requested_limit=250_000, deductible=2_500,
            )

    def test_invalid_coverage_type_raises(self, fresh_engine):
        with pytest.raises(ValueError):
            fresh_engine.add_submission(
                "sub-bad", "cyber", "BadCo",
                employee_count=10, annual_revenue=1_000_000,
                years_in_business=5, industry_risk="low",
                requested_limit=500_000, deductible=5_000,
            )

    def test_invalid_industry_risk_raises(self, fresh_engine):
        with pytest.raises(ValueError):
            fresh_engine.add_submission(
                "sub-bad2", "do", "BadCo2",
                employee_count=10, annual_revenue=1_000_000,
                years_in_business=5, industry_risk="extreme",
                requested_limit=500_000, deductible=5_000,
            )


class TestGetSubmission:
    def test_returns_correct_submission(self, engine):
        s = engine.get_submission("sub-epl")
        assert s["submission_id"] == "sub-epl"
        assert s["company_name"] == "Acme Corp"

    def test_unknown_raises(self, engine):
        with pytest.raises(KeyError):
            engine.get_submission("no-such")


class TestCalculateBasePremium:
    def test_epl_formula(self, fresh_engine):
        # EPL: 1200 + (15 * 50) + (5_000_000 * 0.0008) = 1200 + 750 + 4000 = 5950
        fresh_engine.add_submission(
            "sub-epl-calc", "epl", "Co",
            employee_count=50, annual_revenue=5_000_000,
            years_in_business=5, industry_risk="medium",
            requested_limit=1_000_000, deductible=10_000,
        )
        assert fresh_engine.calculate_base_premium("sub-epl-calc") == 5950

    def test_do_formula(self, fresh_engine):
        # D&O: 2500 + (10_000_000 * 0.0010) = 2500 + 10000 = 12500
        fresh_engine.add_submission(
            "sub-do-calc", "do", "Co",
            employee_count=10, annual_revenue=10_000_000,
            years_in_business=5, industry_risk="medium",
            requested_limit=5_000_000, deductible=10_000,
        )
        assert fresh_engine.calculate_base_premium("sub-do-calc") == 12500

    def test_fiduciary_formula(self, fresh_engine):
        # Fiduciary: 800 + (3_000_000 * 0.0004) = 800 + 1200 = 2000
        fresh_engine.add_submission(
            "sub-fid-calc", "fiduciary", "Co",
            employee_count=100, annual_revenue=3_000_000,
            years_in_business=5, industry_risk="medium",
            requested_limit=1_000_000, deductible=5_000,
        )
        assert fresh_engine.calculate_base_premium("sub-fid-calc") == 2000

    def test_cap_at_3pct_of_limit(self, fresh_engine):
        # EPL: 1200 + (10 * 1000) + (100_000_000 * 0.0008) = 1200 + 10000 + 80000 = 91200
        # Cap: 3% of 100_000 = 3000
        fresh_engine.add_submission(
            "sub-cap", "epl", "BigCo",
            employee_count=1000, annual_revenue=100_000_000,
            years_in_business=5, industry_risk="medium",
            requested_limit=100_000, deductible=1_000,
        )
        assert fresh_engine.calculate_base_premium("sub-cap") == 3000

    def test_floor_at_500(self, fresh_engine):
        # D&O: 2500 + (100 * 0.0010) = 2500 + 0 = 2500... but limit is tiny
        # Use very small revenue so computed < 500
        # EPL: 1200 + (15 * 1) + (100 * 0.0008) = ~1215 > 500, not a good floor test
        # Make limit very small so cap kicks in below floor:
        # EPL: 1200 + (15 * 1) + (1000 * 0.0008) = 1215 + 0 = ~1216
        # cap = 3% of 10_000 = 300 < 500 → floor applies → 500
        fresh_engine.add_submission(
            "sub-floor", "epl", "MicroCo",
            employee_count=1, annual_revenue=1_000,
            years_in_business=5, industry_risk="medium",
            requested_limit=10_000, deductible=500,
        )
        assert fresh_engine.calculate_base_premium("sub-floor") == 500

    def test_unknown_submission_raises(self, engine):
        with pytest.raises(KeyError):
            engine.calculate_base_premium("no-such")


# ---------------------------------------------------------------------------
# PART 2 — Prior claims and final premium with modifiers
# ---------------------------------------------------------------------------

class TestRecordPriorClaim:
    def test_appends_claim(self, engine):
        engine.record_prior_claim("sub-epl", year=2023, amount=50_000, claim_type="epl")
        s = engine.get_submission("sub-epl")
        assert len(s["prior_claims"]) == 1
        assert s["prior_claims"][0]["year"] == 2023
        assert s["prior_claims"][0]["amount"] == 50_000
        assert s["prior_claims"][0]["claim_type"] == "epl"

    def test_multiple_claims_accumulate(self, engine):
        engine.record_prior_claim("sub-epl", year=2023, amount=10_000, claim_type="epl")
        engine.record_prior_claim("sub-epl", year=2022, amount=20_000, claim_type="epl")
        assert len(engine.get_submission("sub-epl")["prior_claims"]) == 2

    def test_unknown_raises(self, engine):
        with pytest.raises(KeyError):
            engine.record_prior_claim("no-such", year=2023, amount=10_000, claim_type="epl")


class TestCalculateFinalPremium:
    def test_medium_risk_no_claims_tenure_3_to_10(self, engine):
        # sub-epl: medium × 1.0, 7 yrs × 1.0, no claims × 1.0
        # base = 5950, modifiers all 1.0 → final = 5950
        result = engine.calculate_final_premium("sub-epl", current_year=2025)
        assert result["base_premium"] == 5950
        assert result["industry_modifier"] == 1.0
        assert result["tenure_modifier"] == 1.0
        assert result["claims_modifier"] == 1.0
        assert result["final_premium"] == 5950

    def test_high_risk_young_company(self, engine):
        # sub-do: high × 1.35, 2 yrs × 1.25, no claims × 1.0
        # base = 2500 + (10_000_000 * 0.001) = 12500
        # final = 12500 * 1.35 * 1.25 = 21093.75 → 21094
        result = engine.calculate_final_premium("sub-do", current_year=2025)
        assert result["base_premium"] == 12500
        assert result["industry_modifier"] == 1.35
        assert result["tenure_modifier"] == 1.25
        assert result["final_premium"] == 21094

    def test_low_risk_veteran_company(self, engine):
        # sub-fid: low × 0.85, 15 yrs × 0.90, no claims × 1.0
        # base = 800 + (3_000_000 * 0.0004) = 2000
        # final = 2000 * 0.85 * 0.90 = 1530
        result = engine.calculate_final_premium("sub-fid", current_year=2025)
        assert result["base_premium"] == 2000
        assert result["industry_modifier"] == 0.85
        assert result["tenure_modifier"] == 0.90
        assert result["final_premium"] == 1530

    def test_one_recent_qualifying_claim(self, engine):
        # sub-epl with 1 EPL claim in current_year - 1 → claims_modifier = 1.15
        engine.record_prior_claim("sub-epl", year=2024, amount=30_000, claim_type="epl")
        result = engine.calculate_final_premium("sub-epl", current_year=2025)
        assert result["claims_modifier"] == 1.15
        assert result["final_premium"] == round(5950 * 1.0 * 1.0 * 1.15)

    def test_non_matching_claim_type_ignored(self, engine):
        # sub-epl + a "do" claim → not qualifying → modifier stays 1.0
        engine.record_prior_claim("sub-epl", year=2024, amount=30_000, claim_type="do")
        result = engine.calculate_final_premium("sub-epl", current_year=2025)
        assert result["claims_modifier"] == 1.0

    def test_claim_type_any_qualifies(self, engine):
        # "any" claim type should qualify regardless of coverage_type
        engine.record_prior_claim("sub-epl", year=2024, amount=10_000, claim_type="any")
        result = engine.calculate_final_premium("sub-epl", current_year=2025)
        assert result["claims_modifier"] == 1.15

    def test_old_claim_outside_3_years_ignored(self, engine):
        # claim from 2020, current_year=2025 → year < 2023 → not qualifying
        engine.record_prior_claim("sub-epl", year=2020, amount=30_000, claim_type="epl")
        result = engine.calculate_final_premium("sub-epl", current_year=2025)
        assert result["claims_modifier"] == 1.0

    def test_claims_modifier_capped_at_1_60(self, engine):
        # 5 EPL claims in last 3 years → cap at ×1.60
        for yr in [2023, 2023, 2024, 2024, 2025]:
            engine.record_prior_claim("sub-epl", year=yr, amount=10_000, claim_type="epl")
        result = engine.calculate_final_premium("sub-epl", current_year=2025)
        assert result["claims_modifier"] == 1.60

    def test_final_premium_floor_from_deductible(self, fresh_engine):
        # tiny revenue → base near floor; deductible//10 > computed final
        fresh_engine.add_submission(
            "sub-flr2", "fiduciary", "MicroCo",
            employee_count=1, annual_revenue=1_000,
            years_in_business=5, industry_risk="medium",
            requested_limit=10_000, deductible=8_000,
        )
        result = fresh_engine.calculate_final_premium("sub-flr2", current_year=2025)
        # deductible // 10 = 800; computed = 500 (floor from base); floor = max(500, 800, 500) = 800
        assert result["final_premium"] == 800

    def test_unknown_raises(self, engine):
        with pytest.raises(KeyError):
            engine.calculate_final_premium("no-such", current_year=2025)


# ---------------------------------------------------------------------------
# PART 3 — Portfolio analytics
# ---------------------------------------------------------------------------

class TestGetSubmissionsByCoverageType:
    def test_groups_correctly(self, engine):
        grouped = engine.get_submissions_by_coverage_type()
        assert "epl" in grouped
        assert "do" in grouped
        assert "fiduciary" in grouped
        epl_ids = [s["submission_id"] for s in grouped["epl"]]
        assert "sub-epl" in epl_ids

    def test_sorted_by_submission_id(self, fresh_engine):
        for sid, ct in [("sub-z", "epl"), ("sub-a", "epl"), ("sub-m", "epl")]:
            fresh_engine.add_submission(
                sid, ct, "Co",
                employee_count=10, annual_revenue=500_000,
                years_in_business=5, industry_risk="medium",
                requested_limit=250_000, deductible=5_000,
            )
        grouped = fresh_engine.get_submissions_by_coverage_type()
        ids = [s["submission_id"] for s in grouped["epl"]]
        assert ids == sorted(ids)

    def test_empty_engine_returns_empty(self, fresh_engine):
        assert fresh_engine.get_submissions_by_coverage_type() == {}


class TestGetPortfolioMetrics:
    def test_total_submissions(self, engine):
        metrics = engine.get_portfolio_metrics(current_year=2025)
        assert metrics["total_submissions"] == 3

    def test_total_and_average_premium(self, engine):
        metrics = engine.get_portfolio_metrics(current_year=2025)
        # sub-epl final=5950, sub-do final=21094, sub-fid final=1530
        expected_total = 5950 + 21094 + 1530
        assert metrics["total_premium"] == expected_total
        assert metrics["average_premium"] == round(expected_total / 3)

    def test_by_coverage_type_counts(self, engine):
        metrics = engine.get_portfolio_metrics(current_year=2025)
        assert metrics["by_coverage_type"]["epl"]["count"] == 1
        assert metrics["by_coverage_type"]["do"]["count"] == 1
        assert metrics["by_coverage_type"]["fiduciary"]["count"] == 1

    def test_empty_engine(self, fresh_engine):
        metrics = fresh_engine.get_portfolio_metrics(current_year=2025)
        assert metrics["total_submissions"] == 0
        assert metrics["total_premium"] == 0
        assert metrics["by_coverage_type"] == {}


class TestGetHighRiskSubmissions:
    def test_returns_above_threshold(self, engine):
        # sub-do has modifier 1.35 × 1.25 = 1.6875 > 1.3 threshold
        results = engine.get_high_risk_submissions(current_year=2025, modifier_threshold=1.30)
        ids = [r["submission_id"] for r in results]
        assert "sub-do" in ids

    def test_excludes_below_threshold(self, engine):
        # sub-epl and sub-fid have effective modifiers ≤ 1.0
        results = engine.get_high_risk_submissions(current_year=2025, modifier_threshold=1.30)
        ids = [r["submission_id"] for r in results]
        assert "sub-epl" not in ids
        assert "sub-fid" not in ids

    def test_result_includes_required_fields(self, engine):
        results = engine.get_high_risk_submissions(current_year=2025, modifier_threshold=1.0)
        for r in results:
            assert "submission_id" in r
            assert "company_name" in r
            assert "coverage_type" in r
            assert "effective_modifier" in r
            assert "final_premium" in r

    def test_sorted_by_modifier_descending(self, engine):
        results = engine.get_high_risk_submissions(current_year=2025, modifier_threshold=0.0)
        modifiers = [r["effective_modifier"] for r in results]
        assert modifiers == sorted(modifiers, reverse=True)

    def test_effective_modifier_rounded_to_4_places(self, engine):
        results = engine.get_high_risk_submissions(current_year=2025, modifier_threshold=0.0)
        for r in results:
            assert r["effective_modifier"] == round(r["effective_modifier"], 4)
