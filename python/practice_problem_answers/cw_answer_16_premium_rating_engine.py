"""
=============================================================================
INTERVIEW PROBLEM 16: Policy Premium Rating Engine
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You're building the premium rating engine for a management and professional
liability insurance platform. Brokers submit applications on behalf of their
clients; underwriters use your engine to calculate base premiums, apply risk
modifiers, and analyze portfolio exposure.

For this problem you are building a PremiumRatingEngine class.
Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between instances — avoid them.
You choose the internal data structures; the public interface is what matters.

DATA MODEL
----------
PolicySubmission:
  {
    "submission_id":    str,
    "coverage_type":    str,       # "epl" | "do" | "fiduciary"
    "company_name":     str,
    "employee_count":   int,
    "annual_revenue":   int,       # dollars
    "years_in_business": int,
    "industry_risk":    str,       # "low" | "medium" | "high"
    "requested_limit":  int,       # policy limit in dollars
    "deductible":       int,       # self-insured retention in dollars
    "prior_claims":     list,      # list of PriorClaim dicts (see below)
  }

PriorClaim:
  {
    "year":        int,    # calendar year of the claim
    "amount":      int,    # dollars paid/reserved
    "claim_type":  str,    # e.g. "epl", "do", "fiduciary"
  }

PREMIUM RATING RULES
--------------------
Base premium by coverage type (Part 1):
  EPL:       $1,200 + ($15 × employee_count) + (annual_revenue × 0.0008)
  D&O:       $2,500 + (annual_revenue × 0.0010)
  Fiduciary: $800   + (annual_revenue × 0.0004)
  Cap:       base_premium may not exceed 3% of requested_limit
  Floor:     base_premium may not be less than $500

Risk modifiers applied to base premium (Part 2):
  industry_risk:
    "low"    → × 0.85
    "medium" → × 1.00
    "high"   → × 1.35
  years_in_business:
    < 3      → × 1.25
    3–10     → × 1.00
    > 10     → × 0.90
  prior_claims in the last 3 years (relative to the current_year argument):
    Each qualifying claim adds +15% (multiplicative modifier cap: +60%)
    Qualifying = claim_type matches coverage_type OR claim_type is "any"

  Final premium = base × industry_modifier × tenure_modifier × claims_modifier
  Final premium floor: max(final, deductible // 10, 500)

# Example
# engine = PremiumRatingEngine()
# engine.add_submission(
#     "sub-001", "epl", "Acme Corp",
#     employee_count=50, annual_revenue=5_000_000,
#     years_in_business=7, industry_risk="medium",
#     requested_limit=1_000_000, deductible=10_000,
# )
# engine.calculate_base_premium("sub-001")
# -> 1200 + (15 * 50) + (5_000_000 * 0.0008) = 1200 + 750 + 4000 = 5950
# engine.record_prior_claim("sub-001", year=2023, amount=50000, claim_type="epl")
# engine.calculate_final_premium("sub-001", current_year=2025)
# -> base=5950, industry=×1.0, tenure=×1.0, claims=×1.15 → final=6842 (rounded)

=============================================================================
PART 1 — Submission management and base premium calculation
=============================================================================

Implement `add_submission`, `get_submission`, and `calculate_base_premium`.

"""

from __future__ import annotations

COVERAGE_TYPES = {"epl", "do", "fiduciary"}
INDUSTRY_RISKS = {"low", "medium", "high"}


