"""Tests for Problem 17: Claims Processing Pipeline

Run from the python/ directory:
    pytest tests/test_problem_17_claims_pipeline.py -v
"""

import pytest
from practice_problems.problem_17_claims_pipeline import ClaimsPipeline


# ---------------------------------------------------------------------------
# Shared timestamps
# ---------------------------------------------------------------------------
D0 = "2025-01-15"                # incident date
T0 = "2025-02-01T09:00:00"       # filed
T1 = "2025-02-03T10:00:00"       # investigating
T2 = "2025-02-10T14:00:00"       # evaluation
T3 = "2025-02-20T11:00:00"       # settled / denied
T4 = "2025-03-01T09:00:00"       # closed


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def fresh_pipeline():
    """Empty ClaimsPipeline."""
    return ClaimsPipeline()


@pytest.fixture
def pipeline():
    """
    Pre-seeded pipeline:
      clm-001  pol-101  epl   claimed=75_000  reserve=50_000  → settled (approved=60_000)
      clm-002  pol-101  epl   claimed=30_000  reserve=30_000  → investigating
      clm-003  pol-102  do    claimed=200_000 reserve=150_000 → evaluation
      clm-004  pol-103  epl   claimed=10_000  reserve=10_000  → denied
    """
    p = ClaimsPipeline()

    # clm-001: filed → investigating → evaluation → settled → closed
    p.file_claim(
        "clm-001", policy_id="pol-101", coverage_type="epl",
        incident_date=D0, filed_at=T0,
        claimed_amount=75_000, reserve_amount=50_000, actor="adj-1",
    )
    p.advance_status("clm-001", "investigating", at=T1, actor="adj-1")
    p.advance_status("clm-001", "evaluation", at=T2, actor="adj-1")
    p.settle_claim("clm-001", approved_amount=60_000, settled_at=T3, actor="adj-1")

    # clm-002: filed → investigating (stays there)
    p.file_claim(
        "clm-002", policy_id="pol-101", coverage_type="epl",
        incident_date=D0, filed_at=T0,
        claimed_amount=30_000, reserve_amount=30_000, actor="adj-2",
    )
    p.advance_status("clm-002", "investigating", at=T1, actor="adj-2")

    # clm-003: filed → investigating → evaluation (stays there)
    p.file_claim(
        "clm-003", policy_id="pol-102", coverage_type="do",
        incident_date=D0, filed_at=T0,
        claimed_amount=200_000, reserve_amount=150_000, actor="adj-1",
    )
    p.advance_status("clm-003", "investigating", at=T1, actor="adj-1")
    p.advance_status("clm-003", "evaluation", at=T2, actor="adj-1")

    # clm-004: filed → investigating → evaluation → denied
    p.file_claim(
        "clm-004", policy_id="pol-103", coverage_type="epl",
        incident_date=D0, filed_at=T0,
        claimed_amount=10_000, reserve_amount=10_000, actor="adj-3",
    )
    p.advance_status("clm-004", "investigating", at=T1, actor="adj-3")
    p.advance_status("clm-004", "evaluation", at=T2, actor="adj-3")
    p.deny_claim("clm-004", reason="Coverage exclusion applies.", denied_at=T3, actor="adj-3")

    return p


# ---------------------------------------------------------------------------
# PART 1 — Claim filing, status transitions, reserve updates
# ---------------------------------------------------------------------------

class TestFileClaim:
    def test_returns_claim_in_filed_state(self, fresh_pipeline):
        c = fresh_pipeline.file_claim(
            "clm-new", policy_id="pol-200", coverage_type="do",
            incident_date=D0, filed_at=T0,
            claimed_amount=50_000, reserve_amount=40_000, actor="adj-1",
        )
        assert c["claim_id"] == "clm-new"
        assert c["status"] == "filed"
        assert c["approved_amount"] is None
        assert c["claimed_amount"] == 50_000
        assert c["reserve_amount"] == 40_000

    def test_initial_event_recorded(self, fresh_pipeline):
        fresh_pipeline.file_claim(
            "clm-evt", policy_id="pol-200", coverage_type="epl",
            incident_date=D0, filed_at=T0,
            claimed_amount=20_000, reserve_amount=15_000, actor="adj-1",
        )
        c = fresh_pipeline.get_claim("clm-evt")
        assert len(c["events"]) == 1
        assert c["events"][0]["action"] == "filed"
        assert c["events"][0]["actor"] == "adj-1"

    def test_duplicate_raises(self, fresh_pipeline):
        fresh_pipeline.file_claim(
            "clm-dup", policy_id="pol-200", coverage_type="do",
            incident_date=D0, filed_at=T0,
            claimed_amount=10_000, reserve_amount=10_000, actor="adj-1",
        )
        with pytest.raises(ValueError):
            fresh_pipeline.file_claim(
                "clm-dup", policy_id="pol-201", coverage_type="do",
                incident_date=D0, filed_at=T1,
                claimed_amount=5_000, reserve_amount=5_000, actor="adj-1",
            )

    def test_zero_claimed_amount_raises(self, fresh_pipeline):
        with pytest.raises(ValueError):
            fresh_pipeline.file_claim(
                "clm-zero", policy_id="pol-200", coverage_type="epl",
                incident_date=D0, filed_at=T0,
                claimed_amount=0, reserve_amount=10_000, actor="adj-1",
            )

    def test_zero_reserve_raises(self, fresh_pipeline):
        with pytest.raises(ValueError):
            fresh_pipeline.file_claim(
                "clm-zero-r", policy_id="pol-200", coverage_type="epl",
                incident_date=D0, filed_at=T0,
                claimed_amount=10_000, reserve_amount=0, actor="adj-1",
            )


