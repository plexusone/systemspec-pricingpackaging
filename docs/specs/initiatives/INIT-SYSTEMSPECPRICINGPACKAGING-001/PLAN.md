# systemspec-pricingpackaging: Plan

**Initiative:** `INIT-SYSTEMSPECPRICINGPACKAGING-001`
**Repository:** `github.com/plexusone/systemspec-pricingpackaging`

Implements the [PRD](PRD.md) and [TRD](TRD.md). The [roadmap](ROADMAP.md)
holds the work items and their status.

## Approach

- **Small first.** The first release covers recurring plans, one-time products,
  a marketplace fee and the tier-to-entitlement map, and no more. Offers that
  need more extend the model with examples behind them.
- **Types, then schema, then rules.** The Go types are specified and frozen,
  the schema generated from them, and validation and lint written against the
  types.
- **Two shapes before freezing.** A subscription product with levelled features
  and a marketplace product both have to be expressible, so the model is not
  frozen against one.
- **Projections are tested against golden files,** so a change to a page, a map
  or a diff is visible in review.
- **Public-ready from the first commit.** Examples are invented; no real
  products, prices or customers.

## Phases

| Phase | Theme | Outcome |
|---|---|---|
| 1 | Design | The model checked against two product shapes, the tier and entitlement vocabulary fixed, and the TRD frozen |
| 2 | The model | The types, the generated schema, validation and lint |
| 3 | Helpers and projections | Plan, feature and limit helpers; the pricing view, entitlement map and diff |
| 4 | Examples and release | Invented examples for two shapes, a guide, and v0.1.0 |

## Order and dependencies

Phase 1 gates the rest. Phase 3 depends on the types from Phase 2. The billing
library that consumes the spec is planned in its own repository and starts once
the types are frozen.

## Verification

- Validation and lint: a test for each rule.
- Helpers: unknown identifiers, level ordering, unlimited versus zero.
- Projections: golden files.
- Schema: generated, linted with the family's tool, and used to validate the
  examples.
- A billing library imports the package and builds a pricing view, a limit
  check and a provider plan from an example before v0.1.0.
