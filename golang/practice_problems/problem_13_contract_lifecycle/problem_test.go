// Tests for Problem 13: Contract Lifecycle State Machine
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_13_contract_lifecycle/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_13_contract_lifecycle.go \
//	  -c go test -v .
package lifecycle

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Shared timestamps
// T0 = base datetime
// T1 = T0 + 1 day
// T2 = T0 + 5 days
// T3 = T0 + 10 days
// T4 = T0 + 40 days  (> 30 days for overdue tests)
// T5 = T0 + 45 days
// ---------------------------------------------------------------------------

const (
	T0 = "2025-01-01T09:00:00"
	T1 = "2025-01-02T10:00:00"
	T2 = "2025-01-06T11:00:00"
	T3 = "2025-01-11T12:00:00"
	T4 = "2025-02-10T09:00:00" // ~40 days after T0
	T5 = "2025-02-15T09:00:00" // ~45 days after T0
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newCL(t *testing.T) ContractLifecycle {
	t.Helper()
	return NewContractLifecycle()
}

// seededCL returns a lifecycle manager pre-seeded with three contracts:
//
//	c-seed-1  "Vendor MSA"     state=approved    (draft→in_review→approved)
//	c-seed-2  "NDA Agreement"  state=draft
//	c-seed-3  "SaaS License"   state=terminated  (full path)
func seededCL(t *testing.T) ContractLifecycle {
	t.Helper()
	cl := NewContractLifecycle()
	// c-seed-1: draft → in_review → approved
	mustCreateContract(t, cl, "c-seed-1", "Vendor MSA", T0, "alice")
	mustTransition(t, cl, "c-seed-1", "in_review", T1, "alice")
	mustTransition(t, cl, "c-seed-1", "approved", T2, "bob")
	// c-seed-2: stays draft
	mustCreateContract(t, cl, "c-seed-2", "NDA Agreement", T0, "alice")
	// c-seed-3: full path to terminated
	mustCreateContract(t, cl, "c-seed-3", "SaaS License", T0, "carol")
	mustTransition(t, cl, "c-seed-3", "in_review", T1, "carol")
	mustTransition(t, cl, "c-seed-3", "approved", T2, "bob")
	mustTransition(t, cl, "c-seed-3", "executed", T3, "carol")
	mustTransition(t, cl, "c-seed-3", "active", T4, "carol")
	mustTransition(t, cl, "c-seed-3", "terminated", T5, "carol")
	return cl
}

func mustCreateContract(t *testing.T, cl ContractLifecycle, id, title, createdAt, actor string) *Contract {
	t.Helper()
	c, err := cl.CreateContract(id, title, createdAt, actor)
	if err != nil {
		t.Fatalf("CreateContract(%q): unexpected error: %v", id, err)
	}
	return c
}

func mustTransition(t *testing.T, cl ContractLifecycle, id, toState, at, actor string) *Contract {
	t.Helper()
	c, err := cl.Transition(id, toState, at, actor)
	if err != nil {
		t.Fatalf("Transition(%q → %q): unexpected error: %v", id, toState, err)
	}
	return c
}

// ---------------------------------------------------------------------------
// PART 1 — Contract creation, field management, transitions
// ---------------------------------------------------------------------------

func TestCreateContract(t *testing.T) {
	t.Run("returns_contract_in_draft", func(t *testing.T) {
		cl := newCL(t)
		c, err := cl.CreateContract("c-cr-1", "Title", T0, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ContractID != "c-cr-1" {
			t.Errorf("ContractID = %q, want 'c-cr-1'", c.ContractID)
		}
		if c.State != "draft" {
			t.Errorf("State = %q, want 'draft'", c.State)
		}
		if c.Title != "Title" {
			t.Errorf("Title = %q, want 'Title'", c.Title)
		}
		if len(c.Fields) != 0 {
			t.Errorf("Fields = %v, want empty map", c.Fields)
		}
	})

	t.Run("stores_created_at", func(t *testing.T) {
		cl := newCL(t)
		c, err := cl.CreateContract("c-cr-ts", "Title", T0, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.CreatedAt != T0 {
			t.Errorf("CreatedAt = %q, want %q", c.CreatedAt, T0)
		}
	})

	t.Run("duplicate_raises_error", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-dup-lc", "A", T0, "alice")
		_, err := cl.CreateContract("c-dup-lc", "B", T1, "alice")
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

}

func TestSetField(t *testing.T) {
	t.Run("sets_field", func(t *testing.T) {
		cl := seededCL(t)
		_, err := cl.SetField("c-seed-1", "value", 100000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		c, _ := cl.GetContract("c-seed-1")
		if c.Fields["value"] != 100000 {
			t.Errorf("Fields[value] = %v, want 100000", c.Fields["value"])
		}
	})

	t.Run("updates_existing_field", func(t *testing.T) {
		cl := seededCL(t)
		cl.SetField("c-seed-1", "value", 50000)
		cl.SetField("c-seed-1", "value", 75000)
		c, _ := cl.GetContract("c-seed-1")
		if c.Fields["value"] != 75000 {
			t.Errorf("Fields[value] = %v, want 75000", c.Fields["value"])
		}
	})

	t.Run("unknown_contract_returns_error", func(t *testing.T) {
		cl := seededCL(t)
		_, err := cl.SetField("no-such", "key", "val")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetContract(t *testing.T) {
	t.Run("returns_contract", func(t *testing.T) {
		cl := seededCL(t)
		c, err := cl.GetContract("c-seed-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ContractID != "c-seed-1" {
			t.Errorf("ContractID = %q, want 'c-seed-1'", c.ContractID)
		}
		if c.State != "approved" {
			t.Errorf("State = %q, want 'approved'", c.State)
		}
	})

	t.Run("unknown_returns_error", func(t *testing.T) {
		cl := seededCL(t)
		_, err := cl.GetContract("nonexistent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestTransition(t *testing.T) {
	t.Run("valid_transition_updates_state", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-tr-1", "T", T0, "alice")
		cl.Transition("c-tr-1", "in_review", T1, "alice")
		c, _ := cl.GetContract("c-tr-1")
		if c.State != "in_review" {
			t.Errorf("State = %q, want 'in_review'", c.State)
		}
	})

	t.Run("invalid_transition_returns_error", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-tr-inv", "T", T0, "alice")
		_, err := cl.Transition("c-tr-inv", "approved", T1, "alice") // draft→approved invalid
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("terminal_state_returns_error", func(t *testing.T) {
		cl := seededCL(t)
		_, err := cl.Transition("c-seed-3", "draft", T5, "alice") // terminated is terminal
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("back_transition_in_review_to_draft", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-back", "T", T0, "alice")
		cl.Transition("c-back", "in_review", T1, "alice")
		cl.Transition("c-back", "draft", T2, "alice")
		c, _ := cl.GetContract("c-back")
		if c.State != "draft" {
			t.Errorf("State = %q, want 'draft'", c.State)
		}
	})

	t.Run("unknown_contract_returns_error", func(t *testing.T) {
		cl := seededCL(t)
		_, err := cl.Transition("no-such", "in_review", T1, "alice")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Audit trail, by-state query, bulk advance
// ---------------------------------------------------------------------------

func TestGetAuditTrail(t *testing.T) {
	t.Run("transition_appends_audit_entry", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-tr-audit", "T", T0, "alice")
		cl.Transition("c-tr-audit", "in_review", T1, "bob")
		trail, _ := cl.GetAuditTrail("c-tr-audit")
		if len(trail) != 2 {
			t.Fatalf("len(trail) = %d, want 2", len(trail))
		}
		last := trail[len(trail)-1]
		if last.FromState == nil || *last.FromState != "draft" {
			t.Errorf("FromState = %v, want 'draft'", last.FromState)
		}
		if last.ToState != "in_review" {
			t.Errorf("ToState = %q, want 'in_review'", last.ToState)
		}
		if last.Actor != "bob" {
			t.Errorf("Actor = %q, want 'bob'", last.Actor)
		}
	})

	t.Run("initial_audit_entry_recorded_on_create", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-audit-init", "T", T0, "alice")
		trail, err := cl.GetAuditTrail("c-audit-init")
		if err != nil {
			t.Fatalf("GetAuditTrail: %v", err)
		}
		if len(trail) != 1 {
			t.Fatalf("len(trail) = %d, want 1", len(trail))
		}
		e := trail[0]
		if e.FromState != nil {
			t.Errorf("FromState = %v, want nil", e.FromState)
		}
		if e.ToState != "draft" {
			t.Errorf("ToState = %q, want 'draft'", e.ToState)
		}
		if e.Actor != "alice" {
			t.Errorf("Actor = %q, want 'alice'", e.Actor)
		}
	})

	t.Run("returns_all_entries_ordered", func(t *testing.T) {
		cl := seededCL(t)
		trail, err := cl.GetAuditTrail("c-seed-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// create(draft) + in_review + approved = 3 entries
		if len(trail) != 3 {
			t.Fatalf("len(trail) = %d, want 3", len(trail))
		}
		want := []string{"draft", "in_review", "approved"}
		for i, e := range trail {
			if e.ToState != want[i] {
				t.Errorf("trail[%d].ToState = %q, want %q", i, e.ToState, want[i])
			}
		}
	})

	t.Run("terminal_contract_full_trail", func(t *testing.T) {
		cl := seededCL(t)
		trail, _ := cl.GetAuditTrail("c-seed-3")
		want := []string{"draft", "in_review", "approved", "executed", "active", "terminated"}
		if len(trail) != len(want) {
			t.Fatalf("len(trail) = %d, want %d", len(trail), len(want))
		}
		for i, e := range trail {
			if e.ToState != want[i] {
				t.Errorf("trail[%d].ToState = %q, want %q", i, e.ToState, want[i])
			}
		}
	})

	t.Run("unknown_contract_returns_error", func(t *testing.T) {
		cl := seededCL(t)
		_, err := cl.GetAuditTrail("no-such")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetContractsByState(t *testing.T) {
	t.Run("returns_correct_contracts", func(t *testing.T) {
		cl := seededCL(t)
		approved := cl.GetContractsByState("approved")
		found := false
		for _, c := range approved {
			if c.ContractID == "c-seed-1" {
				found = true
			}
			if c.ContractID == "c-seed-2" {
				t.Error("c-seed-2 (draft) should not be in approved results")
			}
		}
		if !found {
			t.Error("c-seed-1 not found in approved results")
		}
	})

	t.Run("sorted_by_contract_id", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-z", "Z", T0, "a")
		cl.CreateContract("c-a", "A", T0, "a")
		cl.CreateContract("c-m", "M", T0, "a")
		drafts := cl.GetContractsByState("draft")
		for i := 1; i < len(drafts); i++ {
			if drafts[i].ContractID < drafts[i-1].ContractID {
				t.Errorf("not sorted: %q comes before %q", drafts[i-1].ContractID, drafts[i].ContractID)
			}
		}
	})

	t.Run("empty_for_unused_state", func(t *testing.T) {
		cl := seededCL(t)
		result := cl.GetContractsByState("expired")
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %d contracts", len(result))
		}
	})
}

func TestBulkAdvance(t *testing.T) {
	t.Run("all_succeed", func(t *testing.T) {
		cl := newCL(t)
		for i := 0; i < 3; i++ {
			id := "c-bulk-" + string(rune('0'+i))
			cl.CreateContract(id, "Contract", T0, "alice")
		}
		result := cl.BulkAdvance([]string{"c-bulk-0", "c-bulk-1", "c-bulk-2"}, "in_review", T1, "alice")
		if len(result.Succeeded) != 3 {
			t.Errorf("Succeeded = %d, want 3", len(result.Succeeded))
		}
		if len(result.Failed) != 0 {
			t.Errorf("Failed = %d, want 0", len(result.Failed))
		}
	})

	t.Run("partial_failure_continues", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-ok", "OK", T0, "alice")
		cl.CreateContract("c-bad", "Bad", T0, "alice")
		cl.Transition("c-bad", "in_review", T1, "alice")
		// c-bad is now in_review; trying to move it to in_review again is invalid
		result := cl.BulkAdvance([]string{"c-ok", "c-bad"}, "in_review", T2, "alice")
		foundOK := false
		for _, id := range result.Succeeded {
			if id == "c-ok" {
				foundOK = true
			}
		}
		if !foundOK {
			t.Error("c-ok should be in Succeeded")
		}
		foundBad := false
		for _, f := range result.Failed {
			if f.ContractID == "c-bad" {
				foundBad = true
			}
		}
		if !foundBad {
			t.Error("c-bad should be in Failed")
		}
	})

	t.Run("states_updated_for_successes", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-bs-1", "A", T0, "alice")
		cl.CreateContract("c-bs-2", "B", T0, "alice")
		cl.BulkAdvance([]string{"c-bs-1", "c-bs-2"}, "in_review", T1, "alice")
		c1, _ := cl.GetContract("c-bs-1")
		c2, _ := cl.GetContract("c-bs-2")
		if c1.State != "in_review" {
			t.Errorf("c-bs-1 state = %q, want 'in_review'", c1.State)
		}
		if c2.State != "in_review" {
			t.Errorf("c-bs-2 state = %q, want 'in_review'", c2.State)
		}
	})

	t.Run("failed_entry_includes_reason", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-fail-r", "X", T0, "alice")
		result := cl.BulkAdvance([]string{"c-fail-r"}, "approved", T1, "alice") // draft→approved invalid
		if len(result.Failed) != 1 {
			t.Fatalf("len(Failed) = %d, want 1", len(result.Failed))
		}
		if result.Failed[0].Reason == "" {
			t.Error("Reason should not be empty")
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Lifecycle metrics and overdue contracts
// ---------------------------------------------------------------------------

func TestGetLifecycleMetrics(t *testing.T) {
	t.Run("total_count", func(t *testing.T) {
		cl := seededCL(t)
		m := cl.GetLifecycleMetrics()
		if m.Total != 3 {
			t.Errorf("Total = %d, want 3", m.Total)
		}
	})

	t.Run("by_state_counts", func(t *testing.T) {
		cl := seededCL(t)
		m := cl.GetLifecycleMetrics()
		if m.ByState["approved"] != 1 {
			t.Errorf("ByState[approved] = %d, want 1", m.ByState["approved"])
		}
		if m.ByState["draft"] != 1 {
			t.Errorf("ByState[draft] = %d, want 1", m.ByState["draft"])
		}
		if m.ByState["terminated"] != 1 {
			t.Errorf("ByState[terminated] = %d, want 1", m.ByState["terminated"])
		}
	})

	t.Run("only_nonzero_states_in_by_state", func(t *testing.T) {
		cl := seededCL(t)
		m := cl.GetLifecycleMetrics()
		for state, count := range m.ByState {
			if count <= 0 {
				t.Errorf("ByState[%q] = %d, want > 0", state, count)
			}
		}
	})

	t.Run("terminal_count", func(t *testing.T) {
		cl := seededCL(t)
		m := cl.GetLifecycleMetrics()
		if m.TerminalCount != 1 {
			t.Errorf("TerminalCount = %d, want 1", m.TerminalCount)
		}
	})

	t.Run("empty_manager_returns_zeros", func(t *testing.T) {
		cl := newCL(t)
		m := cl.GetLifecycleMetrics()
		if m.Total != 0 {
			t.Errorf("Total = %d, want 0", m.Total)
		}
		if len(m.ByState) != 0 {
			t.Errorf("ByState = %v, want empty map", m.ByState)
		}
		if m.TerminalCount != 0 {
			t.Errorf("TerminalCount = %d, want 0", m.TerminalCount)
		}
	})
}

func TestGetOverdueContracts(t *testing.T) {
	t.Run("returns_contracts_stuck_over_30_days", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-overdue", "Old Contract", T0, "alice")
		cl.Transition("c-overdue", "in_review", T0, "alice")
		// T4 is ~40 days after T0 → should be overdue
		overdue := cl.GetOverdueContracts(T4)
		found := false
		for _, o := range overdue {
			if o.ContractID == "c-overdue" {
				found = true
			}
		}
		if !found {
			t.Error("c-overdue should be in overdue results")
		}
	})

	t.Run("recent_transition_not_overdue", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-recent", "New Contract", T0, "alice")
		cl.Transition("c-recent", "in_review", T3, "alice")
		// T4 - T3 = ~30 days, not strictly > 30
		overdue := cl.GetOverdueContracts(T4)
		for _, o := range overdue {
			if o.ContractID == "c-recent" {
				t.Error("c-recent should not be overdue (< 30 days)")
			}
		}
	})

	t.Run("terminal_contracts_excluded", func(t *testing.T) {
		cl := seededCL(t)
		overdue := cl.GetOverdueContracts(T5)
		for _, o := range overdue {
			if o.ContractID == "c-seed-3" {
				t.Error("c-seed-3 (terminated) should not appear in overdue")
			}
		}
	})

	t.Run("sorted_by_days_stuck_descending", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-od-a", "A", T0, "alice")
		cl.CreateContract("c-od-b", "B", T0, "alice")
		cl.Transition("c-od-a", "in_review", T0, "alice")
		cl.Transition("c-od-b", "in_review", T1, "alice")
		overdue := cl.GetOverdueContracts(T4)
		for i := 1; i < len(overdue); i++ {
			if overdue[i].DaysStuck > overdue[i-1].DaysStuck {
				t.Errorf("not sorted descending: %d > %d at index %d", overdue[i].DaysStuck, overdue[i-1].DaysStuck, i)
			}
		}
	})

	t.Run("result_includes_required_fields", func(t *testing.T) {
		cl := newCL(t)
		cl.CreateContract("c-od-f", "Fields Test", T0, "alice")
		cl.Transition("c-od-f", "in_review", T0, "alice")
		overdue := cl.GetOverdueContracts(T4)
		var entry *OverdueContract
		for i := range overdue {
			if overdue[i].ContractID == "c-od-f" {
				entry = &overdue[i]
				break
			}
		}
		if entry == nil {
			t.Fatal("c-od-f not found in overdue results")
		}
		if entry.Title == "" {
			t.Error("Title should not be empty")
		}
		if entry.State == "" {
			t.Error("State should not be empty")
		}
		if entry.StuckSince == "" {
			t.Error("StuckSince should not be empty")
		}
		if entry.DaysStuck < 31 {
			t.Errorf("DaysStuck = %d, want >= 31", entry.DaysStuck)
		}
	})

}

