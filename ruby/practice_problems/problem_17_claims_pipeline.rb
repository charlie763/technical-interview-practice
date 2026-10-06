# =============================================================================
# INTERVIEW PROBLEM 17: Claims Processing Pipeline
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# You're building the claims processing system for a management liability
# insurance platform. When a policyholder experiences a covered incident
# (e.g. an employment lawsuit, a D&O action), they file a claim. Claims
# move through a multi-stage review pipeline from filing through
# investigation, evaluation, and ultimately settlement or denial.
#
# For this problem you are building a ClaimsPipeline class. Store all
# state in instance variables initialized in `initialize`. Class-level
# variables will bleed between tests and between instances — avoid them.
# You choose the internal data structures; the public interface is what
# matters.
#
# DATA MODEL
# ----------
# Claim:
#   {
#     claim_id:         String,
#     policy_id:        String,
#     coverage_type:    String,       # "epl" | "do" | "fiduciary"
#     incident_date:    String,       # ISO-8601 date string, e.g. "2025-03-15"
#     filed_at:         String,       # ISO-8601 datetime
#     status:           String,       # current status (see STATE MACHINE below)
#     claimed_amount:   Integer,      # dollars claimed by policyholder
#     reserve_amount:   Integer,      # current reserve estimate
#     approved_amount:  Integer | nil,  # set only when settled
#     events:           Array,        # ordered array of ClaimEvent hashes
#   }
#
# ClaimEvent:
#   {
#     at:      String,  # ISO-8601 datetime
#     actor:   String,  # adjuster ID or "system"
#     action:  String,  # "filed" | "status_change" | "reserve_update" | "settled" | "denied"
#     payload: Hash,    # action-specific data
#   }
#
# STATE MACHINE
# -------------
# Valid transitions:
#   filed          -> investigating
#   investigating  -> evaluation
#   evaluation     -> settled | denied
#   settled        -> closed
#   denied         -> closed
#   closed         -> (terminal)
#
# Timestamps are ISO-8601 strings.
#
# Example
# -------
#   pipeline = ClaimsPipeline.new
#   pipeline.file_claim(
#     "clm-001", policy_id: "pol-101", coverage_type: "epl",
#     incident_date: "2025-01-15", filed_at: "2025-02-01T09:00:00",
#     claimed_amount: 75_000, reserve_amount: 50_000, actor: "adjuster-1",
#   )
#   pipeline.advance_status("clm-001", "investigating", at: "2025-02-03T10:00:00", actor: "adjuster-1")
#   pipeline.get_claim("clm-001")[:status]  # -> "investigating"
#   pipeline.update_reserve("clm-001", 60_000, at: "2025-02-10T14:00:00", actor: "adjuster-1")
#   pipeline.get_claim("clm-001")[:reserve_amount]  # -> 60_000
#
# =============================================================================
# PART 1 — Claim filing, status transitions, and reserve updates
# =============================================================================
# Implement file_claim, get_claim, advance_status, and update_reserve.

CLAIM_VALID_TRANSITIONS = {
  "filed" => ["investigating"],
  "investigating" => ["evaluation"],
  "evaluation" => %w[settled denied],
  "settled" => ["closed"],
  "denied" => ["closed"],
  "closed" => [],
}.freeze

CLAIM_TERMINAL_STATES = ["closed"].freeze

