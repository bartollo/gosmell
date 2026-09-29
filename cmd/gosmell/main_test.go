package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bartollo/gosmell/report"
)

func TestCountLines(t *testing.T) {
	dir := t.TempDir()

	withTrailingNewline := filepath.Join(dir, "a.go")
	if err := os.WriteFile(withTrailingNewline, []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatalf("writing file: %v", err)
	}
	if got := countLines(withTrailingNewline); got != 3 {
		t.Errorf("countLines with trailing newline = %d, want 3", got)
	}

	withoutTrailingNewline := filepath.Join(dir, "b.go")
	if err := os.WriteFile(withoutTrailingNewline, []byte("one\ntwo\nthree"), 0o644); err != nil {
		t.Fatalf("writing file: %v", err)
	}
	if got := countLines(withoutTrailingNewline); got != 3 {
		t.Errorf("countLines without trailing newline = %d, want 3", got)
	}
}

func TestFormatThousands(t *testing.T) {
	cases := map[int]string{
		0:       "0",
		7:       "7",
		999:     "999",
		1000:    "1,000",
		17895:   "17,895",
		1234567: "1,234,567",
	}

	for in, want := range cases {
		if got := formatThousands(in); got != want {
			t.Errorf("formatThousands(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestScoreTier(t *testing.T) {
	cases := map[int]string{
		100: "Pristine",
		90:  "Pristine",
		89:  "Clean",
		75:  "Clean",
		74:  "Sloppy",
		50:  "Sloppy",
		49:  "Messy",
		25:  "Messy",
		24:  "Disaster",
		0:   "Disaster",
	}

	for score, want := range cases {
		if got := scoreTier(score); got != want {
			t.Errorf("scoreTier(%d) = %q, want %q", score, got, want)
		}
	}
}

func TestComputeScore(t *testing.T) {
	cases := []struct {
		high, medium, low int
		want              int
	}{
		{0, 0, 0, 100},
		{1, 0, 0, 95},
		{0, 1, 0, 97},
		{0, 0, 1, 99},
		{2, 2, 1, 100 - (2*5 + 2*3 + 1)},
		{100, 0, 0, 0},
	}

	for _, c := range cases {
		if got := computeScore(c.high, c.medium, c.low); got != c.want {
			t.Errorf("computeScore(%d, %d, %d) = %d, want %d", c.high, c.medium, c.low, got, c.want)
		}
	}
}

func TestCountSeverities(t *testing.T) {
	fileFindings := map[string][]displayFinding{
		"a.go": {
			{meta: report.RuleMeta{Severity: report.High}},
			{meta: report.RuleMeta{Severity: report.Medium}},
		},
		"b.go": {
			{meta: report.RuleMeta{Severity: report.Low}},
			{meta: report.RuleMeta{Severity: report.High}},
		},
	}

	total, high, medium, low := countSeverities(fileFindings)

	if total != 4 {
		t.Errorf("total = %d, want 4", total)
	}
	if high != 2 {
		t.Errorf("high = %d, want 2", high)
	}
	if medium != 1 {
		t.Errorf("medium = %d, want 1", medium)
	}
	if low != 1 {
		t.Errorf("low = %d, want 1", low)
	}
}

func TestSummarizeSeverities(t *testing.T) {
	cases := []struct {
		high, medium, low int
		want              string
	}{
		{0, 0, 0, ""},
		{2, 0, 0, "2 high"},
		{0, 3, 0, "3 medium"},
		{1, 1, 1, "1 high, 1 medium, 1 low"},
	}

	for _, c := range cases {
		if got := summarizeSeverities(c.high, c.medium, c.low); got != c.want {
			t.Errorf("summarizeSeverities(%d, %d, %d) = %q, want %q", c.high, c.medium, c.low, got, c.want)
		}
	}
}

func TestBuildFileFindings_ParsesTextAndFiltersOtherLinters(t *testing.T) {
	issues := &golangciJSON{}
	issues.Issues = append(issues.Issues,
		struct {
			FromLinter string `json:"FromLinter"`
			Text       string `json:"Text"`
			Pos        struct {
				Filename string `json:"Filename"`
				Line     int    `json:"Line"`
			} `json:"Pos"`
		}{
			FromLinter: "gosmell",
			Text:       "longmethod: CS101 (78% confidence): BigOne() spans 14 lines.",
			Pos: struct {
				Filename string `json:"Filename"`
				Line     int    `json:"Line"`
			}{Filename: "pkg/a.go", Line: 5},
		},
		struct {
			FromLinter string `json:"FromLinter"`
			Text       string `json:"Text"`
			Pos        struct {
				Filename string `json:"Filename"`
				Line     int    `json:"Line"`
			} `json:"Pos"`
		}{
			FromLinter: "govet",
			Text:       "some unrelated issue",
			Pos: struct {
				Filename string `json:"Filename"`
				Line     int    `json:"Line"`
			}{Filename: "pkg/a.go", Line: 1},
		},
	)

	got := buildFileFindings(issues, "/does/not/matter")

	findings, ok := got["pkg/a.go"]
	if !ok {
		t.Fatalf("expected findings for pkg/a.go, got %+v", got)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding (govet issue must be filtered out), got %d", len(findings))
	}
	if findings[0].meta.Code != "CS101" {
		t.Errorf("Code = %q, want CS101", findings[0].meta.Code)
	}
	if findings[0].confidence != 78 {
		t.Errorf("confidence = %d, want 78", findings[0].confidence)
	}
	if findings[0].line != 5 {
		t.Errorf("line = %d, want 5", findings[0].line)
	}
}

func TestResolveLintBinary_EnvVarTakesPriority(t *testing.T) {
	t.Setenv("GOSMELL_LINT_BIN", "/custom/path/to/custom-gcl")

	got, err := resolveLintBinary()
	if err != nil {
		t.Fatalf("resolveLintBinary() error = %v", err)
	}
	if got != "/custom/path/to/custom-gcl" {
		t.Errorf("resolveLintBinary() = %q, want the GOSMELL_LINT_BIN value", got)
	}
}

func TestUserSettingsBlock_NoFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()

	got, err := userSettingsBlock(dir)
	if err != nil {
		t.Fatalf("userSettingsBlock() error = %v", err)
	}
	if got != "" {
		t.Errorf("userSettingsBlock() with no file = %q, want empty", got)
	}
}

func TestUserSettingsBlock_ReindentsUserFile(t *testing.T) {
	dir := t.TempDir()
	content := "longMethod:\n  maxStatements: 60\n"
	if err := os.WriteFile(filepath.Join(dir, userSettingsFile), []byte(content), 0o644); err != nil {
		t.Fatalf("writing user settings file: %v", err)
	}

	got, err := userSettingsBlock(dir)
	if err != nil {
		t.Fatalf("userSettingsBlock() error = %v", err)
	}

	want := "        settings:\n          longMethod:\n            maxStatements: 60\n"
	if got != want {
		t.Errorf("userSettingsBlock() =\n%q\nwant\n%q", got, want)
	}
}
