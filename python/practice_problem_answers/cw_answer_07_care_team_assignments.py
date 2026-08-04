from __future__ import annotations

"""
maximum number of patients
pecific role ("coach", "physician", "dietitian",
A patient may have at most
one assigned member per role at a time.
"""

"""
Care Team Assignment Manager
============================

A remote clinical platform supports patients through a care team. Each care
team member has a specific role ("coach", "physician", "dietitian", etc.) and a
maximum number of patients they can hold at one time. A patient may have at most
one assigned member per role at a time. When a patient is reassigned to a
different member, the full history of past assignments is preserved for audit
and care-continuity purposes.

You choose the internal data structures — the public interface is what matters.

Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between CareTeamManager
instances — avoid them.

────────────────────────────────────────────────────────────────────────────────
Part 1 — Basic assignment and lookup
  add_member(member_id, role, max_patients)
  assign(patient_id, member_id, assigned_at)
  get_assignment(patient_id, role)   -> Optional[str]
  get_patients(member_id)            -> list[str]

Part 2 — Capacity enforcement
  assign() now raises CapacityError when the member is at max_patients
  available_members(role)            -> list[str]

Part 3 — Assignment history
  get_history(patient_id, role)               -> list[tuple[str, float, Optional[float]]]
  get_assignment_at(patient_id, role, timestamp) -> Optional[str]
────────────────────────────────────────────────────────────────────────────────

# Example
# mgr = CareTeamManager()
# mgr.add_member("coach_a", "coach", max_patients=2)
# mgr.add_member("dr_main", "physician", max_patients=100)
# mgr.assign("patient_1", "coach_a", assigned_at=1000.0)
# mgr.assign("patient_1", "dr_main", assigned_at=1000.0)
# mgr.get_assignment("patient_1", "coach")      # -> "coach_a"
# mgr.get_assignment("patient_1", "dietitian")  # -> None
# mgr.get_patients("coach_a")                   # -> ["patient_1"]
#
# # Part 2
# mgr.add_member("coach_b", "coach", max_patients=1)
# mgr.assign("patient_2", "coach_b", assigned_at=2000.0)
# mgr.assign("patient_3", "coach_b", assigned_at=3000.0)  # raises CapacityError
# mgr.available_members("coach")                # -> ["coach_a"]
#
# # Part 3 — reassign patient_1 from coach_a to coach_b
# mgr.assign("patient_1", "coach_b", assigned_at=5000.0)
# mgr.get_history("patient_1", "coach")
# # -> [("coach_a", 1000.0, 5000.0), ("coach_b", 5000.0, None)]
# mgr.get_assignment_at("patient_1", "coach",  500.0)   # -> None (before any assignment)
# mgr.get_assignment_at("patient_1", "coach", 3000.0)   # -> "coach_a"
# mgr.get_assignment_at("patient_1", "coach", 6000.0)   # -> "coach_b"
"""

from typing import Optional


class CapacityError(Exception):
    """Raised when assigning a patient to a member who is at their patient capacity."""

    pass


