package pricingspec

import (
	"fmt"
	"slices"
	"sort"
)

// ChangeKind says what happened to the thing that changed.
type ChangeKind string

// Change kinds.
const (
	Added   ChangeKind = "added"
	Removed ChangeKind = "removed"
	Changed ChangeKind = "changed"
)

// Change is one difference between two Specs.
type Change struct {
	Kind ChangeKind `json:"kind"`
	// Subject is what changed, for example "plan pro" or "grant free/reports".
	Subject string `json:"subject"`
	Detail  string `json:"detail"`
	// AffectsExisting is true when the change worsens the offer for accounts
	// already on it: a higher price, a lower allowance, a removed feature, a
	// removed grant, fewer seats, a lowered level, or a removed plan.
	AffectsExisting bool `json:"affectsExisting,omitempty"`
}

func (c Change) String() string {
	mark := ""
	if c.AffectsExisting {
		mark = " [affects existing customers]"
	}
	return fmt.Sprintf("%s %s: %s%s", c.Kind, c.Subject, c.Detail, mark)
}

// Diff compares two Specs and lists what differs, in a stable order.
// Additions and improvements are listed but do not set AffectsExisting.
func Diff(oldSpec, newSpec *Spec) []Change {
	d := &differ{o: oldSpec, n: newSpec}
	d.run()
	return d.out
}

type differ struct {
	o, n *Spec
	out  []Change
}

func (d *differ) add(kind ChangeKind, subject, detail string, affects bool) {
	d.out = append(d.out, Change{Kind: kind, Subject: subject, Detail: detail, AffectsExisting: affects})
}

