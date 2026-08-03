"""Tests for Problem 10: Responder Dispatch Manager

Run from the python/ directory:
    pytest tests/test_problem_10_dispatch_manager.py -v
"""

import pytest
from practice_problems.problem_10_dispatch_manager import DispatchManager

# ---------------------------------------------------------------------------
# Shared timestamps
# ---------------------------------------------------------------------------
T0 = "2024-06-01T10:00:00"
T1 = "2024-06-01T10:01:00"
T2 = "2024-06-01T10:02:00"


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def fresh_dm():
    """Empty DispatchManager."""
    return DispatchManager()


@pytest.fixture
def dm():
    """
    Pre-seeded manager:
      Responders:
        unit-12  (shooting + robbery, capacity=3): has inc-seed-1 open
        unit-14  (car-crash + fire,   capacity=2): no assignments
      Incidents:
        inc-seed-1  shooting  sev=5  T0  → assigned to unit-12
        inc-seed-2  shooting  sev=3  T1  → unassigned
        inc-seed-3  car-crash sev=4  T2  → unassigned
    """
    d = DispatchManager()
    d.register_responder("unit-12", "Alpha Team",
                         subscribed_types=["shooting", "robbery"], capacity=3)
    d.register_responder("unit-14", "Beta Team",
                         subscribed_types=["car-crash", "fire"], capacity=2)
    d.add_incident("inc-seed-1", "shooting",  severity=5, ts=T0)
    d.add_incident("inc-seed-2", "shooting",  severity=3, ts=T1)
    d.add_incident("inc-seed-3", "car-crash", severity=4, ts=T2)
    d.assign_incident("inc-seed-1", "unit-12")
    return d


# ---------------------------------------------------------------------------
# PART 1 — Registration and basic queries
# ---------------------------------------------------------------------------

class TestRegisterResponder:
    def test_stores_and_returns_responder(self, fresh_dm):
        r = fresh_dm.register_responder("unit_reg_test", "Gamma", ["fire"], 2)
        assert r["responder_id"]     == "unit_reg_test"
        assert r["name"]             == "Gamma"
        assert r["subscribed_types"] == ["fire"]
        assert r["capacity"]         == 2

    def test_duplicate_raises(self, dm):
        with pytest.raises(ValueError):
            dm.register_responder("unit-12", "Duplicate", ["fire"], 1)


class TestAddIncident:
    def test_stores_and_returns_incident(self, fresh_dm):
        inc = fresh_dm.add_incident("inc_add_test", "fire", severity=2, ts=T0)
        assert inc["incident_id"]   == "inc_add_test"
        assert inc["incident_type"] == "fire"
        assert inc["severity"]      == 2
        assert inc["responder_id"]  is None
        assert inc["resolved"]      is False

    def test_duplicate_raises(self, dm):
        with pytest.raises(ValueError):
            dm.add_incident("inc-seed-1", "fire", severity=1, ts=T0)


class TestGetIncidentsForResponder:
    def test_returns_subscribed_types_only(self, dm):
        incidents = dm.get_incidents_for_responder("unit-12")
        assert all(i["incident_type"] in ["shooting", "robbery"] for i in incidents)

    def test_sorted_severity_desc_then_ts_asc(self, dm):
        # inc-seed-1 sev=5, inc-seed-2 sev=3 — both are shooting
        incidents = dm.get_incidents_for_responder("unit-12")
        ids = [i["incident_id"] for i in incidents]
        assert ids[0] == "inc-seed-1"   # higher severity first
        assert ids[1] == "inc-seed-2"

    def test_same_severity_sorted_by_ts(self, fresh_dm):
        fresh_dm.register_responder("u_ts_test", "T", ["fire"], 5)
        fresh_dm.add_incident("inc_ts_early", "fire", severity=3, ts=T0)
        fresh_dm.add_incident("inc_ts_late",  "fire", severity=3, ts=T1)
        incidents = fresh_dm.get_incidents_for_responder("u_ts_test")
        ids = [i["incident_id"] for i in incidents]
        assert ids == ["inc_ts_early", "inc_ts_late"]

    def test_missing_responder_raises(self, dm):
        with pytest.raises(KeyError):
            dm.get_incidents_for_responder("ghost")


