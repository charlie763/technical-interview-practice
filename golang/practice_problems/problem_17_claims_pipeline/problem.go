// =============================================================================
// INTERVIEW PROBLEM 17: Claims Processing Pipeline
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the claims processing system for a management liability
// insurance platform. When a policyholder experiences a covered incident
// (e.g. an employment lawsuit, a D&O action), they file a claim. Claims move
// through a multi-stage review pipeline from filing through investigation,
// evaluation, and ultimately settlement or denial.
//
// You are implementing a ClaimsPipeline. All state must be stored in struct
// fields — do NOT use package-level variables, as they bleed state between
// instances and between test runs. You choose the internal data structures;
// the public interface (ClaimsPipeline) is what matters.
//
// DATA MODEL
// ----------
// Claim:
//   ClaimID        string
//   PolicyID       string
//   CoverageType   string   // "epl" | "do" | "fiduciary"
//   IncidentDate   string   // ISO-8601 date, e.g. "2025-03-15"
//   FiledAt        string   // ISO-8601 datetime
//   Status         string   // current status (see STATE MACHINE below)
//   ClaimedAmount  int      // dollars claimed by policyholder
//   ReserveAmount  int      // current reserve estimate
//   ApprovedAmount *int     // nil until settled
//   Events         []ClaimEvent
//
// ClaimEvent:
//   At      string // ISO-8601 datetime
//   Actor   string // adjuster ID or "system"
//   Action  string // "filed" | "status_change" | "reserve_update" | "settled" | "denied"
//   Payload map[string]interface{} // action-specific data
//
// STATE MACHINE
// -------------
// Valid transitions (see ValidTransitions package variable):
//   filed         → investigating
//   investigating → evaluation
//   evaluation    → settled | denied
//   settled       → closed
//   denied        → closed
//   closed        → (terminal — no further transitions)
//
// EXAMPLE
// -------
//   pipeline := NewClaimsPipeline()
//   pipeline.FileClaim(
//       "clm-001", "pol-101", "epl",
//       "2025-01-15", "2025-02-01T09:00:00",
//       75_000, 50_000, "adjuster-1",
//   )
//   pipeline.AdvanceStatus("clm-001", "investigating", "2025-02-03T10:00:00", "adjuster-1")
//   pipeline.GetClaim("clm-001") // status == "investigating"
//   pipeline.UpdateReserve("clm-001", 60_000, "2025-02-10T14:00:00", "adjuster-1")
//   pipeline.GetClaim("clm-001") // reserve_amount == 60_000
// =============================================================================

package claims

import "errors"

// ValidTransitions defines the allowed state-machine edges.
// A transition is only valid if the target status appears in the list for the
// current status. An empty list means the state is terminal.
var ValidTransitions = map[string][]string{
	"filed":         {"investigating"},
	"investigating": {"evaluation"},
	"evaluation":    {"settled", "denied"},
	"settled":       {"closed"},
	"denied":        {"closed"},
	"closed":        {},
}

// TerminalStates is the set of states from which no transitions are allowed.
var TerminalStates = map[string]bool{
	"closed": true,
}

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// ErrInvalidTransition is returned when the requested status transition is not
// allowed by the state machine.
var ErrInvalidTransition = errors.New("invalid transition")

// ErrInvalidAmount is returned when a monetary amount fails a validation check
// (e.g. must be > 0, or approved must not exceed claimed).
var ErrInvalidAmount = errors.New("invalid amount")

// ErrTerminalState is returned when an operation is attempted on a claim that
// has already reached a terminal state.
var ErrTerminalState = errors.New("claim is in terminal state")

// ErrWrongStatus is returned when a dedicated operation (e.g. SettleClaim)
// is called on a claim that is not in the required status.
var ErrWrongStatus = errors.New("claim is not in the required status")

// ClaimEvent records a single action taken on a claim.
type ClaimEvent struct {
	At      string
	Actor   string
	Action  string
	Payload map[string]interface{}
}

// Claim represents an insurance claim and its full event history.
type Claim struct {
	ClaimID        string
	PolicyID       string
	CoverageType   string
	IncidentDate   string
	FiledAt        string
	Status         string
	ClaimedAmount  int
	ReserveAmount  int
	ApprovedAmount *int // nil until settled
	Events         []ClaimEvent
}

// ReserveAdequacy is returned by GetReserveAdequacy.
type ReserveAdequacy struct {
	TotalReserves      int
	TotalApproved      int
	UnderReservedCount int
	UnderReservedGap   int
}

// ClaimsMetrics is returned by GetClaimsMetrics.
type ClaimsMetrics struct {
	Total               int
	ByStatus            map[string]int // only statuses with count > 0
	TotalClaimed        int
	TotalPaid           int
	AvgSettlementRatio  float64 // rounded to 4 decimal places; 0.0 if no settled claims
}

// PolicyLossHistory is returned by GetPolicyLossHistory.
type PolicyLossHistory struct {
	PolicyID     string
	ClaimCount   int
	TotalClaimed int
	TotalPaid    int
	LossRatio    float64 // rounded to 4 decimal places; 0.0 if no claims
}

