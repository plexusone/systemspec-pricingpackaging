# ROADMAP — systemspec-pricingpackaging: Pricing and Packaging Spec

**Initiative:** `INIT-SYSTEMSPECPRICINGPACKAGING-001`
**Repository:** `github.com/plexusone/systemspec-pricingpackaging`

Specs: [PRD](PRD.md), [TRD](TRD.md), [PLAN](PLAN.md). Work items carry stable
IDs (`RMI-SYSTEMSPECPRICINGPACKAGING-NNN`), are reviewed and executed by phase,
and commits implementing one carry the trailer
`Refs: RMI-SYSTEMSPECPRICINGPACKAGING-<NNN>`.

Later, not scheduled: prices in several currencies, seat-based and metered
prices, coupons, regional pricing, and a command-line tool.

## Phase 1 — Design

**Theme:** The model is checked against a subscription product and a marketplace product, the tier and entitlement vocabulary is fixed, and the TRD is frozen.

- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-001` Check the proposed model against two product shapes (a subscription tool with levelled features, and a marketplace with one-time sales) and record what does not fit
- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-002` Fix the tier vocabulary and the tier-to-entitlement map, including what is a grant and what is looked up in the spec at check time
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-001`
- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-003` Write SPEC.md, the normative description of the model, and freeze the TRD
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-002`

## Phase 2 — The model

**Theme:** The types exist, the schema is generated from them, and a spec that grants more than it states is rejected.

- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-004` Module, package layout and the Go types
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-003`
- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-005` JSON Schema generated from the types and linted with the family's tool
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-004`
- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-006` Loader and validation: every error in the TRD, each with a test
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-004`
- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-007` Lint warnings: no free tier, unused features and limits, tiers without grants, unlimited free limits without fair-use
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-006`

## Phase 3 — Helpers and projections

**Theme:** A host can resolve a plan, check a feature or a limit, render a pricing page and see what changed between two offers.

- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-008` Helpers: plan or default, feature and level checks, limit checks with allowed, warn and refuse
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-006`
- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-009` Pricing view for rendering, with golden files
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-008`
- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-010` Entitlement map from tier to grants
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-008`
- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-011` Diff between two specs, marking price increases and removed grants as affecting existing customers
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-008`

## Phase 4 — Examples and release

**Theme:** Two invented examples exercise the model, a billing library has built from it, and the package is released.

- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-012` Examples: a subscription tool with levelled features and a marketplace with one-time sales
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-009`, `RMI-SYSTEMSPECPRICINGPACKAGING-010`
- [x] `RMI-SYSTEMSPECPRICINGPACKAGING-013` User guide: writing a spec, reading it, and the projections
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-012`
- [ ] `RMI-SYSTEMSPECPRICINGPACKAGING-014` A billing library's trial: pricing view, limit check and provider plan built from an example
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-012`
- [ ] `RMI-SYSTEMSPECPRICINGPACKAGING-015` Review; fix findings; record known limitations; v0.1.0 release with changelog, notes and tag after CI
  - Depends on: `RMI-SYSTEMSPECPRICINGPACKAGING-013`, `RMI-SYSTEMSPECPRICINGPACKAGING-014`
