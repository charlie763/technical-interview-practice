// Tests for Problem 11: Sensor Coverage Tracker
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_11_coverage_tracker/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_11_coverage_tracker.go \
//	  -c go test -v .
package coverage

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Shared timestamps
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

func newCT(t *testing.T) CoverageTracker {
	t.Helper()
	return NewCoverageTracker()
}

// seededCT returns a pre-seeded CoverageTracker:
//
//	sta-seed-1  North Tower  downtown  last hb: T0
//	sta-seed-2  South Tower  downtown  last hb: T1
//	sta-seed-3  East Hub     eastside  (never sent a heartbeat)
func seededCT(t *testing.T) CoverageTracker {
	t.Helper()
	ct := NewCoverageTracker()
	mustRegisterStation(t, ct, "sta-seed-1", "North Tower", "downtown")
	mustRegisterStation(t, ct, "sta-seed-2", "South Tower", "downtown")
	mustRegisterStation(t, ct, "sta-seed-3", "East Hub", "eastside")
	mustOK(t, ct.RecordHeartbeat("sta-seed-1", T0))
	mustOK(t, ct.RecordHeartbeat("sta-seed-2", T1))
	// sta-seed-3 intentionally has no heartbeat
	return ct
}

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustRegisterStation(t *testing.T, ct CoverageTracker, id, name, region string) *Station {
	t.Helper()
	s, err := ct.RegisterStation(id, name, region)
	if err != nil {
		t.Fatalf("RegisterStation %q: %v", id, err)
	}
	return s
}

// ---------------------------------------------------------------------------
// PART 1 — Station registration and heartbeats
// ---------------------------------------------------------------------------

