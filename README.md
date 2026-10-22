# systemspec-pricingpackaging

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/systemspec-pricingpackaging/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/systemspec-pricingpackaging/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/systemspec-pricingpackaging/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/systemspec-pricingpackaging/actions/workflows/go-lint.yaml
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/systemspec-pricingpackaging/blob/main/LICENSE

A Go-first, machine-readable description of a product's pricing and packaging:
the tiers of its ladder, the plans in them with their prices, the features
(a switch or an ordered level) and limits each plan sets, one-time products, a
marketplace fee, and the application access each tier grants.

The Spec is the source of truth. A pricing page, a limit check, a payment
provider's prices and an account's entitlements are computed from it.

- [SPEC.md](SPEC.md): the normative description of the model.
- `pricingspec`: the Go types, a loader, validation (`E001` to `E013`) and lint
  (`W001` to `W006`).
- `schema`: the JSON Schema generated from the types.

```go
spec, err := pricingspec.Load(f)
if err != nil {
	return err // a *pricingspec.ValidationError lists every error
}
for _, w := range spec.Lint() {
	fmt.Println(w) // warnings for review
}
```

Helpers (plan, feature and limit checks) and projections (a pricing view, an
entitlement map, a diff) are planned; see
[the roadmap](docs/specs/initiatives/INIT-SYSTEMSPECPRICINGPACKAGING-001/ROADMAP.md).

## Releases

[v0.1.0](docs/releases/v0.1.0.md) · [Changelog](CHANGELOG.md)

## Regenerating the schema

```sh
go generate ./schema/...
schemakit lint --property-case camelCase schema/pricing.schema.json
```
