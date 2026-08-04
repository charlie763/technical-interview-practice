"""Tests for Problem 11: Sensor Coverage Tracker

Run from the python/ directory:
    pytest tests/test_problem_11_coverage_tracker.py -v
"""

import pytest
from practice_problems.problem_11_coverage_tracker import CoverageTracker

# ---------------------------------------------------------------------------
# Shared timestamps
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
def fresh_ct():
    """Empty CoverageTracker."""
    return CoverageTracker()


@pytest.fixture
def ct():
    """
    Pre-seeded tracker:
      sta-seed-1  North Tower  downtown  last hb: T0
      sta-seed-2  South Tower  downtown  last hb: T1
      sta-seed-3  East Hub     eastside  (never sent a heartbeat)
    """
    c = CoverageTracker()
    c.register_station("sta-seed-1", "North Tower", region="downtown")
    c.register_station("sta-seed-2", "South Tower", region="downtown")
    c.register_station("sta-seed-3", "East Hub",    region="eastside")
    c.record_heartbeat("sta-seed-1", T0)
    c.record_heartbeat("sta-seed-2", T1)
    # sta-seed-3 intentionally has no heartbeat
    return c


# ---------------------------------------------------------------------------
# PART 1 — Station registration and heartbeats
# ---------------------------------------------------------------------------

class TestRegisterStation:
    def test_stores_and_returns_station(self, fresh_ct):
        s = fresh_ct.register_station("sta_reg_test", "Tower", "north")
        assert s["station_id"] == "sta_reg_test"
        assert s["name"]       == "Tower"
        assert s["region"]     == "north"

    def test_duplicate_raises(self, ct):
        with pytest.raises(ValueError):
            ct.register_station("sta-seed-1", "Dup", "downtown")


class TestRecordHeartbeat:
    def test_updates_last_heartbeat(self, ct):
        ct.record_heartbeat("sta-seed-1", T2)
        assert ct.get_last_heartbeat("sta-seed-1") == T2

    def test_missing_station_raises(self, ct):
        with pytest.raises(KeyError):
            ct.record_heartbeat("ghost", T0)

    def test_same_ts_raises(self, ct):
        # sta-seed-1 last hb is T0; equal ts should be rejected
        with pytest.raises(ValueError):
            ct.record_heartbeat("sta-seed-1", T0)

    def test_earlier_ts_raises(self, ct):
        # sta-seed-2 last hb is T1; T0 < T1 should be rejected
        with pytest.raises(ValueError):
            ct.record_heartbeat("sta-seed-2", T0)

    def test_first_heartbeat_accepted(self, ct):
        # sta-seed-3 has never sent one
        ct.record_heartbeat("sta-seed-3", T0)
        assert ct.get_last_heartbeat("sta-seed-3") == T0


class TestGetLastHeartbeat:
    def test_returns_none_if_never_sent(self, ct):
        assert ct.get_last_heartbeat("sta-seed-3") is None

    def test_returns_latest_ts(self, ct):
        assert ct.get_last_heartbeat("sta-seed-1") == T0

    def test_missing_station_raises(self, ct):
        with pytest.raises(KeyError):
            ct.get_last_heartbeat("ghost")


class TestGetStations:
    def test_no_filter_returns_all(self, ct):
        assert len(ct.get_stations()) == 3

    def test_filter_by_region(self, ct):
        downtown = ct.get_stations(region="downtown")
        assert len(downtown) == 2
        assert all(s["region"] == "downtown" for s in downtown)

    def test_unknown_region_returns_empty(self, ct):
        assert ct.get_stations(region="nowhere") == []

    def test_sorted_by_station_id(self, ct):
        stations = ct.get_stations()
        ids = [s["station_id"] for s in stations]
        assert ids == sorted(ids)


# ---------------------------------------------------------------------------
# PART 2 — Staleness detection and outage tracking
# ---------------------------------------------------------------------------

