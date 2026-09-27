# =============================================================================
# INTERVIEW PROBLEM 16: Policy Premium Rating Engine
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# You're building the premium rating engine for a management and
# professional liability insurance platform. Brokers submit applications
# on behalf of their clients; underwriters use your engine to calculate
# base premiums, apply risk modifiers, and analyze portfolio exposure.
#
# For this problem you are building a PremiumRatingEngine class. Store all
# state in instance variables initialized in `initialize`. Class-level
# variables will bleed between tests and between instances — avoid them.
# You choose the internal data structures; the public interface is what
# matters.
#
# DATA MODEL
# ----------
# PolicySubmission:
#   {
#     submission_id:      String,
#     coverage_type:      String,       # "epl" | "do" | "fiduciary"
#     company_name:       String,
#     employee_count:     Integer,
#     annual_revenue:     Integer,      # dollars
#     years_in_business:  Integer,
#     industry_risk:      String,       # "low" | "medium" | "high"
#     requested_limit:    Integer,      # policy limit in dollars
#     deductible:         Integer,      # self-insured retention in dollars
#     prior_claims:       Array,        # array of PriorClaim hashes (see below)
#   }
#
# PriorClaim:
#   {
#     year:        Integer,  # calendar year of the claim
#     amount:      Integer,  # dollars paid/reserved
#     claim_type:  String,   # e.g. "epl", "do", "fiduciary"
#   }
#
# PREMIUM RATING RULES
# ---------------------
# Base premium by coverage type (Part 1):
#   EPL:       $1,200 + ($15 x employee_count) + (annual_revenue x 0.0008)
#   D&O:       $2,500 + (annual_revenue x 0.0010)
#   Fiduciary: $800   + (annual_revenue x 0.0004)
#   Cap:       base_premium may not exceed 3% of requested_limit
#   Floor:     base_premium may not be less than $500
#
# Risk modifiers applied to base premium (Part 2):
#   industry_risk:
#     "low"    -> x 0.85
#     "medium" -> x 1.00
#     "high"   -> x 1.35
#   years_in_business:
#     < 3      -> x 1.25
#     3-10     -> x 1.00
#     > 10     -> x 0.90
#   prior_claims in the last 3 years (relative to the current_year argument):
#     Each qualifying claim adds +15% (multiplicative modifier cap: +60%)
#     Qualifying = claim_type matches coverage_type OR claim_type is "any"
#
#   Final premium = base x industry_modifier x tenure_modifier x claims_modifier
#   Final premium floor: max(final, deductible / 10, 500)
#
# Example
# -------
#   engine = PremiumRatingEngine.new
#   engine.add_submission(
#     "sub-001", "epl", "Acme Corp",
#     employee_count: 50, annual_revenue: 5_000_000,
#     years_in_business: 7, industry_risk: "medium",
#     requested_limit: 1_000_000, deductible: 10_000,
#   )
#   engine.calculate_base_premium("sub-001")
#   # -> 1200 + (15 * 50) + (5_000_000 * 0.0008) = 1200 + 750 + 4000 = 5950
#   engine.record_prior_claim("sub-001", year: 2023, amount: 50_000, claim_type: "epl")
#   engine.calculate_final_premium("sub-001", current_year: 2025)
#   # -> base=5950, industry=x1.0, tenure=x1.0, claims=x1.15 -> final=6842 (rounded)
#
# =============================================================================
# PART 1 — Submission management and base premium calculation
# =============================================================================
# Implement add_submission, get_submission, and calculate_base_premium.

require "set"

COVERAGE_TYPES = Set.new(%w[epl do fiduciary]).freeze
INDUSTRY_RISKS = Set.new(%w[low medium high]).freeze

