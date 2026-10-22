# systemspec-pricingpackaging

A Go-first, machine-readable description of a product's pricing and packaging.
A **Spec** holds the tiers of a product's ladder, the plans in them with their
prices, the features and limits each plan sets, one-time products, a marketplace
fee, and the application access each tier grants. A pricing page, a limit check,
a payment provider's prices and an account's entitlements are computed from it.

- [Concepts](guides/index.md): tiers, plans, features, limits and grants.
- [Writing a spec](guides/writing-a-spec.md): the document, with a complete
  example.
- [Using a spec](guides/using-a-spec.md): loading, checking features and limits,
  rendering a pricing page, and seeing what changed.
- The normative description is [SPEC.md](https://github.com/plexusone/systemspec-pricingpackaging/blob/main/SPEC.md).