class PremiumRatingEngine:
    """
    Calculates premiums for management and professional liability submissions.
    """

    def __init__(self):
        raise NotImplementedError

    # ── Part 1 ────────────────────────────────────────────────────────────────

    def add_submission(
        self,
        submission_id: str,
        coverage_type: str,
        company_name: str,
        employee_count: int,
        annual_revenue: int,
        years_in_business: int,
        industry_risk: str,
        requested_limit: int,
        deductible: int,
    ) -> dict:
        """
        Register a new policy submission.

        Parameters
        ----------
        submission_id : str
            Unique identifier for the submission.
        coverage_type : str
            One of "epl", "do", or "fiduciary".
        company_name : str
        employee_count : int
        annual_revenue : int
            Annual revenue in dollars.
        years_in_business : int
        industry_risk : str
            One of "low", "medium", or "high".
        requested_limit : int
            Requested policy limit in dollars.
        deductible : int
            Self-insured retention in dollars.

        Returns
        -------
        dict
            The stored submission dict (prior_claims starts as empty list).

        Raises
        ------
        ValueError
            If submission_id already exists, coverage_type is not one of
            "epl"/"do"/"fiduciary", or industry_risk is not one of
            "low"/"medium"/"high".
        """
        raise NotImplementedError

    def get_submission(self, submission_id: str) -> dict:
        """
        Return the submission dict.

        Raises
        ------
        KeyError
            If submission_id does not exist.
        """
        raise NotImplementedError

    def calculate_base_premium(self, submission_id: str) -> int:
        """
        Calculate and return the base premium (integer, dollars) using the
        PREMIUM RATING RULES defined in the module docstring.

        Apply the cap (3% of requested_limit) and floor ($500) after computing
        the type-specific formula.

        Does NOT persist the result — call this as many times as you like.

        Raises
        ------
        KeyError
            If submission_id does not exist.
        """
        raise NotImplementedError

    # ── Part 2 ────────────────────────────────────────────────────────────────

    def record_prior_claim(
        self,
        submission_id: str,
        year: int,
        amount: int,
        claim_type: str,
    ) -> dict:
        """
        Append a PriorClaim to the submission's prior_claims list.

        Parameters
        ----------
        submission_id : str
        year : int
            Calendar year of the claim.
        amount : int
            Dollars paid or reserved.
        claim_type : str
            The coverage line the claim was made under (e.g. "epl", "do",
            "fiduciary"). Use "any" to match all coverage types.

        Returns
        -------
        dict
            The updated submission dict.

        Raises
        ------
        KeyError
            If submission_id does not exist.
        """
        raise NotImplementedError

    def calculate_final_premium(
        self,
        submission_id: str,
        current_year: int,
    ) -> dict:
        """
        Apply risk modifiers to the base premium and return a breakdown.

        Calls `calculate_base_premium` internally — do not duplicate its logic.

        A "qualifying" prior claim is one whose claim_type matches the
        submission's coverage_type OR whose claim_type is "any", AND whose
        year is within the last 3 years (i.e. year >= current_year - 2).

        Each qualifying claim adds ×1.15 to the claims modifier, capped at
        ×1.60 total (i.e. max 4 qualifying claims make a difference).

        Final premium floor: max(computed_final, deductible // 10, 500).
        Round to the nearest dollar (int).

        Parameters
        ----------
        submission_id : str
        current_year : int
            Reference year for evaluating "last 3 years" of prior claims.

        Returns
        -------
        dict
            {
              "base_premium":        int,
              "industry_modifier":   float,
              "tenure_modifier":     float,
              "claims_modifier":     float,
              "final_premium":       int,
            }

        Raises
        ------
        KeyError
            If submission_id does not exist.
        """
        raise NotImplementedError

    # ── Part 3 ────────────────────────────────────────────────────────────────

    def get_submissions_by_coverage_type(self) -> dict[str, list[dict]]:
        """
        Return all submissions grouped by coverage_type.

        Returns
        -------
        dict
            Keys are coverage types that have at least one submission.
            Values are lists of submission dicts, sorted by submission_id.
        """
        raise NotImplementedError

    def get_portfolio_metrics(self, current_year: int) -> dict:
        """
        Aggregate statistics across all submissions.

        Calls `calculate_final_premium` for each submission — do not duplicate
        its logic.

        Returns
        -------
        dict
            {
              "total_submissions":  int,
              "total_premium":      int,   # sum of all final premiums
              "average_premium":    int,   # rounded to nearest dollar
              "by_coverage_type":   {
                  coverage_type: {
                      "count":           int,
                      "total_premium":   int,
                  },
                  ...                      # only types with submissions
              },
            }
        """
        raise NotImplementedError

    def get_high_risk_submissions(
        self,
        current_year: int,
        modifier_threshold: float,
    ) -> list[dict]:
        """
        Return submissions whose effective overall modifier exceeds the
        threshold.

        Effective modifier = final_premium / base_premium.

        Calls `calculate_final_premium` internally.

        Parameters
        ----------
        current_year : int
        modifier_threshold : float
            E.g. 1.30 — return submissions with a modifier > 1.30.

        Returns
        -------
        list[dict]
            Each item:
            {
              "submission_id":    str,
              "company_name":     str,
              "coverage_type":    str,
              "effective_modifier": float,  # rounded to 4 decimal places
              "final_premium":    int,
            }
            Sorted by effective_modifier descending, then submission_id ascending.
        """
        raise NotImplementedError
