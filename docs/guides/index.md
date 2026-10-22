# Concepts

A product's offer is described once, as data, and everything else is computed
from it.

| Term | Meaning |
|---|---|
| **Tier** | A position on the product's ladder, with a rank. "Free", "Plus", "Business". An account's entitlement carries its tier. |
| **Plan** | What an account can be on. A plan belongs to one tier, may have monthly and yearly prices, and sets features and limits. Free and arranged plans have no price. |
| **Feature** | Something a plan switches on, or sets to a level. A *switch* is on or off; a *ladder* has ordered levels such as `community`, `email`, `priority`. |
| **Limit** | Something a plan caps: a count, optionally per day or month. It can be unlimited, or fair use. |
| **Product** | A one-time purchase: a fixed price, or a price the seller sets with a floor. |
| **Marketplace** | The platform's share of sales by sellers, in basis points, optionally overridden per plan. |
| **Grant** | What a tier gives an account: access to an application, with a number of members or no limit. |
| **Default plan** | The free plan an unrecognised plan value resolves to, and the plan for an account when billing is not configured. |

## Access is a grant, allowances are looked up

Two different things follow from an account's plan:

- **Access.** Which applications the account may open and how many members it
  may have. This is the tier's *grant*, and it travels with the account's
  entitlement.
- **Allowances.** Which features are on, at what level, and how many boards,
  items or runs. These are *not* copied into the entitlement. The host stores the
  account's plan and asks the Spec at the moment of a check.

So a price or an allowance can change without reissuing anything, and two plans
in one tier can differ in allowances while giving the same access.

## Safe by construction

- An unknown plan is the default plan.
- An unset feature is off, and an unset limit is zero.
- "Unlimited" is its own field and never a zero or a missing value.
- The default plan must be free, and its tier must grant something.
- A Spec that breaks a rule is rejected with every error at once, each with a
  stable code (E001 to E013). Warnings (W001 to W006) are for review.

## What it is not

It is not a billing system: charging, subscriptions and webhooks are a runtime
library's work, which reads a Spec. It does not decide who may buy or what a
purchase issues; those stay with the product. It does not recommend prices.
