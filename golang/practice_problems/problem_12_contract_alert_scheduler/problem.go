// =============================================================================
// INTERVIEW PROBLEM 12: Contract Expiration Alert Scheduler
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the alert-scheduling subsystem for a contract lifecycle
// management (CLM) platform used by legal and operations teams. Contracts have
// expiration dates, and stakeholders need to be notified days in advance so they
// can act before expiry.
//
// Store all state in struct fields set in your constructor.
// Package-level variables will bleed state between instances and test runs —
// avoid them. You choose the internal data structures; the public interface
// (ContractAlertScheduler) is what matters.
//
// DATA MODEL
// ----------
// Contract:
//   ID         string
//   Title      string
//   OwnerEmail string
//   ExpiresOn  string  // ISO-8601 date string, e.g. "2025-03-15"
//
// AlertConfig:
//   ID          string
//   DaysBefore  int    // how many days before expiry to trigger the alert
//   Label       string // e.g. "30-day notice", "final warning"
//
// AlertScheduleEntry (returned by ComputeAlertSchedule):
//   ConfigID string
//   Label    string
//   AlertOn  string  // ISO-8601 date string
//
// DueAlert (returned by GetDueAlerts):
//   ContractID string
//   ConfigID   string
//   Label      string
//   AlertOn    string
//   OwnerEmail string
//   ExpiresOn  string
//
// SentRecord (returned by RecordAlertSent):
//   ContractID string
//   ConfigID   string
//   SentOn     string  // ISO-8601 date string
//
// UpcomingAlert (returned by GetUpcomingAlerts):
//   ConfigID string
//   Label    string
//   AlertOn  string
//   Sent     bool
//
// DATE FORMAT
// -----------
// All dates are ISO-8601 date-only strings, e.g. "2025-03-15".
// Use time.Parse with layout "2006-01-02" for parsing and arithmetic.
//
// EXAMPLE
// -------
//   scheduler := NewContractAlertScheduler()
//   scheduler.AddContract("c-001", "Vendor MSA", "legal@acme.com", "2025-06-30")
//   scheduler.AddAlertConfig("cfg-30", 30, "30-day notice")
//   scheduler.AddAlertConfig("cfg-7",  7,  "final warning")
//
//   scheduler.GetContractsExpiringBetween("2025-06-01", "2025-06-30")
//   // -> [Contract{ID:"c-001", ...}]
//
//   scheduler.ComputeAlertSchedule("c-001")
//   // -> [{ConfigID:"cfg-30", Label:"30-day notice", AlertOn:"2025-05-31"},
//   //     {ConfigID:"cfg-7",  Label:"final warning",  AlertOn:"2025-06-23"}]
//
//   scheduler.GetDueAlerts("2025-06-01")
//   // -> [{ContractID:"c-001", ConfigID:"cfg-30", AlertOn:"2025-05-31",
//   //      OwnerEmail:"legal@acme.com", ExpiresOn:"2025-06-30", ...}]
// =============================================================================

package alertscheduler

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// Contract represents a business contract with an expiration date.
type Contract struct {
	ID         string
	Title      string
	OwnerEmail string
	ExpiresOn  string // ISO-8601 date string
}

// AlertConfig defines a global alert trigger: fire N days before expiry.
type AlertConfig struct {
	ID         string
	DaysBefore int
	Label      string
}

// AlertScheduleEntry is one entry in a contract's computed alert schedule.
type AlertScheduleEntry struct {
	ConfigID string
	Label    string
	AlertOn  string // ISO-8601 date string
}

// DueAlert is a scheduled alert that is on or before a given date.
type DueAlert struct {
	ContractID string
	ConfigID   string
	Label      string
	AlertOn    string
	OwnerEmail string
	ExpiresOn  string
}

// SentRecord records that an alert was dispatched for a contract/config pair.
type SentRecord struct {
	ContractID string
	ConfigID   string
	SentOn     string // ISO-8601 date string
}

// UpcomingAlert is an entry in get_upcoming_alerts — includes a Sent flag.
type UpcomingAlert struct {
	ConfigID string
	Label    string
	AlertOn  string
	Sent     bool
}

// ContractAlertScheduler is the interface candidates must implement.
//
// Implement it by defining your own struct and a NewContractAlertScheduler
// constructor. Store all state in struct fields — do NOT use package-level
// variables, as they bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myCAS struct {
//	    contracts    map[string]*Contract
//	    alertConfigs map[string]*AlertConfig
//	    sentRecords  []*SentRecord
//	    // ... add whatever fields you need
//	}
//
//	func NewContractAlertScheduler() ContractAlertScheduler {
//	    return &myCAS{
//	        contracts:    make(map[string]*Contract),
//	        alertConfigs: make(map[string]*AlertConfig),
//	    }
//	}
type ContractAlertScheduler interface {
	// -------------------------------------------------------------------------
	// PART 1 — Contract and alert-config management  (~10 min)
	// -------------------------------------------------------------------------

	// AddContract registers a contract and returns it.
	// Returns ErrAlreadyExists if contractID already exists.
	AddContract(contractID, title, ownerEmail, expiresOn string) (*Contract, error)

	// AddAlertConfig registers a global alert configuration and returns it.
	// Returns ErrAlreadyExists if configID already exists.
	AddAlertConfig(configID string, daysBefore int, label string) (*AlertConfig, error)

	// GetContractsExpiringBetween returns all contracts whose expiration date falls
	// within [startDate, endDate], inclusive on both ends.
	// Results are sorted by ExpiresOn ascending.
	GetContractsExpiringBetween(startDate, endDate string) ([]*Contract, error)

	// -------------------------------------------------------------------------
	// PART 2 — Alert schedule computation  (~15 min)
	// -------------------------------------------------------------------------

	// ComputeAlertSchedule computes the full alert schedule for a contract by
	// applying every registered AlertConfig.
	//
	// For each AlertConfig, the alert fires on:
	//     ExpiresOn - DaysBefore days
	//
	// Results are sorted by AlertOn ascending.
	// Returns ErrNotFound if contractID does not exist.
	ComputeAlertSchedule(contractID string) ([]*AlertScheduleEntry, error)

	// GetDueAlerts returns all alert schedule entries whose AlertOn date is on or
	// before asOfDate. Calls ComputeAlertSchedule per contract internally.
	//
	// Results are sorted by AlertOn ascending, then ContractID ascending.
	GetDueAlerts(asOfDate string) ([]*DueAlert, error)

	// -------------------------------------------------------------------------
	// PART 3 — Sent records and upcoming alerts  (~20 min)
	// -------------------------------------------------------------------------

	// RecordAlertSent records that an alert was sent for a contract/config pair
	// and returns the SentRecord.
	// Returns ErrNotFound if contractID or configID does not exist.
	RecordAlertSent(contractID, configID, sentOn string) (*SentRecord, error)

	// GetUpcomingAlerts returns the alert schedule for a contract, enriched with a
	// Sent flag indicating whether that alert has already been sent.
	// Uses ComputeAlertSchedule internally.
	//
	// Exclude alerts whose AlertOn is strictly before asOfDate (they are in the past).
	// Results are sorted by AlertOn ascending.
	// Returns ErrNotFound if contractID does not exist.
	GetUpcomingAlerts(contractID, asOfDate string) ([]*UpcomingAlert, error)
}

// NewContractAlertScheduler returns a fresh, empty ContractAlertScheduler
// implementation. Candidates implement their own struct type in the answer file.
func NewContractAlertScheduler() ContractAlertScheduler {
	panic("not implemented")
}
