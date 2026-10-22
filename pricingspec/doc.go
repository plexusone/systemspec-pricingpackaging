// Package pricingspec is a machine-readable description of a product's pricing
// and packaging: the tiers of its ladder, the plans in them with their prices,
// the features and limits each plan sets, its one-time products, an optional
// marketplace fee, and the grants each tier gives an account.
//
// A Spec is the source of truth. A pricing page, a limit check, a payment
// provider's prices and an account's entitlements are computed from it. The
// normative description is SPEC.md at the root of the repository; the types in
// this package are its Go form, and the JSON Schema in the schema package is
// generated from them.
//
// Nothing here depends on a payment provider or an application framework.
package pricingspec
