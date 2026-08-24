// =============================================================================
// INTERVIEW PROBLEM 19: Walkathon Pledge Tracker
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the pledge management module for a school walk-a-thon
// fundraiser. Before the event, student participants collect pledges from
// sponsors — either a flat donation or a per-lap amount. After the event, the
// school records how many laps each student completed and the platform
// calculates what each sponsor owes and produces class-level leaderboards.
//
// Store all state in struct fields initialised in your constructor.
// You choose the internal data structures; the public interface is what matters.
//
// DATA MODEL
// ----------
// Participant: { ID, Name, ClassID, LapsCompleted (starts 0) }
// Pledge:
//   {
//     ID:            string    — auto-generated ("pledge-1", "pledge-2", …)
//     ParticipantID: string
//     SponsorName:   string
//     SponsorEmail:  string
//     PledgeType:    PledgeType  ("per_lap" or "flat")
//     Amount:        float64
//   }
//
// PLEDGE CALCULATION
// ------------------
//   participant.TotalRaised = (sum of per_lap pledges × LapsCompleted)
//                           + (sum of flat pledges)
//
// EXAMPLE
// -------
//   pt := NewPledgeTracker()
//   pt.RegisterParticipant("p1", "Emma Torres", "class-5a")
//   pt.AddPledge("p1", "Grandma Rose", "rose@example.com", PledgeTypePerLap, 5)
//   pt.AddPledge("p1", "Uncle Joe",    "joe@example.com",  PledgeTypeFlat,   50)
//   pt.RecordLaps("p1", 10)
//   total, _ := pt.GetParticipantTotal("p1")
//   // total.TotalRaised == 100  (10 laps × $5 + $50 flat)
//   // total.PerLapTotal == 50
//   // total.FlatTotal   == 50
// =============================================================================

package pledges

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// ErrInvalidAmount is returned when a pledge amount is <= 0.
var ErrInvalidAmount = errors.New("amount must be greater than zero")

// ErrInvalidLaps is returned when lapsCompleted is negative.
var ErrInvalidLaps = errors.New("lapsCompleted must be >= 0")

// PledgeType represents the type of a pledge.
type PledgeType string

const (
	// PledgeTypePerLap indicates the sponsor pledges an amount per lap completed.
	PledgeTypePerLap PledgeType = "per_lap"
	// PledgeTypeFlat indicates the sponsor pledges a fixed amount regardless of laps.
	PledgeTypeFlat PledgeType = "flat"
)

// Participant represents a student participant in the walkathon.
type Participant struct {
	ID            string
	Name          string
	ClassID       string
	LapsCompleted int
}

// Pledge represents a single sponsor pledge for a participant.
type Pledge struct {
	ID            string
	ParticipantID string
	SponsorName   string
	SponsorEmail  string
	PledgeType    PledgeType
	Amount        float64
}

// ParticipantTotal holds the computed fundraising totals for a participant.
type ParticipantTotal struct {
	ParticipantID string
	Name          string
	ClassID       string
	LapsCompleted int
	PerLapTotal   float64 // sum of per_lap pledges × LapsCompleted
	FlatTotal     float64 // sum of flat pledges
	TotalRaised   float64 // PerLapTotal + FlatTotal
}

// ClassEntry holds the leaderboard entry for a class.
type ClassEntry struct {
	ClassID          string
	TotalRaised      float64
	ParticipantCount int
	TopFundraiser    TopFundraiserSummary
}

// TopFundraiserSummary identifies the top-raising participant within a class.
type TopFundraiserSummary struct {
	ParticipantID string
	Name          string
	TotalRaised   float64
}

// ParticipantRank is a ParticipantTotal with an added rank field.
type ParticipantRank struct {
	ParticipantTotal
	Rank int // 1-based position
}

// PledgeTracker is the interface candidates must implement.
//
// Implement it by defining your own struct and a NewPledgeTracker constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myPT struct {
//	    participants  map[string]*Participant
//	    pledges       []*Pledge
//	    pledgeCounter int
//	    // ... add whatever fields you need
//	}
//
//	func NewPledgeTracker() PledgeTracker {
//	    return &myPT{
//	        participants: make(map[string]*Participant),
//	    }
//	}
type PledgeTracker interface {
	// -------------------------------------------------------------------------
	// PART 1 — Participant registration and pledge collection
	// -------------------------------------------------------------------------

	// RegisterParticipant registers a new participant. LapsCompleted starts at 0.
	// Returns ErrAlreadyExists if id already exists.
	RegisterParticipant(id, name, classID string) (*Participant, error)

	// AddPledge adds a pledge for a participant.
	// Pledge IDs are auto-generated ("pledge-1", "pledge-2", …) in insertion order.
	// Returns ErrNotFound if participantID does not exist.
	// Returns ErrInvalidAmount if amount <= 0.
	AddPledge(participantID, sponsorName, sponsorEmail string, pledgeType PledgeType, amount float64) (*Pledge, error)

	// GetParticipant returns the participant object for the given ID.
	// Returns ErrNotFound if participantID does not exist.
	GetParticipant(participantID string) (*Participant, error)

	// -------------------------------------------------------------------------
	// PART 2 — Lap recording and totals
	// -------------------------------------------------------------------------

	// RecordLaps sets the number of laps a participant completed (not an increment).
	// Returns ErrNotFound if participantID does not exist.
	// Returns ErrInvalidLaps if lapsCompleted < 0.
	RecordLaps(participantID string, lapsCompleted int) (*Participant, error)

	// GetParticipantTotal calculates the total raised for a single participant.
	// Returns ErrNotFound if participantID does not exist.
	GetParticipantTotal(participantID string) (*ParticipantTotal, error)

	// GetCampaignTotal returns the sum of TotalRaised across ALL participants.
	GetCampaignTotal() float64

	// -------------------------------------------------------------------------
	// PART 3 — Class leaderboard and participant ranking
	// -------------------------------------------------------------------------

	// GetClassLeaderboard returns a leaderboard of classes sorted by TotalRaised
	// descending. On a tie, sort by ClassID ascending.
	// Each entry includes ClassID, TotalRaised, ParticipantCount, and TopFundraiser
	// (the participant with the highest TotalRaised; ties broken by ParticipantID
	// ascending).
	GetClassLeaderboard() ([]*ClassEntry, error)

	// GetParticipantLeaderboard returns participants ranked by TotalRaised
	// descending. Ties broken by ParticipantID ascending.
	// Returns all participants when limit <= 0 or limit > total participants.
	// Each entry includes all ParticipantTotal fields plus a 1-based Rank.
	GetParticipantLeaderboard(limit int) ([]*ParticipantRank, error)
}

// NewPledgeTracker returns a fresh, empty PledgeTracker implementation.
//
// You must define your own struct type that implements the PledgeTracker
// interface and return it from this constructor.
//
// All state must be stored in struct fields — do NOT use package-level variables.
func NewPledgeTracker() PledgeTracker {
	panic("not implemented")
}
