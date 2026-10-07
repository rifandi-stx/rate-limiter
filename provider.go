package ratelimit

// RuleSpec is a rule as it appears in config (parsed from JSON).
type RuleSpec struct {
	Name             string         `json:"name"`
	Scope            string         `json:"scope"`
	KeyField         string         `json:"keyField"`
	Match            MatchSpec      `json:"match"`
	Limit            int            `json:"limit"`
	WindowSeconds    int            `json:"windowSeconds"`
	Algorithm        string         `json:"algorithm"`
	DistributionMode string         `json:"distributionMode"`
	Params           map[string]any `json:"params,omitempty"`
}

// MatchSpec defines when a rule applies.
// Exactly one operator should be set: Always, Exists, or Equals.
type MatchSpec struct {
	Always *bool   `json:"always,omitempty"`
	Field  string  `json:"field,omitempty"`
	Exists *bool   `json:"exists,omitempty"`
	Equals *string `json:"equals,omitempty"`
}

// Config is the parsed rule set from config.
type Config struct {
	Rules []RuleSpec `json:"rules"`
}

// DeclarativeProvider implements RuleProvider by interpreting config rules.
// v1 is first-match-wins: returns the first rule whose match is satisfied
// AND whose keyField resolves to a non-empty value.
type DeclarativeProvider struct {
	rules func() []RuleSpec
}

// NewDeclarativeProvider creates a provider that reads rules from the given function.
// Typically backed by a config poller: func() []RuleSpec { return poller.Current().Rules }
func NewDeclarativeProvider(rules func() []RuleSpec) *DeclarativeProvider {
	return &DeclarativeProvider{rules: rules}
}

// RulesFor returns the applicable rules for a request.
// v1 returns at most one rule (first-match-wins).
func (d *DeclarativeProvider) RulesFor(ctx RequestContext) []Rule {
	for _, spec := range d.rules() {
		if !matches(spec.Match, ctx) {
			continue
		}
		val, ok := ctx.Field(spec.KeyField)
		if !ok {
			continue // matched but no key value: skip
		}
		return []Rule{{
			Name:             spec.Name,
			Scope:            spec.Scope,
			KeyValue:         val,
			Limit:            spec.Limit,
			WindowSeconds:    spec.WindowSeconds,
			Algorithm:        spec.Algorithm,
			DistributionMode: spec.DistributionMode,
			Params:           spec.Params,
		}}
	}
	return nil // no match: allow-by-default
}

func matches(m MatchSpec, ctx RequestContext) bool {
	if m.Always != nil {
		return *m.Always
	}
	if m.Field == "" {
		return false // malformed: no field for exists/equals
	}
	val, present := ctx.Field(m.Field)
	switch {
	case m.Exists != nil:
		return present == *m.Exists
	case m.Equals != nil:
		return present && val == *m.Equals
	}
	return false
}
