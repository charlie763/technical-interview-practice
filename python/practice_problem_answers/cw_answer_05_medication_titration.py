"""
user: coaches and physicians

uses
- reduce medication as blood sugar improves
- produce care and compliance reports

data:
- medication history
- blood sugar levels

question
- for TitrationEvent it sounds like dose_mg when direction is increase/decrease still tracks static number, not delta?
- can you start after stopping
"""

"""
=============================================================================
INTERVIEW PROBLEM 5: Medication Titration Tracker
Difficulty: Senior Software Engineer | Estimated time: 45 min
Company context: Health Tech
=============================================================================

CONTEXT
-------
<health tech co>'s remote clinical care program often de-escalates (reduces or
stops) diabetes medications as patients' blood sugar improves. Coaches and
physicians need to track each patient's medication history — when doses were
changed and why — so they can coordinate care and generate compliance reports.

You are building the TitrationTracker — the class that ingests a stream of
titration events and answers questions about each patient's medication history.

PRE-GIVEN (do not modify these)
--------------------------------
TitrationEvent and Medication are fully-implemented dataclasses.
You implement TitrationTracker.

Titration direction vocabulary:
  "increase"   — dose or frequency was raised
  "decrease"   — dose or frequency was lowered (de-escalation)
  "stop"       — medication discontinued entirely
  "start"      — new medication introduced

EXAMPLE
-------
  events = [
      TitrationEvent("pt1", "metformin",  "start",    500.0, date(2024, 1, 1)),
      TitrationEvent("pt1", "metformin",  "increase", 1000.0, date(2024, 2, 1)),
      TitrationEvent("pt1", "metformin",  "decrease",  500.0, date(2024, 3, 1)),
      TitrationEvent("pt1", "metformin",  "stop",        0.0, date(2024, 4, 1)),
  ]
  t = TitrationTracker(events)
  t.current_medications("pt1")
  # -> []
  t.titration_count("pt1", "metformin", direction="decrease")
  # -> 1

NOTES
-----
  - Events are not guaranteed to arrive in chronological order — sort by date.
  - A medication is "active" if the most recent event for it is NOT "stop".
  - All mutable state must be stored in instance variables set in __init__.
  - You choose the internal data structures — the public interface is what matters.
=============================================================================
"""

from dataclasses import dataclass
from datetime import date
from typing import Optional

# ---------------------------------------------------------------------------
# PRE-GIVEN — do not modify
# ---------------------------------------------------------------------------


@dataclass
class TitrationEvent:
    """A single medication change recorded by a Virta clinical coach or physician."""

    patient_id: str
    medication: str  # e.g. "metformin", "glipizide", "insulin_glargine"
    direction: str  # "start" | "increase" | "decrease" | "stop"
    dose_mg: float  # dose in milligrams at the time of this event (0.0 for "stop")
    recorded_on: date


@dataclass
class Medication:
    """Summary of a patient's current relationship with a single medication."""

    name: str
    current_dose: float
    last_changed: date
    total_changes: int  # total number of titration events (including start/stop)


# ---------------------------------------------------------------------------
# YOUR IMPLEMENTATION
# ---------------------------------------------------------------------------


