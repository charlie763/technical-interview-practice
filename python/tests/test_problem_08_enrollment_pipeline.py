"""Tests for problem_08_enrollment_pipeline."""
import pytest
from practice_problems.problem_08_enrollment_pipeline import EnrollmentPipeline


# ── Fixtures ─────────────────────────────────────────────────────────────────

@pytest.fixture
def fresh_pipeline():
    """Empty EnrollmentPipeline with no patients."""
    return EnrollmentPipeline()


@pytest.fixture
def pipeline():
    """
    Pre-seeded pipeline with patients in various states (all timestamps in seconds):
      seed_active:      referred(0) → screened(100) → enrolled(200) → active(300)
      seed_graduated:   referred(0) → screened(50)  → enrolled(150) → active(250) → graduated(1000)
      seed_ineligible:  referred(0) → screened(10)  → ineligible(20)
      seed_referred:    referred(0)  [no further transitions]
    """
    p = EnrollmentPipeline()

    p.add_patient("seed_active", timestamp=0.0)
    p.transition("seed_active", "screened",  timestamp=100.0)
    p.transition("seed_active", "enrolled",  timestamp=200.0)
    p.transition("seed_active", "active",    timestamp=300.0)

    p.add_patient("seed_graduated", timestamp=0.0)
    p.transition("seed_graduated", "screened",  timestamp=50.0)
    p.transition("seed_graduated", "enrolled",  timestamp=150.0)
    p.transition("seed_graduated", "active",    timestamp=250.0)
    p.transition("seed_graduated", "graduated", timestamp=1000.0)

    p.add_patient("seed_ineligible", timestamp=0.0)
    p.transition("seed_ineligible", "screened",   timestamp=10.0)
    p.transition("seed_ineligible", "ineligible", timestamp=20.0)

    p.add_patient("seed_referred", timestamp=0.0)
    return p


# ── Part 1: State tracking ────────────────────────────────────────────────────

class TestAddPatient:
    def test_new_patient_starts_in_referred(self, fresh_pipeline):
        fresh_pipeline.add_patient("add_p1", timestamp=0.0)
        assert fresh_pipeline.get_state("add_p1") == "referred"

    def test_duplicate_patient_id_raises(self, fresh_pipeline):
        fresh_pipeline.add_patient("dup_p", timestamp=0.0)
        with pytest.raises(ValueError):
            fresh_pipeline.add_patient("dup_p", timestamp=10.0)


class TestTransition:
    def test_valid_transition_changes_state(self, fresh_pipeline):
        fresh_pipeline.add_patient("trans_p1", timestamp=0.0)
        fresh_pipeline.transition("trans_p1", "screened", timestamp=100.0)
        assert fresh_pipeline.get_state("trans_p1") == "screened"

    def test_invalid_transition_raises(self, fresh_pipeline):
        fresh_pipeline.add_patient("trans_p2", timestamp=0.0)
        with pytest.raises(ValueError):
            # Cannot jump from referred straight to graduated
            fresh_pipeline.transition("trans_p2", "graduated", timestamp=100.0)

    def test_skipping_states_raises(self, fresh_pipeline):
        fresh_pipeline.add_patient("trans_p3", timestamp=0.0)
        fresh_pipeline.transition("trans_p3", "screened", timestamp=10.0)
        with pytest.raises(ValueError):
            # Cannot skip enrolled → jump from screened to active
            fresh_pipeline.transition("trans_p3", "active", timestamp=20.0)

    def test_unknown_patient_raises(self, fresh_pipeline):
        with pytest.raises(ValueError):
            fresh_pipeline.transition("ghost", "screened", timestamp=100.0)

    def test_transition_from_terminal_state_raises(self, fresh_pipeline):
        fresh_pipeline.add_patient("trans_p4", timestamp=0.0)
        fresh_pipeline.transition("trans_p4", "screened",   timestamp=10.0)
        fresh_pipeline.transition("trans_p4", "ineligible", timestamp=20.0)
        with pytest.raises(ValueError):
            fresh_pipeline.transition("trans_p4", "enrolled", timestamp=30.0)

    def test_both_branches_from_screened_are_valid(self, fresh_pipeline):
        fresh_pipeline.add_patient("branch_p1", timestamp=0.0)
        fresh_pipeline.transition("branch_p1", "screened",  timestamp=10.0)
        fresh_pipeline.transition("branch_p1", "enrolled",  timestamp=20.0)
        assert fresh_pipeline.get_state("branch_p1") == "enrolled"

        fresh_pipeline.add_patient("branch_p2", timestamp=0.0)
        fresh_pipeline.transition("branch_p2", "screened",   timestamp=10.0)
        fresh_pipeline.transition("branch_p2", "ineligible", timestamp=20.0)
        assert fresh_pipeline.get_state("branch_p2") == "ineligible"


