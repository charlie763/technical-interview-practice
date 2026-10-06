load_problem("16_premium_rating_engine")

RSpec.describe PremiumRatingEngine do
  let(:engine) { described_class.new }

  # Pre-seeded engine with three submissions:
  #   sub-epl  "Acme Corp"       epl,       50 employees, $5M revenue,  7 yrs, medium risk, $1M limit,   $10k deductible
  #   sub-do   "Bright Ventures" do,        10 employees, $10M revenue, 2 yrs, high risk,   $2M limit,   $25k deductible
  #   sub-fid  "Serene Health"   fiduciary, 200 employees, $3M revenue, 15 yrs, low risk,   $500k limit, $5k deductible
  let(:seeded_engine) do
    e = described_class.new
    e.add_submission(
      "sub-epl", "epl", "Acme Corp",
      employee_count: 50, annual_revenue: 5_000_000,
      years_in_business: 7, industry_risk: "medium",
      requested_limit: 1_000_000, deductible: 10_000
    )
    e.add_submission(
      "sub-do", "do", "Bright Ventures",
      employee_count: 10, annual_revenue: 10_000_000,
      years_in_business: 2, industry_risk: "high",
      requested_limit: 2_000_000, deductible: 25_000
    )
    e.add_submission(
      "sub-fid", "fiduciary", "Serene Health",
      employee_count: 200, annual_revenue: 3_000_000,
      years_in_business: 15, industry_risk: "low",
      requested_limit: 500_000, deductible: 5_000
    )
    e
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Submission management and base premium calculation
  # ---------------------------------------------------------------------------
  describe "#add_submission" do
    it "returns the submission hash" do
      s = engine.add_submission(
        "sub-cr-1", "epl", "TestCo",
        employee_count: 10, annual_revenue: 1_000_000,
        years_in_business: 5, industry_risk: "low",
        requested_limit: 500_000, deductible: 5_000
      )
      expect(s[:submission_id]).to eq("sub-cr-1")
      expect(s[:coverage_type]).to eq("epl")
      expect(s[:company_name]).to eq("TestCo")
      expect(s[:prior_claims]).to eq([])
    end

    it "raises ArgumentError on a duplicate" do
      engine.add_submission(
        "sub-dup", "do", "A",
        employee_count: 5, annual_revenue: 500_000,
        years_in_business: 3, industry_risk: "medium",
        requested_limit: 250_000, deductible: 2_500
      )
      expect do
        engine.add_submission(
          "sub-dup", "do", "B",
          employee_count: 5, annual_revenue: 500_000,
          years_in_business: 3, industry_risk: "medium",
          requested_limit: 250_000, deductible: 2_500
        )
      end.to raise_error(ArgumentError)
    end

    it "raises ArgumentError for an invalid coverage_type" do
      expect do
        engine.add_submission(
          "sub-bad", "cyber", "BadCo",
          employee_count: 10, annual_revenue: 1_000_000,
          years_in_business: 5, industry_risk: "low",
          requested_limit: 500_000, deductible: 5_000
        )
      end.to raise_error(ArgumentError)
    end

    it "raises ArgumentError for an invalid industry_risk" do
      expect do
        engine.add_submission(
          "sub-bad2", "do", "BadCo2",
          employee_count: 10, annual_revenue: 1_000_000,
          years_in_business: 5, industry_risk: "extreme",
          requested_limit: 500_000, deductible: 5_000
        )
      end.to raise_error(ArgumentError)
    end
  end

  describe "#get_submission" do
    it "returns the correct submission" do
      s = seeded_engine.get_submission("sub-epl")
      expect(s[:submission_id]).to eq("sub-epl")
      expect(s[:company_name]).to eq("Acme Corp")
    end

    it "raises KeyError for an unknown submission" do
      expect { seeded_engine.get_submission("no-such") }.to raise_error(KeyError)
    end
  end

  describe "#calculate_base_premium" do
    it "applies the EPL formula" do
      # EPL: 1200 + (15 * 50) + (5_000_000 * 0.0008) = 1200 + 750 + 4000 = 5950
      engine.add_submission(
        "sub-epl-calc", "epl", "Co",
        employee_count: 50, annual_revenue: 5_000_000,
        years_in_business: 5, industry_risk: "medium",
        requested_limit: 1_000_000, deductible: 10_000
      )
      expect(engine.calculate_base_premium("sub-epl-calc")).to eq(5950)
    end

    it "applies the D&O formula" do
      # D&O: 2500 + (10_000_000 * 0.0010) = 2500 + 10000 = 12500
      engine.add_submission(
        "sub-do-calc", "do", "Co",
        employee_count: 10, annual_revenue: 10_000_000,
        years_in_business: 5, industry_risk: "medium",
        requested_limit: 5_000_000, deductible: 10_000
      )
      expect(engine.calculate_base_premium("sub-do-calc")).to eq(12500)
    end

    it "applies the fiduciary formula" do
      # Fiduciary: 800 + (3_000_000 * 0.0004) = 800 + 1200 = 2000
      engine.add_submission(
        "sub-fid-calc", "fiduciary", "Co",
        employee_count: 100, annual_revenue: 3_000_000,
        years_in_business: 5, industry_risk: "medium",
        requested_limit: 1_000_000, deductible: 5_000
      )
      expect(engine.calculate_base_premium("sub-fid-calc")).to eq(2000)
    end

    it "caps the premium at 3% of the requested limit" do
      # EPL: 1200 + (15 * 1000) + (100_000_000 * 0.0008) = 1200 + 15000 + 80000 = 96200
      # Cap: 3% of 100_000 = 3000
      engine.add_submission(
        "sub-cap", "epl", "BigCo",
        employee_count: 1000, annual_revenue: 100_000_000,
        years_in_business: 5, industry_risk: "medium",
        requested_limit: 100_000, deductible: 1_000
      )
      expect(engine.calculate_base_premium("sub-cap")).to eq(3000)
    end

    it "floors the premium at 500" do
      # cap = 3% of 10_000 = 300 < 500 -> floor applies -> 500
      engine.add_submission(
        "sub-floor", "epl", "MicroCo",
        employee_count: 1, annual_revenue: 1_000,
        years_in_business: 5, industry_risk: "medium",
        requested_limit: 10_000, deductible: 500
      )
      expect(engine.calculate_base_premium("sub-floor")).to eq(500)
    end

    it "raises KeyError for an unknown submission" do
      expect { seeded_engine.calculate_base_premium("no-such") }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Prior claims and final premium with modifiers
  # ---------------------------------------------------------------------------
  describe "#record_prior_claim" do
    it "appends the claim" do
      seeded_engine.record_prior_claim("sub-epl", year: 2023, amount: 50_000, claim_type: "epl")
      s = seeded_engine.get_submission("sub-epl")
      expect(s[:prior_claims].size).to eq(1)
      expect(s[:prior_claims].first[:year]).to eq(2023)
      expect(s[:prior_claims].first[:amount]).to eq(50_000)
      expect(s[:prior_claims].first[:claim_type]).to eq("epl")
    end

    it "accumulates multiple claims" do
      seeded_engine.record_prior_claim("sub-epl", year: 2023, amount: 10_000, claim_type: "epl")
      seeded_engine.record_prior_claim("sub-epl", year: 2022, amount: 20_000, claim_type: "epl")
      expect(seeded_engine.get_submission("sub-epl")[:prior_claims].size).to eq(2)
    end

    it "raises KeyError for an unknown submission" do
      expect { seeded_engine.record_prior_claim("no-such", year: 2023, amount: 10_000, claim_type: "epl") }.to raise_error(KeyError)
    end
  end

  describe "#calculate_final_premium" do
    it "applies no modifiers for medium risk, mid tenure, no claims" do
      # sub-epl: medium x 1.0, 7 yrs x 1.0, no claims x 1.0
      result = seeded_engine.calculate_final_premium("sub-epl", current_year: 2025)
      expect(result[:base_premium]).to eq(5950)
      expect(result[:industry_modifier]).to eq(1.0)
      expect(result[:tenure_modifier]).to eq(1.0)
      expect(result[:claims_modifier]).to eq(1.0)
      expect(result[:final_premium]).to eq(5950)
    end

    it "compounds high risk and a young company" do
      # sub-do: high x 1.35, 2 yrs x 1.25, no claims x 1.0
      # final = 12500 * 1.35 * 1.25 = 21093.75 -> 21094
      result = seeded_engine.calculate_final_premium("sub-do", current_year: 2025)
      expect(result[:base_premium]).to eq(12500)
      expect(result[:industry_modifier]).to eq(1.35)
      expect(result[:tenure_modifier]).to eq(1.25)
      expect(result[:final_premium]).to eq(21094)
    end

    it "discounts low risk and a veteran company" do
      # sub-fid: low x 0.85, 15 yrs x 0.90, no claims x 1.0
      # final = 2000 * 0.85 * 0.90 = 1530
      result = seeded_engine.calculate_final_premium("sub-fid", current_year: 2025)
      expect(result[:base_premium]).to eq(2000)
      expect(result[:industry_modifier]).to eq(0.85)
      expect(result[:tenure_modifier]).to eq(0.90)
      expect(result[:final_premium]).to eq(1530)
    end

    it "applies a 15% bump for one recent qualifying claim" do
      seeded_engine.record_prior_claim("sub-epl", year: 2024, amount: 30_000, claim_type: "epl")
      result = seeded_engine.calculate_final_premium("sub-epl", current_year: 2025)
      expect(result[:claims_modifier]).to eq(1.15)
      expect(result[:final_premium]).to eq((5950 * 1.0 * 1.0 * 1.15).round)
    end

    it "ignores a claim of a non-matching type" do
      seeded_engine.record_prior_claim("sub-epl", year: 2024, amount: 30_000, claim_type: "do")
      result = seeded_engine.calculate_final_premium("sub-epl", current_year: 2025)
      expect(result[:claims_modifier]).to eq(1.0)
    end

    it "qualifies a claim_type of 'any' regardless of coverage" do
      seeded_engine.record_prior_claim("sub-epl", year: 2024, amount: 10_000, claim_type: "any")
      result = seeded_engine.calculate_final_premium("sub-epl", current_year: 2025)
      expect(result[:claims_modifier]).to eq(1.15)
    end

    it "ignores a claim older than 3 years" do
      seeded_engine.record_prior_claim("sub-epl", year: 2020, amount: 30_000, claim_type: "epl")
      result = seeded_engine.calculate_final_premium("sub-epl", current_year: 2025)
      expect(result[:claims_modifier]).to eq(1.0)
    end

    it "caps the claims modifier at 1.60" do
      # 5 EPL claims in last 3 years -> cap at x1.60
      [2023, 2023, 2024, 2024, 2025].each do |yr|
        seeded_engine.record_prior_claim("sub-epl", year: yr, amount: 10_000, claim_type: "epl")
      end
      result = seeded_engine.calculate_final_premium("sub-epl", current_year: 2025)
      expect(result[:claims_modifier]).to eq(1.60)
    end

    it "applies the deductible-derived floor" do
      # deductible/10 = 800; computed = 500 (floor from base); floor = max(500, 800, 500) = 800
      engine.add_submission(
        "sub-flr2", "fiduciary", "MicroCo",
        employee_count: 1, annual_revenue: 1_000,
        years_in_business: 5, industry_risk: "medium",
        requested_limit: 10_000, deductible: 8_000
      )
      result = engine.calculate_final_premium("sub-flr2", current_year: 2025)
      expect(result[:final_premium]).to eq(800)
    end

    it "raises KeyError for an unknown submission" do
      expect { seeded_engine.calculate_final_premium("no-such", current_year: 2025) }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Portfolio analytics
  # ---------------------------------------------------------------------------
  describe "#get_submissions_by_coverage_type" do
    it "groups submissions correctly" do
      grouped = seeded_engine.get_submissions_by_coverage_type
      expect(grouped).to include("epl", "do", "fiduciary")
      epl_ids = grouped["epl"].map { |s| s[:submission_id] }
      expect(epl_ids).to include("sub-epl")
    end

    it "sorts each group by submission_id" do
      [["sub-z", "epl"], ["sub-a", "epl"], ["sub-m", "epl"]].each do |sid, ct|
        engine.add_submission(
          sid, ct, "Co",
          employee_count: 10, annual_revenue: 500_000,
          years_in_business: 5, industry_risk: "medium",
          requested_limit: 250_000, deductible: 5_000
        )
      end
      ids = engine.get_submissions_by_coverage_type["epl"].map { |s| s[:submission_id] }
      expect(ids).to eq(ids.sort)
    end

    it "returns empty for an empty engine" do
      expect(engine.get_submissions_by_coverage_type).to eq({})
    end
  end

  describe "#get_portfolio_metrics" do
    it "reports the total submission count" do
      expect(seeded_engine.get_portfolio_metrics(current_year: 2025)[:total_submissions]).to eq(3)
    end

    it "reports total and average premium" do
      metrics = seeded_engine.get_portfolio_metrics(current_year: 2025)
      # sub-epl final=5950, sub-do final=21094, sub-fid final=1530
      expected_total = 5950 + 21094 + 1530
      expect(metrics[:total_premium]).to eq(expected_total)
      expect(metrics[:average_premium]).to eq((expected_total / 3.0).round)
    end

    it "reports per-coverage-type counts" do
      metrics = seeded_engine.get_portfolio_metrics(current_year: 2025)
      expect(metrics[:by_coverage_type]["epl"][:count]).to eq(1)
      expect(metrics[:by_coverage_type]["do"][:count]).to eq(1)
      expect(metrics[:by_coverage_type]["fiduciary"][:count]).to eq(1)
    end

    it "returns zeros for an empty engine" do
      metrics = engine.get_portfolio_metrics(current_year: 2025)
      expect(metrics[:total_submissions]).to eq(0)
      expect(metrics[:total_premium]).to eq(0)
      expect(metrics[:by_coverage_type]).to eq({})
    end
  end

  describe "#get_high_risk_submissions" do
    it "returns submissions above the threshold" do
      # sub-do has modifier 1.35 x 1.25 = 1.6875 > 1.3 threshold
      ids = seeded_engine.get_high_risk_submissions(current_year: 2025, modifier_threshold: 1.30).map { |r| r[:submission_id] }
      expect(ids).to include("sub-do")
    end

    it "excludes submissions below the threshold" do
      # sub-epl and sub-fid have effective modifiers <= 1.0
      ids = seeded_engine.get_high_risk_submissions(current_year: 2025, modifier_threshold: 1.30).map { |r| r[:submission_id] }
      expect(ids).not_to include("sub-epl", "sub-fid")
    end

    it "includes all required fields" do
      results = seeded_engine.get_high_risk_submissions(current_year: 2025, modifier_threshold: 1.0)
      results.each do |r|
        expect(r).to include(:submission_id, :company_name, :coverage_type, :effective_modifier, :final_premium)
      end
    end

    it "sorts by modifier descending" do
      modifiers = seeded_engine.get_high_risk_submissions(current_year: 2025, modifier_threshold: 0.0).map { |r| r[:effective_modifier] }
      expect(modifiers).to eq(modifiers.sort.reverse)
    end

    it "rounds the effective modifier to 4 decimal places" do
      results = seeded_engine.get_high_risk_submissions(current_year: 2025, modifier_threshold: 0.0)
      results.each do |r|
        expect(r[:effective_modifier]).to eq(r[:effective_modifier].round(4))
      end
    end
  end
end
