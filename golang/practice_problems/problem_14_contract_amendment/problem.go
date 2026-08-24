// =============================================================================
// INTERVIEW PROBLEM 14: Contract Amendment Manager
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the amendment-tracking module for a contract platform.
// After a contract is signed, its terms can be modified through formal amendments.
// Each amendment specifies a set of field overrides that take effect from a given
// date. To know the effective terms on any given date, you apply the base contract
// fields and then overlay amendments in chronological order up to that date.
//
// Store all state in struct fields initialized in NewContractAmendmentManager.
// Package-level variables will bleed state between instances and test runs —
// avoid them. You choose the internal data structures; the public interface
// (ContractAmendmentManager) is what matters.
//
// DATA MODEL
// ----------
// Contract (base):
//   ContractID  string
//   Title       string
//   Fields      map[string]interface{}  // e.g. {"value": 50000, "payment_terms": "net-30"}
//
// Amendment:
//   AmendmentID  string
//   ContractID   string
//   EffectiveOn  string            // ISO-8601 date string ("2006-01-02")
//   Overrides    map[string]interface{}  // field key → new value (may be a subset of fields)
//   Note         string            // human-readable reason for the amendment
//
// Dates are ISO-8601 date strings (date-only, format "2006-01-02").
// Use string comparison (lexicographic) or time.Parse for date ordering.
//
// EXAMPLE
// -------
//   mgr := NewContractAmendmentManager()
//   mgr.AddContract("c-001", "Vendor MSA", map[string]interface{}{"value": 50000, "payment_terms": "net-30"})
//
//   mgr.AddAmendment("amd-1", "c-001", "2025-03-01", map[string]interface{}{"payment_terms": "net-45"}, "extended terms")
//   mgr.AddAmendment("amd-2", "c-001", "2025-06-01", map[string]interface{}{"value": 75000}, "scope increase")
//
//   mgr.GetEffectiveContract("c-001", "2025-01-01")
//   // -> {"value": 50000, "payment_terms": "net-30"}  (no amendments yet)
//
//   mgr.GetEffectiveContract("c-001", "2025-04-15")
//   // -> {"value": 50000, "payment_terms": "net-45"}  (amd-1 applied)
//
//   mgr.GetEffectiveContract("c-001", "2025-07-01")
//   // -> {"value": 75000, "payment_terms": "net-45"}  (amd-1 + amd-2 applied)
// =============================================================================

package amendments

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// BaseContract holds a contract's original terms.
type BaseContract struct {
	ContractID string
	Title      string
	Fields     map[string]interface{}
}

// Amendment records a set of field overrides that take effect from EffectiveOn.
type Amendment struct {
	AmendmentID string
	ContractID  string
	EffectiveOn string // ISO-8601 date "2006-01-02"
	Overrides   map[string]interface{}
	Note        string
}

// FieldHistoryEntry records the value of a single field at a point in time.
type FieldHistoryEntry struct {
	EffectiveOn string      // ISO-8601 date, or "base" for the original value
	Value       interface{}
	Source      string // "base" | amendment_id
}

// AmendmentSummary is returned by GetAmendmentSummary.
type AmendmentSummary struct {
	ContractID      string
	AmendmentCount  int
	FieldsAmended   []string // unique field names ever overridden, sorted ascending
	LatestAmendment *string  // ISO-8601 date of most recent amendment, or nil if none
	CurrentFields   map[string]interface{}
}

// ContractAmendmentManager is the interface candidates must implement.
//
// Implement it by defining your own struct and a NewContractAmendmentManager constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myCAM struct {
//	    contracts  map[string]*BaseContract
//	    amendments map[string][]*Amendment  // contractID → ordered list
//	    amendIDs   map[string]bool          // global amendment ID set
//	}
//
//	func NewContractAmendmentManager() ContractAmendmentManager {
//	    return &myCAM{
//	        contracts:  make(map[string]*BaseContract),
//	        amendments: make(map[string][]*Amendment),
//	        amendIDs:   make(map[string]bool),
//	    }
//	}
type ContractAmendmentManager interface {
	// -------------------------------------------------------------------------
	// PART 1 — Base contract management
	// -------------------------------------------------------------------------

	// AddContract registers a base contract.
	// Fields must be stored as a copy — do not hold a reference to the caller's map.
	// Returns ErrAlreadyExists if contractID already exists.
	AddContract(contractID, title string, fields map[string]interface{}) (*BaseContract, error)

	// GetBaseContract returns the base contract (original fields, no amendments applied).
	// Returns ErrNotFound if contractID does not exist.
	GetBaseContract(contractID string) (*BaseContract, error)

	// -------------------------------------------------------------------------
	// PART 2 — Amendments and effective contract
	// -------------------------------------------------------------------------

	// AddAmendment registers an amendment for a contract.
	// Overrides must be stored as a copy — do not hold a reference to the caller's map.
	// Returns ErrAlreadyExists if amendmentID already exists globally.
	// Returns ErrNotFound if contractID does not exist.
	AddAmendment(amendmentID, contractID, effectiveOn string, overrides map[string]interface{}, note string) (*Amendment, error)

	// GetAmendments returns all amendments for a contract, sorted by EffectiveOn
	// ascending, then AmendmentID ascending (for deterministic ordering when dates match).
	// Returns ErrNotFound if contractID does not exist.
	GetAmendments(contractID string) ([]*Amendment, error)

	// GetEffectiveContract returns the resolved field values for a contract as of asOfDate.
	// Starts with base fields from GetBaseContract, then applies amendments in chronological
	// order (earliest first) where EffectiveOn <= asOfDate, overlaying their overrides.
	// Uses GetBaseContract and GetAmendments internally — does not duplicate their logic.
	// Returns ErrNotFound if contractID does not exist.
	GetEffectiveContract(contractID, asOfDate string) (map[string]interface{}, error)

	// -------------------------------------------------------------------------
	// PART 3 — Value history and amendment summary
	// -------------------------------------------------------------------------

	// GetValueHistory returns the full history of a specific field's value across
	// the base and all amendments that touched it, in chronological order.
	// The base entry always comes first with Source="base".
	// Returns ErrNotFound if contractID does not exist, or if the field is not present
	// in the base contract or any amendment for that contract.
	GetValueHistory(contractID, field string) ([]FieldHistoryEntry, error)

	// GetAmendmentSummary returns a summary of amendment activity for a contract.
	// Uses GetAmendments and GetEffectiveContract internally — does not duplicate their logic.
	// For CurrentFields, use the most recent amendment's EffectiveOn date if amendments exist,
	// otherwise use "2099-12-31" as a far-future sentinel so all amendments are included.
	// Returns ErrNotFound if contractID does not exist.
	GetAmendmentSummary(contractID string) (*AmendmentSummary, error)
}

// NewContractAmendmentManager returns a fresh, empty ContractAmendmentManager implementation.
// Candidates implement their own struct type in the answer file.
func NewContractAmendmentManager() ContractAmendmentManager {
	panic("not implemented")
}
