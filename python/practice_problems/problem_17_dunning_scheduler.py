"""
=============================================================================
INTERVIEW PROBLEM 17: Subscription Dunning & Retry Scheduler
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You're building the recurring-billing retry logic for a subscription
platform. When a scheduled charge fails (card declined, insufficient
funds), the platform must not give up immediately — it retries on a
backoff schedule, and if all retries fail, moves the subscription into
dunning and eventually cancels it.

For this problem you are building a DunningManager class.
Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between DunningManager
instances — avoid them.
You choose the internal data structures; the public interface is what
matters.

Timestamps are ISO-8601 strings without timezone offset, e.g.
"2024-01-01T00:00:00". Use datetime.fromisoformat()/timedelta for
arithmetic.

RETRY SCHEDULE
--------------
RETRY_INTERVALS_DAYS = [1, 3, 7]
After the Nth consecutive failed attempt (N = 1, 2, 3, ...), the next retry
is scheduled RETRY_INTERVALS_DAYS[N - 1] days after that failed attempt's
timestamp. Once N exceeds len(RETRY_INTERVALS_DAYS) (i.e. a 4th consecutive
failure has occurred), the schedule is exhausted — no further retry is
scheduled.

# Example
# dm = DunningManager()
# dm.create_subscription("sub_1", amount=29.99)
# dm.record_attempt("sub_1", "2024-01-01T00:00:00", succeeded=False)
# # -> {"status": "past_due", "consecutive_failures": 1}
# dm.get_next_retry_time("sub_1")
# # -> "2024-01-02T00:00:00"   (1 day later)
# dm.record_attempt("sub_1", "2024-01-02T00:00:00", succeeded=True)
# # -> {"status": "active", "consecutive_failures": 0}
=============================================================================
"""

from typing import Optional


class DunningManager:
    def __init__(self):
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 1 — Attempt tracking  (~12 min)
    # -------------------------------------------------------------------------

    def create_subscription(self, subscription_id: str, amount: float) -> None:
        """
        Register a new subscription with status "active" and 0 consecutive
        failures.
        Raise ValueError if subscription_id already exists.
        """
        raise NotImplementedError

    def get_subscription_status(self, subscription_id: str) -> str:
        """
        Return the current status: "active" | "past_due" | "canceled".
        Raise KeyError if subscription_id does not exist.
        """
        raise NotImplementedError

    def record_attempt(
        self, subscription_id: str, timestamp: str, succeeded: bool
    ) -> dict:
        """
        Record the outcome of a charge attempt for subscription_id and return
        {"status": str, "consecutive_failures": int}.

        - Raise KeyError if subscription_id does not exist.
        - If succeeded is True: consecutive_failures resets to 0 and status
          becomes "active".
        - If succeeded is False: consecutive_failures increments by 1 and
          status becomes "past_due" (whether it was already "active" or
          already "past_due").
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 2 — Backoff scheduling  (~15 min)
    # -------------------------------------------------------------------------

    def get_next_retry_time(self, subscription_id: str) -> Optional[str]:
        """
        Return the ISO-8601 timestamp of the next scheduled retry for
        subscription_id, computed from RETRY_INTERVALS_DAYS (see module
        docstring) and the timestamp + consecutive_failures recorded by
        record_attempt (Part 1).

        - Raise KeyError if subscription_id does not exist.
        - Return None if the subscription is "active" (nothing to retry),
          "canceled", or if consecutive_failures exceeds
          len(RETRY_INTERVALS_DAYS) (schedule exhausted).
        """
        raise NotImplementedError

    def is_retry_due(self, subscription_id: str, current_time: str) -> bool:
        """
        Return True if get_next_retry_time(subscription_id) is not None and
        is at or before current_time. Return False otherwise (including when
        get_next_retry_time returns None).
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 3 — Auto-cancellation  (~18 min)
    # -------------------------------------------------------------------------

    def cancel_subscription(self, subscription_id: str, timestamp: str) -> dict:
        """
        Manually cancel subscription_id and return
        {"status": "canceled", "consecutive_failures": int}.
        Raise KeyError if subscription_id does not exist.
        Idempotent: canceling an already-canceled subscription just returns
        its current state without error.
        """
        raise NotImplementedError

    # Extend record_attempt (Part 1) once more:
    #   - Raise ValueError if record_attempt is called on a subscription whose
    #     status is already "canceled".
    #   - After recording a failed attempt (succeeded=False), check whether
    #     the retry schedule is now exhausted for this subscription (i.e.
    #     get_next_retry_time would return None because consecutive_failures
    #     exceeds len(RETRY_INTERVALS_DAYS)). If so, the subscription's status
    #     becomes "canceled" instead of "past_due", and the returned dict
    #     reflects that.
