from __future__ import annotations

"""
MY NOTES
assumptions
 - it's a linear progression
 - SLA = service layer agreement?
 - timestamp in seconds? since when? 1970 mark?

questions
 - what does "-> 86400.0  (172800 - 86400; already exited, as_of ignored)" example mean below?
 - what needs to go in init?

improvement ideas
 - transition should be patient_transition
 - get_state too broad a name
 - enforcing typing, i.e Patient
 - more specific error messaging
 - use enums for state
 - keep state history or just STATE as separate record
"""

"""
Patient Enrollment Pipeline
===========================

A clinical care program moves patients through a structured enrollment pipeline.
Each patient begins in the "referred" state and advances through a predefined set
of transitions until reaching a terminal state.

Allowed transitions (defined in ALLOWED_TRANSITIONS below):
  referred   →  screened
  screened   →  enrolled  |  ineligible
  enrolled   →  active    |  withdrawn
  active     →  graduated |  churned  |  withdrawn

Terminal states: ineligible, withdrawn, graduated, churned

You choose the internal data structures — the public interface is what matters.

Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between EnrollmentPipeline
instances — avoid them.

────────────────────────────────────────────────────────────────────────────────
Part 1 — State tracking
  add_patient(patient_id, timestamp)
  transition(patient_id, new_state, timestamp)
  get_state(patient_id)              -> str
  get_patients_in_state(state)       -> list[str]

Part 2 — Duration and conversion metrics
  time_in_state(patient_id, state, as_of) -> float
  conversion_rate(from_state, to_state)   -> float

Part 3 — SLA monitoring
  patients_overdue(state, max_seconds, as_of) -> list[str]
  average_time_in_state(state, as_of)         -> float
────────────────────────────────────────────────────────────────────────────────

# Example
# pipeline = EnrollmentPipeline()
# pipeline.add_patient("p_001", timestamp=0.0)
# pipeline.transition("p_001", "screened", timestamp=86400.0)   # 1 day later
# pipeline.transition("p_001", "enrolled", timestamp=172800.0)  # 2 days later
# pipeline.get_state("p_001")                   # -> "enrolled"
# pipeline.get_patients_in_state("enrolled")    # -> ["p_001"]
#
# pipeline.add_patient("p_002", timestamp=0.0)
# pipeline.transition("p_002", "screened",   timestamp=43200.0)
# pipeline.transition("p_002", "ineligible", timestamp=86400.0)
# pipeline.get_patients_in_state("screened")    # -> []  (both have moved on)
#
# # Part 2
# pipeline.time_in_state("p_001", "screened", as_of=999999.0)
# # -> 86400.0  (172800 - 86400; already exited, as_of ignored)
# pipeline.conversion_rate("screened", "enrolled")
# # -> 0.5  (p_001 enrolled, p_002 ineligible; one of two converted)
#
# # Part 3
# pipeline.transition("p_001", "active", timestamp=259200.0)
# pipeline.patients_overdue("active", max_seconds=3600.0, as_of=270000.0)
# # -> ["p_001"]  (has been active 10800 s > 3600 s threshold)
# pipeline.average_time_in_state("screened", as_of=999999.0)
# # -> 64800.0  ((86400 + 43200) / 2; both p_001 and p_002 have exited screened)
"""

ALLOWED_TRANSITIONS: dict[str, set[str]] = {
    "referred": {"screened"},
    "screened": {"enrolled", "ineligible"},
    "enrolled": {"active", "withdrawn"},
    "active": {"graduated", "churned", "withdrawn"},
}

# TRANSITION_ORDERING = {
#     "referred": 0,
#     "screened": 1,
#     "enrolled": 2,
#     "active": {"graduated", "churned", "withdrawn"},
# }

TERMINAL_STATES: set[str] = {"ineligible", "withdrawn", "graduated", "churned"}


