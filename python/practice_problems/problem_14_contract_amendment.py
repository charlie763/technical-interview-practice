"""
=============================================================================
INTERVIEW PROBLEM 14: Contract Amendment Manager
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You're building the amendment-tracking module for a contract platform.
After a contract is signed, its terms can be modified through formal amendments.
Each amendment specifies a set of field overrides that take effect from a given
date. To know the effective terms on any given date, you apply the base contract
fields and then overlay amendments in chronological order up to that date.

For this problem you are building a ContractAmendmentManager class.
Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between
ContractAmendmentManager instances — avoid them.
You choose the internal data structures; the public interface is what matters.

DATA MODEL
----------
Contract (base):
  {
    "contract_id":  str,
    "title":        str,
    "fields":       dict,  # e.g. {"value": 50000, "payment_terms": "net-30"}
  }

Amendment:
  {
    "amendment_id":  str,
    "contract_id":   str,
    "effective_on":  str,   # ISO-8601 date string ("YYYY-MM-DD")
    "overrides":     dict,  # field key → new value (may be a subset of fields)
    "note":          str,   # human-readable reason for the amendment
  }

Dates are ISO-8601 date strings (date-only).
Use datetime.date.fromisoformat() for comparisons.

# Example
# mgr = ContractAmendmentManager()
# mgr.add_contract("c-001", "Vendor MSA", fields={"value": 50000, "payment_terms": "net-30"})
#
# mgr.add_amendment("amd-1", "c-001", effective_on="2025-03-01",
#                   overrides={"payment_terms": "net-45"}, note="extended terms")
# mgr.add_amendment("amd-2", "c-001", effective_on="2025-06-01",
#                   overrides={"value": 75000}, note="scope increase")
#
# mgr.get_effective_contract("c-001", as_of_date="2025-01-01")
# # -> {"value": 50000, "payment_terms": "net-30"}   (no amendments yet)
#
# mgr.get_effective_contract("c-001", as_of_date="2025-04-15")
# # -> {"value": 50000, "payment_terms": "net-45"}   (amd-1 applied)
#
# mgr.get_effective_contract("c-001", as_of_date="2025-07-01")
# # -> {"value": 75000, "payment_terms": "net-45"}   (amd-1 + amd-2 applied)

=============================================================================
PART 1 — Base contract management
=============================================================================

Implement `add_contract` and `get_base_contract`.

"""


class ContractAmendmentManager:
    """
    Tracks contract base terms and their amendments, and resolves the effective
    contract state as of any given date.
    """

    def __init__(self):
        raise NotImplementedError

    # ── Part 1 ────────────────────────────────────────────────────────────────

    def add_contract(
        self,
        contract_id: str,
        title: str,
        fields: dict,
    ) -> dict:
        """
        Register a base contract.

        Parameters
        ----------
        contract_id : str
            Unique identifier.
        title : str
        fields : dict
            Initial field values (e.g. {"value": 50000, "payment_terms": "net-30"}).
            Store a copy — do not hold a reference to the caller's dict.

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

    def get_base_contract(self, contract_id: str) -> dict:
        """
        Return the base contract dict (original fields, no amendments applied).

        Raises
        ------
        KeyError
            If contract_id does not exist.
        """
        raise NotImplementedError

    # ── Part 2 ────────────────────────────────────────────────────────────────

    def add_amendment(
        self,
        amendment_id: str,
        contract_id: str,
        effective_on: str,
        overrides: dict,
        note: str,
    ) -> dict:
        """
        Register an amendment for a contract.

        Parameters
        ----------
        amendment_id : str
            Unique amendment identifier.
        contract_id : str
        effective_on : str
            ISO-8601 date string from which this amendment takes effect.
        overrides : dict
            Partial field updates. Keys may be a subset of the base contract
            fields, or introduce new fields.
            Store a copy — do not hold a reference to the caller's dict.
        note : str
            Human-readable reason.

        Returns
        -------
        dict
            The stored amendment dict.

        Raises
        ------
        ValueError
            If amendment_id already exists.
        KeyError
            If contract_id does not exist.
        """
        raise NotImplementedError

    def get_amendments(self, contract_id: str) -> list[dict]:
        """
        Return all amendments for a contract, sorted by effective_on ascending,
        then amendment_id ascending (for deterministic ordering when dates match).

        Parameters
        ----------
        contract_id : str

        Returns
        -------
        list[dict]
            List of amendment dicts.

        Raises
        ------
        KeyError
            If contract_id does not exist.
        """
        raise NotImplementedError

    def get_effective_contract(self, contract_id: str, as_of_date: str) -> dict:
        """
        Return the resolved field values for a contract as of as_of_date.

        Start with the base fields from `get_base_contract`, then apply
        amendments in chronological order (earliest first) where
        effective_on <= as_of_date, overlaying their overrides.

        Uses get_base_contract and get_amendments internally — do not
        duplicate their logic.

        Parameters
        ----------
        contract_id : str
        as_of_date : str
            ISO-8601 date string.

        Returns
        -------
        dict
            Resolved {field: value} mapping at as_of_date.

        Raises
        ------
        KeyError
            If contract_id does not exist.
        """
        raise NotImplementedError

    # ── Part 3 ────────────────────────────────────────────────────────────────

    def get_value_history(self, contract_id: str, field: str) -> list[dict]:
        """
        Return the full history of a specific field's value across base and
        all amendments that touched it, in chronological order.

        Parameters
        ----------
        contract_id : str
        field : str
            The field name to trace.

        Returns
        -------
        list[dict]
            Each item:
            {
              "effective_on":   str,          # ISO-8601 date; "base" for the original value
              "value":          any,
              "source":         str,          # "base" | amendment_id
            }
            Sorted by effective_on ascending (base always first).

        Raises
        ------
        KeyError
            If contract_id does not exist, or if the field is not present in
            the base contract or any amendment for that contract.
        """
        raise NotImplementedError

    def get_amendment_summary(self, contract_id: str) -> dict:
        """
        Return a summary of amendment activity for a contract.
        Uses get_amendments and get_effective_contract internally — do not
        duplicate their logic.

        Parameters
        ----------
        contract_id : str

        Returns
        -------
        dict
            {
              "contract_id":       str,
              "amendment_count":   int,
              "fields_amended":    list[str],   # unique field names ever overridden,
                                                # sorted ascending
              "latest_amendment":  str | None,  # ISO-8601 date of most recent
                                                # amendment, or None if no amendments
              "current_fields":    dict,        # result of get_effective_contract
                                                # called with today's date
            }
            For "today" use the most recent amendment's effective_on date if
            any amendments exist, otherwise use "2099-12-31" as a far-future
            sentinel so all amendments are included.

        Raises
        ------
        KeyError
            If contract_id does not exist.
        """
        raise NotImplementedError