class CareTeamManager:
    """
    Manages patient-to-care-team-member assignments for a remote clinical platform.

    You choose the internal data structures — the public interface is what matters.

    Store all state in instance variables initialized in `__init__`.
    Class-level variables will bleed between tests and between CareTeamManager
    instances — avoid them.
    """

    def __init__(self) -> None:
        self.members = (
            {}
        )  # {id: {"id": member_id, "role": role, "max_patients": max_patients}}
        self.history = {}
        # { 'patient_id': {
        #         "current_member_ids": list[str],
        #         "past_assignments":list[("coach_a", 1000.0, 5000.0)]
        #     }
        # }

    def _get_or_raise_for_member(self, member_id: str):
        assigned_member = self.members.get(member_id)
        if not assigned_member:
            raise ValueError
        return assigned_member

    # ── Part 1: Basic assignment and lookup ───────────────────────────────────

    def add_member(self, member_id: str, role: str, max_patients: int) -> None:
        # improvment/question: make sure we don't add duplicate members?
        self.members[member_id] = {
            "id": member_id,
            "role": role,
            "max_patients": max_patients,
        }

    def assign(self, patient_id: str, member_id: str, assigned_at: float) -> None:
        """
        Assign a patient to a care team member (assigned_at is Unix seconds).

        A patient may have at most one assigned member per role at a time.
        If the patient already has a member with the same role, that assignment
        is replaced — the new assignment takes effect at assigned_at.

        Raises ValueError  if member_id has not been registered via add_member.

        Part 2 addition: raises CapacityError if the member is already at
        max_patients and the patient is not currently assigned to that exact member.
        (Reassigning a patient who is already on this member does not count as
        adding a new patient — it is a no-op for capacity purposes.)
        """
        # improvement: more descriptive error messaging
        # improvement: pre-sorting
        # what do we do if patient is not registered yet?
        assigned_member = self._get_or_raise_for_member(member_id=member_id)
        patient_data = self.history.get(
            patient_id, {"current_member_ids": [], "past_assignments": []}
        )
        current_patient_members = filter(
            lambda member: member["id"] in patient_data.get("current_member_ids", [])
            and member["role"] != assigned_member["role"],
            self.members.values(),
        )
        current_patient_member_ids = [
            member["id"] for member in current_patient_members
        ]
        current_patient_member_ids.append(member_id)
        patient_data["current_member_ids"] = current_patient_member_ids
        patient_data["past_assignments"].append((member_id, assigned_at, None))
        self.history[patient_id] = patient_data

    def get_assignment(self, patient_id: str, role: str) -> Optional[str]:
        """
        Return the member_id currently assigned to this patient for the given
        role, or None if no member of that role is currently assigned.
        """
        patient_data = self.history.get(patient_id, {})

        current_patient_member_with_role = filter(
            lambda member: member["id"] in patient_data.get("current_member_ids", [])
            and member["role"] == role,
            self.members.values(),
        )
        try:
            member = next(current_patient_member_with_role)
            return member["id"]
        except:  # figure out exact error
            return None

    def get_patients(self, member_id: str) -> list[str]:
        """
        Return a sorted list of patient_ids currently assigned to this member.
        Raises ValueError if member_id has not been registered.
        """
        assigned_member = self._get_or_raise_for_member(member_id=member_id)
        patients_with_member = []
        for patient_id, patient_data in self.history.items():
            if assigned_member["id"] in patient_data["current_member_ids"]:
                patients_with_member.append(patient_id)
        return sorted(patients_with_member)

    # ── Part 2: Capacity enforcement ──────────────────────────────────────────

    def available_members(self, role: str) -> list[str]:
        """
        Return a sorted list of member_ids with the given role that still have
        open capacity (current patient count < max_patients).
        """
        # filter member sby role
        # use get_patients to see if over max capacity
        # improvment have non-sorted bet patients for performance
        av_member_ids = []
        for member in self.members.values():
            if member["role"] != role:
                continue
            member_patients = self.get_patients(member_id=member["id"])
            if len(member_patients) < member["max_patients"]:
                av_member_ids.append(member["id"])
        return sorted(av_member_ids)

    # ── Part 3: Assignment history ────────────────────────────────────────────

    def get_history(
        self, patient_id: str, role: str
    ) -> list[tuple[str, float, Optional[float]]]:
        """
        Return the full assignment history for the patient's given role as a list
        of (member_id, assigned_at, unassigned_at) tuples sorted by assigned_at.

        - unassigned_at is None for the current (still-active) assignment.
        - unassigned_at equals the assigned_at of the subsequent assignment for
          past entries.
        - Returns [] if the patient has never been assigned a member of this role.
        """
        # go back and add unassigned at time
        patient_history = self.history.get(patient_id, {}).get("past_assignments", [])
        print(f"patient_history: {patient_history}")
        patient_history_for_role = []
        for assignment in patient_history:
            assignment_member = self._get_or_raise_for_member(member_id=assignment[0])
            if assignment_member["role"] == role:
                patient_history_for_role.append(assignment)
        return sorted(patient_history_for_role, key=lambda history: history[1])

    def get_assignment_at(
        self, patient_id: str, role: str, timestamp: float
    ) -> Optional[str]:
        """
        Return the member_id assigned to the patient for the given role at the
        given timestamp, or None if no assignment was active at that time.

        An assignment is active during the interval [assigned_at, unassigned_at).
        The current assignment (unassigned_at is None) is active from assigned_at
        onward.

        Implement this by calling get_history() — do not duplicate the lookup logic.
        """
        raise NotImplementedError
