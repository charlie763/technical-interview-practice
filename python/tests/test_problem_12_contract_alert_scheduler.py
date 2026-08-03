"""Tests for Problem 12: Contract Expiration Alert Scheduler

Run from the python/ directory:
    pytest tests/test_problem_12_contract_alert_scheduler.py -v
"""

import pytest
from practice_problems.problem_12_contract_alert_scheduler import ContractAlertScheduler

# ---------------------------------------------------------------------------
# Shared dates
# D0  = base date
# D30 = D0 + 30 days
# D60 = D0 + 60 days
# D90 = D0 + 90 days
# ---------------------------------------------------------------------------
D0  = "2025-01-01"
D30 = "2025-01-31"
D60 = "2025-03-02"   # Jan has 31 days, Feb 2025 has 28 days → Jan 1 + 60 = Mar 2
D90 = "2025-04-01"


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def fresh_sched():
    """Empty ContractAlertScheduler."""
    return ContractAlertScheduler()


@pytest.fixture
def sched():
    """
    Pre-seeded scheduler:
      c-seed-1  "Vendor MSA"        owner=legal@acme.com  expires=2025-06-30
      c-seed-2  "SaaS Subscription" owner=ops@acme.com    expires=2025-09-15
      c-seed-3  "NDA Agreement"     owner=legal@acme.com  expires=2025-12-31

    Alert configs:
      cfg-30   days_before=30  label="30-day notice"
      cfg-7    days_before=7   label="final warning"
    """
    s = ContractAlertScheduler()
    s.add_contract("c-seed-1", "Vendor MSA",        "legal@acme.com", expires_on="2025-06-30")
    s.add_contract("c-seed-2", "SaaS Subscription", "ops@acme.com",   expires_on="2025-09-15")
    s.add_contract("c-seed-3", "NDA Agreement",     "legal@acme.com", expires_on="2025-12-31")
    s.add_alert_config("cfg-30", days_before=30, label="30-day notice")
    s.add_alert_config("cfg-7",  days_before=7,  label="final warning")
    return s


# ---------------------------------------------------------------------------
# PART 1 — Contract and alert-config management
# ---------------------------------------------------------------------------

class TestAddContract:
    def test_returns_contract_dict(self, fresh_sched):
        c = fresh_sched.add_contract("c-add-1", "Test Contract", "a@b.com", expires_on="2025-06-01")
        assert c["contract_id"] == "c-add-1"
        assert c["title"] == "Test Contract"
        assert c["owner_email"] == "a@b.com"
        assert c["expires_on"] == "2025-06-01"

    def test_duplicate_raises(self, fresh_sched):
        fresh_sched.add_contract("c-dup-1", "A", "a@b.com", expires_on="2025-06-01")
        with pytest.raises(ValueError):
            fresh_sched.add_contract("c-dup-1", "B", "b@c.com", expires_on="2025-07-01")

    def test_multiple_contracts_stored(self, sched):
        # sched has 3 contracts; retrieve their expiry dates via expiring_between
        results = sched.get_contracts_expiring_between("2025-01-01", "2025-12-31")
        ids = [r["contract_id"] for r in results]
        assert "c-seed-1" in ids
        assert "c-seed-2" in ids
        assert "c-seed-3" in ids


class TestAddAlertConfig:
    def test_returns_config_dict(self, fresh_sched):
        cfg = fresh_sched.add_alert_config("cfg-add-1", days_before=14, label="two-week notice")
        assert cfg["config_id"] == "cfg-add-1"
        assert cfg["days_before"] == 14
        assert cfg["label"] == "two-week notice"

    def test_duplicate_raises(self, fresh_sched):
        fresh_sched.add_alert_config("cfg-dup-1", days_before=30, label="notice")
        with pytest.raises(ValueError):
            fresh_sched.add_alert_config("cfg-dup-1", days_before=60, label="other")


class TestGetContractsExpiringBetween:
    def test_exact_range_match(self, sched):
        results = sched.get_contracts_expiring_between("2025-06-30", "2025-06-30")
        assert len(results) == 1
        assert results[0]["contract_id"] == "c-seed-1"

    def test_range_spans_multiple(self, sched):
        results = sched.get_contracts_expiring_between("2025-06-01", "2025-09-30")
        ids = [r["contract_id"] for r in results]
        assert "c-seed-1" in ids
        assert "c-seed-2" in ids
        assert "c-seed-3" not in ids

    def test_sorted_ascending(self, sched):
        results = sched.get_contracts_expiring_between("2025-01-01", "2025-12-31")
        dates = [r["expires_on"] for r in results]
        assert dates == sorted(dates)

    def test_empty_when_none_in_range(self, sched):
        results = sched.get_contracts_expiring_between("2024-01-01", "2024-12-31")
        assert results == []

    def test_inclusive_start_boundary(self, sched):
        # c-seed-1 expires exactly on 2025-06-30; start = 2025-06-30 should include it
        results = sched.get_contracts_expiring_between("2025-06-30", "2025-12-31")
        ids = [r["contract_id"] for r in results]
        assert "c-seed-1" in ids

    def test_inclusive_end_boundary(self, sched):
        results = sched.get_contracts_expiring_between("2025-01-01", "2025-06-30")
        ids = [r["contract_id"] for r in results]
        assert "c-seed-1" in ids


# ---------------------------------------------------------------------------
# PART 2 — Alert schedule computation
# ---------------------------------------------------------------------------

