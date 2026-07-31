"""
NOTES

unknowns/questions
- A submission clears the specific deadline it satisfies (the earliest
    uncleared deadline on or after the submission date). ???
- deadline can be date or something like "quarterly"?
"""

"""
=============================================================================
INTERVIEW PROBLEM 6: Lab Cadence Compliance Monitor
Difficulty: Senior Software Engineer | Estimated time: 45 min
Company context: Health Tech
=============================================================================

CONTEXT
-------
<health tech co> requires patients to submit lab work at regular intervals so
clinicians can track metabolic health markers (HbA1c, fasting glucose, lipids,
kidney function, etc.). Patients who miss lab deadlines need follow-up from
their health coach.

You are building the LabCadenceMonitor — a function-based module (no class)
that manages a registry of patients, their required lab types, submission
deadlines, and actual submissions.

PRE-GIVEN (do not modify these)
--------------------------------
make_monitor() creates and returns the data store you will work with.
All functions receive the monitor dict as their first argument.

EXAMPLE
-------
  m = make_monitor()
  register_patient(m, "alice", required_labs=["hba1c", "bmp"])
  set_lab_deadline(m, "alice", "hba1c", due_date=date(2024, 3, 31))
  record_submission(m, "alice", "hba1c", submitted_on=date(2024, 3, 28))
  is_overdue(m, "alice", "hba1c", as_of=date(2024, 4, 1))  # -> False (submitted on time)
  is_overdue(m, "alice", "bmp",   as_of=date(2024, 4, 1))  # -> True  (no deadline set yet,
                                                              #          but bmp is required)

NOTES
-----
  - "overdue" means: a required lab has a deadline that has passed (as_of > due_date)
    AND no submission exists on or before the due_date.
  - If a required lab has no deadline set, it is NOT considered overdue.
  - A submission clears the specific deadline it satisfies (the earliest
    uncleared deadline on or after the submission date).
  - Patients can have multiple deadlines per lab type (e.g. quarterly HbA1c).
  - All state lives inside the dict returned by make_monitor(). No global state.
=============================================================================
"""

from datetime import date

# ---------------------------------------------------------------------------
# PRE-GIVEN — do not modify
# ---------------------------------------------------------------------------


def make_monitor() -> dict:
    """
    Return a fresh monitor data store.

    Schema (you may add keys as needed):
      {
        "patients": {
            patient_id: {
                "required_labs": set[str],
                "deadlines":     { lab_type: [date, ...] },   # sorted ascending
                "submissions":   { lab_type: [date, ...] },   # sorted ascending
            }
        }
      }
    """
    return {"patients": {}}


# ---------------------------------------------------------------------------
# YOUR IMPLEMENTATION
# ---------------------------------------------------------------------------


# -------------------------------------------------------------------------
# PART 1 — Patient & lab registration  (~10 min)
# -------------------------------------------------------------------------


def _get_patient_or_raise(monitor: dict, patient_id: str):
    patient = monitor["patients"].get(patient_id)
    if not patient:
        raise KeyError
    return patient


def register_patient(monitor: dict, patient_id: str, required_labs: list[str]) -> None:
    """
    Register a new patient with a list of required lab types.

    If the patient already exists, add any new lab types to their required set
    (do not remove existing ones). Idempotent for labs already in the set.

    Raise ValueError if required_labs is empty.
    """
    # improvement valid lab strings are of a certain certain kind
    # be consistent about types
    if len(required_labs) == 0:
        raise ValueError
    patient = monitor["patients"].get(patient_id)
    if patient:
        patient["required_labs"] = patient["required_labs"] | set(required_labs)
    else:
        monitor["patients"][patient_id] = {
            "required_labs": set(required_labs),
            "deadlines": {},
            "submissions": {},
        }


def add_required_lab(monitor: dict, patient_id: str, lab_type: str) -> None:
    """
    Add a single required lab type to an existing patient's requirements.

    Raise KeyError if the patient doesn't exist.
    No-op if the lab is already required.
    """
    patient = _get_patient_or_raise(monitor=monitor, patient_id=patient_id)
    patient["required_labs"].add(lab_type)


def get_required_labs(monitor: dict, patient_id: str) -> set[str]:
    """
    Return the set of required lab types for the patient.
    Raise KeyError if the patient doesn't exist.
    """
    patient = _get_patient_or_raise(monitor=monitor, patient_id=patient_id)
    return patient["required_labs"]


# -------------------------------------------------------------------------
# PART 2 — Deadlines and submissions  (~15 min)
# -------------------------------------------------------------------------


def _get_and_validate_required_labs_and_patient(
    monitor: dict, patient_id: str, lab_type: str
):
    patient = _get_patient_or_raise(monitor=monitor, patient_id=patient_id)
    required_labs = patient["required_labs"]
    if lab_type not in required_labs:
        raise ValueError
    return (patient, required_labs)


def set_lab_deadline(
    monitor: dict, patient_id: str, lab_type: str, due_date: date
) -> None:
    """
    Add a deadline for a specific lab type for the patient.

    A patient may have multiple deadlines for the same lab (e.g. quarterly).
    Duplicate deadlines (same patient + lab + date) are ignored.

    Raise KeyError if the patient doesn't exist.
    Raise ValueError if lab_type is not in the patient's required_labs.
    """
    # what does ignored mean, like don't add or duplicates are okay
    # set or list for deadlines, example suggest list, but set would make hitting specs easier
    # improvement - pre-sort deadlines so you don't have to do it over and over
    patient, _ = _get_and_validate_required_labs_and_patient(
        monitor=monitor, patient_id=patient_id, lab_type=lab_type
    )
    lab_type_deadlines = patient["deadlines"].get(lab_type)
    if not lab_type_deadlines:
        patient["deadlines"][lab_type] = {due_date}
    else:
        lab_type_deadlines.add(due_date)


