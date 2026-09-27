# =============================================================================
# INTERVIEW PROBLEM 4: Biomarker Alert Monitor
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# This health tech company delivers remote clinical care for chronic condition
# management. Patients use connected glucometers and ketone meters that sync
# readings into the app several times per day. The care team dashboard needs
# to surface patients whose numbers have been out of target range for
# multiple consecutive days, so health coaches can prioritize outreach.
#
# You are building the BiomarkerMonitor — the class that processes a stream
# of patient readings and answers questions about trends and outreach
# priority.
#
# PRE-GIVEN (do not modify)
# --------------------------
# BiomarkerReading is a fully-implemented Struct. You implement BiomarkerMonitor.
#
# Target ranges:
#   Glucose: 70–180 mg/dL   (< 70 is a dangerous low, > 180 is hyperglycemia)
#   Ketone:  0.5–3.0 mmol/L (below = not in ketosis, above = monitor)
#   Weight:  no absolute target — never flagged as out-of-range
#
# Example
# -------
#   readings = [
#     BiomarkerReading.new("alice", "glucose", 195.0, Date.new(2024, 1, 1)),
#     BiomarkerReading.new("alice", "glucose", 202.0, Date.new(2024, 1, 2)),
#     BiomarkerReading.new("alice", "glucose", 188.0, Date.new(2024, 1, 3)),
#   ]
#   m = BiomarkerMonitor.new(readings)
#   m.max_consecutive_out_of_range_days("alice", "glucose")  # -> 3
#   m.get_outreach_list(min_consecutive_days: 3)
#   # -> [{ patient_id: "alice", reading_type: "glucose",
#   #       consecutive_days: 3, latest_value: 188.0 }]
#
# NOTES
# -----
#   - Multiple readings on the same calendar day count as ONE day. A day is
#     "out-of-range" if ANY reading that day is out of range.
#   - Store all state in instance variables initialized in `initialize`.
# =============================================================================

require "date"

# ---------------------------------------------------------------------------
# PRE-GIVEN — do not modify
# ---------------------------------------------------------------------------

GLUCOSE_RANGE = (70.0..180.0) # mg/dL, inclusive
KETONE_RANGE = (0.5..3.0)     # mmol/L, inclusive

# A single biomarker measurement from a patient device or manual entry.
# reading_type is one of "glucose" | "ketone" | "weight".
BiomarkerReading = Struct.new(:patient_id, :reading_type, :value, :recorded_on)

# ---------------------------------------------------------------------------
# YOUR IMPLEMENTATION
# ---------------------------------------------------------------------------

class BiomarkerMonitor
  # readings: Array[BiomarkerReading]
  def initialize(readings)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 1 — single-reading classification (~5 min)
  # ---------------------------------------------------------------------------

  # Return true if the reading falls outside the target range for its type.
  #   - Glucose: outside GLUCOSE_RANGE
  #   - Ketone:  outside KETONE_RANGE
  #   - Weight:  never out-of-range (return false)
  def is_out_of_range(reading)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 2 — streak detection (~15 min)
  # ---------------------------------------------------------------------------

  # Return the length of the longest streak of *consecutive calendar days* on
  # which the patient had at least one out-of-range reading of the given type.
  #
  # Return 0 if the patient has no out-of-range readings of that type.
  #
  # Consecutive means no gap: Jan 1, Jan 2, Jan 3 is a streak of 3. Jan 1,
  # Jan 3 (skipping Jan 2) is two separate streaks of 1. Multiple readings on
  # the same day collapse to one day.
  def max_consecutive_out_of_range_days(patient_id, reading_type)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 3 — outreach list (~10 min)
  # ---------------------------------------------------------------------------

  # Return a list of patients who need proactive coach outreach because they
  # have been out-of-range for at least min_consecutive_days in a row.
  #
  # Each entry in the returned array is a hash:
  #   {
  #     patient_id:       String,
  #     reading_type:     String,
  #     consecutive_days: Integer,  # their max streak length
  #     latest_value:     Float,    # most recent out-of-range reading value
  #   }
  #
  # A patient can appear more than once if multiple reading types cross the
  # threshold (e.g., both glucose and ketone streaks).
  #
  # Sort the list by consecutive_days descending (most urgent first).
  def get_outreach_list(min_consecutive_days: 3)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 4 — deduplication on ingestion (~10 min)
  # ---------------------------------------------------------------------------

  # Add a new reading to the monitor's internal state.
  # Return true if the reading was added successfully.
  # Return false (without adding) if it is a duplicate.
  #
  # A duplicate is: same patient_id, reading_type, and recorded_on, with a
  # value within ±0.5 of an existing reading on that day.
  def add_reading(reading)
    raise NotImplementedError
  end
end
