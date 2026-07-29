"""Tests for problem_07_care_team_assignments."""
import pytest
from practice_problems.problem_07_care_team_assignments import CareTeamManager, CapacityError


# ── Fixtures ─────────────────────────────────────────────────────────────────

@pytest.fixture
def fresh_mgr():
    """Empty CareTeamManager with no members or assignments."""
    return CareTeamManager()


@pytest.fixture
def mgr():
    """
    Pre-seeded manager:
      - coach_alpha  (coach,     max=3): assigned patient_seed_1, patient_seed_2
      - coach_beta   (coach,     max=1): no patients
      - dr_omega     (physician, max=100): assigned patient_seed_1
    """
    m = CareTeamManager()
    m.add_member("coach_alpha", "coach", max_patients=3)
    m.add_member("coach_beta",  "coach", max_patients=1)
    m.add_member("dr_omega",    "physician", max_patients=100)
    m.assign("patient_seed_1", "coach_alpha", assigned_at=1000.0)
    m.assign("patient_seed_1", "dr_omega",    assigned_at=1000.0)
    m.assign("patient_seed_2", "coach_alpha", assigned_at=2000.0)
    return m


# ── Part 1: Basic assignment and lookup ──────────────────────────────────────

class TestAddMember:
    def test_registered_member_can_receive_assignment(self, fresh_mgr):
        fresh_mgr.add_member("member_reg_test", "coach", max_patients=5)
        # Should not raise — member is known
        fresh_mgr.assign("patient_reg_test", "member_reg_test", assigned_at=0.0)

    def test_unregistered_member_raises_value_error(self, fresh_mgr):
        with pytest.raises(ValueError):
            fresh_mgr.assign("patient_unreg_test", "ghost_member", assigned_at=0.0)


class TestAssign:
    def test_basic_assignment_recorded(self, fresh_mgr):
        fresh_mgr.add_member("coach_basic", "coach", max_patients=5)
        fresh_mgr.assign("patient_basic", "coach_basic", assigned_at=100.0)
        assert fresh_mgr.get_assignment("patient_basic", "coach") == "coach_basic"

    def test_reassignment_replaces_current_member(self, fresh_mgr):
        fresh_mgr.add_member("coach_orig", "coach", max_patients=5)
        fresh_mgr.add_member("coach_new",  "coach", max_patients=5)
        fresh_mgr.assign("patient_reassign", "coach_orig", assigned_at=100.0)
        fresh_mgr.assign("patient_reassign", "coach_new",  assigned_at=200.0)
        assert fresh_mgr.get_assignment("patient_reassign", "coach") == "coach_new"

    def test_reassignment_removes_patient_from_old_member(self, fresh_mgr):
        fresh_mgr.add_member("coach_from", "coach", max_patients=5)
        fresh_mgr.add_member("coach_to",   "coach", max_patients=5)
        fresh_mgr.assign("patient_move", "coach_from", assigned_at=100.0)
        fresh_mgr.assign("patient_move", "coach_to",   assigned_at=200.0)
        assert "patient_move" not in fresh_mgr.get_patients("coach_from")

    def test_assignments_across_roles_are_independent(self, fresh_mgr):
        fresh_mgr.add_member("coach_ind", "coach",     max_patients=5)
        fresh_mgr.add_member("dr_ind",    "physician", max_patients=5)
        fresh_mgr.assign("patient_ind", "coach_ind", assigned_at=100.0)
        fresh_mgr.assign("patient_ind", "dr_ind",    assigned_at=100.0)
        assert fresh_mgr.get_assignment("patient_ind", "coach")     == "coach_ind"
        assert fresh_mgr.get_assignment("patient_ind", "physician") == "dr_ind"


class TestGetAssignment:
    def test_returns_current_member(self, mgr):
        assert mgr.get_assignment("patient_seed_1", "coach") == "coach_alpha"

    def test_returns_none_for_unassigned_role(self, mgr):
        assert mgr.get_assignment("patient_seed_1", "dietitian") is None

    def test_returns_none_for_unknown_patient(self, fresh_mgr):
        assert fresh_mgr.get_assignment("no_such_patient", "coach") is None


