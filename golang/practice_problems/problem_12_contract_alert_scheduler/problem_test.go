// Tests for Problem 12: Contract Expiration Alert Scheduler
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_12_contract_alert_scheduler/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_12_contract_alert_scheduler.go \
//	  -c go test -v .
package alertscheduler

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Shared dates
// D0  = base date
// D30 = D0 + 30 days
// D60 = D0 + 60 days
// D90 = D0 + 90 days
// ---------------------------------------------------------------------------

const (
	D0  = "2025-01-01"
	D30 = "2025-01-31"
	D60 = "2025-03-02" // Jan has 31 days, Feb 2025 has 28 days → Jan 1 + 60 = Mar 2
	D90 = "2025-04-01"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newSched(t *testing.T) ContractAlertScheduler {
	t.Helper()
	return NewContractAlertScheduler()
}

// seededSched returns a pre-seeded ContractAlertScheduler:
//
//	Contracts:
//	  c-seed-1  "Vendor MSA"        legal@acme.com  expires=2025-06-30
//	  c-seed-2  "SaaS Subscription" ops@acme.com    expires=2025-09-15
//	  c-seed-3  "NDA Agreement"     legal@acme.com  expires=2025-12-31
//
//	Alert configs:
//	  cfg-30   days_before=30  label="30-day notice"
//	  cfg-7    days_before=7   label="final warning"
func seededSched(t *testing.T) ContractAlertScheduler {
	t.Helper()
	s := NewContractAlertScheduler()
	mustAddContract(t, s, "c-seed-1", "Vendor MSA", "legal@acme.com", "2025-06-30")
	mustAddContract(t, s, "c-seed-2", "SaaS Subscription", "ops@acme.com", "2025-09-15")
	mustAddContract(t, s, "c-seed-3", "NDA Agreement", "legal@acme.com", "2025-12-31")
	mustAddAlertConfig(t, s, "cfg-30", 30, "30-day notice")
	mustAddAlertConfig(t, s, "cfg-7", 7, "final warning")
	return s
}

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustAddContract(t *testing.T, s ContractAlertScheduler, id, title, email, expires string) *Contract {
	t.Helper()
	c, err := s.AddContract(id, title, email, expires)
	if err != nil {
		t.Fatalf("AddContract %q: %v", id, err)
	}
	return c
}

func mustAddAlertConfig(t *testing.T, s ContractAlertScheduler, id string, days int, label string) *AlertConfig {
	t.Helper()
	cfg, err := s.AddAlertConfig(id, days, label)
	if err != nil {
		t.Fatalf("AddAlertConfig %q: %v", id, err)
	}
	return cfg
}

// ---------------------------------------------------------------------------
// PART 1 — Contract and alert-config management
// ---------------------------------------------------------------------------

func TestAddContract(t *testing.T) {
	t.Run("returns_contract_struct", func(t *testing.T) {
		s := newSched(t)
		c, err := s.AddContract("c-add-1", "Test Contract", "a@b.com", "2025-06-01")
		mustOK(t, err)
		if c.ID != "c-add-1" {
			t.Errorf("ID = %q, want 'c-add-1'", c.ID)
		}
		if c.Title != "Test Contract" {
			t.Errorf("Title = %q, want 'Test Contract'", c.Title)
		}
		if c.OwnerEmail != "a@b.com" {
			t.Errorf("OwnerEmail = %q, want 'a@b.com'", c.OwnerEmail)
		}
		if c.ExpiresOn != "2025-06-01" {
			t.Errorf("ExpiresOn = %q, want '2025-06-01'", c.ExpiresOn)
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		s := newSched(t)
		mustAddContract(t, s, "c-dup-1", "A", "a@b.com", "2025-06-01")
		_, err := s.AddContract("c-dup-1", "B", "b@c.com", "2025-07-01")
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("multiple_contracts_stored", func(t *testing.T) {
		s := seededSched(t)
		results, err := s.GetContractsExpiringBetween("2025-01-01", "2025-12-31")
		mustOK(t, err)
		ids := make(map[string]bool)
		for _, c := range results {
			ids[c.ID] = true
		}
		if !ids["c-seed-1"] || !ids["c-seed-2"] || !ids["c-seed-3"] {
			t.Errorf("missing contracts in results: %v", ids)
		}
	})
}

func TestAddAlertConfig(t *testing.T) {
	t.Run("returns_config_struct", func(t *testing.T) {
		s := newSched(t)
		cfg, err := s.AddAlertConfig("cfg-add-1", 14, "two-week notice")
		mustOK(t, err)
		if cfg.ID != "cfg-add-1" {
			t.Errorf("ID = %q, want 'cfg-add-1'", cfg.ID)
		}
		if cfg.DaysBefore != 14 {
			t.Errorf("DaysBefore = %d, want 14", cfg.DaysBefore)
		}
		if cfg.Label != "two-week notice" {
			t.Errorf("Label = %q, want 'two-week notice'", cfg.Label)
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		s := newSched(t)
		mustAddAlertConfig(t, s, "cfg-dup-1", 30, "notice")
		_, err := s.AddAlertConfig("cfg-dup-1", 60, "other")
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestGetContractsExpiringBetween(t *testing.T) {
	t.Run("exact_range_match", func(t *testing.T) {
		s := seededSched(t)
		results, err := s.GetContractsExpiringBetween("2025-06-30", "2025-06-30")
		mustOK(t, err)
		if len(results) != 1 || results[0].ID != "c-seed-1" {
			t.Errorf("expected [c-seed-1], got %v", results)
		}
	})

	t.Run("range_spans_multiple", func(t *testing.T) {
		s := seededSched(t)
		results, err := s.GetContractsExpiringBetween("2025-06-01", "2025-09-30")
		mustOK(t, err)
		ids := make(map[string]bool)
		for _, c := range results {
			ids[c.ID] = true
		}
		if !ids["c-seed-1"] || !ids["c-seed-2"] {
			t.Error("expected c-seed-1 and c-seed-2 in results")
		}
		if ids["c-seed-3"] {
			t.Error("c-seed-3 should not be in range")
		}
	})

	t.Run("sorted_ascending", func(t *testing.T) {
		s := seededSched(t)
		results, err := s.GetContractsExpiringBetween("2025-01-01", "2025-12-31")
		mustOK(t, err)
		for i := 1; i < len(results); i++ {
			if results[i].ExpiresOn < results[i-1].ExpiresOn {
				t.Errorf("not sorted: %q before %q", results[i-1].ExpiresOn, results[i].ExpiresOn)
			}
		}
	})

	t.Run("empty_when_none_in_range", func(t *testing.T) {
		s := seededSched(t)
		results, err := s.GetContractsExpiringBetween("2024-01-01", "2024-12-31")
		mustOK(t, err)
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("inclusive_start_boundary", func(t *testing.T) {
		s := seededSched(t)
		// c-seed-1 expires exactly on 2025-06-30; start = 2025-06-30 should include it
		results, err := s.GetContractsExpiringBetween("2025-06-30", "2025-12-31")
		mustOK(t, err)
		ids := make(map[string]bool)
		for _, c := range results {
			ids[c.ID] = true
		}
		if !ids["c-seed-1"] {
			t.Error("c-seed-1 should be included at start boundary")
		}
	})

	t.Run("inclusive_end_boundary", func(t *testing.T) {
		s := seededSched(t)
		results, err := s.GetContractsExpiringBetween("2025-01-01", "2025-06-30")
		mustOK(t, err)
		ids := make(map[string]bool)
		for _, c := range results {
			ids[c.ID] = true
		}
		if !ids["c-seed-1"] {
			t.Error("c-seed-1 should be included at end boundary")
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Alert schedule computation
// ---------------------------------------------------------------------------

func TestComputeAlertSchedule(t *testing.T) {
	t.Run("returns_entry_per_config", func(t *testing.T) {
		s := seededSched(t)
		schedule, err := s.ComputeAlertSchedule("c-seed-1")
		mustOK(t, err)
		// 2 configs registered → 2 schedule entries
		if len(schedule) != 2 {
			t.Errorf("expected 2 schedule entries, got %d", len(schedule))
		}
	})

	t.Run("alert_on_dates_correct", func(t *testing.T) {
		s := seededSched(t)
		// c-seed-1 expires 2025-06-30
		// cfg-30: 2025-06-30 - 30d = 2025-05-31
		// cfg-7:  2025-06-30 - 7d  = 2025-06-23
		schedule, err := s.ComputeAlertSchedule("c-seed-1")
		mustOK(t, err)
		byCfg := make(map[string]*AlertScheduleEntry)
		for _, e := range schedule {
			byCfg[e.ConfigID] = e
		}
		if byCfg["cfg-30"] == nil || byCfg["cfg-30"].AlertOn != "2025-05-31" {
			t.Errorf("cfg-30 AlertOn = %v, want '2025-05-31'", byCfg["cfg-30"])
		}
		if byCfg["cfg-7"] == nil || byCfg["cfg-7"].AlertOn != "2025-06-23" {
			t.Errorf("cfg-7 AlertOn = %v, want '2025-06-23'", byCfg["cfg-7"])
		}
	})

	t.Run("sorted_by_alert_on_ascending", func(t *testing.T) {
		s := seededSched(t)
		schedule, err := s.ComputeAlertSchedule("c-seed-1")
		mustOK(t, err)
		for i := 1; i < len(schedule); i++ {
			if schedule[i].AlertOn < schedule[i-1].AlertOn {
				t.Errorf("schedule not sorted: %q before %q", schedule[i-1].AlertOn, schedule[i].AlertOn)
			}
		}
	})

	t.Run("includes_label", func(t *testing.T) {
		s := seededSched(t)
		schedule, err := s.ComputeAlertSchedule("c-seed-1")
		mustOK(t, err)
		labels := make(map[string]bool)
		for _, e := range schedule {
			labels[e.Label] = true
		}
		if !labels["30-day notice"] {
			t.Error("expected '30-day notice' label in schedule")
		}
		if !labels["final warning"] {
			t.Error("expected 'final warning' label in schedule")
		}
	})

	t.Run("unknown_contract_returns_error", func(t *testing.T) {
		s := seededSched(t)
		_, err := s.ComputeAlertSchedule("no-such-contract")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("no_configs_returns_empty_list", func(t *testing.T) {
		s := newSched(t)
		mustAddContract(t, s, "c-no-cfg", "Bare Contract", "a@b.com", "2025-06-01")
		schedule, err := s.ComputeAlertSchedule("c-no-cfg")
		mustOK(t, err)
		if len(schedule) != 0 {
			t.Errorf("expected empty schedule, got %d entries", len(schedule))
		}
	})
}

func TestGetDueAlerts(t *testing.T) {
	t.Run("returns_alerts_on_or_before_date", func(t *testing.T) {
		s := seededSched(t)
		// cfg-30 for c-seed-1 fires on 2025-05-31
		due, err := s.GetDueAlerts("2025-05-31")
		mustOK(t, err)
		found := false
		for _, e := range due {
			if e.ContractID == "c-seed-1" && e.ConfigID == "cfg-30" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected (c-seed-1, cfg-30) in due alerts")
		}
	})

	t.Run("excludes_future_alerts", func(t *testing.T) {
		s := seededSched(t)
		due, err := s.GetDueAlerts("2025-01-01")
		mustOK(t, err)
		if len(due) != 0 {
			t.Errorf("expected 0 due alerts, got %d", len(due))
		}
	})

	t.Run("includes_owner_email_and_expires_on", func(t *testing.T) {
		s := seededSched(t)
		due, err := s.GetDueAlerts("2025-05-31")
		mustOK(t, err)
		var entry *DueAlert
		for _, e := range due {
			if e.ContractID == "c-seed-1" && e.ConfigID == "cfg-30" {
				entry = e
				break
			}
		}
		if entry == nil {
			t.Fatal("(c-seed-1, cfg-30) not found in due alerts")
		}
		if entry.OwnerEmail != "legal@acme.com" {
			t.Errorf("OwnerEmail = %q, want 'legal@acme.com'", entry.OwnerEmail)
		}
		if entry.ExpiresOn != "2025-06-30" {
			t.Errorf("ExpiresOn = %q, want '2025-06-30'", entry.ExpiresOn)
		}
	})

	t.Run("sorted_by_alert_on_then_contract_id", func(t *testing.T) {
		s := seededSched(t)
		due, err := s.GetDueAlerts("2025-12-31")
		mustOK(t, err)
		for i := 1; i < len(due); i++ {
			if due[i].AlertOn < due[i-1].AlertOn {
				t.Errorf("due alerts not sorted by AlertOn: %q before %q", due[i-1].AlertOn, due[i].AlertOn)
			}
			if due[i].AlertOn == due[i-1].AlertOn && due[i].ContractID < due[i-1].ContractID {
				t.Errorf("due alerts not sorted by ContractID on tie: %q before %q", due[i-1].ContractID, due[i].ContractID)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Sent records and upcoming alerts
// ---------------------------------------------------------------------------

func TestRecordAlertSent(t *testing.T) {
	t.Run("returns_sent_record", func(t *testing.T) {
		s := seededSched(t)
		rec, err := s.RecordAlertSent("c-seed-1", "cfg-30", "2025-05-31")
		mustOK(t, err)
		if rec.ContractID != "c-seed-1" {
			t.Errorf("ContractID = %q, want 'c-seed-1'", rec.ContractID)
		}
		if rec.ConfigID != "cfg-30" {
			t.Errorf("ConfigID = %q, want 'cfg-30'", rec.ConfigID)
		}
		if rec.SentOn != "2025-05-31" {
			t.Errorf("SentOn = %q, want '2025-05-31'", rec.SentOn)
		}
	})

	t.Run("unknown_contract_returns_error", func(t *testing.T) {
		s := seededSched(t)
		_, err := s.RecordAlertSent("no-contract", "cfg-30", "2025-05-31")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("unknown_config_returns_error", func(t *testing.T) {
		s := seededSched(t)
		_, err := s.RecordAlertSent("c-seed-1", "no-cfg", "2025-05-31")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("multiple_sends_stored", func(t *testing.T) {
		s := seededSched(t)
		_, err := s.RecordAlertSent("c-seed-1", "cfg-30", "2025-05-31")
		mustOK(t, err)
		_, err = s.RecordAlertSent("c-seed-1", "cfg-7", "2025-06-23")
		mustOK(t, err)
		upcoming, err := s.GetUpcomingAlerts("c-seed-1", "2025-05-01")
		mustOK(t, err)
		sentIDs := make(map[string]bool)
		for _, e := range upcoming {
			if e.Sent {
				sentIDs[e.ConfigID] = true
			}
		}
		if !sentIDs["cfg-30"] {
			t.Error("cfg-30 should be marked as sent")
		}
		if !sentIDs["cfg-7"] {
			t.Error("cfg-7 should be marked as sent")
		}
	})
}

func TestGetUpcomingAlerts(t *testing.T) {
	t.Run("excludes_past_alerts", func(t *testing.T) {
		s := seededSched(t)
		// as_of_date = 2025-06-01; cfg-30 alert_on=2025-05-31 is in the past
		upcoming, err := s.GetUpcomingAlerts("c-seed-1", "2025-06-01")
		mustOK(t, err)
		for _, e := range upcoming {
			if e.ConfigID == "cfg-30" {
				t.Error("cfg-30 (alert_on 2025-05-31) should be excluded as past")
			}
		}
	})

	t.Run("includes_future_alerts", func(t *testing.T) {
		s := seededSched(t)
		// cfg-7 alert_on=2025-06-23 is still upcoming from 2025-06-01
		upcoming, err := s.GetUpcomingAlerts("c-seed-1", "2025-06-01")
		mustOK(t, err)
		found := false
		for _, e := range upcoming {
			if e.ConfigID == "cfg-7" {
				found = true
				break
			}
		}
		if !found {
			t.Error("cfg-7 (alert_on 2025-06-23) should be in upcoming")
		}
	})

	t.Run("sent_flag_false_by_default", func(t *testing.T) {
		s := seededSched(t)
		upcoming, err := s.GetUpcomingAlerts("c-seed-1", "2025-05-01")
		mustOK(t, err)
		for _, e := range upcoming {
			if e.Sent {
				t.Errorf("Sent should be false by default for %q", e.ConfigID)
			}
		}
	})

	t.Run("sent_flag_true_after_record", func(t *testing.T) {
		s := seededSched(t)
		_, err := s.RecordAlertSent("c-seed-1", "cfg-30", "2025-05-31")
		mustOK(t, err)
		upcoming, err := s.GetUpcomingAlerts("c-seed-1", "2025-05-01")
		mustOK(t, err)
		var entry *UpcomingAlert
		for _, e := range upcoming {
			if e.ConfigID == "cfg-30" {
				entry = e
				break
			}
		}
		if entry == nil {
			t.Fatal("cfg-30 not found in upcoming alerts")
		}
		if !entry.Sent {
			t.Error("Sent should be true after RecordAlertSent")
		}
	})

	t.Run("sorted_by_alert_on_ascending", func(t *testing.T) {
		s := seededSched(t)
		upcoming, err := s.GetUpcomingAlerts("c-seed-1", "2025-01-01")
		mustOK(t, err)
		for i := 1; i < len(upcoming); i++ {
			if upcoming[i].AlertOn < upcoming[i-1].AlertOn {
				t.Errorf("upcoming alerts not sorted: %q before %q", upcoming[i-1].AlertOn, upcoming[i].AlertOn)
			}
		}
	})

	t.Run("unknown_contract_returns_error", func(t *testing.T) {
		s := seededSched(t)
		_, err := s.GetUpcomingAlerts("no-such", "2025-01-01")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
