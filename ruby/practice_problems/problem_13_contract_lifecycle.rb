# =============================================================================
# INTERVIEW PROBLEM 13: Contract Lifecycle State Machine
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# You're building the lifecycle management module for a contract platform.
# Every contract moves through a defined set of states from creation to
# completion or termination. Only specific transitions are allowed —
# illegal transitions must be rejected. All changes are audit-logged.
#
# For this problem you are building a ContractLifecycle class. Store all
# state in instance variables initialized in `initialize`. Class-level
# variables will bleed between tests and between ContractLifecycle
# instances — avoid them. You choose the internal data structures; the
# public interface is what matters.
#
# DATA MODEL
# ----------
# Contract:
#   {
#     contract_id:  String,
#     title:        String,
#     state:        String,  # current lifecycle state (see STATE MACHINE below)
#     created_at:   String,  # ISO-8601 datetime when the contract was created
#     fields:       Hash,    # arbitrary key/value metadata
#   }
#
# AuditEntry:
#   {
#     contract_id:  String,
#     from_state:   String | nil,  # nil for the initial "created" entry
#     to_state:     String,
#     at:           String,        # ISO-8601 datetime of the transition
#     actor:        String,        # user or system identifier
#   }
#
# STATE MACHINE
# -------------
# Valid states and allowed forward transitions:
#
#   draft          -> in_review
#   in_review      -> approved | draft            (can be sent back for revisions)
#   approved       -> executed
#   executed       -> active
#   active         -> expiring_soon | terminated
#   expiring_soon  -> expired | active | terminated
#   expired        -> (terminal — no outgoing transitions)
#   terminated     -> (terminal — no outgoing transitions)
#
# Timestamps are ISO-8601 strings without timezone offset. Use Time.parse
# (from the "time" library) for arithmetic.
#
# Example
# -------
#   cl = ContractLifecycle.new
#   cl.create_contract("c-001", "Vendor MSA", "2025-01-01T09:00:00", "alice")
#   cl.get_contract("c-001")[:state]  # -> "draft"
#   cl.transition("c-001", "in_review", "2025-01-02T10:00:00", "alice")
#   cl.transition("c-001", "approved", "2025-01-03T11:00:00", "bob")
#   cl.get_contract("c-001")[:state]  # -> "approved"
#   cl.transition("c-001", "draft", "2025-01-04T09:00:00", "bob")
#   # raises ArgumentError — approved -> draft is not a valid transition
#
# =============================================================================
# PART 1 — Contract creation, field management, and transitions
# =============================================================================
# Implement create_contract, set_field, get_contract, and transition.

require "time"

# Valid forward transitions from each state.
VALID_TRANSITIONS = {
  "draft" => ["in_review"],
  "in_review" => %w[approved draft],
  "approved" => ["executed"],
  "executed" => ["active"],
  "active" => %w[expiring_soon terminated],
  "expiring_soon" => %w[expired active terminated],
  "expired" => [],
  "terminated" => [],
}.freeze

CONTRACT_TERMINAL_STATES = %w[expired terminated].freeze

class ContractLifecycle
  def initialize
    raise NotImplementedError
  end

  # ── Part 1 ────────────────────────────────────────────────────────────────

  # Create a new contract in the "draft" state and record the initial audit
  # entry (from_state: nil, to_state: "draft").
  # Returns the stored contract hash (fields starts empty).
  # Raise ArgumentError if contract_id already exists.
  def create_contract(contract_id, title, created_at, actor)
    raise NotImplementedError
  end

  # Set or update a field on the contract's `fields` hash.
  # Returns the updated contract hash.
  # Raise KeyError if contract_id does not exist.
  def set_field(contract_id, key, value)
    raise NotImplementedError
  end

  # Return the contract hash. Raise KeyError if contract_id does not exist.
  def get_contract(contract_id)
    raise NotImplementedError
  end

  # Move a contract to a new state if the transition is valid, and append
  # an AuditEntry. Returns the updated contract hash.
  #
  # Raise KeyError if contract_id does not exist.
  # Raise ArgumentError if the transition from the current state to
  # to_state is not allowed (consult VALID_TRANSITIONS).
  def transition(contract_id, to_state, at, actor)
    raise NotImplementedError
  end

  # ── Part 2 ────────────────────────────────────────────────────────────────

  # Return the full ordered audit trail for a contract (oldest first), as an
  # array of AuditEntry hashes.
  # Raise KeyError if contract_id does not exist.
  def get_audit_trail(contract_id)
    raise NotImplementedError
  end

  # Return all contracts currently in the given state, sorted by
  # contract_id ascending.
  def get_contracts_by_state(state)
    raise NotImplementedError
  end

  # Attempt to transition each contract in contract_ids to to_state. Calls
  # transition internally — do not duplicate its logic.
  #
  # Continue processing remaining contracts even if one fails; collect all
  # failures.
  #
  # Returns:
  #   {
  #     succeeded: [contract_id, ...],  # successfully transitioned
  #     failed: [                        # failed transitions
  #       { contract_id: String, reason: String },
  #       ...
  #     ],
  #   }
  def bulk_advance(contract_ids, to_state, at, actor)
    raise NotImplementedError
  end

  # ── Part 3 ────────────────────────────────────────────────────────────────

  # Return aggregate counts across all contracts:
  #   {
  #     total:          Integer,
  #     by_state:       { state => count, ... },  # only states with count > 0
  #     terminal_count: Integer,  # contracts in "expired" or "terminated"
  #   }
  def get_lifecycle_metrics
    raise NotImplementedError
  end

  # Return contracts that have been stuck in the same non-terminal state
  # for more than 30 days without any transition.
  #
  # "Stuck since" is the `at` timestamp of the most recent AuditEntry for
  # the contract.
  #
  # Uses get_audit_trail internally — do not duplicate its logic.
  #
  # Each item (sorted by days_stuck descending):
  #   {
  #     contract_id:  String,
  #     title:        String,
  #     state:        String,
  #     stuck_since:  String,   # ISO-8601 datetime of last transition
  #     days_stuck:   Integer,
  #   }
  def get_overdue_contracts(as_of)
    raise NotImplementedError
  end
end
