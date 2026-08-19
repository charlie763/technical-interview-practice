"""Tests for Problem 17: Subscription Dunning & Retry Scheduler

Run from the python/ directory:
    pytest tests/test_problem_17_dunning_scheduler.py -v
"""

import pytest

from practice_problems.problem_17_dunning_scheduler import DunningManager

# ---------------------------------------------------------------------------
# Shared timestamps
# D0        = base
# D0_PLUS_1 = D0 + 1 day   (1st retry interval)
# D0_PLUS_4 = D0_PLUS_1 + 3 days  (2nd retry interval)
# D0_PLUS_11 = D0_PLUS_4 + 7 days (3rd retry interval)
# ---------------------------------------------------------------------------
D0 = "2024-01-01T00:00:00"
D0_PLUS_1 = "2024-01-02T00:00:00"
D0_PLUS_4 = "2024-01-05T00:00:00"
D0_PLUS_11 = "2024-01-12T00:00:00"


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def fresh_dm():
    """Empty DunningManager."""
    return DunningManager()


@pytest.fixture
def dm():
    """
    Pre-seeded manager:
      sub_active:   active, no failed attempts.
      sub_past_due: 1 failed attempt at D0 (past_due, consecutive_failures=1).
    """
    m = DunningManager()
    m.create_subscription("sub_active", amount=19.99)
    m.create_subscription("sub_past_due", amount=29.99)
    m.record_attempt("sub_past_due", D0, succeeded=False)
    return m


# ---------------------------------------------------------------------------
# PART 1 — Attempt tracking
# ---------------------------------------------------------------------------

class TestAttemptTracking:
    def test_create_subscription_starts_active(self, fresh_dm):
        fresh_dm.create_subscription("sub_create_test", amount=10.0)
        assert fresh_dm.get_subscription_status("sub_create_test") == "active"

    def test_duplicate_subscription_raises(self, fresh_dm):
        fresh_dm.create_subscription("sub_dup_test", amount=10.0)
        with pytest.raises(ValueError):
            fresh_dm.create_subscription("sub_dup_test", amount=10.0)

    def test_get_status_unknown_raises(self, fresh_dm):
        with pytest.raises(KeyError):
            fresh_dm.get_subscription_status("sub_ghost")

    def test_record_attempt_unknown_subscription_raises(self, fresh_dm):
        with pytest.raises(KeyError):
            fresh_dm.record_attempt("sub_ghost", D0, succeeded=True)

    def test_failed_attempt_moves_to_past_due(self, fresh_dm):
        fresh_dm.create_subscription("sub_fail_test", amount=10.0)
        result = fresh_dm.record_attempt("sub_fail_test", D0, succeeded=False)
        assert result == {"status": "past_due", "consecutive_failures": 1}
        assert fresh_dm.get_subscription_status("sub_fail_test") == "past_due"

    def test_successful_attempt_keeps_active(self, fresh_dm):
        fresh_dm.create_subscription("sub_success_test", amount=10.0)
        result = fresh_dm.record_attempt("sub_success_test", D0, succeeded=True)
        assert result == {"status": "active", "consecutive_failures": 0}

    def test_success_resets_failure_streak(self, fresh_dm):
        fresh_dm.create_subscription("sub_reset_test", amount=10.0)
        fresh_dm.record_attempt("sub_reset_test", D0, succeeded=False)
        fresh_dm.record_attempt("sub_reset_test", D0_PLUS_1, succeeded=False)
        result = fresh_dm.record_attempt("sub_reset_test", D0_PLUS_4, succeeded=True)
        assert result == {"status": "active", "consecutive_failures": 0}

    def test_consecutive_failures_increment(self, fresh_dm):
        fresh_dm.create_subscription("sub_incr_test", amount=10.0)
        fresh_dm.record_attempt("sub_incr_test", D0, succeeded=False)
        result = fresh_dm.record_attempt("sub_incr_test", D0_PLUS_1, succeeded=False)
        assert result == {"status": "past_due", "consecutive_failures": 2}

    def test_preseeded_active_subscription_status(self, dm):
        assert dm.get_subscription_status("sub_active") == "active"

    def test_preseeded_past_due_subscription_status(self, dm):
        assert dm.get_subscription_status("sub_past_due") == "past_due"


# ---------------------------------------------------------------------------
# PART 2 — Backoff scheduling
# ---------------------------------------------------------------------------