class TestGetState:
    def test_returns_current_state_for_each_patient(self, pipeline):
        assert pipeline.get_state("seed_active")     == "active"
        assert pipeline.get_state("seed_graduated")  == "graduated"
        assert pipeline.get_state("seed_ineligible") == "ineligible"
        assert pipeline.get_state("seed_referred")   == "referred"

    def test_unknown_patient_raises(self, fresh_pipeline):
        with pytest.raises(ValueError):
            fresh_pipeline.get_state("nobody")


class TestGetPatientsInState:
    def test_returns_patients_in_state(self, pipeline):
        active_patients = pipeline.get_patients_in_state("active")
        assert "seed_active" in active_patients
        assert "seed_graduated" not in active_patients

    def test_result_is_sorted(self, pipeline):
        result = pipeline.get_patients_in_state("referred")
        assert result == sorted(result)

    def test_returns_empty_for_unpopulated_state(self, pipeline):
        assert pipeline.get_patients_in_state("withdrawn") == []

    def test_patient_absent_after_leaving_state(self, pipeline):
        # seed_ineligible passed through screened but is no longer there
        assert "seed_ineligible" not in pipeline.get_patients_in_state("screened")


# ── Part 2: Duration and conversion metrics ───────────────────────────────────

class TestTimeInState:
    def test_completed_state_returns_exact_duration(self, fresh_pipeline):
        fresh_pipeline.add_patient("dur_p1", timestamp=0.0)
        fresh_pipeline.transition("dur_p1", "screened", timestamp=1000.0)
        fresh_pipeline.transition("dur_p1", "enrolled", timestamp=4000.0)
        # Spent exactly 3000 s in screened; as_of is ignored for completed states
        assert fresh_pipeline.time_in_state("dur_p1", "screened", as_of=99999.0) == 3000.0

    def test_current_state_counts_up_to_as_of(self, fresh_pipeline):
        fresh_pipeline.add_patient("dur_p2", timestamp=0.0)
        fresh_pipeline.transition("dur_p2", "screened", timestamp=1000.0)
        assert fresh_pipeline.time_in_state("dur_p2", "screened", as_of=4000.0) == 3000.0

    def test_returns_zero_for_state_never_visited(self, fresh_pipeline):
        fresh_pipeline.add_patient("dur_p3", timestamp=0.0)
        assert fresh_pipeline.time_in_state("dur_p3", "enrolled", as_of=99999.0) == 0.0

    def test_initial_referred_state_timed_from_add_patient_timestamp(self, fresh_pipeline):
        fresh_pipeline.add_patient("dur_p4", timestamp=500.0)
        fresh_pipeline.transition("dur_p4", "screened", timestamp=1500.0)
        # Spent 1000 s in referred (1500 - 500)
        assert fresh_pipeline.time_in_state("dur_p4", "referred", as_of=99999.0) == 1000.0


class TestConversionRate:
    def test_fifty_percent_conversion(self, fresh_pipeline):
        fresh_pipeline.add_patient("conv_p1", timestamp=0.0)
        fresh_pipeline.transition("conv_p1", "screened",  timestamp=10.0)
        fresh_pipeline.transition("conv_p1", "enrolled",  timestamp=20.0)

        fresh_pipeline.add_patient("conv_p2", timestamp=0.0)
        fresh_pipeline.transition("conv_p2", "screened",   timestamp=10.0)
        fresh_pipeline.transition("conv_p2", "ineligible", timestamp=20.0)

        assert fresh_pipeline.conversion_rate("screened", "enrolled") == pytest.approx(0.5)

    def test_excludes_patients_still_in_from_state(self, fresh_pipeline):
        fresh_pipeline.add_patient("conv_p3", timestamp=0.0)
        fresh_pipeline.transition("conv_p3", "screened", timestamp=10.0)
        # conv_p3 is still in screened — must not be counted

        fresh_pipeline.add_patient("conv_p4", timestamp=0.0)
        fresh_pipeline.transition("conv_p4", "screened",  timestamp=10.0)
        fresh_pipeline.transition("conv_p4", "enrolled",  timestamp=20.0)

        # Only conv_p4 has exited; they enrolled → rate is 1.0
        assert fresh_pipeline.conversion_rate("screened", "enrolled") == pytest.approx(1.0)

    def test_returns_zero_when_no_one_has_exited_from_state(self, fresh_pipeline):
        fresh_pipeline.add_patient("conv_p5", timestamp=0.0)
        # conv_p5 is still in referred
        assert fresh_pipeline.conversion_rate("referred", "screened") == pytest.approx(0.0)

    def test_hundred_percent_when_all_converted(self, fresh_pipeline):
        for pid in ("conv_all_1", "conv_all_2"):
            fresh_pipeline.add_patient(pid, timestamp=0.0)
            fresh_pipeline.transition(pid, "screened", timestamp=10.0)
            fresh_pipeline.transition(pid, "enrolled", timestamp=20.0)
        assert fresh_pipeline.conversion_rate("screened", "enrolled") == pytest.approx(1.0)


