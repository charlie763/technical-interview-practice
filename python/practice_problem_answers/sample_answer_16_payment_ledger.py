"""
Reference solution for Problem 16: Payment Ledger & Reconciliation

See python/practice_problems/problem_16_payment_ledger.py for the full
problem statement and method contracts.
"""


class LedgerService:
    def __init__(self):
        self._balances = {}
        self._transfers = {}  # transaction_id -> transfer record (applied only)

    # -------------------------------------------------------------------------
    # PART 1 — Accounts & idempotent transfers
    # -------------------------------------------------------------------------

    def open_account(self, account_id, initial_balance=0.0):
        if account_id in self._balances:
            raise ValueError(f"account already exists: {account_id}")
        if initial_balance < 0:
            raise ValueError("initial_balance must be >= 0")
        self._balances[account_id] = initial_balance

    def get_balance(self, account_id):
        if account_id not in self._balances:
            raise KeyError(account_id)
        return self._balances[account_id]

    def record_transfer(self, transaction_id, from_account, to_account, amount, timestamp):
        existing = self._transfers.get(transaction_id)
        if existing is not None:
            return {"status": "duplicate", "reason": None}

        if from_account not in self._balances or to_account not in self._balances:
            return {"status": "rejected", "reason": "unknown_account"}
        if amount <= 0:
            return {"status": "rejected", "reason": "invalid_amount"}
        if self._balances[from_account] < amount:
            return {"status": "rejected", "reason": "insufficient_funds"}

        self._balances[from_account] -= amount
        self._balances[to_account] += amount
        self._transfers[transaction_id] = {
            "from_account": from_account,
            "to_account": to_account,
            "amount": amount,
            "timestamp": timestamp,
            "reversed": False,
        }
        return {"status": "applied", "reason": None}

    # -------------------------------------------------------------------------
    # PART 2 — Reversals
    # -------------------------------------------------------------------------

    def reverse_transfer(self, transaction_id, reversal_id, timestamp):
        original = self._transfers.get(transaction_id)
        if original is None:
            raise KeyError(transaction_id)
        if original["reversed"]:
            return {"status": "rejected", "reason": "already_reversed"}

        result = self.record_transfer(
            reversal_id,
            original["to_account"],
            original["from_account"],
            original["amount"],
            timestamp,
        )
        if result["status"] == "applied":
            original["reversed"] = True
        return result

    # -------------------------------------------------------------------------
    # PART 3 — Reconciliation
    # -------------------------------------------------------------------------

    def reconcile(self, statement_lines):
        matched = []
        mismatched = []
        missing_from_ledger = []
        seen_transaction_ids = set()

        for line in statement_lines:
            if line["type"] != "debit":
                continue
            transaction_id = line["transaction_id"]
            seen_transaction_ids.add(transaction_id)
            transfer = self._transfers.get(transaction_id)

            if transfer is None:
                missing_from_ledger.append(transaction_id)
                continue

            if transfer["from_account"] != line["account_id"]:
                mismatched.append(
                    {"transaction_id": transaction_id, "reason": "account_mismatch"}
                )
            elif transfer["amount"] != line["amount"]:
                mismatched.append(
                    {"transaction_id": transaction_id, "reason": "amount_mismatch"}
                )
            else:
                matched.append(transaction_id)

        missing_from_statement = [
            transaction_id
            for transaction_id in self._transfers
            if transaction_id not in seen_transaction_ids
        ]

        return {
            "matched": matched,
            "mismatched": mismatched,
            "missing_from_ledger": missing_from_ledger,
            "missing_from_statement": missing_from_statement,
        }
