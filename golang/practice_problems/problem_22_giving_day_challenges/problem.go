// =============================================================================
// INTERVIEW PROBLEM 22: Giving Day Challenge Engine
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the challenge engine for a school giving-day fundraiser.
// Sponsors set up challenges: "if the school raises $X by deadline Y, we'll
// add a $Z bonus." As donations roll in, the engine checks whether any
// challenges have been completed. At the end of the day, advancement staff
// review an hourly breakdown of activity and which challenges fired when.
//
// Store all state in struct fields initialised in the constructor.
// You choose the internal data structures; the public interface is what matters.
//
// CHALLENGE SEMANTICS
// -------------------
// A challenge is COMPLETE when:
//   sum of all donations whose timestamp <= challenge.Deadline >= challenge.ThresholdAmount
//
// Assume donations are always recorded in chronological order.
// Once a challenge is marked complete it stays complete.
//
// DATA MODEL
// ----------
//
//   Challenge — stored per challenge:
//     ID, Name, ThresholdAmount, Deadline (ISO-8601 datetime),
//     BonusType (BonusType), BonusValue (float64),
//     Completed (bool), CompletedAt (*string, nil until completed)
//
//   DonationEntry — a recorded donation:
//     ID (auto-generated "don-1", "don-2", …), Amount (float64), Timestamp (string)
//
//   HourlyBucket (Part 3) — one bucket per calendar hour with at least one donation:
//     Hour (ISO-8601 start of the hour, e.g. "2025-10-01T09:00:00"),
//     DonationCount (int),
//     AmountRaised (float64 — total donated this hour),
//     CumulativeTotal (float64 — running sum up to and including this hour),
//     ChallengesCompleted ([]string — IDs of challenges that first completed
//                          due to a donation in this hour)
//
// EXAMPLE
// -------
//   engine := NewGivingDayEngine()
//   engine.AddChallenge("c1", "Morning Boost", 1000, "2025-10-01T12:00:00", BonusTypeFlat, 500)
//   engine.ProcessDonation(400, "2025-10-01T09:00:00")  // cum=400 < 1000, no completion
//   engine.ProcessDonation(300, "2025-10-01T10:00:00")  // cum=700 < 1000, no completion
//   engine.ProcessDonation(400, "2025-10-01T11:00:00")  // cum=1100 >= 1000, c1 completes!
//   c, _ := engine.GetChallenge("c1")
//   // c.Completed == true, c.CompletedAt == "2025-10-01T11:00:00"
//
// =============================================================================

package challenges

import "errors"

// ErrAlreadyExists is returned when adding a challenge with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a challenge or donation that does not exist.
var ErrNotFound = errors.New("not found")

// ErrInvalidAmount is returned when a donation or threshold amount is <= 0.
var ErrInvalidAmount = errors.New("amount must be greater than zero")

// BonusType describes the kind of bonus a challenge awards upon completion.
type BonusType string

const (
	BonusTypeFlat     BonusType = "flat_bonus" // sponsor contributes a fixed dollar amount
	BonusTypeMatchPct BonusType = "match_pct"  // sponsor matches a percentage of total raised
)

// Challenge holds the configuration and completion state for a single challenge.
type Challenge struct {
	ID              string
	Name            string
	ThresholdAmount float64
	Deadline        string    // ISO-8601 datetime, e.g. "2025-10-01T12:00:00"
	BonusType       BonusType
	BonusValue      float64
	Completed       bool
	CompletedAt     *string // nil until the challenge completes
}

// DonationEntry records a single donation.
type DonationEntry struct {
	ID        string  // auto-generated ("don-1", "don-2", …)
	Amount    float64
	Timestamp string // ISO-8601 datetime
}

// ProcessDonationResult is returned by ProcessDonation.
type ProcessDonationResult struct {
	Donation       *DonationEntry
	NewlyCompleted []*Challenge
}

