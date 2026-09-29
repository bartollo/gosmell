package rules

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

var DeadCodeAnalyzer = &analysis.Analyzer{
	Name: "deadcode",
	Doc:  "CS105: detects unexported functions and methods with no visible reference anywhere in the declaring package",
	Run:  runDeadCode,
}

const deadCodeConfidence = 80

func runDeadCode(pass *analysis.Pass) (any, error) {
	reportFindings(pass, CollectDeadCodeFindings(pass))
	return nil, nil
}

func CollectDeadCodeFindings(pass *analysis.Pass) []report.Finding {
	declared := collectUnexportedDeclarations(pass)
	if len(declared) == 0 {
		return nil
	}

	refCount := countIdentifierReferences(pass, declared)

	var findings []report.Finding
	for obj, fn := range declared {
		if refCount[obj] > 0 {
			continue
		}
		findings = append(findings, report.Finding{
			Pos:        fn.Pos(),
			Code:       "CS105",
			Confidence: deadCodeConfidence,
			Message: fmt.Sprintf(
				"unexported %s has no visible reference anywhere in this package; consider removing it.", fn.Name.Name),
		})
	}

	return findings
}

func collectUnexportedDeclarations(pass *analysis.Pass) map[types.Object]*ast.FuncDecl {
	declared := map[types.Object]*ast.FuncDecl{}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.IsExported() || fn.Name.Name == "init" || fn.Name.Name == "main" {
				continue
			}
			obj := pass.TypesInfo.Defs[fn.Name]
			if obj == nil {
				continue
			}
			declared[obj] = fn
		}
	}

	return declared
}

func countIdentifierReferences(pass *analysis.Pass, declared map[types.Object]*ast.FuncDecl) map[types.Object]int {
	refCount := map[types.Object]int{}
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			obj := pass.TypesInfo.Uses[id]
			if obj == nil {
				return true
			}
			if _, isDeclared := declared[obj]; isDeclared {
				refCount[obj]++
			}
			return true
		})
	}
	return refCount
}
