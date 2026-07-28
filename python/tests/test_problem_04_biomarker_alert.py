"""
Tests for Problem 4: Biomarker Alert Monitor

Run from the python/ directory:
    pytest tests/test_problem_04_biomarker_alert.py -v

Or use the test runner from the repo root:
    ./run_tests.sh python/practice_problem_answers/cw_answer_04_biomarker_alert.py
"""

import pytest
from datetime import date

# from practice_problems.problem_04_biomarker_alert import BiomarkerMonitor, BiomarkerReading
from python.practice_problem_answers.cw_answer_04_biomarker_alert import (
    BiomarkerMonitor,
    BiomarkerReading,
)

# ---------------------------------------------------------------------------
# Pre-existing data — represents a snapshot of patient readings already in the
# system when the monitor is initialized. Tests below build on this dataset.
# ---------------------------------------------------------------------------

ALICE_READINGS = [
    # All in range — no outreach needed
    BiomarkerReading("alice", "glucose", 105.0, date(2024, 1, 1)),
    BiomarkerReading("alice", "glucose", 98.0, date(2024, 1, 2)),
    BiomarkerReading("alice", "glucose", 112.0, date(2024, 1, 3)),
    BiomarkerReading("alice", "glucose", 91.0, date(2024, 1, 4)),
]

BOB_READINGS = [
    # 4 consecutive high-glucose days → needs outreach
    BiomarkerReading("bob", "glucose", 195.0, date(2024, 1, 1)),
    BiomarkerReading("bob", "glucose", 210.0, date(2024, 1, 2)),
    BiomarkerReading("bob", "glucose", 188.0, date(2024, 1, 3)),
    BiomarkerReading("bob", "glucose", 202.0, date(2024, 1, 4)),
]

CAROL_READINGS = [
    # Streak of 1 (Jan 1), in-range on Jan 2, then streak of 2 (Jan 3–4) → max=2
    BiomarkerReading("carol", "glucose", 190.0, date(2024, 1, 1)),
    BiomarkerReading("carol", "glucose", 150.0, date(2024, 1, 2)),  # in range
    BiomarkerReading("carol", "glucose", 185.0, date(2024, 1, 3)),
    BiomarkerReading("carol", "glucose", 191.0, date(2024, 1, 4)),
]

DAVE_READINGS = [
    # 2 consecutive high-glucose days — below default threshold of 3
    BiomarkerReading("dave", "glucose", 199.0, date(2024, 1, 3)),
    BiomarkerReading("dave", "glucose", 205.0, date(2024, 1, 4)),
]

EVE_READINGS = [
    # One dangerous low glucose (below 70) on Jan 1 — streak of 1 for glucose
    BiomarkerReading("eve", "glucose", 62.0, date(2024, 1, 1)),
    # 3 consecutive days with ketones below target (< 0.5) → ketone outreach
    BiomarkerReading("eve", "ketone", 0.3, date(2024, 1, 2)),
    BiomarkerReading("eve", "ketone", 0.2, date(2024, 1, 3)),
    BiomarkerReading("eve", "ketone", 0.4, date(2024, 1, 4)),
    # Weight readings — never out of range
    BiomarkerReading("eve", "weight", 165.0, date(2024, 1, 1)),
    BiomarkerReading("eve", "weight", 164.5, date(2024, 1, 2)),
]

ALL_READINGS = (
    ALICE_READINGS + BOB_READINGS + CAROL_READINGS + DAVE_READINGS + EVE_READINGS
)


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


@pytest.fixture
def monitor():
    """BiomarkerMonitor seeded with the full pre-existing dataset."""
    return BiomarkerMonitor(ALL_READINGS)


@pytest.fixture
def fresh_monitor():
    """Empty BiomarkerMonitor."""
    return BiomarkerMonitor([])


# ---------------------------------------------------------------------------
# PART 1 — Single-reading classification
# ---------------------------------------------------------------------------


