# =============================================================================
# INTERVIEW PROBLEM 10: Responder Dispatch Manager
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# You're building the dispatch assignment layer for an emergency-response
# platform. Incident alerts stream in and need to be routed to available
# field responders. Responders specialize in certain incident types and
# have a capacity limit — the maximum number of simultaneous open
# (unresolved) incidents they can handle.
#
# For this problem you are building a DispatchManager class. Store all
# state in instance variables initialized in `initialize`. Class-level
# variables will bleed between tests and between DispatchManager instances
# — avoid them. You choose the internal data structures; the public
# interface is what matters.
#
# DATA MODEL
# ----------
# Responder:
#   {
#     responder_id:     String,
#     name:              String,
#     subscribed_types:  Array[String],  # incident types this responder handles
#     capacity:          Integer,        # max simultaneous open assignments
#   }
#
# Incident:
#   {
#     incident_id:   String,
#     incident_type: String,       # e.g. "shooting", "car-crash", "fire"
#     severity:      Integer,      # 1 (low) - 5 (critical)
#     ts:            String,       # ISO-8601 timestamp, when reported
#     responder_id:  String | nil, # nil until assigned
#     resolved:      true | false, # false until resolve_incident is called
#   }
#
# Example
# -------
#   dm = DispatchManager.new
#   dm.register_responder("unit-12", "Alpha Team", ["shooting", "robbery"], 3)
#   dm.register_responder("unit-14", "Beta Team", ["car-crash", "fire"], 2)
#   dm.add_incident("inc-001", "shooting", 5, "2024-01-01T10:00:00")
#   dm.add_incident("inc-002", "car-crash", 3, "2024-01-01T10:01:00")
#   dm.get_incidents_for_responder("unit-12")   # -> [inc-001 hash]
#   dm.assign_incident("inc-001", "unit-12")
#   dm.get_open_assignments("unit-12")          # -> [inc-001 hash]
#   dm.auto_assign("inc-002")                   # -> "unit-14"
# =============================================================================

class DispatchManager
  def initialize
    raise NotImplementedError
  end

  # -------------------------------------------------------------------------
  # PART 1 — Registration and basic queries (~10 min)
  # -------------------------------------------------------------------------

  # Register a new responder and return it.
  # Raise ArgumentError if responder_id already exists.
  def register_responder(responder_id, name, subscribed_types, capacity)
    raise NotImplementedError
  end

  # Add a new incident (unassigned, unresolved) and return it.
  # Raise ArgumentError if incident_id already exists.
  def add_incident(incident_id, incident_type, severity, ts)
    raise NotImplementedError
  end

  # Return all incidents whose incident_type appears in the responder's
  # subscribed_types list, regardless of whether the incident has been
  # assigned yet.
  #
  # Sort order: severity descending (5 first), then ts ascending (oldest
  # first within the same severity).
  #
  # Raise KeyError if responder_id does not exist.
  def get_incidents_for_responder(responder_id)
    raise NotImplementedError
  end

  # -------------------------------------------------------------------------
  # PART 2 — Assignment and resolution (~15 min)
  # -------------------------------------------------------------------------

  # Assign an incident to a responder.
  # - Raise KeyError if incident_id or responder_id does not exist.
  # - Raise ArgumentError if the incident already has a responder assigned.
  # - Raise ArgumentError if the responder is at capacity. A responder is at
  #   capacity when their count of open assignments (assigned + not yet
  #   resolved) equals their capacity.
  # - On success, sets incident[:responder_id] = responder_id.
  def assign_incident(incident_id, responder_id)
    raise NotImplementedError
  end

  # Mark an incident as resolved (sets resolved = true), freeing the
  # assigned responder's capacity slot.
  # - Raise KeyError if incident_id does not exist.
  # - Raise ArgumentError if the incident is already resolved.
  def resolve_incident(incident_id)
    raise NotImplementedError
  end

  # Return all incidents that are assigned to this responder and not yet
  # resolved.
  # Sort order: severity descending, then ts ascending.
  # Raise KeyError if responder_id does not exist.
  def get_open_assignments(responder_id)
    raise NotImplementedError
  end

  # -------------------------------------------------------------------------
  # PART 3 — Auto-assignment (~20 min)
  # -------------------------------------------------------------------------

  # Automatically assign an incident to the best available responder:
  #
  # Eligibility (both must hold):
  #   1. The responder's subscribed_types includes the incident's
  #      incident_type.
  #   2. The responder's current open-assignment count is less than their
  #      capacity.
  #
  # Selection — among eligible responders, prefer:
  #   1. Fewest open assignments (least loaded).
  #   2. Tie-break: highest capacity (largest capacity value).
  #   3. Tie-break: responder_id lexicographically ascending.
  #
  # - Raise KeyError if incident_id does not exist.
  # - Raise ArgumentError if the incident is already assigned.
  # - Raise ArgumentError if no eligible responder is available.
  #
  # Call assign_incident to perform the assignment and return the
  # responder_id of the chosen responder.
  def auto_assign(incident_id)
    raise NotImplementedError
  end

  # Return an array of hashes — one per registered responder — with fields:
  #   responder_id:        String
  #   name:                String
  #   capacity:             Integer
  #   open_count:           Integer  (current open assignments)
  #   available_capacity:   Integer  (capacity - open_count)
  #
  # Sorted by responder_id ascending.
  def get_dispatch_summary
    raise NotImplementedError
  end
end
