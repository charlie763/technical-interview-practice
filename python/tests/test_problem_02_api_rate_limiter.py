"""
Tests for Problem 2: Tiered API Rate Limiter

Run from the python/ directory:
    pytest tests/test_problem_02_api_rate_limiter.py -v
"""

import pytest
from practice_problems.problem_02_api_rate_limiter import (
    make_gateway,
    create_key,
    revoke_key,
    update_plan,
    _count_in_window,
    is_allowed,
    record_request,
    handle_request,
    get_usage,
)

BASE_TIME = 1_700_000_000.0  # arbitrary fixed "now" for deterministic tests


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def gw():
    """Gateway with a known set of plans."""
    plans = {
        "free":  {"rpm": 3,   "rpd": 10},
        "pro":   {"rpm": 100, "rpd": 5_000},
        "unlimited": {"rpm": None, "rpd": None},
    }
    return make_gateway(plans)


@pytest.fixture
def gw_with_key(gw):
    create_key(gw, "key_abc", "alice", "pro")
    return gw


# ---------------------------------------------------------------------------
# PART 1 — Key management
# ---------------------------------------------------------------------------

class TestCreateKey:
    def test_creates_key(self, gw):
        k = create_key(gw, "k1", "alice", "free")
        assert gw["keys"]["k1"] is k
        assert k["owner"] == "alice"
        assert k["plan"] == "free"
        assert k["enabled"] is True
        assert k["request_log"] == []

    def test_duplicate_key_raises(self, gw):
        create_key(gw, "k1", "alice", "free")
        with pytest.raises(ValueError):
            create_key(gw, "k1", "bob", "pro")

    def test_invalid_plan_raises(self, gw):
        with pytest.raises(ValueError):
            create_key(gw, "k1", "alice", "enterprise")


class TestRevokeKey:
    def test_disables_key(self, gw_with_key):
        revoke_key(gw_with_key, "key_abc")
        assert gw_with_key["keys"]["key_abc"]["enabled"] is False

    def test_missing_key_raises(self, gw):
        with pytest.raises(KeyError):
            revoke_key(gw, "ghost")


class TestUpdatePlan:
    def test_changes_plan(self, gw_with_key):
        update_plan(gw_with_key, "key_abc", "free")
        assert gw_with_key["keys"]["key_abc"]["plan"] == "free"

    def test_preserves_request_log(self, gw_with_key):
        gw_with_key["keys"]["key_abc"]["request_log"] = [BASE_TIME - 5]
        update_plan(gw_with_key, "key_abc", "free")
        assert gw_with_key["keys"]["key_abc"]["request_log"] == [BASE_TIME - 5]

    def test_invalid_plan_raises(self, gw_with_key):
        with pytest.raises(ValueError):
            update_plan(gw_with_key, "key_abc", "nonexistent")

    def test_missing_key_raises(self, gw):
        with pytest.raises(KeyError):
            update_plan(gw, "ghost", "free")


# ---------------------------------------------------------------------------
# PART 2 — _count_in_window
# ---------------------------------------------------------------------------

class TestCountInWindow:
    def test_empty_log(self):
        assert _count_in_window([], BASE_TIME, 60) == 0

    def test_all_within_window(self):
        log = [BASE_TIME - 30, BASE_TIME - 10, BASE_TIME]
        assert _count_in_window(log, BASE_TIME, 60) == 3

    def test_some_outside_window(self):
        log = [BASE_TIME - 120, BASE_TIME - 61, BASE_TIME - 30, BASE_TIME]
        assert _count_in_window(log, BASE_TIME, 60) == 2

    def test_exactly_on_window_edge_excluded(self):
        # window is (now - window_seconds, now] — left side is exclusive
        log = [BASE_TIME - 60]
        assert _count_in_window(log, BASE_TIME, 60) == 0

    def test_one_second_inside(self):
        log = [BASE_TIME - 59.999]
        assert _count_in_window(log, BASE_TIME, 60) == 1


# ---------------------------------------------------------------------------
# PART 2 — is_allowed
# ---------------------------------------------------------------------------

class TestIsAllowed:
    def test_allowed_when_under_limits(self, gw_with_key):
        assert is_allowed(gw_with_key, "key_abc", BASE_TIME) is True

    def test_denied_when_key_not_found(self, gw):
        assert is_allowed(gw, "ghost", BASE_TIME) is False

    def test_denied_when_key_disabled(self, gw_with_key):
        revoke_key(gw_with_key, "key_abc")
        assert is_allowed(gw_with_key, "key_abc", BASE_TIME) is False

    def test_denied_when_rpm_exceeded(self, gw):
        create_key(gw, "k1", "alice", "free")  # rpm=3
        # Seed 3 requests in the last minute
        gw["keys"]["k1"]["request_log"] = [BASE_TIME - 30, BASE_TIME - 20, BASE_TIME - 10]
        assert is_allowed(gw, "k1", BASE_TIME) is False

    def test_allowed_when_rpm_window_has_rolled_off(self, gw):
        create_key(gw, "k1", "alice", "free")  # rpm=3
        # 3 requests but all > 60s ago — they're outside the window
        gw["keys"]["k1"]["request_log"] = [
            BASE_TIME - 90, BASE_TIME - 80, BASE_TIME - 70
        ]
        assert is_allowed(gw, "k1", BASE_TIME) is True

    def test_denied_when_rpd_exceeded(self, gw):
        create_key(gw, "k1", "alice", "free")  # rpd=10
        gw["keys"]["k1"]["request_log"] = [BASE_TIME - i * 100 for i in range(10)]
        assert is_allowed(gw, "k1", BASE_TIME) is False

    def test_unlimited_plan_always_allowed(self, gw):
        create_key(gw, "k1", "alice", "unlimited")
        # Even with a huge log, unlimited plan is never blocked
        gw["keys"]["k1"]["request_log"] = [BASE_TIME - i for i in range(1000)]
        assert is_allowed(gw, "k1", BASE_TIME) is True


