// =============================================================================
// INTERVIEW PROBLEM 13: Contract Lifecycle State Machine
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the lifecycle management module for a contract platform.
// Every contract moves through a defined set of states from creation to
// completion or termination. Only specific transitions are allowed — illegal
// transitions must be rejected. All changes are audit-logged.
//
// Store all state in struct fields initialized in NewContractLifecycle.
// Package-level variables will bleed state between instances and test runs —
// avoid them. You choose the internal data structures; the public interface
// (ContractLifecycle) is what matters.
//
// DATA MODEL
// ----------
// Contract:
//   ContractID  string
//   Title       string
//   State       string  // current lifecycle state (see STATE MACHINE below)
//   CreatedAt   string  // ISO-8601 datetime when the contract was created
//   Fields      map[string]interface{}  // arbitrary key/value metadata
//
// AuditEntry:
//   ContractID  string
//   FromState   *string  // nil for the initial "created" entry
//   ToState     string
//   At          string   // ISO-8601 datetime of the transition
//   Actor       string   // user or system identifier
//
// STATE MACHINE
// -------------
// Valid states and allowed forward transitions:
//
//   draft        → in_review
//   in_review    → approved | draft   (can be sent back for revisions)
//   approved     → executed
//   executed     → active
//   active       → expiring_soon | terminated
//   expiring_soon → expired | active | terminated
//   expired      → (terminal — no outgoing transitions)
//   terminated   → (terminal — no outgoing transitions)
//
// EXAMPLE
// -------
//   cl := NewContractLifecycle()
//   cl.CreateContract("c-001", "Vendor MSA", "2025-01-01T09:00:00", "alice")
//   // contract is in "draft" state
//   cl.Transition("c-001", "in_review", "2025-01-02T10:00:00", "alice")
//   cl.Transition("c-001", "approved",  "2025-01-03T11:00:00", "bob")
//   // cl.Transition("c-001", "draft", "2025-01-04T09:00:00", "bob")
//   // returns ErrInvalidTransition — approved → draft is not a valid transition
// =============================================================================

package lifecycle

import "errors"

// ValidTransitions maps each state to the list of states it may transition to.
// Terminal states (expired, terminated) map to empty slices.
var ValidTransitions = map[string][]string{
	"draft":         {"in_review"},
	"in_review":     {"approved", "draft"},
	"approved":      {"executed"},
	"executed":      {"active"},
	"active":        {"expiring_soon", "terminated"},
	"expiring_soon": {"expired", "active", "terminated"},
	"expired":       {},
	"terminated":    {},
}

// TerminalStates is the set of states from which no further transitions are allowed.
var TerminalStates = map[string]bool{
	"expired":    true,
	"terminated": true,
}

// ErrAlreadyExists is returned when creating a contract with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a contract that does not exist.
var ErrNotFound = errors.New("not found")

// ErrInvalidTransition is returned when a state transition is not allowed.
var ErrInvalidTransition = errors.New("invalid transition")

// Contract holds the current state of a contract.
type Contract struct {
	ContractID string
	Title      string
	State      string
	CreatedAt  string
	Fields     map[string]interface{}
}

// AuditEntry records a single state transition (or creation) event.
type AuditEntry struct {
	ContractID string
	FromState  *string // nil for the initial creation entry
	ToState    string
	At         string
	Actor      string
}

// BulkResult is returned by BulkAdvance.
type BulkResult struct {
	Succeeded []string      // contract IDs successfully transitioned
	Failed    []BulkFailure // details for each failure
}

// BulkFailure describes a single failed transition in a BulkAdvance call.
type BulkFailure struct {
	ContractID string
	Reason     string
}

// OverdueContract is returned by GetOverdueContracts.
type OverdueContract struct {
	ContractID string
	Title      string
	State      string
	StuckSince string // ISO-8601 datetime of last transition
	DaysStuck  int
}

// LifecycleMetrics is returned by GetLifecycleMetrics.
type LifecycleMetrics struct {
	Total         int
	ByState       map[string]int // only states with count > 0
	TerminalCount int
}

// ContractLifecycle is the interface candidates must implement.
//
// Implement it by defining your own struct and a NewContractLifecycle constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myCL struct {
//	    contracts map[string]*Contract
//	    audit     map[string][]*AuditEntry
//	}
//
//	func NewContractLifecycle() ContractLifecycle {
//	    return &myCL{
//	        contracts: make(map[string]*Contract),
//	        audit:     make(map[string][]*AuditEntry),
//	    }
//	}
type ContractLifecycle interface {
	// -------------------------------------------------------------------------
	// PART 1 — Contract creation, field management, and transitions
	// -------------------------------------------------------------------------

	// CreateContract creates a new contract in "draft" state and records the
	// initial audit entry (FromState=nil, ToState="draft").
	// Returns ErrAlreadyExists if contractID already exists.
	CreateContract(contractID, title, createdAt, actor string) (*Contract, error)

	// SetField sets or updates a field on the contract's Fields map.
	// Returns ErrNotFound if contractID does not exist.
	SetField(contractID, key string, value interface{}) (*Contract, error)

	// GetContract returns the contract.
	// Returns ErrNotFound if contractID does not exist.
	GetContract(contractID string) (*Contract, error)

	// Transition moves a contract to a new state if the transition is valid,
	// and appends an AuditEntry.
	// Returns ErrNotFound if contractID does not exist.
	// Returns ErrInvalidTransition if the transition from the current state to
	// toState is not allowed (consult ValidTransitions).
	Transition(contractID, toState, at, actor string) (*Contract, error)

	// -------------------------------------------------------------------------
	// PART 2 — Audit trail, by-state query, bulk advance
	// -------------------------------------------------------------------------

	// GetAuditTrail returns the full ordered audit trail for a contract (oldest first).
	// Returns ErrNotFound if contractID does not exist.
	GetAuditTrail(contractID string) ([]*AuditEntry, error)

	// GetContractsByState returns all contracts currently in the given state,
	// sorted by ContractID ascending.
	GetContractsByState(state string) []*Contract

	// BulkAdvance attempts to transition each contract in contractIDs to toState.
	// Calls Transition internally — does not duplicate its logic.
	// Continues processing remaining contracts even if one fails; collects all failures.
	BulkAdvance(contractIDs []string, toState, at, actor string) BulkResult

	// -------------------------------------------------------------------------
	// PART 3 — Lifecycle metrics and overdue contracts
	// -------------------------------------------------------------------------

	// GetLifecycleMetrics returns aggregate counts across all contracts.
	GetLifecycleMetrics() LifecycleMetrics

	// GetOverdueContracts returns contracts that have been stuck in the same
	// non-terminal state for more than 30 days without any transition.
	// "Stuck since" is the At timestamp of the most recent AuditEntry.
	// Uses GetAuditTrail internally — does not duplicate its logic.
	// Results are sorted by DaysStuck descending.
	GetOverdueContracts(asOf string) []OverdueContract
}

// NewContractLifecycle returns a fresh, empty ContractLifecycle implementation.
// Candidates implement their own struct type in the answer file.
func NewContractLifecycle() ContractLifecycle {
	panic("not implemented")
}