class EnrollmentPipeline:
    """
    Tracks patients moving through a structured clinical enrollment pipeline.

    You choose the internal data structures — the public interface is what matters.

    Store all state in instance variables initialized in `__init__`.
    Class-level variables will bleed between tests and between EnrollmentPipeline
    instances — avoid them.
    """

    def __init__(self) -> None:
        self.patients: dict[str, dict] = {}

    def _get_existing_patient(self, patient_id: str):
        existing_patient = self.patients.get(patient_id)
        if not existing_patient:
            raise ValueError
        return existing_patient

    # ── Part 1: State tracking ────────────────────────────────────────────────

    def add_patient(self, patient_id: str, timestamp: float = 0.0) -> None:
        """
        Register a patient in the pipeline at the "referred" state.
        timestamp is when they entered the "referred" state (Unix seconds).
        Raises ValueError if patient_id is already registered.
        """
        existing_patient = self.patients.get(patient_id)
        if existing_patient:
            raise ValueError
        self.patients[patient_id] = {
            "id": patient_id,
            "added_at": timestamp,
            "last_state_change": timestamp,
            "current_state": "referred",
            "state_history": {},  # prev state as key: total time spent in that state as value
        }

    def transition(self, patient_id: str, new_state: str, timestamp: float) -> None:
        """
        Advance a patient to new_state at the given timestamp.

        Raises ValueError if:
          - patient_id is not registered.
          - new_state is not a valid next state from the patient's current state
            (consult ALLOWED_TRANSITIONS).
          - the patient is already in a terminal state.
        """
        existing_patient = self._get_existing_patient(patient_id=patient_id)
        current_state = existing_patient["current_state"]
        if (
            current_state in TERMINAL_STATES
            or new_state not in ALLOWED_TRANSITIONS[current_state]
        ):
            raise ValueError
        existing_patient["state_history"][current_state] = (
            timestamp - existing_patient["last_state_change"]
        )
        existing_patient["current_state"] = new_state
        existing_patient["last_state_change"] = timestamp

    def get_state(self, patient_id: str) -> str:
        """Return the patient's current state. Raises ValueError if not registered."""
        existing_patient = self._get_existing_patient(patient_id=patient_id)
        return existing_patient["current_state"]

    def get_patients_in_state(self, state: str) -> list[str]:
        """Return a sorted list of patient_ids currently in the given state."""
        # do we need to raise an error if state is not valid?
        patients_in_state = filter(
            lambda patient: patient["current_state"] == state, self.patients.values()
        )
        return sorted(patients_in_state, key=lambda patient: patient["id"])

    # ── Part 2: Duration and conversion metrics ───────────────────────────────

    def time_in_state(self, patient_id: str, state: str, as_of: float) -> float:
        """
        Return the total seconds the patient has spent in the given state.

        - If the patient is currently in that state, count time from state entry
          up to as_of.
        - If the patient has already left that state, return the exact duration
          spent there (as_of is ignored).
        - Returns 0.0 if the patient has never been in that state.

        With the allowed transitions above, each state is visited at most once,
        so there is no ambiguity about multiple visits.
        """
        existing_patient = self._get_existing_patient(patient_id=patient_id)
        state_history_time = existing_patient["state_history"].get(state)
        if state_history_time:
            return state_history_time
        elif existing_patient["current_state"] == state:
            return as_of - existing_patient["last_state_change"]
        else:
            return 0.0

    def conversion_rate(self, from_state: str, to_state: str) -> float:
        """
        Of all patients who have exited from_state, return the fraction that
        transitioned directly to to_state.

        - Only patients who have already left from_state are counted; patients
          currently sitting in from_state are excluded (still undecided).
        - Returns 0.0 if no patients have exited from_state yet.

        Example: conversion_rate("screened", "enrolled") returns the share of
        screened patients who went on to enroll (vs. being marked ineligible).
        """
        # exited from state would me that that state is in state_history -> filter
        # transitioned directly to to_state. just means to_state is current state or to_state is also in state_history
        #  patients who have already left from_state are counted; -> just means from state is not current state (but that's impiciet)
        # Returns 0.0 if no patients have exited from_state yet. -> zero division check?
        all_patients_exited_from_state = list(
            filter(
                lambda patient: from_state in patient["state_history"].keys(),
                self.patients.values(),
            )
        )
        num_patients_exited_from_state = len(all_patients_exited_from_state)
        if num_patients_exited_from_state == 0:
            return 0.0
        patients_with_to_state = list(
            filter(
                lambda patient: patient["current_state"] == to_state
                or to_state in patient["state_history"].keys(),
                all_patients_exited_from_state,
            )
        )
        return len(patients_with_to_state) / num_patients_exited_from_state

    # ── Part 3: SLA monitoring ────────────────────────────────────────────────

    def patients_overdue(
        self, state: str, max_seconds: float, as_of: float
    ) -> list[str]:
        """
        Return patient_ids currently in state who have spent more than
        max_seconds there, sorted by time spent descending (longest-waiting first).

        Call time_in_state() for each patient's duration — do not re-implement
        the duration logic here.
        """
        # filter by over max_seconds
        # and then sort/map?
        # improvement: filter and sort at same time
        patients_over_max = filter(
            lambda patient: patient["current_state"] == state
            and self.time_in_state(patient_id=patient["id"], state=state, as_of=as_of)
            > max_seconds,
            self.patients.values(),
        )
        sorted_patients_over_max = sorted(
            patients_over_max,
            key=lambda patient: self.time_in_state(
                patient_id=patient["id"], state=state, as_of=as_of
            ),
            reverse=True,
        )
        return [patient["id"] for patient in sorted_patients_over_max]

    def average_time_in_state(self, state: str, as_of: float) -> float:
        """
        Return the mean seconds spent in state across all patients who have fully
        exited that state (their time is complete and will not grow further).

        - Patients currently in state are excluded from the average.
        - Returns 0.0 if no patients have fully exited state yet.

        Call time_in_state() for each patient's duration.
        """
        # improvement: helper for all_patients_exited_from_state?
        all_patients_exited_from_state = list(
            filter(
                lambda patient: state in patient["state_history"].keys(),
                self.patients.values(),
            )
        )
        num_patients_exited_from_state = len(all_patients_exited_from_state)
        if num_patients_exited_from_state == 0:
            return 0.0
        total_time_in_state = sum(
            patient["state_history"][state]
            for patient in all_patients_exited_from_state
        )
        return total_time_in_state / num_patients_exited_from_state
