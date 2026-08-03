"""
=============================================================================
INTERVIEW PROBLEM 10: Responder Dispatch Manager
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You're building the dispatch assignment layer for an emergency-response
platform. Incident alerts stream in and need to be routed to available field
responders. Responders specialize in certain incident types and have a
capacity limit — the maximum number of simultaneous open (unresolved)
incidents they can handle.

For this problem you are building a DispatchManager class.
Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between DispatchManager
instances — avoid them.
You choose the internal data structures; the public interface is what matters.

DATA MODEL
----------
Responder:
  {
    "responder_id":     str,
    "name":             str,
    "subscribed_types": list[str],  # incident types this responder handles
    "capacity":         int,        # max simultaneous open assignments
  }

Incident:
  {
    "incident_id":   str,
    "incident_type": str,         # e.g. "shooting", "car-crash", "fire"
    "severity":      int,         # 1 (low) – 5 (critical)
    "ts":            str,         # ISO-8601 timestamp, when reported
    "responder_id":  str | None,  # None until assigned
    "resolved":      bool,        # False until resolve_incident is called
  }

# Example
# dm = DispatchManager()
# dm.register_responder("unit-12", "Alpha Team",
#                        subscribed_types=["shooting", "robbery"], capacity=3)
# dm.register_responder("unit-14", "Beta Team",
#                        subscribed_types=["car-crash", "fire"], capacity=2)
# dm.add_incident("inc-001", "shooting", severity=5, ts="2024-01-01T10:00:00")
# dm.add_incident("inc-002", "car-crash", severity=3, ts="2024-01-01T10:01:00")
# dm.get_incidents_for_responder("unit-12")   # -> [inc-001 dict]
# dm.assign_incident("inc-001", "unit-12")
# dm.get_open_assignments("unit-12")          # -> [inc-001 dict]
# dm.auto_assign("inc-002")                   # -> "unit-14"
=============================================================================
"""

from typing import Optional


class DispatchManager:
    def __init__(self):
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 1 — Registration and basic queries  (~10 min)
    # -------------------------------------------------------------------------

    def register_responder(
        self,
        responder_id: str,
        name: str,
        subscribed_types: list,
        capacity: int,
    ) -> dict:
        """
        Register a new responder and return it.
        Raise ValueError if responder_id already exists.
        """
        raise NotImplementedError

    def add_incident(
        self,
        incident_id: str,
        incident_type: str,
        severity: int,
        ts: str,
    ) -> dict:
        """
        Add a new incident (unassigned, unresolved) and return it.
        Raise ValueError if incident_id already exists.
        """
        raise NotImplementedError

    def get_incidents_for_responder(self, responder_id: str) -> list:
        """
        Return all incidents whose incident_type appears in the responder's
        subscribed_types list, regardless of whether the incident has been
        assigned yet.

        Sort order: severity descending (5 first), then ts ascending (oldest
        first within the same severity).

        Raise KeyError if responder_id does not exist.
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 2 — Assignment and resolution  (~15 min)
    # -------------------------------------------------------------------------

    def assign_incident(self, incident_id: str, responder_id: str) -> None:
        """
        Assign an incident to a responder.
        - Raise KeyError   if incident_id or responder_id does not exist.
        - Raise ValueError if the incident already has a responder assigned.
        - Raise ValueError if the responder is at capacity.  A responder is at
          capacity when their count of open assignments (assigned + not yet
          resolved) equals their capacity.
        - On success, sets incident["responder_id"] = responder_id.
        """
        raise NotImplementedError

    def resolve_incident(self, incident_id: str) -> None:
        """
        Mark an incident as resolved (sets resolved = True), freeing the
        assigned responder's capacity slot.
        - Raise KeyError   if incident_id does not exist.
        - Raise ValueError if the incident is already resolved.
        """
        raise NotImplementedError

    def get_open_assignments(self, responder_id: str) -> list:
        """
        Return all incidents that are assigned to this responder and not yet
        resolved.
        Sort order: severity descending, then ts ascending.
        Raise KeyError if responder_id does not exist.
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 3 — Auto-assignment  (~20 min)
    # -------------------------------------------------------------------------

    def auto_assign(self, incident_id: str) -> str:
        """
        Automatically assign an incident to the best available responder:

        Eligibility (both must hold):
          1. The responder's subscribed_types includes the incident's
             incident_type.
          2. The responder's current open-assignment count is less than their
             capacity.

        Selection — among eligible responders, prefer:
          1. Fewest open assignments (least loaded).
          2. Tie-break: highest capacity (largest capacity value).
          3. Tie-break: responder_id lexicographically ascending.

        - Raise KeyError   if incident_id does not exist.
        - Raise ValueError if the incident is already assigned.
        - Raise ValueError if no eligible responder is available.

        Call assign_incident to perform the assignment and return the
        responder_id of the chosen responder.
        """
        raise NotImplementedError

    def get_dispatch_summary(self) -> list:
        """
        Return a list of dicts — one per registered responder — with fields:
          "responder_id":       str
          "name":               str
          "capacity":           int
          "open_count":         int  (current open assignments)
          "available_capacity": int  (capacity - open_count)

        Sorted by responder_id ascending.
        """
        raise NotImplementedError
