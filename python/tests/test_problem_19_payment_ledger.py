"""Tests for Problem 16: Payment Ledger & Reconciliation

Run from the python/ directory:
    pytest tests/test_problem_19_payment_ledger.py -v
"""

import pytest

from practice_problems.problem_19_payment_ledger import LedgerService

# ---------------------------------------------------------------------------
# Shared timestamps
# ---------------------------------------------------------------------------
T0 = "2024-06-01T10:00:00"
T0_1H = "2024-06-01T11:00:00"
T0_2H = "2024-06-01T12:00:00"


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def fresh_ledger():
    """Empty LedgerService."""
    return LedgerService()


@pytest.fixture
def ledger():
    """
    Pre-seeded ledger:
      Accounts: wallet:alice (200.0), wallet:bob (0.0), platform:fees (0.0)
      Applied transfers: tx_seed_1 (alice -> bob, 50.0, T0)
                          tx_seed_2 (alice -> platform:fees, 5.0, T0_1H)
    """
    l = LedgerService()
    l.open_account("wallet:alice", initial_balance=200.0)
    l.open_account("wallet:bob")
    l.open_account("platform:fees")
    l.record_transfer("tx_seed_1", "wallet:alice", "wallet:bob", 50.0, T0)
    l.record_transfer("tx_seed_2", "wallet:alice", "platform:fees", 5.0, T0_1H)
    return l


# ---------------------------------------------------------------------------
# PART 1 — Accounts & idempotent transfers
# ---------------------------------------------------------------------------

class TestAccountsAndTransfers:
    def test_open_account_sets_initial_balance(self, fresh_ledger):
        fresh_ledger.open_account("wallet:open_test", initial_balance=25.0)
        assert fresh_ledger.get_balance("wallet:open_test") == 25.0

    def test_open_account_defaults_to_zero(self, fresh_ledger):
        fresh_ledger.open_account("wallet:default_test")
        assert fresh_ledger.get_balance("wallet:default_test") == 0.0

    def test_duplicate_account_raises(self, fresh_ledger):
        fresh_ledger.open_account("wallet:dup_test")
        with pytest.raises(ValueError):
            fresh_ledger.open_account("wallet:dup_test")

    def test_negative_initial_balance_raises(self, fresh_ledger):
        with pytest.raises(ValueError):
            fresh_ledger.open_account("wallet:negative_test", initial_balance=-5.0)

    def test_get_balance_unknown_account_raises(self, fresh_ledger):
        with pytest.raises(KeyError):
            fresh_ledger.get_balance("wallet:ghost")

    def test_record_transfer_applies_and_updates_balances(self, fresh_ledger):
        fresh_ledger.open_account("wallet:a_apply", initial_balance=100.0)
        fresh_ledger.open_account("wallet:b_apply")
        result = fresh_ledger.record_transfer(
            "tx_apply_test", "wallet:a_apply", "wallet:b_apply", 40.0, T0
        )
        assert result == {"status": "applied", "reason": None}
        assert fresh_ledger.get_balance("wallet:a_apply") == 60.0
        assert fresh_ledger.get_balance("wallet:b_apply") == 40.0

    def test_duplicate_transaction_id_does_not_reapply(self, fresh_ledger):
        fresh_ledger.open_account("wallet:a_dup", initial_balance=100.0)
        fresh_ledger.open_account("wallet:b_dup")
        fresh_ledger.record_transfer("tx_dup_test", "wallet:a_dup", "wallet:b_dup", 40.0, T0)
        second = fresh_ledger.record_transfer(
            "tx_dup_test", "wallet:a_dup", "wallet:b_dup", 40.0, T0_1H
        )
        assert second == {"status": "duplicate", "reason": None}
        assert fresh_ledger.get_balance("wallet:a_dup") == 60.0
        assert fresh_ledger.get_balance("wallet:b_dup") == 40.0

    def test_unknown_account_rejected(self, fresh_ledger):
        fresh_ledger.open_account("wallet:a_unknown", initial_balance=100.0)
        result = fresh_ledger.record_transfer(
            "tx_unknown_test", "wallet:a_unknown", "wallet:ghost", 10.0, T0
        )
        assert result == {"status": "rejected", "reason": "unknown_account"}

    def test_invalid_amount_rejected(self, fresh_ledger):
        fresh_ledger.open_account("wallet:a_invalid", initial_balance=100.0)
        fresh_ledger.open_account("wallet:b_invalid")
        result = fresh_ledger.record_transfer(
            "tx_invalid_test", "wallet:a_invalid", "wallet:b_invalid", 0.0, T0
        )
        assert result == {"status": "rejected", "reason": "invalid_amount"}

    def test_insufficient_funds_rejected(self, fresh_ledger):
        fresh_ledger.open_account("wallet:a_short", initial_balance=10.0)
        fresh_ledger.open_account("wallet:b_short")
        result = fresh_ledger.record_transfer(
            "tx_short_test", "wallet:a_short", "wallet:b_short", 40.0, T0
        )
        assert result == {"status": "rejected", "reason": "insufficient_funds"}
        assert fresh_ledger.get_balance("wallet:a_short") == 10.0

    def test_retrying_a_rejected_transaction_id_is_evaluated_fresh(self, fresh_ledger):
        fresh_ledger.open_account("wallet:a_retry", initial_balance=0.0)
        fresh_ledger.open_account("wallet:b_retry")
        fresh_ledger.open_account("wallet:c_retry", initial_balance=100.0)

        first = fresh_ledger.record_transfer(
            "tx_retry_test", "wallet:a_retry", "wallet:b_retry", 50.0, T0
        )
        assert first == {"status": "rejected", "reason": "insufficient_funds"}

        fresh_ledger.record_transfer(
            "tx_retry_topup", "wallet:c_retry", "wallet:a_retry", 50.0, T0
        )
        second = fresh_ledger.record_transfer(
            "tx_retry_test", "wallet:a_retry", "wallet:b_retry", 50.0, T0_1H
        )
        assert second == {"status": "applied", "reason": None}
        assert fresh_ledger.get_balance("wallet:b_retry") == 50.0


