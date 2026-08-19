"""Tests for Problem 15: Card Transaction Risk Engine

Run from the python/ directory:
    pytest tests/test_problem_15_card_risk_engine.py -v
"""

import pytest

from practice_problems.problem_15_card_risk_engine import RiskEngine

# ---------------------------------------------------------------------------
# Shared timestamps  (all naive ISO-8601, lexicographically sortable)
# ---------------------------------------------------------------------------
T0 = "2024-06-01T10:00:00"
T0_10S = "2024-06-01T10:00:10"
T0_20S = "2024-06-01T10:00:20"
T0_30S = "2024-06-01T10:00:30"
T0_1H = "2024-06-01T11:00:00"
T0_2H = "2024-06-01T12:00:00"
T0_2H30M = "2024-06-01T12:30:00"
T0_PLUS_27H = "2024-06-02T13:00:00"  # 25h after T0_2H — outside the 24h block window

# Spaced 70s apart: wide enough that no 60s window ever contains more than
# one of these, so velocity_count_exceeded never fires.
TA0 = "2024-06-01T09:00:00"
TA1 = "2024-06-01T09:01:10"
TA2 = "2024-06-01T09:02:20"
TA3 = "2024-06-01T09:03:30"
TA4 = "2024-06-01T09:04:40"


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def fresh_engine():
    """Empty RiskEngine."""
    return RiskEngine()


@pytest.fixture
def engine():
    """Pre-seeded engine: card_seed has one approved transaction at T0."""
    e = RiskEngine()
    e.evaluate_transaction("card_seed", 100.0, "grocery", T0)
    return e


# ---------------------------------------------------------------------------
# PART 1 — Static rule evaluation
# ---------------------------------------------------------------------------

class TestStaticRules:
    def test_approves_normal_transaction(self, fresh_engine):
        result = fresh_engine.evaluate_transaction("card_static_approve", 200.0, "grocery", T0)
        assert result == {"decision": "approve", "reasons": []}

    def test_amount_over_limit_declines(self, fresh_engine):
        result = fresh_engine.evaluate_transaction("card_amount_limit", 6000.0, "electronics", T0)
        assert result == {"decision": "decline", "reasons": ["amount_limit_exceeded"]}

    def test_blocked_merchant_category_declines(self, fresh_engine):
        result = fresh_engine.evaluate_transaction("card_blocked_merchant", 50.0, "gambling", T0)
        assert result == {"decision": "decline", "reasons": ["blocked_merchant_category"]}

    def test_high_amount_reviews(self, fresh_engine):
        result = fresh_engine.evaluate_transaction("card_high_amount", 1500.0, "travel", T0)
        assert result == {"decision": "review", "reasons": ["high_amount"]}

    def test_amount_at_high_amount_lower_bound(self, fresh_engine):
        result = fresh_engine.evaluate_transaction("card_boundary_low", 1000.0, "travel", T0)
        assert result == {"decision": "review", "reasons": ["high_amount"]}

    def test_amount_at_limit_is_high_amount_not_decline(self, fresh_engine):
        result = fresh_engine.evaluate_transaction("card_boundary_high", 5000.0, "travel", T0)
        assert result == {"decision": "review", "reasons": ["high_amount"]}

    def test_amount_below_high_amount_threshold_approves(self, fresh_engine):
        result = fresh_engine.evaluate_transaction("card_below_threshold", 999.0, "travel", T0)
        assert result == {"decision": "approve", "reasons": []}

    def test_multiple_static_rules_combine(self, fresh_engine):
        result = fresh_engine.evaluate_transaction("card_multi_rule", 6000.0, "gambling", T0)
        assert result["decision"] == "decline"
        assert result["reasons"] == ["amount_limit_exceeded", "blocked_merchant_category"]

    def test_new_card_unaffected_by_other_cards_history(self, engine):
        result = engine.evaluate_transaction("card_unrelated", 200.0, "grocery", T0)
        assert result == {"decision": "approve", "reasons": []}


# ---------------------------------------------------------------------------
# PART 2 — Velocity checks
# ---------------------------------------------------------------------------