# ── Part 3: SLA monitoring ────────────────────────────────────────────────────

class TestPatientsOverdue:
    def test_returns_patients_exceeding_threshold(self, fresh_pipeline):
        fresh_pipeline.add_patient("over_p1", timestamp=0.0)
        fresh_pipeline.transition("over_p1", "screened", timestamp=0.0)
        fresh_pipeline.transition("over_p1", "enrolled", timestamp=0.0)
        fresh_pipeline.transition("over_p1", "active",   timestamp=0.0)   # 10000 s in active

        fresh_pipeline.add_patient("over_p2", timestamp=0.0)
        fresh_pipeline.transition("over_p2", "screened", timestamp=0.0)
        fresh_pipeline.transition("over_p2", "enrolled", timestamp=0.0)
        fresh_pipeline.transition("over_p2", "active",   timestamp=5000.0)  # 5000 s in active

        overdue = fresh_pipeline.patients_overdue("active", max_seconds=6000.0, as_of=10000.0)
        assert "over_p1" in overdue
        assert "over_p2" not in overdue

    def test_sorted_by_duration_descending(self, fresh_pipeline):
        fresh_pipeline.add_patient("sort_p1", timestamp=0.0)
        fresh_pipeline.transition("sort_p1", "screened", timestamp=0.0)
        fresh_pipeline.transition("sort_p1", "enrolled", timestamp=0.0)
        fresh_pipeline.transition("sort_p1", "active",   timestamp=0.0)    # 10000 s

        fresh_pipeline.add_patient("sort_p2", timestamp=0.0)
        fresh_pipeline.transition("sort_p2", "screened", timestamp=0.0)
        fresh_pipeline.transition("sort_p2", "enrolled", timestamp=0.0)
        fresh_pipeline.transition("sort_p2", "active",   timestamp=3000.0) # 7000 s

        overdue = fresh_pipeline.patients_overdue("active", max_seconds=5000.0, as_of=10000.0)
        assert overdue == ["sort_p1", "sort_p2"]

    def test_excludes_patients_not_currently_in_state(self, pipeline):
        # seed_graduated has left active — must not appear in active overdue list
        overdue = pipeline.patients_overdue("active", max_seconds=0.0, as_of=2000.0)
        assert "seed_graduated" not in overdue

    def test_returns_empty_when_no_one_is_overdue(self, fresh_pipeline):
        fresh_pipeline.add_patient("noover_p", timestamp=0.0)
        fresh_pipeline.transition("noover_p", "screened", timestamp=0.0)
        fresh_pipeline.transition("noover_p", "enrolled", timestamp=0.0)
        fresh_pipeline.transition("noover_p", "active",   timestamp=9900.0)  # only 100 s in active
        assert fresh_pipeline.patients_overdue("active", max_seconds=500.0, as_of=10000.0) == []


class TestAverageTimeInState:
    def test_average_over_all_exited_patients(self, fresh_pipeline):
        # avg_pa: 1000 s in screened; avg_pb: 3000 s in screened → average = 2000
        fresh_pipeline.add_patient("avg_pa", timestamp=0.0)
        fresh_pipeline.transition("avg_pa", "screened", timestamp=0.0)
        fresh_pipeline.transition("avg_pa", "enrolled", timestamp=1000.0)

        fresh_pipeline.add_patient("avg_pb", timestamp=0.0)
        fresh_pipeline.transition("avg_pb", "screened", timestamp=0.0)
        fresh_pipeline.transition("avg_pb", "enrolled", timestamp=3000.0)

        assert fresh_pipeline.average_time_in_state("screened", as_of=99999.0) == pytest.approx(2000.0)

    def test_excludes_patients_still_in_state(self, fresh_pipeline):
        fresh_pipeline.add_patient("avg_pc", timestamp=0.0)
        fresh_pipeline.transition("avg_pc", "screened", timestamp=0.0)
        fresh_pipeline.transition("avg_pc", "enrolled", timestamp=1000.0)  # exited: 1000 s

        fresh_pipeline.add_patient("avg_pd", timestamp=0.0)
        fresh_pipeline.transition("avg_pd", "screened", timestamp=0.0)
        # avg_pd is still in screened at as_of=5000 — must be excluded

        assert fresh_pipeline.average_time_in_state("screened", as_of=5000.0) == pytest.approx(1000.0)

    def test_returns_zero_when_no_one_has_exited(self, fresh_pipeline):
        fresh_pipeline.add_patient("avg_pe", timestamp=0.0)
        fresh_pipeline.transition("avg_pe", "screened", timestamp=0.0)
        # avg_pe is still in screened
        assert fresh_pipeline.average_time_in_state("screened", as_of=5000.0) == pytest.approx(0.0)