class TestGetClaim:
    def test_returns_existing_claim(self, pipeline):
        c = pipeline.get_claim("clm-001")
        assert c["claim_id"] == "clm-001"

    def test_unknown_raises(self, pipeline):
        with pytest.raises(KeyError):
            pipeline.get_claim("no-such")


class TestAdvanceStatus:
    def test_valid_transition_updates_status(self, fresh_pipeline):
        fresh_pipeline.file_claim(
            "clm-adv", policy_id="pol-200", coverage_type="epl",
            incident_date=D0, filed_at=T0,
            claimed_amount=20_000, reserve_amount=15_000, actor="adj-1",
        )
        fresh_pipeline.advance_status("clm-adv", "investigating", at=T1, actor="adj-1")
        assert fresh_pipeline.get_claim("clm-adv")["status"] == "investigating"

    def test_invalid_transition_raises(self, fresh_pipeline):
        fresh_pipeline.file_claim(
            "clm-inv", policy_id="pol-200", coverage_type="epl",
            incident_date=D0, filed_at=T0,
            claimed_amount=20_000, reserve_amount=15_000, actor="adj-1",
        )
        with pytest.raises(ValueError):
            fresh_pipeline.advance_status("clm-inv", "evaluation", at=T1, actor="adj-1")

    def test_settling_via_advance_status_raises(self, pipeline):
        # "settled" must go through settle_claim, not advance_status
        with pytest.raises(ValueError):
            pipeline.advance_status("clm-003", "settled", at=T3, actor="adj-1")

    def test_denying_via_advance_status_raises(self, pipeline):
        with pytest.raises(ValueError):
            pipeline.advance_status("clm-003", "denied", at=T3, actor="adj-1")

    def test_appends_status_change_event(self, fresh_pipeline):
        fresh_pipeline.file_claim(
            "clm-ev2", policy_id="pol-200", coverage_type="epl",
            incident_date=D0, filed_at=T0,
            claimed_amount=20_000, reserve_amount=15_000, actor="adj-1",
        )
        fresh_pipeline.advance_status("clm-ev2", "investigating", at=T1, actor="adj-2")
        events = fresh_pipeline.get_claim("clm-ev2")["events"]
        last = events[-1]
        assert last["action"] == "status_change"
        assert last["payload"]["from_status"] == "filed"
        assert last["payload"]["to_status"] == "investigating"

    def test_unknown_raises(self, pipeline):
        with pytest.raises(KeyError):
            pipeline.advance_status("no-such", "investigating", at=T1, actor="adj-1")


class TestUpdateReserve:
    def test_updates_reserve_amount(self, pipeline):
        pipeline.update_reserve("clm-002", 35_000, at=T2, actor="adj-2")
        assert pipeline.get_claim("clm-002")["reserve_amount"] == 35_000

    def test_appends_reserve_update_event(self, pipeline):
        pipeline.update_reserve("clm-002", 35_000, at=T2, actor="adj-2")
        events = pipeline.get_claim("clm-002")["events"]
        last = events[-1]
        assert last["action"] == "reserve_update"
        assert last["payload"]["old_reserve"] == 30_000
        assert last["payload"]["new_reserve"] == 35_000

    def test_zero_reserve_raises(self, pipeline):
        with pytest.raises(ValueError):
            pipeline.update_reserve("clm-002", 0, at=T2, actor="adj-2")

    def test_closed_claim_raises(self, fresh_pipeline):
        fresh_pipeline.file_claim(
            "clm-cls", policy_id="pol-200", coverage_type="epl",
            incident_date=D0, filed_at=T0,
            claimed_amount=20_000, reserve_amount=15_000, actor="adj-1",
        )
        fresh_pipeline.advance_status("clm-cls", "investigating", at=T1, actor="adj-1")
        fresh_pipeline.advance_status("clm-cls", "evaluation", at=T2, actor="adj-1")
        fresh_pipeline.settle_claim("clm-cls", approved_amount=10_000, settled_at=T3, actor="adj-1")
        fresh_pipeline.advance_status("clm-cls", "closed", at=T4, actor="adj-1")
        with pytest.raises(ValueError):
            fresh_pipeline.update_reserve("clm-cls", 5_000, at=T4, actor="adj-1")

    def test_unknown_raises(self, pipeline):
        with pytest.raises(KeyError):
            pipeline.update_reserve("no-such", 10_000, at=T2, actor="adj-1")


