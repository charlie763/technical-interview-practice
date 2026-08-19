"""
Reference solution for Problem 17: Subscription Dunning & Retry Scheduler

See python/practice_problems/problem_17_dunning_scheduler.py for the full
problem statement, retry schedule, and method contracts.
"""

from datetime import datetime, timedelta

RETRY_INTERVALS_DAYS = [1, 3, 7]


class DunningManager:
    def __init__(self):
        self._subscriptions = {}

    def _get(self, subscription_id):
        if subscription_id not in self._subscriptions:
            raise KeyError(subscription_id)
        return self._subscriptions[subscription_id]

    # -------------------------------------------------------------------------
    # PART 1 — Attempt tracking
    # -------------------------------------------------------------------------

    def create_subscription(self, subscription_id, amount):
        if subscription_id in self._subscriptions:
            raise ValueError(f"subscription already exists: {subscription_id}")
        self._subscriptions[subscription_id] = {
            "amount": amount,
            "status": "active",
            "consecutive_failures": 0,
            "last_attempt_ts": None,
        }

    def get_subscription_status(self, subscription_id):
        return self._get(subscription_id)["status"]

    def record_attempt(self, subscription_id, timestamp, succeeded):
        sub = self._get(subscription_id)
        if sub["status"] == "canceled":
            raise ValueError(f"subscription is canceled: {subscription_id}")

        sub["last_attempt_ts"] = timestamp
        if succeeded:
            sub["consecutive_failures"] = 0
            sub["status"] = "active"
        else:
            sub["consecutive_failures"] += 1
            sub["status"] = "past_due"
            if self.get_next_retry_time(subscription_id) is None:
                sub["status"] = "canceled"

        return {
            "status": sub["status"],
            "consecutive_failures": sub["consecutive_failures"],
        }

    # -------------------------------------------------------------------------
    # PART 2 — Backoff scheduling
    # -------------------------------------------------------------------------

    def get_next_retry_time(self, subscription_id):
        sub = self._get(subscription_id)
        if sub["status"] in ("active", "canceled"):
            return None

        failures = sub["consecutive_failures"]
        if failures < 1 or failures > len(RETRY_INTERVALS_DAYS):
            return None

        interval_days = RETRY_INTERVALS_DAYS[failures - 1]
        last_ts = datetime.fromisoformat(sub["last_attempt_ts"])
        return (last_ts + timedelta(days=interval_days)).isoformat()

    def is_retry_due(self, subscription_id, current_time):
        next_retry = self.get_next_retry_time(subscription_id)
        if next_retry is None:
            return False
        return datetime.fromisoformat(next_retry) <= datetime.fromisoformat(current_time)

    # -------------------------------------------------------------------------
    # PART 3 — Auto-cancellation
    # -------------------------------------------------------------------------

    def cancel_subscription(self, subscription_id, timestamp):
        sub = self._get(subscription_id)
        sub["status"] = "canceled"
        return {
            "status": "canceled",
            "consecutive_failures": sub["consecutive_failures"],
        }
