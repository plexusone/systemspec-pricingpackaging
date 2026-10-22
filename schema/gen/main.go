//go:build ignore

// Command gen generates the JSON Schema document from the pricingspec.Spec Go
// type, which is the source of truth. Run via `go generate ./schema/...` from
// the repository root. An optional argument names the directory to write to;
// the default is the current directory.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"

	"github.com/plexusone/systemspec-pricingpackaging/pricingspec"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	reflector := &jsonschema.Reflector{ExpandedStruct: true}
	schema := reflector.Reflect(&pricingspec.Spec{})
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}
	data = append(data, '\n')
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := os.WriteFile(filepath.Join(dir, "pricing.schema.json"), data, 0o600); err != nil {
		return fmt.Errorf("write pricing.schema.json: %w", err)
	}
	return nil
}
