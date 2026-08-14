"""
=============================================================================
INTERVIEW PROBLEM 17: Claims Processing Pipeline
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You're building the claims processing system for a management liability
insurance platform. When a policyholder experiences a covered incident
(e.g. an employment lawsuit, a D&O action), they file a claim. Claims move
through a multi-stage review pipeline from filing through investigation,
evaluation, and ultimately settlement or denial.

For this problem you are building a ClaimsPipeline class.
Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between instances — avoid them.
You choose the internal data structures; the public interface is what matters.

DATA MODEL
----------
Claim:
  {
    "claim_id":        str,
    "policy_id":       str,
    "coverage_type":   str,       # "epl" | "do" | "fiduciary"
    "incident_date":   str,       # ISO-8601 date string, e.g. "2025-03-15"
    "filed_at":        str,       # ISO-8601 datetime
    "status":          str,       # current status (see STATE MACHINE below)
    "claimed_amount":  int,       # dollars claimed by policyholder
    "reserve_amount":  int,       # current reserve estimate
    "approved_amount": int | None,  # set only when settled
    "events":          list,      # ordered list of ClaimEvent dicts
  }

ClaimEvent:
  {
    "at":      str,   # ISO-8601 datetime
    "actor":   str,   # adjuster ID or "system"
    "action":  str,   # "filed" | "status_change" | "reserve_update" | "settled" | "denied"
    "payload": dict,  # action-specific data
  }

STATE MACHINE
-------------
Valid transitions:
  filed  →  investigating
  investigating  →  evaluation
  evaluation  →  settled | denied
  settled  →  closed
  denied   →  closed
  closed   →  (terminal)

Timestamps are ISO-8601 strings.
Use datetime.fromisoformat() for any date arithmetic.

# Example
# pipeline = ClaimsPipeline()
# pipeline.file_claim(
#     "clm-001", policy_id="pol-101", coverage_type="epl",
#     incident_date="2025-01-15", filed_at="2025-02-01T09:00:00",
#     claimed_amount=75_000, reserve_amount=50_000, actor="adjuster-1",
# )
# pipeline.advance_status("clm-001", "investigating", at="2025-02-03T10:00:00", actor="adjuster-1")
# pipeline.get_claim("clm-001")["status"]  # -> "investigating"
# pipeline.update_reserve("clm-001", 60_000, at="2025-02-10T14:00:00", actor="adjuster-1")
# pipeline.get_claim("clm-001")["reserve_amount"]  # -> 60_000

=============================================================================
PART 1 — Claim filing, status transitions, and reserve updates
=============================================================================

Implement `file_claim`, `get_claim`, `advance_status`, and `update_reserve`.

"""

from __future__ import annotations

VALID_TRANSITIONS: dict[str, list[str]] = {
    "filed":         ["investigating"],
    "investigating": ["evaluation"],
    "evaluation":    ["settled", "denied"],
    "settled":       ["closed"],
    "denied":        ["closed"],
    "closed":        [],
}

TERMINAL_STATES = {"closed"}


