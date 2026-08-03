"""
=============================================================================
INTERVIEW PROBLEM 13: Contract Lifecycle State Machine
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You're building the lifecycle management module for a contract platform.
Every contract moves through a defined set of states from creation to
completion or termination. Only specific transitions are allowed — illegal
transitions must be rejected. All changes are audit-logged.

For this problem you are building a ContractLifecycle class.
Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between ContractLifecycle
instances — avoid them.
You choose the internal data structures; the public interface is what matters.

DATA MODEL
----------
Contract:
  {
    "contract_id":  str,
    "title":        str,
    "state":        str,     # current lifecycle state (see STATE MACHINE below)
    "created_at":   str,     # ISO-8601 datetime when the contract was created
    "fields":       dict,    # arbitrary key/value metadata
  }

AuditEntry:
  {
    "contract_id":  str,
    "from_state":   str | None,   # None for the initial "created" entry
    "to_state":     str,
    "at":           str,          # ISO-8601 datetime of the transition
    "actor":        str,          # user or system identifier
  }

STATE MACHINE
-------------
Valid states and allowed forward transitions:

  draft  →  in_review
  in_review  →  approved | draft     (can be sent back for revisions)
  approved  →  executed
  executed  →  active
  active  →  expiring_soon | terminated
  expiring_soon  →  expired | active | terminated
  expired  →  (terminal — no outgoing transitions)
  terminated  →  (terminal — no outgoing transitions)

Timestamps are ISO-8601 strings without timezone offset.
Use datetime.fromisoformat() for arithmetic.

# Example
# cl = ContractLifecycle()
# cl.create_contract("c-001", "Vendor MSA", created_at="2025-01-01T09:00:00", actor="alice")
# cl.get_contract("c-001")["state"]  # -> "draft"
# cl.transition("c-001", "in_review", at="2025-01-02T10:00:00", actor="alice")
# cl.transition("c-001", "approved",  at="2025-01-03T11:00:00", actor="bob")
# cl.get_contract("c-001")["state"]  # -> "approved"
# cl.transition("c-001", "draft", at="2025-01-04T09:00:00", actor="bob")
# # raises ValueError — approved → draft is not a valid transition

=============================================================================
PART 1 — Contract creation, field management, and transitions
=============================================================================

Implement `create_contract`, `set_field`, `get_contract`, and `transition`.

"""


# Valid forward transitions from each state.
VALID_TRANSITIONS: dict[str, list[str]] = {
    "draft":         ["in_review"],
    "in_review":     ["approved", "draft"],
    "approved":      ["executed"],
    "executed":      ["active"],
    "active":        ["expiring_soon", "terminated"],
    "expiring_soon": ["expired", "active", "terminated"],
    "expired":       [],
    "terminated":    [],
}


class ContractLifecycle:
    """
    Manages lifecycle state transitions and audit history for contracts.
    """

    def __init__(self):
        raise NotImplementedError

    # ── Part 1 ────────────────────────────────────────────────────────────────

    def create_contract(
        self,
        contract_id: str,
        title: str,
        created_at: str,
        actor: str,
    ) -> dict:
        """
        Create a new contract in the "draft" state and record the initial
        audit entry (from_state=None, to_state="draft").

        Parameters
        ----------
        contract_id : str
            Unique identifier.
        title : str
        created_at : str
            ISO-8601 datetime.
        actor : str
            User/system creating the contract.

        Returns
        -------
        dict
            The stored contract dict (fields starts empty).

        Raises
        ------
        ValueError
            If contract_id already exists.
        """
        raise NotImplementedError

    def set_field(self, contract_id: str, key: str, value) -> dict:
        """
        Set or update a field on the contract's `fields` dict.

        Parameters
        ----------
        contract_id : str
        key : str
        value : any JSON-serialisable value.

        Returns
        -------
        dict
            The updated contract dict.

        Raises
        ------
        KeyError
            If contract_id does not exist.
        """
        raise NotImplementedError

    def get_contract(self, contract_id: str) -> dict:
        """
        Return the contract dict.

        Raises
        ------
        KeyError
            If contract_id does not exist.
        """
        raise NotImplementedError

    def transition(
        self,
        contract_id: str,
        to_state: str,
        at: str,
        actor: str,
    ) -> dict:
        """
        Move a contract to a new state if the transition is valid, and append
        an AuditEntry.

        Parameters
        ----------
        contract_id : str
        to_state : str
            Target state.
        at : str
            ISO-8601 datetime of the transition.
        actor : str

        Returns
        -------
        dict
            The updated contract dict.

        Raises
        ------
        KeyError
            If contract_id does not exist.
        ValueError
            If the transition from the current state to to_state is not allowed
            (consult VALID_TRANSITIONS).
        """
        raise NotImplementedError

    # ── Part 2 ────────────────────────────────────────────────────────────────

    def get_audit_trail(self, contract_id: str) -> list[dict]:
        """
        Return the full ordered audit trail for a contract (oldest first).

        Parameters
        ----------
        contract_id : str

        Returns
        -------
        list[dict]
            List of AuditEntry dicts in chronological order.

        Raises
        ------
        KeyError
            If contract_id does not exist.
        """
        raise NotImplementedError

    def get_contracts_by_state(self, state: str) -> list[dict]:
        """
        Return all contracts currently in the given state, sorted by
        contract_id ascending.

        Parameters
        ----------
        state : str

        Returns
        -------
        list[dict]
        """
        raise NotImplementedError

    def bulk_advance(
        self,
        contract_ids: list[str],
        to_state: str,
        at: str,
        actor: str,
    ) -> dict:
        """
        Attempt to transition each contract in contract_ids to to_state.
        Calls `transition` internally — do not duplicate its logic.

        Continue processing remaining contracts even if one fails; collect all
        failures.

        Parameters
        ----------
        contract_ids : list[str]
        to_state : str
        at : str
            ISO-8601 datetime applied to all transitions.
        actor : str

        Returns
        -------
        dict
            {
              "succeeded": [contract_id, ...],   # successfully transitioned
              "failed":    [                      # failed transitions
                {
                  "contract_id": str,
                  "reason":      str,  # error message
                },
                ...
              ],
            }
        """
        raise NotImplementedError

    # ── Part 3 ────────────────────────────────────────────────────────────────

    def get_lifecycle_metrics(self) -> dict:
        """
        Return aggregate counts across all contracts.

        Returns
        -------
        dict
            {
              "total":           int,
              "by_state":        {state: count, ...},  # only states with count > 0
              "terminal_count":  int,   # contracts in "expired" or "terminated"
            }
        """
        raise NotImplementedError

    def get_overdue_contracts(self, as_of: str) -> list[dict]:
        """
        Return contracts that have been stuck in the same non-terminal state
        for more than 30 days without any transition.

        "Stuck since" is the `at` timestamp of the most recent AuditEntry for
        the contract.

        Uses get_audit_trail internally — do not duplicate its logic.

        Parameters
        ----------
        as_of : str
            ISO-8601 datetime to measure staleness against.

        Returns
        -------
        list[dict]
            Each item:
            {
              "contract_id":  str,
              "title":        str,
              "state":        str,
              "stuck_since":  str,  # ISO-8601 datetime of last transition
              "days_stuck":   int,
            }
            Sorted by days_stuck descending.
        """
        raise NotImplementedError
