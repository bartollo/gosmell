package gosmell

import (
	"testing"

	"github.com/bartollo/gosmell/rules"
)

func TestNew_DefaultsWhenNoSettings(t *testing.T) {
	plugin, err := New(nil)
	if err != nil {
		t.Fatalf("New(nil) error = %v", err)
	}

	p, ok := plugin.(*gosmellPlugin)
	if !ok {
		t.Fatalf("New(nil) returned %T, want *gosmellPlugin", plugin)
	}
	if p.cfg != rules.DefaultConfig() {
		t.Errorf("cfg = %+v, want DefaultConfig()", p.cfg)
	}
}

func TestNew_AppliesOverrides(t *testing.T) {
	settings := map[string]any{
		"longMethod": map[string]any{
			"maxStatements": 5,
		},
	}

	plugin, err := New(settings)
	if err != nil {
		t.Fatalf("New(settings) error = %v", err)
	}

	p := plugin.(*gosmellPlugin)
	if p.cfg.LongMethod.MaxStatements != 5 {
		t.Errorf("LongMethod.MaxStatements = %d, want 5", p.cfg.LongMethod.MaxStatements)
	}

	if p.cfg.LargeClass != rules.DefaultConfig().LargeClass {
		t.Errorf("LargeClass = %+v, want unchanged default", p.cfg.LargeClass)
	}
}

func TestNew_RejectsUnknownSettingsField(t *testing.T) {
	settings := map[string]any{
		"longMethod": map[string]any{
			"maxStatments": 5,
		},
	}

	_, err := New(settings)
	if err == nil {
		t.Fatal("New(settings) with an unknown field succeeded, want an error")
	}
}

func TestGosmellPlugin_BuildAnalyzers(t *testing.T) {
	plugin, err := New(nil)
	if err != nil {
		t.Fatalf("New(nil) error = %v", err)
	}

	analyzers, err := plugin.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() error = %v", err)
	}

	wantNames := []string{
		"longmethod", "largeclass", "excessivenesting", "duplicatecode",
		"deadcode", "unusedctordep", "swallowederror", "redundantcondition",
		"comments", "defensivenoise", "copypastedrift", "placeholderimpl",
	}
	if len(analyzers) != len(wantNames) {
		t.Fatalf("BuildAnalyzers() returned %d analyzers, want %d", len(analyzers), len(wantNames))
	}

	seen := map[string]bool{}
	for _, a := range analyzers {
		if a.Name == "" {
			t.Error("an analyzer has an empty Name")
		}
		if seen[a.Name] {
			t.Errorf("duplicate analyzer name %q", a.Name)
		}
		seen[a.Name] = true
		if a.Run == nil {
			t.Errorf("analyzer %q has a nil Run func", a.Name)
		}
	}
	for _, name := range wantNames {
		if !seen[name] {
			t.Errorf("BuildAnalyzers() is missing analyzer %q", name)
		}
	}
}

func TestGosmellPlugin_GetLoadMode(t *testing.T) {
	plugin, err := New(nil)
	if err != nil {
		t.Fatalf("New(nil) error = %v", err)
	}

	if got := plugin.GetLoadMode(); got != "typesinfo" {
		t.Errorf("GetLoadMode() = %q, want %q", got, "typesinfo")
	}
}
