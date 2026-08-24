// =============================================================================
// INTERVIEW PROBLEM 16: Policy Premium Rating Engine
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You're building the premium rating engine for a management and professional
// liability insurance platform. Brokers submit applications on behalf of their
// clients; underwriters use your engine to calculate base premiums, apply risk
// modifiers, and analyze portfolio exposure.
//
// You are implementing a PremiumRatingEngine. All state must be stored in
// struct fields — do NOT use package-level variables, as they bleed state
// between instances and between test runs. You choose the internal data
// structures; the public interface (PremiumRatingEngine) is what matters.
//
// DATA MODEL
// ----------
// PolicySubmission:
//   SubmissionID    string
//   CoverageType    string   // "epl" | "do" | "fiduciary"
//   CompanyName     string
//   EmployeeCount   int
//   AnnualRevenue   int      // dollars
//   YearsInBusiness int
//   IndustryRisk    string   // "low" | "medium" | "high"
//   RequestedLimit  int      // policy limit in dollars
//   Deductible      int      // self-insured retention in dollars
//   PriorClaims     []PriorClaim
//
// PriorClaim:
//   Year      int    // calendar year
//   Amount    int    // dollars paid/reserved
//   ClaimType string // e.g. "epl", "do", "fiduciary", "any"
//
// PREMIUM RATING RULES
// --------------------
// Base premium by coverage type (Part 1):
//   EPL:       $1,200 + ($15 × EmployeeCount) + (AnnualRevenue × 0.0008)
//   D&O:       $2,500 + (AnnualRevenue × 0.0010)
//   Fiduciary: $800   + (AnnualRevenue × 0.0004)
//   Cap:   base_premium may not exceed 3% of RequestedLimit
//   Floor: base_premium may not be less than $500
//
// Risk modifiers applied to base premium (Part 2):
//   IndustryRisk:
//     "low"    → × 0.85
//     "medium" → × 1.00
//     "high"   → × 1.35
//   YearsInBusiness:
//     < 3      → × 1.25
//     3–10     → × 1.00
//     > 10     → × 0.90
//   PriorClaims in the last 3 years (relative to currentYear argument):
//     Each qualifying claim adds ×1.15 (multiplicative modifier cap: ×1.60)
//     Qualifying = ClaimType matches CoverageType OR ClaimType is "any"
//     AND year >= currentYear - 2
//
//   Final premium = base × industryModifier × tenureModifier × claimsModifier
//   Final premium floor: max(computed, Deductible/10, 500)
//   Round to the nearest dollar.
//
// EXAMPLE
// -------
//   engine := NewPremiumRatingEngine()
//   engine.AddSubmission("sub-001", "epl", "Acme Corp",
//       50, 5_000_000, 7, "medium", 1_000_000, 10_000)
//   // -> base = 1200 + (15*50) + (5000000*0.0008) = 1200 + 750 + 4000 = 5950
//   engine.RecordPriorClaim("sub-001", 2023, 50000, "epl")
//   engine.CalculateFinalPremium("sub-001", 2025)
//   // -> base=5950, industry=×1.0, tenure=×1.0, claims=×1.15 → final=6842
// =============================================================================

package premiumrating

import "errors"

// CoverageType constants.
const (
	CoverageEPL       = "epl"
	CoverageDO        = "do"
	CoverageFiduciary = "fiduciary"
)

// IndustryRisk constants.
const (
	RiskLow    = "low"
	RiskMedium = "medium"
	RiskHigh   = "high"
)

// Base premium formula constants.
const (
	EPLBase           = 1200.0
	EPLPerEmployee    = 15.0
	EPLRevenueRate    = 0.0008
	DOBase            = 2500.0
	DORevenueRate     = 0.0010
	FiduciaryBase     = 800.0
	FiduciaryRevenueRate = 0.0004
	BasePremiumCapPct = 0.03
	BasePremiumFloor  = 500.0
)

// Industry risk multipliers.
const (
	RiskMultiplierLow    = 0.85
	RiskMultiplierMedium = 1.00
	RiskMultiplierHigh   = 1.35
)

// Tenure (years in business) multipliers.
const (
	TenureMultiplierYoung   = 1.25 // < 3 years
	TenureMultiplierMid     = 1.00 // 3–10 years
	TenureMultiplierVeteran = 0.90 // > 10 years
)

// Claims modifier constants.
const (
	ClaimsModifierPerClaim = 1.15
	ClaimsModifierCap      = 1.60
	ClaimsLookbackYears    = 2 // year >= currentYear - 2  (last 3 years inclusive)
)

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// ErrInvalidCoverageType is returned when coverage_type is not one of the valid values.
var ErrInvalidCoverageType = errors.New("invalid coverage type")

// ErrInvalidIndustryRisk is returned when industry_risk is not one of the valid values.
var ErrInvalidIndustryRisk = errors.New("invalid industry risk")

// PriorClaim represents a historical claim attached to a submission.
type PriorClaim struct {
	Year      int
	Amount    int
	ClaimType string
}

// PolicySubmission holds all data for a single policy application.
type PolicySubmission struct {
	SubmissionID    string
	CoverageType    string
	CompanyName     string
	EmployeeCount   int
	AnnualRevenue   int
	YearsInBusiness int
	IndustryRisk    string
	RequestedLimit  int
	Deductible      int
	PriorClaims     []PriorClaim
}