# ---------------------------------------------------------------------------
# PART 2 — Reversals
# ---------------------------------------------------------------------------

class TestReversals:
    def test_reverses_applied_transfer(self, fresh_ledger):
        fresh_ledger.open_account("wallet:alice_rev", initial_balance=100.0)
        fresh_ledger.open_account("wallet:bob_rev")
        fresh_ledger.record_transfer(
            "tx_rev_base", "wallet:alice_rev", "wallet:bob_rev", 40.0, T0
        )
        result = fresh_ledger.reverse_transfer("tx_rev_base", "tx_rev_base_r", T0_1H)
        assert result == {"status": "applied", "reason": None}
        assert fresh_ledger.get_balance("wallet:alice_rev") == 100.0
        assert fresh_ledger.get_balance("wallet:bob_rev") == 0.0

    def test_reverse_unknown_transaction_raises(self, fresh_ledger):
        with pytest.raises(KeyError):
            fresh_ledger.reverse_transfer("nonexistent_tx", "rev_of_nonexistent", T0)

    def test_reverse_already_reversed_rejected(self, fresh_ledger):
        fresh_ledger.open_account("wallet:alice_double", initial_balance=100.0)
        fresh_ledger.open_account("wallet:bob_double")
        fresh_ledger.record_transfer(
            "tx_double_rev", "wallet:alice_double", "wallet:bob_double", 40.0, T0
        )
        fresh_ledger.reverse_transfer("tx_double_rev", "tx_double_rev_r1", T0_1H)
        result = fresh_ledger.reverse_transfer("tx_double_rev", "tx_double_rev_r2", T0_2H)
        assert result == {"status": "rejected", "reason": "already_reversed"}

    def test_reversal_fails_if_funds_already_spent(self, fresh_ledger):
        fresh_ledger.open_account("wallet:alice_spend", initial_balance=100.0)
        fresh_ledger.open_account("wallet:bob_spend")
        fresh_ledger.open_account("wallet:carol_spend")
        fresh_ledger.record_transfer(
            "tx_spend_base", "wallet:alice_spend", "wallet:bob_spend", 40.0, T0
        )
        # bob spends the money elsewhere before the reversal is attempted
        fresh_ledger.record_transfer(
            "tx_spend_away", "wallet:bob_spend", "wallet:carol_spend", 40.0, T0_1H
        )
        result = fresh_ledger.reverse_transfer("tx_spend_base", "tx_spend_base_r", T0_2H)
        assert result == {"status": "rejected", "reason": "insufficient_funds"}

    def test_reversal_is_itself_a_regular_idempotent_transfer(self, fresh_ledger):
        fresh_ledger.open_account("wallet:alice_idem", initial_balance=100.0)
        fresh_ledger.open_account("wallet:bob_idem")
        fresh_ledger.record_transfer(
            "tx_idem_base", "wallet:alice_idem", "wallet:bob_idem", 40.0, T0
        )
        fresh_ledger.reverse_transfer("tx_idem_base", "tx_idem_base_r", T0_1H)
        # Replaying the reversal's own transaction_id directly must be a
        # no-op duplicate, proving reverse_transfer moved money via
        # record_transfer rather than a separate code path.
        replay = fresh_ledger.record_transfer(
            "tx_idem_base_r", "wallet:bob_idem", "wallet:alice_idem", 40.0, T0_2H
        )
        assert replay == {"status": "duplicate", "reason": None}
        assert fresh_ledger.get_balance("wallet:alice_idem") == 100.0