def record_submission(
    monitor: dict, patient_id: str, lab_type: str, submitted_on: date
) -> None:
    """
    Record that the patient submitted a lab result on submitted_on.

    Clears the earliest uncleared deadline for this lab type that is
    >= submitted_on. If no such deadline exists, the submission is still
    recorded (it may satisfy a future deadline or serve as history).

    Raise KeyError if the patient doesn't exist.
    Raise ValueError if lab_type is not in the patient's required_labs.
    """
    patient, _ = _get_and_validate_required_labs_and_patient(
        monitor=monitor, patient_id=patient_id, lab_type=lab_type
    )
    lab_type_submissions = patient["submissions"].get(lab_type)
    if not lab_type_submissions:
        patient["submissions"][lab_type] = [submitted_on]
    else:
        lab_type_submissions.append(submitted_on)
    lab_type_deadlines = patient["deadlines"].get(lab_type)
    if not lab_type_deadlines:
        return

    sorted_deadlines = sorted(lab_type_deadlines)
    cleared_deadlines = []
    deadline_has_cleared = False
    for deadline in sorted_deadlines:
        if not deadline_has_cleared and deadline >= submitted_on:
            deadline_has_cleared = True
        else:
            cleared_deadlines.append(deadline)
    patient["deadlines"][lab_type] = cleared_deadlines


def is_overdue(monitor: dict, patient_id: str, lab_type: str, as_of: date) -> bool:
    """
    Return True if the patient has at least one uncleared deadline for
    lab_type that has passed as of `as_of` (i.e., due_date < as_of).

    Return False if:
      - The patient doesn't exist.
      - lab_type is not required for the patient.
      - No deadline has been set for that lab.
      - All past deadlines have been cleared by a submission.
    """
    patient = monitor["patients"].get(patient_id)
    if not patient:
        return False
    required_labs = patient["required_labs"]
    if lab_type not in required_labs:
        return False
    lab_type_deadlines = patient["deadlines"].get(lab_type)
    if not lab_type_deadlines:
        return False
    return any([deadlines < as_of for deadlines in lab_type_deadlines])


# -------------------------------------------------------------------------
# PART 3 — Compliance reporting  (~15 min)
# -------------------------------------------------------------------------


def overdue_labs(monitor: dict, patient_id: str, as_of: date) -> list[str]:
    """
    Return a sorted list of lab type names that are currently overdue for
    the patient as of `as_of`.

    Return an empty list if the patient doesn't exist or has no overdue labs.

    """
    # use is_overdue
    patient = monitor["patients"].get(patient_id)
    if not patient:
        return []
    overdue_lab_types = []
    for lab_type in patient["required_labs"]:
        if is_overdue(
            monitor=monitor, patient_id=patient_id, lab_type=lab_type, as_of=as_of
        ):
            overdue_lab_types.append(lab_type)
    return sorted(overdue_lab_types)


def compliance_report(monitor: dict, as_of: date) -> list[dict]:
    """
    Return a report of all patients with at least one overdue lab as of `as_of`.

    Each entry in the list is a dict:
    {
        "patient_id":    str,
        "overdue_labs":  list[str],   # sorted lab names
        "overdue_count": int,
    }

    Sort the list by overdue_count descending (most overdue first), then
    alphabetically by patient_id for ties.

    Return an empty list if no patient has overdue labs.

    """
    #  {
    #         "patients": {
    #             patient_id: {
    #                 "required_labs": set[str],
    #                 "deadlines":     { lab_type: [date, ...] },   # sorted ascending
    #                 "submissions":   { lab_type: [date, ...] },   # sorted ascending
    #             }
    #         }
    #       }
    #     """
    #     return {"patients": {}}
    report = []
    print(f"monitor: {monitor}")
    for patient_id in monitor["patients"].keys():
        patient_overdue_labs = overdue_labs(
            monitor=monitor, patient_id=patient_id, as_of=as_of
        )
        print(f"patient_overdue_labs: {patient_overdue_labs}")
        if patient_overdue_labs:
            report.append(
                {
                    "patient_id": patient_id,
                    "overdue_labs": patient_overdue_labs,
                    "overdue_count": len(patient_overdue_labs),
                }
            )
    return sorted(
        report,
        key=lambda patient_report: (
            patient_report["overdue_count"],
            patient_report["patient_id"],
        ),
        reverse=True,
    )


# -------------------------------------------------------------------------
# PART 4 — Submission history  (~5 min)
# -------------------------------------------------------------------------


def submission_history(monitor: dict, patient_id: str, lab_type: str) -> list[date]:
    """
    Return a chronologically sorted list of all submission dates for
    (patient_id, lab_type).

    Return an empty list if the patient doesn't exist or has no submissions
    for that lab type.
    """
    raise NotImplementedError


def days_since_last_submission(
    monitor: dict, patient_id: str, lab_type: str, as_of: date
) -> int | None:
    """
    Return the number of days between the patient's most recent submission
    for lab_type and `as_of`.

    Return None if the patient has never submitted that lab type.

    """
    raise NotImplementedError
