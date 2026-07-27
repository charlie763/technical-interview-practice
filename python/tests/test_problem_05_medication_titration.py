"""
Tests for Problem 5: Medication Titration Tracker

Run from the python/ directory:
    pytest tests/test_problem_05_medication_titration.py -v

Or use the test runner from the repo root:
    ./run_tests.sh python/practice_problem_answers/cw_answer_05_medication_titration.py
"""

import pytest
from datetime import date

from practice_problems.problem_05_medication_titration import (
    TitrationTracker,
    TitrationEvent,
    Medication,
)

# ---------------------------------------------------------------------------
# Pre-existing patient event histories
# ---------------------------------------------------------------------------

# Maria: successfully de-escalated off metformin, still on low-dose glipizide
MARIA_EVENTS = [
    TitrationEvent("maria", "metformin",  "start",    500.0, date(2023, 6,  1)),
    TitrationEvent("maria", "metformin",  "increase", 1000.0, date(2023, 8,  1)),
    TitrationEvent("maria", "metformin",  "decrease",  500.0, date(2023, 11, 1)),
    TitrationEvent("maria", "metformin",  "stop",        0.0, date(2024,  2,  1)),
    TitrationEvent("maria", "glipizide",  "start",      5.0, date(2023, 6,  1)),
    TitrationEvent("maria", "glipizide",  "decrease",   2.5, date(2024,  1, 15)),
]

# James: still on two active medications, one recently increased
JAMES_EVENTS = [
    TitrationEvent("james", "metformin",        "start",     500.0, date(2023, 9,  1)),
    TitrationEvent("james", "metformin",        "increase", 1000.0, date(2023, 12, 1)),
    TitrationEvent("james", "insulin_glargine", "start",      10.0, date(2023, 9,  1)),
    TitrationEvent("james", "insulin_glargine", "increase",   15.0, date(2024,  1,  1)),
    TitrationEvent("james", "insulin_glargine", "decrease",   10.0, date(2024,  3,  1)),
]

# Susan: completely off all medications
SUSAN_EVENTS = [
    TitrationEvent("susan", "metformin",  "start",    500.0, date(2023, 3,  1)),
    TitrationEvent("susan", "metformin",  "increase", 750.0, date(2023, 5,  1)),
    TitrationEvent("susan", "metformin",  "stop",       0.0, date(2023, 10, 1)),
    TitrationEvent("susan", "glipizide",  "start",     5.0, date(2023, 3,  1)),
    TitrationEvent("susan", "glipizide",  "stop",       0.0, date(2023, 9,  1)),
]

ALL_EVENTS = MARIA_EVENTS + JAMES_EVENTS + SUSAN_EVENTS


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


@pytest.fixture
def tracker():
    """TitrationTracker seeded with the full pre-existing dataset."""
    return TitrationTracker(ALL_EVENTS)


@pytest.fixture
def fresh_tracker():
    """Empty TitrationTracker."""
    return TitrationTracker([])


# ---------------------------------------------------------------------------
# PART 1 — Current medication snapshot
# ---------------------------------------------------------------------------


class TestCurrentMedications:
    def test_active_medications_returned(self, tracker):
        meds = tracker.current_medications("maria")
        names = {m.name for m in meds}
        assert "glipizide" in names

    def test_stopped_medication_excluded(self, tracker):
        meds = tracker.current_medications("maria")
        names = {m.name for m in meds}
        assert "metformin" not in names

    def test_all_stopped_returns_empty(self, tracker):
        assert tracker.current_medications("susan") == []

    def test_unknown_patient_returns_empty(self, tracker):
        assert tracker.current_medications("nobody") == []

    def test_medication_fields(self, tracker):
        meds = tracker.current_medications("maria")
        glip = next(m for m in meds if m.name == "glipizide")
        assert glip.current_dose == 2.5
        assert glip.last_changed == date(2024, 1, 15)
        assert glip.total_changes == 2  # start + decrease

    def test_multiple_active_medications(self, tracker):
        meds = tracker.current_medications("james")
        names = {m.name for m in meds}
        assert "metformin" in names
        assert "insulin_glargine" in names

    def test_events_out_of_order_handled(self, fresh_tracker):
        """Events should be sorted by date internally — order of insertion must not matter."""
        events = [
            TitrationEvent("pt_x", "metformin", "increase", 1000.0, date(2024, 3, 1)),
            TitrationEvent("pt_x", "metformin", "start",     500.0, date(2024, 1, 1)),
            TitrationEvent("pt_x", "metformin", "decrease",  750.0, date(2024, 2, 1)),
        ]
        t = TitrationTracker(events)
        meds = t.current_medications("pt_x")
        assert len(meds) == 1
        assert meds[0].current_dose == 1000.0

    def test_get_medication_history_sorted(self, tracker):
        history = tracker.get_medication_history("maria", "metformin")
        dates = [e.recorded_on for e in history]
        assert dates == sorted(dates)

    def test_get_medication_history_correct_events(self, tracker):
        history = tracker.get_medication_history("maria", "metformin")
        directions = [e.direction for e in history]
        assert directions == ["start", "increase", "decrease", "stop"]

    def test_get_medication_history_unknown_returns_empty(self, tracker):
        assert tracker.get_medication_history("nobody", "metformin") == []


# ---------------------------------------------------------------------------
# PART 2 — Titration counts
# ---------------------------------------------------------------------------


