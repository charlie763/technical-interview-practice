require "set"

COVERAGE_TYPES = Set.new(%w[epl do fiduciary]).freeze
INDUSTRY_RISKS = Set.new(%w[low medium high]).freeze

class PremiumRatingEngine
  def initialize
    @submissions = {}
  end

  def add_submission(
    submission_id, coverage_type, company_name,
    employee_count:, annual_revenue:, years_in_business:,
    industry_risk:, requested_limit:, deductible:
  )
    raise ArgumentError, "duplicate submission_id: #{submission_id}" if @submissions.key?(submission_id)
    raise ArgumentError, "invalid coverage_type: #{coverage_type}" unless COVERAGE_TYPES.include?(coverage_type)
    raise ArgumentError, "invalid industry_risk: #{industry_risk}" unless INDUSTRY_RISKS.include?(industry_risk)

    submission = {
      submission_id: submission_id,
      coverage_type: coverage_type,
      company_name: company_name,
      employee_count: employee_count,
      annual_revenue: annual_revenue,
      years_in_business: years_in_business,
      industry_risk: industry_risk,
      requested_limit: requested_limit,
      deductible: deductible,
      prior_claims: [],
    }
    @submissions[submission_id] = submission
    submission
  end

  def get_submission(submission_id)
    fetch_submission(submission_id)
  end

  def calculate_base_premium(submission_id)
    s = fetch_submission(submission_id)

    raw = case s[:coverage_type]
          when "epl" then 1200 + (15 * s[:employee_count]) + (s[:annual_revenue] * 0.0008)
          when "do" then 2500 + (s[:annual_revenue] * 0.0010)
          when "fiduciary" then 800 + (s[:annual_revenue] * 0.0004)
          end

    capped = [raw, s[:requested_limit] * 0.03].min
    [capped, 500].max.round
  end

  def record_prior_claim(submission_id, year:, amount:, claim_type:)
    s = fetch_submission(submission_id)
    s[:prior_claims] << { year: year, amount: amount, claim_type: claim_type }
    s
  end

  def calculate_final_premium(submission_id, current_year:)
    s = fetch_submission(submission_id)
    base = calculate_base_premium(submission_id)

    industry_modifier = { "low" => 0.85, "medium" => 1.00, "high" => 1.35 }.fetch(s[:industry_risk])
    tenure_modifier =
      if s[:years_in_business] < 3
        1.25
      elsif s[:years_in_business] > 10
        0.90
      else
        1.00
      end

    qualifying_claims = s[:prior_claims].count do |c|
      (c[:claim_type] == s[:coverage_type] || c[:claim_type] == "any") && c[:year] >= current_year - 2
    end
    claims_modifier = [1.0 + (0.15 * qualifying_claims), 1.60].min

    raw_final = base * industry_modifier * tenure_modifier * claims_modifier
    final_premium = [raw_final, s[:deductible] / 10, 500].max.round

    {
      base_premium: base,
      industry_modifier: industry_modifier,
      tenure_modifier: tenure_modifier,
      claims_modifier: claims_modifier,
      final_premium: final_premium,
    }
  end

  def get_submissions_by_coverage_type
    @submissions.values
      .group_by { |s| s[:coverage_type] }
      .transform_values { |list| list.sort_by { |s| s[:submission_id] } }
  end

  def get_portfolio_metrics(current_year:)
    by_coverage = {}
    total_premium = 0

    @submissions.each_value do |s|
      final = calculate_final_premium(s[:submission_id], current_year: current_year)[:final_premium]
      total_premium += final
      entry = (by_coverage[s[:coverage_type]] ||= { count: 0, total_premium: 0 })
      entry[:count] += 1
      entry[:total_premium] += final
    end

    {
      total_submissions: @submissions.size,
      total_premium: total_premium,
      average_premium: @submissions.empty? ? 0 : (total_premium.to_f / @submissions.size).round,
      by_coverage_type: by_coverage,
    }
  end

  def get_high_risk_submissions(current_year:, modifier_threshold:)
    @submissions.values
      .filter_map do |s|
        result = calculate_final_premium(s[:submission_id], current_year: current_year)
        modifier = result[:final_premium].to_f / result[:base_premium]
        next if modifier <= modifier_threshold

        {
          submission_id: s[:submission_id],
          company_name: s[:company_name],
          coverage_type: s[:coverage_type],
          effective_modifier: modifier.round(4),
          final_premium: result[:final_premium],
        }
      end
      .sort_by { |e| [-e[:effective_modifier], e[:submission_id]] }
  end

  private

  def fetch_submission(submission_id)
    @submissions.fetch(submission_id) { raise KeyError, "submission not found: #{submission_id}" }
  end
end