func TestRegisterStation(t *testing.T) {
	t.Run("stores_and_returns_station", func(t *testing.T) {
		ct := newCT(t)
		s, err := ct.RegisterStation("sta-reg", "Tower", "north")
		mustOK(t, err)
		if s.ID != "sta-reg" {
			t.Errorf("ID = %q, want 'sta-reg'", s.ID)
		}
		if s.Name != "Tower" {
			t.Errorf("Name = %q, want 'Tower'", s.Name)
		}
		if s.Region != "north" {
			t.Errorf("Region = %q, want 'north'", s.Region)
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		ct := seededCT(t)
		_, err := ct.RegisterStation("sta-seed-1", "Dup", "downtown")
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestRecordHeartbeat(t *testing.T) {
	t.Run("updates_last_heartbeat", func(t *testing.T) {
		ct := seededCT(t)
		mustOK(t, ct.RecordHeartbeat("sta-seed-1", T2))
		hb, err := ct.GetLastHeartbeat("sta-seed-1")
		mustOK(t, err)
		if hb != T2 {
			t.Errorf("GetLastHeartbeat = %q, want %q", hb, T2)
		}
	})

	t.Run("missing_station_returns_error", func(t *testing.T) {
		ct := seededCT(t)
		err := ct.RecordHeartbeat("ghost", T0)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("same_ts_returns_error", func(t *testing.T) {
		ct := seededCT(t)
		// sta-seed-1 last hb is T0; equal ts should be rejected
		err := ct.RecordHeartbeat("sta-seed-1", T0)
		if !errors.Is(err, ErrOutOfOrder) {
			t.Errorf("expected ErrOutOfOrder, got %v", err)
		}
	})

	t.Run("earlier_ts_returns_error", func(t *testing.T) {
		ct := seededCT(t)
		// sta-seed-2 last hb is T1; T0 < T1 should be rejected
		err := ct.RecordHeartbeat("sta-seed-2", T0)
		if !errors.Is(err, ErrOutOfOrder) {
			t.Errorf("expected ErrOutOfOrder, got %v", err)
		}
	})

	t.Run("first_heartbeat_accepted", func(t *testing.T) {
		ct := seededCT(t)
		// sta-seed-3 has never sent one
		mustOK(t, ct.RecordHeartbeat("sta-seed-3", T0))
		hb, err := ct.GetLastHeartbeat("sta-seed-3")
		mustOK(t, err)
		if hb != T0 {
			t.Errorf("GetLastHeartbeat = %q, want %q", hb, T0)
		}
	})
}

func TestGetLastHeartbeat(t *testing.T) {
	t.Run("returns_empty_if_never_sent", func(t *testing.T) {
		ct := seededCT(t)
		hb, err := ct.GetLastHeartbeat("sta-seed-3")
		mustOK(t, err)
		if hb != "" {
			t.Errorf("GetLastHeartbeat = %q, want empty string", hb)
		}
	})

	t.Run("returns_latest_ts", func(t *testing.T) {
		ct := seededCT(t)
		hb, err := ct.GetLastHeartbeat("sta-seed-1")
		mustOK(t, err)
		if hb != T0 {
			t.Errorf("GetLastHeartbeat = %q, want %q", hb, T0)
		}
	})

	t.Run("missing_station_returns_error", func(t *testing.T) {
		ct := seededCT(t)
		_, err := ct.GetLastHeartbeat("ghost")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetStations(t *testing.T) {
	t.Run("no_filter_returns_all", func(t *testing.T) {
		ct := seededCT(t)
		stations, err := ct.GetStations("")
		mustOK(t, err)
		if len(stations) != 3 {
			t.Errorf("expected 3 stations, got %d", len(stations))
		}
	})

	t.Run("filter_by_region", func(t *testing.T) {
		ct := seededCT(t)
		stations, err := ct.GetStations("downtown")
		mustOK(t, err)
		if len(stations) != 2 {
			t.Errorf("expected 2 downtown stations, got %d", len(stations))
		}
		for _, s := range stations {
			if s.Region != "downtown" {
				t.Errorf("station %q has region %q, want 'downtown'", s.ID, s.Region)
			}
		}
	})

	t.Run("unknown_region_returns_empty", func(t *testing.T) {
		ct := seededCT(t)
		stations, err := ct.GetStations("nowhere")
		mustOK(t, err)
		if len(stations) != 0 {
			t.Errorf("expected 0 stations for unknown region, got %d", len(stations))
		}
	})

	t.Run("sorted_by_station_id", func(t *testing.T) {
		ct := seededCT(t)
		stations, err := ct.GetStations("")
		mustOK(t, err)
		for i := 1; i < len(stations); i++ {
			if stations[i].ID < stations[i-1].ID {
				t.Errorf("stations not sorted: %q before %q", stations[i-1].ID, stations[i].ID)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Staleness detection and outage tracking
// ---------------------------------------------------------------------------

func TestGetStaleStations(t *testing.T) {
	t.Run("never_heartbeat_is_stale", func(t *testing.T) {
		ct := seededCT(t)
		stale, err := ct.GetStaleStations(T2, 120)
		mustOK(t, err)
		found := false
		for _, s := range stale {
			if s.ID == "sta-seed-3" {
				found = true
				break
			}
		}
		if !found {
			t.Error("sta-seed-3 (no heartbeat) should be stale")
		}
	})

	t.Run("recent_heartbeat_not_stale", func(t *testing.T) {
		ct := seededCT(t)
		// sta-seed-2 hb=T1; T2-T1=60s <= 120s → not stale
		stale, err := ct.GetStaleStations(T2, 120)
		mustOK(t, err)
		for _, s := range stale {
			if s.ID == "sta-seed-2" {
				t.Error("sta-seed-2 should not be stale (recent heartbeat)")
			}
		}
	})

	t.Run("old_heartbeat_is_stale", func(t *testing.T) {
		ct := seededCT(t)
		// sta-seed-1 hb=T0; T3-T0=300s > 120s → stale
		stale, err := ct.GetStaleStations(T3, 120)
		mustOK(t, err)
		found := false
		for _, s := range stale {
			if s.ID == "sta-seed-1" {
				found = true
				break
			}
		}
		if !found {
			t.Error("sta-seed-1 should be stale (old heartbeat)")
		}
	})

	t.Run("all_stale_when_threshold_is_tiny", func(t *testing.T) {
		ct := seededCT(t)
		stale, err := ct.GetStaleStations(T4, 5)
		mustOK(t, err)
		if len(stale) != 3 {
			t.Errorf("expected all 3 stations stale, got %d", len(stale))
		}
	})

	t.Run("sorted_by_station_id", func(t *testing.T) {
		ct := seededCT(t)
		stale, err := ct.GetStaleStations(T4, 5)
		mustOK(t, err)
		for i := 1; i < len(stale); i++ {
			if stale[i].ID < stale[i-1].ID {
				t.Errorf("stale stations not sorted: %q before %q", stale[i-1].ID, stale[i].ID)
			}
		}
	})
}

func TestRecordOutageStartEnd(t *testing.T) {
	t.Run("open_outage_recorded", func(t *testing.T) {
		ct := newCT(t)
		mustRegisterStation(t, ct, "sta-out", "T", "north")
		mustOK(t, ct.RecordOutageStart("sta-out", T0))
		outages, err := ct.GetOutages("sta-out")
		mustOK(t, err)
		if len(outages) != 1 {
			t.Fatalf("expected 1 outage, got %d", len(outages))
		}
		if outages[0].StartTs != T0 {
			t.Errorf("StartTs = %q, want %q", outages[0].StartTs, T0)
		}
		if outages[0].EndTs != nil {
			t.Errorf("EndTs = %v, want nil (still open)", outages[0].EndTs)
		}
	})

	t.Run("duplicate_open_outage_raises", func(t *testing.T) {
		ct := newCT(t)
		mustRegisterStation(t, ct, "sta-dup-out", "T", "north")
		mustOK(t, ct.RecordOutageStart("sta-dup-out", T0))
		err := ct.RecordOutageStart("sta-dup-out", T1)
		if !errors.Is(err, ErrOpenOutageExists) {
			t.Errorf("expected ErrOpenOutageExists, got %v", err)
		}
	})

	t.Run("end_closes_outage", func(t *testing.T) {
		ct := newCT(t)
		mustRegisterStation(t, ct, "sta-end-out", "T", "north")
		mustOK(t, ct.RecordOutageStart("sta-end-out", T0))
		mustOK(t, ct.RecordOutageEnd("sta-end-out", T1))
		outages, err := ct.GetOutages("sta-end-out")
		mustOK(t, err)
		if outages[0].EndTs == nil || *outages[0].EndTs != T1 {
			t.Errorf("EndTs = %v, want %q", outages[0].EndTs, T1)
		}
	})

	t.Run("end_with_no_open_outage_returns_error", func(t *testing.T) {
		ct := seededCT(t)
		// sta-seed-1 has no open outage
		err := ct.RecordOutageEnd("sta-seed-1", T2)
		if !errors.Is(err, ErrNoOpenOutage) {
			t.Errorf("expected ErrNoOpenOutage, got %v", err)
		}
	})

	t.Run("start_on_missing_station_returns_error", func(t *testing.T) {
		ct := seededCT(t)
		err := ct.RecordOutageStart("ghost", T0)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("end_on_missing_station_returns_error", func(t *testing.T) {
		ct := seededCT(t)
		err := ct.RecordOutageEnd("ghost", T0)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("second_outage_allowed_after_first_closed", func(t *testing.T) {
		ct := newCT(t)
		mustRegisterStation(t, ct, "sta-2nd-out", "T", "north")
		mustOK(t, ct.RecordOutageStart("sta-2nd-out", T0))
		mustOK(t, ct.RecordOutageEnd("sta-2nd-out", T1))
		mustOK(t, ct.RecordOutageStart("sta-2nd-out", T2)) // should not error
		outages, err := ct.GetOutages("sta-2nd-out")
		mustOK(t, err)
		if len(outages) != 2 {
			t.Errorf("expected 2 outages, got %d", len(outages))
		}
	})

	t.Run("outages_sorted_by_start_ts", func(t *testing.T) {
		ct := newCT(t)
		mustRegisterStation(t, ct, "sta-sort-out", "T", "north")
		mustOK(t, ct.RecordOutageStart("sta-sort-out", T0))
		mustOK(t, ct.RecordOutageEnd("sta-sort-out", T1))
		mustOK(t, ct.RecordOutageStart("sta-sort-out", T2))
		mustOK(t, ct.RecordOutageEnd("sta-sort-out", T3))
		outages, err := ct.GetOutages("sta-sort-out")
		mustOK(t, err)
		if outages[0].StartTs != T0 {
			t.Errorf("first outage StartTs = %q, want %q", outages[0].StartTs, T0)
		}
		if outages[1].StartTs != T2 {
			t.Errorf("second outage StartTs = %q, want %q", outages[1].StartTs, T2)
		}
	})

	t.Run("get_outages_missing_station_returns_error", func(t *testing.T) {
		ct := seededCT(t)
		_, err := ct.GetOutages("ghost")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Coverage analysis
// ---------------------------------------------------------------------------

func TestGetRegionCoverage(t *testing.T) {
	t.Run("basic_partial_coverage", func(t *testing.T) {
		ct := seededCT(t)
		// as_of=T2, threshold=90s
		// sta-seed-1: T2-T0=120s > 90s → stale
		// sta-seed-2: T2-T1=60s  ≤ 90s → healthy
		cov, err := ct.GetRegionCoverage("downtown", T2, 90)
		mustOK(t, err)
		if cov.Region != "downtown" {
			t.Errorf("Region = %q, want 'downtown'", cov.Region)
		}
		if cov.Total != 2 {
			t.Errorf("Total = %d, want 2", cov.Total)
		}
		if cov.Healthy != 1 {
			t.Errorf("Healthy = %d, want 1", cov.Healthy)
		}
		if cov.Stale != 1 {
			t.Errorf("Stale = %d, want 1", cov.Stale)
		}
		if !cov.HasCoverage {
			t.Error("HasCoverage should be true (1 healthy)")
		}
	})

	t.Run("full_coverage", func(t *testing.T) {
		ct := seededCT(t)
		// as_of=T2, threshold=300s → both hbs are within window
		cov, err := ct.GetRegionCoverage("downtown", T2, 300)
		mustOK(t, err)
		if cov.Healthy != 2 {
			t.Errorf("Healthy = %d, want 2", cov.Healthy)
		}
		if cov.Stale != 0 {
			t.Errorf("Stale = %d, want 0", cov.Stale)
		}
		if !cov.HasCoverage {
			t.Error("HasCoverage should be true")
		}
	})

	t.Run("no_coverage_all_stale", func(t *testing.T) {
		ct := seededCT(t)
		// as_of=T4, threshold=30s → both hbs are way older than 30s
		cov, err := ct.GetRegionCoverage("downtown", T4, 30)
		mustOK(t, err)
		if cov.Healthy != 0 {
			t.Errorf("Healthy = %d, want 0", cov.Healthy)
		}
		if cov.HasCoverage {
			t.Error("HasCoverage should be false (all stale)")
		}
	})

	t.Run("empty_region_returns_zeros", func(t *testing.T) {
		ct := newCT(t)
		cov, err := ct.GetRegionCoverage("ghost-region", T0, 60)
		mustOK(t, err)
		if cov.Total != 0 {
			t.Errorf("Total = %d, want 0", cov.Total)
		}
		if cov.Healthy != 0 {
			t.Errorf("Healthy = %d, want 0", cov.Healthy)
		}
		if cov.HasCoverage {
			t.Error("HasCoverage should be false for empty region")
		}
	})

	t.Run("correct_total_for_region", func(t *testing.T) {
		ct := seededCT(t)
		cov, err := ct.GetRegionCoverage("eastside", T0, 60)
		mustOK(t, err)
		if cov.Total != 1 {
			t.Errorf("Total = %d, want 1", cov.Total)
		}
	})
}

func TestGetOutageSummary(t *testing.T) {
	t.Run("no_outages", func(t *testing.T) {
		ct := seededCT(t)
		summary, err := ct.GetOutageSummary("sta-seed-1", T4)
		mustOK(t, err)
		if summary.StationID != "sta-seed-1" {
			t.Errorf("StationID = %q, want 'sta-seed-1'", summary.StationID)
		}
		if summary.TotalOutages != 0 {
			t.Errorf("TotalOutages = %d, want 0", summary.TotalOutages)
		}
		if summary.OpenOutage {
			t.Error("OpenOutage should be false")
		}
		if summary.TotalOutageSecs != 0 {
			t.Errorf("TotalOutageSecs = %d, want 0", summary.TotalOutageSecs)
		}
	})

	t.Run("closed_outage_duration", func(t *testing.T) {
		ct := newCT(t)
		mustRegisterStation(t, ct, "sta-dur", "T", "north")
		mustOK(t, ct.RecordOutageStart("sta-dur", T0))
		mustOK(t, ct.RecordOutageEnd("sta-dur", T2)) // T2 - T0 = 120s
		summary, err := ct.GetOutageSummary("sta-dur", T3)
		mustOK(t, err)
		if summary.TotalOutages != 1 {
			t.Errorf("TotalOutages = %d, want 1", summary.TotalOutages)
		}
		if summary.OpenOutage {
			t.Error("OpenOutage should be false")
		}
		if summary.TotalOutageSecs != 120 {
			t.Errorf("TotalOutageSecs = %d, want 120", summary.TotalOutageSecs)
		}
	})

	t.Run("open_outage_counts_to_as_of", func(t *testing.T) {
		ct := newCT(t)
		mustRegisterStation(t, ct, "sta-open", "T", "north")
		mustOK(t, ct.RecordOutageStart("sta-open", T0))
		// T3 - T0 = 300s
		summary, err := ct.GetOutageSummary("sta-open", T3)
		mustOK(t, err)
		if !summary.OpenOutage {
			t.Error("OpenOutage should be true")
		}
		if summary.TotalOutageSecs != 300 {
			t.Errorf("TotalOutageSecs = %d, want 300", summary.TotalOutageSecs)
		}
	})

	t.Run("multiple_closed_outages_cumulative", func(t *testing.T) {
		ct := newCT(t)
		mustRegisterStation(t, ct, "sta-cumul", "T", "north")
		// Outage 1: T0 → T1 = 60s
		mustOK(t, ct.RecordOutageStart("sta-cumul", T0))
		mustOK(t, ct.RecordOutageEnd("sta-cumul", T1))
		// Outage 2: T2 → T3 = 180s
		mustOK(t, ct.RecordOutageStart("sta-cumul", T2))
		mustOK(t, ct.RecordOutageEnd("sta-cumul", T3))
		summary, err := ct.GetOutageSummary("sta-cumul", T4)
		mustOK(t, err)
		if summary.TotalOutages != 2 {
			t.Errorf("TotalOutages = %d, want 2", summary.TotalOutages)
		}
		if summary.OpenOutage {
			t.Error("OpenOutage should be false")
		}
		if summary.TotalOutageSecs != 240 { // 60 + 180
			t.Errorf("TotalOutageSecs = %d, want 240", summary.TotalOutageSecs)
		}
	})

	t.Run("missing_station_returns_error", func(t *testing.T) {
		ct := seededCT(t)
		_, err := ct.GetOutageSummary("ghost", T0)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
