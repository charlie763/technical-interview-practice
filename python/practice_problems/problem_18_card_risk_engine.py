"""
=============================================================================
INTERVIEW PROBLEM 18: Card Transaction Risk Engine
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You're building the real-time risk engine for a card payments platform.
Every authorization request must be evaluated against a set of fraud rules
before the platform decides whether to approve, flag, or decline it. Rules
range from simple static thresholds to velocity checks that look at a
card's recent transaction history.

For this problem you are building a RiskEngine class.
Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between RiskEngine
instances — avoid them.
You choose the internal data structures; the public interface is what
matters.

DECISION MODEL
--------------
evaluate_transaction() always returns a dict shaped like:
  {
    "decision": "approve" | "review" | "decline",
    "reasons":  list[str],   # every rule code that matched, in the order
                              # the rules are listed below; [] if approve
  }

Severity ordering when multiple rules match: decline > review > approve.
The returned "decision" is the most severe outcome of any matched rule;
"reasons" always lists every matched rule code, not just the most severe.

Timestamps are ISO-8601 strings without timezone offset, e.g.
"2024-01-01T10:00:00". Use datetime.fromisoformat() for arithmetic.

# Example
# engine = RiskEngine()
# engine.evaluate_transaction("card_1", 200.0, "grocery", "2024-01-01T10:00:00")
# # -> {"decision": "approve", "reasons": []}
# engine.evaluate_transaction("card_1", 6000.0, "electronics", "2024-01-01T10:00:05")
# # -> {"decision": "decline", "reasons": ["amount_limit_exceeded"]}
# engine.evaluate_transaction("card_1", 1500.0, "travel", "2024-01-01T10:00:10")
# # -> {"decision": "review", "reasons": ["high_amount"]}
=============================================================================
"""

class RiskEngine:
    def __init__(self):
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 1 — Static rule evaluation  (~15 min)
    # -------------------------------------------------------------------------

    def evaluate_transaction(
        self,
        card_id: str,
        amount: float,
        merchant_category: str,
        timestamp: str,
    ) -> dict:
        """
        Record the transaction against card_id, then evaluate it against the
        fixed rules below and return a decision dict (see DECISION MODEL).

        Static rules (Part 1):
          - amount > 5000
                -> matches "amount_limit_exceeded" (decline)
          - merchant_category in {"gambling", "cash_advance", "crypto_exchange"}
                -> matches "blocked_merchant_category" (decline)
          - 1000 <= amount <= 5000
                -> matches "high_amount" (review)
                   (only relevant if nothing else already declined it)

        A transaction can match more than one rule; "reasons" must include
        every rule code that matched.
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 2 — Velocity checks  (~15 min)
    # -------------------------------------------------------------------------

    # No new public method for Part 2. Extend evaluate_transaction (above) to
    # also check the two velocity rules below, using the per-card transaction
    # history recorded in Part 1. Velocity rules are evaluated in addition to
    # the Part 1 static rules; "reasons" and "decision" combine matches from
    # both.
    #
    # Velocity rules (Part 2), evaluated against card_id's history *including*
    # the current transaction:
    #   - More than 3 transactions within a trailing 60-second window
    #         -> matches "velocity_count_exceeded" (decline)
    #   - Sum of amounts within a trailing 600-second window exceeds 10000
    #         -> matches "velocity_amount_exceeded" (review)
    #
    # "Trailing window" means: transactions with timestamp in
    # [current_ts - window_secs, current_ts].

    # -------------------------------------------------------------------------
    # PART 3 — Auto-blocking  (~15 min)
    # -------------------------------------------------------------------------

    def is_card_blocked(self, card_id: str) -> bool:
        """
        Return True if card_id is currently blocked: it has accumulated 3 or
        more "decline" decisions (from Parts 1 or 2, for any reason) within a
        trailing 24-hour (86400s) window, measured from the most recent
        transaction evaluated for that card.
        Return False for unknown cards.
        """
        raise NotImplementedError

    # Extend evaluate_transaction once more: if is_card_blocked(card_id) is
    # True *before* evaluating the new transaction, evaluate_transaction must
    # still record the transaction (so history/velocity stay accurate) but
    # short-circuit the rule checks above and return:
    #     {"decision": "decline", "reasons": ["card_blocked"]}
    # without re-running the Part 1/2 rules.