class TestIsOutOfRange:
    def test_glucose_above_range(self, fresh_monitor):
        r = BiomarkerReading("p1", "glucose", 181.0, date(2024, 1, 1))
        assert fresh_monitor.is_out_of_range(r) is True

    def test_glucose_below_range(self, fresh_monitor):
        r = BiomarkerReading("p1", "glucose", 69.9, date(2024, 1, 1))
        assert fresh_monitor.is_out_of_range(r) is True

    def test_glucose_at_upper_boundary(self, fresh_monitor):
        r = BiomarkerReading("p1", "glucose", 180.0, date(2024, 1, 1))
        assert fresh_monitor.is_out_of_range(r) is False

    def test_glucose_at_lower_boundary(self, fresh_monitor):
        r = BiomarkerReading("p1", "glucose", 70.0, date(2024, 1, 1))
        assert fresh_monitor.is_out_of_range(r) is False

    def test_glucose_in_range(self, fresh_monitor):
        r = BiomarkerReading("p1", "glucose", 120.0, date(2024, 1, 1))
        assert fresh_monitor.is_out_of_range(r) is False

    def test_ketone_above_range(self, fresh_monitor):
        r = BiomarkerReading("p1", "ketone", 3.1, date(2024, 1, 1))
        assert fresh_monitor.is_out_of_range(r) is True

    def test_ketone_below_range(self, fresh_monitor):
        r = BiomarkerReading("p1", "ketone", 0.4, date(2024, 1, 1))
        assert fresh_monitor.is_out_of_range(r) is True

    def test_ketone_in_range(self, fresh_monitor):
        r = BiomarkerReading("p1", "ketone", 1.5, date(2024, 1, 1))
        assert fresh_monitor.is_out_of_range(r) is False

    def test_ketone_at_boundaries(self, fresh_monitor):
        lo = BiomarkerReading("p1", "ketone", 0.5, date(2024, 1, 1))
        hi = BiomarkerReading("p1", "ketone", 3.0, date(2024, 1, 1))
        assert fresh_monitor.is_out_of_range(lo) is False
        assert fresh_monitor.is_out_of_range(hi) is False

    def test_weight_never_out_of_range(self, fresh_monitor):
        r = BiomarkerReading("p1", "weight", 9999.0, date(2024, 1, 1))
        assert fresh_monitor.is_out_of_range(r) is False

    def test_classification_uses_preloaded_data(self, monitor):
        """is_out_of_range works on any BiomarkerReading regardless of monitor state."""
        in_range = BiomarkerReading("alice", "glucose", 100.0, date(2024, 1, 5))
        out_of_range = BiomarkerReading("bob", "glucose", 250.0, date(2024, 1, 5))
        assert monitor.is_out_of_range(in_range) is False
        assert monitor.is_out_of_range(out_of_range) is True


# ---------------------------------------------------------------------------
# PART 2 — Streak detection
# ---------------------------------------------------------------------------


class TestMaxConsecutiveOutOfRangeDays:
    def test_no_readings_returns_zero(self, monitor):
        assert (
            monitor.max_consecutive_out_of_range_days("unknown_patient", "glucose") == 0
        )

    def test_all_in_range_returns_zero(self, monitor):
        assert monitor.max_consecutive_out_of_range_days("alice", "glucose") == 0

    def test_consecutive_streak(self, monitor):
        assert monitor.max_consecutive_out_of_range_days("bob", "glucose") == 4

    def test_streak_broken_by_in_range_day(self, monitor):
        # carol: streak of 1, break, streak of 2 → max = 2
        assert monitor.max_consecutive_out_of_range_days("carol", "glucose") == 2

    def test_short_streak(self, monitor):
        assert monitor.max_consecutive_out_of_range_days("dave", "glucose") == 2

    def test_glucose_streak_of_one_for_low(self, monitor):
        assert monitor.max_consecutive_out_of_range_days("eve", "glucose") == 1

    def test_ketone_streak(self, monitor):
        assert monitor.max_consecutive_out_of_range_days("eve", "ketone") == 3

    def test_weight_never_streaks(self, monitor):
        assert monitor.max_consecutive_out_of_range_days("eve", "weight") == 0

    def test_multiple_readings_same_day_count_as_one(self, fresh_monitor):
        """Two out-of-range readings on the same day = 1 day in the streak."""
        readings = [
            BiomarkerReading("frank", "glucose", 200.0, date(2024, 2, 1)),
            BiomarkerReading("frank", "glucose", 210.0, date(2024, 2, 1)),  # same day
            BiomarkerReading("frank", "glucose", 195.0, date(2024, 2, 2)),
        ]
        m = BiomarkerMonitor(readings)
        assert m.max_consecutive_out_of_range_days("frank", "glucose") == 2

    def test_gap_breaks_streak(self, fresh_monitor):
        """Jan 1 and Jan 3 (no Jan 2) are two separate streaks of 1."""
        readings = [
            BiomarkerReading("grace", "glucose", 200.0, date(2024, 3, 1)),
            BiomarkerReading("grace", "glucose", 200.0, date(2024, 3, 3)),  # skip Mar 2
        ]
        m = BiomarkerMonitor(readings)
        assert m.max_consecutive_out_of_range_days("grace", "glucose") == 1

    def test_wrong_reading_type_ignored(self, monitor):
        """Ketone readings don't count toward glucose streak and vice versa."""
        assert monitor.max_consecutive_out_of_range_days("eve", "glucose") == 1
        assert monitor.max_consecutive_out_of_range_days("eve", "ketone") == 3


# ---------------------------------------------------------------------------
# PART 3 — Outreach list
# ---------------------------------------------------------------------------


