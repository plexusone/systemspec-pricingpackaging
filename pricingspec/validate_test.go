package pricingspec_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/plexusone/systemspec-pricingpackaging/pricingspec"
)

func valid(t *testing.T) *pricingspec.Spec {
	t.Helper()
	data, err := os.ReadFile("../testdata/valid.json")
	if err != nil {
		t.Fatal(err)
	}
	s, err := pricingspec.Parse(data)
	if err != nil {
		t.Fatalf("testdata/valid.json: %v", err)
	}
	return s
}

func codes(issues []pricingspec.Issue) []string {
	var out []string
	for _, i := range issues {
		out = append(out, i.Code)
	}
	return out
}

func TestValidSpecHasNoErrorsAndNoWarnings(t *testing.T) {
	s := valid(t)
	if is := s.Validate(); len(is) != 0 {
		t.Fatalf("errors: %v", is)
	}
	if is := s.Lint(); len(is) != 0 {
		t.Fatalf("warnings: %v", is)
	}
}

func TestRoundTrip(t *testing.T) {
	s := valid(t)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	back, err := pricingspec.Parse(data)
	if err != nil {
		t.Fatalf("parse marshalled spec: %v", err)
	}
	again, _ := json.Marshal(back)
	if !bytes.Equal(data, again) {
		t.Fatalf("round trip changed the document:\n%s\n%s", data, again)
	}
}

func TestErrorRules(t *testing.T) {
	one := 1
	bad := 10001
	tests := []struct {
		name   string
		mutate func(s *pricingspec.Spec)
		want   string
	}{
		{"E001 format", func(s *pricingspec.Spec) { s.FormatVersion = "2" }, "E001"},
		{"E002 app", func(s *pricingspec.Spec) { s.App = "Bad App" }, "E002"},
		{"E002 currency", func(s *pricingspec.Spec) { s.Currency = "USD" }, "E002"},
		{"E003 malformed id", func(s *pricingspec.Spec) { s.Tiers[0].ID = "Free!" }, "E003"},
		{"E003 duplicate id", func(s *pricingspec.Spec) { s.Plans[1].ID = "free" }, "E003"},
		{"E004 zero rank", func(s *pricingspec.Spec) { s.Tiers[0].Rank = 0 }, "E004"},
		{"E004 duplicate rank", func(s *pricingspec.Spec) { s.Tiers[1].Rank = 1 }, "E004"},
		{"E005 undeclared tier", func(s *pricingspec.Spec) { s.Plans[1].Tier = "gold" }, "E005"},
		{"E006 undeclared feature", func(s *pricingspec.Spec) { s.Plans[0].Features[0].Feature = "nope" }, "E006"},
		{"E006 set twice", func(s *pricingspec.Spec) {
			s.Plans[0].Features = append(s.Plans[0].Features, s.Plans[0].Features[0])
		}, "E006"},
		{"E006 switch with level", func(s *pricingspec.Spec) {
			s.Plans[1].Features[1] = pricingspec.FeatureValue{Feature: "custom_domain", Level: "on"}
		}, "E006"},
		{"E006 ladder enabled", func(s *pricingspec.Spec) {
			s.Plans[0].Features[0] = pricingspec.FeatureValue{Feature: "refresh", Enabled: true}
		}, "E006"},
		{"E006 unknown level", func(s *pricingspec.Spec) { s.Plans[0].Features[0].Level = "weekly" }, "E006"},
		{"E007 one level", func(s *pricingspec.Spec) { s.Features[0].Levels = []string{"daily"} }, "E007"},
		{"E007 duplicate level", func(s *pricingspec.Spec) { s.Features[0].Levels = []string{"a", "a"} }, "E007"},
		{"E008 undeclared limit", func(s *pricingspec.Spec) { s.Plans[0].Limits[0].Limit = "nope" }, "E008"},
		{"E008 negative", func(s *pricingspec.Spec) { s.Plans[0].Limits[0].Max = -1 }, "E008"},
		{"E008 unlimited with max", func(s *pricingspec.Spec) {
			s.Plans[2].Limits[0] = pricingspec.LimitValue{Limit: "projects", Unlimited: true, Max: 5}
		}, "E008"},
		{"E009 zero price", func(s *pricingspec.Spec) { s.Plans[1].Prices[0].AmountCents = 0 }, "E009"},
		{"E009 same interval twice", func(s *pricingspec.Spec) { s.Plans[1].Prices[1].Interval = pricingspec.IntervalMonth }, "E009"},
		{"E009 bad interval", func(s *pricingspec.Spec) { s.Plans[1].Prices[0].Interval = "week" }, "E009"},
		{"E009 trial without prices", func(s *pricingspec.Spec) { s.Plans[0].TrialDays = 7 }, "E009"},
		{"E009 free self-serve plan other than default", func(s *pricingspec.Spec) { s.Plans[1].Prices = nil; s.Plans[1].TrialDays = 0 }, "E009"},
		{"E010 default missing", func(s *pricingspec.Spec) { s.DefaultPlan = "nope" }, "E010"},
		{"E010 default priced", func(s *pricingspec.Spec) { s.DefaultPlan = "pro" }, "E010"},
		{"E010 default tier without grant", func(s *pricingspec.Spec) { s.Entitlements = s.Entitlements[1:] }, "E010"},
		{"E011 fixed without amount", func(s *pricingspec.Spec) { s.Products[0].AmountCents = 0 }, "E011"},
		{"E011 seller-set without minimum", func(s *pricingspec.Spec) { s.Products[1].MinAmountCents = 0 }, "E011"},
		{"E011 seller-set with amount", func(s *pricingspec.Spec) { s.Products[1].AmountCents = 500 }, "E011"},
		{"E011 fixed with minimum", func(s *pricingspec.Spec) { s.Products[0].MinAmountCents = 5 }, "E011"},
		{"E012 marketplace fee", func(s *pricingspec.Spec) { s.Marketplace.PlatformFeeBps = bad }, "E012"},
		{"E012 plan fee", func(s *pricingspec.Spec) { s.Plans[1].PlatformFeeBps = &bad }, "E012"},
		{"E013 undeclared tier", func(s *pricingspec.Spec) { s.Entitlements[0].Tier = "gold" }, "E013"},
		{"E013 undeclared application", func(s *pricingspec.Spec) { s.Entitlements[0].Application = "nope" }, "E013"},
		{"E013 zero seats", func(s *pricingspec.Spec) { s.Entitlements[0].Seats = 0 }, "E013"},
		{"E013 both seats", func(s *pricingspec.Spec) {
			s.Entitlements[0].UnlimitedSeats = true
			s.Entitlements[0].Seats = one
		}, "E013"},
		{"E013 duplicate grant", func(s *pricingspec.Spec) { s.Entitlements = append(s.Entitlements, s.Entitlements[0]) }, "E013"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := valid(t)
			tc.mutate(s)
			got := codes(s.Validate())
			if !slices.Contains(got, tc.want) {
				t.Fatalf("want %s among %v", tc.want, got)
			}
		})
	}
}

