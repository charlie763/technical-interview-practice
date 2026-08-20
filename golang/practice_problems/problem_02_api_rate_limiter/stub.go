// =============================================================================
// INTERVIEW PROBLEM 2: Tiered API Rate Limiter
// Difficulty: Senior Software Engineer | Estimated time: 40 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the rate-limiting layer for a developer-facing API platform
// (think: Stripe, Vercel, Twilio — any product where external developers call
// your API and pay for usage tiers).
//
// Each API key belongs to a plan with per-minute and per-day request caps.
// Rate limiting uses a sliding window: a request at time T is within the
// per-minute cap if fewer than RPM requests occurred in the window (T-60, T],
// and within the per-day cap if fewer than RPD requests occurred in (T-86400, T].
//
// All timestamps are Unix time (float64 seconds).
//
// DATA MODEL (see types.go)
// -------------------------
// Plan       — RPM *int, RPD *int  (nil = unlimited)
// APIKey     — ID, Owner, Plan string; Enabled bool; RequestLog []float64
// Gateway    — Plans map[string]*Plan, Keys map[string]*APIKey
// RequestResult — Allowed bool, KeyID string, Reason string
// UsageStats    — KeyID, Plan string; RPMUsed, RPMLimit *int; RPDUsed, RPDLimit *int
//
// DEFAULT PLANS (used by MakeGateway when plans is nil)
// ------------------------------------------------------
//   "free":       RPM=60,    RPD=1_000
//   "starter":    RPM=300,   RPD=25_000
//   "pro":        RPM=1_000, RPD=200_000
//   "enterprise": RPM=nil,   RPD=nil  (unlimited)
//
// EXAMPLE
// -------
//   gw := MakeGateway(nil)
//   CreateKey(gw, "key_abc", "alice", "pro")
//   result := HandleRequest(gw, "key_abc", 1700000000.0)
//   // result.Allowed == true
//   stats, _ := GetUsage(gw, "key_abc", 1700000000.0)
//   // stats.RPMUsed == 1
// =============================================================================

package ratelimiter

// DefaultPlans is used by MakeGateway when no custom plans are provided.
var DefaultPlans = map[string]*Plan{
	"free":       {RPM: IntPtr(60), RPD: IntPtr(1_000)},
	"starter":    {RPM: IntPtr(300), RPD: IntPtr(25_000)},
	"pro":        {RPM: IntPtr(1_000), RPD: IntPtr(200_000)},
	"enterprise": {RPM: nil, RPD: nil},
}

// MakeGateway returns a fresh Gateway state.
// If plans is nil, DefaultPlans is used.
func MakeGateway(plans map[string]*Plan) *Gateway {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 1 — Key management  (~10 min)
// ---------------------------------------------------------------------------

// CreateKey registers a new API key and returns it.
// Returns ErrAlreadyExists if keyID already exists.
// Returns ErrInvalidPlan if plan is not in gw.Plans.
func CreateKey(gw *Gateway, keyID, owner, plan string) (*APIKey, error) {
	panic("not implemented")
}

// RevokeKey sets key.Enabled = false.
// Keys are disabled rather than deleted so historical logs are preserved.
// Returns ErrNotFound if keyID is not found.
func RevokeKey(gw *Gateway, keyID string) error {
	panic("not implemented")
}

// UpdatePlan changes a key's plan.
// The RequestLog is preserved (no reset on plan change).
// Returns ErrNotFound if keyID is not found.
// Returns ErrInvalidPlan if newPlan is not in gw.Plans.
func UpdatePlan(gw *Gateway, keyID, newPlan string) error {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 2 — Sliding-window rate check  (~15 min)
// ---------------------------------------------------------------------------

// countInWindow returns the number of entries in requestLog that fall within
// the half-open window (now - windowSeconds, now].
// Helper — feel free to call this from IsAllowed and HandleRequest, or inline it.
func countInWindow(requestLog []float64, now, windowSeconds float64) int {
	panic("not implemented")
}

// IsAllowed returns true if the key is allowed to make a request at time now.
//
// A request is NOT allowed if any of the following:
//   - keyID doesn't exist in gw.Keys
//   - key.Enabled is false
//   - the per-minute sliding window count >= plan's RPM cap
//   - the per-day sliding window count >= plan's RPD cap
//
// A nil limit means that dimension is unlimited.
func IsAllowed(gw *Gateway, keyID string, now float64) bool {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 3 — Recording requests + pruning  (~5 min)
// ---------------------------------------------------------------------------

// RecordRequest appends now to key.RequestLog and prunes any entries older
// than 25 hours (90_000 seconds) to bound memory usage.
// Call this only AFTER confirming the request is allowed.
// Returns ErrNotFound if keyID is not found.
func RecordRequest(gw *Gateway, keyID string, now float64) error {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// PART 4 — Combined handler + usage stats  (~10 min)
// ---------------------------------------------------------------------------

// HandleRequest attempts to process a request.
//
// On success (request allowed):
//
//	&RequestResult{Allowed: true, KeyID: keyID}
//	Side effect: records the request via RecordRequest.
//
// On failure:
//
//	&RequestResult{Allowed: false, Reason: <one of the strings below>}
//	No side effects.
//
// Reason strings:
//
//	"key_not_found" — keyID not in gw.Keys
//	"key_disabled"  — key exists but Enabled == false
//	"rpm_exceeded"  — per-minute cap reached
//	"rpd_exceeded"  — per-day cap reached (only when rpm is within limit)
func HandleRequest(gw *Gateway, keyID string, now float64) *RequestResult {
	panic("not implemented")
}

// GetUsage returns current usage stats for a key.
// Returns ErrNotFound if keyID is not found.
func GetUsage(gw *Gateway, keyID string, now float64) (*UsageStats, error) {
	panic("not implemented")
}
