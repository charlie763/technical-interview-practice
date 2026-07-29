"""
NOTES TO SELF
goal:
- are numbers out of range
- health coaches know to prioritze outreach
- need to break down problem more

for AI
- get rid of Hints
"""

"""
=============================================================================
INTERVIEW PROBLEM 4: Biomarker Alert Monitor
Difficulty: Senior Software Engineer | Estimated time: 45 min
Company context: Health Tech
=============================================================================

CONTEXT
-------
<health tech co> delivers remote clinical care for type 2 diabetes reversal.
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

GLUCOSE_RANGE = (70.0, 180.0)  # mg/dL, inclusive
KETONE_RANGE = (0.5, 3.0)  # mmol/L, inclusive


@dataclass
class BiomarkerReading:
    """A single biomarker measurement from a patient device or manual entry."""

    patient_id: str  # often patient name, at least in test data
    reading_type: str  # "glucose" | "ketone" | "weight"
    value: float
    recorded_on: date


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
        self.readings = readings

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
        if reading.reading_type == "glucose" and (
            reading.value < GLUCOSE_RANGE[0] or reading.value > GLUCOSE_RANGE[1]
        ):
            return True
        if reading.reading_type == "ketone" and (
            reading.value < KETONE_RANGE[0] or reading.value > KETONE_RANGE[1]
        ):
            return True
        return False

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

        Hint: is_out_of_range can be helpful here.
        """
        patient_readings = list(
            filter(
                lambda reading: reading.patient_id == patient_id
                and reading.reading_type == reading_type,
                self.readings,
            )
        )
        print(f"patient_readings: {len(patient_readings)}")
        if len(patient_readings) == 0:
            return 0
        sorted_patient_readings = sorted(
            patient_readings, key=lambda reading: reading.recorded_on
        )
        current_day = sorted_patient_readings[0].recorded_on.day
        grouped_patient_readings = []
        current_reading_group = []
        for reading in sorted_patient_readings:
            if reading.recorded_on.day == current_day:
                current_reading_group.append(reading)
            else:
                current_day = reading.recorded_on.day
                grouped_patient_readings.append(current_reading_group)
                current_reading_group = [reading]
        grouped_patient_readings.append(current_reading_group)
        first_reading = sorted_patient_readings[0]
        longest_streak = 0
        current_streak = 0
        most_recent_day = first_reading.recorded_on.day
        for reading_group in grouped_patient_readings:
            group_day = reading_group[0].recorded_on.day
            any_reading_out_of_range = any(
                self.is_out_of_range(reading=reading) for reading in reading_group
            )
            if group_day - most_recent_day <= 1 and any_reading_out_of_range:
                current_streak += 1
            else:
                current_streak = 0
            if current_streak > longest_streak:
                longest_streak = current_streak
            most_recent_day = group_day
        return longest_streak

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
        # break up readings by patient id and reading_type,
        # pass into max_consecutive_out_of_range_days
        # keep track of results that come out of that
        # normalize data to shape in method description -  i.e. latest_value
        # sort based on longest streak/consecutive_days
        # question: if patient can show up multiple times, what does it mean for them to be sorted (should their entries be still grouped by patient id even if one is less urgent)
        # improvement note: make this logic and max_consecutive_out_of_range_days not redundant in terms of looping through readings
        # improvement note: sort during instead of after the fact
        # improvement note: add latest_value to return of max_consecutive_out_of_range_days
        # improvement note: make helper function for getting most recent out of range value
        # improvement note: get rid of list wrapper around filter
        outreach_list = []
        patient_ids = set()
        for reading in self.readings:
            patient_ids.add(reading.patient_id)
        for patient_id in patient_ids:
            glucose_max_streak = self.max_consecutive_out_of_range_days(
                patient_id=patient_id, reading_type="glucose"
            )
            ketone_max_streak = self.max_consecutive_out_of_range_days(
                patient_id=patient_id, reading_type="ketone"
            )
            if glucose_max_streak >= min_consecutive_days:
                patient_readings_out_of_range = list(
                    filter(
                        lambda reading: reading.patient_id == patient_id
                        and reading.reading_type == "glucose"
                        and self.is_out_of_range(reading=reading),
                        self.readings,
                    )
                )
                max_reading = max(
                    patient_readings_out_of_range,
                    key=lambda reading: reading.recorded_on.day,
                )
                outreach_list.append(
                    {
                        "patient_id": patient_id,
                        "reading_type": "glucose",
                        "consecutive_days": glucose_max_streak,
                        "latest_value": max_reading.value,
                    }
                )
            if ketone_max_streak >= min_consecutive_days:
                patient_readings_out_of_range = list(
                    filter(
                        lambda reading: reading.patient_id == patient_id
                        and reading.reading_type == "ketone"
                        and self.is_out_of_range(reading=reading),
                        self.readings,
                    )
                )
                max_reading = max(
                    patient_readings_out_of_range,
                    key=lambda reading: reading.recorded_on.day,
                )
                outreach_list.append(
                    {
                        "patient_id": patient_id,
                        "reading_type": "ketone",
                        "consecutive_days": ketone_max_streak,
                        "latest_value": max_reading.value,
                    }
                )
        sorted_outreach_list = sorted(
            outreach_list,
            key=lambda outreach_entry: outreach_entry["consecutive_days"],
            reverse=True,
        )
        return sorted_outreach_list

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
        # improvement note: just insert readings already sorted so we don't have to sort every time
        possible_duplicate_readings = filter(
            lambda existing_reading: existing_reading.patient_id == reading.patient_id
            and existing_reading.reading_type == reading.reading_type
            and existing_reading.recorded_on == reading.recorded_on,
            self.readings,
        )
        is_duplicate = any(
            reading.value < possible_duplicate.value + 0.5
            and reading.value > possible_duplicate.value - 0.5
            for possible_duplicate in possible_duplicate_readings
        )
        print(f"is_duplicate: {is_duplicate}")
        if is_duplicate:
            return False
        else:
            print("got here")
            self.readings.append(reading)
            return True