class TestGetStaleStations:
    def test_never_heartbeat_is_stale(self, ct):
        stale_ids = {s["station_id"] for s in ct.get_stale_stations(T2, stale_after_secs=120)}
        assert "sta-seed-3" in stale_ids

    def test_recent_heartbeat_not_stale(self, ct):
        # sta-seed-2 hb=T1; T2-T1=60s <= 120s → not stale
        stale_ids = {s["station_id"] for s in ct.get_stale_stations(T2, stale_after_secs=120)}
        assert "sta-seed-2" not in stale_ids

    def test_old_heartbeat_is_stale(self, ct):
        # sta-seed-1 hb=T0; T3-T0=300s > 120s → stale
        stale_ids = {s["station_id"] for s in ct.get_stale_stations(T3, stale_after_secs=120)}
        assert "sta-seed-1" in stale_ids

    def test_all_stale_when_threshold_is_tiny(self, ct):
        # 5s threshold: all stations stale at T4
        stale = ct.get_stale_stations(T4, stale_after_secs=5)
        assert len(stale) == 3

    def test_sorted_by_station_id(self, ct):
        stale = ct.get_stale_stations(T4, stale_after_secs=5)
        ids = [s["station_id"] for s in stale]
        assert ids == sorted(ids)


class TestRecordOutageStartEnd:
    def test_open_outage_recorded(self, fresh_ct):
        fresh_ct.register_station("sta_out_test", "T", "north")
        fresh_ct.record_outage_start("sta_out_test", T0)
        outages = fresh_ct.get_outages("sta_out_test")
        assert len(outages) == 1
        assert outages[0]["start_ts"] == T0
        assert outages[0]["end_ts"] is None

    def test_duplicate_open_outage_raises(self, fresh_ct):
        fresh_ct.register_station("sta_dup_out", "T", "north")
        fresh_ct.record_outage_start("sta_dup_out", T0)
        with pytest.raises(ValueError):
            fresh_ct.record_outage_start("sta_dup_out", T1)

    def test_end_closes_outage(self, fresh_ct):
        fresh_ct.register_station("sta_end_out", "T", "north")
        fresh_ct.record_outage_start("sta_end_out", T0)
        fresh_ct.record_outage_end("sta_end_out", T1)
        outages = fresh_ct.get_outages("sta_end_out")
        assert outages[0]["end_ts"] == T1

    def test_end_with_no_open_outage_raises(self, ct):
        # sta-seed-1 has no open outage
        with pytest.raises(ValueError):
            ct.record_outage_end("sta-seed-1", T2)

    def test_start_on_missing_station_raises(self, ct):
        with pytest.raises(KeyError):
            ct.record_outage_start("ghost", T0)

    def test_end_on_missing_station_raises(self, ct):
        with pytest.raises(KeyError):
            ct.record_outage_end("ghost", T0)

    def test_second_outage_allowed_after_first_closed(self, fresh_ct):
        fresh_ct.register_station("sta_2nd_out", "T", "north")
        fresh_ct.record_outage_start("sta_2nd_out", T0)
        fresh_ct.record_outage_end("sta_2nd_out", T1)
        fresh_ct.record_outage_start("sta_2nd_out", T2)   # should not raise
        outages = fresh_ct.get_outages("sta_2nd_out")
        assert len(outages) == 2

    def test_outages_sorted_by_start_ts(self, fresh_ct):
        fresh_ct.register_station("sta_sort_out", "T", "north")
        fresh_ct.record_outage_start("sta_sort_out", T0)
        fresh_ct.record_outage_end("sta_sort_out", T1)
        fresh_ct.record_outage_start("sta_sort_out", T2)
        fresh_ct.record_outage_end("sta_sort_out", T3)
        outages = fresh_ct.get_outages("sta_sort_out")
        assert outages[0]["start_ts"] == T0
        assert outages[1]["start_ts"] == T2

    def test_get_outages_missing_station_raises(self, ct):
        with pytest.raises(KeyError):
            ct.get_outages("ghost")


