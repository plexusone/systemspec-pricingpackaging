package pricingspec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

// Severity says whether an Issue rejects a Spec.
type Severity string

// Severities.
const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Issue is one finding about a Spec. Its Code is stable (E001 to E013 for
// errors, W001 to W006 for warnings) and documented in SPEC.md.
type Issue struct {
	Code     string   `json:"code"`
	Severity Severity `json:"severity"`
	// Path locates the finding, for example "plans[2].limits[0]".
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (i Issue) String() string {
	return fmt.Sprintf("%s %s: %s: %s", i.Code, i.Severity, i.Path, i.Message)
}

// ValidationError is returned by Load and Parse for a Spec that breaks an
// error rule. It carries every error found, not only the first.
type ValidationError struct {
	Issues []Issue
}

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Issues))
	for i, is := range e.Issues {
		parts[i] = is.String()
	}
	return "pricingspec: invalid spec:\n  " + strings.Join(parts, "\n  ")
}

// ErrUnknownFormat is returned for a document whose formatVersion this package
// does not read.
var ErrUnknownFormat = errors.New("pricingspec: unknown format version")

// Load reads a Spec from JSON, rejecting unknown fields, and validates it.
func Load(r io.Reader) (*Spec, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("pricingspec: read: %w", err)
	}
	return Parse(data)
}

