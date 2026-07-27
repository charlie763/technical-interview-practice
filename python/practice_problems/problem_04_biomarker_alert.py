"""
=============================================================================
INTERVIEW PROBLEM 4: Biomarker Alert Monitor
Difficulty: Senior Software Engineer | Estimated time: 45 min
Company context: Virta Health — Care Delivery Engineering
=============================================================================

CONTEXT
-------
Virta Health delivers remote clinical care for type 2 diabetes reversal.
Patients use connected glucometers and ketone meters that sync readings into
the Virta app several times per day. The care team dashboard needs to surface
patients whose numbers have been out of target range for multiple consecutive
days, so health coaches can prioritize outreach.

You are building the BiomarkerMonitor — the class that processes a stream of
patient readings and answers questions about trends and outreach priority.

PRE-GIVEN (do not modify these)
--------------------------------
BiomarkerReading is a fully-implemented dataclass. You implement BiomarkerMonitor.

Target ranges:
  Glucose: 70–180 mg/dL  (< 70 is a dangerous low, > 180 is hyperglycemia)
  Ketone:  0.5–3.0 mmol/L  (below = not in ketosis, above = monitor)
  Weight:  no absolute target — not flagged as out-of-range

EXAMPLE
-------
  readings = [
      BiomarkerReading("alice", "glucose", 195.0, date(2024, 1, 1)),
      BiomarkerReading("alice", "glucose", 202.0, date(2024, 1, 2)),
      BiomarkerReading("alice", "glucose", 188.0, date(2024, 1, 3)),
  ]
  m = BiomarkerMonitor(readings)
  m.max_consecutive_out_of_range_days("alice", "glucose")  # -> 3
  m.get_outreach_list(min_consecutive_days=3)
  # -> [{"patient_id": "alice", "reading_type": "glucose",
  #       "consecutive_days": 3, "latest_value": 188.0}]

NOTES
-----
  - Multiple readings on the same calendar day count as ONE day.
    A day is "out-of-range" if ANY reading that day is out of range.
  - All mutable state must be stored in instance variables set in __init__.
=============================================================================
"""

from dataclasses import dataclass, field
from datetime import date
from typing import Optional


# ---------------------------------------------------------------------------
# PRE-GIVEN — do not modify
# ---------------------------------------------------------------------------

GLUCOSE_RANGE = (70.0, 180.0)   # mg/dL, inclusive
KETONE_RANGE  = (0.5, 3.0)      # mmol/L, inclusive


@dataclass
class BiomarkerReading:
    """A single biomarker measurement from a patient device or manual entry."""
    patient_id:   str
    reading_type: str    # "glucose" | "ketone" | "weight"
    value:        float
    recorded_on:  date


# ---------------------------------------------------------------------------
# YOUR IMPLEMENTATION
# ---------------------------------------------------------------------------

class BiomarkerMonitor:
    """
    Ingests a list of BiomarkerReadings and answers questions about
    out-of-range trends across the patient population.
    """

    def __init__(self, readings: list[BiomarkerReading]):
        # TODO: store and organize readings however makes the methods efficient.
        # All state must be in instance variables (not class variables).
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 1 — single-reading classification  (~5 min)
    # -------------------------------------------------------------------------

    def is_out_of_range(self, reading: BiomarkerReading) -> bool:
        """
        Return True if the reading falls outside the target range for its type.
        - Glucose: outside GLUCOSE_RANGE
        - Ketone:  outside KETONE_RANGE
        - Weight:  never out-of-range (return False)
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 2 — streak detection  (~15 min)
    # -------------------------------------------------------------------------

    def max_consecutive_out_of_range_days(
        self, patient_id: str, reading_type: str
    ) -> int:
        """
        Return the length of the longest streak of *consecutive calendar days*
        on which the patient had at least one out-of-range reading of the given type.

        Return 0 if the patient has no out-of-range readings of that type.

        Consecutive means no gap: Jan 1, Jan 2, Jan 3 is a streak of 3.
        Jan 1, Jan 3 (skipping Jan 2) is two separate streaks of 1.
        Multiple readings on the same day collapse to one day.
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 3 — outreach list  (~10 min)
    # -------------------------------------------------------------------------

    def get_outreach_list(self, min_consecutive_days: int = 3) -> list[dict]:
        """
        Return a list of patients who need proactive coach outreach because
        they have been out-of-range for at least min_consecutive_days in a row.

        Each entry in the returned list is a dict:
        {
            "patient_id":       str,
            "reading_type":     str,
            "consecutive_days": int,   # their max streak length
            "latest_value":     float, # most recent out-of-range reading value
        }

        A patient can appear more than once if multiple reading types cross the
        threshold (e.g., both glucose and ketone streaks).

        Sort the list by consecutive_days descending (most urgent first).
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 4 — deduplication on ingestion  (~10 min)
    # -------------------------------------------------------------------------

    def add_reading(self, reading: BiomarkerReading) -> bool:
        """
        Add a new reading to the monitor's internal state.
        Return True if the reading was added successfully.
        Return False (without adding) if it is a duplicate.

        A duplicate is: same patient_id, reading_type, and recorded_on,
        with a value within ±0.5 of an existing reading on that day.
        """
        raise NotImplementedError