# ---------------------------------------------------------------------------
# PART 2 — Assignment and resolution
# ---------------------------------------------------------------------------

class TestAssignIncident:
    def test_sets_responder_id(self, dm):
        dm.assign_incident("inc-seed-2", "unit-12")
        open_ids = {i["incident_id"] for i in dm.get_open_assignments("unit-12")}
        assert "inc-seed-2" in open_ids

    def test_missing_incident_raises(self, dm):
        with pytest.raises(KeyError):
            dm.assign_incident("ghost-inc", "unit-12")

    def test_missing_responder_raises(self, dm):
        with pytest.raises(KeyError):
            dm.assign_incident("inc-seed-2", "ghost-unit")

    def test_already_assigned_raises(self, dm):
        # inc-seed-1 is already assigned to unit-12
        with pytest.raises(ValueError):
            dm.assign_incident("inc-seed-1", "unit-12")

    def test_at_capacity_raises(self, fresh_dm):
        fresh_dm.register_responder("cap_unit", "Cap", ["fire"], capacity=1)
        fresh_dm.add_incident("cap_inc_1", "fire", severity=1, ts=T0)
        fresh_dm.add_incident("cap_inc_2", "fire", severity=1, ts=T1)
        fresh_dm.assign_incident("cap_inc_1", "cap_unit")
        with pytest.raises(ValueError):
            fresh_dm.assign_incident("cap_inc_2", "cap_unit")


class TestResolveIncident:
    def test_marks_resolved_and_removes_from_open(self, dm):
        dm.resolve_incident("inc-seed-1")
        open_ids = {i["incident_id"] for i in dm.get_open_assignments("unit-12")}
        assert "inc-seed-1" not in open_ids

    def test_already_resolved_raises(self, dm):
        dm.resolve_incident("inc-seed-1")
        with pytest.raises(ValueError):
            dm.resolve_incident("inc-seed-1")

    def test_missing_raises(self, dm):
        with pytest.raises(KeyError):
            dm.resolve_incident("ghost")

    def test_frees_capacity_for_next_assignment(self, fresh_dm):
        fresh_dm.register_responder("cap2_unit", "Cap2", ["fire"], capacity=1)
        fresh_dm.add_incident("cap2_inc_1", "fire", severity=1, ts=T0)
        fresh_dm.add_incident("cap2_inc_2", "fire", severity=1, ts=T1)
        fresh_dm.assign_incident("cap2_inc_1", "cap2_unit")
        fresh_dm.resolve_incident("cap2_inc_1")
        # Should no longer raise — capacity was freed
        fresh_dm.assign_incident("cap2_inc_2", "cap2_unit")


class TestGetOpenAssignments:
    def test_returns_open_assignments(self, dm):
        open_list = dm.get_open_assignments("unit-12")
        assert len(open_list) == 1
        assert open_list[0]["incident_id"] == "inc-seed-1"

    def test_resolved_not_in_open(self, dm):
        dm.resolve_incident("inc-seed-1")
        assert dm.get_open_assignments("unit-12") == []

    def test_missing_responder_raises(self, dm):
        with pytest.raises(KeyError):
            dm.get_open_assignments("ghost")

    def test_sorted_severity_desc_then_ts_asc(self, fresh_dm):
        fresh_dm.register_responder("u_sort", "S", ["fire"], 5)
        fresh_dm.add_incident("inc_sort_low",  "fire", severity=2, ts=T0)
        fresh_dm.add_incident("inc_sort_high", "fire", severity=5, ts=T1)
        fresh_dm.assign_incident("inc_sort_low",  "u_sort")
        fresh_dm.assign_incident("inc_sort_high", "u_sort")
        open_list = fresh_dm.get_open_assignments("u_sort")
        assert open_list[0]["incident_id"] == "inc_sort_high"


# ---------------------------------------------------------------------------
# PART 3 — Auto-assignment
# ---------------------------------------------------------------------------