class TestGetPatients:
    def test_returns_all_assigned_patients(self, mgr):
        patients = mgr.get_patients("coach_alpha")
        assert "patient_seed_1" in patients
        assert "patient_seed_2" in patients

    def test_result_is_sorted(self, mgr):
        patients = mgr.get_patients("coach_alpha")
        assert patients == sorted(patients)

    def test_returns_empty_list_for_member_with_no_patients(self, mgr):
        assert mgr.get_patients("coach_beta") == []

    def test_unregistered_member_raises_value_error(self, fresh_mgr):
        with pytest.raises(ValueError):
            fresh_mgr.get_patients("nobody")


# ── Part 2: Capacity enforcement ─────────────────────────────────────────────

class TestCapacityEnforcement:
    def test_raises_capacity_error_when_member_is_full(self, fresh_mgr):
        fresh_mgr.add_member("coach_full", "coach", max_patients=1)
        fresh_mgr.assign("patient_cap_1", "coach_full", assigned_at=100.0)
        with pytest.raises(CapacityError):
            fresh_mgr.assign("patient_cap_2", "coach_full", assigned_at=200.0)

    def test_reassigning_existing_patient_to_same_member_does_not_raise(self, fresh_mgr):
        fresh_mgr.add_member("coach_same", "coach", max_patients=1)
        fresh_mgr.assign("patient_same", "coach_same", assigned_at=100.0)
        # Member is "full" but the patient is already theirs — must not raise
        fresh_mgr.assign("patient_same", "coach_same", assigned_at=200.0)

    def test_reassigning_patient_away_frees_capacity(self, fresh_mgr):
        fresh_mgr.add_member("coach_donor", "coach", max_patients=1)
        fresh_mgr.add_member("coach_recv",  "coach", max_patients=5)
        fresh_mgr.assign("patient_freed", "coach_donor", assigned_at=100.0)
        # Move the patient away — coach_donor now has a free slot
        fresh_mgr.assign("patient_freed", "coach_recv", assigned_at=200.0)
        # A new patient should now fit on coach_donor
        fresh_mgr.assign("patient_new", "coach_donor", assigned_at=300.0)
        assert fresh_mgr.get_assignment("patient_new", "coach") == "coach_donor"


class TestAvailableMembers:
    def test_returns_members_with_open_capacity(self, fresh_mgr):
        fresh_mgr.add_member("avail_coach_open", "coach", max_patients=2)
        fresh_mgr.add_member("avail_coach_full", "coach", max_patients=1)
        fresh_mgr.assign("avail_patient_1", "avail_coach_full", assigned_at=100.0)
        available = fresh_mgr.available_members("coach")
        assert "avail_coach_open" in available
        assert "avail_coach_full" not in available

    def test_full_member_is_excluded(self, fresh_mgr):
        fresh_mgr.add_member("only_coach", "coach", max_patients=1)
        fresh_mgr.assign("only_patient", "only_coach", assigned_at=100.0)
        assert "only_coach" not in fresh_mgr.available_members("coach")

    def test_result_is_sorted(self, fresh_mgr):
        fresh_mgr.add_member("sort_coach_z", "coach", max_patients=5)
        fresh_mgr.add_member("sort_coach_a", "coach", max_patients=5)
        result = fresh_mgr.available_members("coach")
        assert result == sorted(result)

    def test_empty_when_no_members_with_role(self, fresh_mgr):
        fresh_mgr.add_member("solo_physician", "physician", max_patients=100)
        assert fresh_mgr.available_members("coach") == []


# ── Part 3: Assignment history ────────────────────────────────────────────────