class TestBackoffScheduling:
    def test_next_retry_time_none_when_active(self, fresh_dm):
        fresh_dm.create_subscription("sub_retry_active_test", amount=10.0)
        assert fresh_dm.get_next_retry_time("sub_retry_active_test") is None

    def test_next_retry_time_after_first_failure(self, fresh_dm):
        fresh_dm.create_subscription("sub_retry_1_test", amount=10.0)
        fresh_dm.record_attempt("sub_retry_1_test", D0, succeeded=False)
        assert fresh_dm.get_next_retry_time("sub_retry_1_test") == D0_PLUS_1

    def test_next_retry_time_after_second_failure(self, fresh_dm):
        fresh_dm.create_subscription("sub_retry_2_test", amount=10.0)
        fresh_dm.record_attempt("sub_retry_2_test", D0, succeeded=False)
        fresh_dm.record_attempt("sub_retry_2_test", D0_PLUS_1, succeeded=False)
        assert fresh_dm.get_next_retry_time("sub_retry_2_test") == D0_PLUS_4

    def test_next_retry_time_after_third_failure(self, fresh_dm):
        fresh_dm.create_subscription("sub_retry_3_test", amount=10.0)
        fresh_dm.record_attempt("sub_retry_3_test", D0, succeeded=False)
        fresh_dm.record_attempt("sub_retry_3_test", D0_PLUS_1, succeeded=False)
        fresh_dm.record_attempt("sub_retry_3_test", D0_PLUS_4, succeeded=False)
        assert fresh_dm.get_next_retry_time("sub_retry_3_test") == D0_PLUS_11

    def test_unknown_subscription_raises_for_next_retry(self, fresh_dm):
        with pytest.raises(KeyError):
            fresh_dm.get_next_retry_time("sub_ghost_retry")

    def test_is_retry_due_true_when_time_reached(self, fresh_dm):
        fresh_dm.create_subscription("sub_due_test", amount=10.0)
        fresh_dm.record_attempt("sub_due_test", D0, succeeded=False)
        assert fresh_dm.is_retry_due("sub_due_test", D0_PLUS_1) is True
        assert fresh_dm.is_retry_due("sub_due_test", D0_PLUS_4) is True

    def test_is_retry_due_false_before_scheduled_time(self, fresh_dm):
        fresh_dm.create_subscription("sub_not_due_test", amount=10.0)
        fresh_dm.record_attempt("sub_not_due_test", D0, succeeded=False)
        assert fresh_dm.is_retry_due("sub_not_due_test", "2024-01-01T12:00:00") is False

    def test_is_retry_due_false_when_active(self, fresh_dm):
        fresh_dm.create_subscription("sub_active_due_test", amount=10.0)
        assert fresh_dm.is_retry_due("sub_active_due_test", D0) is False

    def test_preseeded_past_due_has_scheduled_retry(self, dm):
        assert dm.get_next_retry_time("sub_past_due") == D0_PLUS_1


# ---------------------------------------------------------------------------
# PART 3 — Auto-cancellation
# ---------------------------------------------------------------------------

class TestAutoCancellation:
    def test_manual_cancel_sets_status(self, fresh_dm):
        fresh_dm.create_subscription("sub_cancel_test", amount=10.0)
        result = fresh_dm.cancel_subscription("sub_cancel_test", D0)
        assert result == {"status": "canceled", "consecutive_failures": 0}
        assert fresh_dm.get_subscription_status("sub_cancel_test") == "canceled"

    def test_manual_cancel_is_idempotent(self, fresh_dm):
        fresh_dm.create_subscription("sub_cancel_idem_test", amount=10.0)
        fresh_dm.cancel_subscription("sub_cancel_idem_test", D0)
        result = fresh_dm.cancel_subscription("sub_cancel_idem_test", D0_PLUS_1)
        assert result["status"] == "canceled"

    def test_cancel_unknown_subscription_raises(self, fresh_dm):
        with pytest.raises(KeyError):
            fresh_dm.cancel_subscription("sub_ghost_cancel", D0)

    def test_record_attempt_on_canceled_subscription_raises(self, fresh_dm):
        fresh_dm.create_subscription("sub_canceled_attempt_test", amount=10.0)
        fresh_dm.cancel_subscription("sub_canceled_attempt_test", D0)
        with pytest.raises(ValueError):
            fresh_dm.record_attempt("sub_canceled_attempt_test", D0_PLUS_1, succeeded=True)

    def test_auto_cancels_after_retry_schedule_exhausted(self, fresh_dm):
        sub = "sub_auto_cancel_test"
        fresh_dm.create_subscription(sub, amount=10.0)
        fresh_dm.record_attempt(sub, D0, succeeded=False)          # 1
        fresh_dm.record_attempt(sub, D0_PLUS_1, succeeded=False)   # 2
        fresh_dm.record_attempt(sub, D0_PLUS_4, succeeded=False)   # 3
        result = fresh_dm.record_attempt(sub, D0_PLUS_11, succeeded=False)  # 4 — exhausted
        assert result == {"status": "canceled", "consecutive_failures": 4}
        assert fresh_dm.get_subscription_status(sub) == "canceled"
        assert fresh_dm.get_next_retry_time(sub) is None

    def test_not_canceled_before_schedule_exhausted(self, fresh_dm):
        sub = "sub_not_yet_canceled_test"
        fresh_dm.create_subscription(sub, amount=10.0)
        fresh_dm.record_attempt(sub, D0, succeeded=False)
        fresh_dm.record_attempt(sub, D0_PLUS_1, succeeded=False)
        result = fresh_dm.record_attempt(sub, D0_PLUS_4, succeeded=False)
        assert result["status"] == "past_due"