func TestLintRules(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(s *pricingspec.Spec)
		want   string
	}{
		{"W001 default unlimited without fair use", func(s *pricingspec.Spec) {
			s.Plans[0].Limits[0] = pricingspec.LimitValue{Limit: "projects", Unlimited: true}
		}, "W001"},
		{"W002 higher tier cheaper", func(s *pricingspec.Spec) {
			s.Plans = append(s.Plans, pricingspec.Plan{ID: "plus", Name: "Plus", Tier: "enterprise", SelfServe: true,
				Prices: []pricingspec.Price{{Interval: pricingspec.IntervalMonth, AmountCents: 100}}})
		}, "W002"},
		{"W003 unused feature", func(s *pricingspec.Spec) {
			s.Features = append(s.Features, pricingspec.Feature{ID: "extra", Name: "Extra"})
		}, "W003"},
		{"W003 unused limit", func(s *pricingspec.Spec) {
			s.Limits = append(s.Limits, pricingspec.Limit{ID: "extra", Name: "Extra", Period: pricingspec.PeriodNone})
		}, "W003"},
		{"W004 tier without grants", func(s *pricingspec.Spec) {
			s.Tiers = append(s.Tiers, pricingspec.Tier{ID: "team", Name: "Team", Rank: 4})
			s.Plans = append(s.Plans, pricingspec.Plan{ID: "team", Name: "Team", Tier: "team", SelfServe: false})
		}, "W004"},
		{"W005 levels read as a list", func(s *pricingspec.Spec) { s.Features[0].Levels = []string{"daily", "hourly", "hourly or weekly"} }, "W005"},
		{"W006 annual not cheaper", func(s *pricingspec.Spec) { s.Plans[1].Prices[1].AmountCents = 22800 }, "W006"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := valid(t)
			tc.mutate(s)
			if errs := s.Validate(); len(errs) != 0 {
				t.Fatalf("the mutation should leave the spec valid: %v", errs)
			}
			got := codes(s.Lint())
			if !slices.Contains(got, tc.want) {
				t.Fatalf("want %s among %v", tc.want, got)
			}
		})
	}
}

func TestParseRejects(t *testing.T) {
	good, _ := os.ReadFile("../testdata/valid.json")
	t.Run("unknown field", func(t *testing.T) {
		bad := strings.Replace(string(good), `"currency": "usd",`, `"currency": "usd", "surprise": true,`, 1)
		if _, err := pricingspec.Parse([]byte(bad)); err == nil || !strings.Contains(err.Error(), "surprise") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("unknown format", func(t *testing.T) {
		bad := strings.Replace(string(good), `"formatVersion": "1"`, `"formatVersion": "9"`, 1)
		if _, err := pricingspec.Parse([]byte(bad)); !errors.Is(err, pricingspec.ErrUnknownFormat) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("not json", func(t *testing.T) {
		if _, err := pricingspec.Parse([]byte("nope")); err == nil {
			t.Fatal("accepted")
		}
	})
	t.Run("validation error lists every error", func(t *testing.T) {
		bad := strings.Replace(string(good), `"app": "example-suite"`, `"app": "Bad"`, 1)
		bad = strings.Replace(bad, `"currency": "usd"`, `"currency": "US"`, 1)
		_, err := pricingspec.Parse([]byte(bad))
		var ve *pricingspec.ValidationError
		if !errors.As(err, &ve) || len(ve.Issues) < 2 {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestLoad(t *testing.T) {
	f, err := os.Open("../testdata/valid.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := pricingspec.Load(f); err != nil {
		t.Fatal(err)
	}
}