class TestGetOutreachList:
    def test_default_threshold_filters_correctly(self, monitor):
        """With min_consecutive_days=3 (default), only bob and eve (ketone) qualify."""
        result = monitor.get_outreach_list()
        patient_types = {(e["patient_id"], e["reading_type"]) for e in result}
        assert ("bob", "glucose") in patient_types
        assert ("eve", "ketone") in patient_types
        # alice, carol, dave don't qualify at threshold 3
        assert ("alice", "glucose") not in patient_types
        assert ("carol", "glucose") not in patient_types
        assert ("dave", "glucose") not in patient_types

    def test_weight_never_in_outreach(self, monitor):
        result = monitor.get_outreach_list(min_consecutive_days=1)
        for entry in result:
            assert entry["reading_type"] != "weight"

    def test_result_sorted_by_consecutive_days_descending(self, monitor):
        result = monitor.get_outreach_list(min_consecutive_days=1)
        days = [e["consecutive_days"] for e in result]
        assert days == sorted(days, reverse=True)

    def test_entry_fields(self, monitor):
        result = monitor.get_outreach_list()
        bob_entry = next(e for e in result if e["patient_id"] == "bob")
        assert set(bob_entry.keys()) == {
            "patient_id",
            "reading_type",
            "consecutive_days",
            "latest_value",
        }
        assert bob_entry["consecutive_days"] == 4
        assert bob_entry["reading_type"] == "glucose"
        assert isinstance(bob_entry["latest_value"], float)

    def test_latest_value_is_most_recent_out_of_range(self, monitor):
        result = monitor.get_outreach_list()
        bob_entry = next(e for e in result if e["patient_id"] == "bob")
        # Bob's most recent OOR reading is Jan 4: 202.0
        assert bob_entry["latest_value"] == 202.0

    def test_patient_can_appear_twice_for_different_types(self, fresh_monitor):
        readings = [
            # 3-day glucose streak
            BiomarkerReading("hank", "glucose", 200.0, date(2024, 1, 1)),
            BiomarkerReading("hank", "glucose", 210.0, date(2024, 1, 2)),
            BiomarkerReading("hank", "glucose", 195.0, date(2024, 1, 3)),
            # 3-day ketone streak
            BiomarkerReading("hank", "ketone", 0.2, date(2024, 1, 1)),
            BiomarkerReading("hank", "ketone", 0.3, date(2024, 1, 2)),
            BiomarkerReading("hank", "ketone", 0.1, date(2024, 1, 3)),
        ]
        m = BiomarkerMonitor(readings)
        result = m.get_outreach_list(min_consecutive_days=3)
        patient_types = [(e["patient_id"], e["reading_type"]) for e in result]
        assert ("hank", "glucose") in patient_types
        assert ("hank", "ketone") in patient_types

    def test_custom_threshold(self, monitor):
        result_2 = monitor.get_outreach_list(min_consecutive_days=2)
        patient_types_2 = {(e["patient_id"], e["reading_type"]) for e in result_2}
        assert ("carol", "glucose") in patient_types_2
        assert ("dave", "glucose") in patient_types_2

    def test_empty_monitor_returns_empty(self, fresh_monitor):
        assert fresh_monitor.get_outreach_list() == []


# ---------------------------------------------------------------------------
# PART 4 — Deduplication on ingestion
# ---------------------------------------------------------------------------


class TestAddReading:
    def test_new_reading_returns_true(self, monitor):
        r = BiomarkerReading("alice", "glucose", 100.0, date(2024, 1, 10))
        assert monitor.add_reading(r) is True

    def test_duplicate_exact_same_returns_false(self, monitor):
        r = BiomarkerReading("alice", "glucose", 105.0, date(2024, 1, 1))
        assert monitor.add_reading(r) is False

    def test_duplicate_within_tolerance_returns_false(self, monitor):
        # alice jan 1 = 105.0; 105.4 is within ±0.5
        r = BiomarkerReading("alice", "glucose", 105.4, date(2024, 1, 1))
        assert monitor.add_reading(r) is False

    def test_just_outside_tolerance_returns_true(self, monitor):
        # alice jan 1 = 105.0; 105.6 is outside ±0.5
        r = BiomarkerReading("alice", "glucose", 105.6, date(2024, 1, 1))
        assert monitor.add_reading(r) is True

    def test_different_date_not_duplicate(self, monitor):
        r = BiomarkerReading("alice", "glucose", 105.0, date(2024, 1, 5))
        assert monitor.add_reading(r) is True

    def test_different_type_not_duplicate(self, monitor):
        r = BiomarkerReading("alice", "ketone", 1.0, date(2024, 1, 1))
        assert monitor.add_reading(r) is True

    def test_different_patient_not_duplicate(self, monitor):
        r = BiomarkerReading("frank", "glucose", 105.0, date(2024, 1, 1))
        assert monitor.add_reading(r) is True

    def test_added_reading_affects_streak(self, monitor):
        """After adding a new out-of-range day, the streak should update."""
        # bob's current streak is Jan 1-4 (4 days). Add Jan 5.
        r = BiomarkerReading("bob", "glucose", 195.0, date(2024, 1, 5))
        monitor.add_reading(r)
        assert monitor.max_consecutive_out_of_range_days("bob", "glucose") == 5

    def test_adding_duplicate_does_not_affect_streak(self, monitor):
        """A rejected duplicate must not inflate the streak."""
        r = BiomarkerReading("bob", "glucose", 195.0, date(2024, 1, 1))
        monitor.add_reading(r)  # duplicate, rejected
        assert monitor.max_consecutive_out_of_range_days("bob", "glucose") == 4
