"""
Tests for Problem 6: Lab Cadence Compliance Monitor

Run from the python/ directory:
    pytest tests/test_problem_06_lab_cadence.py -v

Or use the test runner from the repo root:
    ./run_tests.sh python/practice_problem_answers/cw_answer_06_lab_cadence.py
"""

import pytest
from datetime import date

from practice_problems.problem_06_lab_cadence import (
    make_monitor,
    register_patient,
    add_required_lab,
    get_required_labs,
    set_lab_deadline,
    record_submission,
    is_overdue,
    overdue_labs,
    compliance_report,
    submission_history,
    days_since_last_submission,
)


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


@pytest.fixture
def m():
    """Bare monitor with no patients."""
    return make_monitor()


@pytest.fixture
def seeded(m):
    """Monitor pre-loaded with a small realistic patient set."""
    register_patient(m, "alice", ["hba1c", "bmp", "lipids"])
    register_patient(m, "bob",   ["hba1c", "bmp"])
    register_patient(m, "carol", ["hba1c"])

    # Alice: hba1c quarterly deadlines; submitted first one, missed second
    set_lab_deadline(m, "alice", "hba1c", date(2024, 3, 31))
    set_lab_deadline(m, "alice", "hba1c", date(2024, 6, 30))
    record_submission(m, "alice", "hba1c", date(2024, 3, 28))

    # Alice: bmp due but no submission
    set_lab_deadline(m, "alice", "bmp", date(2024, 4, 15))

    # Bob: hba1c submitted on time; bmp overdue
    set_lab_deadline(m, "bob", "hba1c", date(2024, 3, 31))
    record_submission(m, "bob", "hba1c", date(2024, 3, 25))
    set_lab_deadline(m, "bob", "bmp", date(2024, 3, 31))

    # Carol: all labs submitted on time
    set_lab_deadline(m, "carol", "hba1c", date(2024, 3, 31))
    record_submission(m, "carol", "hba1c", date(2024, 3, 15))

    return m


# ---------------------------------------------------------------------------
# PART 1 — Patient & lab registration
# ---------------------------------------------------------------------------


class TestRegisterPatient:
    def test_registers_new_patient(self, m):
        register_patient(m, "dave", ["hba1c"])
        assert get_required_labs(m, "dave") == {"hba1c"}

    def test_empty_required_labs_raises(self, m):
        with pytest.raises(ValueError):
            register_patient(m, "eve", [])

    def test_idempotent_for_existing_lab(self, m):
        register_patient(m, "frank", ["hba1c"])
        register_patient(m, "frank", ["hba1c"])
        assert get_required_labs(m, "frank") == {"hba1c"}

    def test_adds_new_labs_to_existing_patient(self, m):
        register_patient(m, "grace", ["hba1c"])
        register_patient(m, "grace", ["bmp"])
        assert get_required_labs(m, "grace") == {"hba1c", "bmp"}

    def test_does_not_remove_existing_labs(self, m):
        register_patient(m, "hank", ["hba1c", "bmp"])
        register_patient(m, "hank", ["lipids"])
        assert "hba1c" in get_required_labs(m, "hank")
        assert "bmp" in get_required_labs(m, "hank")


class TestAddRequiredLab:
    def test_adds_lab_to_existing_patient(self, m):
        register_patient(m, "iris", ["hba1c"])
        add_required_lab(m, "iris", "bmp")
        assert "bmp" in get_required_labs(m, "iris")

    def test_idempotent(self, m):
        register_patient(m, "jack", ["hba1c"])
        add_required_lab(m, "jack", "hba1c")  # already required
        assert get_required_labs(m, "jack") == {"hba1c"}

    def test_unknown_patient_raises(self, m):
        with pytest.raises(KeyError):
            add_required_lab(m, "nobody", "hba1c")


class TestGetRequiredLabs:
    def test_returns_correct_set(self, m):
        register_patient(m, "kate", ["hba1c", "lipids"])
        assert get_required_labs(m, "kate") == {"hba1c", "lipids"}

    def test_unknown_patient_raises(self, m):
        with pytest.raises(KeyError):
            get_required_labs(m, "nobody")


