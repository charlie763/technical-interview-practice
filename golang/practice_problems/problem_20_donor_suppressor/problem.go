// =============================================================================
// INTERVIEW PROBLEM 20: Donor Communication Suppressor
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the donor suppression engine for a school fundraising
// platform. Before each email campaign goes out, the platform filters the
// send list to avoid contacting donors who have already given, recently gave,
// or have opted out of communications. Getting this wrong wastes goodwill —
// double-soliciting recent donors is a common and costly mistake.
//
// Store all state in struct fields initialised in your constructor.
// You choose the internal data structures; the public interface is what matters.
//
// SUPPRESSION RULES (applied in priority order)
// -----------------------------------------------
// 1. opted_out       — donor explicitly unsubscribed
// 2. already_donated — donor has a donation recorded for this specific campaignID
// 3. recency_cooldown — donor donated to ANY campaign within the last cooldownDays
//                       days relative to asOf (i.e. donatedAt >= asOf minus cooldownDays)
//
// A donor is suppressed if ANY rule applies. Return the highest-priority reason.
//
// EXAMPLE
// -------
//   s := NewDonorSuppressor()
//   s.OptOut("alice@example.com")
//   s.RecordDonation("bob@example.com", "camp-fall", "2025-08-01T10:00:00")
//   r1, _ := s.CheckSuppression("alice@example.com", "camp-fall", "2025-08-20T00:00:00", SuppressOptions{})
//   // r1.Suppressed == true, r1.Reason == SuppressReasonOptedOut
//   r2, _ := s.CheckSuppression("bob@example.com", "camp-fall", "2025-08-20T00:00:00", SuppressOptions{})
//   // r2.Suppressed == true, r2.Reason == SuppressReasonAlreadyDonated
//   r3, _ := s.CheckSuppression("carol@example.com", "camp-fall", "2025-08-20T00:00:00", SuppressOptions{})
//   // r3.Suppressed == false, r3.Reason == ""
// =============================================================================

package suppressor

// SuppressReason is the reason a donor was suppressed.
type SuppressReason string

const (
	// SuppressReasonOptedOut indicates the donor explicitly unsubscribed.
	SuppressReasonOptedOut SuppressReason = "opted_out"
	// SuppressReasonAlreadyDonated indicates the donor has a donation for this
	// specific campaign.
	SuppressReasonAlreadyDonated SuppressReason = "already_donated"
	// SuppressReasonRecencyCooldown indicates the donor donated to any campaign
	// within the cooldown window.
	SuppressReasonRecencyCooldown SuppressReason = "recency_cooldown"
)

// SuppressResult holds the outcome of a suppression check for a single email.
type SuppressResult struct {
	Email      string
	Suppressed bool
	Reason     SuppressReason // empty string when Suppressed == false
}

// SuppressOptions configures optional suppression rules for a check.
type SuppressOptions struct {
	// CooldownDays suppresses donors who donated within the last N days relative
	// to asOf. Zero means the recency_cooldown rule is not applied.
	CooldownDays int
}

// BulkFilterResult is the output of a bulk campaign list filter.
type BulkFilterResult struct {
	Eligible   []string         // emails that passed all suppression rules
	Suppressed []*SuppressResult // full result for each suppressed email
}

// SuppressedSummary contains aggregate counts from a bulk suppression filter.
type SuppressedSummary struct {
	Total          int
	EligibleCount  int
	SuppressedCount int
	ByReason       map[SuppressReason]int // count per suppress reason
}

// DonorSuppressor is the interface candidates must implement.
//
// Implement it by defining your own struct and a NewDonorSuppressor constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myDS struct {
//	    optedOut  map[string]bool
//	    donations map[string][]donationRecord // email → records
//	    // ... add whatever fields you need
//	}
//
//	func NewDonorSuppressor() DonorSuppressor {
//	    return &myDS{
//	        optedOut:  make(map[string]bool),
//	        donations: make(map[string][]donationRecord),
//	    }
//	}
type DonorSuppressor interface {
	// -------------------------------------------------------------------------
	// PART 1 — Opt-out management and basic suppression check
	// -------------------------------------------------------------------------

	// OptOut adds an email to the global opt-out list. Idempotent.
	OptOut(email string)

	// OptIn removes an email from the opt-out list. Idempotent (no-op if not
	// opted out).
	OptIn(email string)

	// IsOptedOut returns true if the email is on the opt-out list.
	IsOptedOut(email string) bool

	// CheckSuppression checks whether a donor should be suppressed from a
	// campaign send.
	//
	// Part 1: only applies the opted_out rule.
	// Part 2: also applies already_donated and recency_cooldown (when
	//   options.CooldownDays > 0).
	//
	// Rules are applied in priority order: opted_out > already_donated >
	// recency_cooldown. Returns the highest-priority matching reason.
	//
	// asOf is an ISO-8601 datetime string representing "now" for date arithmetic.
	CheckSuppression(email, campaignID, asOf string, options SuppressOptions) *SuppressResult

	// -------------------------------------------------------------------------
	// PART 2 — Donation-based suppression rules
	// -------------------------------------------------------------------------

	// RecordDonation records that an email donated to a campaign at a given
	// datetime. Multiple calls for the same email/campaign pair are allowed.
	// The most recent donatedAt is used for recency calculations.
	// donatedAt is an ISO-8601 datetime string.
	RecordDonation(email, campaignID, donatedAt string)

	// -------------------------------------------------------------------------
	// PART 3 — Bulk campaign filtering and summary
	// -------------------------------------------------------------------------

	// FilterCampaignList filters a list of emails into eligible and suppressed
	// groups. Calls CheckSuppression for each email — do not duplicate its logic.
	// asOf is an ISO-8601 datetime string representing "now".
	FilterCampaignList(emails []string, campaignID, asOf string, options SuppressOptions) *BulkFilterResult

	// GetSuppressedSummary returns a summary of suppression reasons across a
	// list of emails. Calls FilterCampaignList internally — do not duplicate
	// its logic.
	GetSuppressedSummary(emails []string, campaignID, asOf string, options SuppressOptions) *SuppressedSummary
}

// NewDonorSuppressor returns a fresh, empty DonorSuppressor implementation.
//
// You must define your own struct type that implements the DonorSuppressor
// interface and return it from this constructor.
//
// All state must be stored in struct fields — do NOT use package-level variables.
func NewDonorSuppressor() DonorSuppressor {
	panic("not implemented")
}
