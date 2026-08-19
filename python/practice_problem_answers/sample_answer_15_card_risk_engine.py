"""
Reference solution for Problem 15: Card Transaction Risk Engine

See python/practice_problems/problem_15_card_risk_engine.py for the full
problem statement, decision model, and rule definitions.
"""

from datetime import datetime, timedelta

BLOCKED_MERCHANT_CATEGORIES = {"gambling", "cash_advance", "crypto_exchange"}
AMOUNT_LIMIT = 5000
HIGH_AMOUNT_MIN = 1000

VELOCITY_COUNT_WINDOW_SECS = 60
VELOCITY_COUNT_MAX = 3
VELOCITY_AMOUNT_WINDOW_SECS = 600
VELOCITY_AMOUNT_MAX = 10000

BLOCK_WINDOW_SECS = 24 * 60 * 60
BLOCK_DECLINE_THRESHOLD = 3

_SEVERITY = {"approve": 0, "review": 1, "decline": 2}


class RiskEngine:
    def __init__(self):
        self._history = {}  # card_id -> list of {"amount", "ts", "decision"}

    # -------------------------------------------------------------------------
    # PART 1 + 2 — Static rules + velocity checks
    # -------------------------------------------------------------------------

    def evaluate_transaction(self, card_id, amount, merchant_category, timestamp):
        ts = datetime.fromisoformat(timestamp)
        history = self._history.setdefault(card_id, [])

        if self._is_blocked(history, ts):
            history.append({"amount": amount, "ts": ts, "decision": "decline"})
            return {"decision": "decline", "reasons": ["card_blocked"]}

        reasons = []
        decision = "approve"

        def flag(reason, level):
            nonlocal decision
            reasons.append(reason)
            if _SEVERITY[level] > _SEVERITY[decision]:
                decision = level

        if amount > AMOUNT_LIMIT:
            flag("amount_limit_exceeded", "decline")
        if merchant_category in BLOCKED_MERCHANT_CATEGORIES:
            flag("blocked_merchant_category", "decline")
        if HIGH_AMOUNT_MIN <= amount <= AMOUNT_LIMIT:
            flag("high_amount", "review")

        count_window = self._window(history, ts, VELOCITY_COUNT_WINDOW_SECS)
        if len(count_window) + 1 > VELOCITY_COUNT_MAX:
            flag("velocity_count_exceeded", "decline")

        amount_window = self._window(history, ts, VELOCITY_AMOUNT_WINDOW_SECS)
        total = sum(h["amount"] for h in amount_window) + amount
        if total > VELOCITY_AMOUNT_MAX:
            flag("velocity_amount_exceeded", "review")

        history.append({"amount": amount, "ts": ts, "decision": decision})
        return {"decision": decision, "reasons": reasons}

    @staticmethod
    def _window(history, ts, window_secs):
        earliest = ts - timedelta(seconds=window_secs)
        return [h for h in history if earliest <= h["ts"] <= ts]

    @classmethod
    def _is_blocked(cls, history, ts):
        declines = [
            h
            for h in cls._window(history, ts, BLOCK_WINDOW_SECS)
            if h["decision"] == "decline"
        ]
        return len(declines) >= BLOCK_DECLINE_THRESHOLD

    # -------------------------------------------------------------------------
    # PART 3 — Auto-blocking
    # -------------------------------------------------------------------------

    def is_card_blocked(self, card_id):
        history = self._history.get(card_id)
        if not history:
            return False
        latest_ts = history[-1]["ts"]
        return self._is_blocked(history, latest_ts)
