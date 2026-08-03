"""Tests for Problem 9: Multi-Source Incident Aggregator

Run from the python/ directory:
    pytest tests/test_problem_09_incident_aggregator.py -v
"""

import pytest
from practice_problems.problem_09_incident_aggregator import IncidentAggregator

# ---------------------------------------------------------------------------
# Shared timestamps  (all naive ISO-8601, lexicographically sortable)
# T0 = base
# T1 = T0 + 60s
# T2 = T0 + 120s
# T3 = T0 + 300s
# T4 = T0 + 600s
# ---------------------------------------------------------------------------
T0 = "2024-06-01T10:00:00"
T1 = "2024-06-01T10:01:00"
T2 = "2024-06-01T10:02:00"
T3 = "2024-06-01T10:05:00"
T4 = "2024-06-01T10:10:00"


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def fresh_agg():
    """Empty IncidentAggregator."""
    return IncidentAggregator()


@pytest.fixture
def agg():
    """
    Pre-seeded aggregator:
      Reports: r1 (shooting/downtown/T0), r2 (shooting/downtown/T1),
               r3 (car-crash/midtown/T2)
      Incidents: inc-001 (shooting/downtown) containing r1 and r2
      r3 is unassigned.
    """
    a = IncidentAggregator()
    a.ingest_report("r1", "radio-north", "shooting", "downtown", T0)
    a.ingest_report("r2", "radio-south", "shooting", "downtown", T1)
    a.ingest_report("r3", "social-feed", "car-crash", "midtown",  T2)
    a.create_incident("inc-001", "shooting", "downtown")
    a.add_report_to_incident("inc-001", "r1")
    a.add_report_to_incident("inc-001", "r2")
    return a


# ---------------------------------------------------------------------------
# PART 1 — Report ingestion
# ---------------------------------------------------------------------------

class TestIngestReport:
    def test_stores_and_returns_report(self, fresh_agg):
        r = fresh_agg.ingest_report("r_store", "radio-north", "shooting", "downtown", T0)
        assert r["report_id"]    == "r_store"
        assert r["source_id"]    == "radio-north"
        assert r["event_type"]   == "shooting"
        assert r["location_key"] == "downtown"
        assert r["ts"]           == T0
        assert r["incident_id"]  is None

    def test_duplicate_report_id_raises(self, fresh_agg):
        fresh_agg.ingest_report("r_dup", "src-a", "fire", "east", T0)
        with pytest.raises(ValueError):
            fresh_agg.ingest_report("r_dup", "src-b", "fire", "east", T1)

    def test_get_report_returns_none_if_missing(self, fresh_agg):
        assert fresh_agg.get_report("nonexistent") is None

    def test_get_report_returns_stored(self, agg):
        r = agg.get_report("r1")
        assert r is not None
        assert r["event_type"] == "shooting"


class TestGetReports:
    def test_no_filter_returns_all(self, agg):
        assert len(agg.get_reports()) == 3

    def test_filter_by_location(self, agg):
        reports = agg.get_reports(location_key="downtown")
        assert len(reports) == 2
        assert all(r["location_key"] == "downtown" for r in reports)

    def test_filter_by_event_type(self, agg):
        reports = agg.get_reports(event_type="car-crash")
        assert len(reports) == 1
        assert reports[0]["report_id"] == "r3"

    def test_filter_both(self, agg):
        reports = agg.get_reports(location_key="downtown", event_type="shooting")
        assert len(reports) == 2

    def test_empty_when_no_match(self, agg):
        assert agg.get_reports(location_key="mars") == []

    def test_sorted_by_ts_ascending(self, fresh_agg):
        fresh_agg.ingest_report("r_sort_b", "src", "fire", "zone-1", T1)
        fresh_agg.ingest_report("r_sort_a", "src", "fire", "zone-1", T0)
        reports = fresh_agg.get_reports()
        tss = [r["ts"] for r in reports]
        assert tss == sorted(tss)


# ---------------------------------------------------------------------------
# PART 2 — Manual incident grouping
# ---------------------------------------------------------------------------

class TestCreateIncident:
    def test_creates_empty_incident(self, fresh_agg):
        inc = fresh_agg.create_incident("inc_create_test", "fire", "east-side")
        assert inc["incident_id"]  == "inc_create_test"
        assert inc["event_type"]   == "fire"
        assert inc["location_key"] == "east-side"
        assert inc["report_count"] == 0
        assert inc["report_ids"]   == []
        assert inc["latest_ts"]    is None

    def test_duplicate_incident_id_raises(self, agg):
        with pytest.raises(ValueError):
            agg.create_incident("inc-001", "shooting", "downtown")


