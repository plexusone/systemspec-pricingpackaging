package pricingspec

import (
	"fmt"
	"strings"
)

// Lint returns the warnings (W001 to W006) for a Spec. Warnings are for review
// and never reject a Spec. Call Validate first: Lint assumes references resolve.
func (s *Spec) Lint() []Issue {
	var out []Issue
	warn := func(code, path, format string, a ...any) {
		out = append(out, Issue{Code: code, Severity: SeverityWarning, Path: path, Message: fmt.Sprintf(format, a...)})
	}

	// W001: the default plan's unlimited limits carry a fair-use note.
	for i, p := range s.Plans {
		if p.ID != s.DefaultPlan {
			continue
		}
		for j, lv := range p.Limits {
			if lv.Unlimited && !lv.FairUse {
				warn("W001", fmt.Sprintf("plans[%d].limits[%d]", i, j), "the default plan sets %q unlimited without fair use", lv.Limit)
			}
		}
	}

	// W002: a higher tier should not be cheaper than a lower one.
	rank := map[string]int{}
	for _, t := range s.Tiers {
		rank[t.ID] = t.Rank
	}
	for i, a := range s.Plans {
		for _, ap := range a.Prices {
			for j, b := range s.Plans {
				if rank[b.Tier] >= rank[a.Tier] {
					continue
				}
				for _, bp := range b.Prices {
					if bp.Interval == ap.Interval && ap.AmountCents < bp.AmountCents {
						warn("W002", fmt.Sprintf("plans[%d].prices", i),
							"plan %q (tier %q) costs less per %s than plan %q in a lower tier", a.ID, a.Tier, ap.Interval, s.Plans[j].ID)
					}
				}
			}
		}
	}

	// W003: unused features and limits.
	usedF, usedL := map[string]bool{}, map[string]bool{}
	for _, p := range s.Plans {
		for _, fv := range p.Features {
			usedF[fv.Feature] = true
		}
		for _, lv := range p.Limits {
			usedL[lv.Limit] = true
		}
	}
	for i, f := range s.Features {
		if !usedF[f.ID] {
			warn("W003", fmt.Sprintf("features[%d]", i), "feature %q is declared and no plan sets it", f.ID)
		}
	}
	for i, l := range s.Limits {
		if !usedL[l.ID] {
			warn("W003", fmt.Sprintf("limits[%d]", i), "limit %q is declared and no plan sets it", l.ID)
		}
	}

	// W004: a tier with plans and no grants.
	granted := map[string]bool{}
	for _, g := range s.Entitlements {
		granted[g.Tier] = true
	}
	planned := map[string]bool{}
	for _, p := range s.Plans {
		planned[p.Tier] = true
	}
	for i, t := range s.Tiers {
		if planned[t.ID] && !granted[t.ID] {
			warn("W004", fmt.Sprintf("tiers[%d]", i), "tier %q has plans and grants no application", t.ID)
		}
	}

	// W005: level names that read as a list of options.
	for i, f := range s.Features {
		for _, l := range f.Levels {
			if strings.ContainsAny(l, ",/") || strings.Contains(l, " or ") || strings.Contains(l, " and ") {
				warn("W005", fmt.Sprintf("features[%d].levels", i), "level %q reads as several options; a ladder's levels each include the one below, and independent options are separate switches", l)
				break
			}
		}
	}

	// W006: an annual price should be less than twelve monthly ones.
	for i, p := range s.Plans {
		var month, year int64
		for _, pr := range p.Prices {
			switch pr.Interval {
			case IntervalMonth:
				month = pr.AmountCents
			case IntervalYear:
				year = pr.AmountCents
			}
		}
		if month > 0 && year >= 12*month {
			warn("W006", fmt.Sprintf("plans[%d].prices", i), "the annual price is not less than twelve monthly payments")
		}
	}
	return out
}
