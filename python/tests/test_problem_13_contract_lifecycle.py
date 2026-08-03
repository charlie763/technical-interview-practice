"""Tests for Problem 13: Contract Lifecycle State Machine

Run from the python/ directory:
    pytest tests/test_problem_13_contract_lifecycle.py -v
"""

import pytest
from practice_problems.problem_13_contract_lifecycle import ContractLifecycle

# ---------------------------------------------------------------------------
# Shared timestamps
# T0 = base datetime
# T1 = T0 + 1 day
# T2 = T0 + 5 days
# T3 = T0 + 10 days
# T4 = T0 + 40 days  (> 30 days for overdue tests)
# T5 = T0 + 45 days
# ---------------------------------------------------------------------------
T0 = "2025-01-01T09:00:00"
T1 = "2025-01-02T10:00:00"
T2 = "2025-01-06T11:00:00"
T3 = "2025-01-11T12:00:00"
T4 = "2025-02-10T09:00:00"   # 40 days after T0
T5 = "2025-02-15T09:00:00"   # 45 days after T0


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def fresh_cl():
    """Empty ContractLifecycle."""
    return ContractLifecycle()


@pytest.fixture
def cl():
    """
    Pre-seeded lifecycle manager:
      c-seed-1  "Vendor MSA"        state=approved    (draft→in_review→approved)
      c-seed-2  "NDA Agreement"     state=draft
      c-seed-3  "SaaS License"      state=terminated  (draft→in_review→approved→executed→active→terminated)
    """
    c = ContractLifecycle()
    # c-seed-1: draft → in_review → approved
    c.create_contract("c-seed-1", "Vendor MSA",    created_at=T0, actor="alice")
    c.transition("c-seed-1", "in_review", at=T1, actor="alice")
    c.transition("c-seed-1", "approved",  at=T2, actor="bob")
    # c-seed-2: stays draft
    c.create_contract("c-seed-2", "NDA Agreement", created_at=T0, actor="alice")
    # c-seed-3: full path to terminated
    c.create_contract("c-seed-3", "SaaS License",  created_at=T0, actor="carol")
    c.transition("c-seed-3", "in_review", at=T1, actor="carol")
    c.transition("c-seed-3", "approved",  at=T2, actor="bob")
    c.transition("c-seed-3", "executed",  at=T3, actor="carol")
    c.transition("c-seed-3", "active",    at=T4, actor="carol")
    c.transition("c-seed-3", "terminated",at=T5, actor="carol")
    return c


# ---------------------------------------------------------------------------
# PART 1 — Contract creation, field management, transitions
# ---------------------------------------------------------------------------

class TestCreateContract:
    def test_returns_contract_in_draft(self, fresh_cl):
        c = fresh_cl.create_contract("c-cr-1", "Title", created_at=T0, actor="alice")
        assert c["contract_id"] == "c-cr-1"
        assert c["state"] == "draft"
        assert c["title"] == "Title"
        assert c["fields"] == {}

    def test_stores_created_at(self, fresh_cl):
        c = fresh_cl.create_contract("c-cr-ts", "Title", created_at=T0, actor="alice")
        assert c["created_at"] == T0

    def test_duplicate_raises(self, fresh_cl):
        fresh_cl.create_contract("c-dup-lc", "A", created_at=T0, actor="alice")
        with pytest.raises(ValueError):
            fresh_cl.create_contract("c-dup-lc", "B", created_at=T1, actor="alice")

    def test_initial_audit_entry_recorded(self, fresh_cl):
        fresh_cl.create_contract("c-audit-init", "T", created_at=T0, actor="alice")
        trail = fresh_cl.get_audit_trail("c-audit-init")
        assert len(trail) == 1
        assert trail[0]["from_state"] is None
        assert trail[0]["to_state"] == "draft"
        assert trail[0]["actor"] == "alice"


class TestSetField:
    def test_sets_field(self, cl):
        cl.set_field("c-seed-1", "value", 100000)
        assert cl.get_contract("c-seed-1")["fields"]["value"] == 100000

    def test_updates_existing_field(self, cl):
        cl.set_field("c-seed-1", "value", 50000)
        cl.set_field("c-seed-1", "value", 75000)
        assert cl.get_contract("c-seed-1")["fields"]["value"] == 75000

    def test_unknown_contract_raises(self, cl):
        with pytest.raises(KeyError):
            cl.set_field("no-such", "key", "val")


class TestGetContract:
    def test_returns_contract(self, cl):
        c = cl.get_contract("c-seed-1")
        assert c["contract_id"] == "c-seed-1"
        assert c["state"] == "approved"

    def test_unknown_raises(self, cl):
        with pytest.raises(KeyError):
            cl.get_contract("nonexistent")


