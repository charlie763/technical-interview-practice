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
        raise NotImplementedError

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
        raise NotImplementedError

    def get_medication_history(
        self, patient_id: str, medication: str
    ) -> list[TitrationEvent]:
        """
        Return all TitrationEvents for (patient_id, medication), sorted
        chronologically (earliest first).

        Return an empty list if no events exist for that combination.
        """
        raise NotImplementedError

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

        Hint: get_medication_history can be helpful here.
        """
        raise NotImplementedError

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

        Hint: titration_count can be helpful here.
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 3 — Population-level queries  (~15 min)
    # -------------------------------------------------------------------------

    def patients_on_medication(self, medication: str) -> list[str]:
        """
        Return a sorted list of patient_ids who currently have the given
        medication active (latest event is NOT "stop").

        Hint: current_medications can be helpful here.
        """
        raise NotImplementedError

    def most_titrated_medications(self, top_n: int = 3) -> list[tuple[str, int]]:
        """
        Return the top_n medications (by total titration event count across ALL
        patients) as a list of (medication_name, total_count) tuples,
        sorted descending by count.

        If fewer than top_n medications exist, return all of them.
        Ties may appear in any order.

        Hint: titration_count can be helpful here.
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
