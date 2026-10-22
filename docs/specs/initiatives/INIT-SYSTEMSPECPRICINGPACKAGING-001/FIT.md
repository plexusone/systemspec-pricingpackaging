# systemspec-pricingpackaging: Fit of the Model to Three Product Shapes

**Initiative:** `INIT-SYSTEMSPECPRICINGPACKAGING-001`
**Work item:** `RMI-SYSTEMSPECPRICINGPACKAGING-001`

The model in the [TRD](TRD.md) was written from one product's plans. This
record checks it against three invented shapes that between them cover what the
first release must express, and lists what did not fit and how the model
changes. The shapes are composites; none is a real product.

## The three shapes

**A. A tools suite for small teams.** Three applications behind one sign-in.
Tiers Free, Starter, Team and an arranged Enterprise. Limits are counts
(projects, records) and a daily rate; one feature is a ladder (sync: daily,
hourly).
Seats differ by tier. One platform account, several applications.

**B. A publishing platform for sellers.** Tiers Basic, Plus, Pro. Many
differences are levels, not switches: branding (required, optional, removed),
report and analytics depth, export styling, API access (none, limited, full).
Item count is a limit, "viewers" is advertised as unlimited with a fair-use
policy, and exports per month is a fair-use allowance. Sellers sell items to
buyers and the platform takes a fee; selling is unavailable on Basic. Media
hosting differs by tier as a set of providers.

**C. A household app.** A free tier that nobody pays for, a household
subscription paid by one adult, a one-time purchase that issues a verified
record, a marketplace where sellers set prices and the platform takes a share,
and an arranged institution licence. Rules say what payment may never buy.

## What fit

| Need | In the model |
|---|---|
| Free, monthly and yearly plans; arranged plans | `Plan` with `Prices`, `SelfServe` |
| Counts with explicit unlimited | `LimitValue` with `Unlimited` |
| A daily or monthly allowance | `Limit.Period` |
| A ladder of levels (sync, branding, report and analytics depth, export style, API access) | `Feature.Levels`, ordered; unset is off |
| A switch (custom domain, selling, team management) | A feature without levels |
| A trial | `Plan.TrialDays` |
| A one-time purchase with a fixed price | `Product` |
| A seller-set price with a floor | `Product.SellerSet`, `MinAmountCents` |
| A marketplace share | `Marketplace.PlatformFeeBps` |
| Several applications behind one account | Grants per tier and application |
| Seats differing by tier | `TierGrant.Seats` or `UnlimitedSeats` |

## What did not fit, and the change

1. **The marketplace fee can differ by plan.** In shape B a seller on a higher
   plan may pay a lower share. A single `Marketplace.PlatformFeeBps` cannot say
   that. *Change:* `Plan.PlatformFeeBps` is an optional override of the
   marketplace default.
2. **Advertised "unlimited" with a fair-use policy.** In B, viewers are
   unlimited but subject to anti-abuse limits. `Unlimited` and a numeric
   fair-use threshold are different things. *Change:* `LimitValue.FairUse` may
   accompany `Unlimited`; then it means "advertised as unlimited, subject to the
   host's fair-use policy", and `Max` is ignored. With a `Max`, `FairUse` means
   crossing it warns and flags rather than refuses.
3. **A set of providers is not a ladder.** B's media hosting is "one provider",
   then "that one or an external one", then "external or integrated". Where the
   sets are nested this is a ladder, but it need not be. *Change:* the rule is
   that **levels are only for monotonic ladders**; a capability that is a set
   (media providers, payment methods) is several features without levels. A lint
   warns when a levelled feature's levels read as names of independent options.
4. **Applications were an undeclared string.** A grant named an application with
   nothing to check it against. *Change:* the document declares `Applications`,
   and grants refer to them. This closes the open question in the TRD.
5. **Seats had two homes.** A plan limit and a tier grant could both say how many
   members. *Change:* seats are in grants only. A product that wants a member
   cap as a limit expresses it as a grant; there is one place to enforce it.
6. **A one-time product that issues something.** In C a purchase issues a
   verified record. The model records a price, not what the purchase gives.
   *Decision:* what a product grants is the host's, identified by the product ID
   it receives when the purchase completes. The spec stays a description of the
   offer, not of fulfilment. Recorded as out of scope for the first release.
7. **Who may pay, and what payment may never buy.** C's rules (one adult pays,
   payment never buys an advantage) are policy. *Decision:* policy stays with the
   host. A feature may carry an advisory `Category` in a later release so a lint
   can flag a paid-only feature in a category the product has promised to keep
   free; it is not in the first release.
8. **Audience.** Pricing pages group plans for individuals, families, teams and
   institutions. *Decision:* deferred; a page can group by tier.

## Result

Two additions to the types (`Plan.PlatformFeeBps`, `Applications`), one change
of meaning (`FairUse` with `Unlimited`), and two rules (levels only for
ladders; seats only in grants). Three things are recorded as out of scope for
the first release: what a product grants, payer policy, and audience. The model
otherwise holds for all three shapes without extension.