# ---------------------------------------------------------------------------
# PART 2 — Deadlines and submissions
# ---------------------------------------------------------------------------


class TestSetLabDeadline:
    def test_sets_deadline(self, seeded):
        # alice already has deadlines set — just verify no exception
        set_lab_deadline(seeded, "alice", "lipids", date(2024, 5, 1))

    def test_duplicate_deadline_ignored(self, seeded):
        set_lab_deadline(seeded, "alice", "lipids", date(2024, 5, 1))
        set_lab_deadline(seeded, "alice", "lipids", date(2024, 5, 1))  # duplicate
        # is_overdue should work normally — no duplication side effects
        assert is_overdue(seeded, "alice", "lipids", date(2024, 5, 2)) is True

    def test_unknown_patient_raises(self, m):
        with pytest.raises(KeyError):
            set_lab_deadline(m, "nobody", "hba1c", date(2024, 3, 31))

    def test_unrequired_lab_raises(self, m):
        register_patient(m, "leo", ["hba1c"])
        with pytest.raises(ValueError):
            set_lab_deadline(m, "leo", "bmp", date(2024, 3, 31))


class TestRecordSubmission:
    def test_submission_clears_deadline(self, seeded):
        # carol submitted hba1c on Mar 15, deadline was Mar 31 → not overdue
        assert is_overdue(seeded, "carol", "hba1c", date(2024, 4, 1)) is False

    def test_submission_clears_earliest_applicable_deadline(self, seeded):
        # alice: hba1c deadlines Mar 31 (cleared) and Jun 30 (not cleared)
        # as of Jul 1, the Jun 30 deadline should be overdue
        assert is_overdue(seeded, "alice", "hba1c", date(2024, 7, 1)) is True

    def test_unknown_patient_raises(self, m):
        with pytest.raises(KeyError):
            record_submission(m, "nobody", "hba1c", date(2024, 3, 28))

    def test_unrequired_lab_raises(self, m):
        register_patient(m, "mia", ["hba1c"])
        with pytest.raises(ValueError):
            record_submission(m, "mia", "bmp", date(2024, 3, 28))

    def test_late_submission_does_not_clear_past_deadline(self, m):
        """A submission AFTER the due_date should not clear it."""
        register_patient(m, "noah", ["hba1c"])
        set_lab_deadline(m, "noah", "hba1c", date(2024, 3, 31))
        record_submission(m, "noah", "hba1c", date(2024, 4, 15))  # submitted late
        # deadline Mar 31 should still be considered overdue as of Apr 1
        assert is_overdue(m, "noah", "hba1c", date(2024, 4, 1)) is True


class TestIsOverdue:
    def test_overdue_deadline_no_submission(self, seeded):
        # bob: bmp deadline Mar 31, no submission
        assert is_overdue(seeded, "bob", "bmp", date(2024, 4, 1)) is True

    def test_not_overdue_when_submitted(self, seeded):
        # bob: hba1c submitted on time
        assert is_overdue(seeded, "bob", "hba1c", date(2024, 4, 1)) is False

    def test_not_overdue_when_deadline_not_yet_passed(self, seeded):
        # alice: bmp deadline Apr 15; as of Apr 14 not yet overdue
        assert is_overdue(seeded, "alice", "bmp", date(2024, 4, 14)) is False

    def test_overdue_on_exact_day_after_deadline(self, seeded):
        # alice: bmp deadline Apr 15; Apr 16 → overdue
        assert is_overdue(seeded, "alice", "bmp", date(2024, 4, 16)) is True

    def test_not_overdue_when_no_deadline_set(self, seeded):
        # alice: lipids has no deadline set
        assert is_overdue(seeded, "alice", "lipids", date(2024, 6, 1)) is False

    def test_unknown_patient_returns_false(self, seeded):
        assert is_overdue(seeded, "nobody", "hba1c", date(2024, 4, 1)) is False


