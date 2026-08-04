"""Tests for Problem 14: Contract Amendment Manager

Run from the python/ directory:
    pytest tests/test_problem_14_contract_amendment.py -v
"""

import pytest
from practice_problems.problem_14_contract_amendment import ContractAmendmentManager

# ---------------------------------------------------------------------------
# Shared dates
# ---------------------------------------------------------------------------
D_BASE  = "2025-01-01"   # base / before any amendments
D_AMD1  = "2025-03-01"   # amendment 1 effective date
D_AMD2  = "2025-06-01"   # amendment 2 effective date
D_AMD3  = "2025-09-01"   # amendment 3 effective date
D_AFTER = "2025-12-31"   # well after all amendments


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def fresh_mgr():
    """Empty ContractAmendmentManager."""
    return ContractAmendmentManager()


@pytest.fixture
def mgr():
    """
    Pre-seeded manager:
      c-seed-1  "Vendor MSA"
                fields = {"value": 50000, "payment_terms": "net-30", "currency": "USD"}
                amd-s1 (2025-03-01): overrides = {"payment_terms": "net-45"}
                amd-s2 (2025-06-01): overrides = {"value": 75000}

      c-seed-2  "NDA Agreement"
                fields = {"term_years": 2, "auto_renew": True}
                (no amendments)
    """
    m = ContractAmendmentManager()
    m.add_contract(
        "c-seed-1", "Vendor MSA",
        fields={"value": 50000, "payment_terms": "net-30", "currency": "USD"},
    )
    m.add_contract(
        "c-seed-2", "NDA Agreement",
        fields={"term_years": 2, "auto_renew": True},
    )
    m.add_amendment(
        "amd-s1", "c-seed-1",
        effective_on=D_AMD1,
        overrides={"payment_terms": "net-45"},
        note="extended payment terms",
    )
    m.add_amendment(
        "amd-s2", "c-seed-1",
        effective_on=D_AMD2,
        overrides={"value": 75000},
        note="scope increase",
    )
    return m


# ---------------------------------------------------------------------------
# PART 1 — Base contract management
# ---------------------------------------------------------------------------

class TestAddContract:
    def test_returns_contract_dict(self, fresh_mgr):
        c = fresh_mgr.add_contract("c-add-1", "Test", fields={"x": 1})
        assert c["contract_id"] == "c-add-1"
        assert c["title"] == "Test"
        assert c["fields"]["x"] == 1

    def test_stores_copy_of_fields(self, fresh_mgr):
        original = {"x": 1}
        fresh_mgr.add_contract("c-copy-1", "Test", fields=original)
        original["x"] = 999
        assert fresh_mgr.get_base_contract("c-copy-1")["fields"]["x"] == 1

    def test_duplicate_raises(self, fresh_mgr):
        fresh_mgr.add_contract("c-dup-am", "A", fields={})
        with pytest.raises(ValueError):
            fresh_mgr.add_contract("c-dup-am", "B", fields={})


class TestGetBaseContract:
    def test_returns_original_fields(self, mgr):
        base = mgr.get_base_contract("c-seed-1")
        # Even though amendments exist, base fields are unchanged
        assert base["fields"]["payment_terms"] == "net-30"
        assert base["fields"]["value"] == 50000

    def test_unknown_raises(self, mgr):
        with pytest.raises(KeyError):
            mgr.get_base_contract("no-such")


# ---------------------------------------------------------------------------
# PART 2 — Amendments and effective contract
# ---------------------------------------------------------------------------

class TestAddAmendment:
    def test_returns_amendment_dict(self, mgr):
        amd = mgr.add_amendment(
            "amd-add-1", "c-seed-1",
            effective_on=D_AMD3,
            overrides={"currency": "EUR"},
            note="switch currency",
        )
        assert amd["amendment_id"] == "amd-add-1"
        assert amd["contract_id"]  == "c-seed-1"
        assert amd["effective_on"] == D_AMD3
        assert amd["overrides"] == {"currency": "EUR"}
        assert amd["note"] == "switch currency"

    def test_stores_copy_of_overrides(self, fresh_mgr):
        fresh_mgr.add_contract("c-amd-copy", "T", fields={"x": 1})
        overrides = {"x": 2}
        fresh_mgr.add_amendment("amd-copy-1", "c-amd-copy", effective_on=D_AMD1, overrides=overrides, note="")
        overrides["x"] = 999
        amendments = fresh_mgr.get_amendments("c-amd-copy")
        assert amendments[0]["overrides"]["x"] == 2

    def test_duplicate_amendment_raises(self, mgr):
        with pytest.raises(ValueError):
            mgr.add_amendment("amd-s1", "c-seed-1", effective_on=D_AMD3, overrides={}, note="")

    def test_unknown_contract_raises(self, mgr):
        with pytest.raises(KeyError):
            mgr.add_amendment("amd-new", "no-such", effective_on=D_AMD1, overrides={}, note="")