class TestComputeAlertSchedule:
    def test_returns_entry_per_config(self, sched):
        schedule = sched.compute_alert_schedule("c-seed-1")
        # 2 configs registered → 2 schedule entries
        assert len(schedule) == 2

    def test_alert_on_dates_correct(self, sched):
        # c-seed-1 expires 2025-06-30
        # cfg-30: 2025-06-30 - 30d = 2025-05-31
        # cfg-7:  2025-06-30 - 7d  = 2025-06-23
        schedule = sched.compute_alert_schedule("c-seed-1")
        by_cfg = {e["config_id"]: e for e in schedule}
        assert by_cfg["cfg-30"]["alert_on"] == "2025-05-31"
        assert by_cfg["cfg-7"]["alert_on"]  == "2025-06-23"

    def test_sorted_by_alert_on_ascending(self, sched):
        schedule = sched.compute_alert_schedule("c-seed-1")
        dates = [e["alert_on"] for e in schedule]
        assert dates == sorted(dates)

    def test_includes_label(self, sched):
        schedule = sched.compute_alert_schedule("c-seed-1")
        labels = {e["label"] for e in schedule}
        assert "30-day notice" in labels
        assert "final warning" in labels

    def test_unknown_contract_raises(self, sched):
        with pytest.raises(KeyError):
            sched.compute_alert_schedule("no-such-contract")

    def test_no_configs_returns_empty_list(self, fresh_sched):
        fresh_sched.add_contract("c-no-cfg", "Bare Contract", "a@b.com", expires_on="2025-06-01")
        assert fresh_sched.compute_alert_schedule("c-no-cfg") == []


class TestGetDueAlerts:
    def test_returns_alerts_on_or_before_date(self, sched):
        # cfg-30 for c-seed-1 fires on 2025-05-31
        due = sched.get_due_alerts(as_of_date="2025-05-31")
        entries = [(e["contract_id"], e["config_id"]) for e in due]
        assert ("c-seed-1", "cfg-30") in entries

    def test_excludes_future_alerts(self, sched):
        # 2025-05-01 → no alerts yet for any contract
        due = sched.get_due_alerts(as_of_date="2025-01-01")
        assert due == []

    def test_includes_owner_email_and_expires_on(self, sched):
        due = sched.get_due_alerts(as_of_date="2025-05-31")
        entry = next(e for e in due if e["contract_id"] == "c-seed-1" and e["config_id"] == "cfg-30")
        assert entry["owner_email"] == "legal@acme.com"
        assert entry["expires_on"]  == "2025-06-30"

    def test_sorted_by_alert_on_then_contract_id(self, sched):
        # Ask for a date far enough in the future to capture many alerts
        due = sched.get_due_alerts(as_of_date="2025-12-31")
        dates = [e["alert_on"] for e in due]
        assert dates == sorted(dates)


# ---------------------------------------------------------------------------
# PART 3 — Sent records and upcoming alerts
# ---------------------------------------------------------------------------

class TestRecordAlertSent:
    def test_returns_sent_record(self, sched):
        rec = sched.record_alert_sent("c-seed-1", "cfg-30", sent_on="2025-05-31")
        assert rec["contract_id"] == "c-seed-1"
        assert rec["config_id"]   == "cfg-30"
        assert rec["sent_on"]     == "2025-05-31"

    def test_unknown_contract_raises(self, sched):
        with pytest.raises(KeyError):
            sched.record_alert_sent("no-contract", "cfg-30", sent_on="2025-05-31")

    def test_unknown_config_raises(self, sched):
        with pytest.raises(KeyError):
            sched.record_alert_sent("c-seed-1", "no-cfg", sent_on="2025-05-31")

    def test_multiple_sends_stored(self, sched):
        sched.record_alert_sent("c-seed-1", "cfg-30", sent_on="2025-05-31")
        sched.record_alert_sent("c-seed-1", "cfg-7",  sent_on="2025-06-23")
        upcoming = sched.get_upcoming_alerts("c-seed-1", as_of_date="2025-05-01")
        sent_ids = {e["config_id"] for e in upcoming if e["sent"]}
        assert "cfg-30" in sent_ids
        assert "cfg-7" in sent_ids


class TestGetUpcomingAlerts:
    def test_excludes_past_alerts(self, sched):
        # as_of_date = 2025-06-01; cfg-30 alert_on=2025-05-31 is in the past
        upcoming = sched.get_upcoming_alerts("c-seed-1", as_of_date="2025-06-01")
        config_ids = [e["config_id"] for e in upcoming]
        assert "cfg-30" not in config_ids

    def test_includes_future_alerts(self, sched):
        # cfg-7 alert_on=2025-06-23 is still upcoming from 2025-06-01
        upcoming = sched.get_upcoming_alerts("c-seed-1", as_of_date="2025-06-01")
        config_ids = [e["config_id"] for e in upcoming]
        assert "cfg-7" in config_ids

    def test_sent_flag_false_by_default(self, sched):
        upcoming = sched.get_upcoming_alerts("c-seed-1", as_of_date="2025-05-01")
        for e in upcoming:
            assert e["sent"] is False

    def test_sent_flag_true_after_record(self, sched):
        sched.record_alert_sent("c-seed-1", "cfg-30", sent_on="2025-05-31")
        upcoming = sched.get_upcoming_alerts("c-seed-1", as_of_date="2025-05-01")
        entry = next(e for e in upcoming if e["config_id"] == "cfg-30")
        assert entry["sent"] is True

    def test_sorted_by_alert_on_ascending(self, sched):
        upcoming = sched.get_upcoming_alerts("c-seed-1", as_of_date="2025-01-01")
        dates = [e["alert_on"] for e in upcoming]
        assert dates == sorted(dates)

    def test_unknown_contract_raises(self, sched):
        with pytest.raises(KeyError):
            sched.get_upcoming_alerts("no-such", as_of_date="2025-01-01")