// HourlyBucket summarises donation activity within a single calendar hour.
type HourlyBucket struct {
	Hour                string   // ISO-8601 start of the hour, e.g. "2025-10-01T09:00:00"
	DonationCount       int
	AmountRaised        float64  // total donated during this hour
	CumulativeTotal     float64  // running sum of all donations up to and including this hour
	ChallengesCompleted []string // IDs of challenges whose CompletedAt falls in this hour
}

// GivingDayEngine is the interface candidates must implement.
//
// Implement it by defining your own struct type that satisfies this interface
// and returning it from NewGivingDayEngine.
//
// All state must be stored in struct fields — do NOT use package-level variables.
//
// Example skeleton:
//
//	type myEngine struct {
//	    challenges   map[string]*Challenge
//	    donations    []*DonationEntry
//	    donationSeq  int
//	}
//
//	func NewGivingDayEngine() GivingDayEngine {
//	    return &myEngine{
//	        challenges: make(map[string]*Challenge),
//	    }
//	}
type GivingDayEngine interface {
	// -------------------------------------------------------------------------
	// PART 1 — Challenge setup and donation recording
	// -------------------------------------------------------------------------

	// AddChallenge registers a challenge and returns the stored Challenge object
	// (Completed=false, CompletedAt=nil).
	// Returns ErrAlreadyExists if id already exists.
	// Returns ErrInvalidAmount if thresholdAmount <= 0.
	AddChallenge(id, name string, thresholdAmount float64, deadline string, bonusType BonusType, bonusValue float64) (*Challenge, error)

	// AddDonation records a donation without triggering challenge completion checks.
	// Donation IDs are auto-generated as "don-1", "don-2", … in order.
	// Returns ErrInvalidAmount if amount <= 0.
	// Use ProcessDonation to record a donation and check challenges simultaneously.
	AddDonation(amount float64, timestamp string) (*DonationEntry, error)

	// GetChallenge returns a challenge by ID.
	// Returns ErrNotFound if id does not exist.
	GetChallenge(id string) (*Challenge, error)

	// GetTotalRaised returns the sum of all donations whose timestamp <= asOf.
	// When asOf is empty (""), it returns the total of ALL donations.
	GetTotalRaised(asOf string) float64

	// -------------------------------------------------------------------------
	// PART 2 — Challenge completion via processDonation
	// -------------------------------------------------------------------------

	// ProcessDonation records a donation (via AddDonation) and checks all
	// incomplete challenges for completion.
	//
	// A challenge is "newly completed" if:
	//   - it was incomplete before this donation, AND
	//   - GetTotalRaised(asOf=challenge.Deadline) >= challenge.ThresholdAmount
	//
	// Newly completed challenges have Completed set to true and CompletedAt set
	// to the timestamp of this donation.
	//
	// Returns the stored donation and the slice of newly completed challenges
	// (empty, not nil, when none complete).
	// Returns ErrInvalidAmount if amount <= 0.
	ProcessDonation(amount float64, timestamp string) (*ProcessDonationResult, error)

	// -------------------------------------------------------------------------
	// PART 3 — Hourly breakdown
	// -------------------------------------------------------------------------

	// GetHourlyBreakdown returns one HourlyBucket per calendar hour that had at
	// least one donation, ordered by hour ascending.
	//
	// The Hour field is the ISO-8601 start of that hour, e.g. "2025-10-01T09:00:00".
	// ChallengesCompleted lists IDs of challenges whose CompletedAt starts with
	// the same "YYYY-MM-DDTHH" prefix as the bucket hour.
	//
	// Returns an empty slice when no donations have been recorded.
	GetHourlyBreakdown() []*HourlyBucket
}

// NewGivingDayEngine returns a fresh, empty GivingDayEngine implementation.
//
// You must define your own struct type that implements the GivingDayEngine
// interface and return it from this constructor.
//
// All state must be stored in struct fields — do NOT use package-level variables.
func NewGivingDayEngine() GivingDayEngine {
	panic("not implemented")
}
