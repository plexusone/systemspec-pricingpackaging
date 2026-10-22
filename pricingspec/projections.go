package pricingspec

import "sort"

// PricingView is a Spec arranged for a page to render: plans in order, and a
// matrix of features and limits across them. It carries data, not display text.
type PricingView struct {
	App      string       `json:"app"`
	Currency string       `json:"currency"`
	Plans    []PlanColumn `json:"plans"`
	Features []FeatureRow `json:"features"`
	Limits   []LimitRow   `json:"limits"`
	Products []Product    `json:"products,omitempty"`
}

// PlanColumn is one plan in a pricing view.
type PlanColumn struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Tier      string  `json:"tier"`
	TierName  string  `json:"tierName"`
	TierRank  int     `json:"tierRank"`
	SelfServe bool    `json:"selfServe"`
	Free      bool    `json:"free"`
	Default   bool    `json:"default"`
	TrialDays int     `json:"trialDays,omitempty"`
	Prices    []Price `json:"prices,omitempty"`
}

// FeatureRow is one feature across the plans, in the same order as the plans.
type FeatureRow struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Levels      []string      `json:"levels,omitempty"`
	Cells       []FeatureCell `json:"cells"`
}

// FeatureCell is a plan's setting of a feature. Included is true for an enabled
// switch or a ladder at any level; Level is the ladder's level.
type FeatureCell struct {
	Plan     string `json:"plan"`
	Included bool   `json:"included"`
	Level    string `json:"level,omitempty"`
}

// LimitRow is one limit across the plans, in the same order as the plans.
type LimitRow struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Unit        string      `json:"unit,omitempty"`
	Period      Period      `json:"period"`
	Description string      `json:"description,omitempty"`
	Cells       []LimitCell `json:"cells"`
}

// LimitCell is a plan's allowance for a limit. Max is zero when the plan does
// not set it.
type LimitCell struct {
	Plan      string `json:"plan"`
	Max       int    `json:"max"`
	Unlimited bool   `json:"unlimited,omitempty"`
	FairUse   bool   `json:"fairUse,omitempty"`
}

// PricingView arranges the Spec for display. Plans are ordered by tier rank,
// then by monthly price (plans without one first), then by their order in the
// Spec.
func (s *Spec) PricingView() PricingView {
	tierName := map[string]string{}
	for _, t := range s.Tiers {
		tierName[t.ID] = t.Name
	}
	monthly := func(p Plan) int64 {
		for _, pr := range p.Prices {
			if pr.Interval == IntervalMonth {
				return pr.AmountCents
			}
		}
		return 0
	}
	plans := make([]Plan, len(s.Plans))
	copy(plans, s.Plans)
	sort.SliceStable(plans, func(i, j int) bool {
		ri, rj := s.TierRank(plans[i].Tier), s.TierRank(plans[j].Tier)
		if ri != rj {
			return ri < rj
		}
		return monthly(plans[i]) < monthly(plans[j])
	})

	v := PricingView{App: s.App, Currency: s.Currency, Products: s.Products}
	for _, p := range plans {
		v.Plans = append(v.Plans, PlanColumn{
			ID: p.ID, Name: p.Name, Tier: p.Tier, TierName: tierName[p.Tier], TierRank: s.TierRank(p.Tier),
			SelfServe: p.SelfServe, Free: len(p.Prices) == 0 && p.SelfServe, Default: p.ID == s.DefaultPlan,
			TrialDays: p.TrialDays, Prices: p.Prices,
		})
	}
	for _, f := range s.Features {
		row := FeatureRow{ID: f.ID, Name: f.Name, Description: f.Description, Levels: f.Levels}
		for _, p := range plans {
			row.Cells = append(row.Cells, FeatureCell{Plan: p.ID, Included: p.Has(f.ID), Level: s.levelOf(p, f.ID)})
		}
		v.Features = append(v.Features, row)
	}
	for _, l := range s.Limits {
		row := LimitRow{ID: l.ID, Name: l.Name, Unit: l.Unit, Period: l.Period, Description: l.Description}
		for _, p := range plans {
			cell := LimitCell{Plan: p.ID}
			for _, lv := range p.Limits {
				if lv.Limit == l.ID {
					cell.Max, cell.Unlimited, cell.FairUse = lv.Max, lv.Unlimited, lv.FairUse
				}
			}
			row.Cells = append(row.Cells, cell)
		}
		v.Limits = append(v.Limits, row)
	}
	return v
}

func (s *Spec) levelOf(p Plan, feature string) string {
	for _, fv := range p.Features {
		if fv.Feature == feature {
			return fv.Level
		}
	}
	return ""
}

// EntitlementMap is what each tier grants, in tier rank order.
type EntitlementMap struct {
	Tiers []TierEntitlements `json:"tiers"`
}

// TierEntitlements is one tier's grants.
type TierEntitlements struct {
	Tier   string  `json:"tier"`
	Name   string  `json:"name"`
	Rank   int     `json:"rank"`
	Grants []Grant `json:"grants"`
}

// Grant is access to one application.
type Grant struct {
	Application     string `json:"application"`
	ApplicationName string `json:"applicationName"`
	Seats           int    `json:"seats,omitempty"`
	UnlimitedSeats  bool   `json:"unlimitedSeats,omitempty"`
}

// EntitlementMap returns what each tier grants. Tiers with no grant are listed
// with an empty Grants.
func (s *Spec) EntitlementMap() EntitlementMap {
	appName := map[string]string{}
	for _, a := range s.Applications {
		appName[a.ID] = a.Name
	}
	tiers := make([]Tier, len(s.Tiers))
	copy(tiers, s.Tiers)
	sort.SliceStable(tiers, func(i, j int) bool { return tiers[i].Rank < tiers[j].Rank })

	var m EntitlementMap
	for _, t := range tiers {
		te := TierEntitlements{Tier: t.ID, Name: t.Name, Rank: t.Rank, Grants: []Grant{}}
		for _, g := range s.Entitlements {
			if g.Tier == t.ID {
				te.Grants = append(te.Grants, Grant{
					Application: g.Application, ApplicationName: appName[g.Application],
					Seats: g.Seats, UnlimitedSeats: g.UnlimitedSeats,
				})
			}
		}
		m.Tiers = append(m.Tiers, te)
	}
	return m
}
