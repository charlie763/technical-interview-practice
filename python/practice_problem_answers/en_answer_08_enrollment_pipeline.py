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
from __future__ import annotations
import statistics

ALLOWED_TRANSITIONS: dict[str, set[str]] = {
    "referred":  {"screened"},
    "screened":  {"enrolled", "ineligible"},
    "enrolled":  {"active", "withdrawn"},
    "active":    {"graduated", "churned", "withdrawn"},
}

TERMINAL_STATES: set[str] = {"ineligible", "withdrawn", "graduated", "churned"}


class EnrollmentPipeline:
    """
    Tracks patients moving through a structured clinical enrollment pipeline.

    You choose the internal data structures — the public interface is what matters.

    Store all state in instance variables initialized in `__init__`.
    Class-level variables will bleed between tests and between EnrollmentPipeline
    instances — avoid them.

    Patient:
      {
        "state": str,
        "referred": timestamp (float) | None,
        "screened": timestamp (float) | None,
        "enrolled": timestamp (float) | None,
        "active": timpstamp (float) | None

      }

    PatientTrackerState:
      {
        "patients": {patient_id: Patient},
        "referred": [patient_ids: str],
        "screened": [patient_ids: str],
        "enrolled": [patient_ids: str],
        "active": [patient_ids: str],

      }
    """

    def __init__(self) -> None:
        self.patients = {}
        # self.referred set()
        # self.screened = set()
        # self.enrolled = set()
        # self.active = set()

    # ── Part 1: State tracking ────────────────────────────────────────────────

    def add_patient(self, patient_id: str, timestamp: float = 0.0) -> None:
        """
        Register a patient in the pipeline at the "referred" state.
        timestamp is when they entered the "referred" state (Unix seconds).
        Raises ValueError if patient_id is already registered.
        """
        try:
          if self.patients.get(patient_id): # already in system/pipeline
            raise ValueError
          self.patients[patient_id] = {
            "id": patient_id,
            "state": "referred",
            "referred": (timestamp, 0.0),
            "screened": None,
            "enrolled": None,
            "active": None,
            "ineligible": None,
            "withdrawn": None,
            "graduated": None,
            "churned": None
          }
        except ValueError as error:
          print(f'Patient already registered: {error}')
          raise
        return None

    def transition(self, patient_id: str, new_state: str, timestamp: float) -> None:
        """
        Advance a patient to new_state at the given timestamp.

        Raises ValueError if:
          - patient_id is not registered.
          - new_state is not a valid next state from the patient's current state
            (consult ALLOWED_TRANSITIONS).
          - the patient is already in a terminal state.
        """
        try:
          if not self.patients.get(patient_id):
            raise ValueError
        except ValueError as error:
          print(f'Patient not registered: {error}')
          raise

        try:
          if self.patients[patient_id]["state"] in TERMINAL_STATES:
            raise ValueError
        except ValueError as error:
          print(f'Patient in terminal state: {error}')
          raise

        try:
          current_state = self.patients[patient_id]["state"]
          if new_state not in ALLOWED_TRANSITIONS[current_state]:
            raise ValueError
        except ValueError as error:
          print(f'new state not in a valid next state: {error}')
          raise

        current_state = self.get_state(patient_id)

        # update how long patient was in the previous state
        start_timestamp, time_spent_in_state = self.patients[patient_id][current_state]
        self.patients[patient_id][current_state] = (start_timestamp, timestamp - start_timestamp)

        # transition patient state; we just got into this state so 0 time spent
        self.patients[patient_id][new_state] = (timestamp, 0.0)
        self.patients[patient_id]["state"] = new_state
        return None

    def get_state(self, patient_id: str) -> str:
        """Return the patient's current state. Raises ValueError if not registered."""
        try:
          if not self.patients.get(patient_id):
            raise ValueError
          return self.patients[patient_id]['state']
        except ValueError as error:
          print(f'Patient not registered: {error}')
          raise

    def get_patients_in_state(self, state: str) -> list[str]:
        """Return a sorted list of patient_ids currently in the given state."""
        return [patient["id"] for patient in self.patients.values() if patient['state'] == state]

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
        patient_state = self.get_state(patient_id)
        if not self.patients[patient_id][state]: # never been in state
          return 0.0

        start_timestamp, time_spent_in_state = self.patients[patient_id][state]
        if state != patient_state: # no longer in state
          return time_spent_in_state

        return as_of - start_timestamp

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
        # should use self.get_state instead of get('state') to ensure data is sanitized
        num_patients_exited_from_state = sum(1 for patient in self.patients.values()
                                            if patient.get('state') != from_state and
                                                patient.get(from_state)
                                            )
        if num_patients_exited_from_state == 0:
          return 0.0
        num_patients_in_to_state = sum(1 for patient in self.patients.values()
                                      if patient.get(from_state) and
                                          patient.get('state') == to_state
                                    )
        return (num_patients_in_to_state / num_patients_exited_from_state)

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
        patients_list = {}
        sorted_patients = []
        patients_in_state = self.get_patients_in_state(state)

        for patient_id in patients_in_state:
          time_spent = self.time_in_state(patient_id, state, as_of)
          if time_spent > max_seconds:
            patients_list[patient_id] = time_spent

        sorted_patients = sorted(patients_list, key=patients_list.get, reverse=True)
        return sorted_patients

    def average_time_in_state(self, state: str, as_of: float) -> float:
        """
        Return the mean seconds spent in state across all patients who have fully
        exited that state (their time is complete and will not grow further).

        - Patients currently in state are excluded from the average.
        - Returns 0.0 if no patients have fully exited state yet.

        Call time_in_state() for each patient's duration.
        """
        times_of_valid_patients = []
        for patient in self.patients.values():
          if patient.get(state) and patient['state'] != state:
            times_of_valid_patients.append(self.time_in_state(patient['id'], state, as_of))

        if not times_of_valid_patients:
          return 0.0
        return statistics.fmean(times_of_valid_patients)
