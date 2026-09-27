# =============================================================================
# INTERVIEW PROBLEM 5: Medication Titration Tracker
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# This health tech company's remote clinical care program often de-escalates
# (reduces or stops) medications as patients' health metrics improve. Coaches
# and physicians need to track each patient's medication history — when
# doses were changed and why — so they can coordinate care and generate
# compliance reports.
#
# You are building the TitrationTracker — the class that ingests a stream of
# titration events and answers questions about each patient's medication
# history.
#
# PRE-GIVEN (do not modify)
# --------------------------
# TitrationEvent and Medication are fully-implemented Structs. You implement
# TitrationTracker.
#
# Titration direction vocabulary:
#   "increase"   — dose or frequency was raised
#   "decrease"   — dose or frequency was lowered (de-escalation)
#   "stop"       — medication discontinued entirely
#   "start"      — new medication introduced
#
# Example
# -------
#   events = [
#     TitrationEvent.new("pt1", "metformin", "start",     500.0, Date.new(2024, 1, 1)),
#     TitrationEvent.new("pt1", "metformin", "increase", 1000.0, Date.new(2024, 2, 1)),
#     TitrationEvent.new("pt1", "metformin", "decrease",  500.0, Date.new(2024, 3, 1)),
#     TitrationEvent.new("pt1", "metformin", "stop",        0.0, Date.new(2024, 4, 1)),
#   ]
#   t = TitrationTracker.new(events)
#   t.current_medications("pt1")                                   # -> []
#   t.titration_count("pt1", "metformin", direction: "decrease")   # -> 1
#
# NOTES
# -----
#   - Events are not guaranteed to arrive in chronological order — sort by date.
#   - A medication is "active" if the most recent event for it is NOT "stop".
#   - Store all state in instance variables initialized in `initialize`.
#   - You choose the internal data structures — the public interface is what matters.
# =============================================================================

require "date"

# ---------------------------------------------------------------------------
# PRE-GIVEN — do not modify
# ---------------------------------------------------------------------------

# A single medication change recorded by a clinical coach or physician.
# direction is one of "start" | "increase" | "decrease" | "stop".
# dose_mg is the dose in milligrams at the time of this event (0.0 for "stop").
TitrationEvent = Struct.new(:patient_id, :medication, :direction, :dose_mg, :recorded_on)

# Summary of a patient's current relationship with a single medication.
# total_changes is the total number of titration events (including start/stop).
Medication = Struct.new(:name, :current_dose, :last_changed, :total_changes)

# ---------------------------------------------------------------------------
# YOUR IMPLEMENTATION
# ---------------------------------------------------------------------------

class TitrationTracker
  # events: Array[TitrationEvent]
  def initialize(events)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Current medication snapshot (~10 min)
  # ---------------------------------------------------------------------------

  # Return an array of Medication objects for all currently active
  # medications for the given patient (i.e., medications whose latest event
  # is NOT "stop").
  #
  # Each Medication reflects:
  #   - name:          the medication name
  #   - current_dose:  dose_mg from the most recent event for that medication
  #   - last_changed:  date of the most recent event
  #   - total_changes: total number of TitrationEvents recorded for this medication
  #
  # Return an empty array if the patient has no events or all medications
  # have been stopped. The array may be returned in any order.
  def current_medications(patient_id)
    raise NotImplementedError
  end

  # Return all TitrationEvents for (patient_id, medication), sorted
  # chronologically (earliest first).
  # Return an empty array if no events exist for that combination.
  def get_medication_history(patient_id, medication)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Titration counts (~10 min)
  # ---------------------------------------------------------------------------

  # Return the number of titration events for (patient_id, medication).
  #
  # If direction is provided (one of "start", "increase", "decrease", "stop"),
  # return only events with that direction.
  #
  # Return 0 if the patient or medication is unknown.
  def titration_count(patient_id, medication, direction: nil)
    raise NotImplementedError
  end

  # Return a hash mapping each medication name to the number of "decrease" or
  # "stop" events recorded for that patient.
  #
  # Only include medications that have at least one decrease or stop event.
  # Return an empty hash if the patient has no such events.
  #
  # Example:
  #   { "metformin" => 2, "glipizide" => 1 }
  def de_escalation_summary(patient_id)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Population-level queries (~15 min)
  # ---------------------------------------------------------------------------

  # Return a sorted array of patient_ids who currently have the given
  # medication active (latest event is NOT "stop").
  def patients_on_medication(medication)
    raise NotImplementedError
  end

  # Return the top_n medications (by total titration event count across ALL
  # patients) as an array of [medication_name, total_count] pairs, sorted
  # descending by count.
  #
  # If fewer than top_n medications exist, return all of them.
  # Ties may appear in any order.
  def most_titrated_medications(top_n: 3)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 4 — Live ingestion (~10 min)
  # ---------------------------------------------------------------------------

  # Add a new TitrationEvent to the tracker.
  #
  # If an event with the same (patient_id, medication, recorded_on) already
  # exists in the tracker, overwrite it with the new event (last-write wins).
  def add_event(event)
    raise NotImplementedError
  end
end
