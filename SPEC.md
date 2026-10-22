# systemspec-pricingpackaging Specification v0.1

This document is the normative specification of the systemspec-pricingpackaging
data model: the `pricingspec` package's types and the JSON documents they marshal
to and from.

A **Spec** describes a product's pricing and packaging as data: the **tiers** of
its ladder, the **plans** in them with their prices, the **features** (a switch
or an ordered level) and **limits** (counts, with an optional period) each plan
sets, its **one-time products**, an optional **marketplace** fee, and the
**grants** each tier gives an account. A pricing page, a limit check, a payment
provider's prices and an account's entitlements are computed from it; none of
them is the model.

Three invariants govern the model:

- **Nothing grants more than the Spec states.** An unknown plan is the default
  plan; an unset feature is off; a limit that is absent is zero; "unlimited" is
  its own field and never a zero or a missing value.
- **Access is a grant; allowances are looked up.** What an account may open and
  how many members it may have is a *grant* of its tier and travels with the
  account's entitlement. Features and limits are *not* copied into entitlements:
  they are read from the Spec by the account's plan at the moment of a check.
- **The Spec is the offer, not the policy.** It records what is offered. Who may
  buy, what payment may never buy, and what a purchase issues are the host's.

Every type conforms to the PlexusOne static-type profile: no `oneOf`, `anyOf`,
`allOf` or other schema composition, and no untyped or heterogeneous maps.
Controlled vocabularies are string enums and optional data is explicit named
fields. The generated schema is `schema/pricing.schema.json`; it passes
`schemakit lint --property-case camelCase` with zero findings.

## Spec

| Field | Type | Meaning |
|---|---|---|
| `formatVersion` | string | `"1"`. A reader rejects a version it does not know. |
| `app` | string | Identifies the product. Everything a billing system holds for it is tagged with this. `^[a-z][a-z0-9-]*$`. |
| `currency` | string | Lowercase ISO 4217 code. One currency per Spec. |
| `defaultPlan` | string | A plan ID: the plan for any unrecognised plan value, and for an account when billing is not configured. |
| `applications` | []Application | The applications an account can be granted. |
| `tiers` | []Tier | The ladder. |
| `features` | []Feature | |
| `limits` | []Limit | |
| `plans` | []Plan | |
| `products` | []Product | One-time purchases. |
| `marketplace` | Marketplace | Optional. |
| `entitlements` | []TierGrant | What each tier grants. |

IDs (`^[a-z][a-z0-9_]*$` except `app`) are unique within their kind.

## Application

`{ id, name }`. An application a tier can be granted: a product area or module
of the offering. Grants refer to applications by `id`.

## Tier

`{ id, name, rank }`. A position on the product's ladder.

- `rank` is a positive integer, unique within the Spec, lowest first. Ranks are
  meaningful only within one Spec; tiers of different products are not
  comparable.
- A tier ID is the word used for an account's tier in its entitlement and
  subscription. Conventional IDs such as `free` and `enterprise` carry no
  special meaning to the model.

## Feature

`{ id, name, description, levels }`.

- With no `levels` the feature is a **switch**: on or off.
- With `levels` it is a **ladder**: an ordered list, lowest first, of at least
  two distinct names. A plan sets a ladder to one of its levels, or leaves it
  unset (off, below the lowest level).
- Levels are **only for monotonic ladders**, where each level includes the one
  below. A capability that is a set of independent options (video providers,
  payment methods) is several switches.

## Limit

`{ id, name, unit, period, description }`. `period` is `none`, `day` or
`month`. The host counts usage, including per period; the Spec defines the
allowance and the check.

## Plan

| Field | Type | Meaning |
|---|---|---|
| `id`, `name` | string | |
| `tier` | string | The plan's tier. Exactly one. |
| `selfServe` | bool | The plan can be bought through checkout. False means arranged. |
| `trialDays` | int | Zero or more. Only on a plan with prices. |
| `prices` | []Price | Empty for a free or arranged plan. |
| `features` | []FeatureValue | |
| `limits` | []LimitValue | |
| `platformFeeBps` | int, optional | Overrides the marketplace fee for sales by an account on this plan. |

### Price

`{ interval, amountCents }`. `interval` is `month` or `year`; `amountCents` is a
positive integer in the currency's minor units. A plan has at most one price per
interval.

### FeatureValue

`{ feature, enabled, level }`. For a switch, `enabled` states it; `level` must be
empty. For a ladder, `level` names one of the feature's levels; `enabled` must
be false.

### LimitValue

`{ limit, max, unlimited, fairUse }`.