// FinalPremiumResult is returned by CalculateFinalPremium.
type FinalPremiumResult struct {
	BasePremium      int
	IndustryModifier float64
	TenureModifier   float64
	ClaimsModifier   float64
	FinalPremium     int
}

// HighRiskResult is a single entry returned by GetHighRiskSubmissions.
type HighRiskResult struct {
	SubmissionID      string
	CompanyName       string
	CoverageType      string
	EffectiveModifier float64
	FinalPremium      int
}

// CoverageMetrics is used inside PortfolioMetrics.
type CoverageMetrics struct {
	Count        int
	TotalPremium int
}

// PortfolioMetrics is returned by GetPortfolioMetrics.
type PortfolioMetrics struct {
	TotalSubmissions int
	TotalPremium     int
	AveragePremium   int
	ByCoverageType   map[string]*CoverageMetrics
}

// PremiumRatingEngine is the interface all implementations must satisfy.
//
// Implement it by defining your own struct and a NewPremiumRatingEngine constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myEngine struct {
//	    submissions map[string]*PolicySubmission
//	    // ... add whatever fields you need
//	}
//
//	func NewPremiumRatingEngine() PremiumRatingEngine {
//	    return &myEngine{
//	        submissions: make(map[string]*PolicySubmission),
//	    }
//	}
type PremiumRatingEngine interface {
	// -------------------------------------------------------------------------
	// PART 1 — Submission management and base premium calculation
	// -------------------------------------------------------------------------

	// AddSubmission registers a new policy submission.
	//
	// Returns the stored submission (PriorClaims starts as an empty slice).
	//
	// Returns ErrAlreadyExists if submissionID already exists.
	// Returns ErrInvalidCoverageType if coverageType is not "epl", "do", or "fiduciary".
	// Returns ErrInvalidIndustryRisk if industryRisk is not "low", "medium", or "high".
	AddSubmission(
		submissionID, coverageType, companyName string,
		employeeCount, annualRevenue, yearsInBusiness int,
		industryRisk string,
		requestedLimit, deductible int,
	) (*PolicySubmission, error)

	// GetSubmission returns the submission for the given ID.
	// Returns ErrNotFound if submissionID does not exist.
	GetSubmission(submissionID string) (*PolicySubmission, error)

	// CalculateBasePremium calculates and returns the base premium (dollars)
	// using the type-specific formula, then applies cap and floor.
	// Does NOT persist the result.
	// Returns ErrNotFound if submissionID does not exist.
	CalculateBasePremium(submissionID string) (int, error)

	// -------------------------------------------------------------------------
	// PART 2 — Prior claims and final premium with risk modifiers
	// -------------------------------------------------------------------------

	// RecordPriorClaim appends a PriorClaim to the submission's PriorClaims slice.
	// Returns ErrNotFound if submissionID does not exist.
	RecordPriorClaim(submissionID string, year, amount int, claimType string) (*PolicySubmission, error)

	// CalculateFinalPremium applies risk modifiers to the base premium and
	// returns a FinalPremiumResult breakdown.
	//
	// Calls CalculateBasePremium internally — do not duplicate its logic.
	//
	// A "qualifying" prior claim:
	//   - ClaimType matches the submission's CoverageType OR ClaimType == "any"
	//   - AND year >= currentYear - 2  (i.e. within the last 3 calendar years)
	//
	// Each qualifying claim multiplies the claims modifier by 1.15, capped at 1.60.
	// Final premium floor: max(computed, Deductible/10, 500). Round to nearest int.
	// Returns ErrNotFound if submissionID does not exist.
	CalculateFinalPremium(submissionID string, currentYear int) (*FinalPremiumResult, error)

	// -------------------------------------------------------------------------
	// PART 3 — Portfolio analytics
	// -------------------------------------------------------------------------

	// GetSubmissionsByCoverageType returns all submissions grouped by coverage type.
	// Keys are coverage types that have at least one submission.
	// Values are slices of submissions sorted by SubmissionID ascending.
	GetSubmissionsByCoverageType() map[string][]*PolicySubmission

	// GetPortfolioMetrics aggregates statistics across all submissions.
	// Calls CalculateFinalPremium for each submission — do not duplicate its logic.
	// AveragePremium is rounded to the nearest dollar (0 if no submissions).
	// ByCoverageType only includes types that have at least one submission.
	GetPortfolioMetrics(currentYear int) (*PortfolioMetrics, error)

	// GetHighRiskSubmissions returns submissions whose effective overall modifier
	// exceeds modifierThreshold.
	//
	// EffectiveModifier = FinalPremium / BasePremium, rounded to 4 decimal places.
	// Calls CalculateFinalPremium internally.
	//
	// Results are sorted by EffectiveModifier descending, then SubmissionID ascending.
	// Returns ErrNotFound if any submission cannot be found (should not normally happen).
	GetHighRiskSubmissions(currentYear int, modifierThreshold float64) ([]*HighRiskResult, error)
}

// NewPremiumRatingEngine returns a new PremiumRatingEngine implementation.
//
// You must define your own struct type that implements the PremiumRatingEngine
// interface and return it from this constructor.
//
// All state must be stored in struct fields — do NOT use package-level variables.
func NewPremiumRatingEngine() PremiumRatingEngine {
	panic("not implemented")
}