class TestGetHistory:
    def test_single_assignment_has_open_end(self, fresh_mgr):
        fresh_mgr.add_member("hist_coach_1", "coach", max_patients=5)
        fresh_mgr.assign("hist_patient_1", "hist_coach_1", assigned_at=1000.0)
        assert fresh_mgr.get_history("hist_patient_1", "coach") == [
            ("hist_coach_1", 1000.0, None)
        ]

    def test_reassignment_closes_previous_entry(self, fresh_mgr):
        fresh_mgr.add_member("hist_coach_a", "coach", max_patients=5)
        fresh_mgr.add_member("hist_coach_b", "coach", max_patients=5)
        fresh_mgr.assign("hist_patient_2", "hist_coach_a", assigned_at=1000.0)
        fresh_mgr.assign("hist_patient_2", "hist_coach_b", assigned_at=3000.0)
        assert fresh_mgr.get_history("hist_patient_2", "coach") == [
            ("hist_coach_a", 1000.0, 3000.0),
            ("hist_coach_b", 3000.0, None),
        ]

    def test_multiple_reassignments_sorted_chronologically(self, fresh_mgr):
        for name in ("hist_cx", "hist_cy", "hist_cz"):
            fresh_mgr.add_member(name, "coach", max_patients=5)
        fresh_mgr.assign("hist_patient_3", "hist_cx", assigned_at=100.0)
        fresh_mgr.assign("hist_patient_3", "hist_cy", assigned_at=200.0)
        fresh_mgr.assign("hist_patient_3", "hist_cz", assigned_at=300.0)
        history = fresh_mgr.get_history("hist_patient_3", "coach")
        assert [entry[1] for entry in history] == [100.0, 200.0, 300.0]
        assert history[-1][2] is None

    def test_returns_empty_for_unassigned_role(self, fresh_mgr):
        fresh_mgr.add_member("hist_coach_only", "coach", max_patients=5)
        fresh_mgr.assign("hist_patient_4", "hist_coach_only", assigned_at=100.0)
        assert fresh_mgr.get_history("hist_patient_4", "physician") == []

    def test_returns_empty_for_unknown_patient(self, fresh_mgr):
        assert fresh_mgr.get_history("hist_nobody", "coach") == []


class TestGetAssignmentAt:
    def test_returns_member_during_active_window(self, fresh_mgr):
        fresh_mgr.add_member("at_coach_1", "coach", max_patients=5)
        fresh_mgr.assign("at_patient_1", "at_coach_1", assigned_at=1000.0)
        assert fresh_mgr.get_assignment_at("at_patient_1", "coach", 2000.0) == "at_coach_1"

    def test_returns_none_before_first_assignment(self, fresh_mgr):
        fresh_mgr.add_member("at_coach_2", "coach", max_patients=5)
        fresh_mgr.assign("at_patient_2", "at_coach_2", assigned_at=1000.0)
        assert fresh_mgr.get_assignment_at("at_patient_2", "coach", 500.0) is None

    def test_returns_original_member_before_reassignment(self, fresh_mgr):
        fresh_mgr.add_member("at_coach_old", "coach", max_patients=5)
        fresh_mgr.add_member("at_coach_new", "coach", max_patients=5)
        fresh_mgr.assign("at_patient_3", "at_coach_old", assigned_at=1000.0)
        fresh_mgr.assign("at_patient_3", "at_coach_new", assigned_at=3000.0)
        assert fresh_mgr.get_assignment_at("at_patient_3", "coach", 2000.0) == "at_coach_old"

    def test_returns_new_member_after_reassignment(self, fresh_mgr):
        fresh_mgr.add_member("at_coach_prev", "coach", max_patients=5)
        fresh_mgr.add_member("at_coach_curr", "coach", max_patients=5)
        fresh_mgr.assign("at_patient_4", "at_coach_prev", assigned_at=1000.0)
        fresh_mgr.assign("at_patient_4", "at_coach_curr", assigned_at=3000.0)
        assert fresh_mgr.get_assignment_at("at_patient_4", "coach", 4000.0) == "at_coach_curr"

    def test_returns_none_for_unassigned_role(self, fresh_mgr):
        fresh_mgr.add_member("at_coach_role", "coach", max_patients=5)
        fresh_mgr.assign("at_patient_5", "at_coach_role", assigned_at=1000.0)
        assert fresh_mgr.get_assignment_at("at_patient_5", "physician", 2000.0) is None