- `unlimited: true` means no numeric limit; `max` must be zero.
- Otherwise `max` is the allowance and **zero means zero**.
- `fairUse` changes the meaning of exceeding the allowance. With a `max`,
  exceeding it **warns and flags the account for review** instead of refusing.
  With `unlimited`, it records that the allowance is advertised as unlimited and
  subject to the host's fair-use policy; the Spec sets no threshold.

## Product

A one-time purchase: `{ id, name, amountCents, sellerSet, minAmountCents }`.

- A fixed product has `amountCents > 0`, and `sellerSet` false.
- A seller-set product has `sellerSet: true`, `amountCents` zero and
  `minAmountCents > 0`.

What a purchase issues is the host's and is identified by the product ID.

## Marketplace

`{ platformFeeBps }`: the platform's share of a sale, in basis points, 0 to
10000. A plan's `platformFeeBps` overrides it.

## TierGrant

`{ tier, application, seats, unlimitedSeats }`: the account on that tier may use
that application, with that many members or without a limit.

- A grant has `seats >= 1` or `unlimitedSeats: true`, never both.
- A tier has at most one grant per application. To give a tier no access to an
  application, omit the grant.
- Seats live **only** in grants. A plan does not also carry a member limit.

## Validation

A Spec that breaks an error rule is rejected. Each rule has a stable code.

| Code | Rule |
|---|---|
| E001 | `formatVersion` is known. |
| E002 | `app` and `currency` are well formed. |
| E003 | IDs are well formed and unique within their kind. |
| E004 | Tier ranks are positive and unique. |
| E005 | Every plan names a declared tier. |
| E006 | Every feature value names a declared feature, once per plan; a switch value has no level; a ladder value names one of its levels. |
| E007 | A feature with `levels` has at least two distinct, non-empty levels. |
| E008 | Every limit value names a declared limit, once per plan; `unlimited` is set with `max` zero; `max` is not negative. |
| E009 | Prices are positive integers, at most one per interval; a self-serve plan other than the default has at least one price; a plan with `trialDays` has prices. |
| E010 | `defaultPlan` names a plan that is self-serve and has no price, and its tier has a grant. |
| E011 | A product is fixed (positive amount) or seller-set (zero amount, positive minimum), not both. |
| E012 | Marketplace and plan fees are within 0 to 10000. |
| E013 | Every grant names a declared tier and application; seats are positive or unlimited, not both; at most one grant per tier and application. |

Lint warnings are reviewable and do not reject:

| Code | Warning |
|---|---|
| W001 | The default plan sets an unlimited limit without `fairUse`. |
| W002 | A plan in a higher tier is cheaper than a plan in a lower tier at the same interval. |
| W003 | A feature or limit is declared and never used by a plan. |
| W004 | A tier has plans but no grants. |
| W005 | A ladder's level names read as a list of independent options (they contain a comma, a slash, or the words "or" or "and"). |
| W006 | A plan has an annual price that is not less than twelve times the monthly price. |

## Resolution

For an identifier that is not in the Spec, every function resolves to the
safest answer and never fails.

- `PlanOrDefault(id)`: the plan, or the default plan.
- `Has(feature)`: true for a switch that is enabled; true for a ladder set to any
  level.
- `AtLeast(feature, level)`: true when the plan's level is at or above `level`
  in the feature's order. False for an unset ladder, an unknown feature or an
  unknown level.
- `Check(limit, after)`: `after` is the usage **after** the action being
  checked. `Allowed` when the limit is `unlimited`, or `after <= max`. Over the
  maximum, `Warn` when `fairUse` and otherwise `Refuse`. A limit the plan does
  not set has a maximum of zero, so any `after >= 1` is `Refuse`. An `unlimited`
  limit with `fairUse` is always `Allowed`; the Spec sets no threshold and the
  host applies its fair-use policy.

## Entitlement mapping

An account has a **plan** (stored with its subscription) and a **tier** (the
plan's tier). The mapping to entitlements is:

1. The entitlement the account holds for each application is the tier's grant:
   the application, the seats, and a validity that ends with the subscription's
   current period (or never for a free plan).
2. The entitlement's `tier` is the plan's tier ID, unchanged.
3. Features and limits are read from the plan, which the host stores, and not
   from the tier: two plans in one tier may set different allowances while
   giving the same access.
4. An account whose plan is unknown, whose subscription has ended, or that has no
   billing is on the default plan.

## Projections

- **PricingView:** plans in tier-rank order with prices by interval and a matrix
  of features and limits with display names. Suitable for rendering.
- **EntitlementMap:** tier to its grants, in rank order.
- **Diff:** two Specs compared into added, removed and changed tiers, plans,
  prices, features, limits and grants. A price increase, a lowered allowance, a
  removed feature and a removed grant are marked as affecting existing
  customers.

## Versioning

`formatVersion` is bumped for a change to the document format. A new optional
field is an addition and does not bump it. A reader rejects a version it does
not know.
