package pricingspec_test

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/oscompat/testutil"

	"github.com/plexusone/systemspec-pricingpackaging/pricingspec"
)

var update = flag.Bool("update", false, "rewrite the golden files")

func golden(t *testing.T, name string, v any) {
	t.Helper()
	got, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	testutil.AssertGolden(t, filepath.Join("..", "testdata", "golden", name), got, *update)
}

func load(t *testing.T, name string) *pricingspec.Spec {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	s, err := pricingspec.Parse(data)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return s
}

func TestPricingViewGolden(t *testing.T) {
	golden(t, "pricing_view.json", valid(t).PricingView())
}

func TestPricingViewOrderAndAlignment(t *testing.T) {
	v := valid(t).PricingView()
	var ids []string
	for _, p := range v.Plans {
		ids = append(ids, p.ID)
	}
	if strings.Join(ids, ",") != "free,pro,enterprise" {
		t.Fatalf("plan order = %v", ids)
	}
	if !v.Plans[0].Free || !v.Plans[0].Default || v.Plans[1].Free || v.Plans[2].SelfServe {
		t.Fatalf("flags: %+v", v.Plans)
	}
	for _, row := range v.Features {
		if len(row.Cells) != len(v.Plans) {
			t.Fatalf("feature %s has %d cells for %d plans", row.ID, len(row.Cells), len(v.Plans))
		}
		for i, c := range row.Cells {
			if c.Plan != v.Plans[i].ID {
				t.Fatalf("feature %s cell %d is for %s, want %s", row.ID, i, c.Plan, v.Plans[i].ID)
			}
		}
	}
	for _, row := range v.Limits {
		if len(row.Cells) != len(v.Plans) {
			t.Fatalf("limit %s has %d cells", row.ID, len(row.Cells))
		}
	}
}

func TestEntitlementMapGolden(t *testing.T) {
	golden(t, "entitlement_map.json", valid(t).EntitlementMap())
}

func TestEntitlementMapListsEveryTierInRankOrder(t *testing.T) {
	s := valid(t)
	s.Tiers[0], s.Tiers[2] = s.Tiers[2], s.Tiers[0] // declared out of order
	m := s.EntitlementMap()
	if len(m.Tiers) != 3 || m.Tiers[0].Tier != "free" || m.Tiers[2].Tier != "enterprise" {
		t.Fatalf("tiers = %+v", m.Tiers)
	}
	if m.Tiers[1].Grants[0].ApplicationName == "" {
		t.Fatalf("no application name: %+v", m.Tiers[1])
	}
}

func TestDiffGolden(t *testing.T) {
	golden(t, "diff.json", pricingspec.Diff(load(t, "valid.json"), load(t, "valid_v2.json")))
}

func TestDiffMarksWhatAffectsExistingCustomers(t *testing.T) {
	changes := pricingspec.Diff(load(t, "valid.json"), load(t, "valid_v2.json"))
	affects := map[string]bool{}
	for _, c := range changes {
		key := c.Subject + ": " + c.Detail
		affects[key] = c.AffectsExisting
	}
	for _, tc := range []struct {
		contains string
		want     bool
	}{
		{"plan pro: month price 1900 to 2400", true},                     // price up
		{"plan pro: limit projects 20 to 10", true},                      // allowance down
		{"plan enterprise: feature refresh level hourly to daily", true}, // level down
		{"grant pro/studio: 5 seats", true},                              // grant removed
		{"grant pro/reports: seats 5 seats to 10 seats", false},          // more seats
		{"plan free: feature custom_domain", false},                      // feature added
		{"plan team: tier pro", false},                                   // plan added
		{"marketplace: platform fee 1000 to 1200", true},                 // fee up
	} {
		got, ok := affects[tc.contains]
		if !ok {
			t.Errorf("no change %q in %v", tc.contains, changes)
			continue
		}
		if got != tc.want {
			t.Errorf("%q affects existing = %t, want %t", tc.contains, got, tc.want)
		}
	}
}

func TestDiffOfEqualSpecsIsEmpty(t *testing.T) {
	if c := pricingspec.Diff(load(t, "valid.json"), load(t, "valid.json")); len(c) != 0 {
		t.Fatalf("changes = %v", c)
	}
}

func TestDiffIsStable(t *testing.T) {
	a := pricingspec.Diff(load(t, "valid.json"), load(t, "valid_v2.json"))
	b := pricingspec.Diff(load(t, "valid.json"), load(t, "valid_v2.json"))
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Fatal("two diffs of the same specs differ")
	}
}

func TestDiffRemovalsAndLoweredLimits(t *testing.T) {
	old := valid(t)
	n := valid(t)
	n.Plans = n.Plans[:2] // enterprise removed
	n.Plans[0].Limits[0] = pricingspec.LimitValue{Limit: "projects", Unlimited: true}
	n.Plans[1].Limits[0] = pricingspec.LimitValue{Limit: "projects", Max: 5}
	n.Products = n.Products[:1]
	n.Marketplace = nil
	var got []string
	for _, c := range pricingspec.Diff(old, n) {
		got = append(got, c.String())
	}
	all := strings.Join(got, "\n")
	for _, want := range []string{
		"removed plan enterprise",
		"changed plan free: limit projects 2 to unlimited",
		"changed plan pro: limit projects 20 to 5 [affects existing customers]",
		"removed product template",
		"changed marketplace: platform fee 1000 to default",
	} {
		if !strings.Contains(all, want) {
			t.Errorf("missing %q in\n%s", want, all)
		}
	}
}