# ---------------------------------------------------------------------------
# PART 3 — Compliance reporting
# ---------------------------------------------------------------------------


class TestOverdueLabs:
    def test_returns_sorted_overdue_labs(self, seeded):
        # alice as of Jul 1: bmp (missed Apr 15 deadline) and hba1c (missed Jun 30 deadline)
        labs = overdue_labs(seeded, "alice", date(2024, 7, 1))
        assert sorted(labs) == labs
        assert "bmp" in labs
        assert "hba1c" in labs

    def test_no_overdue_returns_empty(self, seeded):
        assert overdue_labs(seeded, "carol", date(2024, 4, 1)) == []

    def test_unknown_patient_returns_empty(self, seeded):
        assert overdue_labs(seeded, "nobody", date(2024, 4, 1)) == []


class TestComplianceReport:
    def test_report_contains_overdue_patients(self, seeded):
        report = compliance_report(seeded, as_of=date(2024, 4, 1))
        patient_ids = [e["patient_id"] for e in report]
        assert "alice" in patient_ids  # bmp overdue
        assert "bob" in patient_ids    # bmp overdue

    def test_compliant_patients_excluded(self, seeded):
        report = compliance_report(seeded, as_of=date(2024, 4, 1))
        patient_ids = [e["patient_id"] for e in report]
        assert "carol" not in patient_ids

    def test_entry_fields(self, seeded):
        report = compliance_report(seeded, as_of=date(2024, 4, 1))
        bob_entry = next(e for e in report if e["patient_id"] == "bob")
        assert set(bob_entry.keys()) == {"patient_id", "overdue_labs", "overdue_count"}
        assert bob_entry["overdue_count"] == len(bob_entry["overdue_labs"])

    def test_sorted_by_overdue_count_descending(self, seeded):
        report = compliance_report(seeded, as_of=date(2024, 7, 1))
        counts = [e["overdue_count"] for e in report]
        assert counts == sorted(counts, reverse=True)

    def test_empty_report_when_no_overdue(self, m):
        register_patient(m, "perfectly_compliant", ["hba1c"])
        set_lab_deadline(m, "perfectly_compliant", "hba1c", date(2024, 3, 31))
        record_submission(m, "perfectly_compliant", "hba1c", date(2024, 3, 20))
        report = compliance_report(m, as_of=date(2024, 4, 1))
        assert report == []


# ---------------------------------------------------------------------------
# PART 4 — Submission history
# ---------------------------------------------------------------------------


class TestSubmissionHistory:
    def test_returns_sorted_dates(self, m):
        register_patient(m, "quinn", ["hba1c"])
        set_lab_deadline(m, "quinn", "hba1c", date(2024, 3, 31))
        set_lab_deadline(m, "quinn", "hba1c", date(2024, 6, 30))
        record_submission(m, "quinn", "hba1c", date(2024, 3, 20))
        record_submission(m, "quinn", "hba1c", date(2024, 6, 15))
        history = submission_history(m, "quinn", "hba1c")
        assert history == sorted(history)
        assert len(history) == 2

    def test_no_submissions_returns_empty(self, seeded):
        # alice has no lipids submissions
        assert submission_history(seeded, "alice", "lipids") == []

    def test_unknown_patient_returns_empty(self, m):
        assert submission_history(m, "nobody", "hba1c") == []


class TestDaysSinceLastSubmission:
    def test_correct_days(self, m):
        register_patient(m, "rita", ["hba1c"])
        set_lab_deadline(m, "rita", "hba1c", date(2024, 3, 31))
        record_submission(m, "rita", "hba1c", date(2024, 3, 20))
        assert days_since_last_submission(m, "rita", "hba1c", date(2024, 4, 20)) == 31

    def test_no_submission_returns_none(self, seeded):
        assert days_since_last_submission(seeded, "alice", "lipids", date(2024, 5, 1)) is None

    def test_unknown_patient_returns_none(self, m):
        assert days_since_last_submission(m, "nobody", "hba1c", date(2024, 4, 1)) is None
