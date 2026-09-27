# =============================================================================
# INTERVIEW PROBLEM 14: Contract Amendment Manager
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# You're building the amendment-tracking module for a contract platform.
# After a contract is signed, its terms can be modified through formal
# amendments. Each amendment specifies a set of field overrides that take
# effect from a given date. To know the effective terms on any given date,
# you apply the base contract fields and then overlay amendments in
# chronological order up to that date.
#
# For this problem you are building a ContractAmendmentManager class.
# Store all state in instance variables initialized in `initialize`.
# Class-level variables will bleed between tests and between
# ContractAmendmentManager instances — avoid them. You choose the internal
# data structures; the public interface is what matters.
#
# DATA MODEL
# ----------
# Contract (base):
#   {
#     contract_id:  String,
#     title:        String,
#     fields:       Hash,  # e.g. { "value" => 50000, "payment_terms" => "net-30" }
#   }
#
# Amendment:
#   {
#     amendment_id:  String,
#     contract_id:   String,
#     effective_on:  String,  # ISO-8601 date string ("YYYY-MM-DD")
#     overrides:     Hash,    # field key -> new value (may be a subset of fields)
#     note:          String,  # human-readable reason for the amendment
#   }
#
# Dates are ISO-8601 date strings (date-only). Plain string comparison is
# sufficient since ISO-8601 dates sort lexicographically.
#
# Example
# -------
#   mgr = ContractAmendmentManager.new
#   mgr.add_contract("c-001", "Vendor MSA", { "value" => 50000, "payment_terms" => "net-30" })
#
#   mgr.add_amendment("amd-1", "c-001", "2025-03-01",
#                      { "payment_terms" => "net-45" }, "extended terms")
#   mgr.add_amendment("amd-2", "c-001", "2025-06-01",
#                      { "value" => 75000 }, "scope increase")
#
#   mgr.get_effective_contract("c-001", "2025-01-01")
#   # -> { "value" => 50000, "payment_terms" => "net-30" }   (no amendments yet)
#
#   mgr.get_effective_contract("c-001", "2025-04-15")
#   # -> { "value" => 50000, "payment_terms" => "net-45" }   (amd-1 applied)
#
#   mgr.get_effective_contract("c-001", "2025-07-01")
#   # -> { "value" => 75000, "payment_terms" => "net-45" }   (amd-1 + amd-2 applied)
#
# =============================================================================
# PART 1 — Base contract management
# =============================================================================
# Implement add_contract and get_base_contract.

class ContractAmendmentManager
  def initialize
    raise NotImplementedError
  end

  # ── Part 1 ────────────────────────────────────────────────────────────────

  # Register a base contract.
  #
  # fields: initial field values (e.g. { "value" => 50000 }). Store a copy —
  # do not hold a reference to the caller's hash.
  #
  # Returns the stored contract hash.
  # Raise ArgumentError if contract_id already exists.
  def add_contract(contract_id, title, fields)
    raise NotImplementedError
  end

  # Return the base contract hash (original fields, no amendments applied).
  # Raise KeyError if contract_id does not exist.
  def get_base_contract(contract_id)
    raise NotImplementedError
  end

  # ── Part 2 ────────────────────────────────────────────────────────────────

  # Register an amendment for a contract.
  #
  # overrides: partial field updates. Keys may be a subset of the base
  # contract fields, or introduce new fields. Store a copy — do not hold a
  # reference to the caller's hash.
  #
  # Returns the stored amendment hash.
  # Raise ArgumentError if amendment_id already exists.
  # Raise KeyError if contract_id does not exist.
  def add_amendment(amendment_id, contract_id, effective_on, overrides, note)
    raise NotImplementedError
  end

  # Return all amendments for a contract, sorted by effective_on ascending,
  # then amendment_id ascending (for deterministic ordering when dates match).
  # Raise KeyError if contract_id does not exist.
  def get_amendments(contract_id)
    raise NotImplementedError
  end

  # Return the resolved field values for a contract as of as_of_date.
  #
  # Start with the base fields from get_base_contract, then apply
  # amendments in chronological order (earliest first) where
  # effective_on <= as_of_date, overlaying their overrides.
  #
  # Uses get_base_contract and get_amendments internally — do not duplicate
  # their logic.
  #
  # Raise KeyError if contract_id does not exist.
  def get_effective_contract(contract_id, as_of_date)
    raise NotImplementedError
  end

  # ── Part 3 ────────────────────────────────────────────────────────────────

  # Return the full history of a specific field's value across base and all
  # amendments that touched it, in chronological order.
  #
  # Each item:
  #   {
  #     effective_on:  String,  # ISO-8601 date; "base" for the original value
  #     value:         Object,
  #     source:        String,  # "base" | amendment_id
  #   }
  # Sorted by effective_on ascending (base always first).
  #
  # Raise KeyError if contract_id does not exist, or if the field is not
  # present in the base contract or any amendment for that contract.
  def get_value_history(contract_id, field)
    raise NotImplementedError
  end

  # Return a summary of amendment activity for a contract. Uses
  # get_amendments and get_effective_contract internally — do not duplicate
  # their logic.
  #
  #   {
  #     contract_id:       String,
  #     amendment_count:   Integer,
  #     fields_amended:    Array[String],  # unique field names ever overridden,
  #                                        # sorted ascending
  #     latest_amendment:  String | nil,   # ISO-8601 date of most recent
  #                                        # amendment, or nil if no amendments
  #     current_fields:    Hash,           # result of get_effective_contract
  #                                        # called with today's date
  #   }
  #   For "today" use the most recent amendment's effective_on date if any
  #   amendments exist, otherwise use "2099-12-31" as a far-future sentinel
  #   so all amendments are included.
  #
  # Raise KeyError if contract_id does not exist.
  def get_amendment_summary(contract_id)
    raise NotImplementedError
  end
end
