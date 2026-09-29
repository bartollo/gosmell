package rules

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

func NewDuplicateCodeAnalyzer(cfg Config) *analysis.Analyzer {
	return newAnalyzer("duplicatecode",
		"CS104: detects methods whose bodies are structurally identical to another method in the project",
		func(pass *analysis.Pass) []report.Finding {
			return CollectDuplicateCodeFindings(pass, cfg.DuplicateCode)
		})
}

type duplicateCandidate struct {
	name       string
	pos        token.Pos
	statements int
}

func CollectDuplicateCodeFindings(pass *analysis.Pass, cfg DuplicateCodeConfig) []report.Finding {
	groups := map[string][]duplicateCandidate{}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			statements := countStatements(fn.Body)
			if statements < cfg.MinStatements {
				continue
			}
			shape := structuralShape(fn.Body)
			groups[shape] = append(groups[shape], duplicateCandidate{name: fn.Name.Name, pos: fn.Pos(), statements: statements})
		}
	}

	var findings []report.Finding

	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		original := group[0]
		for _, dup := range group[1:] {
			findings = append(findings, report.Finding{
				Pos:        dup.pos,
				Code:       "CS104",
				Confidence: report.ConfidenceFromExcess(float64(dup.statements) / float64(cfg.MinStatements)),
				Message: fmt.Sprintf(
					"%s is structurally identical to %s; consider extracting the shared logic.",
					dup.name, original.name),
			})
		}
	}

	return findings
}

func structuralShape(n ast.Node) string {
	shape := ""
	ast.Inspect(n, func(node ast.Node) bool {
		if node == nil {
			return false
		}
		shape += fmt.Sprintf("%T|", node)
		if bin, ok := node.(*ast.BinaryExpr); ok {
			shape += bin.Op.String() + "|"
		}
		return true
	})
	return shape
}
