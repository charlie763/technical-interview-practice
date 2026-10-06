# =============================================================================
# Care Team Assignment Manager
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# A remote clinical platform supports patients through a care team. Each
# care team member has a specific role ("coach", "physician", "dietitian",
# etc.) and a maximum number of patients they can hold at one time. A
# patient may have at most one assigned member per role at a time. When a
# patient is reassigned to a different member, the full history of past
# assignments is preserved for audit and care-continuity purposes.
#
# You choose the internal data structures — the public interface is what
# matters. Store all state in instance variables initialized in
# `initialize`. Class-level variables will bleed between tests and between
# CareTeamManager instances — avoid them.
#
# -------------------------------------------------------------------------
# Part 1 — Basic assignment and lookup
#   add_member(member_id, role, max_patients)
#   assign(patient_id, member_id, assigned_at)
#   get_assignment(patient_id, role)   -> String | nil
#   get_patients(member_id)            -> Array[String]
#
# Part 2 — Capacity enforcement
#   assign now raises CapacityError when the member is at max_patients
#   available_members(role)            -> Array[String]
#
# Part 3 — Assignment history
#   get_history(patient_id, role)               -> Array[[String, Float, Float | nil]]
#   get_assignment_at(patient_id, role, timestamp) -> String | nil
# -------------------------------------------------------------------------
#
# Example
# -------
#   mgr = CareTeamManager.new
#   mgr.add_member("coach_a", "coach", 2)
#   mgr.add_member("dr_main", "physician", 100)
#   mgr.assign("patient_1", "coach_a", 1000.0)
#   mgr.assign("patient_1", "dr_main", 1000.0)
#   mgr.get_assignment("patient_1", "coach")      # -> "coach_a"
#   mgr.get_assignment("patient_1", "dietitian")  # -> nil
#   mgr.get_patients("coach_a")                   # -> ["patient_1"]
#
#   # Part 2
#   mgr.add_member("coach_b", "coach", 1)
#   mgr.assign("patient_2", "coach_b", 2000.0)
#   mgr.assign("patient_3", "coach_b", 3000.0)  # raises CapacityError
#   mgr.available_members("coach")              # -> ["coach_a"]
#
#   # Part 3 — reassign patient_1 from coach_a to coach_b
#   mgr.assign("patient_1", "coach_b", 5000.0)
#   mgr.get_history("patient_1", "coach")
#   # -> [["coach_a", 1000.0, 5000.0], ["coach_b", 5000.0, nil]]
#   mgr.get_assignment_at("patient_1", "coach",  500.0)  # -> nil (before any assignment)
#   mgr.get_assignment_at("patient_1", "coach", 3000.0)  # -> "coach_a"
#   mgr.get_assignment_at("patient_1", "coach", 6000.0)  # -> "coach_b"
# =============================================================================

# Raised when assigning a patient to a member who is at their patient capacity.
class CapacityError < StandardError; end

class CareTeamManager
  def initialize
    raise NotImplementedError
  end

  # ── Part 1: Basic assignment and lookup ───────────────────────────────────

  # Register a care team member with the given role and patient capacity.
  def add_member(member_id, role, max_patients)
    raise NotImplementedError
  end

  # Assign a patient to a care team member (assigned_at is Unix seconds).
  #
  # A patient may have at most one assigned member per role at a time. If the
  # patient already has a member with the same role, that assignment is
  # replaced — the new assignment takes effect at assigned_at.
  #
  # Raises KeyError if member_id has not been registered via add_member.
  #
  # Part 2 addition: raises CapacityError if the member is already at
  # max_patients and the patient is not currently assigned to that exact
  # member. (Reassigning a patient who is already on this member does not
  # count as adding a new patient — it is a no-op for capacity purposes.)
  def assign(patient_id, member_id, assigned_at)
    raise NotImplementedError
  end

  # Return the member_id currently assigned to this patient for the given
  # role, or nil if no member of that role is currently assigned.
  def get_assignment(patient_id, role)
    raise NotImplementedError
  end

  # Return a sorted array of patient_ids currently assigned to this member.
  # Raises KeyError if member_id has not been registered.
  def get_patients(member_id)
    raise NotImplementedError
  end

  # ── Part 2: Capacity enforcement ──────────────────────────────────────────

  # Return a sorted array of member_ids with the given role that still have
  # open capacity (current patient count < max_patients).
  def available_members(role)
    raise NotImplementedError
  end

  # ── Part 3: Assignment history ────────────────────────────────────────────

  # Return the full assignment history for the patient's given role as an
  # array of [member_id, assigned_at, unassigned_at] triples sorted by
  # assigned_at.
  #
  # - unassigned_at is nil for the current (still-active) assignment.
  # - unassigned_at equals the assigned_at of the subsequent assignment for
  #   past entries.
  # - Returns [] if the patient has never been assigned a member of this role.
  def get_history(patient_id, role)
    raise NotImplementedError
  end

  # Return the member_id assigned to the patient for the given role at the
  # given timestamp, or nil if no assignment was active at that time.
  #
  # An assignment is active during the interval [assigned_at, unassigned_at).
  # The current assignment (unassigned_at is nil) is active from assigned_at
  # onward.
  #
  # Implement this by calling get_history — do not duplicate the lookup logic.
  def get_assignment_at(patient_id, role, timestamp)
    raise NotImplementedError
  end
end
