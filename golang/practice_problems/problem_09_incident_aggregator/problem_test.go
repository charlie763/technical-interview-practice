// Tests for Problem 9: Multi-Source Incident Aggregator
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_09_incident_aggregator/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_09_incident_aggregator.go \
//	  -c go test -v .
package incidents

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Shared timestamps (all naive ISO-8601, lexicographically sortable)
// T0 = base
// T1 = T0 + 60s
// T2 = T0 + 120s
// T3 = T0 + 300s
// T4 = T0 + 600s
// ---------------------------------------------------------------------------

const (
	T0 = "2024-06-01T10:00:00"
	T1 = "2024-06-01T10:01:00"
	T2 = "2024-06-01T10:02:00"
	T3 = "2024-06-01T10:05:00"
	T4 = "2024-06-01T10:10:00"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func freshAgg(t *testing.T) IncidentAggregator {
	t.Helper()
	return NewIncidentAggregator()
}

// seededAgg returns:
//
//	Reports: r1 (shooting/downtown/T0), r2 (shooting/downtown/T1),
//	         r3 (car-crash/midtown/T2)
//	Incidents: inc-001 (shooting/downtown) containing r1 and r2
//	r3 is unassigned.
func seededAgg(t *testing.T) IncidentAggregator {
	t.Helper()
	a := NewIncidentAggregator()
	mustIngestReport(t, a, "r1", "radio-north", "shooting", "downtown", T0)
	mustIngestReport(t, a, "r2", "radio-south", "shooting", "downtown", T1)
	mustIngestReport(t, a, "r3", "social-feed", "car-crash", "midtown", T2)
	mustCreateIncident(t, a, "inc-001", "shooting", "downtown")
	mustAddReport(t, a, "inc-001", "r1")
	mustAddReport(t, a, "inc-001", "r2")
	return a
}

func mustIngestReport(t *testing.T, a IncidentAggregator, rid, src, et, loc, ts string) *Report {
	t.Helper()
	r, err := a.IngestReport(rid, src, et, loc, ts)
	if err != nil {
		t.Fatalf("IngestReport(%q): %v", rid, err)
	}
	return r
}

func mustCreateIncident(t *testing.T, a IncidentAggregator, iid, et, loc string) *Incident {
	t.Helper()
	inc, err := a.CreateIncident(iid, et, loc)
	if err != nil {
		t.Fatalf("CreateIncident(%q): %v", iid, err)
	}
	return inc
}

func mustAddReport(t *testing.T, a IncidentAggregator, iid, rid string) {
	t.Helper()
	if err := a.AddReportToIncident(iid, rid); err != nil {
		t.Fatalf("AddReportToIncident(%q, %q): %v", iid, rid, err)
	}
}

// ---------------------------------------------------------------------------
// PART 1 — Report ingestion
// ---------------------------------------------------------------------------

func TestIngestReport(t *testing.T) {
	t.Run("stores_and_returns_report", func(t *testing.T) {
		a := freshAgg(t)
		r, err := a.IngestReport("r_store", "radio-north", "shooting", "downtown", T0)
		if err != nil {
			t.Fatalf("IngestReport: %v", err)
		}
		if r.ReportID != "r_store" {
			t.Errorf("ReportID = %q, want 'r_store'", r.ReportID)
		}
		if r.SourceID != "radio-north" {
			t.Errorf("SourceID = %q, want 'radio-north'", r.SourceID)
		}
		if r.EventType != "shooting" {
			t.Errorf("EventType = %q, want 'shooting'", r.EventType)
		}
		if r.LocationKey != "downtown" {
			t.Errorf("LocationKey = %q, want 'downtown'", r.LocationKey)
		}
		if r.Ts != T0 {
			t.Errorf("Ts = %q, want %q", r.Ts, T0)
		}
		if r.IncidentID != nil {
			t.Errorf("IncidentID = %v, want nil", r.IncidentID)
		}
	})

	t.Run("duplicate_report_id_returns_error", func(t *testing.T) {
		a := freshAgg(t)
		mustIngestReport(t, a, "r_dup", "src-a", "fire", "east", T0)
		if _, err := a.IngestReport("r_dup", "src-b", "fire", "east", T1); !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("get_report_returns_nil_if_missing", func(t *testing.T) {
		a := freshAgg(t)
		if got := a.GetReport("nonexistent"); got != nil {
			t.Errorf("GetReport = %v, want nil", got)
		}
	})

	t.Run("get_report_returns_stored", func(t *testing.T) {
		a := seededAgg(t)
		r := a.GetReport("r1")
		if r == nil {
			t.Fatal("GetReport returned nil")
		}
		if r.EventType != "shooting" {
			t.Errorf("EventType = %q, want 'shooting'", r.EventType)
		}
	})
}

func TestGetReports(t *testing.T) {
	t.Run("no_filter_returns_all", func(t *testing.T) {
		a := seededAgg(t)
		if got := a.GetReports("", ""); len(got) != 3 {
			t.Errorf("len(GetReports) = %d, want 3", len(got))
		}
	})

	t.Run("filter_by_location", func(t *testing.T) {
		a := seededAgg(t)
		reports := a.GetReports("downtown", "")
		if len(reports) != 2 {
			t.Fatalf("len = %d, want 2", len(reports))
		}
		for _, r := range reports {
			if r.LocationKey != "downtown" {
				t.Errorf("unexpected LocationKey = %q", r.LocationKey)
			}
		}
	})

	t.Run("filter_by_event_type", func(t *testing.T) {
		a := seededAgg(t)
		reports := a.GetReports("", "car-crash")
		if len(reports) != 1 {
			t.Fatalf("len = %d, want 1", len(reports))
		}
		if reports[0].ReportID != "r3" {
			t.Errorf("ReportID = %q, want 'r3'", reports[0].ReportID)
		}
	})

	t.Run("filter_both", func(t *testing.T) {
		a := seededAgg(t)
		reports := a.GetReports("downtown", "shooting")
		if len(reports) != 2 {
			t.Errorf("len = %d, want 2", len(reports))
		}
	})

	t.Run("empty_when_no_match", func(t *testing.T) {
		a := seededAgg(t)
		if got := a.GetReports("mars", ""); len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("sorted_by_ts_ascending", func(t *testing.T) {
		a := freshAgg(t)
		mustIngestReport(t, a, "r_sort_b", "src", "fire", "zone-1", T1)
		mustIngestReport(t, a, "r_sort_a", "src", "fire", "zone-1", T0)
		reports := a.GetReports("", "")
		for i := 1; i < len(reports); i++ {
			if reports[i].Ts < reports[i-1].Ts {
				t.Errorf("reports not sorted by Ts: %v", reports)
				break
			}
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Manual incident grouping
// ---------------------------------------------------------------------------

func TestCreateIncident(t *testing.T) {
	t.Run("creates_empty_incident", func(t *testing.T) {
		a := freshAgg(t)
		inc, err := a.CreateIncident("inc_create_test", "fire", "east-side")
		if err != nil {
			t.Fatalf("CreateIncident: %v", err)
		}
		if inc.IncidentID != "inc_create_test" {
			t.Errorf("IncidentID = %q, want 'inc_create_test'", inc.IncidentID)
		}
		if inc.EventType != "fire" {
			t.Errorf("EventType = %q, want 'fire'", inc.EventType)
		}
		if inc.LocationKey != "east-side" {
			t.Errorf("LocationKey = %q, want 'east-side'", inc.LocationKey)
		}
		if inc.ReportCount != 0 {
			t.Errorf("ReportCount = %d, want 0", inc.ReportCount)
		}
		if len(inc.ReportIDs) != 0 {
			t.Errorf("ReportIDs = %v, want empty", inc.ReportIDs)
		}
		if inc.LatestTs != nil {
			t.Errorf("LatestTs = %v, want nil", inc.LatestTs)
		}
	})

	t.Run("duplicate_incident_id_returns_error", func(t *testing.T) {
		a := seededAgg(t)
		if _, err := a.CreateIncident("inc-001", "shooting", "downtown"); !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestAddReportToIncident(t *testing.T) {
	t.Run("assigns_report", func(t *testing.T) {
		a := freshAgg(t)
		mustIngestReport(t, a, "r_assign", "src", "fire", "east", T0)
		mustCreateIncident(t, a, "inc_assign_test", "fire", "east")
		mustAddReport(t, a, "inc_assign_test", "r_assign")
		inc := a.GetIncident("inc_assign_test")
		if inc == nil {
			t.Fatal("GetIncident returned nil")
		}
		found := false
		for _, rid := range inc.ReportIDs {
			if rid == "r_assign" {
				found = true
			}
		}
		if !found {
			t.Error("r_assign not found in ReportIDs")
		}
		if inc.ReportCount != 1 {
			t.Errorf("ReportCount = %d, want 1", inc.ReportCount)
		}
		if inc.LatestTs == nil || *inc.LatestTs != T0 {
			t.Errorf("LatestTs = %v, want %q", inc.LatestTs, T0)
		}
	})

	t.Run("updates_report_incident_id", func(t *testing.T) {
		a := freshAgg(t)
		mustIngestReport(t, a, "r_update", "src", "fire", "east", T0)
		mustCreateIncident(t, a, "inc_update_test", "fire", "east")
		mustAddReport(t, a, "inc_update_test", "r_update")
		r := a.GetReport("r_update")
		if r == nil || r.IncidentID == nil || *r.IncidentID != "inc_update_test" {
			t.Errorf("Report.IncidentID = %v, want 'inc_update_test'", r.IncidentID)
		}
	})

	t.Run("already_assigned_returns_error", func(t *testing.T) {
		a := seededAgg(t)
		// r1 is already in inc-001
		if err := a.AddReportToIncident("inc-001", "r1"); !errors.Is(err, ErrAlreadyAssigned) {
			t.Errorf("expected ErrAlreadyAssigned, got %v", err)
		}
	})

	t.Run("bad_incident_returns_not_found", func(t *testing.T) {
		a := freshAgg(t)
		mustIngestReport(t, a, "r_bad_inc", "src", "fire", "east", T0)
		if err := a.AddReportToIncident("nonexistent_inc", "r_bad_inc"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("bad_report_returns_not_found", func(t *testing.T) {
		a := seededAgg(t)
		if err := a.AddReportToIncident("inc-001", "nonexistent_report"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("report_ids_sorted_by_ts", func(t *testing.T) {
		a := freshAgg(t)
		mustIngestReport(t, a, "r_ts_b", "src", "fire", "east", T1)
		mustIngestReport(t, a, "r_ts_a", "src", "fire", "east", T0)
		mustCreateIncident(t, a, "inc_ts_test", "fire", "east")
		mustAddReport(t, a, "inc_ts_test", "r_ts_b")
		mustAddReport(t, a, "inc_ts_test", "r_ts_a")
		inc := a.GetIncident("inc_ts_test")
		if inc == nil {
			t.Fatal("GetIncident returned nil")
		}
		if len(inc.ReportIDs) != 2 || inc.ReportIDs[0] != "r_ts_a" || inc.ReportIDs[1] != "r_ts_b" {
			t.Errorf("ReportIDs = %v, want ['r_ts_a', 'r_ts_b']", inc.ReportIDs)
		}
	})

	t.Run("latest_ts_updates_to_newest", func(t *testing.T) {
		a := freshAgg(t)
		mustIngestReport(t, a, "r_lt_a", "src", "fire", "east", T0)
		mustIngestReport(t, a, "r_lt_b", "src", "fire", "east", T2)
		mustCreateIncident(t, a, "inc_lt_test", "fire", "east")
		mustAddReport(t, a, "inc_lt_test", "r_lt_a")
		mustAddReport(t, a, "inc_lt_test", "r_lt_b")
		inc := a.GetIncident("inc_lt_test")
		if inc == nil || inc.LatestTs == nil || *inc.LatestTs != T2 {
			t.Errorf("LatestTs = %v, want %q", inc.LatestTs, T2)
		}
	})
}

func TestGetIncident(t *testing.T) {
	t.Run("returns_nil_if_not_found", func(t *testing.T) {
		a := freshAgg(t)
		if got := a.GetIncident("ghost"); got != nil {
			t.Errorf("GetIncident = %v, want nil", got)
		}
	})

	t.Run("returns_incident", func(t *testing.T) {
		a := seededAgg(t)
		inc := a.GetIncident("inc-001")
		if inc == nil {
			t.Fatal("GetIncident returned nil")
		}
		if inc.ReportCount != 2 {
			t.Errorf("ReportCount = %d, want 2", inc.ReportCount)
		}
	})
}

func TestGetUnassignedReports(t *testing.T) {
	t.Run("returns_unassigned", func(t *testing.T) {
		a := seededAgg(t)
		unassigned := a.GetUnassignedReports()
		if len(unassigned) != 1 {
			t.Fatalf("len = %d, want 1", len(unassigned))
		}
		if unassigned[0].ReportID != "r3" {
			t.Errorf("ReportID = %q, want 'r3'", unassigned[0].ReportID)
		}
	})

	t.Run("empty_when_all_assigned", func(t *testing.T) {
		a := freshAgg(t)
		mustIngestReport(t, a, "r_all", "src", "fire", "east", T0)
		mustCreateIncident(t, a, "inc_all_test", "fire", "east")
		mustAddReport(t, a, "inc_all_test", "r_all")
		if got := a.GetUnassignedReports(); len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("sorted_by_ts_ascending", func(t *testing.T) {
		a := freshAgg(t)
		mustIngestReport(t, a, "r_ua_b", "src", "fire", "east", T1)
		mustIngestReport(t, a, "r_ua_a", "src", "fire", "east", T0)
		unassigned := a.GetUnassignedReports()
		for i := 1; i < len(unassigned); i++ {
			if unassigned[i].Ts < unassigned[i-1].Ts {
				t.Errorf("unassigned not sorted by Ts: %v", unassigned)
				break
			}
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Automatic deduplication
// ---------------------------------------------------------------------------

func TestAutoIngestReport(t *testing.T) {
	t.Run("creates_new_incident_when_no_match", func(t *testing.T) {
		a := freshAgg(t)
		incidentID, err := a.AutoIngestReport("r_auto_new", "src", "fire", "east-side", T0, 120)
		if err != nil {
			t.Fatalf("AutoIngestReport: %v", err)
		}
		if incidentID == "" {
			t.Fatal("returned empty incidentID")
		}
		inc := a.GetIncident(incidentID)
		if inc == nil {
			t.Fatalf("GetIncident(%q) returned nil", incidentID)
		}
		if inc.ReportCount != 1 {
			t.Errorf("ReportCount = %d, want 1", inc.ReportCount)
		}
		found := false
		for _, rid := range inc.ReportIDs {
			if rid == "r_auto_new" {
				found = true
			}
		}
		if !found {
			t.Error("r_auto_new not found in ReportIDs")
		}
	})

	t.Run("merges_into_existing_when_within_window", func(t *testing.T) {
		a := seededAgg(t)
		// inc-001 latest_ts = T1; new ts = T2; T2 - T1 = 60s <= 120s
		result, err := a.AutoIngestReport("r_merge", "radio-east", "shooting", "downtown", T2, 120)
		if err != nil {
			t.Fatalf("AutoIngestReport: %v", err)
		}
		if result != "inc-001" {
			t.Errorf("incidentID = %q, want 'inc-001'", result)
		}
		inc := a.GetIncident("inc-001")
		if inc.ReportCount != 3 {
			t.Errorf("ReportCount = %d, want 3", inc.ReportCount)
		}
	})

	t.Run("no_merge_different_event_type", func(t *testing.T) {
		a := seededAgg(t)
		result, _ := a.AutoIngestReport("r_diff_type", "src", "car-crash", "downtown", T2, 120)
		if result == "inc-001" {
			t.Error("should not merge into inc-001 (different event type)")
		}
	})

	t.Run("no_merge_different_location", func(t *testing.T) {
		a := seededAgg(t)
		result, _ := a.AutoIngestReport("r_diff_loc", "src", "shooting", "uptown", T2, 120)
		if result == "inc-001" {
			t.Error("should not merge into inc-001 (different location)")
		}
	})

	t.Run("no_merge_outside_window", func(t *testing.T) {
		a := seededAgg(t)
		// inc-001 latest_ts = T1; new ts = T4; T4 - T1 = 540s > 120s
		result, _ := a.AutoIngestReport("r_expired", "src", "shooting", "downtown", T4, 120)
		if result == "inc-001" {
			t.Error("should not merge into inc-001 (outside time window)")
		}
	})

	t.Run("new_report_stored", func(t *testing.T) {
		a := freshAgg(t)
		a.AutoIngestReport("r_stored", "src", "fire", "east", T0, 60)
		if got := a.GetReport("r_stored"); got == nil {
			t.Error("report was not stored after AutoIngestReport")
		}
	})

	t.Run("picks_most_recently_active_match", func(t *testing.T) {
		a := freshAgg(t)
		// Two incidents for the same type+location, both within the window
		mustIngestReport(t, a, "ra1", "src", "shooting", "downtown", T0)
		mustIngestReport(t, a, "ra2", "src", "shooting", "downtown", T1)
		mustCreateIncident(t, a, "inc_older", "shooting", "downtown")
		mustAddReport(t, a, "inc_older", "ra1") // latest_ts = T0
		mustCreateIncident(t, a, "inc_newer", "shooting", "downtown")
		mustAddReport(t, a, "inc_newer", "ra2") // latest_ts = T1

		// T2 - T0 = 120s <= 300s and T2 - T1 = 60s <= 300s → both active
		result, err := a.AutoIngestReport("r_pick", "src", "shooting", "downtown", T2, 300)
		if err != nil {
			t.Fatalf("AutoIngestReport: %v", err)
		}
		// inc_newer is more recently active (T1 > T0)
		if result != "inc_newer" {
			t.Errorf("incidentID = %q, want 'inc_newer'", result)
		}
	})

	t.Run("auto_generated_id_does_not_clash", func(t *testing.T) {
		a := freshAgg(t)
		id1, _ := a.AutoIngestReport("r_id1", "src", "fire", "west", T0, 0)
		id2, _ := a.AutoIngestReport("r_id2", "src", "fire", "east", T1, 0)
		if id1 == id2 {
			t.Errorf("auto-generated IDs should be unique, both got %q", id1)
		}
	})
}

func TestGetActiveIncidents(t *testing.T) {
	t.Run("returns_active_incidents", func(t *testing.T) {
		a := seededAgg(t)
		// inc-001 latest_ts=T1; T2-T1=60s <= 120s → active
		active := a.GetActiveIncidents(T2, 120)
		found := false
		for _, inc := range active {
			if inc.IncidentID == "inc-001" {
				found = true
			}
		}
		if !found {
			t.Error("expected inc-001 to be active")
		}
	})

	t.Run("excludes_stale_incidents", func(t *testing.T) {
		a := seededAgg(t)
		// inc-001 latest_ts=T1; T4-T1=540s > 120s → not active
		active := a.GetActiveIncidents(T4, 120)
		for _, inc := range active {
			if inc.IncidentID == "inc-001" {
				t.Error("inc-001 should be stale and excluded")
			}
		}
	})

	t.Run("excludes_empty_incidents", func(t *testing.T) {
		a := freshAgg(t)
		mustCreateIncident(t, a, "inc_empty_active", "fire", "north")
		active := a.GetActiveIncidents(T0, 300)
		for _, inc := range active {
			if inc.IncidentID == "inc_empty_active" {
				t.Error("empty incident (no reports) should be excluded from active")
			}
		}
	})

	t.Run("sorted_by_latest_ts_descending", func(t *testing.T) {
		a := freshAgg(t)
		mustIngestReport(t, a, "ri1", "src", "fire", "east", T0)
		mustIngestReport(t, a, "ri2", "src", "fire", "west", T2)
		mustCreateIncident(t, a, "inc_sort_a", "fire", "east")
		mustAddReport(t, a, "inc_sort_a", "ri1") // latest_ts = T0
		mustCreateIncident(t, a, "inc_sort_b", "fire", "west")
		mustAddReport(t, a, "inc_sort_b", "ri2") // latest_ts = T2

		active := a.GetActiveIncidents(T3, 600)
		idxA, idxB := -1, -1
		for i, inc := range active {
			if inc.IncidentID == "inc_sort_a" {
				idxA = i
			}
			if inc.IncidentID == "inc_sort_b" {
				idxB = i
			}
		}
		if idxA == -1 || idxB == -1 {
			t.Fatalf("one or both incidents missing from active list: %v", active)
		}
		if idxB > idxA {
			t.Error("inc_sort_b (newer) should come before inc_sort_a (older) in descending order")
		}
	})
}
