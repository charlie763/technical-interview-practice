package ratelimiter

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// ErrInvalidPlan is returned when referencing a plan name not in Gateway.Plans.
var ErrInvalidPlan = errors.New("invalid plan")

// IntPtr is a convenience helper for creating *int literals in plan definitions.
func IntPtr(i int) *int { return &i }

// Plan holds per-minute and per-day request caps. nil means unlimited.
type Plan struct {
	RPM *int // requests per minute; nil = unlimited
	RPD *int // requests per day;    nil = unlimited
}

// APIKey represents a registered API key.
type APIKey struct {
	ID         string
	Owner      string
	Plan       string    // must be a key in Gateway.Plans
	Enabled    bool
	RequestLog []float64 // sorted Unix timestamps of recent requests
}

// Gateway holds all gateway state.
type Gateway struct {
	Plans map[string]*Plan
	Keys  map[string]*APIKey
}

// RequestResult is returned by HandleRequest.
type RequestResult struct {
	Allowed bool
	KeyID   string // set when Allowed == true
	Reason  string // set when Allowed == false: "key_not_found" | "key_disabled" | "rpm_exceeded" | "rpd_exceeded"
}

// UsageStats is returned by GetUsage.
type UsageStats struct {
	KeyID    string
	Plan     string
	RPMUsed  int
	RPMLimit *int // nil if unlimited
	RPDUsed  int
	RPDLimit *int // nil if unlimited
}