class TestVelocityChecks:
    def test_velocity_count_exceeded_after_four_rapid_transactions(self, fresh_engine):
        card = "card_vel_count"
        r1 = fresh_engine.evaluate_transaction(card, 50.0, "grocery", T0)
        r2 = fresh_engine.evaluate_transaction(card, 50.0, "grocery", T0_10S)
        r3 = fresh_engine.evaluate_transaction(card, 50.0, "grocery", T0_20S)
        r4 = fresh_engine.evaluate_transaction(card, 50.0, "grocery", T0_30S)
        assert r1["decision"] == "approve"
        assert r2["decision"] == "approve"
        assert r3["decision"] == "approve"
        assert r4["decision"] == "decline"
        assert "velocity_count_exceeded" in r4["reasons"]

    def test_velocity_count_check_is_per_card(self, fresh_engine):
        for ts in (T0, T0_10S, T0_20S, T0_30S):
            fresh_engine.evaluate_transaction("card_vel_isolated_a", 50.0, "grocery", ts)
        result = fresh_engine.evaluate_transaction("card_vel_isolated_b", 50.0, "grocery", T0_30S)
        assert result["decision"] == "approve"

    def test_velocity_amount_exceeded_after_cumulative_spend(self, fresh_engine):
        card = "card_vel_amount"
        results = [
            fresh_engine.evaluate_transaction(card, 2200.0, "travel", ts)
            for ts in (TA0, TA1, TA2, TA3, TA4)
        ]
        for r in results[:4]:
            assert r["reasons"] == ["high_amount"]
        assert "velocity_amount_exceeded" in results[4]["reasons"]
        assert results[4]["decision"] == "review"

    def test_velocity_amount_check_ignores_wide_gaps(self, fresh_engine):
        # Same per-transaction amount as above, but 20 min apart — far outside
        # the 600s trailing window, so the total never accumulates.
        card = "card_vel_amount_wide"
        ts_list = [
            "2024-06-01T09:00:00",
            "2024-06-01T09:20:00",
            "2024-06-01T09:40:00",
            "2024-06-01T10:00:00",
            "2024-06-01T10:20:00",
        ]
        results = [
            fresh_engine.evaluate_transaction(card, 2200.0, "travel", ts) for ts in ts_list
        ]
        assert all("velocity_amount_exceeded" not in r["reasons"] for r in results)


# ---------------------------------------------------------------------------
# PART 3 — Auto-blocking
# ---------------------------------------------------------------------------

class TestAutoBlocking:
    def test_card_not_blocked_initially(self, fresh_engine):
        assert fresh_engine.is_card_blocked("card_never_seen") is False

    def test_not_blocked_before_threshold(self, fresh_engine):
        card = "card_block_partial"
        fresh_engine.evaluate_transaction(card, 6000.0, "electronics", T0)
        fresh_engine.evaluate_transaction(card, 6000.0, "electronics", T0_1H)
        assert fresh_engine.is_card_blocked(card) is False

    def test_card_blocked_after_three_declines(self, fresh_engine):
        card = "card_block_full"
        fresh_engine.evaluate_transaction(card, 6000.0, "electronics", T0)
        fresh_engine.evaluate_transaction(card, 6000.0, "electronics", T0_1H)
        third = fresh_engine.evaluate_transaction(card, 6000.0, "electronics", T0_2H)
        assert third["reasons"] == ["amount_limit_exceeded"]
        assert fresh_engine.is_card_blocked(card) is True

    def test_blocked_card_short_circuits_future_rules(self, fresh_engine):
        card = "card_block_shortcircuit"
        fresh_engine.evaluate_transaction(card, 6000.0, "electronics", T0)
        fresh_engine.evaluate_transaction(card, 6000.0, "electronics", T0_1H)
        fresh_engine.evaluate_transaction(card, 6000.0, "electronics", T0_2H)
        result = fresh_engine.evaluate_transaction(card, 50.0, "grocery", T0_2H30M)
        assert result == {"decision": "decline", "reasons": ["card_blocked"]}

    def test_block_expires_once_declines_age_out(self, fresh_engine):
        card = "card_block_expires"
        fresh_engine.evaluate_transaction(card, 6000.0, "electronics", T0)
        fresh_engine.evaluate_transaction(card, 6000.0, "electronics", T0_1H)
        fresh_engine.evaluate_transaction(card, 6000.0, "electronics", T0_2H)
        assert fresh_engine.is_card_blocked(card) is True

        result = fresh_engine.evaluate_transaction(card, 50.0, "grocery", T0_PLUS_27H)
        assert result == {"decision": "approve", "reasons": []}
        assert fresh_engine.is_card_blocked(card) is False