class TestTransition:
    def test_valid_transition_updates_state(self, fresh_cl):
        fresh_cl.create_contract("c-tr-1", "T", created_at=T0, actor="alice")
        fresh_cl.transition("c-tr-1", "in_review", at=T1, actor="alice")
        assert fresh_cl.get_contract("c-tr-1")["state"] == "in_review"

    def test_invalid_transition_raises(self, fresh_cl):
        fresh_cl.create_contract("c-tr-inv", "T", created_at=T0, actor="alice")
        with pytest.raises(ValueError):
            fresh_cl.transition("c-tr-inv", "approved", at=T1, actor="alice")  # draft→approved invalid

    def test_terminal_state_raises(self, cl):
        with pytest.raises(ValueError):
            cl.transition("c-seed-3", "draft", at=T5, actor="alice")  # terminated is terminal

    def test_back_transition_draft_to_in_review_to_draft(self, fresh_cl):
        fresh_cl.create_contract("c-back", "T", created_at=T0, actor="alice")
        fresh_cl.transition("c-back", "in_review", at=T1, actor="alice")
        fresh_cl.transition("c-back", "draft",     at=T2, actor="alice")
        assert fresh_cl.get_contract("c-back")["state"] == "draft"

    def test_transition_appends_audit_entry(self, fresh_cl):
        fresh_cl.create_contract("c-tr-audit", "T", created_at=T0, actor="alice")
        fresh_cl.transition("c-tr-audit", "in_review", at=T1, actor="bob")
        trail = fresh_cl.get_audit_trail("c-tr-audit")
        assert len(trail) == 2
        last = trail[-1]
        assert last["from_state"] == "draft"
        assert last["to_state"]   == "in_review"
        assert last["actor"]      == "bob"

    def test_unknown_contract_raises(self, cl):
        with pytest.raises(KeyError):
            cl.transition("no-such", "in_review", at=T1, actor="alice")


# ---------------------------------------------------------------------------
# PART 2 — Audit trail, by-state query, bulk advance
# ---------------------------------------------------------------------------

class TestGetAuditTrail:
    def test_returns_all_entries_ordered(self, cl):
        trail = cl.get_audit_trail("c-seed-1")
        # create(draft) + in_review + approved = 3 entries
        assert len(trail) == 3
        to_states = [e["to_state"] for e in trail]
        assert to_states == ["draft", "in_review", "approved"]

    def test_terminal_contract_full_trail(self, cl):
        trail = cl.get_audit_trail("c-seed-3")
        to_states = [e["to_state"] for e in trail]
        assert to_states == ["draft", "in_review", "approved", "executed", "active", "terminated"]

    def test_unknown_contract_raises(self, cl):
        with pytest.raises(KeyError):
            cl.get_audit_trail("no-such")


class TestGetContractsByState:
    def test_returns_correct_contracts(self, cl):
        approved = cl.get_contracts_by_state("approved")
        ids = [c["contract_id"] for c in approved]
        assert "c-seed-1" in ids
        assert "c-seed-2" not in ids

    def test_sorted_by_contract_id(self, fresh_cl):
        fresh_cl.create_contract("c-z", "Z", created_at=T0, actor="a")
        fresh_cl.create_contract("c-a", "A", created_at=T0, actor="a")
        fresh_cl.create_contract("c-m", "M", created_at=T0, actor="a")
        drafts = fresh_cl.get_contracts_by_state("draft")
        ids = [c["contract_id"] for c in drafts]
        assert ids == sorted(ids)

    def test_empty_for_unused_state(self, cl):
        assert cl.get_contracts_by_state("expired") == []


class TestBulkAdvance:
    def test_all_succeed(self, fresh_cl):
        for i in range(3):
            cid = f"c-bulk-{i}"
            fresh_cl.create_contract(cid, f"Contract {i}", created_at=T0, actor="alice")
        result = fresh_cl.bulk_advance(
            ["c-bulk-0", "c-bulk-1", "c-bulk-2"], "in_review", at=T1, actor="alice"
        )
        assert len(result["succeeded"]) == 3
        assert len(result["failed"])    == 0

    def test_partial_failure_continues(self, fresh_cl):
        # c-ok is draft; c-bad is already in_review (can't go back to draft via bulk)
        fresh_cl.create_contract("c-ok",  "OK",  created_at=T0, actor="alice")
        fresh_cl.create_contract("c-bad", "Bad", created_at=T0, actor="alice")
        fresh_cl.transition("c-bad", "in_review", at=T1, actor="alice")
        result = fresh_cl.bulk_advance(
            ["c-ok", "c-bad"], "in_review", at=T2, actor="alice"
        )
        assert "c-ok" in result["succeeded"]
        assert any(f["contract_id"] == "c-bad" for f in result["failed"])

    def test_states_updated_for_successes(self, fresh_cl):
        fresh_cl.create_contract("c-bs-1", "A", created_at=T0, actor="alice")
        fresh_cl.create_contract("c-bs-2", "B", created_at=T0, actor="alice")
        fresh_cl.bulk_advance(["c-bs-1", "c-bs-2"], "in_review", at=T1, actor="alice")
        assert fresh_cl.get_contract("c-bs-1")["state"] == "in_review"
        assert fresh_cl.get_contract("c-bs-2")["state"] == "in_review"

    def test_failed_entry_includes_reason(self, fresh_cl):
        fresh_cl.create_contract("c-fail-r", "X", created_at=T0, actor="alice")
        result = fresh_cl.bulk_advance(["c-fail-r"], "approved", at=T1, actor="alice")
        assert len(result["failed"]) == 1
        assert result["failed"][0]["reason"] != ""