class ClaimsPipeline
  def initialize
    raise NotImplementedError
  end

  # ── Part 1 ────────────────────────────────────────────────────────────────

  # Register a new claim in the "filed" state and append an initial
  # ClaimEvent with action: "filed".
  # Returns the stored claim hash (approved_amount starts as nil).
  #
  # Raise ArgumentError if claim_id already exists.
  # Raise ArgumentError if claimed_amount <= 0 or reserve_amount <= 0.
  def file_claim(
    claim_id, policy_id:, coverage_type:, incident_date:, filed_at:,
    claimed_amount:, reserve_amount:, actor:
  )
    raise NotImplementedError
  end

  # Return the claim hash. Raise KeyError if claim_id does not exist.
  def get_claim(claim_id)
    raise NotImplementedError
  end

  # Move a claim to a new status if the transition is valid, and append a
  # ClaimEvent with action: "status_change" and
  # payload: { from_status: ..., to_status: ... }.
  #
  # Do NOT use this method to settle or deny — use settle_claim and
  # deny_claim from Part 2 for those transitions.
  #
  # to_status must be reachable via CLAIM_VALID_TRANSITIONS from the
  # current status.
  #
  # Raise KeyError if claim_id does not exist.
  # Raise ArgumentError if to_status is "settled" or "denied" (use the
  # dedicated methods).
  # Raise ArgumentError if the transition is not in CLAIM_VALID_TRANSITIONS
  # from the current state.
  def advance_status(claim_id, to_status, at:, actor:)
    raise NotImplementedError
  end

  # Update the claim's reserve_amount and append a ClaimEvent with
  # action: "reserve_update" and payload: { old_reserve: ..., new_reserve: ... }.
  #
  # new_reserve must be > 0.
  #
  # Raise KeyError if claim_id does not exist.
  # Raise ArgumentError if new_reserve <= 0.
  # Raise ArgumentError if the claim is in a terminal state ("closed").
  def update_reserve(claim_id, new_reserve, at:, actor:)
    raise NotImplementedError
  end

  # ── Part 2 ────────────────────────────────────────────────────────────────

  # Settle a claim: transition it from "evaluation" -> "settled", set
  # approved_amount, and append a ClaimEvent with action: "settled" and
  # payload: { approved_amount: ... }.
  #
  # approved_amount must be > 0 and <= claimed_amount.
  #
  # Raise KeyError if claim_id does not exist.
  # Raise ArgumentError if the claim is not in "evaluation" status.
  # Raise ArgumentError if approved_amount <= 0 or approved_amount > claimed_amount.
  def settle_claim(claim_id, approved_amount:, settled_at:, actor:)
    raise NotImplementedError
  end

  # Deny a claim: transition it from "evaluation" -> "denied" and append a
  # ClaimEvent with action: "denied" and payload: { reason: ... }.
  #
  # Raise KeyError if claim_id does not exist.
  # Raise ArgumentError if the claim is not in "evaluation" status.
  def deny_claim(claim_id, reason:, denied_at:, actor:)
    raise NotImplementedError
  end

  # Return all claims for the given policy_id, sorted by filed_at ascending.
  # Returns an empty array if no claims exist for that policy.
  def get_claims_by_policy(policy_id)
    raise NotImplementedError
  end

  # Return all claims that are NOT in a terminal state ("closed") and NOT
  # denied. Sorted by filed_at ascending.
  def get_open_claims
    raise NotImplementedError
  end

  # ── Part 3 ────────────────────────────────────────────────────────────────

  # Analyse whether reserves cover settled amounts across all claims.
  #
  # For settled claims, compare the reserve_amount at the time of
  # settlement (the current reserve_amount field) with the approved_amount.
  # A claim is "under-reserved" if reserve_amount < approved_amount.
  #
  # Returns:
  #   {
  #     total_reserves:        Integer,  # sum of reserve_amount for ALL claims
  #     total_approved:        Integer,  # sum of approved_amount for settled claims
  #     under_reserved_count:  Integer,  # count of settled claims where reserve < approved
  #     under_reserved_gap:    Integer,  # sum of (approved - reserve) for under-reserved claims
  #   }
  def get_reserve_adequacy
    raise NotImplementedError
  end

  # Return aggregate statistics across all claims.
  #
  # settlement_ratio for a single claim = approved_amount / claimed_amount.
  # avg_settlement_ratio = mean of settlement_ratios for all SETTLED claims
  # (0.0 if no settled claims). Round to 4 decimal places.
  #
  # Returns:
  #   {
  #     total:                Integer,
  #     by_status:            { status => count, ... },  # only statuses with count > 0
  #     total_claimed:        Integer,  # sum of claimed_amount across all claims
  #     total_paid:           Integer,  # sum of approved_amount for settled claims
  #     avg_settlement_ratio: Float,    # rounded to 4 decimal places
  #   }
  def get_claims_metrics
    raise NotImplementedError
  end

  # Summarise claim history for a specific policy.
  #
  # loss_ratio = total_paid / total_claimed (0.0 if total_claimed == 0).
  # Round loss_ratio to 4 decimal places.
  #
  # Returns:
  #   {
  #     policy_id:      String,
  #     claim_count:    Integer,
  #     total_claimed:  Integer,
  #     total_paid:     Integer,  # sum of approved_amount for settled claims
  #     loss_ratio:     Float,
  #   }
  #
  # Note: returns a hash with zero values if no claims exist for the policy.
  def get_policy_loss_history(policy_id)
    raise NotImplementedError
  end
end
