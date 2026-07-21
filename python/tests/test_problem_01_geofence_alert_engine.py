"""
Tests for Problem 1: Geofence Alert Rule Engine

Run from the python/ directory:
    pytest tests/test_problem_01_geofence_alert_engine.py -v
"""

import pytest
from practice_problems.problem_01_geofence_alert_engine import (
    make_tracker,
    is_in_zone,
    get_current_zone_id,
    process_location_update,
    add_zone,
    remove_zone,
    add_asset,
    add_alert_rule,
)


# ---------------------------------------------------------------------------
# Shared fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def state():
    """A tracker with two adjacent zones and two assets."""
    s = make_tracker()
    # Warehouse: lat [35.00, 35.10], lng [-106.70, -106.60]
    add_zone(s, "warehouse", "Warehouse A", 35.00, 35.10, -106.70, -106.60)
    # Loading dock: lat [35.10, 35.20], lng [-106.70, -106.60]
    add_zone(s, "loading_dock", "Loading Dock", 35.10, 35.20, -106.70, -106.60)
    add_asset(s, "forklift_1", "Forklift #1")
    add_asset(s, "drone_1", "Drone #1")
    return s


# ---------------------------------------------------------------------------
# PART 1 — is_in_zone
# ---------------------------------------------------------------------------

class TestIsInZone:
    def _zone(self, min_lat, max_lat, min_lng, max_lng):
        return {
            "id": "z1", "name": "Z",
            "bounds": {"min_lat": min_lat, "max_lat": max_lat,
                       "min_lng": min_lng, "max_lng": max_lng},
        }

    def _asset(self, lat, lng):
        return {"id": "a1", "name": "A", "lat": lat, "lng": lng, "zone_id": None}

    def test_inside(self):
        assert is_in_zone(self._asset(35.05, -106.65), self._zone(35.0, 35.1, -106.7, -106.6)) is True

    def test_on_min_corner(self):
        assert is_in_zone(self._asset(35.0, -106.7), self._zone(35.0, 35.1, -106.7, -106.6)) is True

    def test_on_max_corner(self):
        assert is_in_zone(self._asset(35.1, -106.6), self._zone(35.0, 35.1, -106.7, -106.6)) is True

    def test_outside_lat(self):
        assert is_in_zone(self._asset(35.15, -106.65), self._zone(35.0, 35.1, -106.7, -106.6)) is False

    def test_outside_lng(self):
        assert is_in_zone(self._asset(35.05, -106.5), self._zone(35.0, 35.1, -106.7, -106.6)) is False

    def test_no_lat(self):
        assert is_in_zone(self._asset(None, -106.65), self._zone(35.0, 35.1, -106.7, -106.6)) is False

    def test_no_lng(self):
        assert is_in_zone(self._asset(35.05, None), self._zone(35.0, 35.1, -106.7, -106.6)) is False

    def test_both_none(self):
        assert is_in_zone(self._asset(None, None), self._zone(35.0, 35.1, -106.7, -106.6)) is False


# ---------------------------------------------------------------------------
# PART 2 — get_current_zone_id
# ---------------------------------------------------------------------------

class TestGetCurrentZoneId:
    def test_asset_in_warehouse(self, state):
        state["assets"]["forklift_1"]["lat"] = 35.05
        state["assets"]["forklift_1"]["lng"] = -106.65
        assert get_current_zone_id(state, "forklift_1") == "warehouse"

    def test_asset_in_loading_dock(self, state):
        state["assets"]["forklift_1"]["lat"] = 35.15
        state["assets"]["forklift_1"]["lng"] = -106.65
        assert get_current_zone_id(state, "forklift_1") == "loading_dock"

    def test_asset_outside_all_zones(self, state):
        state["assets"]["forklift_1"]["lat"] = 36.0
        state["assets"]["forklift_1"]["lng"] = -106.65
        assert get_current_zone_id(state, "forklift_1") is None

    def test_asset_no_location(self, state):
        assert get_current_zone_id(state, "forklift_1") is None

    def test_unknown_asset_id(self, state):
        assert get_current_zone_id(state, "ghost") is None


# ---------------------------------------------------------------------------
# PART 3 — process_location_update
# ---------------------------------------------------------------------------

