# =============================================================================
# INTERVIEW PROBLEM 12: Contract Expiration Alert Scheduler
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# You're building the alert-scheduling subsystem for a contract lifecycle
# management (CLM) platform used by legal and operations teams. Contracts
# have expiration dates, and stakeholders need to be notified days in
# advance so they can act before expiry.
#
# For this problem you are building a ContractAlertScheduler class. Store
# all state in instance variables initialized in `initialize`. Class-level
# variables will bleed between tests and between ContractAlertScheduler
# instances — avoid them. You choose the internal data structures; the
# public interface is what matters.
#
# DATA MODEL
# ----------
# Contract:
#   {
#     contract_id:  String,
#     title:        String,
#     owner_email:  String,
#     expires_on:   String,  # ISO-8601 date string, e.g. "2025-03-15"
#   }
#
# AlertConfig:
#   {
#     config_id:    String,
#     days_before:  Integer,  # how many days before expiry to trigger the alert
#     label:        String,   # e.g. "30-day notice", "final warning"
#   }
#
# SentRecord:
#   {
#     contract_id:  String,
#     config_id:    String,
#     sent_on:      String,  # ISO-8601 date string when the alert was sent
#   }
#
# Dates are ISO-8601 strings (date-only, no time component). Use Date.parse
# (from the "date" library) for arithmetic and comparisons.
#
# Example
# -------
#   scheduler = ContractAlertScheduler.new
#   scheduler.add_contract("c-001", "Vendor MSA", "legal@acme.com", "2025-06-30")
#   scheduler.add_alert_config("cfg-30", 30, "30-day notice")
#   scheduler.add_alert_config("cfg-7", 7, "final warning")
#
#   scheduler.get_contracts_expiring_between("2025-06-01", "2025-06-30")
#   # -> [{ contract_id: "c-001", title: "Vendor MSA", ... }]
#
#   scheduler.compute_alert_schedule("c-001")
#   # -> [
#   #      { config_id: "cfg-30", label: "30-day notice", alert_on: "2025-05-31" },
#   #      { config_id: "cfg-7",  label: "final warning",  alert_on: "2025-06-23" },
#   #    ]
#
#   scheduler.get_due_alerts("2025-06-01")
#   # -> [{ contract_id: "c-001", config_id: "cfg-30", alert_on: "2025-05-31", ... }]
#
# =============================================================================
# PART 1 — Contract and alert-config management
# =============================================================================
# Implement add_contract, add_alert_config, and get_contracts_expiring_between.

require "date"

class ContractAlertScheduler
  def initialize
    raise NotImplementedError
  end

  # ── Part 1 ────────────────────────────────────────────────────────────────

  # Register a contract and return the stored contract hash.
  # Raise ArgumentError if contract_id already exists.
  def add_contract(contract_id, title, owner_email, expires_on)
    raise NotImplementedError
  end

  # Register a global alert configuration and return the stored config hash.
  # days_before is the number of days before contract expiry to trigger the
  # alert.
  # Raise ArgumentError if config_id already exists.
  def add_alert_config(config_id, days_before, label)
    raise NotImplementedError
  end

  # Return all contracts whose expiration date falls within
  # [start_date, end_date], inclusive on both ends. Results are sorted by
  # expires_on ascending.
  def get_contracts_expiring_between(start_date, end_date)
    raise NotImplementedError
  end

  # ── Part 2 ────────────────────────────────────────────────────────────────

  # Compute the full alert schedule for a contract by applying every
  # registered alert config.
  #
  # For each AlertConfig, the alert fires on: expires_on - days_before days.
  #
  # Results are sorted by alert_on date ascending. Each item:
  #   { config_id: String, label: String, alert_on: String }
  #
  # Raise KeyError if contract_id does not exist.
  def compute_alert_schedule(contract_id)
    raise NotImplementedError
  end

  # Return all alert schedule entries whose alert_on date is on or before
  # as_of_date. Uses compute_alert_schedule internally — call it per
  # contract; do not duplicate its logic here.
  #
  # Results are sorted by alert_on ascending, then contract_id ascending.
  # Each item:
  #   {
  #     contract_id:  String,
  #     config_id:    String,
  #     label:        String,
  #     alert_on:     String,
  #     owner_email:  String,
  #     expires_on:   String,
  #   }
  def get_due_alerts(as_of_date)
    raise NotImplementedError
  end

  # ── Part 3 ────────────────────────────────────────────────────────────────

  # Record that an alert was sent for a specific contract/config pair.
  # Returns the SentRecord: { contract_id:, config_id:, sent_on: }.
  # Raise KeyError if contract_id or config_id does not exist.
  def record_alert_sent(contract_id, config_id, sent_on)
    raise NotImplementedError
  end

  # Return the alert schedule for a contract, enriched with a "sent" flag
  # indicating whether that alert has already been sent.
  #
  # Uses compute_alert_schedule and record_alert_sent state internally.
  # Exclude alerts whose alert_on is strictly before as_of_date (they are
  # in the past).
  #
  # Each item (sorted by alert_on ascending):
  #   { config_id: String, label: String, alert_on: String, sent: true | false }
  #
  # Raise KeyError if contract_id does not exist.
  def get_upcoming_alerts(contract_id, as_of_date)
    raise NotImplementedError
  end
end
