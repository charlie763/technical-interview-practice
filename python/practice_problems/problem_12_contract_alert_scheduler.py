"""
=============================================================================
INTERVIEW PROBLEM 12: Contract Expiration Alert Scheduler
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You're building the alert-scheduling subsystem for a contract lifecycle
management (CLM) platform used by legal and operations teams. Contracts have
expiration dates, and stakeholders need to be notified days in advance so they
can act before expiry.

For this problem you are building a ContractAlertScheduler class.
Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between ContractAlertScheduler
instances — avoid them.
You choose the internal data structures; the public interface is what matters.

DATA MODEL
----------
Contract:
  {
    "contract_id":  str,
    "title":        str,
    "owner_email":  str,
    "expires_on":   str,  # ISO-8601 date string, e.g. "2025-03-15"
  }

AlertConfig:
  {
    "config_id":    str,
    "days_before":  int,  # how many days before expiry to trigger the alert
    "label":        str,  # e.g. "30-day notice", "final warning"
  }

SentRecord:
  {
    "contract_id":  str,
    "config_id":    str,
    "sent_on":      str,  # ISO-8601 date string when the alert was sent
  }

Dates are ISO-8601 strings (date-only, no time component).
Use datetime.date.fromisoformat() for arithmetic and comparisons.

# Example
# scheduler = ContractAlertScheduler()
# scheduler.add_contract("c-001", "Vendor MSA", "legal@acme.com", expires_on="2025-06-30")
# scheduler.add_alert_config("cfg-30", days_before=30, label="30-day notice")
# scheduler.add_alert_config("cfg-7",  days_before=7,  label="final warning")
#
# scheduler.get_contracts_expiring_between("2025-06-01", "2025-06-30")
# # -> [{"contract_id": "c-001", "title": "Vendor MSA", ...}]
#
# scheduler.compute_alert_schedule("c-001")
# # -> [
# #      {"config_id": "cfg-30", "label": "30-day notice", "alert_on": "2025-05-31"},
# #      {"config_id": "cfg-7",  "label": "final warning",  "alert_on": "2025-06-23"},
# #    ]
#
# scheduler.get_due_alerts(as_of_date="2025-06-01")
# # -> [{"contract_id": "c-001", "config_id": "cfg-30", "alert_on": "2025-05-31", ...}]

=============================================================================
PART 1 — Contract and alert-config management
=============================================================================

Implement `add_contract`, `add_alert_config`, and `get_contracts_expiring_between`.

"""
from datetime import date


class ContractAlertScheduler:
    """
    Schedules and tracks expiration alerts for contracts.
    """

    def __init__(self):
        raise NotImplementedError

    # ── Part 1 ────────────────────────────────────────────────────────────────

    def add_contract(
        self,
        contract_id: str,
        title: str,
        owner_email: str,
        expires_on: str,
    ) -> dict:
        """
        Register a contract.

        Parameters
        ----------
        contract_id : str
            Unique identifier.
        title : str
            Human-readable contract name.
        owner_email : str
            Primary point of contact.
        expires_on : str
            Expiration date in ISO-8601 date format ("YYYY-MM-DD").

        Returns
        -------
        dict
            The stored contract dict.

        Raises
        ------
        ValueError
            If contract_id already exists.
        """
        raise NotImplementedError

    def add_alert_config(self, config_id: str, days_before: int, label: str) -> dict:
        """
        Register a global alert configuration.

        Parameters
        ----------
        config_id : str
            Unique identifier.
        days_before : int
            Number of days before contract expiry to trigger the alert.
        label : str
            Human-readable description (e.g. "30-day notice").

        Returns
        -------
        dict
            The stored alert config dict.

        Raises
        ------
        ValueError
            If config_id already exists.
        """
        raise NotImplementedError

    def get_contracts_expiring_between(self, start_date: str, end_date: str) -> list[dict]:
        """
        Return all contracts whose expiration date falls within [start_date, end_date],
        inclusive on both ends. Results are sorted by expires_on ascending.

        Parameters
        ----------
        start_date : str
            ISO-8601 date string.
        end_date : str
            ISO-8601 date string.

        Returns
        -------
        list[dict]
            List of matching contract dicts, sorted by expires_on ascending.
        """
        raise NotImplementedError

    # ── Part 2 ────────────────────────────────────────────────────────────────

    def compute_alert_schedule(self, contract_id: str) -> list[dict]:
        """
        Compute the full alert schedule for a contract by applying every
        registered alert config.

        For each AlertConfig, the alert fires on:
            expires_on - timedelta(days=days_before)

        Results are sorted by alert_on date ascending.

        Parameters
        ----------
        contract_id : str

        Returns
        -------
        list[dict]
            Each item:
            {
              "config_id": str,
              "label":     str,
              "alert_on":  str,   # ISO-8601 date string
            }

        Raises
        ------
        KeyError
            If contract_id does not exist.
        """
        raise NotImplementedError

    def get_due_alerts(self, as_of_date: str) -> list[dict]:
        """
        Return all alert schedule entries whose alert_on date is on or before
        as_of_date. Uses compute_alert_schedule internally — call it per contract;
        do not duplicate its logic here.

        Results are sorted by alert_on ascending, then contract_id ascending.

        Parameters
        ----------
        as_of_date : str
            ISO-8601 date string representing "today".

        Returns
        -------
        list[dict]
            Each item:
            {
              "contract_id":  str,
              "config_id":    str,
              "label":        str,
              "alert_on":     str,
              "owner_email":  str,
              "expires_on":   str,
            }
        """
        raise NotImplementedError

    # ── Part 3 ────────────────────────────────────────────────────────────────

    def record_alert_sent(self, contract_id: str, config_id: str, sent_on: str) -> dict:
        """
        Record that an alert was sent for a specific contract/config pair.

        Parameters
        ----------
        contract_id : str
        config_id : str
        sent_on : str
            ISO-8601 date string.

        Returns
        -------
        dict
            The SentRecord: {"contract_id", "config_id", "sent_on"}.

        Raises
        ------
        KeyError
            If contract_id or config_id does not exist.
        """
        raise NotImplementedError

    def get_upcoming_alerts(self, contract_id: str, as_of_date: str) -> list[dict]:
        """
        Return the alert schedule for a contract, enriched with a "sent" flag
        indicating whether that alert has already been sent.

        Uses compute_alert_schedule and record_alert_sent state internally.

        Parameters
        ----------
        contract_id : str
        as_of_date : str
            ISO-8601 date string. Exclude alerts whose alert_on is strictly
            before as_of_date (they are in the past).

        Returns
        -------
        list[dict]
            Each item (sorted by alert_on ascending):
            {
              "config_id":  str,
              "label":      str,
              "alert_on":   str,
              "sent":       bool,
            }

        Raises
        ------
        KeyError
            If contract_id does not exist.
        """
        raise NotImplementedError