class TestProcessLocationUpdate:
    def test_no_alert_on_first_update_outside_zone(self, state):
        alerts = process_location_update(state, "forklift_1", 36.0, -106.65, "t1")
        assert alerts == []
        assert state["assets"]["forklift_1"]["lat"] == 36.0
        assert state["assets"]["forklift_1"]["zone_id"] is None

    def test_zone_entry_triggers_matching_rule(self, state):
        add_alert_rule(state, "rule_entry", None, "warehouse", None)
        alerts = process_location_update(state, "forklift_1", 35.05, -106.65, "t1")
        assert len(alerts) == 1
        assert alerts[0]["rule_id"] == "rule_entry"
        assert alerts[0]["asset_id"] == "forklift_1"
        assert alerts[0]["from_zone_id"] is None
        assert alerts[0]["to_zone_id"] == "warehouse"
        assert alerts[0]["timestamp"] == "t1"

    def test_zone_exit_triggers_rule(self, state):
        # Pre-position asset in warehouse
        state["assets"]["forklift_1"]["lat"] = 35.05
        state["assets"]["forklift_1"]["lng"] = -106.65
        state["assets"]["forklift_1"]["zone_id"] = "warehouse"
        add_alert_rule(state, "rule_exit", "warehouse", None, None)
        alerts = process_location_update(state, "forklift_1", 36.0, -106.65, "t2")
        assert len(alerts) == 1
        assert alerts[0]["from_zone_id"] == "warehouse"
        assert alerts[0]["to_zone_id"] is None

    def test_zone_to_zone_transition(self, state):
        state["assets"]["forklift_1"]["lat"] = 35.05
        state["assets"]["forklift_1"]["lng"] = -106.65
        state["assets"]["forklift_1"]["zone_id"] = "warehouse"
        add_alert_rule(state, "rule_wh_to_dock", "warehouse", "loading_dock", None)
        alerts = process_location_update(state, "forklift_1", 35.15, -106.65, "t3")
        assert len(alerts) == 1
        assert alerts[0]["from_zone_id"] == "warehouse"
        assert alerts[0]["to_zone_id"] == "loading_dock"

    def test_no_alert_when_zone_unchanged(self, state):
        state["assets"]["forklift_1"]["lat"] = 35.05
        state["assets"]["forklift_1"]["lng"] = -106.65
        state["assets"]["forklift_1"]["zone_id"] = "warehouse"
        add_alert_rule(state, "rule_any", None, None, None)
        alerts = process_location_update(state, "forklift_1", 35.06, -106.65, "t4")
        assert alerts == []

    def test_asset_specific_rule_ignores_other_assets(self, state):
        add_alert_rule(state, "rule_drone_only", None, "warehouse", "drone_1")
        alerts = process_location_update(state, "forklift_1", 35.05, -106.65, "t5")
        assert alerts == []

    def test_asset_specific_rule_fires_for_correct_asset(self, state):
        add_alert_rule(state, "rule_forklift", None, "warehouse", "forklift_1")
        alerts = process_location_update(state, "forklift_1", 35.05, -106.65, "t6")
        assert len(alerts) == 1

    def test_multiple_matching_rules_all_fire(self, state):
        add_alert_rule(state, "rule_a", None, "warehouse", None)
        add_alert_rule(state, "rule_b", None, None, None)
        alerts = process_location_update(state, "forklift_1", 35.05, -106.65, "t7")
        assert len(alerts) == 2

    def test_alerts_appended_to_log(self, state):
        add_alert_rule(state, "rule_1", None, "warehouse", None)
        process_location_update(state, "forklift_1", 35.05, -106.65, "t8")
        assert len(state["alert_log"]) == 1

    def test_unknown_asset_raises_key_error(self, state):
        with pytest.raises(KeyError):
            process_location_update(state, "ghost_asset", 35.05, -106.65, "t9")

    def test_lat_lng_updated_even_when_no_zone_change(self, state):
        state["assets"]["forklift_1"]["lat"] = 35.05
        state["assets"]["forklift_1"]["lng"] = -106.65
        state["assets"]["forklift_1"]["zone_id"] = "warehouse"
        process_location_update(state, "forklift_1", 35.06, -106.64, "t10")
        assert state["assets"]["forklift_1"]["lat"] == 35.06
        assert state["assets"]["forklift_1"]["lng"] == -106.64


# ---------------------------------------------------------------------------
# PART 4 — CRUD helpers
# ---------------------------------------------------------------------------

class TestAddZone:
    def test_adds_zone(self, state):
        z = add_zone(state, "yard", "Yard", 35.3, 35.4, -106.7, -106.6)
        assert state["zones"]["yard"] == z
        assert z["name"] == "Yard"

    def test_duplicate_raises_value_error(self, state):
        with pytest.raises(ValueError):
            add_zone(state, "warehouse", "Duplicate", 0, 1, 0, 1)


class TestRemoveZone:
    def test_removes_zone(self, state):
        remove_zone(state, "warehouse")
        assert "warehouse" not in state["zones"]

    def test_clears_zone_id_on_assets(self, state):
        state["assets"]["forklift_1"]["zone_id"] = "warehouse"
        remove_zone(state, "warehouse")
        assert state["assets"]["forklift_1"]["zone_id"] is None

    def test_does_not_affect_assets_in_other_zones(self, state):
        state["assets"]["forklift_1"]["zone_id"] = "loading_dock"
        remove_zone(state, "warehouse")
        assert state["assets"]["forklift_1"]["zone_id"] == "loading_dock"

    def test_missing_zone_raises_key_error(self, state):
        with pytest.raises(KeyError):
            remove_zone(state, "nonexistent")


class TestAddAsset:
    def test_adds_asset(self, state):
        a = add_asset(state, "scanner_1", "Scanner #1")
        assert state["assets"]["scanner_1"] == a
        assert a["lat"] is None
        assert a["lng"] is None
        assert a["zone_id"] is None

    def test_duplicate_raises_value_error(self, state):
        with pytest.raises(ValueError):
            add_asset(state, "forklift_1", "Duplicate")


class TestAddAlertRule:
    def test_adds_rule(self, state):
        r = add_alert_rule(state, "r1", "warehouse", "loading_dock", None)
        assert r in state["alert_rules"]
        assert r["id"] == "r1"

    def test_duplicate_raises_value_error(self, state):
        add_alert_rule(state, "r1", None, None, None)
        with pytest.raises(ValueError):
            add_alert_rule(state, "r1", None, None, None)
