package report

import (
	"math"
	"testing"
)

func TestRules_KeysMatchCodeAndAreWellFormed(t *testing.T) {
	wantCodes := []string{
		"CS101", "CS102", "CS103", "CS104", "CS105", "CS106",
		"CS107", "CS108", "CS109", "CS110", "CS111", "CS112",
	}

	if len(Rules) != len(wantCodes) {
		t.Fatalf("Rules has %d entries, want %d", len(Rules), len(wantCodes))
	}

	for _, code := range wantCodes {
		meta, ok := Rules[code]
		if !ok {
			t.Errorf("Rules missing entry for %s", code)
			continue
		}
		if meta.Code != code {
			t.Errorf("Rules[%q].Code = %q, want %q", code, meta.Code, code)
		}
		if meta.Name == "" {
			t.Errorf("Rules[%q].Name is empty", code)
		}
		if meta.Suggestion == "" {
			t.Errorf("Rules[%q].Suggestion is empty", code)
		}
		switch meta.Severity {
		case High, Medium, Low:
		default:
			t.Errorf("Rules[%q].Severity = %q, want High, Medium or Low", code, meta.Severity)
		}
	}
}

func TestConfidenceFromExcess_NoExcessReturnsBaseline(t *testing.T) {
	got := ConfidenceFromExcess(0.5, 0.9, 1.0)
	if got != 60 {
		t.Errorf("ConfidenceFromExcess with no ratio over 1 = %d, want 60", got)
	}
}

func TestConfidenceFromExcess_IncreasesWithMoreExceededDimensions(t *testing.T) {
	one := ConfidenceFromExcess(1.5)
	two := ConfidenceFromExcess(1.5, 1.5)

	if two <= one {
		t.Errorf("expected confidence to increase with more exceeded dimensions: one=%d, two=%d", one, two)
	}
}

func TestConfidenceFromExcess_CapsAt97(t *testing.T) {
	got := ConfidenceFromExcess(10, 10, 10, 10, 10)
	if got != 97 {
		t.Errorf("ConfidenceFromExcess with large ratios = %d, want capped at 97", got)
	}
}

func TestConfidenceFromExcess_HandlesInfinityAndNaN(t *testing.T) {
	if got := ConfidenceFromExcess(math.Inf(1)); got != 97 {
		t.Errorf("ConfidenceFromExcess(+Inf) = %d, want 97", got)
	}
	if got := ConfidenceFromExcess(math.NaN()); got != 97 {
		t.Errorf("ConfidenceFromExcess(NaN) = %d, want 97", got)
	}
	if got := ConfidenceFromExcess(1.5, math.Inf(1)); got != 97 {
		t.Errorf("ConfidenceFromExcess(1.5, +Inf) = %d, want 97", got)
	}
}
