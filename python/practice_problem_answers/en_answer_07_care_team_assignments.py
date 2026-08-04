from __future__ import annotations

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
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)


class CareTeamManager:
    """
    Manages patient-to-care-team-member assignments for a remote clinical platform.

    You choose the internal data structures — the public interface is what matters.

    Store all state in instance variables initialized in `__init__`.
    Class-level variables will bleed between tests and between CareTeamManager
    instances — avoid them.

    Members:
    {
        dict[member_id(str), Member]
    }

    Member:
    {
        role: str,
        num_patients: int,
        max_patients: int,
        patients: set[str]
    }

    Roles:
    {
        dict[role(str), Role]
    }

    Role:
    {
        available_members: set[member_id (str)]
        unavailable_members: set[member_id (str)]
    }

    # a list is ordered in a way a set isn't and you can keep track of
    # if the patient moves back to a previous team member
    Patients:
    {
        patient_id: {
            current_members: dict[role (str), member_id (str)],
            member_history: dict[role (str): dict[member_id (str), list[timestamps(int)]]
        }
    }

    """

    def __init__(self) -> None:
        #self.roles: dict[str, dict] = {}
        self.members: dict[str,dict] = {}
        self.patients: dict[str, dict] = {}

    # ── Part 1: Basic assignment and lookup ───────────────────────────────────

    def add_member(self, member_id: str, role: str, max_patients: int) -> None:
        """Register a care team member with the given role and patient capacity."""
        if self.members.get(member_id):
            # assumes you cannot update a member's role in the system
            raise ValueError('member already exists')
        self.members[member_id] = {
            "role": role,
            "num_patients": 0,
            "max_patients": max_patients,
            "patients": set()
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
        if self.members.get(member_id) is None:
            raise ValueError("member doesn't exist")

        if patient_id in self.members[member_id]["patients"]:
            return None

        if self.members[member_id]["num_patients"] == self.members[member_id]["max_patients"]:
            raise CapacityError('At capacity')

        # is user already assigned to member with same role, remove patient from prev member
        role = self.members[member_id]["role"]
        if self.patients.get(patient_id) is None: # existing patient
            self.patients[patient_id] = {
                "current_members": { role: member_id },
                "member_history": {
                    role: {
                        member_id: [assigned_at, None],
                    }
                }
            }
        else:
            member_id_with_role = self.patients[patient_id]["current_members"].get(role)
            if member_id_with_role:
                old_assigned_at, old_unassigned_at = self.patients[patient_id]["member_history"][role][member_id_with_role]
                self.patients[patient_id]["member_history"][role][member_id_with_role] = [old_assigned_at, assigned_at]
                self.members[member_id_with_role]["patients"].discard(patient_id)
                self.members[member_id_with_role]["num_patients"] -= 1
            if self.patients[patient_id]["member_history"].get(role) is None:
                self.patients[patient_id]["member_history"][role] = { member_id: [assigned_at, None] }
            else:
                self.patients[patient_id]["member_history"][role][member_id] = [assigned_at, None]
            self.patients[patient_id]["current_members"][role] = member_id

        self.members[member_id]["patients"].add(patient_id)
        self.members[member_id]["num_patients"] += 1

    def get_assignment(self, patient_id: str, role: str) -> Optional[str]:
        """
        Return the member_id currently assigned to this patient for the given
        role, or None if no member of that role is currently assigned.
        """
        if self.patients.get(patient_id) is None:
            return None
        return self.patients[patient_id]["current_members"].get(role)


    def get_patients(self, member_id: str) -> list[str]:
        """
        Return a sorted list of patient_ids currently assigned to this member.
        Raises ValueError if member_id has not been registered.
        """
        if self.members.get(member_id) is None:
            raise ValueError("member doesn't exist")
        return(sorted(self.members[member_id]["patients"]))

    # ── Part 2: Capacity enforcement ──────────────────────────────────────────

    def available_members(self, role: str) -> list[str]:
        """
        Return a sorted list of member_ids with the given role that still have
        open capacity (current patient count < max_patients).
        """
        available_members = []
        for member_id, member_contents in self.members.items():
            if member_contents["role"] == role:
                if member_contents["num_patients"] < member_contents["max_patients"]:
                    available_members.append(member_id)
        return sorted(available_members)

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
        history = []
        if self.patients.get(patient_id) and self.patients[patient_id]["member_history"].get(role):
            for member_id, assignment in self.patients[patient_id]["member_history"][role].items():
                assigned_at, unassigned_at = assignment
                history.append((member_id, assigned_at, unassigned_at))
        if history:
            return sorted(history, key=lambda x: x[1])
        return []

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
        if self.patients.get(patient_id) is None:
            return None
        history = self.get_history(patient_id, role)
        for member_id, assigned_at, unassigned_at in history:
            if assigned_at <= timestamp and (unassigned_at is None or timestamp < unassigned_at):
                return member_id
        return None