package pricingspec

import "slices"

// Decision is the answer to a limit check.
type Decision string

// Decisions.
const (
	// Allowed means the action is within the allowance.
	Allowed Decision = "allowed"
	// Warn means the allowance is exceeded but marked fair use: allow the
	// action and flag the account for review.
	Warn Decision = "warn"
	// Refuse means the allowance is exceeded.
	Refuse Decision = "refuse"
)

// Plan returns the plan with the given ID.
func (s *Spec) Plan(id string) (Plan, bool) {
	i := slices.IndexFunc(s.Plans, func(p Plan) bool { return p.ID == id })
	if i < 0 {
		return Plan{}, false
	}
	return s.Plans[i], true
}

// PlanOrDefault returns the plan with the given ID, or the default plan when
// the ID is not a plan, so an unrecognised value never grants more. It returns
// the zero Plan only for a Spec with no default plan, which Validate rejects.
func (s *Spec) PlanOrDefault(id string) Plan {
	if p, ok := s.Plan(id); ok {
		return p
	}
	p, _ := s.Plan(s.DefaultPlan)
	return p
}

// TierRank returns the rank of a tier, or zero when it is not declared.
func (s *Spec) TierRank(tier string) int {
	for _, t := range s.Tiers {
		if t.ID == tier {
			return t.Rank
		}
	}
	return 0
}

// Has reports whether the plan switches the feature on, or sets it to any
// level. It does not resolve an unknown plan; see Spec.Has.
func (p Plan) Has(feature string) bool {
	for _, fv := range p.Features {
		if fv.Feature == feature {
			return fv.Enabled || fv.Level != ""
		}
	}
	return false
}

// Has reports whether the plan (or the default plan, for an unknown ID) has the
// feature on, or at any level.
func (s *Spec) Has(planID, feature string) bool {
	return s.PlanOrDefault(planID).Has(feature)
}

// Level returns the level the plan sets for a ladder feature, or "" when the
// plan does not set it, or the feature is a switch or unknown.
func (s *Spec) Level(planID, feature string) string {
	for _, fv := range s.PlanOrDefault(planID).Features {
		if fv.Feature == feature {
			return fv.Level
		}
	}
	return ""
}

// AtLeast reports whether the plan sets a ladder feature at or above level. It
// is false for an unset ladder, a switch, an unknown feature or an unknown
// level.
func (s *Spec) AtLeast(planID, feature, level string) bool {
	var levels []string
	for _, f := range s.Features {
		if f.ID == feature {
			levels = f.Levels
		}
	}
	want := slices.Index(levels, level)
	have := slices.Index(levels, s.Level(planID, feature))
	return want >= 0 && have >= want
}

// Allowance returns the plan's allowance for a limit. A limit the plan does not
// set has a maximum of zero. unlimited is true when the plan sets it unlimited,
// and fairUse when it is marked fair use.
func (s *Spec) Allowance(planID, limit string) (max int, unlimited, fairUse bool) {
	for _, lv := range s.PlanOrDefault(planID).Limits {
		if lv.Limit == limit {
			return lv.Max, lv.Unlimited, lv.FairUse
		}
	}
	return 0, false, false
}

// Check answers whether an action may proceed. after is the usage after the
// action: a plan with a maximum of 3 and 2 items in use checks adding a third as
// 3. The result is Allowed when the limit is unlimited or after is within the
// maximum; over the maximum it is Warn for a fair-use limit and Refuse
// otherwise. A limit the plan does not set has a maximum of zero. An unlimited
// fair-use limit is always Allowed: the Spec sets no threshold and the host
// applies its own policy.
func (s *Spec) Check(planID, limit string, after int) Decision {
	max, unlimited, fairUse := s.Allowance(planID, limit)
	switch {
	case unlimited, after <= max:
		return Allowed
	case fairUse:
		return Warn
	}
	return Refuse
}

// GrantsFor returns what the plan's tier grants an account, in the Spec's order.
// An unknown plan resolves to the default plan.
func (s *Spec) GrantsFor(planID string) []TierGrant {
	tier := s.PlanOrDefault(planID).Tier
	var out []TierGrant
	for _, g := range s.Entitlements {
		if g.Tier == tier {
			out = append(out, g)
		}
	}
	return out
}