class ClaimsPipeline:
    """
    Manages the lifecycle and financial tracking of insurance claims.
    """

    def __init__(self):
        raise NotImplementedError

    # ── Part 1 ────────────────────────────────────────────────────────────────

    def file_claim(
        self,
        claim_id: str,
        policy_id: str,
        coverage_type: str,
        incident_date: str,
        filed_at: str,
        claimed_amount: int,
        reserve_amount: int,
        actor: str,
    ) -> dict:
        """
        Register a new claim in the "filed" state and append an initial
        ClaimEvent with action="filed".

        Parameters
        ----------
        claim_id : str
            Unique identifier.
        policy_id : str
        coverage_type : str
        incident_date : str
            ISO-8601 date (e.g. "2025-03-15").
        filed_at : str
            ISO-8601 datetime when the claim was submitted.
        claimed_amount : int
            Dollar amount the policyholder is claiming.
        reserve_amount : int
            Initial reserve estimate set by the adjuster.
        actor : str

        Returns
        -------
        dict
            The stored claim dict (approved_amount starts as None).

        Raises
        ------
        ValueError
            If claim_id already exists.
        ValueError
            If claimed_amount <= 0 or reserve_amount <= 0.
        """
        raise NotImplementedError

    def get_claim(self, claim_id: str) -> dict:
        """
        Return the claim dict.

        Raises
        ------
        KeyError
            If claim_id does not exist.
        """
        raise NotImplementedError

    def advance_status(
        self,
        claim_id: str,
        to_status: str,
        at: str,
        actor: str,
    ) -> dict:
        """
        Move a claim to a new status if the transition is valid, and append
        a ClaimEvent with action="status_change" and
        payload={"from_status": ..., "to_status": ...}.

        Do NOT use this method to settle or deny — use `settle_claim` and
        `deny_claim` from Part 2 for those transitions.

        Parameters
        ----------
        claim_id : str
        to_status : str
            Target status. Must be reachable via VALID_TRANSITIONS from the
            current status. Raises ValueError for "settled" or "denied" —
            those are handled by dedicated methods.
        at : str
            ISO-8601 datetime.
        actor : str

        Returns
        -------
        dict
            The updated claim dict.

        Raises
        ------
        KeyError
            If claim_id does not exist.
        ValueError
            If to_status is "settled" or "denied" (use the dedicated methods).
        ValueError
            If the transition is not in VALID_TRANSITIONS from the current state.
        """
        raise NotImplementedError

    def update_reserve(
        self,
        claim_id: str,
        new_reserve: int,
        at: str,
        actor: str,
    ) -> dict:
        """
        Update the claim's reserve_amount and append a ClaimEvent with
        action="reserve_update" and
        payload={"old_reserve": ..., "new_reserve": ...}.

        Parameters
        ----------
        claim_id : str
        new_reserve : int
            New reserve amount in dollars. Must be > 0.
        at : str
        actor : str

        Returns
        -------
        dict
            The updated claim dict.

        Raises
        ------
        KeyError
            If claim_id does not exist.
        ValueError
            If new_reserve <= 0.
        ValueError
            If the claim is in a terminal state ("closed").
        """
        raise NotImplementedError

    # ── Part 2 ────────────────────────────────────────────────────────────────

    def settle_claim(
        self,
        claim_id: str,
        approved_amount: int,
        settled_at: str,
        actor: str,
    ) -> dict:
        """
        Settle a claim: transition it from "evaluation" → "settled",
        set approved_amount, and append a ClaimEvent with action="settled"
        and payload={"approved_amount": ...}.

        Parameters
        ----------
        claim_id : str
        approved_amount : int
            Amount approved for payment. Must be > 0 and <= claimed_amount.
        settled_at : str
        actor : str

        Returns
        -------
        dict
            The updated claim dict.

        Raises
        ------
        KeyError
            If claim_id does not exist.
        ValueError
            If the claim is not in "evaluation" status.
        ValueError
            If approved_amount <= 0 or approved_amount > claimed_amount.
        """
        raise NotImplementedError

    def deny_claim(
        self,
        claim_id: str,
        reason: str,
        denied_at: str,
        actor: str,
    ) -> dict:
        """
        Deny a claim: transition it from "evaluation" → "denied" and append
        a ClaimEvent with action="denied" and payload={"reason": ...}.

        Parameters
        ----------
        claim_id : str
        reason : str
            Free-text reason for denial.
        denied_at : str
        actor : str

        Returns
        -------
        dict
            The updated claim dict.

        Raises
        ------
        KeyError
            If claim_id does not exist.
        ValueError
            If the claim is not in "evaluation" status.
        """
        raise NotImplementedError

    def get_claims_by_policy(self, policy_id: str) -> list[dict]:
        """
        Return all claims for the given policy_id, sorted by filed_at ascending.

        Returns an empty list if no claims exist for that policy.
        """
        raise NotImplementedError

    def get_open_claims(self) -> list[dict]:
        """
        Return all claims that are NOT in a terminal state ("closed") and
        NOT denied.

        Sorted by filed_at ascending.
        """
        raise NotImplementedError

    # ── Part 3 ────────────────────────────────────────────────────────────────

    def get_reserve_adequacy(self) -> dict:
        """
        Analyse whether reserves cover settled amounts across all claims.

        For settled claims, compare the reserve_amount at the time of
        settlement (the current reserve_amount field) with the approved_amount.
        A claim is "under-reserved" if reserve_amount < approved_amount.

        Returns
        -------
        dict
            {
              "total_reserves":       int,   # sum of reserve_amount for ALL claims
              "total_approved":       int,   # sum of approved_amount for settled claims
              "under_reserved_count": int,   # count of settled claims where reserve < approved
              "under_reserved_gap":   int,   # sum of (approved - reserve) for under-reserved claims
            }
        """
        raise NotImplementedError

    def get_claims_metrics(self) -> dict:
        """
        Return aggregate statistics across all claims.

        settlement_ratio for a single claim = approved_amount / claimed_amount.
        avg_settlement_ratio = mean of settlement_ratios for all SETTLED claims
        (0.0 if no settled claims). Round to 4 decimal places.

        Returns
        -------
        dict
            {
              "total":                 int,
              "by_status":             {status: count, ...},  # only statuses with count > 0
              "total_claimed":         int,   # sum of claimed_amount across all claims
              "total_paid":            int,   # sum of approved_amount for settled claims
              "avg_settlement_ratio":  float, # rounded to 4 decimal places
            }
        """
        raise NotImplementedError

    def get_policy_loss_history(self, policy_id: str) -> dict:
        """
        Summarise claim history for a specific policy.

        loss_ratio = total_paid / total_claimed (0.0 if total_claimed == 0).
        Round loss_ratio to 4 decimal places.

        Parameters
        ----------
        policy_id : str

        Returns
        -------
        dict
            {
              "policy_id":     str,
              "claim_count":   int,
              "total_claimed": int,
              "total_paid":    int,   # sum of approved_amount for settled claims
              "loss_ratio":    float,
            }

        Note: returns a dict with zero values if no claims exist for the policy.
        """
        raise NotImplementedError