class TestTitrationCount:
    def test_total_count(self, tracker):
        assert tracker.titration_count("maria", "metformin") == 4

    def test_directional_count_decrease(self, tracker):
        assert tracker.titration_count("maria", "metformin", direction="decrease") == 1

    def test_directional_count_stop(self, tracker):
        assert tracker.titration_count("susan", "metformin", direction="stop") == 1

    def test_directional_count_zero(self, tracker):
        assert tracker.titration_count("james", "metformin", direction="stop") == 0

    def test_unknown_patient_returns_zero(self, tracker):
        assert tracker.titration_count("nobody", "metformin") == 0

    def test_unknown_medication_returns_zero(self, tracker):
        assert tracker.titration_count("maria", "insulin_glargine") == 0


class TestDeEscalationSummary:
    def test_summary_includes_decrease_and_stop(self, tracker):
        summary = tracker.de_escalation_summary("maria")
        # metformin: 1 decrease + 1 stop = 2; glipizide: 1 decrease = 1
        assert summary["metformin"] == 2
        assert summary["glipizide"] == 1

    def test_no_de_escalations_returns_empty(self, fresh_tracker):
        events = [
            TitrationEvent("pt_y", "metformin", "start",    500.0, date(2024, 1, 1)),
            TitrationEvent("pt_y", "metformin", "increase", 1000.0, date(2024, 2, 1)),
        ]
        t = TitrationTracker(events)
        assert t.de_escalation_summary("pt_y") == {}

    def test_only_de_escalated_meds_included(self, tracker):
        """james has increases but also one decrease on insulin_glargine."""
        summary = tracker.de_escalation_summary("james")
        assert "insulin_glargine" in summary
        assert summary["insulin_glargine"] == 1
        # metformin only has start+increase — not included
        assert "metformin" not in summary

    def test_unknown_patient_returns_empty(self, tracker):
        assert tracker.de_escalation_summary("nobody") == {}


# ---------------------------------------------------------------------------
# PART 3 — Population-level queries
# ---------------------------------------------------------------------------


class TestPopulationQueries:
    def test_patients_on_medication_active_only(self, tracker):
        patients = tracker.patients_on_medication("metformin")
        # maria stopped metformin, susan stopped metformin, james still active
        assert patients == ["james"]

    def test_patients_on_medication_sorted(self, tracker):
        # add two patients both on glipizide to verify sorting
        extra = [
            TitrationEvent("zoe",  "glipizide", "start", 5.0, date(2024, 1, 1)),
            TitrationEvent("anna", "glipizide", "start", 5.0, date(2024, 1, 1)),
        ]
        t = TitrationTracker(ALL_EVENTS + extra)
        patients = t.patients_on_medication("glipizide")
        assert patients == sorted(patients)
        assert "maria" in patients
        assert "anna" in patients
        assert "zoe" in patients
        assert "susan" not in patients  # susan stopped glipizide

    def test_patients_on_medication_none_active(self, tracker):
        patients = tracker.patients_on_medication("glipizide_xr")  # unknown med
        assert patients == []

    def test_most_titrated_medications(self, tracker):
        top = tracker.most_titrated_medications(top_n=3)
        names = [name for name, _ in top]
        counts = [count for _, count in top]
        # metformin: maria(4) + james(2) + susan(3) = 9 events total
        assert "metformin" in names
        assert counts == sorted(counts, reverse=True)

    def test_most_titrated_fewer_than_n(self, fresh_tracker):
        events = [
            TitrationEvent("pt_z", "drug_a", "start", 10.0, date(2024, 1, 1)),
        ]
        t = TitrationTracker(events)
        top = t.most_titrated_medications(top_n=5)
        assert len(top) == 1

    def test_most_titrated_returns_correct_count(self, tracker):
        top_dict = {name: count for name, count in tracker.most_titrated_medications(top_n=10)}
        assert top_dict["metformin"] == 9   # 4 + 2 + 3


# ---------------------------------------------------------------------------
# PART 4 — Live ingestion
# ---------------------------------------------------------------------------


class TestAddEvent:
    def test_new_event_updates_current_medications(self, tracker):
        """Adding a stop event removes the medication from current_medications."""
        event = TitrationEvent("james", "metformin", "stop", 0.0, date(2024, 6, 1))
        tracker.add_event(event)
        names = {m.name for m in tracker.current_medications("james")}
        assert "metformin" not in names

    def test_new_event_updates_history(self, tracker):
        event = TitrationEvent("maria", "glipizide", "stop", 0.0, date(2024, 6, 1))
        tracker.add_event(event)
        history = tracker.get_medication_history("maria", "glipizide")
        directions = [e.direction for e in history]
        assert "stop" in directions

    def test_overwrite_same_date(self, tracker):
        """An event on the same (patient, medication, date) overwrites the existing one."""
        event = TitrationEvent("james", "metformin", "stop", 0.0, date(2023, 12, 1))
        tracker.add_event(event)
        history = tracker.get_medication_history("james", "metformin")
        dec_event = next(e for e in history if e.recorded_on == date(2023, 12, 1))
        assert dec_event.direction == "stop"

    def test_add_event_new_patient(self, tracker):
        event = TitrationEvent("new_pt", "metformin", "start", 500.0, date(2024, 5, 1))
        tracker.add_event(event)
        meds = tracker.current_medications("new_pt")
        assert len(meds) == 1
        assert meds[0].name == "metformin"

    def test_add_event_affects_population_query(self, tracker):
        """After adding a new patient on metformin, they appear in patients_on_medication."""
        event = TitrationEvent("new_pt2", "metformin", "start", 500.0, date(2024, 5, 1))
        tracker.add_event(event)
        assert "new_pt2" in tracker.patients_on_medication("metformin")
