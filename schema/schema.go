// Package schema embeds the JSON Schema document generated from the
// pricingspec.Spec Go type. The Go structs are the source of truth; this schema
// is a generated, embeddable artifact for non-Go consumers and for validating
// hand-authored pricing documents.
package schema

import _ "embed"

//go:generate go run gen/main.go

// SpecJSON is the JSON Schema for a pricingspec.Spec document.
//
//go:embed pricing.schema.json
var SpecJSON []byte