# ---------------------------------------------------------------------------
# PART 2 — Settlement, denial, and query methods
# ---------------------------------------------------------------------------

class TestSettleClaim:
    def test_settles_claim_and_sets_approved_amount(self, pipeline):
        # clm-003 is in evaluation
        pipeline.settle_claim("clm-003", approved_amount=150_000, settled_at=T3, actor="adj-1")
        c = pipeline.get_claim("clm-003")
        assert c["status"] == "settled"
        assert c["approved_amount"] == 150_000

    def test_appends_settled_event(self, pipeline):
        pipeline.settle_claim("clm-003", approved_amount=150_000, settled_at=T3, actor="adj-1")
        events = pipeline.get_claim("clm-003")["events"]
        last = events[-1]
        assert last["action"] == "settled"
        assert last["payload"]["approved_amount"] == 150_000

    def test_wrong_state_raises(self, pipeline):
        # clm-002 is in investigating, not evaluation
        with pytest.raises(ValueError):
            pipeline.settle_claim("clm-002", approved_amount=20_000, settled_at=T3, actor="adj-2")

    def test_approved_exceeds_claimed_raises(self, pipeline):
        with pytest.raises(ValueError):
            pipeline.settle_claim("clm-003", approved_amount=300_000, settled_at=T3, actor="adj-1")

    def test_zero_approved_raises(self, pipeline):
        with pytest.raises(ValueError):
            pipeline.settle_claim("clm-003", approved_amount=0, settled_at=T3, actor="adj-1")

    def test_unknown_raises(self, pipeline):
        with pytest.raises(KeyError):
            pipeline.settle_claim("no-such", approved_amount=10_000, settled_at=T3, actor="adj-1")


class TestDenyClaim:
    def test_denies_claim(self, pipeline):
        # clm-003 is in evaluation
        pipeline.deny_claim("clm-003", reason="Outside coverage period.", denied_at=T3, actor="adj-1")
        assert pipeline.get_claim("clm-003")["status"] == "denied"

    def test_appends_denied_event(self, pipeline):
        pipeline.deny_claim("clm-003", reason="Exclusion.", denied_at=T3, actor="adj-1")
        events = pipeline.get_claim("clm-003")["events"]
        last = events[-1]
        assert last["action"] == "denied"
        assert last["payload"]["reason"] == "Exclusion."

    def test_wrong_state_raises(self, pipeline):
        with pytest.raises(ValueError):
            pipeline.deny_claim("clm-002", reason="Reason.", denied_at=T3, actor="adj-2")

    def test_unknown_raises(self, pipeline):
        with pytest.raises(KeyError):
            pipeline.deny_claim("no-such", reason="Reason.", denied_at=T3, actor="adj-1")


class TestGetClaimsByPolicy:
    def test_returns_claims_for_policy(self, pipeline):
        claims = pipeline.get_claims_by_policy("pol-101")
        ids = [c["claim_id"] for c in claims]
        assert "clm-001" in ids
        assert "clm-002" in ids
        assert "clm-003" not in ids

    def test_sorted_by_filed_at(self, fresh_pipeline):
        # Both filed at T0; filing order should match
        fresh_pipeline.file_claim(
            "clm-b", policy_id="pol-sort", coverage_type="epl",
            incident_date=D0, filed_at=T1,
            claimed_amount=10_000, reserve_amount=8_000, actor="adj-1",
        )
        fresh_pipeline.file_claim(
            "clm-a", policy_id="pol-sort", coverage_type="epl",
            incident_date=D0, filed_at=T0,
            claimed_amount=10_000, reserve_amount=8_000, actor="adj-1",
        )
        claims = fresh_pipeline.get_claims_by_policy("pol-sort")
        filed_ats = [c["filed_at"] for c in claims]
        assert filed_ats == sorted(filed_ats)

    def test_unknown_policy_returns_empty(self, pipeline):
        assert pipeline.get_claims_by_policy("pol-999") == []


