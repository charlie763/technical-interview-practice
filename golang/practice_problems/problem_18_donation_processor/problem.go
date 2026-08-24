// =============================================================================
// INTERVIEW PROBLEM 18: Donation Processor
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the core donation module for a K-12 school fundraising
// platform. Schools run campaigns — annual funds, giving days, capital
// campaigns — and donations flow in from families, alumni, and community
// supporters. Each campaign has a dollar goal, and you need to track progress,
// identify top donors, and handle edge cases like refunds and matching gifts.
//
// Store all state in struct fields initialised in your constructor.
// You choose the internal data structures; the public interface is what matters.
//
// DATA MODEL
// ----------
// Campaign: { ID, Name, Goal, StartDate }
// Donation:
//   {
//     ID:          string   — auto-generated ("don-1", "don-2", …)
//     CampaignID:  string
//     DonorName:   string
//     DonorEmail:  string
//     Amount:      float64  — dollars, must be > 0
//     Timestamp:   string   — ISO-8601 datetime
//     Refunded:    bool     — starts false
//   }
//
// EXAMPLE
// -------
//   dp := NewDonationProcessor()
//   dp.CreateCampaign("camp-1", "Annual Fund", 50_000, "2025-09-01")
//   dp.AddDonation("camp-1", "Alice Smith", "alice@school.org", 500, "2025-09-05T10:00:00")
//   dp.AddDonation("camp-1", "Bob Jones",   "bob@school.org",  1_000, "2025-09-06T11:00:00")
//   stats, _ := dp.GetCampaignStats("camp-1")
//   // stats.TotalRaised == 1500, stats.DonorCount == 2, stats.GoalPercent == 3.0
// =============================================================================

package donations

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// ErrInvalidAmount is returned when a donation amount is <= 0.
var ErrInvalidAmount = errors.New("amount must be greater than zero")

// ErrAlreadyRefunded is returned when trying to refund a donation that is
// already refunded.
var ErrAlreadyRefunded = errors.New("donation already refunded")

// ErrSourceRefunded is returned when trying to create a matching gift against
// a refunded source donation.
var ErrSourceRefunded = errors.New("source donation is refunded")

// Campaign represents a fundraising campaign.
type Campaign struct {
	ID        string
	Name      string
	Goal      float64 // dollars
	StartDate string  // ISO-8601 date
}

// Donation represents a single donation to a campaign.
type Donation struct {
	ID          string
	CampaignID  string
	DonorName   string
	DonorEmail  string
	Amount      float64 // dollars
	Timestamp   string  // ISO-8601 datetime
	Refunded    bool
}

// CampaignStats holds aggregate statistics for a campaign, excluding refunded
// donations.
type CampaignStats struct {
	TotalRaised   float64 // sum of non-refunded donation amounts
	DonorCount    int     // unique donor emails across non-refunded donations
	DonationCount int     // non-refunded donation count
	GoalPercent   float64 // (TotalRaised / Goal) * 100, rounded to 1 decimal place
}

// DonorSummary holds aggregated per-donor stats for a campaign.
type DonorSummary struct {
	DonorEmail    string
	DonorName     string  // name from the most recent donation by this email
	TotalAmount   float64 // sum of non-refunded donations from this email for the campaign
	DonationCount int     // non-refunded count
}

// DonationProcessor is the interface candidates must implement.
//
// Implement it by defining your own struct and a NewDonationProcessor constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myDP struct {
//	    campaigns  map[string]*Campaign
//	    donations  map[string]*Donation
//	    donCounter int
//	    // ... add whatever fields you need
//	}
//
//	func NewDonationProcessor() DonationProcessor {
//	    return &myDP{
//	        campaigns: make(map[string]*Campaign),
//	        donations: make(map[string]*Donation),
//	    }
//	}
type DonationProcessor interface {
	// -------------------------------------------------------------------------
	// PART 1 — Campaign setup, donation recording, and basic stats
	// -------------------------------------------------------------------------

	// CreateCampaign creates a new campaign and returns it.
	// Returns ErrAlreadyExists if id already exists.
	CreateCampaign(id, name string, goal float64, startDate string) (*Campaign, error)

	// AddDonation records a donation to an existing campaign.
	// Donation IDs are auto-generated ("don-1", "don-2", …) in insertion order.
	// Returns ErrNotFound if campaignID does not exist.
	// Returns ErrInvalidAmount if amount <= 0.
	AddDonation(campaignID, donorName, donorEmail string, amount float64, timestamp string) (*Donation, error)

	// GetDonation returns a donation by ID.
	// Returns ErrNotFound if donationID does not exist.
	GetDonation(donationID string) (*Donation, error)

	// GetCampaignStats returns aggregate stats for a campaign, excluding
	// refunded donations.
	// GoalPercent = math.Round((TotalRaised/Goal)*1000) / 10
	// Returns ErrNotFound if campaignID does not exist.
	GetCampaignStats(campaignID string) (*CampaignStats, error)

	// -------------------------------------------------------------------------
	// PART 2 — Top donors and date-range search
	// -------------------------------------------------------------------------

	// GetTopDonors returns the top `limit` donors for a campaign, ranked by
	// total non-refunded amount descending. Aggregates multiple donations from
	// the same email. On a tie in TotalAmount, sort by DonorEmail ascending.
	// Returns ErrNotFound if campaignID does not exist.
	GetTopDonors(campaignID string, limit int) ([]*DonorSummary, error)

	// GetDonationsInRange returns all non-refunded donations for a campaign
	// whose Timestamp falls within [from, to] (inclusive), sorted by Timestamp
	// ascending.
	// Returns ErrNotFound if campaignID does not exist.
	GetDonationsInRange(campaignID, from, to string) ([]*Donation, error)

	// -------------------------------------------------------------------------
	// PART 3 — Refunds and matching gifts
	// -------------------------------------------------------------------------

	// RefundDonation marks a donation as refunded.
	// Refunded donations are excluded from stats and leaderboards.
	// Returns ErrNotFound if donationID does not exist.
	// Returns ErrAlreadyRefunded if the donation is already refunded.
	RefundDonation(donationID string) (*Donation, error)

	// AddMatchingGift creates a matched donation in the same campaign as
	// sourceDonationID.
	//   matchedAmount = min(sourceDonation.Amount * matchRatio, maxAmount)
	// Matched donations appear in stats and leaderboards like any other
	// donation. The source donation must not be refunded.
	// Returns ErrNotFound if sourceDonationID does not exist.
	// Returns ErrSourceRefunded if the source donation is refunded.
	AddMatchingGift(sourceDonationID, matcherName, matcherEmail string, matchRatio, maxAmount float64) (*Donation, error)
}

// NewDonationProcessor returns a fresh, empty DonationProcessor implementation.
//
// You must define your own struct type that implements the DonationProcessor
// interface and return it from this constructor.
//
// All state must be stored in struct fields — do NOT use package-level variables.
func NewDonationProcessor() DonationProcessor {
	panic("not implemented")
}