class TestAutoAssign:
    def test_assigns_to_eligible_responder(self, dm):
        # inc-seed-3 is car-crash → only unit-14 is subscribed
        result = dm.auto_assign("inc-seed-3")
        assert result == "unit-14"
        open_ids = {i["incident_id"] for i in dm.get_open_assignments("unit-14")}
        assert "inc-seed-3" in open_ids

    def test_missing_incident_raises(self, dm):
        with pytest.raises(KeyError):
            dm.auto_assign("ghost")

    def test_already_assigned_raises(self, dm):
        with pytest.raises(ValueError):
            dm.auto_assign("inc-seed-1")   # already assigned to unit-12

    def test_no_eligible_responder_raises(self, fresh_dm):
        fresh_dm.register_responder("only_unit", "Only", ["shooting"], 1)
        fresh_dm.add_incident("inc_no_sub", "fire", severity=1, ts=T0)
        with pytest.raises(ValueError):
            fresh_dm.auto_assign("inc_no_sub")

    def test_full_capacity_responder_excluded(self, fresh_dm):
        fresh_dm.register_responder("full_unit", "Full", ["fire"], capacity=1)
        fresh_dm.register_responder("open_unit", "Open", ["fire"], capacity=2)
        fresh_dm.add_incident("inc_cap_fill", "fire", severity=1, ts=T0)
        fresh_dm.add_incident("inc_cap_new",  "fire", severity=1, ts=T1)
        fresh_dm.assign_incident("inc_cap_fill", "full_unit")
        result = fresh_dm.auto_assign("inc_cap_new")
        assert result == "open_unit"

    def test_picks_least_loaded_responder(self, fresh_dm):
        fresh_dm.register_responder("u_loaded", "Loaded", ["fire"], capacity=3)
        fresh_dm.register_responder("u_free",   "Free",   ["fire"], capacity=3)
        fresh_dm.add_incident("inc_load_seed", "fire", severity=1, ts=T0)
        fresh_dm.add_incident("inc_load_new",  "fire", severity=1, ts=T1)
        # Give u_loaded one open incident
        fresh_dm.assign_incident("inc_load_seed", "u_loaded")
        result = fresh_dm.auto_assign("inc_load_new")
        assert result == "u_free"   # fewer open assignments

    def test_tiebreak_by_highest_capacity(self, fresh_dm):
        # Equal open assignments (0 each); tiebreak → higher capacity wins
        fresh_dm.register_responder("u_low_cap",  "Low",  ["fire"], capacity=1)
        fresh_dm.register_responder("u_high_cap", "High", ["fire"], capacity=5)
        fresh_dm.add_incident("inc_cap_tb", "fire", severity=1, ts=T0)
        result = fresh_dm.auto_assign("inc_cap_tb")
        assert result == "u_high_cap"

    def test_calls_assign_incident_internally(self, fresh_dm):
        """auto_assign must delegate to assign_incident (not duplicate logic)."""
        fresh_dm.register_responder("u_delegate", "D", ["fire"], capacity=1)
        fresh_dm.add_incident("inc_delegate_1", "fire", severity=1, ts=T0)
        fresh_dm.add_incident("inc_delegate_2", "fire", severity=1, ts=T1)
        fresh_dm.auto_assign("inc_delegate_1")
        # Capacity should now be full; assigning again must raise via assign_incident
        with pytest.raises(ValueError):
            fresh_dm.auto_assign("inc_delegate_2")


class TestGetDispatchSummary:
    def test_returns_all_responders(self, dm):
        summary = dm.get_dispatch_summary()
        ids = {s["responder_id"] for s in summary}
        assert "unit-12" in ids
        assert "unit-14" in ids

    def test_sorted_by_responder_id(self, dm):
        summary = dm.get_dispatch_summary()
        ids = [s["responder_id"] for s in summary]
        assert ids == sorted(ids)

    def test_open_count_and_available_capacity(self, dm):
        # unit-12 has 1 open assignment; capacity=3
        summary = dm.get_dispatch_summary()
        u12 = next(s for s in summary if s["responder_id"] == "unit-12")
        assert u12["open_count"]         == 1
        assert u12["available_capacity"] == 2

    def test_zero_load_responder(self, dm):
        # unit-14 has no assignments
        summary = dm.get_dispatch_summary()
        u14 = next(s for s in summary if s["responder_id"] == "unit-14")
        assert u14["open_count"]         == 0
        assert u14["available_capacity"] == 2
