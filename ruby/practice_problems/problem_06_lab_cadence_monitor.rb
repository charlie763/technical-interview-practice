# =============================================================================
# INTERVIEW PROBLEM 6: Lab Cadence Compliance Monitor
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# This health tech company requires patients to submit lab work at regular
# intervals so clinicians can track metabolic health markers (HbA1c, fasting
# glucose, lipids, kidney function, etc.). Patients who miss lab deadlines
# need follow-up from their health coach.
#
# You are building the Lab Cadence Monitor — a set of functions (no class)
# that manage a registry of patients, their required lab types, submission
# deadlines, and actual submissions.
#
# PRE-GIVEN (do not modify)
# --------------------------
# make_monitor creates and returns the data store you will work with. Every
# other function receives the monitor hash as its first argument.
#
# Example
# -------
#   m = make_monitor
#   register_patient(m, "alice", ["hba1c", "bmp"])
#   set_lab_deadline(m, "alice", "hba1c", Date.new(2024, 3, 31))
#   record_submission(m, "alice", "hba1c", Date.new(2024, 3, 28))
#   is_overdue(m, "alice", "hba1c", Date.new(2024, 4, 1))  # -> false (submitted on time)
#   is_overdue(m, "alice", "bmp",   Date.new(2024, 4, 1))  # -> true  (no deadline set yet,
#                                                           #          but bmp is required)
#
# NOTES
# -----
#   - "overdue" means: a required lab has a deadline that has passed (as_of > due_date)
#     AND no submission exists on or before the due_date.
#   - If a required lab has no deadline set, it is NOT considered overdue.
#   - A submission clears the specific deadline it satisfies (the earliest
#     uncleared deadline on or after the submission date).
#   - Patients can have multiple deadlines per lab type (e.g. quarterly HbA1c).
#   - All state lives inside the hash returned by make_monitor. No global state.
# =============================================================================

require "date"
require "set"

# ---------------------------------------------------------------------------
# PRE-GIVEN — do not modify
# ---------------------------------------------------------------------------

# Return a fresh monitor data store.
#
# Schema (you may add keys as needed):
#   {
#     patients: {
#       patient_id => {
#         required_labs: Set[String],
#         deadlines:     { lab_type => [Date, ...] },  # sorted ascending
#         submissions:   { lab_type => [Date, ...] },  # sorted ascending
#       }
#     }
#   }
def make_monitor
  { patients: {} }
end

# ---------------------------------------------------------------------------
# YOUR IMPLEMENTATION
# ---------------------------------------------------------------------------

# ---------------------------------------------------------------------------
# PART 1 — Patient & lab registration (~10 min)
# ---------------------------------------------------------------------------

# Register a new patient with an array of required lab types.
#
# If the patient already exists, add any new lab types to their required set
# (do not remove existing ones). Idempotent for labs already in the set.
#
# Raise ArgumentError if required_labs is empty.
def register_patient(monitor, patient_id, required_labs)
  raise NotImplementedError
end

# Add a single required lab type to an existing patient's requirements.
#
# Raise KeyError if the patient doesn't exist.
# No-op if the lab is already required.
def add_required_lab(monitor, patient_id, lab_type)
  raise NotImplementedError
end

# Return the set of required lab types for the patient.
# Raise KeyError if the patient doesn't exist.
def get_required_labs(monitor, patient_id)
  raise NotImplementedError
end

# ---------------------------------------------------------------------------
# PART 2 — Deadlines and submissions (~15 min)
# ---------------------------------------------------------------------------

# Add a deadline for a specific lab type for the patient.
#
# A patient may have multiple deadlines for the same lab (e.g. quarterly).
# Duplicate deadlines (same patient + lab + date) are ignored.
#
# Raise KeyError if the patient doesn't exist.
# Raise ArgumentError if lab_type is not in the patient's required_labs.
def set_lab_deadline(monitor, patient_id, lab_type, due_date)
  raise NotImplementedError
end

# Record that the patient submitted a lab result on submitted_on.
#
# Clears the earliest uncleared deadline for this lab type that is
# >= submitted_on. If no such deadline exists, the submission is still
# recorded (it may satisfy a future deadline or serve as history).
#
# Raise KeyError if the patient doesn't exist.
# Raise ArgumentError if lab_type is not in the patient's required_labs.
def record_submission(monitor, patient_id, lab_type, submitted_on)
  raise NotImplementedError
end

# Return true if the patient has at least one uncleared deadline for
# lab_type that has passed as of `as_of` (i.e., due_date < as_of).
#
# Return false if:
#   - The patient doesn't exist.
#   - lab_type is not required for the patient.
#   - No deadline has been set for that lab.
#   - All past deadlines have been cleared by a submission.
def is_overdue(monitor, patient_id, lab_type, as_of)
  raise NotImplementedError
end

# ---------------------------------------------------------------------------
# PART 3 — Compliance reporting (~15 min)
# ---------------------------------------------------------------------------

# Return a sorted array of lab type names that are currently overdue for
# the patient as of `as_of`.
#
# Return an empty array if the patient doesn't exist or has no overdue labs.
def overdue_labs(monitor, patient_id, as_of)
  raise NotImplementedError
end

# Return a report of all patients with at least one overdue lab as of `as_of`.
#
# Each entry in the array is a hash:
#   {
#     patient_id:    String,
#     overdue_labs:  Array[String],  # sorted lab names
#     overdue_count: Integer,
#   }
#
# Sort the array by overdue_count descending (most overdue first), then
# alphabetically by patient_id for ties.
#
# Return an empty array if no patient has overdue labs.
def compliance_report(monitor, as_of:)
  raise NotImplementedError
end

# ---------------------------------------------------------------------------
# PART 4 — Submission history (~5 min)
# ---------------------------------------------------------------------------

# Return a chronologically sorted array of all submission dates for
# (patient_id, lab_type).
#
# Return an empty array if the patient doesn't exist or has no submissions
# for that lab type.
def submission_history(monitor, patient_id, lab_type)
  raise NotImplementedError
end

# Return the number of days between the patient's most recent submission for
# lab_type and `as_of`.
#
# Return nil if the patient has never submitted that lab type.
def days_since_last_submission(monitor, patient_id, lab_type, as_of)
  raise NotImplementedError
end
