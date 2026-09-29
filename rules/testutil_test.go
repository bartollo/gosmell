package rules

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"

	"github.com/bartollo/gosmell/report"
)

func loadPass(t *testing.T, src string) *analysis.Pass {
	t.Helper()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module testpkg\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dir, "pkg.go"), src)

	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("loading test package: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	for _, e := range pkgs[0].Errors {
		t.Fatalf("package error: %v", e)
	}

	return &analysis.Pass{
		Fset:      pkgs[0].Fset,
		Files:     pkgs[0].Syntax,
		Pkg:       pkgs[0].Types,
		TypesInfo: pkgs[0].TypesInfo,
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func requireFindingCount(t *testing.T, findings []report.Finding, want int) {
	t.Helper()
	if len(findings) != want {
		t.Fatalf("expected %d finding(s), got %d: %+v", want, len(findings), findings)
	}
}

func requireCode(t *testing.T, findings []report.Finding, want string) {
	t.Helper()
	for _, f := range findings {
		if f.Code != want {
			t.Errorf("Code = %q, want %q", f.Code, want)
		}
	}
}

func requireSingleFinding(t *testing.T, findings []report.Finding, wantCode string) {
	t.Helper()
	requireFindingCount(t, findings, 1)
	requireCode(t, findings, wantCode)
}