class TestAddReportToIncident:
    def test_assigns_report(self, fresh_agg):
        fresh_agg.ingest_report("r_assign", "src", "fire", "east", T0)
        fresh_agg.create_incident("inc_assign_test", "fire", "east")
        fresh_agg.add_report_to_incident("inc_assign_test", "r_assign")
        inc = fresh_agg.get_incident("inc_assign_test")
        assert "r_assign" in inc["report_ids"]
        assert inc["report_count"] == 1
        assert inc["latest_ts"]    == T0

    def test_updates_report_incident_id(self, fresh_agg):
        fresh_agg.ingest_report("r_update", "src", "fire", "east", T0)
        fresh_agg.create_incident("inc_update_test", "fire", "east")
        fresh_agg.add_report_to_incident("inc_update_test", "r_update")
        assert fresh_agg.get_report("r_update")["incident_id"] == "inc_update_test"

    def test_already_assigned_raises(self, agg):
        # r1 is already in inc-001
        with pytest.raises(ValueError):
            agg.add_report_to_incident("inc-001", "r1")

    def test_bad_incident_raises(self, fresh_agg):
        fresh_agg.ingest_report("r_bad_inc", "src", "fire", "east", T0)
        with pytest.raises(KeyError):
            fresh_agg.add_report_to_incident("nonexistent_inc", "r_bad_inc")

    def test_bad_report_raises(self, agg):
        with pytest.raises(KeyError):
            agg.add_report_to_incident("inc-001", "nonexistent_report")

    def test_report_ids_sorted_by_ts(self, fresh_agg):
        fresh_agg.ingest_report("r_ts_b", "src", "fire", "east", T1)
        fresh_agg.ingest_report("r_ts_a", "src", "fire", "east", T0)
        fresh_agg.create_incident("inc_ts_test", "fire", "east")
        fresh_agg.add_report_to_incident("inc_ts_test", "r_ts_b")
        fresh_agg.add_report_to_incident("inc_ts_test", "r_ts_a")
        inc = fresh_agg.get_incident("inc_ts_test")
        assert inc["report_ids"] == ["r_ts_a", "r_ts_b"]

    def test_latest_ts_updates_to_newest(self, fresh_agg):
        fresh_agg.ingest_report("r_lt_a", "src", "fire", "east", T0)
        fresh_agg.ingest_report("r_lt_b", "src", "fire", "east", T2)
        fresh_agg.create_incident("inc_lt_test", "fire", "east")
        fresh_agg.add_report_to_incident("inc_lt_test", "r_lt_a")
        fresh_agg.add_report_to_incident("inc_lt_test", "r_lt_b")
        assert fresh_agg.get_incident("inc_lt_test")["latest_ts"] == T2


class TestGetIncident:
    def test_returns_none_if_not_found(self, fresh_agg):
        assert fresh_agg.get_incident("ghost") is None

    def test_returns_incident(self, agg):
        inc = agg.get_incident("inc-001")
        assert inc is not None
        assert inc["report_count"] == 2


class TestGetUnassignedReports:
    def test_returns_unassigned(self, agg):
        unassigned = agg.get_unassigned_reports()
        assert len(unassigned) == 1
        assert unassigned[0]["report_id"] == "r3"

    def test_empty_when_all_assigned(self, fresh_agg):
        fresh_agg.ingest_report("r_all", "src", "fire", "east", T0)
        fresh_agg.create_incident("inc_all_test", "fire", "east")
        fresh_agg.add_report_to_incident("inc_all_test", "r_all")
        assert fresh_agg.get_unassigned_reports() == []

    def test_sorted_by_ts_ascending(self, fresh_agg):
        fresh_agg.ingest_report("r_ua_b", "src", "fire", "east", T1)
        fresh_agg.ingest_report("r_ua_a", "src", "fire", "east", T0)
        unassigned = fresh_agg.get_unassigned_reports()
        tss = [r["ts"] for r in unassigned]
        assert tss == sorted(tss)


# ---------------------------------------------------------------------------
# PART 3 — Automatic deduplication
# ---------------------------------------------------------------------------

