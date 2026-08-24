// =============================================================================
// INTERVIEW PROBLEM 21: Platform Fee Calculator
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the fee calculation module for a school fundraising platform.
// The platform charges a fee on every transaction, and payment processors add
// their own fees on top. A key donor experience feature is "donor covers fee":
// the donor can opt to pay a higher gross amount so the school receives the
// full intended net donation.
//
// Store all state in struct fields initialised in the constructor.
// You choose the internal data structures; the public interface is what matters.
//
// FEE SCHEDULE (all fees calculated on the gross amount)
// -------------------------------------------------------
//   Platform fee:  3.5% of gross
//   Processing fee by method:
//     "card":  2.9% of gross + $0.30
//     "ach":   0.8% of gross, capped at $5.00
//     "check": $0.00
//
//   schoolReceives = gross − platformFee − processingFee
//
// Round all dollar values to the nearest cent: math.Round(x * 100) / 100
//
// EXAMPLE
// -------
//   calc := NewFeeCalculator()
//   f, _ := calc.CalculateFees(100, "card")
//   // f.Gross == 100, f.PlatformFee == 3.50, f.ProcessingFee == 3.20
//   // f.TotalFees == 6.70, f.SchoolReceives == 93.30
//
//   f2, _ := calc.GrossForNet(100, "card")
//   // f2.SchoolReceives ≈ 100 (±$0.02 rounding tolerance)
//   // f2.Gross is the amount the donor pays to cover all fees
//
// =============================================================================

package fees

import "errors"

// ErrInvalidAmount is returned when a gross or netTarget amount is <= 0.
var ErrInvalidAmount = errors.New("amount must be greater than zero")

// ErrUnknownCampaign is returned when getCampaignFeeSummary is called for a
// campaign that has no recorded transactions.
var ErrUnknownCampaign = errors.New("campaign not found")

// Fee rate constants — all fees are calculated on the gross amount.
const (
	PlatformRate  = 0.035 // 3.5% platform fee
	CardPctRate   = 0.029 // 2.9% card processing fee
	CardFixedFee  = 0.30  // $0.30 fixed card processing fee
	ACHPctRate    = 0.008 // 0.8% ACH processing fee
	ACHMaxFee     = 5.00  // $5.00 cap on ACH processing fee
)

// PaymentMethod is the type for accepted payment methods.
type PaymentMethod string

const (
	PaymentMethodCard  PaymentMethod = "card"
	PaymentMethodACH   PaymentMethod = "ach"
	PaymentMethodCheck PaymentMethod = "check"
)

// FeeBreakdown holds the result of a fee calculation.
type FeeBreakdown struct {
	Gross          float64 // dollar amount the donor pays
	PlatformFee    float64 // 3.5% of gross
	ProcessingFee  float64 // depends on payment method
	TotalFees      float64 // PlatformFee + ProcessingFee
	SchoolReceives float64 // Gross − TotalFees
}

// TransactionRecord is stored for each recorded transaction.
type TransactionRecord struct {
	ID             string        // auto-generated ("tx-1", "tx-2", …)
	CampaignID     string
	Gross          float64
	SchoolReceives float64
	TotalFees      float64
	Method         PaymentMethod
	DonorCoversFee bool
}

// MethodSummary aggregates stats for a single payment method within a campaign.
type MethodSummary struct {
	Count               int
	TotalGross          float64
	TotalSchoolReceives float64
	TotalFees           float64
}

// CampaignFeeSummary holds aggregated statistics for a campaign.
type CampaignFeeSummary struct {
	CampaignID          string
	TransactionCount    int
	TotalGross          float64
	TotalSchoolReceives float64
	TotalFees           float64
	FeeCoverageRate     float64 // fraction where DonorCoversFee=true, 2 decimal places
	ByMethod            map[PaymentMethod]*MethodSummary
}

// FeeCalculator is the interface candidates must implement.
//
// Implement it by defining your own struct type that implements the interface
// and returning it from NewFeeCalculator.
//
// All state must be stored in struct fields — do NOT use package-level variables.
//
// Example skeleton:
//
//	type myCalc struct {
//	    transactions []*TransactionRecord
//	    txCounter    int
//	}
//
//	func NewFeeCalculator() FeeCalculator {
//	    return &myCalc{}
//	}
type FeeCalculator interface {
	// -------------------------------------------------------------------------
	// PART 1 — Fee calculation for a given gross amount
	// -------------------------------------------------------------------------

	// CalculateFees returns the fee breakdown for a given gross donation amount.
	// Returns ErrInvalidAmount if gross <= 0.
	// All dollar values are rounded to the nearest cent.
	CalculateFees(gross float64, method PaymentMethod) (*FeeBreakdown, error)

	// -------------------------------------------------------------------------
	// PART 2 — Gross-for-net ("donor covers fee")
	// -------------------------------------------------------------------------

	// GrossForNet calculates the gross amount a donor must pay so the school
	// receives at least netTarget (within ±$0.02 rounding tolerance).
	//
	// Use math.Ceil to the nearest cent when computing the gross so the school
	// always receives >= netTarget. Then derive the FeeBreakdown by calling
	// CalculateFees on that gross — do not duplicate the fee logic.
	//
	// Derived formulas (for reference):
	//   card:  gross = ceil((netTarget + CardFixedFee) / (1 − PlatformRate − CardPctRate), 2)
	//   ach (processingFee would be < ACHMaxFee):
	//          gross = ceil(netTarget / (1 − PlatformRate − ACHPctRate), 2)
	//   ach (processingFee would reach ACHMaxFee, i.e. gross × ACHPctRate >= ACHMaxFee):
	//          gross = ceil((netTarget + ACHMaxFee) / (1 − PlatformRate), 2)
	//   check: gross = ceil(netTarget / (1 − PlatformRate), 2)
	//
	// Returns ErrInvalidAmount if netTarget <= 0.
	GrossForNet(netTarget float64, method PaymentMethod) (*FeeBreakdown, error)

	// -------------------------------------------------------------------------
	// PART 3 — Transaction recording and campaign aggregation
	// -------------------------------------------------------------------------

	// RecordTransaction records a transaction for a campaign.
	// Uses CalculateFees(gross, method) to compute SchoolReceives and TotalFees.
	// Transaction IDs are auto-generated as "tx-1", "tx-2", … in order.
	// Returns ErrInvalidAmount if gross <= 0.
	RecordTransaction(campaignID string, gross float64, method PaymentMethod, donorCoversFee bool) (*TransactionRecord, error)

	// GetCampaignFeeSummary returns aggregated fee statistics for a campaign.
	// FeeCoverageRate = coveredCount / totalCount, rounded to 2 decimal places.
	// ByMethod only includes methods that have at least one transaction.
	// Returns ErrUnknownCampaign if campaignID has no recorded transactions.
	GetCampaignFeeSummary(campaignID string) (*CampaignFeeSummary, error)
}

// NewFeeCalculator returns a new FeeCalculator implementation.
//
// You must define your own struct type that implements the FeeCalculator
// interface and return it from this constructor.
//
// All state must be stored in struct fields — do NOT use package-level variables.
func NewFeeCalculator() FeeCalculator {
	panic("not implemented")
}