# ---------------------------------------------------------------------------
# PART 3 — Reconciliation
# ---------------------------------------------------------------------------

class TestReconciliation:
    def test_matches_statement_line_that_agrees(self, ledger):
        statement = [
            {"transaction_id": "tx_seed_1", "account_id": "wallet:alice", "amount": 50.0, "type": "debit"}
        ]
        report = ledger.reconcile(statement)
        assert report["matched"] == ["tx_seed_1"]
        assert report["mismatched"] == []
        assert report["missing_from_ledger"] == []
        assert "tx_seed_2" in report["missing_from_statement"]

    def test_flags_amount_mismatch(self, ledger):
        statement = [
            {"transaction_id": "tx_seed_1", "account_id": "wallet:alice", "amount": 999.0, "type": "debit"}
        ]
        report = ledger.reconcile(statement)
        assert report["mismatched"] == [{"transaction_id": "tx_seed_1", "reason": "amount_mismatch"}]

    def test_flags_account_mismatch(self, ledger):
        statement = [
            {"transaction_id": "tx_seed_1", "account_id": "wallet:bob", "amount": 50.0, "type": "debit"}
        ]
        report = ledger.reconcile(statement)
        assert report["mismatched"] == [{"transaction_id": "tx_seed_1", "reason": "account_mismatch"}]

    def test_account_mismatch_takes_priority_over_amount_mismatch(self, ledger):
        statement = [
            {"transaction_id": "tx_seed_1", "account_id": "wallet:bob", "amount": 999.0, "type": "debit"}
        ]
        report = ledger.reconcile(statement)
        assert report["mismatched"] == [{"transaction_id": "tx_seed_1", "reason": "account_mismatch"}]

    def test_flags_missing_from_ledger(self, ledger):
        statement = [
            {"transaction_id": "tx_unknown", "account_id": "wallet:alice", "amount": 10.0, "type": "debit"}
        ]
        report = ledger.reconcile(statement)
        assert report["missing_from_ledger"] == ["tx_unknown"]

    def test_flags_missing_from_statement(self, ledger):
        report = ledger.reconcile([])
        assert set(report["missing_from_statement"]) == {"tx_seed_1", "tx_seed_2"}

    def test_ignores_credit_lines(self, ledger):
        statement = [
            {"transaction_id": "tx_seed_1", "account_id": "wallet:bob", "amount": 50.0, "type": "credit"}
        ]
        report = ledger.reconcile(statement)
        assert report["matched"] == []
        assert "tx_seed_1" in report["missing_from_statement"]

    def test_reconciliation_includes_reversed_transfers(self, fresh_ledger):
        fresh_ledger.open_account("wallet:x_rec", initial_balance=100.0)
        fresh_ledger.open_account("wallet:y_rec")
        fresh_ledger.record_transfer("tx_rec_rev", "wallet:x_rec", "wallet:y_rec", 20.0, T0)
        fresh_ledger.reverse_transfer("tx_rec_rev", "tx_rec_rev_r", T0_1H)
        statement = [
            {"transaction_id": "tx_rec_rev", "account_id": "wallet:x_rec", "amount": 20.0, "type": "debit"},
            {"transaction_id": "tx_rec_rev_r", "account_id": "wallet:y_rec", "amount": 20.0, "type": "debit"},
        ]
        report = fresh_ledger.reconcile(statement)
        assert set(report["matched"]) == {"tx_rec_rev", "tx_rec_rev_r"}
