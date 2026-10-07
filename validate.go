package ratelimit

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyRules        = errors.New("rules array is empty")
	ErrMissingName       = errors.New("rule missing name")
	ErrMissingScope      = errors.New("rule missing scope")
	ErrMissingKeyField   = errors.New("rule missing keyField")
	ErrInvalidLimit      = errors.New("limit must be >= 1")
	ErrInvalidWindow     = errors.New("windowSeconds must be >= 1")
	ErrInvalidAlgorithm  = errors.New("invalid algorithm")
	ErrInvalidDistMode   = errors.New("distributionMode must be LOCAL or REMOTE")
	ErrInvalidMatch      = errors.New("match must have exactly one operator: always, exists+field, or equals+field")
	ErrMissingBurst      = errors.New("token_bucket requires params.burst >= 1")
	ErrMissingQueueSize  = errors.New("leaky_bucket requires params.queueSize >= 1")
	ErrUnexpectedParams  = errors.New("window algorithms must not have params")
)

var validAlgorithms = map[string]bool{
	"fixed_window":           true,
	"sliding_window_log":     true,
	"sliding_window_counter": true,
	"token_bucket":           true,
	"leaky_bucket":           true,
}

// Validate checks that a Config conforms to the rule schema.
// Returns nil if valid, or an error describing the first violation.
func Validate(cfg Config) error {
	if len(cfg.Rules) == 0 {
		return ErrEmptyRules
	}

	for i, r := range cfg.Rules {
		if err := validateRule(r); err != nil {
			return fmt.Errorf("rule[%d] %q: %w", i, r.Name, err)
		}
	}
	return nil
}

func validateRule(r RuleSpec) error {
	if r.Name == "" {
		return ErrMissingName
	}
	if r.Scope == "" {
		return ErrMissingScope
	}
	if r.KeyField == "" {
		return ErrMissingKeyField
	}
	if r.Limit < 1 {
		return ErrInvalidLimit
	}
	if r.WindowSeconds < 1 {
		return ErrInvalidWindow
	}
	if !validAlgorithms[r.Algorithm] {
		return fmt.Errorf("%w: %s", ErrInvalidAlgorithm, r.Algorithm)
	}
	if r.DistributionMode != "LOCAL" && r.DistributionMode != "REMOTE" {
		return ErrInvalidDistMode
	}
	if err := validateMatch(r.Match); err != nil {
		return err
	}
	if err := validateParams(r.Algorithm, r.Params); err != nil {
		return err
	}
	return nil
}

func validateMatch(m MatchSpec) error {
	hasAlways := m.Always != nil
	hasExists := m.Exists != nil
	hasEquals := m.Equals != nil

	operators := 0
	if hasAlways {
		operators++
	}
	if hasExists {
		operators++
	}
	if hasEquals {
		operators++
	}

	if operators != 1 {
		return ErrInvalidMatch
	}

	// exists and equals require field
	if (hasExists || hasEquals) && m.Field == "" {
		return ErrInvalidMatch
	}

	return nil
}

func validateParams(algorithm string, params map[string]any) error {
	switch algorithm {
	case "token_bucket":
		burst, ok := getIntParam(params, "burst")
		if !ok || burst < 1 {
			return ErrMissingBurst
		}

	case "leaky_bucket":
		queueSize, ok := getIntParam(params, "queueSize")
		if !ok || queueSize < 1 {
			return ErrMissingQueueSize
		}

	case "fixed_window", "sliding_window_log", "sliding_window_counter":
		if len(params) > 0 {
			return ErrUnexpectedParams
		}
	}

	return nil
}

func getIntParam(params map[string]any, key string) (int, bool) {
	if params == nil {
		return 0, false
	}
	v, ok := params[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case int:
		return val, true
	case float64:
		return int(val), true
	case int64:
		return int(val), true
	}
	return 0, false
}
