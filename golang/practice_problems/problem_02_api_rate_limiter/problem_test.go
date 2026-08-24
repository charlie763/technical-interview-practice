// Tests for Problem 2: Tiered API Rate Limiter
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_02_api_rate_limiter/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_02_api_rate_limiter.go \
//	  -c go test -v .
package ratelimiter

import (
	"errors"
	"testing"
)

const baseTime = 1_700_000_000.0

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

func smallPlans() map[string]*Plan {
	return map[string]*Plan{
		"free":      {RPM: IntPtr(3), RPD: IntPtr(10)},
		"pro":       {RPM: IntPtr(100), RPD: IntPtr(5_000)},
		"unlimited": {RPM: nil, RPD: nil},
	}
}

func newGW(t *testing.T) *Gateway {
	t.Helper()
	return MakeGateway(smallPlans())
}

func gwWithKey(t *testing.T) *Gateway {
	t.Helper()
	gw := newGW(t)
	if _, err := CreateKey(gw, "key_abc", "alice", "pro"); err != nil {
		t.Fatalf("CreateKey: %v", err)
	}
	return gw
}

// ---------------------------------------------------------------------------
// PART 1 — Key management
// ---------------------------------------------------------------------------

func TestCreateKey(t *testing.T) {
	t.Run("creates_key", func(t *testing.T) {
		gw := newGW(t)
		k, err := CreateKey(gw, "k1", "alice", "free")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gw.Keys["k1"] != k {
			t.Error("key not stored in gateway")
		}
		if k.Owner != "alice" {
			t.Errorf("Owner = %q, want 'alice'", k.Owner)
		}
		if k.Plan != "free" {
			t.Errorf("Plan = %q, want 'free'", k.Plan)
		}
		if !k.Enabled {
			t.Error("Enabled should be true")
		}
		if len(k.RequestLog) != 0 {
			t.Errorf("RequestLog should be empty, got %v", k.RequestLog)
		}
	})

	t.Run("duplicate_key_returns_error", func(t *testing.T) {
		gw := newGW(t)
		CreateKey(gw, "k1", "alice", "free")
		_, err := CreateKey(gw, "k1", "bob", "pro")
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("invalid_plan_returns_error", func(t *testing.T) {
		gw := newGW(t)
		_, err := CreateKey(gw, "k1", "alice", "enterprise")
		if !errors.Is(err, ErrInvalidPlan) {
			t.Errorf("expected ErrInvalidPlan, got %v", err)
		}
	})
}

func TestRevokeKey(t *testing.T) {
	t.Run("disables_key", func(t *testing.T) {
		gw := gwWithKey(t)
		if err := RevokeKey(gw, "key_abc"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gw.Keys["key_abc"].Enabled {
			t.Error("key should be disabled")
		}
	})

	t.Run("missing_key_returns_error", func(t *testing.T) {
		gw := newGW(t)
		if err := RevokeKey(gw, "ghost"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestUpdatePlan(t *testing.T) {
	t.Run("changes_plan", func(t *testing.T) {
		gw := gwWithKey(t)
		if err := UpdatePlan(gw, "key_abc", "free"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gw.Keys["key_abc"].Plan != "free" {
			t.Errorf("Plan = %q, want 'free'", gw.Keys["key_abc"].Plan)
		}
	})

	t.Run("preserves_request_log", func(t *testing.T) {
		gw := gwWithKey(t)
		gw.Keys["key_abc"].RequestLog = []float64{baseTime - 5}
		UpdatePlan(gw, "key_abc", "free")
		if len(gw.Keys["key_abc"].RequestLog) != 1 || gw.Keys["key_abc"].RequestLog[0] != baseTime-5 {
			t.Error("RequestLog was modified during plan change")
		}
	})

	t.Run("invalid_plan_returns_error", func(t *testing.T) {
		gw := gwWithKey(t)
		if err := UpdatePlan(gw, "key_abc", "nonexistent"); !errors.Is(err, ErrInvalidPlan) {
			t.Errorf("expected ErrInvalidPlan, got %v", err)
		}
	})

	t.Run("missing_key_returns_error", func(t *testing.T) {
		gw := newGW(t)
		if err := UpdatePlan(gw, "ghost", "free"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — countInWindow
// ---------------------------------------------------------------------------

func TestCountInWindow(t *testing.T) {
	tests := []struct {
		name    string
		log     []float64
		window  float64
		want    int
	}{
		{"empty_log", []float64{}, 60, 0},
		{"all_within_window", []float64{baseTime - 30, baseTime - 10, baseTime}, 60, 3},
		{"some_outside_window", []float64{baseTime - 120, baseTime - 61, baseTime - 30, baseTime}, 60, 2},
		{"exactly_on_left_edge_excluded", []float64{baseTime - 60}, 60, 0},
		{"one_second_inside", []float64{baseTime - 59.999}, 60, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := countInWindow(tc.log, baseTime, tc.window)
			if got != tc.want {
				t.Errorf("countInWindow = %d, want %d", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PART 2 — IsAllowed
// ---------------------------------------------------------------------------

func TestIsAllowed(t *testing.T) {
	t.Run("allowed_when_under_limits", func(t *testing.T) {
		gw := gwWithKey(t)
		if !IsAllowed(gw, "key_abc", baseTime) {
			t.Error("expected allowed")
		}
	})

	t.Run("denied_when_key_not_found", func(t *testing.T) {
		gw := newGW(t)
		if IsAllowed(gw, "ghost", baseTime) {
			t.Error("expected denied")
		}
	})

	t.Run("denied_when_key_disabled", func(t *testing.T) {
		gw := gwWithKey(t)
		RevokeKey(gw, "key_abc")
		if IsAllowed(gw, "key_abc", baseTime) {
			t.Error("expected denied")
		}
	})

	t.Run("denied_when_rpm_exceeded", func(t *testing.T) {
		gw := newGW(t)
		CreateKey(gw, "k1", "alice", "free") // rpm=3
		gw.Keys["k1"].RequestLog = []float64{baseTime - 30, baseTime - 20, baseTime - 10}
		if IsAllowed(gw, "k1", baseTime) {
			t.Error("expected denied (rpm exceeded)")
		}
	})

	t.Run("allowed_when_rpm_window_rolled_off", func(t *testing.T) {
		gw := newGW(t)
		CreateKey(gw, "k1", "alice", "free") // rpm=3
		gw.Keys["k1"].RequestLog = []float64{baseTime - 90, baseTime - 80, baseTime - 70}
		if !IsAllowed(gw, "k1", baseTime) {
			t.Error("expected allowed (old requests outside window)")
		}
	})

	t.Run("denied_when_rpd_exceeded", func(t *testing.T) {
		gw := newGW(t)
		CreateKey(gw, "k1", "alice", "free") // rpd=10
		log := make([]float64, 10)
		for i := range log {
			log[i] = baseTime - float64(i*100)
		}
		gw.Keys["k1"].RequestLog = log
		if IsAllowed(gw, "k1", baseTime) {
			t.Error("expected denied (rpd exceeded)")
		}
	})

	t.Run("unlimited_plan_always_allowed", func(t *testing.T) {
		gw := newGW(t)
		CreateKey(gw, "k1", "alice", "unlimited")
		log := make([]float64, 1000)
		for i := range log {
			log[i] = baseTime - float64(i)
		}
		gw.Keys["k1"].RequestLog = log
		if !IsAllowed(gw, "k1", baseTime) {
			t.Error("unlimited plan should always be allowed")
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — RecordRequest
// ---------------------------------------------------------------------------

func TestRecordRequest(t *testing.T) {
	t.Run("appends_timestamp", func(t *testing.T) {
		gw := gwWithKey(t)
		if err := RecordRequest(gw, "key_abc", baseTime); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		found := false
		for _, ts := range gw.Keys["key_abc"].RequestLog {
			if ts == baseTime {
				found = true
				break
			}
		}
		if !found {
			t.Error("timestamp not appended to RequestLog")
		}
	})

	t.Run("prunes_old_entries", func(t *testing.T) {
		gw := gwWithKey(t)
		old := baseTime - 90_001 // older than 25h
		gw.Keys["key_abc"].RequestLog = []float64{old}
		RecordRequest(gw, "key_abc", baseTime)
		for _, ts := range gw.Keys["key_abc"].RequestLog {
			if ts == old {
				t.Error("old entry should have been pruned")
			}
		}
	})

	t.Run("keeps_recent_entries", func(t *testing.T) {
		gw := gwWithKey(t)
		recent := baseTime - 3600
		gw.Keys["key_abc"].RequestLog = []float64{recent}
		RecordRequest(gw, "key_abc", baseTime)
		found := false
		for _, ts := range gw.Keys["key_abc"].RequestLog {
			if ts == recent {
				found = true
				break
			}
		}
		if !found {
			t.Error("recent entry should be preserved")
		}
	})

	t.Run("missing_key_returns_error", func(t *testing.T) {
		gw := newGW(t)
		if err := RecordRequest(gw, "ghost", baseTime); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 4 — HandleRequest
// ---------------------------------------------------------------------------

func TestHandleRequest(t *testing.T) {
	t.Run("success_records_request", func(t *testing.T) {
		gw := gwWithKey(t)
		result := HandleRequest(gw, "key_abc", baseTime)
		if !result.Allowed {
			t.Errorf("expected Allowed=true, got reason=%q", result.Reason)
		}
		found := false
		for _, ts := range gw.Keys["key_abc"].RequestLog {
			if ts == baseTime {
				found = true
				break
			}
		}
		if !found {
			t.Error("request not recorded on success")
		}
	})

	t.Run("key_not_found", func(t *testing.T) {
		gw := newGW(t)
		result := HandleRequest(gw, "ghost", baseTime)
		if result.Allowed || result.Reason != "key_not_found" {
			t.Errorf("got Allowed=%v Reason=%q, want Allowed=false Reason='key_not_found'", result.Allowed, result.Reason)
		}
	})

	t.Run("key_disabled", func(t *testing.T) {
		gw := gwWithKey(t)
		RevokeKey(gw, "key_abc")
		result := HandleRequest(gw, "key_abc", baseTime)
		if result.Allowed || result.Reason != "key_disabled" {
			t.Errorf("got Allowed=%v Reason=%q, want Allowed=false Reason='key_disabled'", result.Allowed, result.Reason)
		}
	})

	t.Run("rpm_exceeded", func(t *testing.T) {
		gw := newGW(t)
		CreateKey(gw, "k1", "alice", "free") // rpm=3
		gw.Keys["k1"].RequestLog = []float64{baseTime - 10, baseTime - 5, baseTime - 1}
		result := HandleRequest(gw, "k1", baseTime)
		if result.Allowed || result.Reason != "rpm_exceeded" {
			t.Errorf("got Allowed=%v Reason=%q, want Allowed=false Reason='rpm_exceeded'", result.Allowed, result.Reason)
		}
	})

	t.Run("rpd_exceeded_not_recorded", func(t *testing.T) {
		gw := newGW(t)
		CreateKey(gw, "k1", "alice", "free") // rpd=10
		log := make([]float64, 10)
		for i := range log {
			log[i] = baseTime - float64(i*100)
		}
		gw.Keys["k1"].RequestLog = log
		logBefore := make([]float64, len(log))
		copy(logBefore, log)
		result := HandleRequest(gw, "k1", baseTime)
		if result.Allowed || result.Reason != "rpd_exceeded" {
			t.Errorf("got Allowed=%v Reason=%q, want Allowed=false Reason='rpd_exceeded'", result.Allowed, result.Reason)
		}
		// log must not be modified on failure
		if len(gw.Keys["k1"].RequestLog) != len(logBefore) {
			t.Error("RequestLog was modified on failure")
		}
	})

	t.Run("rpd_checked_after_rpm", func(t *testing.T) {
		gw := newGW(t)
		CreateKey(gw, "k1", "alice", "free") // rpm=3, rpd=10
		// rpm is OK (0 in last minute), rpd is exceeded
		log := make([]float64, 10)
		for i := range log {
			log[i] = baseTime - float64(3600*(i+1))
		}
		gw.Keys["k1"].RequestLog = log
		result := HandleRequest(gw, "k1", baseTime)
		if result.Reason != "rpd_exceeded" {
			t.Errorf("expected Reason='rpd_exceeded', got %q", result.Reason)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 4 — GetUsage
// ---------------------------------------------------------------------------

func TestGetUsage(t *testing.T) {
	t.Run("returns_correct_counts", func(t *testing.T) {
		gw := gwWithKey(t)
		gw.Keys["key_abc"].RequestLog = []float64{
			baseTime - 30,   // within both minute and day windows
			baseTime - 3600, // within day window only
		}
		stats, err := GetUsage(gw, "key_abc", baseTime)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats.RPMUsed != 1 {
			t.Errorf("RPMUsed = %d, want 1", stats.RPMUsed)
		}
		if stats.RPDUsed != 2 {
			t.Errorf("RPDUsed = %d, want 2", stats.RPDUsed)
		}
		if stats.Plan != "pro" {
			t.Errorf("Plan = %q, want 'pro'", stats.Plan)
		}
		if stats.RPMLimit == nil || *stats.RPMLimit != 100 {
			t.Errorf("RPMLimit = %v, want 100", stats.RPMLimit)
		}
		if stats.RPDLimit == nil || *stats.RPDLimit != 5_000 {
			t.Errorf("RPDLimit = %v, want 5000", stats.RPDLimit)
		}
	})

	t.Run("unlimited_plan_shows_nil_limits", func(t *testing.T) {
		gw := newGW(t)
		CreateKey(gw, "k1", "alice", "unlimited")
		stats, err := GetUsage(gw, "k1", baseTime)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats.RPMLimit != nil {
			t.Errorf("RPMLimit = %v, want nil", stats.RPMLimit)
		}
		if stats.RPDLimit != nil {
			t.Errorf("RPDLimit = %v, want nil", stats.RPDLimit)
		}
	})

	t.Run("missing_key_returns_error", func(t *testing.T) {
		gw := newGW(t)
		_, err := GetUsage(gw, "ghost", baseTime)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