# ---------------------------------------------------------------------------
# PART 3 — record_request
# ---------------------------------------------------------------------------

class TestRecordRequest:
    def test_appends_timestamp(self, gw_with_key):
        record_request(gw_with_key, "key_abc", BASE_TIME)
        assert BASE_TIME in gw_with_key["keys"]["key_abc"]["request_log"]

    def test_prunes_old_entries(self, gw_with_key):
        old = BASE_TIME - 90_001   # older than 25h
        gw_with_key["keys"]["key_abc"]["request_log"] = [old]
        record_request(gw_with_key, "key_abc", BASE_TIME)
        assert old not in gw_with_key["keys"]["key_abc"]["request_log"]

    def test_keeps_recent_entries(self, gw_with_key):
        recent = BASE_TIME - 3600
        gw_with_key["keys"]["key_abc"]["request_log"] = [recent]
        record_request(gw_with_key, "key_abc", BASE_TIME)
        assert recent in gw_with_key["keys"]["key_abc"]["request_log"]

    def test_missing_key_raises(self, gw):
        with pytest.raises(KeyError):
            record_request(gw, "ghost", BASE_TIME)


# ---------------------------------------------------------------------------
# PART 4 — handle_request
# ---------------------------------------------------------------------------

class TestHandleRequest:
    def test_success_records_request(self, gw_with_key):
        result = handle_request(gw_with_key, "key_abc", BASE_TIME)
        assert result["allowed"] is True
        assert BASE_TIME in gw_with_key["keys"]["key_abc"]["request_log"]

    def test_key_not_found(self, gw):
        result = handle_request(gw, "ghost", BASE_TIME)
        assert result["allowed"] is False
        assert result["reason"] == "key_not_found"

    def test_key_disabled(self, gw_with_key):
        revoke_key(gw_with_key, "key_abc")
        result = handle_request(gw_with_key, "key_abc", BASE_TIME)
        assert result["allowed"] is False
        assert result["reason"] == "key_disabled"

    def test_rpm_exceeded(self, gw):
        create_key(gw, "k1", "alice", "free")  # rpm=3
        gw["keys"]["k1"]["request_log"] = [BASE_TIME - 10, BASE_TIME - 5, BASE_TIME - 1]
        result = handle_request(gw, "k1", BASE_TIME)
        assert result["allowed"] is False
        assert result["reason"] == "rpm_exceeded"

    def test_rpd_exceeded_not_recorded(self, gw):
        create_key(gw, "k1", "alice", "free")  # rpd=10
        gw["keys"]["k1"]["request_log"] = [BASE_TIME - i * 100 for i in range(10)]
        log_before = list(gw["keys"]["k1"]["request_log"])
        result = handle_request(gw, "k1", BASE_TIME)
        assert result["allowed"] is False
        assert result["reason"] == "rpd_exceeded"
        # log must not be modified on failure
        assert gw["keys"]["k1"]["request_log"] == log_before

    def test_rpd_checked_after_rpm(self, gw):
        """rpd_exceeded should only appear when rpm is within limits."""
        create_key(gw, "k1", "alice", "free")  # rpm=3, rpd=10
        # rpm is OK (0 in last minute), but rpd is blown
        gw["keys"]["k1"]["request_log"] = [BASE_TIME - 3600 * i for i in range(1, 11)]
        result = handle_request(gw, "k1", BASE_TIME)
        assert result["reason"] == "rpd_exceeded"


# ---------------------------------------------------------------------------
# PART 4 — get_usage
# ---------------------------------------------------------------------------

class TestGetUsage:
    def test_returns_correct_counts(self, gw_with_key):
        gw_with_key["keys"]["key_abc"]["request_log"] = [
            BASE_TIME - 30,   # within both minute and day windows
            BASE_TIME - 3600, # within day window only
        ]
        stats = get_usage(gw_with_key, "key_abc", BASE_TIME)
        assert stats["rpm_used"] == 1
        assert stats["rpd_used"] == 2
        assert stats["plan"] == "pro"
        assert stats["rpm_limit"] == 100
        assert stats["rpd_limit"] == 5_000

    def test_unlimited_plan_shows_none_limits(self, gw):
        create_key(gw, "k1", "alice", "unlimited")
        stats = get_usage(gw, "k1", BASE_TIME)
        assert stats["rpm_limit"] is None
        assert stats["rpd_limit"] is None

    def test_missing_key_raises(self, gw):
        with pytest.raises(KeyError):
            get_usage(gw, "ghost", BASE_TIME)