class TestGetAmendments:
    def test_sorted_by_effective_on(self, mgr):
        amendments = mgr.get_amendments("c-seed-1")
        dates = [a["effective_on"] for a in amendments]
        assert dates == sorted(dates)

    def test_empty_when_none(self, mgr):
        assert mgr.get_amendments("c-seed-2") == []

    def test_unknown_raises(self, mgr):
        with pytest.raises(KeyError):
            mgr.get_amendments("no-such")

    def test_same_date_ordered_by_amendment_id(self, fresh_mgr):
        fresh_mgr.add_contract("c-same-dt", "T", fields={"x": 0})
        fresh_mgr.add_amendment("amd-z", "c-same-dt", effective_on=D_AMD1, overrides={"x": 2}, note="")
        fresh_mgr.add_amendment("amd-a", "c-same-dt", effective_on=D_AMD1, overrides={"x": 1}, note="")
        amendments = fresh_mgr.get_amendments("c-same-dt")
        ids = [a["amendment_id"] for a in amendments]
        assert ids == sorted(ids)


class TestGetEffectiveContract:
    def test_before_any_amendments(self, mgr):
        fields = mgr.get_effective_contract("c-seed-1", as_of_date=D_BASE)
        assert fields["payment_terms"] == "net-30"
        assert fields["value"] == 50000

    def test_after_first_amendment(self, mgr):
        fields = mgr.get_effective_contract("c-seed-1", as_of_date="2025-04-15")
        assert fields["payment_terms"] == "net-45"
        assert fields["value"] == 50000   # amd-s2 not yet effective

    def test_after_all_amendments(self, mgr):
        fields = mgr.get_effective_contract("c-seed-1", as_of_date=D_AFTER)
        assert fields["payment_terms"] == "net-45"
        assert fields["value"] == 75000

    def test_original_unamended_fields_preserved(self, mgr):
        fields = mgr.get_effective_contract("c-seed-1", as_of_date=D_AFTER)
        assert fields["currency"] == "USD"

    def test_no_amendments_returns_base(self, mgr):
        fields = mgr.get_effective_contract("c-seed-2", as_of_date=D_AFTER)
        assert fields["term_years"] == 2
        assert fields["auto_renew"] is True

    def test_exact_effective_date_is_inclusive(self, mgr):
        # amd-s1 effective_on = D_AMD1; querying exactly D_AMD1 should apply it
        fields = mgr.get_effective_contract("c-seed-1", as_of_date=D_AMD1)
        assert fields["payment_terms"] == "net-45"

    def test_does_not_mutate_base(self, mgr):
        mgr.get_effective_contract("c-seed-1", as_of_date=D_AFTER)
        base = mgr.get_base_contract("c-seed-1")
        assert base["fields"]["payment_terms"] == "net-30"

    def test_unknown_raises(self, mgr):
        with pytest.raises(KeyError):
            mgr.get_effective_contract("no-such", as_of_date=D_AFTER)


# ---------------------------------------------------------------------------
# PART 3 — Value history and amendment summary
# ---------------------------------------------------------------------------

class TestGetValueHistory:
    def test_base_value_included_first(self, mgr):
        history = mgr.get_value_history("c-seed-1", "value")
        assert history[0]["source"] == "base"
        assert history[0]["value"] == 50000

    def test_amendment_override_included(self, mgr):
        history = mgr.get_value_history("c-seed-1", "payment_terms")
        sources = [e["source"] for e in history]
        assert "amd-s1" in sources

    def test_only_amendments_that_touched_field(self, mgr):
        # amd-s2 changes "value", not "payment_terms"
        history = mgr.get_value_history("c-seed-1", "payment_terms")
        sources = [e["source"] for e in history]
        assert "amd-s2" not in sources

    def test_sorted_chronologically(self, mgr):
        history = mgr.get_value_history("c-seed-1", "value")
        # base first, then amendment entries in date order
        assert history[0]["source"] == "base"
        dates = [e["effective_on"] for e in history if e["source"] != "base"]
        assert dates == sorted(dates)

    def test_field_not_present_raises(self, mgr):
        with pytest.raises(KeyError):
            mgr.get_value_history("c-seed-1", "nonexistent_field")

    def test_unknown_contract_raises(self, mgr):
        with pytest.raises(KeyError):
            mgr.get_value_history("no-such", "value")


class TestGetAmendmentSummary:
    def test_amendment_count(self, mgr):
        summary = mgr.get_amendment_summary("c-seed-1")
        assert summary["amendment_count"] == 2

    def test_fields_amended_sorted(self, mgr):
        summary = mgr.get_amendment_summary("c-seed-1")
        assert summary["fields_amended"] == sorted(["payment_terms", "value"])

    def test_latest_amendment_date(self, mgr):
        summary = mgr.get_amendment_summary("c-seed-1")
        assert summary["latest_amendment"] == D_AMD2

    def test_current_fields_reflect_all_amendments(self, mgr):
        summary = mgr.get_amendment_summary("c-seed-1")
        assert summary["current_fields"]["payment_terms"] == "net-45"
        assert summary["current_fields"]["value"] == 75000

    def test_no_amendments_returns_none_latest(self, mgr):
        summary = mgr.get_amendment_summary("c-seed-2")
        assert summary["latest_amendment"] is None
        assert summary["amendment_count"] == 0
        assert summary["fields_amended"] == []

    def test_unknown_raises(self, mgr):
        with pytest.raises(KeyError):
            mgr.get_amendment_summary("no-such")