class PremiumRatingEngine
  def initialize
    raise NotImplementedError
  end

  # ── Part 1 ────────────────────────────────────────────────────────────────

  # Register a new policy submission.
  # Returns the stored submission hash (prior_claims starts as an empty array).
  # Raise ArgumentError if submission_id already exists, coverage_type is
  # not one of "epl"/"do"/"fiduciary", or industry_risk is not one of
  # "low"/"medium"/"high".
  def add_submission(
    submission_id, coverage_type, company_name,
    employee_count:, annual_revenue:, years_in_business:,
    industry_risk:, requested_limit:, deductible:
  )
    raise NotImplementedError
  end

  # Return the submission hash. Raise KeyError if submission_id does not exist.
  def get_submission(submission_id)
    raise NotImplementedError
  end

  # Calculate and return the base premium (Integer, dollars) using the
  # PREMIUM RATING RULES defined above.
  #
  # Apply the cap (3% of requested_limit) and floor ($500) after computing
  # the type-specific formula.
  #
  # Does NOT persist the result — call this as many times as you like.
  # Raise KeyError if submission_id does not exist.
  def calculate_base_premium(submission_id)
    raise NotImplementedError
  end

  # ── Part 2 ────────────────────────────────────────────────────────────────

  # Append a PriorClaim to the submission's prior_claims array. claim_type
  # is the coverage line the claim was made under (e.g. "epl", "do",
  # "fiduciary"); use "any" to match all coverage types.
  # Returns the updated submission hash.
  # Raise KeyError if submission_id does not exist.
  def record_prior_claim(submission_id, year:, amount:, claim_type:)
    raise NotImplementedError
  end

  # Apply risk modifiers to the base premium and return a breakdown.
  #
  # Calls calculate_base_premium internally — do not duplicate its logic.
  #
  # A "qualifying" prior claim is one whose claim_type matches the
  # submission's coverage_type OR whose claim_type is "any", AND whose year
  # is within the last 3 years (i.e. year >= current_year - 2).
  #
  # Each qualifying claim adds x1.15 to the claims modifier, capped at
  # x1.60 total (i.e. max 4 qualifying claims make a difference).
  #
  # Final premium floor: max(computed_final, deductible / 10, 500).
  # Round to the nearest dollar (Integer).
  #
  # Returns:
  #   {
  #     base_premium:       Integer,
  #     industry_modifier:  Float,
  #     tenure_modifier:    Float,
  #     claims_modifier:    Float,
  #     final_premium:      Integer,
  #   }
  #
  # Raise KeyError if submission_id does not exist.
  def calculate_final_premium(submission_id, current_year:)
    raise NotImplementedError
  end

  # ── Part 3 ────────────────────────────────────────────────────────────────

  # Return all submissions grouped by coverage_type.
  #
  # Keys are coverage types that have at least one submission.
  # Values are arrays of submission hashes, sorted by submission_id.
  def get_submissions_by_coverage_type
    raise NotImplementedError
  end

  # Aggregate statistics across all submissions.
  #
  # Calls calculate_final_premium for each submission — do not duplicate
  # its logic.
  #
  # Returns:
  #   {
  #     total_submissions:  Integer,
  #     total_premium:      Integer,  # sum of all final premiums
  #     average_premium:    Integer,  # rounded to nearest dollar
  #     by_coverage_type: {
  #       coverage_type => { count: Integer, total_premium: Integer },
  #       ...  # only types with submissions
  #     },
  #   }
  def get_portfolio_metrics(current_year:)
    raise NotImplementedError
  end

  # Return submissions whose effective overall modifier exceeds the threshold.
  #
  # Effective modifier = final_premium / base_premium.
  #
  # Calls calculate_final_premium internally.
  #
  # modifier_threshold: e.g. 1.30 — return submissions with a modifier > 1.30.
  #
  # Each item (sorted by effective_modifier descending, then submission_id
  # ascending):
  #   {
  #     submission_id:        String,
  #     company_name:         String,
  #     coverage_type:        String,
  #     effective_modifier:   Float,  # rounded to 4 decimal places
  #     final_premium:        Integer,
  #   }
  def get_high_risk_submissions(current_year:, modifier_threshold:)
    raise NotImplementedError
  end
end