class TestAutoIngestReport:
    def test_creates_new_incident_when_no_match(self, fresh_agg):
        incident_id = fresh_agg.auto_ingest_report(
            "r_auto_new", "src", "fire", "east-side", T0, time_window_secs=120
        )
        assert incident_id is not None
        inc = fresh_agg.get_incident(incident_id)
        assert inc["report_count"] == 1
        assert "r_auto_new" in inc["report_ids"]

    def test_merges_into_existing_when_within_window(self, agg):
        # inc-001 latest_ts = T1; new ts = T2; T2 - T1 = 60s <= 120s
        result = agg.auto_ingest_report(
            "r_merge", "radio-east", "shooting", "downtown", T2, time_window_secs=120
        )
        assert result == "inc-001"
        assert agg.get_incident("inc-001")["report_count"] == 3

    def test_no_merge_different_event_type(self, agg):
        result = agg.auto_ingest_report(
            "r_diff_type", "src", "car-crash", "downtown", T2, time_window_secs=120
        )
        assert result != "inc-001"

    def test_no_merge_different_location(self, agg):
        result = agg.auto_ingest_report(
            "r_diff_loc", "src", "shooting", "uptown", T2, time_window_secs=120
        )
        assert result != "inc-001"

    def test_no_merge_outside_window(self, agg):
        # inc-001 latest_ts = T1; new ts = T4; T4 - T1 = 540s > 120s
        result = agg.auto_ingest_report(
            "r_expired", "src", "shooting", "downtown", T4, time_window_secs=120
        )
        assert result != "inc-001"

    def test_new_report_stored(self, fresh_agg):
        incident_id = fresh_agg.auto_ingest_report(
            "r_stored", "src", "fire", "east", T0, time_window_secs=60
        )
        assert fresh_agg.get_report("r_stored") is not None

    def test_picks_most_recently_active_match(self, fresh_agg):
        # Two incidents for the same type+location, both within the window
        fresh_agg.ingest_report("ra1", "src", "shooting", "downtown", T0)
        fresh_agg.ingest_report("ra2", "src", "shooting", "downtown", T1)
        fresh_agg.create_incident("inc_older", "shooting", "downtown")
        fresh_agg.add_report_to_incident("inc_older", "ra1")   # latest_ts = T0
        fresh_agg.create_incident("inc_newer", "shooting", "downtown")
        fresh_agg.add_report_to_incident("inc_newer", "ra2")   # latest_ts = T1

        # T2 - T0 = 120s <= 300s and T2 - T1 = 60s <= 300s → both active
        result = fresh_agg.auto_ingest_report(
            "r_pick", "src", "shooting", "downtown", T2, time_window_secs=300
        )
        # inc_newer is more recently active (T1 > T0)
        assert result == "inc_newer"

    def test_auto_generated_id_does_not_clash(self, fresh_agg):
        id1 = fresh_agg.auto_ingest_report(
            "r_id1", "src", "fire", "west", T0, time_window_secs=0
        )
        id2 = fresh_agg.auto_ingest_report(
            "r_id2", "src", "fire", "east", T1, time_window_secs=0
        )
        assert id1 != id2


class TestGetActiveIncidents:
    def test_returns_active_incidents(self, agg):
        # inc-001 latest_ts=T1; T2-T1=60s <= 120s → active
        active = agg.get_active_incidents(T2, time_window_secs=120)
        assert any(i["incident_id"] == "inc-001" for i in active)

    def test_excludes_stale_incidents(self, agg):
        # inc-001 latest_ts=T1; T4-T1=540s > 120s → not active
        active = agg.get_active_incidents(T4, time_window_secs=120)
        assert not any(i["incident_id"] == "inc-001" for i in active)

    def test_excludes_empty_incidents(self, fresh_agg):
        fresh_agg.create_incident("inc_empty_active", "fire", "north")
        active = fresh_agg.get_active_incidents(T0, time_window_secs=300)
        assert not any(i["incident_id"] == "inc_empty_active" for i in active)

    def test_sorted_by_latest_ts_descending(self, fresh_agg):
        fresh_agg.ingest_report("ri1", "src", "fire", "east", T0)
        fresh_agg.ingest_report("ri2", "src", "fire", "west", T2)
        fresh_agg.create_incident("inc_sort_a", "fire", "east")
        fresh_agg.add_report_to_incident("inc_sort_a", "ri1")  # latest_ts=T0
        fresh_agg.create_incident("inc_sort_b", "fire", "west")
        fresh_agg.add_report_to_incident("inc_sort_b", "ri2")  # latest_ts=T2

        active = fresh_agg.get_active_incidents(T3, time_window_secs=600)
        ids = [i["incident_id"] for i in active]
        assert ids.index("inc_sort_b") < ids.index("inc_sort_a")
