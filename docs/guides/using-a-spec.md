# Using a spec

Load the Spec once at startup and keep it. Every function below takes the
account's **plan ID**, which your product stores with the account's
subscription. An unknown ID is treated as the default plan.

## Features and levels

```go
spec.Has(account.Plan, "audit_log")                // a switch, or a ladder at any level
spec.AtLeast(account.Plan, "support", "email")     // a ladder at or above a level
spec.Level(account.Plan, "support")                // "community", "email", ...
```

`AtLeast` is false for an unset ladder, a switch, an unknown feature or an
unknown level, so a check can only grant what the Spec states.

## Limits

```go
switch spec.Check(account.Plan, "boards", boardCount+1) { // usage after the action
case pricingspec.Allowed:
	// proceed
case pricingspec.Warn:
	// a fair-use allowance is exceeded: proceed, and flag the account
case pricingspec.Refuse:
	// tell the person which plan allows more
}
```

`Check` takes the usage *after* the action: with a maximum of 3 and 2 boards in
use, adding a third is checked as 3 and is allowed. You count the usage,
including per-month counts; the Spec defines the allowance. `Allowance` returns
the maximum and the unlimited and fair-use flags, for showing "2 of 3 used".

## Access

```go
for _, g := range spec.GrantsFor(account.Plan) {
	// g.Application, g.Seats, g.UnlimitedSeats
}
```

Use the grants when you turn a subscription into the account's entitlement.
Do not copy features or limits into it.

## A pricing page

```go
view := spec.PricingView()
for _, p := range view.Plans { /* a column: name, tier, prices, free, default */ }
for _, row := range view.Features { /* one row; row.Cells line up with view.Plans */ }
for _, row := range view.Limits { /* the same, with Max, Unlimited, FairUse */ }
```

The view is data, not text: you choose the words. Plans come in tier order, and
the cells of every row are in the same order as the columns.

## What changed

```go
for _, c := range pricingspec.Diff(oldSpec, newSpec) {
	fmt.Println(c) // changed plan plus: month price 1200 to 1500 [affects existing customers]
}
```

Run it on a pull request to see an offer change as a review. A change is marked
as affecting existing customers when it makes things worse for accounts already
on it: a higher price, a lower allowance, a removed feature or grant, or a
removed plan.

## The tier and the entitlement

An account's entitlement carries the plan's tier ID unchanged, and the tier's
grants give it access. The account's plan, which you store, is what you pass to
every function above. See the specification for the full mapping.