// Parse decodes a Spec from JSON, rejecting unknown fields, and validates it.
// A Spec that breaks an error rule returns a *ValidationError.
func Parse(data []byte) (*Spec, error) {
	var probe struct {
		FormatVersion string `json:"formatVersion"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("pricingspec: decode: %w", err)
	}
	if probe.FormatVersion != FormatVersion {
		return nil, fmt.Errorf("%w: %q", ErrUnknownFormat, probe.FormatVersion)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var s Spec
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("pricingspec: decode: %w", err)
	}
	if issues := s.Validate(); len(issues) > 0 {
		return nil, &ValidationError{Issues: issues}
	}
	return &s, nil
}

var (
	appPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	idPattern  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	curPattern = regexp.MustCompile(`^[a-z]{3}$`)
)

// Validate returns every error-level finding (E001 to E013). A Spec with none
// is valid. It does not report warnings; see Lint.
func (s *Spec) Validate() []Issue {
	v := &validator{s: s}
	v.run()
	return v.issues
}

type validator struct {
	s      *Spec
	issues []Issue

	apps, tiers, features, limits, plans map[string]bool
	featureByID                          map[string]Feature
}

func (v *validator) errf(code, path, format string, a ...any) {
	v.issues = append(v.issues, Issue{Code: code, Severity: SeverityError, Path: path, Message: fmt.Sprintf(format, a...)})
}

func (v *validator) run() {
	s := v.s
	if s.FormatVersion != FormatVersion {
		v.errf("E001", "formatVersion", "format version %q is not %q", s.FormatVersion, FormatVersion)
	}
	if !appPattern.MatchString(s.App) {
		v.errf("E002", "app", "app %q must be lowercase letters, digits and hyphens, starting with a letter", s.App)
	}
	if !curPattern.MatchString(s.Currency) {
		v.errf("E002", "currency", "currency %q must be a lowercase three-letter ISO 4217 code", s.Currency)
	}

	v.apps = v.ids("applications", len(s.Applications), func(i int) string { return s.Applications[i].ID })
	v.tiers = v.ids("tiers", len(s.Tiers), func(i int) string { return s.Tiers[i].ID })
	v.features = v.ids("features", len(s.Features), func(i int) string { return s.Features[i].ID })
	v.limits = v.ids("limits", len(s.Limits), func(i int) string { return s.Limits[i].ID })
	v.plans = v.ids("plans", len(s.Plans), func(i int) string { return s.Plans[i].ID })
	v.ids("products", len(s.Products), func(i int) string { return s.Products[i].ID })

	ranks := map[int]bool{}
	for i, t := range s.Tiers {
		p := fmt.Sprintf("tiers[%d].rank", i)
		switch {
		case t.Rank < 1:
			v.errf("E004", p, "rank %d must be positive", t.Rank)
		case ranks[t.Rank]:
			v.errf("E004", p, "rank %d is used by another tier", t.Rank)
		}
		ranks[t.Rank] = true
	}

	v.featureByID = make(map[string]Feature, len(s.Features))
	for i, f := range s.Features {
		v.featureByID[f.ID] = f
		if f.Levels != nil {
			v.levels(fmt.Sprintf("features[%d].levels", i), f.Levels)
		}
	}

	for i, p := range s.Plans {
		v.plan(fmt.Sprintf("plans[%d]", i), p)
	}
	v.defaultPlan()
	for i, pr := range s.Products {
		v.product(fmt.Sprintf("products[%d]", i), pr)
	}
	if m := s.Marketplace; m != nil && (m.PlatformFeeBps < 0 || m.PlatformFeeBps > maxBps) {
		v.errf("E012", "marketplace.platformFeeBps", "fee %d must be within 0 to %d", m.PlatformFeeBps, maxBps)
	}
	v.grants()
}

const maxBps = 10000

// ids reports malformed and duplicate IDs (E003) and returns the set.
func (v *validator) ids(kind string, n int, id func(int) string) map[string]bool {
	seen := make(map[string]bool, n)
	for i := range n {
		x := id(i)
		path := fmt.Sprintf("%s[%d].id", kind, i)
		switch {
		case !idPattern.MatchString(x):
			v.errf("E003", path, "id %q must be lowercase letters, digits and underscores, starting with a letter", x)
		case seen[x]:
			v.errf("E003", path, "id %q is used twice in %s", x, kind)
		}
		seen[x] = true
	}
	return seen
}

func (v *validator) levels(path string, levels []string) {
	seen := map[string]bool{}
	for _, l := range levels {
		if l == "" || seen[l] {
			v.errf("E007", path, "levels must be distinct and non-empty")
			return
		}
		seen[l] = true
	}
	if len(levels) < 2 {
		v.errf("E007", path, "a ladder needs at least two levels; a feature with one state is a switch")
	}
}

func (v *validator) plan(path string, p Plan) {
	if !v.tiers[p.Tier] {
		v.errf("E005", path+".tier", "tier %q is not declared", p.Tier)
	}

	seenF := map[string]bool{}
	for i, fv := range p.Features {
		fp := fmt.Sprintf("%s.features[%d]", path, i)
		f, ok := v.featureByID[fv.Feature]
		switch {
		case !ok:
			v.errf("E006", fp+".feature", "feature %q is not declared", fv.Feature)
		case seenF[fv.Feature]:
			v.errf("E006", fp+".feature", "feature %q is set twice", fv.Feature)
		case len(f.Levels) == 0 && fv.Level != "":
			v.errf("E006", fp+".level", "feature %q is a switch and takes no level", fv.Feature)
		case len(f.Levels) > 0 && fv.Enabled:
			v.errf("E006", fp+".enabled", "feature %q is a ladder and takes a level, not enabled", fv.Feature)
		case len(f.Levels) > 0 && !slices.Contains(f.Levels, fv.Level):
			v.errf("E006", fp+".level", "level %q is not a level of feature %q", fv.Level, fv.Feature)
		}
		seenF[fv.Feature] = true
	}

	seenL := map[string]bool{}
	for i, lv := range p.Limits {
		lp := fmt.Sprintf("%s.limits[%d]", path, i)
		switch {
		case !v.limits[lv.Limit]:
			v.errf("E008", lp+".limit", "limit %q is not declared", lv.Limit)
		case seenL[lv.Limit]:
			v.errf("E008", lp+".limit", "limit %q is set twice", lv.Limit)
		}
		seenL[lv.Limit] = true
		if lv.Max < 0 {
			v.errf("E008", lp+".max", "max %d must not be negative", lv.Max)
		}
		if lv.Unlimited && lv.Max != 0 {
			v.errf("E008", lp+".max", "an unlimited limit must have max zero")
		}
	}

	seenI := map[Interval]bool{}
	for i, pr := range p.Prices {
		pp := fmt.Sprintf("%s.prices[%d]", path, i)
		switch {
		case pr.Interval != IntervalMonth && pr.Interval != IntervalYear:
			v.errf("E009", pp+".interval", "interval %q must be month or year", pr.Interval)
		case seenI[pr.Interval]:
			v.errf("E009", pp+".interval", "interval %q is priced twice", pr.Interval)
		}
		seenI[pr.Interval] = true
		if pr.AmountCents < 1 {
			v.errf("E009", pp+".amountCents", "amount %d must be a positive number of minor units", pr.AmountCents)
		}
	}
	if p.TrialDays < 0 || (p.TrialDays > 0 && len(p.Prices) == 0) {
		v.errf("E009", path+".trialDays", "a trial needs prices and a non-negative length")
	}
	if p.SelfServe && len(p.Prices) == 0 && p.ID != v.s.DefaultPlan {
		v.errf("E009", path, "self-serve plan %q has no price; only the default plan may be free", p.ID)
	}
	if p.PlatformFeeBps != nil && (*p.PlatformFeeBps < 0 || *p.PlatformFeeBps > maxBps) {
		v.errf("E012", path+".platformFeeBps", "fee %d must be within 0 to %d", *p.PlatformFeeBps, maxBps)
	}
}

func (v *validator) defaultPlan() {
	var def *Plan
	for i := range v.s.Plans {
		if v.s.Plans[i].ID == v.s.DefaultPlan {
			def = &v.s.Plans[i]
		}
	}
	switch {
	case def == nil:
		v.errf("E010", "defaultPlan", "default plan %q is not a plan", v.s.DefaultPlan)
		return
	case !def.SelfServe || len(def.Prices) > 0:
		v.errf("E010", "defaultPlan", "default plan %q must be self-serve and free", def.ID)
	}
	for _, g := range v.s.Entitlements {
		if g.Tier == def.Tier {
			return
		}
	}
	v.errf("E010", "defaultPlan", "the default plan's tier %q has no grant", def.Tier)
}

func (v *validator) product(path string, p Product) {
	switch {
	case p.SellerSet && p.AmountCents != 0:
		v.errf("E011", path, "a seller-set product has no fixed amount")
	case p.SellerSet && p.MinAmountCents < 1:
		v.errf("E011", path, "a seller-set product needs a positive minimum")
	case !p.SellerSet && p.AmountCents < 1:
		v.errf("E011", path, "a fixed product needs a positive amount")
	case !p.SellerSet && p.MinAmountCents != 0:
		v.errf("E011", path, "a fixed product has no minimum")
	}
}

func (v *validator) grants() {
	type key struct{ tier, app string }
	seen := map[key]bool{}
	for i, g := range v.s.Entitlements {
		path := fmt.Sprintf("entitlements[%d]", i)
		if !v.tiers[g.Tier] {
			v.errf("E013", path+".tier", "tier %q is not declared", g.Tier)
		}
		if !v.apps[g.Application] {
			v.errf("E013", path+".application", "application %q is not declared", g.Application)
		}
		switch {
		case g.UnlimitedSeats && g.Seats != 0:
			v.errf("E013", path+".seats", "set seats or unlimitedSeats, not both")
		case !g.UnlimitedSeats && g.Seats < 1:
			v.errf("E013", path+".seats", "seats must be at least 1, or unlimitedSeats")
		}
		k := key{g.Tier, g.Application}
		if seen[k] {
			v.errf("E013", path, "tier %q already grants application %q", g.Tier, g.Application)
		}
		seen[k] = true
	}
}
