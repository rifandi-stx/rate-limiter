package ratelimit

import "testing"

func TestDeclarativeProvider_AlwaysMatch(t *testing.T) {
	rules := []RuleSpec{{
		Name:             "catch-all",
		Scope:            "ip",
		KeyField:         "ip",
		Match:            MatchSpec{Always: boolPtr(true)},
		Limit:            100,
		WindowSeconds:    60,
		Algorithm:        "fixed_window",
		DistributionMode: "LOCAL",
	}}
	provider := NewDeclarativeProvider(func() []RuleSpec { return rules })

	ctx := RequestContext{IP: "1.2.3.4"}
	result := provider.RulesFor(ctx)

	if len(result) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(result))
	}
	if result[0].Name != "catch-all" {
		t.Errorf("rule name=%s, want catch-all", result[0].Name)
	}
	if result[0].KeyValue != "1.2.3.4" {
		t.Errorf("keyValue=%s, want 1.2.3.4", result[0].KeyValue)
	}
}

func TestDeclarativeProvider_ExistsMatch(t *testing.T) {
	rules := []RuleSpec{{
		Name:             "per-user",
		Scope:            "user",
		KeyField:         "userId",
		Match:            MatchSpec{Field: "userId", Exists: boolPtr(true)},
		Limit:            100,
		WindowSeconds:    60,
		Algorithm:        "fixed_window",
		DistributionMode: "LOCAL",
	}}
	provider := NewDeclarativeProvider(func() []RuleSpec { return rules })

	// With userId
	ctx1 := RequestContext{UserID: "123"}
	result1 := provider.RulesFor(ctx1)
	if len(result1) != 1 {
		t.Errorf("expected 1 rule for user with ID, got %d", len(result1))
	}

	// Without userId
	ctx2 := RequestContext{IP: "1.2.3.4"}
	result2 := provider.RulesFor(ctx2)
	if len(result2) != 0 {
		t.Errorf("expected 0 rules for user without ID, got %d", len(result2))
	}
}

func TestDeclarativeProvider_EqualsMatch(t *testing.T) {
	rules := []RuleSpec{{
		Name:             "premium-org",
		Scope:            "org",
		KeyField:         "merchantId",
		Match:            MatchSpec{Field: "orgTier", Equals: strPtr("premium")},
		Limit:            1000,
		WindowSeconds:    60,
		Algorithm:        "fixed_window",
		DistributionMode: "LOCAL",
	}}
	provider := NewDeclarativeProvider(func() []RuleSpec { return rules })

	// Premium tier
	ctx1 := RequestContext{
		MerchantID: "org-456",
		Custom:     map[string]string{"orgTier": "premium"},
	}
	result1 := provider.RulesFor(ctx1)
	if len(result1) != 1 {
		t.Errorf("expected 1 rule for premium org, got %d", len(result1))
	}

	// Standard tier
	ctx2 := RequestContext{
		MerchantID: "org-789",
		Custom:     map[string]string{"orgTier": "standard"},
	}
	result2 := provider.RulesFor(ctx2)
	if len(result2) != 0 {
		t.Errorf("expected 0 rules for standard org, got %d", len(result2))
	}
}

func TestDeclarativeProvider_FirstMatchWins(t *testing.T) {
	rules := []RuleSpec{
		{
			Name:             "premium-org",
			Scope:            "org",
			KeyField:         "merchantId",
			Match:            MatchSpec{Field: "orgTier", Equals: strPtr("premium")},
			Limit:            1000,
			WindowSeconds:    60,
			Algorithm:        "fixed_window",
			DistributionMode: "LOCAL",
		},
		{
			Name:             "default-org",
			Scope:            "org",
			KeyField:         "merchantId",
			Match:            MatchSpec{Field: "merchantId", Exists: boolPtr(true)},
			Limit:            100,
			WindowSeconds:    60,
			Algorithm:        "fixed_window",
			DistributionMode: "LOCAL",
		},
	}
	provider := NewDeclarativeProvider(func() []RuleSpec { return rules })

	// Premium matches first rule
	ctx1 := RequestContext{
		MerchantID: "org-456",
		Custom:     map[string]string{"orgTier": "premium"},
	}
	result1 := provider.RulesFor(ctx1)
	if len(result1) != 1 || result1[0].Name != "premium-org" {
		t.Errorf("expected premium-org rule, got %v", result1)
	}

	// Non-premium falls through to second rule
	ctx2 := RequestContext{MerchantID: "org-789"}
	result2 := provider.RulesFor(ctx2)
	if len(result2) != 1 || result2[0].Name != "default-org" {
		t.Errorf("expected default-org rule, got %v", result2)
	}
}

func TestDeclarativeProvider_SkipsIfKeyFieldMissing(t *testing.T) {
	rules := []RuleSpec{{
		Name:             "per-user",
		Scope:            "user",
		KeyField:         "userId",
		Match:            MatchSpec{Always: boolPtr(true)}, // always matches
		Limit:            100,
		WindowSeconds:    60,
		Algorithm:        "fixed_window",
		DistributionMode: "LOCAL",
	}}
	provider := NewDeclarativeProvider(func() []RuleSpec { return rules })

	// No userId, rule matches but keyField is missing
	ctx := RequestContext{IP: "1.2.3.4"}
	result := provider.RulesFor(ctx)
	if len(result) != 0 {
		t.Errorf("expected 0 rules when keyField is missing, got %d", len(result))
	}
}

func strPtr(s string) *string {
	return &s
}