class TitrationTracker:
    """
    Ingests a list of TitrationEvents and answers questions about
    patient medication histories.
    """

    def __init__(self, events: list[TitrationEvent]):
        # TODO: store and organize events however makes the methods efficient.
        # All state must be in instance variables (not class variables).
        # self.events = events
        # improvement: use string literal types
        # "metformin", "glipizide", "insulin_glargine"
        # metformin
        self.patients: dict[str, list[TitrationEvent]] = {}
        for event in events:
            existing_patient_events = self.patients.get(event.patient_id)
            if existing_patient_events:
                existing_patient_events.append(event)
            else:
                self.patients[event.patient_id] = [event]

    # -------------------------------------------------------------------------
    # PART 1 — Current medication snapshot  (~10 min)
    # -------------------------------------------------------------------------

    def current_medications(self, patient_id: str) -> list[Medication]:
        """
        Return a list of Medication objects for all currently active medications
        for the given patient (i.e., medications whose latest event is NOT "stop").

        Each Medication reflects:
          - name:          the medication name
          - current_dose:  dose_mg from the most recent event for that medication
          - last_changed:  date of the most recent event
          - total_changes: total number of TitrationEvents recorded for this medication

        Return an empty list if the patient has no events or all medications have
        been stopped.

        The list may be returned in any order.
        """
        # improvement throw errors i.e. if patient doesn't exist
        existing_patient_events = self.patients.get(patient_id)
        if not existing_patient_events:
            return []
        medications: dict[str, tuple[Medication, TitrationEvent]] = (
            {}
        )  # convent is for titrationEvent to be the last one
        for event in existing_patient_events:
            new_medication_tuple = (
                Medication(
                    name=event.medication,
                    current_dose=event.dose_mg,
                    last_changed=event.recorded_on,
                    total_changes=1,
                ),
                event,
            )
            existing_medication_tuple = medications.get(event.medication)
            if existing_medication_tuple:
                existing_medication = existing_medication_tuple[0]
                new_current_dose = existing_medication.current_dose
                new_last_changed = existing_medication.last_changed
                new_last_event = existing_medication_tuple[1]
                if event.recorded_on > existing_medication.last_changed:
                    new_current_dose = event.dose_mg
                    new_last_changed = event.recorded_on
                    new_last_event = event
                new_medication_tuple = (
                    Medication(
                        name=event.medication,
                        current_dose=new_current_dose,
                        last_changed=new_last_changed,
                        total_changes=existing_medication.total_changes + 1,
                    ),
                    new_last_event,
                )
            medications[event.medication] = new_medication_tuple
        non_stopped_medication_tuples = filter(
            lambda med_tuple: med_tuple[1].direction != "stop", medications.values()
        )
        return [med_tuple[0] for med_tuple in non_stopped_medication_tuples]

    def get_medication_history(
        self, patient_id: str, medication: str
    ) -> list[TitrationEvent]:
        """
        Return all TitrationEvents for (patient_id, medication), sorted
        chronologically (earliest first).

        Return an empty list if no events exist for that combination.
        """
        existing_events = self.patients.get(patient_id, [])
        events_per_medication = filter(
            lambda event: event.medication == medication, existing_events
        )
        return sorted(events_per_medication, key=lambda event: event.recorded_on)

    # -------------------------------------------------------------------------
    # PART 2 — Titration counts  (~10 min)
    # -------------------------------------------------------------------------

    def titration_count(
        self,
        patient_id: str,
        medication: str,
        direction: Optional[str] = None,
    ) -> int:
        """
        Return the number of titration events for (patient_id, medication).

        If direction is provided (one of "start", "increase", "decrease", "stop"),
        return only events with that direction.

        Return 0 if the patient or medication is unknown.

        """
        # use get medication_history to get all events
        # filter out duplicate directions (do we need to handle no direction case)
        # sum
        # improvement: make direction non-optional
        medication_history = self.get_medication_history(
            patient_id=patient_id, medication=medication
        )
        # existing_event_directions = set()
        total_count = 0
        for event in medication_history:
            if direction == None or event.direction == direction:
                # existing_event_directions.add(event.direction)
                total_count += 1
        return total_count

    def de_escalation_summary(self, patient_id: str) -> dict[str, int]:
        """
        Return a dict mapping each medication name to the number of "decrease" or
        "stop" events recorded for that patient.

        Only include medications that have at least one decrease or stop event.
        Return an empty dict if the patient has no such events.

        Example:
            {
                "metformin":       2,   # 1 decrease + 1 stop
                "glipizide":       1,   # 1 stop only
            }

        """
        de_esclation_summary = {}
        existing_patient_events = self.patients.get(patient_id, [])
        for event in existing_patient_events:
            if event.direction in ["stop", "decrease"]:
                current_med_count = de_esclation_summary.get(event.medication, 0)
                de_esclation_summary[event.medication] = current_med_count + 1
        return de_esclation_summary

    # -------------------------------------------------------------------------
    # PART 3 — Population-level queries  (~15 min)
    # -------------------------------------------------------------------------

    def patients_on_medication(self, medication: str) -> list[str]:
        """
        Return a sorted list of patient_ids who currently have the given
        medication active (latest event is NOT "stop").

        """
        # use get_medication_history for all patient ids
        # filter by latest event which should just be the last one in the array
        patient_ids = []
        for patient_id in self.patients.keys():
            med_history = self.get_medication_history(
                patient_id=patient_id, medication=medication
            )
            latest_event = med_history[-1]
            if latest_event.direction != "stop":
                patient_ids.append(patient_id)
        return sorted(patient_ids)

    def most_titrated_medications(self, top_n: int = 3) -> list[tuple[str, int]]:
        """
        Return the top_n medications (by total titration event count across ALL
        patients) as a list of (medication_name, total_count) tuples,
        sorted descending by count.

        If fewer than top_n medications exist, return all of them.
        Ties may appear in any order.

        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 4 — Live ingestion  (~10 min)
    # -------------------------------------------------------------------------

    def add_event(self, event: TitrationEvent) -> None:
        """
        Add a new TitrationEvent to the tracker.

        If an event with the same (patient_id, medication, recorded_on) already
        exists in the tracker, overwrite it with the new event (last-write wins).
        """
        raise NotImplementedError
