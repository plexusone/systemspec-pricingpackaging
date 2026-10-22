# systemspec-pricingpackaging: Product Requirements

**Initiative:** `INIT-SYSTEMSPECPRICINGPACKAGING-001`
**Repository:** `github.com/plexusone/systemspec-pricingpackaging`

## Problem

A product's pricing and packaging (what each tier costs, what it includes,
what it limits, and what access it gives an account) is decided by people and
then scattered into code: a price in a payment dashboard, a feature check in a
handler, a tier name in a database column, a table on a marketing page. The
copies drift. A change to a price or a limit is an edit in several places, no
one can review the whole offer as one thing, and nothing can tell whether the
pages, the checks and the payment provider agree.

Other aspects of a system already have a machine-readable description in this
family of specs: how an API is styled, what cryptography is used, how it is
deployed. Pricing and packaging has none.

## Goals

1. **One description of a product's pricing and packaging,** as data: tiers,
   plans, prices, features (on, off or a level), limits (with periods and
   fair-use), one-time products and marketplace fees.
2. **A mapping from tier to account entitlements:** which applications and
   how many seats a tier grants, so a paid plan can become what an account may
   use, using a shared vocabulary of tiers.
3. **Safe by construction.** Unknown plans resolve to a declared free default;
   "unlimited" is explicit and never a zero; a spec that would grant more than
   it states is rejected.
4. **Validated and reviewable.** A schema, lint rules for common mistakes, and
   a diff that shows what changed in an offer.
5. **Projections, not copies.** A pricing-page view, an entitlement map and a
   description of what a payment provider should hold are computed from the
   spec, never maintained alongside it.
6. **Product-neutral and minimal.** Any product can describe its offer; the
   first release covers recurring plans, one-time products and a marketplace
   fee, and nothing more.

## Non-goals

- **Not a billing system.** Charging, subscriptions, webhooks and payouts are
  a runtime library's work. This is the description it reads.
- **Not a pricing policy.** The spec records decisions; it does not recommend
  prices.
- **Not revenue reporting,** tax, invoicing or accounting.
- **Not a user interface.** It supplies the data a pricing page renders.
- **Not a licence manager** for installed software.

## Users

- **Product owners,** who write and review a product's offer as one document.
- **Billing libraries and platforms,** which read it to create prices, check
  features and limits, and turn a subscription into entitlements.
- **Front ends,** which render a pricing page from the spec's view.
- **Reviewers and agents,** who diff offers and lint them.

## Requirements

### Functional

- A **Spec** document with a format version, a product identifier, a currency,
  declared tiers with an order, declared features (each on/off or with ordered
  levels), declared limits (each with a unit and optionally a period), plans
  and one-time products, an optional marketplace fee, a default plan, and the
  entitlements each tier grants.
- **Validation:** unique identifiers, declared references, one tier per plan,
  integer prices, explicit unlimited, a free default, and ordered levels.
- **Helpers:** resolve a plan or the default, test a feature or a level, and
  check a count against a limit with allowed, warn and refuse results.
- **Projections:** a pricing view for rendering, a tier-to-entitlement map, and
  a diff between two specs.
- A generated **JSON Schema** that passes the family's lint with no findings.

### Quality

- **Go-first.** The Go types are the source of truth; the schema is generated.
- **Static types.** Following the family's profile: no schema composition, no
  untyped or heterogeneous maps; vocabularies are enums and optional data is
  named fields.
- **No payment or framework dependency.** The package imports the standard
  library only.
- **Public-ready.** Examples are invented; no real products, prices or
  customers appear in the repository.

## Success measures

- A product's offer is one reviewable file, and a pricing page, a limit check
  and a provider plan are all computed from it.
- A price or limit change is a diff of that file.
- Two products with different shapes (a subscription tool and a marketplace)
  are both expressed without extending the model.

## Risks

- **Over-modeling.** Offers vary widely. The first release is held to recurring
  plans, one-time products and a marketplace fee, and grows by examples.
- **Static-type constraints** make "a feature is either on or a level" awkward;
  the model uses explicit named fields for each.
- **Tier vocabulary disagreement** between products. The spec makes tiers
  declared and ordered per product and defines the mapping to entitlements, so
  differences are explicit.