// ClaimsPipeline is the interface all implementations must satisfy.
//
// Implement it by defining your own struct and a NewClaimsPipeline constructor.
// Store all state in struct fields — do NOT use package-level variables.
//
// Example skeleton:
//
//	type myPipeline struct {
//	    claims map[string]*Claim
//	    // ... add whatever fields you need
//	}
//
//	func NewClaimsPipeline() ClaimsPipeline {
//	    return &myPipeline{
//	        claims: make(map[string]*Claim),
//	    }
//	}
type ClaimsPipeline interface {
	// -------------------------------------------------------------------------
	// PART 1 — Claim filing, status transitions, and reserve updates
	// -------------------------------------------------------------------------

	// FileClaim registers a new claim in the "filed" state and appends an
	// initial ClaimEvent with Action="filed".
	//
	// Returns ErrAlreadyExists if claimID already exists.
	// Returns ErrInvalidAmount if claimedAmount <= 0 or reserveAmount <= 0.
	FileClaim(
		claimID, policyID, coverageType string,
		incidentDate, filedAt string,
		claimedAmount, reserveAmount int,
		actor string,
	) (*Claim, error)

	// GetClaim returns the claim for the given ID.
	// Returns ErrNotFound if claimID does not exist.
	GetClaim(claimID string) (*Claim, error)

	// AdvanceStatus moves a claim to a new status if the transition is valid,
	// and appends a ClaimEvent with Action="status_change" and
	// Payload={"from_status": ..., "to_status": ...}.
	//
	// Do NOT use this method to settle or deny — use SettleClaim and DenyClaim.
	//
	// Returns ErrNotFound if claimID does not exist.
	// Returns ErrInvalidTransition if toStatus is "settled" or "denied"
	//   (those require the dedicated methods) or if the transition is not
	//   allowed by ValidTransitions from the current status.
	AdvanceStatus(claimID, toStatus, at, actor string) (*Claim, error)

	// UpdateReserve updates the claim's ReserveAmount and appends a ClaimEvent
	// with Action="reserve_update" and
	// Payload={"old_reserve": ..., "new_reserve": ...}.
	//
	// Returns ErrNotFound if claimID does not exist.
	// Returns ErrInvalidAmount if newReserve <= 0.
	// Returns ErrTerminalState if the claim is in a terminal state ("closed").
	UpdateReserve(claimID string, newReserve int, at, actor string) (*Claim, error)

	// -------------------------------------------------------------------------
	// PART 2 — Settlement, denial, and query methods
	// -------------------------------------------------------------------------

	// SettleClaim transitions a claim from "evaluation" → "settled", sets
	// ApprovedAmount, and appends a ClaimEvent with Action="settled" and
	// Payload={"approved_amount": ...}.
	//
	// Returns ErrNotFound if claimID does not exist.
	// Returns ErrWrongStatus if the claim is not in "evaluation" status.
	// Returns ErrInvalidAmount if approvedAmount <= 0 or approvedAmount > ClaimedAmount.
	SettleClaim(claimID string, approvedAmount int, settledAt, actor string) (*Claim, error)

	// DenyClaim transitions a claim from "evaluation" → "denied" and appends
	// a ClaimEvent with Action="denied" and Payload={"reason": ...}.
	//
	// Returns ErrNotFound if claimID does not exist.
	// Returns ErrWrongStatus if the claim is not in "evaluation" status.
	DenyClaim(claimID, reason, deniedAt, actor string) (*Claim, error)

	// GetClaimsByPolicy returns all claims for the given policyID, sorted by
	// FiledAt ascending. Returns an empty slice if no claims exist for that policy.
	GetClaimsByPolicy(policyID string) ([]*Claim, error)

	// GetOpenClaims returns all claims that are NOT in a terminal state ("closed")
	// AND NOT denied. Sorted by FiledAt ascending.
	GetOpenClaims() ([]*Claim, error)

	// -------------------------------------------------------------------------
	// PART 3 — Reserve adequacy and metrics
	// -------------------------------------------------------------------------

	// GetReserveAdequacy analyses whether reserves cover settled amounts across
	// all claims.
	//
	// For settled claims, compare ReserveAmount with ApprovedAmount.
	// A claim is "under-reserved" if ReserveAmount < ApprovedAmount.
	//
	// Returns:
	//   TotalReserves:      sum of ReserveAmount for ALL claims
	//   TotalApproved:      sum of ApprovedAmount for settled claims only
	//   UnderReservedCount: count of settled claims where ReserveAmount < ApprovedAmount
	//   UnderReservedGap:   sum of (ApprovedAmount - ReserveAmount) for under-reserved claims
	GetReserveAdequacy() (*ReserveAdequacy, error)

	// GetClaimsMetrics returns aggregate statistics across all claims.
	//
	// AvgSettlementRatio = mean of (ApprovedAmount/ClaimedAmount) for all settled claims.
	// Returns 0.0 if no settled claims. Rounded to 4 decimal places.
	// ByStatus only includes statuses with count > 0.
	GetClaimsMetrics() (*ClaimsMetrics, error)

	// GetPolicyLossHistory summarises claim history for a specific policy.
	//
	// LossRatio = TotalPaid / TotalClaimed (0.0 if TotalClaimed == 0).
	// Rounded to 4 decimal places.
	// Returns a result with zero values if no claims exist for the policy.
	GetPolicyLossHistory(policyID string) (*PolicyLossHistory, error)
}

// NewClaimsPipeline returns a new ClaimsPipeline implementation.
//
// You must define your own struct type that implements the ClaimsPipeline
// interface and return it from this constructor.
//
// All state must be stored in struct fields — do NOT use package-level variables.
func NewClaimsPipeline() ClaimsPipeline {
	panic("not implemented")
}
