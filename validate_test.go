package ratelimit

import (
	"errors"
	"testing"
)

func TestValidate_ValidConfig(t *testing.T) {
	trueVal := true
	cfg := Config{
		Rules: []RuleSpec{
			{
				Name:             "test-rule",
				Scope:            "user",
				KeyField:         "userId",
				Match:            MatchSpec{Always: &trueVal},
				Limit:            100,
				WindowSeconds:    60,
				Algorithm:        "token_bucket",
				DistributionMode: "LOCAL",
				Params:           map[string]any{"burst": 150},
			},
		},
	}

	if err := Validate(cfg); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestValidate_EmptyRules(t *testing.T) {
	cfg := Config{Rules: []RuleSpec{}}
	if err := Validate(cfg); !errors.Is(err, ErrEmptyRules) {
		t.Errorf("expected ErrEmptyRules, got: %v", err)
	}
}

func TestValidate_MissingName(t *testing.T) {
	trueVal := true
	cfg := Config{
		Rules: []RuleSpec{{
			Scope:            "user",
			KeyField:         "userId",
			Match:            MatchSpec{Always: &trueVal},
			Limit:            100,
			WindowSeconds:    60,
			Algorithm:        "fixed_window",
			DistributionMode: "LOCAL",
		}},
	}
	if err := Validate(cfg); !errors.Is(err, ErrMissingName) {
		t.Errorf("expected ErrMissingName, got: %v", err)
	}
}

func TestValidate_InvalidAlgorithm(t *testing.T) {
	trueVal := true
	cfg := Config{
		Rules: []RuleSpec{{
			Name:             "test",
			Scope:            "user",
			KeyField:         "userId",
			Match:            MatchSpec{Always: &trueVal},
			Limit:            100,
			WindowSeconds:    60,
			Algorithm:        "invalid_algo",
			DistributionMode: "LOCAL",
		}},
	}
	if err := Validate(cfg); !errors.Is(err, ErrInvalidAlgorithm) {
		t.Errorf("expected ErrInvalidAlgorithm, got: %v", err)
	}
}

func TestValidate_InvalidDistributionMode(t *testing.T) {
	trueVal := true
	cfg := Config{
		Rules: []RuleSpec{{
			Name:             "test",
			Scope:            "user",
			KeyField:         "userId",
			Match:            MatchSpec{Always: &trueVal},
			Limit:            100,
			WindowSeconds:    60,
			Algorithm:        "fixed_window",
			DistributionMode: "INVALID",
		}},
	}
	if err := Validate(cfg); !errors.Is(err, ErrInvalidDistMode) {
		t.Errorf("expected ErrInvalidDistMode, got: %v", err)
	}
}

func TestValidate_MatchOperators(t *testing.T) {
	trueVal := true
	falseVal := false
	strVal := "premium"

	tests := []struct {
		name    string
		match   MatchSpec
		wantErr bool
	}{
		{"always true", MatchSpec{Always: &trueVal}, false},
		{"always false", MatchSpec{Always: &falseVal}, false},
		{"exists with field", MatchSpec{Field: "userId", Exists: &trueVal}, false},
		{"equals with field", MatchSpec{Field: "tier", Equals: &strVal}, false},
		{"exists without field", MatchSpec{Exists: &trueVal}, true},
		{"equals without field", MatchSpec{Equals: &strVal}, true},
		{"no operator", MatchSpec{}, true},
		{"multiple operators", MatchSpec{Always: &trueVal, Exists: &trueVal}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Rules: []RuleSpec{{
					Name:             "test",
					Scope:            "user",
					KeyField:         "userId",
					Match:            tt.match,
					Limit:            100,
					WindowSeconds:    60,
					Algorithm:        "fixed_window",
					DistributionMode: "LOCAL",
				}},
			}
			err := Validate(cfg)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidate_TokenBucketRequiresBurst(t *testing.T) {
	trueVal := true
	cfg := Config{
		Rules: []RuleSpec{{
			Name:             "test",
			Scope:            "user",
			KeyField:         "userId",
			Match:            MatchSpec{Always: &trueVal},
			Limit:            100,
			WindowSeconds:    60,
			Algorithm:        "token_bucket",
			DistributionMode: "LOCAL",
			// missing params.burst
		}},
	}
	if err := Validate(cfg); !errors.Is(err, ErrMissingBurst) {
		t.Errorf("expected ErrMissingBurst, got: %v", err)
	}

	// with burst = 0 should also fail
	cfg.Rules[0].Params = map[string]any{"burst": 0}
	if err := Validate(cfg); !errors.Is(err, ErrMissingBurst) {
		t.Errorf("expected ErrMissingBurst for burst=0, got: %v", err)
	}

	// with valid burst should pass
	cfg.Rules[0].Params = map[string]any{"burst": 150}
	if err := Validate(cfg); err != nil {
		t.Errorf("expected no error with valid burst, got: %v", err)
	}
}

func TestValidate_FixedWindowRejectsParams(t *testing.T) {
	trueVal := true
	cfg := Config{
		Rules: []RuleSpec{{
			Name:             "test",
			Scope:            "user",
			KeyField:         "userId",
			Match:            MatchSpec{Always: &trueVal},
			Limit:            100,
			WindowSeconds:    60,
			Algorithm:        "fixed_window",
			DistributionMode: "LOCAL",
			Params:           map[string]any{"burst": 150}, // should not have params
		}},
	}
	if err := Validate(cfg); !errors.Is(err, ErrUnexpectedParams) {
		t.Errorf("expected ErrUnexpectedParams, got: %v", err)
	}
}
