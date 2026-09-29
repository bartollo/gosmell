package rules

import "testing"

func TestDefaultConfig_HasPositiveThresholds(t *testing.T) {
	cfg := DefaultConfig()

	checks := map[string]int{
		"LongMethod.MaxStatements":       cfg.LongMethod.MaxStatements,
		"LongMethod.MaxComplexity":       cfg.LongMethod.MaxComplexity,
		"LongMethod.MaxNesting":          cfg.LongMethod.MaxNesting,
		"LongMethod.MaxCalls":            cfg.LongMethod.MaxCalls,
		"LongMethod.MaxCollaborators":    cfg.LongMethod.MaxCollaborators,
		"LargeClass.MaxFields":           cfg.LargeClass.MaxFields,
		"LargeClass.MaxMethods":          cfg.LargeClass.MaxMethods,
		"LargeClass.MaxCollaborators":    cfg.LargeClass.MaxCollaborators,
		"LargeClass.MaxStatements":       cfg.LargeClass.MaxStatements,
		"LargeClass.MinExceeded":         cfg.LargeClass.MinExceeded,
		"ExcessiveNesting.MaxDepth":      cfg.ExcessiveNesting.MaxDepth,
		"DuplicateCode.MinStatements":    cfg.DuplicateCode.MinStatements,
		"CopyPasteDrift.MinSiblingGroup": cfg.CopyPasteDrift.MinSiblingGroup,
		"CopyPasteDrift.MaxEditDistance": cfg.CopyPasteDrift.MaxEditDistance,
	}

	for name, value := range checks {
		if value <= 0 {
			t.Errorf("DefaultConfig().%s = %d, want > 0", name, value)
		}
	}
}

func TestOverrides_ApplyOnlyOverridesSetFields(t *testing.T) {
	base := DefaultConfig()

	maxStatements := 999
	overrides := Overrides{
		LongMethod: LongMethodOverrides{
			MaxStatements: &maxStatements,
		},
	}

	got := overrides.Apply(base)

	if got.LongMethod.MaxStatements != 999 {
		t.Errorf("LongMethod.MaxStatements = %d, want 999", got.LongMethod.MaxStatements)
	}

	if got.LongMethod.MaxComplexity != base.LongMethod.MaxComplexity {
		t.Errorf("LongMethod.MaxComplexity = %d, want unchanged default %d", got.LongMethod.MaxComplexity, base.LongMethod.MaxComplexity)
	}
	if got.LargeClass != base.LargeClass {
		t.Errorf("LargeClass = %+v, want unchanged default %+v", got.LargeClass, base.LargeClass)
	}
	if got.ExcessiveNesting != base.ExcessiveNesting {
		t.Errorf("ExcessiveNesting = %+v, want unchanged default %+v", got.ExcessiveNesting, base.ExcessiveNesting)
	}
	if got.DuplicateCode != base.DuplicateCode {
		t.Errorf("DuplicateCode = %+v, want unchanged default %+v", got.DuplicateCode, base.DuplicateCode)
	}
	if got.CopyPasteDrift != base.CopyPasteDrift {
		t.Errorf("CopyPasteDrift = %+v, want unchanged default %+v", got.CopyPasteDrift, base.CopyPasteDrift)
	}
}

func TestOverrides_EmptyApplyIsIdentity(t *testing.T) {
	base := DefaultConfig()

	got := Overrides{}.Apply(base)

	if got != base {
		t.Errorf("empty Overrides.Apply changed the config: got %+v, want %+v", got, base)
	}
}