# ---------------------------------------------------------------------------
# PART 3 — Lifecycle metrics and overdue contracts
# ---------------------------------------------------------------------------

class TestGetLifecycleMetrics:
    def test_total_count(self, cl):
        metrics = cl.get_lifecycle_metrics()
        assert metrics["total"] == 3

    def test_by_state_counts(self, cl):
        metrics = cl.get_lifecycle_metrics()
        assert metrics["by_state"].get("approved")    == 1
        assert metrics["by_state"].get("draft")       == 1
        assert metrics["by_state"].get("terminated")  == 1

    def test_only_nonzero_states_in_by_state(self, cl):
        metrics = cl.get_lifecycle_metrics()
        for state, count in metrics["by_state"].items():
            assert count > 0

    def test_terminal_count(self, cl):
        metrics = cl.get_lifecycle_metrics()
        # c-seed-3 is terminated
        assert metrics["terminal_count"] == 1

    def test_empty_manager_returns_zeros(self, fresh_cl):
        metrics = fresh_cl.get_lifecycle_metrics()
        assert metrics["total"] == 0
        assert metrics["by_state"] == {}
        assert metrics["terminal_count"] == 0


class TestGetOverdueContracts:
    def test_returns_contracts_stuck_over_30_days(self, fresh_cl):
        # Create contract that transitions to in_review at T0, then nothing
        fresh_cl.create_contract("c-overdue", "Old Contract", created_at=T0, actor="alice")
        fresh_cl.transition("c-overdue", "in_review", at=T0, actor="alice")
        # T4 = T0 + 40 days → should be overdue
        overdue = fresh_cl.get_overdue_contracts(as_of=T4)
        ids = [o["contract_id"] for o in overdue]
        assert "c-overdue" in ids

    def test_recent_transition_not_overdue(self, fresh_cl):
        fresh_cl.create_contract("c-recent", "New Contract", created_at=T0, actor="alice")
        fresh_cl.transition("c-recent", "in_review", at=T3, actor="alice")
        # T4 = T3 + ~30 days; T4-T3 = T0+40 - T0+10 = 30 days exactly; use T4 + 1 extra day
        overdue = fresh_cl.get_overdue_contracts(as_of=T4)
        ids = [o["contract_id"] for o in overdue]
        assert "c-recent" not in ids

    def test_terminal_contracts_excluded(self, cl):
        overdue = cl.get_overdue_contracts(as_of=T5)
        ids = [o["contract_id"] for o in overdue]
        assert "c-seed-3" not in ids  # terminated → terminal

    def test_sorted_by_days_stuck_descending(self, fresh_cl):
        fresh_cl.create_contract("c-od-a", "A", created_at=T0, actor="alice")
        fresh_cl.create_contract("c-od-b", "B", created_at=T0, actor="alice")
        fresh_cl.transition("c-od-a", "in_review", at=T0, actor="alice")
        fresh_cl.transition("c-od-b", "in_review", at=T1, actor="alice")
        overdue = fresh_cl.get_overdue_contracts(as_of=T4)
        days = [o["days_stuck"] for o in overdue]
        assert days == sorted(days, reverse=True)

    def test_result_includes_required_fields(self, fresh_cl):
        fresh_cl.create_contract("c-od-f", "Fields Test", created_at=T0, actor="alice")
        fresh_cl.transition("c-od-f", "in_review", at=T0, actor="alice")
        overdue = fresh_cl.get_overdue_contracts(as_of=T4)
        entry = next(o for o in overdue if o["contract_id"] == "c-od-f")
        assert "title" in entry
        assert "state" in entry
        assert "stuck_since" in entry
        assert "days_stuck" in entry
        assert entry["days_stuck"] >= 31
