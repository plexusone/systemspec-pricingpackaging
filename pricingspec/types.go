package pricingspec

// FormatVersion is the document format version this package reads and writes.
const FormatVersion = "1"

// Period is how often a limit's usage resets.
type Period string

// Limit periods.
const (
	PeriodNone  Period = "none"
	PeriodDay   Period = "day"
	PeriodMonth Period = "month"
)

// Interval is a billing interval.
type Interval string

// Billing intervals.
const (
	IntervalMonth Interval = "month"
	IntervalYear  Interval = "year"
)

// Spec describes a product's pricing and packaging.
type Spec struct {
	// FormatVersion is "1". A reader rejects a version it does not know.
	FormatVersion string `json:"formatVersion" jsonschema:"description=The document format version,enum=1"`
	// App identifies the product; everything a billing system holds for it is tagged with it.
	App string `json:"app" jsonschema:"description=Identifies the product,pattern=^[a-z][a-z0-9-]*$"`
	// Currency is a lowercase ISO 4217 code. One currency per Spec.
	Currency string `json:"currency" jsonschema:"description=Lowercase ISO 4217 currency code,pattern=^[a-z]{3}$"`
	// DefaultPlan is the plan for any unrecognised plan value and for an account
	// when billing is not configured.
	DefaultPlan string `json:"defaultPlan" jsonschema:"description=Plan ID used for any unrecognised plan and when billing is not configured"`
	// Applications are the applications an account can be granted.
	Applications []Application `json:"applications" jsonschema:"description=The applications a tier can grant"`
	// Tiers are the product's ladder, lowest rank first.
	Tiers []Tier `json:"tiers" jsonschema:"description=The product's ladder"`
	// Features are what a plan can switch on or set to a level.
	Features []Feature `json:"features,omitempty" jsonschema:"description=What a plan can switch on or set to a level"`
	// Limits are what a plan can cap.
	Limits []Limit `json:"limits,omitempty" jsonschema:"description=What a plan can cap"`
	// Plans are what an account can be on.
	Plans []Plan `json:"plans" jsonschema:"description=What an account can be on"`
	// Products are one-time purchases.
	Products []Product `json:"products,omitempty" jsonschema:"description=One-time purchases"`
	// Marketplace is the platform's share of sales by sellers, when there are any.
	Marketplace *Marketplace `json:"marketplace,omitempty" jsonschema:"description=The platform's share of sales by sellers"`
	// Entitlements are what each tier grants an account.
	Entitlements []TierGrant `json:"entitlements" jsonschema:"description=What each tier grants an account"`
}

// Application is an application an account can be granted access to.
type Application struct {
	ID   string `json:"id" jsonschema:"pattern=^[a-z][a-z0-9_]*$"`
	Name string `json:"name"`
}

// Tier is a position on the product's ladder.
type Tier struct {
	ID   string `json:"id" jsonschema:"pattern=^[a-z][a-z0-9_]*$"`
	Name string `json:"name"`
	// Rank is a positive integer, unique within the Spec, lowest first. Ranks
	// of different products are not comparable.
	Rank int `json:"rank" jsonschema:"minimum=1"`
}

// Feature is something a plan switches on, or sets to a level.
type Feature struct {
	ID          string `json:"id" jsonschema:"pattern=^[a-z][a-z0-9_]*$"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Levels, when present, make the feature a ladder: at least two distinct
	// names, lowest first, each including the one below. Without levels the
	// feature is a switch.
	Levels []string `json:"levels,omitempty" jsonschema:"description=Ordered levels of a ladder feature; absent for a switch"`
}

// Limit is something a plan caps.
type Limit struct {
	ID          string `json:"id" jsonschema:"pattern=^[a-z][a-z0-9_]*$"`
	Name        string `json:"name"`
	Unit        string `json:"unit,omitempty"`
	Period      Period `json:"period" jsonschema:"enum=none,enum=day,enum=month"`
	Description string `json:"description,omitempty"`
}

// Plan is what an account can be on.
type Plan struct {
	ID   string `json:"id" jsonschema:"pattern=^[a-z][a-z0-9_]*$"`
	Name string `json:"name"`
	// Tier is the plan's tier. Exactly one.
	Tier string `json:"tier"`
	// SelfServe means the plan can be bought through checkout; false means arranged.
	SelfServe bool `json:"selfServe"`
	// TrialDays is the length of a free trial. Only on a plan with prices.
	TrialDays int `json:"trialDays,omitempty" jsonschema:"minimum=0"`
	// Prices is empty for a free or arranged plan.
	Prices []Price `json:"prices,omitempty"`
	// Features are the features the plan sets; an unset feature is off.
	Features []FeatureValue `json:"features,omitempty"`
	// Limits are the limits the plan sets; an unset limit is zero.
	Limits []LimitValue `json:"limits,omitempty"`
	// PlatformFeeBps overrides the marketplace fee for sales by an account on this plan.
	PlatformFeeBps *int `json:"platformFeeBps,omitempty" jsonschema:"minimum=0,maximum=10000"`
}

// Price is a plan's price for one billing interval.
type Price struct {
	Interval Interval `json:"interval" jsonschema:"enum=month,enum=year"`
	// AmountCents is a positive integer in the currency's minor units.
	AmountCents int64 `json:"amountCents" jsonschema:"minimum=1"`
}

// FeatureValue sets one feature for a plan.
type FeatureValue struct {
	Feature string `json:"feature"`
	// Enabled states a switch. It must be false for a ladder.
	Enabled bool `json:"enabled,omitempty"`
	// Level names one of a ladder's levels. It must be empty for a switch.
	Level string `json:"level,omitempty"`
}

// LimitValue sets one limit for a plan.
type LimitValue struct {
	Limit string `json:"limit"`
	// Max is the allowance; zero means zero. It must be zero when Unlimited.
	Max int `json:"max,omitempty" jsonschema:"minimum=0"`
	// Unlimited means no numeric limit.
	Unlimited bool `json:"unlimited,omitempty"`
	// FairUse: with Max, exceeding it warns and flags the account for review
	// instead of refusing; with Unlimited, the allowance is advertised as
	// unlimited and subject to the host's fair-use policy.
	FairUse bool `json:"fairUse,omitempty"`
}

// Product is a one-time purchase.
type Product struct {
	ID   string `json:"id" jsonschema:"pattern=^[a-z][a-z0-9_]*$"`
	Name string `json:"name"`
	// AmountCents is the fixed price; zero when SellerSet.
	AmountCents int64 `json:"amountCents,omitempty" jsonschema:"minimum=0"`
	// SellerSet means the seller sets the price.
	SellerSet bool `json:"sellerSet,omitempty"`
	// MinAmountCents is the floor for a seller-set price.
	MinAmountCents int64 `json:"minAmountCents,omitempty" jsonschema:"minimum=0"`
}

// Marketplace is the platform's share of sales by sellers.
type Marketplace struct {
	// PlatformFeeBps is the share in basis points, 0 to 10000.
	PlatformFeeBps int `json:"platformFeeBps" jsonschema:"minimum=0,maximum=10000"`
}

// TierGrant is what a tier gives an account for one application.
type TierGrant struct {
	Tier        string `json:"tier"`
	Application string `json:"application"`
	// Seats is the number of members, at least 1, unless UnlimitedSeats.
	Seats          int  `json:"seats,omitempty" jsonschema:"minimum=0"`
	UnlimitedSeats bool `json:"unlimitedSeats,omitempty"`
}