class TestGetOpenClaims:
    def test_excludes_closed_and_denied(self, pipeline):
        open_claims = pipeline.get_open_claims()
        ids = [c["claim_id"] for c in open_claims]
        # clm-001 is settled (not yet closed, so it IS open — but clm-004 is denied → excluded)
        assert "clm-001" in ids   # settled but not closed → open
        assert "clm-002" in ids   # investigating → open
        assert "clm-003" in ids   # evaluation → open
        assert "clm-004" not in ids  # denied → excluded

    def test_sorted_by_filed_at(self, pipeline):
        open_claims = pipeline.get_open_claims()
        filed_ats = [c["filed_at"] for c in open_claims]
        assert filed_ats == sorted(filed_ats)


# ---------------------------------------------------------------------------
# PART 3 — Reserve adequacy and metrics
# ---------------------------------------------------------------------------

class TestGetReserveAdequacy:
    def test_total_reserves_all_claims(self, pipeline):
        adequacy = pipeline.get_reserve_adequacy()
        # clm-001: 50_000  clm-002: 30_000  clm-003: 150_000  clm-004: 10_000
        assert adequacy["total_reserves"] == 240_000

    def test_total_approved_settled_only(self, pipeline):
        # Only clm-001 is settled, approved=60_000
        adequacy = pipeline.get_reserve_adequacy()
        assert adequacy["total_approved"] == 60_000

    def test_under_reserved_count_and_gap(self, pipeline):
        # clm-001: reserve=50_000 vs approved=60_000 → under-reserved by 10_000
        adequacy = pipeline.get_reserve_adequacy()
        assert adequacy["under_reserved_count"] == 1
        assert adequacy["under_reserved_gap"] == 10_000

    def test_empty_pipeline(self, fresh_pipeline):
        adequacy = fresh_pipeline.get_reserve_adequacy()
        assert adequacy["total_reserves"] == 0
        assert adequacy["total_approved"] == 0
        assert adequacy["under_reserved_count"] == 0
        assert adequacy["under_reserved_gap"] == 0


class TestGetClaimsMetrics:
    def test_total_count(self, pipeline):
        metrics = pipeline.get_claims_metrics()
        assert metrics["total"] == 4

    def test_by_status(self, pipeline):
        metrics = pipeline.get_claims_metrics()
        assert metrics["by_status"].get("settled") == 1
        assert metrics["by_status"].get("investigating") == 1
        assert metrics["by_status"].get("evaluation") == 1
        assert metrics["by_status"].get("denied") == 1

    def test_only_nonzero_statuses(self, pipeline):
        metrics = pipeline.get_claims_metrics()
        for status, count in metrics["by_status"].items():
            assert count > 0

    def test_total_claimed(self, pipeline):
        metrics = pipeline.get_claims_metrics()
        assert metrics["total_claimed"] == 75_000 + 30_000 + 200_000 + 10_000

    def test_total_paid(self, pipeline):
        # only clm-001 settled, approved=60_000
        metrics = pipeline.get_claims_metrics()
        assert metrics["total_paid"] == 60_000

    def test_avg_settlement_ratio(self, pipeline):
        # 1 settled claim: approved=60_000 / claimed=75_000 = 0.8
        metrics = pipeline.get_claims_metrics()
        assert metrics["avg_settlement_ratio"] == round(60_000 / 75_000, 4)

    def test_no_settled_claims_avg_ratio_is_zero(self, fresh_pipeline):
        fresh_pipeline.file_claim(
            "clm-ns", policy_id="pol-200", coverage_type="epl",
            incident_date=D0, filed_at=T0,
            claimed_amount=20_000, reserve_amount=15_000, actor="adj-1",
        )
        metrics = fresh_pipeline.get_claims_metrics()
        assert metrics["avg_settlement_ratio"] == 0.0


class TestGetPolicyLossHistory:
    def test_returns_correct_stats(self, pipeline):
        history = pipeline.get_policy_loss_history("pol-101")
        assert history["policy_id"] == "pol-101"
        assert history["claim_count"] == 2
        assert history["total_claimed"] == 105_000   # 75_000 + 30_000
        assert history["total_paid"] == 60_000        # only clm-001 settled
        assert history["loss_ratio"] == round(60_000 / 105_000, 4)

    def test_unknown_policy_returns_zeros(self, pipeline):
        history = pipeline.get_policy_loss_history("pol-999")
        assert history["policy_id"] == "pol-999"
        assert history["claim_count"] == 0
        assert history["total_claimed"] == 0
        assert history["total_paid"] == 0
        assert history["loss_ratio"] == 0.0
