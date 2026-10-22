package pricingspec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/plexusone/systemspec-pricingpackaging/pricingspec"
)

// Every example in examples/ is a valid spec with no warnings, so the
// documentation never shows an offer the validator would reject.
func TestExamplesAreValidAndClean(t *testing.T) {
	files, err := filepath.Glob("../examples/*.json")
	if err != nil || len(files) < 2 {
		t.Fatalf("examples: %v (%d files)", err, len(files))
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			s, err := pricingspec.Parse(data)
			if err != nil {
				t.Fatal(err)
			}
			if w := s.Lint(); len(w) != 0 {
				t.Fatalf("warnings: %v", w)
			}
		})
	}
}

func TestMarketplaceExampleResolves(t *testing.T) {
	data, _ := os.ReadFile("../examples/marketplace.json")
	s, err := pricingspec.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if s.Has("browser", "sell") || !s.Has("seller", "sell") {
		t.Fatal("only sellers sell")
	}
	if !s.AtLeast("pro_seller", "storefront", "custom") || s.AtLeast("seller", "storefront", "custom") {
		t.Fatal("storefront levels")
	}
	if s.Check("browser", "listings", 1) != pricingspec.Refuse {
		t.Fatal("a browser may not list")
	}
	if s.Check("pro_seller", "listings", 10_000) != pricingspec.Allowed {
		t.Fatal("a pro seller has unlimited listings")
	}
}
