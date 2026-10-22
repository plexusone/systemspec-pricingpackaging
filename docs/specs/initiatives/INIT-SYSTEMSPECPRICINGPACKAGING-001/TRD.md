# systemspec-pricingpackaging: Technical Requirements

**Initiative:** `INIT-SYSTEMSPECPRICINGPACKAGING-001`
**Repository:** `github.com/plexusone/systemspec-pricingpackaging`

Status: **frozen for v0.1** by Phase 1. The normative description is
[SPEC.md](../../../../SPEC.md); this document records the reasoning. The model was
checked against three product shapes in [FIT.md](FIT.md), which changed it as
listed under Changes from the fit check.

## Principles

- **The Spec is the source of truth.** Everything else (a page, a check, a
  provider plan) is a projection.
- **Static types.** No `oneOf`, `anyOf` or `allOf`, no untyped or heterogeneous
  maps. A feature value is an explicit pair of named fields, not a value that is
  sometimes a boolean and sometimes a string.
- **Facts, not prose.** Descriptions are display text; rules are fields.
- **Safe defaults.** Nothing grants more than the spec states.
- **Standard library only.**

Package: `pricingspec`. Generated schema: `schema/pricing.schema.json`, linted
with `schemakit lint --property-case camelCase`.

## The model (proposed)

```go
type Spec struct {
    FormatVersion string
    App           string        // identifies the product; tags what a provider holds
    Currency      string        // ISO code, lowercase
    DefaultPlan   string        // a plan ID; the free plan
    Applications  []Application // what grants may refer to
    Tiers         []Tier
    Features      []Feature
    Limits        []Limit
    Plans         []Plan
    Products      []Product     // one-time purchases
    Marketplace   *Marketplace
    Entitlements  []TierGrant   // what each tier grants an account
}
```

### Tiers

```go
type Tier struct {
    ID   string  // "free", "plus", "pro"
    Name string
    Rank int     // order, lowest first; unique
}
```

A tier is a position in the product's ladder, and the word the entitlement and
the subscription use. A plan belongs to exactly one tier. A tier may have
several plans (monthly and yearly are prices on one plan; a separate plan is for
a different offer at the same tier).

### Features

```go
type Feature struct {
    ID          string
    Name        string
    Description string
    Levels      []string  // empty: the feature is on or off; else ordered, lowest first
}

type FeatureValue struct {
    Feature string
    Enabled bool      // for a feature without levels
    Level   string    // for a feature with levels; one of its Levels
}
```

A plan lists a `FeatureValue` for each feature it sets; an unset feature is off
(or at no level). A check asks `Has(feature)` for on/off features or
`AtLeast(feature, level)` for levelled ones.

### Limits

```go
type Limit struct {
    ID          string
    Name        string
    Unit        string   // "board", "seat", "export"
    Period      Period   // none, day, month
    Description string
}

type LimitValue struct {
    Limit     string
    Unlimited bool     // explicit; Max must be zero when true
    Max       int      // zero means zero
    FairUse   bool     // with Max: crossing warns instead of refusing; with Unlimited: advertised as unlimited, subject to the host's policy
}
```

`Unlimited` is a boolean, not a null or a zero, so a mistake never grants more
than intended. The host counts usage (including per period); the spec defines
the allowance and the check.

### Plans and products

```go
type Plan struct {
    ID        string
    Name      string
    Tier      string
    SelfServe bool         // can be bought through checkout; false: arranged
    PlatformFeeBps *int    // optional override of the marketplace fee
    TrialDays int
    Prices    []Price      // empty for a free or arranged plan
    Features  []FeatureValue
    Limits    []LimitValue
}

type Price struct {
    Interval    Interval   // month, year
    AmountCents int64
}

type Product struct {                 // a one-time purchase
    ID          string
    Name        string
    AmountCents int64      // fixed price, or zero when SellerSet
    SellerSet   bool       // the seller sets the price
    MinAmountCents int64   // floor for a seller-set price
}

type Marketplace struct {
    PlatformFeeBps int     // basis points, 0 to 10000
}
```

### Entitlements

```go
type TierGrant struct {
    Tier        string
    Application string   // the application the tier grants access to
    Seats       int      // 0 with Unlimited false means none
    UnlimitedSeats bool
}
```

This is the **tier-to-access map**: what an account on a tier may use. A
runtime that turns a subscription into an entitlement reads the plan's tier and
then this list. Features and limits are **not** copied into entitlements; they
are looked up in the Spec by plan at the moment of a check, so prices and
allowances can change without reissuing anything.

## Validation

Errors (the Spec is rejected):

- Duplicate IDs among tiers, features, limits, plans or products.
- A plan whose tier, or a value whose feature or limit, is not declared.
- A level not among the feature's levels; a feature with levels given a boolean.
- Duplicate or non-positive tier ranks.
- A negative amount, an amount not in minor units, or two prices for one
  interval.
- A default plan that is missing, not self-serve free (no price), or in no tier.
- A `Product` that is seller-set with no minimum, or fixed with no amount.
- A `Marketplace` fee outside 0 to 10000.

Lint warnings (reviewable, not fatal):

- No free tier, or a free plan with unlimited limits and no fair-use note.
- A paid plan with a lower-ranked tier than a cheaper plan.
- A feature declared and never used; a limit declared and never used.
- A tier with no entitlement grant.

## Helpers

```go
spec.PlanOrDefault(id string) Plan
plan.Has(feature string) bool
plan.AtLeast(feature, level string) bool
plan.Check(limit string, after int) Decision // after: usage after the action; Allowed, Warn, Refuse
spec.TierRank(tier string) int
```

An unknown plan ID returns the default plan. Checks never panic and never grant
on an unknown identifier.

## Projections

- **PricingView:** plans in tier order with prices by interval, and a feature
  and limit matrix with display names, for a page to render.
- **EntitlementMap:** tier to its grants.
- **Diff:** two Specs compared into added, removed and changed plans, prices,
  features, limits and grants, with price increases and removed grants marked
  as changes that affect existing customers.

A description of what a payment provider should hold (products and prices by a
stable key) is computed from the Spec by the billing library that syncs it, not
here.

## Versioning

`FormatVersion` is a string, bumped on a change to the document format. New
optional fields are additions. A reader rejects a version it does not know.

## Testing

- Table tests for every validation error and lint warning.
- Round-trip: Go → JSON → Go, and the generated schema validates the examples.
- Helper tests for unknown identifiers, level ordering, unlimited versus zero,
  and fair-use results.
- Golden files for the projections and the diff.
- The family's schema lint with zero findings.

## Changes from the fit check

[FIT.md](FIT.md) checked the draft against three product shapes. The model
changed in these ways, all reflected in SPEC.md:

- `Plan.PlatformFeeBps` overrides the marketplace fee per plan.
- `Applications` are declared, and grants refer to them.
- `FairUse` may accompany `Unlimited`.
- Levels are only for monotonic ladders; a set of independent options is several
  switches.
- Seats live only in grants.
- What a one-time purchase issues, who may pay and audience grouping are out of
  scope for v0.1.

## Tier vocabulary and the mapping to entitlements

Tiers are declared per product with an order; there is no global vocabulary, and
ranks of different products are not comparable. An account's entitlement carries
the plan's tier ID unchanged. Access (applications, seats, validity) is the
tier's grant and travels with the entitlement. Features and limits are read from
the plan the host stores with the subscription, so two plans in one tier can
differ in allowances while giving the same access. The rules are in SPEC.md under
Entitlement mapping.

## Open questions

- Whether prices in several currencies belong in the first release.
- How a product with more than one deployable (several applications) shares
  tiers.