// keys returns the union of two ID sets, sorted.
func keys[T any](a, b map[string]T) []string {
	set := map[string]bool{}
	for k := range a {
		set[k] = true
	}
	for k := range b {
		set[k] = true
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func index[T any](xs []T, id func(T) string) map[string]T {
	m := make(map[string]T, len(xs))
	for _, x := range xs {
		m[id(x)] = x
	}
	return m
}

func (d *differ) run() {
	if d.o.Currency != d.n.Currency {
		d.add(Changed, "currency", fmt.Sprintf("%s to %s", d.o.Currency, d.n.Currency), true)
	}
	if d.o.DefaultPlan != d.n.DefaultPlan {
		d.add(Changed, "default plan", fmt.Sprintf("%s to %s", d.o.DefaultPlan, d.n.DefaultPlan), false)
	}
	d.tiers()
	d.features()
	d.limits()
	d.plans()
	d.products()
	d.marketplace()
	d.grants()
}

func (d *differ) tiers() {
	ot := index(d.o.Tiers, func(t Tier) string { return t.ID })
	nt := index(d.n.Tiers, func(t Tier) string { return t.ID })
	for _, k := range keys(ot, nt) {
		o, inO := ot[k]
		n, inN := nt[k]
		switch {
		case !inO:
			d.add(Added, "tier "+k, fmt.Sprintf("rank %d", n.Rank), false)
		case !inN:
			d.add(Removed, "tier "+k, "", true)
		case o.Rank != n.Rank:
			d.add(Changed, "tier "+k, fmt.Sprintf("rank %d to %d", o.Rank, n.Rank), false)
		}
	}
}

func (d *differ) features() {
	of := index(d.o.Features, func(f Feature) string { return f.ID })
	nf := index(d.n.Features, func(f Feature) string { return f.ID })
	for _, k := range keys(of, nf) {
		o, inO := of[k]
		n, inN := nf[k]
		switch {
		case !inO:
			d.add(Added, "feature "+k, "", false)
		case !inN:
			d.add(Removed, "feature "+k, "", true)
		case !slices.Equal(o.Levels, n.Levels):
			d.add(Changed, "feature "+k, fmt.Sprintf("levels %v to %v", o.Levels, n.Levels), false)
		}
	}
}

func (d *differ) limits() {
	ol := index(d.o.Limits, func(l Limit) string { return l.ID })
	nl := index(d.n.Limits, func(l Limit) string { return l.ID })
	for _, k := range keys(ol, nl) {
		o, inO := ol[k]
		n, inN := nl[k]
		switch {
		case !inO:
			d.add(Added, "limit "+k, "", false)
		case !inN:
			d.add(Removed, "limit "+k, "", false)
		case o.Period != n.Period:
			d.add(Changed, "limit "+k, fmt.Sprintf("period %s to %s", o.Period, n.Period), true)
		}
	}
}

func (d *differ) plans() {
	op := index(d.o.Plans, func(p Plan) string { return p.ID })
	np := index(d.n.Plans, func(p Plan) string { return p.ID })
	for _, k := range keys(op, np) {
		o, inO := op[k]
		n, inN := np[k]
		subj := "plan " + k
		switch {
		case !inO:
			d.add(Added, subj, "tier "+n.Tier, false)
			continue
		case !inN:
			d.add(Removed, subj, "", true)
			continue
		}
		if o.Tier != n.Tier {
			d.add(Changed, subj, fmt.Sprintf("tier %s to %s", o.Tier, n.Tier), d.n.TierRank(n.Tier) < d.o.TierRank(o.Tier))
		}
		if o.SelfServe != n.SelfServe {
			d.add(Changed, subj, fmt.Sprintf("selfServe %t to %t", o.SelfServe, n.SelfServe), false)
		}
		if o.TrialDays != n.TrialDays {
			d.add(Changed, subj, fmt.Sprintf("trial %d to %d days", o.TrialDays, n.TrialDays), false)
		}
		d.prices(subj, o, n)
		d.planFeatures(subj, o, n)
		d.planLimits(subj, o, n)
		if !equalIntPtr(o.PlatformFeeBps, n.PlatformFeeBps) {
			d.add(Changed, subj, fmt.Sprintf("platform fee %s to %s", showPtr(o.PlatformFeeBps), showPtr(n.PlatformFeeBps)), feeRaised(o.PlatformFeeBps, n.PlatformFeeBps, d.o.Marketplace, d.n.Marketplace))
		}
	}
}

func (d *differ) prices(subj string, o, n Plan) {
	op := map[Interval]int64{}
	np := map[Interval]int64{}
	for _, p := range o.Prices {
		op[p.Interval] = p.AmountCents
	}
	for _, p := range n.Prices {
		np[p.Interval] = p.AmountCents
	}
	for _, iv := range []Interval{IntervalMonth, IntervalYear} {
		ov, inO := op[iv]
		nv, inN := np[iv]
		switch {
		case !inO && inN:
			d.add(Added, subj, fmt.Sprintf("%s price %d", iv, nv), false)
		case inO && !inN:
			d.add(Removed, subj, fmt.Sprintf("%s price %d", iv, ov), true)
		case ov != nv:
			d.add(Changed, subj, fmt.Sprintf("%s price %d to %d", iv, ov, nv), nv > ov)
		}
	}
}

func (d *differ) planFeatures(subj string, o, n Plan) {
	of := index(o.Features, func(f FeatureValue) string { return f.Feature })
	nf := index(n.Features, func(f FeatureValue) string { return f.Feature })
	for _, k := range keys(of, nf) {
		ov, inO := of[k]
		nv, inN := nf[k]
		switch {
		case !inO:
			d.add(Added, subj, "feature "+k, false)
		case !inN:
			d.add(Removed, subj, "feature "+k, true)
		case ov.Enabled != nv.Enabled:
			d.add(Changed, subj, fmt.Sprintf("feature %s enabled %t to %t", k, ov.Enabled, nv.Enabled), ov.Enabled && !nv.Enabled)
		case ov.Level != nv.Level:
			d.add(Changed, subj, fmt.Sprintf("feature %s level %s to %s", k, ov.Level, nv.Level), d.levelLowered(k, ov.Level, nv.Level))
		}
	}
}

// levelLowered reports whether going from one level of a feature to another
// takes something away. It checks the new Spec's order, falling back to the
// old one.
func (d *differ) levelLowered(feature, from, to string) bool {
	for _, spec := range []*Spec{d.n, d.o} {
		for _, f := range spec.Features {
			if f.ID == feature {
				fi, ti := slices.Index(f.Levels, from), slices.Index(f.Levels, to)
				if fi >= 0 && ti >= 0 {
					return ti < fi
				}
			}
		}
	}
	return false
}

func (d *differ) planLimits(subj string, o, n Plan) {
	ol := index(o.Limits, func(l LimitValue) string { return l.Limit })
	nl := index(n.Limits, func(l LimitValue) string { return l.Limit })
	for _, k := range keys(ol, nl) {
		ov, inO := ol[k]
		nv, inN := nl[k]
		switch {
		case !inO:
			d.add(Added, subj, "limit "+k+" "+showLimit(nv), false)
		case !inN:
			d.add(Removed, subj, "limit "+k+" "+showLimit(ov), true)
		case ov != nv:
			d.add(Changed, subj, fmt.Sprintf("limit %s %s to %s", k, showLimit(ov), showLimit(nv)), limitLowered(ov, nv))
		}
	}
}

func limitLowered(o, n LimitValue) bool {
	switch {
	case o.Unlimited && !n.Unlimited:
		return true
	case !o.Unlimited && n.Unlimited:
		return false
	case o.Unlimited && n.Unlimited:
		return false
	}
	return n.Max < o.Max
}

func showLimit(l LimitValue) string {
	s := fmt.Sprintf("%d", l.Max)
	if l.Unlimited {
		s = "unlimited"
	}
	if l.FairUse {
		s += " (fair use)"
	}
	return s
}

func (d *differ) products() {
	op := index(d.o.Products, func(p Product) string { return p.ID })
	np := index(d.n.Products, func(p Product) string { return p.ID })
	for _, k := range keys(op, np) {
		o, inO := op[k]
		n, inN := np[k]
		subj := "product " + k
		switch {
		case !inO:
			d.add(Added, subj, "", false)
		case !inN:
			d.add(Removed, subj, "", false)
		case o.AmountCents != n.AmountCents:
			d.add(Changed, subj, fmt.Sprintf("price %d to %d", o.AmountCents, n.AmountCents), n.AmountCents > o.AmountCents)
		case o.SellerSet != n.SellerSet || o.MinAmountCents != n.MinAmountCents:
			d.add(Changed, subj, fmt.Sprintf("seller price floor %d to %d", o.MinAmountCents, n.MinAmountCents), n.MinAmountCents > o.MinAmountCents)
		}
	}
}

func (d *differ) marketplace() {
	var o, n *int
	if d.o.Marketplace != nil {
		o = &d.o.Marketplace.PlatformFeeBps
	}
	if d.n.Marketplace != nil {
		n = &d.n.Marketplace.PlatformFeeBps
	}
	if !equalIntPtr(o, n) {
		d.add(Changed, "marketplace", fmt.Sprintf("platform fee %s to %s", showPtr(o), showPtr(n)), n != nil && (o == nil || *n > *o))
	}
}

func (d *differ) grants() {
	idx := func(s *Spec) map[string]TierGrant {
		m := map[string]TierGrant{}
		for _, g := range s.Entitlements {
			m[g.Tier+"/"+g.Application] = g
		}
		return m
	}
	og, ng := idx(d.o), idx(d.n)
	for _, k := range keys(og, ng) {
		o, inO := og[k]
		n, inN := ng[k]
		subj := "grant " + k
		switch {
		case !inO:
			d.add(Added, subj, showSeats(n), false)
		case !inN:
			d.add(Removed, subj, showSeats(o), true)
		case o.Seats != n.Seats || o.UnlimitedSeats != n.UnlimitedSeats:
			d.add(Changed, subj, fmt.Sprintf("seats %s to %s", showSeats(o), showSeats(n)), seatsLowered(o, n))
		}
	}
}

func showSeats(g TierGrant) string {
	if g.UnlimitedSeats {
		return "unlimited seats"
	}
	return fmt.Sprintf("%d seats", g.Seats)
}

func seatsLowered(o, n TierGrant) bool {
	switch {
	case o.UnlimitedSeats:
		return !n.UnlimitedSeats
	case n.UnlimitedSeats:
		return false
	}
	return n.Seats < o.Seats
}

func equalIntPtr(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func showPtr(p *int) string {
	if p == nil {
		return "default"
	}
	return fmt.Sprintf("%d", *p)
}

// feeRaised reports whether a plan's effective platform fee goes up.
func feeRaised(o, n *int, om, nm *Marketplace) bool {
	eff := func(p *int, m *Marketplace) int {
		if p != nil {
			return *p
		}
		if m != nil {
			return m.PlatformFeeBps
		}
		return 0
	}
	return eff(n, nm) > eff(o, om)
}