# ---------------------------------------------------------------------------
# PART 3 — Coverage analysis
# ---------------------------------------------------------------------------

class TestGetRegionCoverage:
    def test_basic_partial_coverage(self, ct):
        # as_of=T2, threshold=90s
        # sta-seed-1: T2-T0=120s > 90s → stale
        # sta-seed-2: T2-T1=60s  ≤ 90s → healthy
        result = ct.get_region_coverage("downtown", T2, stale_after_secs=90)
        assert result["region"]       == "downtown"
        assert result["total"]        == 2
        assert result["healthy"]      == 1
        assert result["stale"]        == 1
        assert result["has_coverage"] is True

    def test_full_coverage(self, ct):
        # as_of=T2, threshold=300s → both hbs are within window
        result = ct.get_region_coverage("downtown", T2, stale_after_secs=300)
        assert result["healthy"]      == 2
        assert result["stale"]        == 0
        assert result["has_coverage"] is True

    def test_no_coverage_all_stale(self, ct):
        # as_of=T4, threshold=30s → both hbs are way older than 30s
        result = ct.get_region_coverage("downtown", T4, stale_after_secs=30)
        assert result["healthy"]      == 0
        assert result["has_coverage"] is False

    def test_empty_region_returns_zeros(self, fresh_ct):
        result = fresh_ct.get_region_coverage("ghost-region", T0, stale_after_secs=60)
        assert result["total"]        == 0
        assert result["healthy"]      == 0
        assert result["has_coverage"] is False

    def test_correct_total_for_region(self, ct):
        result = ct.get_region_coverage("eastside", T0, stale_after_secs=60)
        assert result["total"] == 1


class TestGetOutageSummary:
    def test_no_outages(self, ct):
        summary = ct.get_outage_summary("sta-seed-1", T4)
        assert summary["station_id"]        == "sta-seed-1"
        assert summary["total_outages"]     == 0
        assert summary["open_outage"]       is False
        assert summary["total_outage_secs"] == 0

    def test_closed_outage_duration(self, fresh_ct):
        fresh_ct.register_station("sta_dur_test", "T", "north")
        fresh_ct.record_outage_start("sta_dur_test", T0)
        fresh_ct.record_outage_end("sta_dur_test", T2)      # T2 - T0 = 120s
        summary = fresh_ct.get_outage_summary("sta_dur_test", T3)
        assert summary["total_outages"]     == 1
        assert summary["open_outage"]       is False
        assert summary["total_outage_secs"] == 120

    def test_open_outage_counts_to_as_of(self, fresh_ct):
        fresh_ct.register_station("sta_open_test", "T", "north")
        fresh_ct.record_outage_start("sta_open_test", T0)
        # T3 - T0 = 300s
        summary = fresh_ct.get_outage_summary("sta_open_test", T3)
        assert summary["open_outage"]       is True
        assert summary["total_outage_secs"] == 300

    def test_multiple_closed_outages_cumulative(self, fresh_ct):
        fresh_ct.register_station("sta_cumul_test", "T", "north")
        # Outage 1: T0 → T1 = 60s
        fresh_ct.record_outage_start("sta_cumul_test", T0)
        fresh_ct.record_outage_end("sta_cumul_test", T1)
        # Outage 2: T2 → T3 = 180s
        fresh_ct.record_outage_start("sta_cumul_test", T2)
        fresh_ct.record_outage_end("sta_cumul_test", T3)
        summary = fresh_ct.get_outage_summary("sta_cumul_test", T4)
        assert summary["total_outages"]     == 2
        assert summary["open_outage"]       is False
        assert summary["total_outage_secs"] == 60 + 180   # 240s total

    def test_missing_station_raises(self, ct):
        with pytest.raises(KeyError):
            ct.get_outage_summary("ghost", T0)
