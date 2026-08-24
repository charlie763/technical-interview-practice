// Tests for Problem 14: Contract Amendment Manager
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_14_contract_amendment/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_14_contract_amendment.go \
//	  -c go test -v .
package amendments

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Shared dates
// ---------------------------------------------------------------------------

const (
	DBase  = "2025-01-01" // base / before any amendments
	DAmd1  = "2025-03-01" // amendment 1 effective date
	DAmd2  = "2025-06-01" // amendment 2 effective date
	DAmd3  = "2025-09-01" // amendment 3 effective date
	DAfter = "2025-12-31" // well after all amendments
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newMgr(t *testing.T) ContractAmendmentManager {
	t.Helper()
	return NewContractAmendmentManager()
}

// seededMgr returns a manager pre-seeded with two contracts:
//
//	c-seed-1  "Vendor MSA"
//	          fields = {"value": 50000, "payment_terms": "net-30", "currency": "USD"}
//	          amd-s1 (2025-03-01): overrides = {"payment_terms": "net-45"}
//	          amd-s2 (2025-06-01): overrides = {"value": 75000}
//
//	c-seed-2  "NDA Agreement"
//	          fields = {"term_years": 2, "auto_renew": true}
//	          (no amendments)
func seededMgr(t *testing.T) ContractAmendmentManager {
	t.Helper()
	m := NewContractAmendmentManager()
	mustAddContract(t, m, "c-seed-1", "Vendor MSA", map[string]interface{}{
		"value": 50000, "payment_terms": "net-30", "currency": "USD",
	})
	mustAddContract(t, m, "c-seed-2", "NDA Agreement", map[string]interface{}{
		"term_years": 2, "auto_renew": true,
	})
	mustAddAmendment(t, m, "amd-s1", "c-seed-1", DAmd1,
		map[string]interface{}{"payment_terms": "net-45"}, "extended payment terms")
	mustAddAmendment(t, m, "amd-s2", "c-seed-1", DAmd2,
		map[string]interface{}{"value": 75000}, "scope increase")
	return m
}

func mustAddContract(t *testing.T, m ContractAmendmentManager, id, title string, fields map[string]interface{}) *BaseContract {
	t.Helper()
	c, err := m.AddContract(id, title, fields)
	if err != nil {
		t.Fatalf("AddContract(%q): unexpected error: %v", id, err)
	}
	return c
}

func mustAddAmendment(t *testing.T, m ContractAmendmentManager, amdID, contractID, effectiveOn string, overrides map[string]interface{}, note string) *Amendment {
	t.Helper()
	a, err := m.AddAmendment(amdID, contractID, effectiveOn, overrides, note)
	if err != nil {
		t.Fatalf("AddAmendment(%q): unexpected error: %v", amdID, err)
	}
	return a
}

// ---------------------------------------------------------------------------
// PART 1 — Base contract management
// ---------------------------------------------------------------------------

func TestAddContract(t *testing.T) {
	t.Run("returns_contract_dict", func(t *testing.T) {
		m := newMgr(t)
		c, err := m.AddContract("c-add-1", "Test", map[string]interface{}{"x": 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ContractID != "c-add-1" {
			t.Errorf("ContractID = %q, want 'c-add-1'", c.ContractID)
		}
		if c.Title != "Test" {
			t.Errorf("Title = %q, want 'Test'", c.Title)
		}
		if c.Fields["x"] != 1 {
			t.Errorf("Fields[x] = %v, want 1", c.Fields["x"])
		}
	})

	t.Run("stores_copy_of_fields", func(t *testing.T) {
		m := newMgr(t)
		original := map[string]interface{}{"x": 1}
		m.AddContract("c-copy-1", "Test", original)
		original["x"] = 999
		base, _ := m.GetBaseContract("c-copy-1")
		if base.Fields["x"] != 1 {
			t.Errorf("Fields[x] = %v after mutating original, want 1 (should be a copy)", base.Fields["x"])
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		m := newMgr(t)
		m.AddContract("c-dup-am", "A", map[string]interface{}{})
		_, err := m.AddContract("c-dup-am", "B", map[string]interface{}{})
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestGetBaseContract(t *testing.T) {
	t.Run("returns_original_fields", func(t *testing.T) {
		m := seededMgr(t)
		base, err := m.GetBaseContract("c-seed-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Even though amendments exist, base fields are unchanged
		if base.Fields["payment_terms"] != "net-30" {
			t.Errorf("Fields[payment_terms] = %v, want 'net-30'", base.Fields["payment_terms"])
		}
		if base.Fields["value"] != 50000 {
			t.Errorf("Fields[value] = %v, want 50000", base.Fields["value"])
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		m := seededMgr(t)
		_, err := m.GetBaseContract("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Amendments and effective contract
// ---------------------------------------------------------------------------

func TestAddAmendment(t *testing.T) {
	t.Run("returns_amendment_dict", func(t *testing.T) {
		m := seededMgr(t)
		amd, err := m.AddAmendment("amd-add-1", "c-seed-1", DAmd3,
			map[string]interface{}{"currency": "EUR"}, "switch currency")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if amd.AmendmentID != "amd-add-1" {
			t.Errorf("AmendmentID = %q, want 'amd-add-1'", amd.AmendmentID)
		}
		if amd.ContractID != "c-seed-1" {
			t.Errorf("ContractID = %q, want 'c-seed-1'", amd.ContractID)
		}
		if amd.EffectiveOn != DAmd3 {
			t.Errorf("EffectiveOn = %q, want %q", amd.EffectiveOn, DAmd3)
		}
		if amd.Overrides["currency"] != "EUR" {
			t.Errorf("Overrides[currency] = %v, want 'EUR'", amd.Overrides["currency"])
		}
		if amd.Note != "switch currency" {
			t.Errorf("Note = %q, want 'switch currency'", amd.Note)
		}
	})

	t.Run("stores_copy_of_overrides", func(t *testing.T) {
		m := newMgr(t)
		m.AddContract("c-amd-copy", "T", map[string]interface{}{"x": 1})
		overrides := map[string]interface{}{"x": 2}
		m.AddAmendment("amd-copy-1", "c-amd-copy", DAmd1, overrides, "")
		overrides["x"] = 999
		amendments, _ := m.GetAmendments("c-amd-copy")
		if amendments[0].Overrides["x"] != 2 {
			t.Errorf("Overrides[x] = %v after mutating original, want 2 (should be a copy)", amendments[0].Overrides["x"])
		}
	})

	t.Run("duplicate_amendment_returns_error", func(t *testing.T) {
		m := seededMgr(t)
		_, err := m.AddAmendment("amd-s1", "c-seed-1", DAmd3, map[string]interface{}{}, "")
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("unknown_contract_returns_error", func(t *testing.T) {
		m := seededMgr(t)
		_, err := m.AddAmendment("amd-new", "no-such", DAmd1, map[string]interface{}{}, "")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetAmendments(t *testing.T) {
	t.Run("sorted_by_effective_on", func(t *testing.T) {
		m := seededMgr(t)
		amendments, err := m.GetAmendments("c-seed-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for i := 1; i < len(amendments); i++ {
			if amendments[i].EffectiveOn < amendments[i-1].EffectiveOn {
				t.Errorf("not sorted: %q before %q at index %d",
					amendments[i-1].EffectiveOn, amendments[i].EffectiveOn, i)
			}
		}
	})

	t.Run("empty_when_none", func(t *testing.T) {
		m := seededMgr(t)
		amendments, err := m.GetAmendments("c-seed-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(amendments) != 0 {
			t.Errorf("expected empty slice, got %d amendments", len(amendments))
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		m := seededMgr(t)
		_, err := m.GetAmendments("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("same_date_ordered_by_amendment_id", func(t *testing.T) {
		m := newMgr(t)
		m.AddContract("c-same-dt", "T", map[string]interface{}{"x": 0})
		m.AddAmendment("amd-z", "c-same-dt", DAmd1, map[string]interface{}{"x": 2}, "")
		m.AddAmendment("amd-a", "c-same-dt", DAmd1, map[string]interface{}{"x": 1}, "")
		amendments, _ := m.GetAmendments("c-same-dt")
		if len(amendments) != 2 {
			t.Fatalf("len(amendments) = %d, want 2", len(amendments))
		}
		if amendments[0].AmendmentID != "amd-a" || amendments[1].AmendmentID != "amd-z" {
			t.Errorf("amendments not sorted by ID: [%q, %q], want ['amd-a', 'amd-z']",
				amendments[0].AmendmentID, amendments[1].AmendmentID)
		}
	})
}

func TestGetEffectiveContract(t *testing.T) {
	t.Run("before_any_amendments", func(t *testing.T) {
		m := seededMgr(t)
		fields, err := m.GetEffectiveContract("c-seed-1", DBase)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fields["payment_terms"] != "net-30" {
			t.Errorf("payment_terms = %v, want 'net-30'", fields["payment_terms"])
		}
		if fields["value"] != 50000 {
			t.Errorf("value = %v, want 50000", fields["value"])
		}
	})

	t.Run("after_first_amendment", func(t *testing.T) {
		m := seededMgr(t)
		fields, _ := m.GetEffectiveContract("c-seed-1", "2025-04-15")
		if fields["payment_terms"] != "net-45" {
			t.Errorf("payment_terms = %v, want 'net-45'", fields["payment_terms"])
		}
		if fields["value"] != 50000 {
			t.Errorf("value = %v, want 50000 (amd-s2 not yet effective)", fields["value"])
		}
	})

	t.Run("after_all_amendments", func(t *testing.T) {
		m := seededMgr(t)
		fields, _ := m.GetEffectiveContract("c-seed-1", DAfter)
		if fields["payment_terms"] != "net-45" {
			t.Errorf("payment_terms = %v, want 'net-45'", fields["payment_terms"])
		}
		if fields["value"] != 75000 {
			t.Errorf("value = %v, want 75000", fields["value"])
		}
	})

	t.Run("original_unamended_fields_preserved", func(t *testing.T) {
		m := seededMgr(t)
		fields, _ := m.GetEffectiveContract("c-seed-1", DAfter)
		if fields["currency"] != "USD" {
			t.Errorf("currency = %v, want 'USD'", fields["currency"])
		}
	})

	t.Run("no_amendments_returns_base", func(t *testing.T) {
		m := seededMgr(t)
		fields, _ := m.GetEffectiveContract("c-seed-2", DAfter)
		if fields["term_years"] != 2 {
			t.Errorf("term_years = %v, want 2", fields["term_years"])
		}
		if fields["auto_renew"] != true {
			t.Errorf("auto_renew = %v, want true", fields["auto_renew"])
		}
	})

	t.Run("exact_effective_date_is_inclusive", func(t *testing.T) {
		m := seededMgr(t)
		// amd-s1 effective_on = DAmd1; querying exactly DAmd1 should apply it
		fields, _ := m.GetEffectiveContract("c-seed-1", DAmd1)
		if fields["payment_terms"] != "net-45" {
			t.Errorf("payment_terms = %v, want 'net-45' (effective date is inclusive)", fields["payment_terms"])
		}
	})

	t.Run("does_not_mutate_base", func(t *testing.T) {
		m := seededMgr(t)
		m.GetEffectiveContract("c-seed-1", DAfter)
		base, _ := m.GetBaseContract("c-seed-1")
		if base.Fields["payment_terms"] != "net-30" {
			t.Errorf("base payment_terms = %v, want 'net-30' (should not be mutated)", base.Fields["payment_terms"])
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		m := seededMgr(t)
		_, err := m.GetEffectiveContract("no-such", DAfter)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Value history and amendment summary
// ---------------------------------------------------------------------------

func TestGetValueHistory(t *testing.T) {
	t.Run("base_value_included_first", func(t *testing.T) {
		m := seededMgr(t)
		history, err := m.GetValueHistory("c-seed-1", "value")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(history) == 0 {
			t.Fatal("history is empty")
		}
		if history[0].Source != "base" {
			t.Errorf("history[0].Source = %q, want 'base'", history[0].Source)
		}
		if history[0].Value != 50000 {
			t.Errorf("history[0].Value = %v, want 50000", history[0].Value)
		}
	})

	t.Run("amendment_override_included", func(t *testing.T) {
		m := seededMgr(t)
		history, _ := m.GetValueHistory("c-seed-1", "payment_terms")
		found := false
		for _, e := range history {
			if e.Source == "amd-s1" {
				found = true
			}
		}
		if !found {
			t.Error("amd-s1 not found in payment_terms history")
		}
	})

	t.Run("only_amendments_that_touched_field", func(t *testing.T) {
		m := seededMgr(t)
		// amd-s2 changes "value", not "payment_terms"
		history, _ := m.GetValueHistory("c-seed-1", "payment_terms")
		for _, e := range history {
			if e.Source == "amd-s2" {
				t.Error("amd-s2 should not appear in payment_terms history (it only changed 'value')")
			}
		}
	})

	t.Run("sorted_chronologically", func(t *testing.T) {
		m := seededMgr(t)
		history, _ := m.GetValueHistory("c-seed-1", "value")
		if len(history) == 0 || history[0].Source != "base" {
			t.Error("base entry should come first")
		}
		// Amendment entries should be in ascending date order
		var prevDate string
		for _, e := range history {
			if e.Source == "base" {
				continue
			}
			if e.EffectiveOn < prevDate {
				t.Errorf("history not sorted: %q < %q", e.EffectiveOn, prevDate)
			}
			prevDate = e.EffectiveOn
		}
	})

	t.Run("field_not_present_raises", func(t *testing.T) {
		m := seededMgr(t)
		_, err := m.GetValueHistory("c-seed-1", "nonexistent_field")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("unknown_contract_raises", func(t *testing.T) {
		m := seededMgr(t)
		_, err := m.GetValueHistory("no-such", "value")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetAmendmentSummary(t *testing.T) {
	t.Run("amendment_count", func(t *testing.T) {
		m := seededMgr(t)
		summary, err := m.GetAmendmentSummary("c-seed-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary.AmendmentCount != 2 {
			t.Errorf("AmendmentCount = %d, want 2", summary.AmendmentCount)
		}
	})

	t.Run("fields_amended_sorted", func(t *testing.T) {
		m := seededMgr(t)
		summary, _ := m.GetAmendmentSummary("c-seed-1")
		want := []string{"payment_terms", "value"}
		if len(summary.FieldsAmended) != len(want) {
			t.Fatalf("FieldsAmended = %v, want %v", summary.FieldsAmended, want)
		}
		for i, f := range summary.FieldsAmended {
			if f != want[i] {
				t.Errorf("FieldsAmended[%d] = %q, want %q", i, f, want[i])
			}
		}
	})

	t.Run("latest_amendment_date", func(t *testing.T) {
		m := seededMgr(t)
		summary, _ := m.GetAmendmentSummary("c-seed-1")
		if summary.LatestAmendment == nil || *summary.LatestAmendment != DAmd2 {
			t.Errorf("LatestAmendment = %v, want %q", summary.LatestAmendment, DAmd2)
		}
	})

	t.Run("current_fields_reflect_all_amendments", func(t *testing.T) {
		m := seededMgr(t)
		summary, _ := m.GetAmendmentSummary("c-seed-1")
		if summary.CurrentFields["payment_terms"] != "net-45" {
			t.Errorf("CurrentFields[payment_terms] = %v, want 'net-45'", summary.CurrentFields["payment_terms"])
		}
		if summary.CurrentFields["value"] != 75000 {
			t.Errorf("CurrentFields[value] = %v, want 75000", summary.CurrentFields["value"])
		}
	})

	t.Run("no_amendments_returns_nil_latest", func(t *testing.T) {
		m := seededMgr(t)
		summary, _ := m.GetAmendmentSummary("c-seed-2")
		if summary.LatestAmendment != nil {
			t.Errorf("LatestAmendment = %v, want nil", summary.LatestAmendment)
		}
		if summary.AmendmentCount != 0 {
			t.Errorf("AmendmentCount = %d, want 0", summary.AmendmentCount)
		}
		if len(summary.FieldsAmended) != 0 {
			t.Errorf("FieldsAmended = %v, want empty", summary.FieldsAmended)
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		m := seededMgr(t)
		_, err := m.GetAmendmentSummary("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
