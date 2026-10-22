package pricingspec_test

import (
	"testing"

	"github.com/plexusone/systemspec-pricingpackaging/pricingspec"
)

func TestPlanOrDefault(t *testing.T) {
	s := valid(t)
	if got := s.PlanOrDefault("pro").ID; got != "pro" {
		t.Fatalf("pro: %s", got)
	}
	for _, id := range []string{"", "nope", "PRO", "enterprise2"} {
		if got := s.PlanOrDefault(id).ID; got != "free" {
			t.Errorf("%q resolved to %q, want the default", id, got)
		}
	}
	if _, ok := s.Plan("nope"); ok {
		t.Fatal("Plan found an unknown plan")
	}
}

func TestTierRank(t *testing.T) {
	s := valid(t)
	if s.TierRank("pro") != 2 || s.TierRank("nope") != 0 {
		t.Fatalf("ranks: %d %d", s.TierRank("pro"), s.TierRank("nope"))
	}
}

func TestHasAndLevels(t *testing.T) {
	s := valid(t)
	tests := []struct {
		plan, feature string
		has           bool
		level         string
		atLeast       map[string]bool
	}{
		{"free", "refresh", true, "daily", map[string]bool{"daily": true, "hourly": false, "weekly": false}},
		{"pro", "refresh", true, "hourly", map[string]bool{"daily": true, "hourly": true}},
		{"free", "custom_domain", false, "", map[string]bool{"on": false}},
		{"pro", "custom_domain", true, "", map[string]bool{"on": false}},     // a switch has no levels
		{"nope", "refresh", true, "daily", map[string]bool{"hourly": false}}, // unknown plan: default
		{"pro", "unknown_feature", false, "", map[string]bool{"daily": false}},
	}
	for _, tc := range tests {
		if got := s.Has(tc.plan, tc.feature); got != tc.has {
			t.Errorf("Has(%s,%s) = %t", tc.plan, tc.feature, got)
		}
		if got := s.Level(tc.plan, tc.feature); got != tc.level {
			t.Errorf("Level(%s,%s) = %q", tc.plan, tc.feature, got)
		}
		for lvl, want := range tc.atLeast {
			if got := s.AtLeast(tc.plan, tc.feature, lvl); got != want {
				t.Errorf("AtLeast(%s,%s,%s) = %t", tc.plan, tc.feature, lvl, got)
			}
		}
	}
}

func TestCheck(t *testing.T) {
	s := valid(t)
	tests := []struct {
		name  string
		plan  string
		limit string
		after int
		want  pricingspec.Decision
	}{
		{"within", "free", "projects", 2, pricingspec.Allowed},
		{"at the maximum is allowed", "free", "projects", 2, pricingspec.Allowed},
		{"over the maximum refuses", "free", "projects", 3, pricingspec.Refuse},
		{"zero usage", "free", "projects", 0, pricingspec.Allowed},
		{"over a fair-use maximum warns", "free", "exports", 21, pricingspec.Warn},
		{"within a fair-use maximum", "free", "exports", 20, pricingspec.Allowed},
		{"unlimited", "enterprise", "projects", 1_000_000, pricingspec.Allowed},
		{"unlimited with fair use is allowed", "pro", "exports", 1_000_000, pricingspec.Allowed},
		{"an unknown plan is the default", "nope", "projects", 3, pricingspec.Refuse},
		{"a limit the plan does not set is zero", "free", "nope", 1, pricingspec.Refuse},
		{"a limit the plan does not set allows nothing used", "free", "nope", 0, pricingspec.Allowed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := s.Check(tc.plan, tc.limit, tc.after); got != tc.want {
				t.Fatalf("Check(%s,%s,%d) = %s, want %s", tc.plan, tc.limit, tc.after, got, tc.want)
			}
		})
	}
}

func TestAllowanceAndGrants(t *testing.T) {
	s := valid(t)
	if max, unl, fair := s.Allowance("free", "exports"); max != 20 || unl || !fair {
		t.Fatalf("allowance = %d %t %t", max, unl, fair)
	}
	if g := s.GrantsFor("pro"); len(g) != 2 || g[0].Application != "reports" {
		t.Fatalf("grants = %+v", g)
	}
	if g := s.GrantsFor("nope"); len(g) != 1 || g[0].Tier != "free" {
		t.Fatalf("unknown plan grants = %+v", g)
	}
}
