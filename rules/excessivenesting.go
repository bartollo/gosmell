package rules

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

func NewExcessiveNestingAnalyzer(cfg Config) *analysis.Analyzer {
	return newAnalyzer("excessivenesting",
		"CS103: detects methods whose conditionals, loops and blocks nest deeper than the configured limit",
		func(pass *analysis.Pass) []report.Finding {
			return CollectExcessiveNestingFindings(pass, cfg.ExcessiveNesting)
		})
}

func CollectExcessiveNestingFindings(pass *analysis.Pass, cfg ExcessiveNestingConfig) []report.Finding {
	var findings []report.Finding

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				return true
			}

			depth := nestingDepth(fn.Body, 0)
			if depth > cfg.MaxDepth {
				findings = append(findings, report.Finding{
					Pos:        fn.Pos(),
					Code:       "CS103",
					Confidence: report.ConfidenceFromExcess(float64(depth) / float64(cfg.MaxDepth)),
					Message: fmt.Sprintf(
						"function %s nests %d levels deep (limit %d); consider early returns or extracting methods.",
						fn.Name.Name, depth, cfg.MaxDepth),
				})
			}

			return false
		})
	}

	return findings
}

func nestingDepth(n ast.Node, current int) int {
	switch stmt := n.(type) {
	case *ast.BlockStmt:
		return maxChildDepth(stmt.List, current)
	case *ast.IfStmt:
		return maxOf(nestingDepth(stmt.Body, current+1), ifElseDepth(stmt, current))
	case *ast.CaseClause:
		return maxChildDepth(stmt.Body, current)
	case *ast.CommClause:
		return maxChildDepth(stmt.Body, current)
	default:
		if body, ok := loopOrBranchBody(n); ok {
			return nestingDepth(body, current+1)
		}
		return current
	}
}

func ifElseDepth(stmt *ast.IfStmt, current int) int {
	if stmt.Else == nil {
		return current
	}
	elseDepth := current + 1
	if _, isElseIf := stmt.Else.(*ast.IfStmt); isElseIf {
		elseDepth = current
	}
	return nestingDepth(stmt.Else, elseDepth)
}

func loopOrBranchBody(n ast.Node) (*ast.BlockStmt, bool) {
	switch stmt := n.(type) {
	case *ast.ForStmt:
		return stmt.Body, true
	case *ast.RangeStmt:
		return stmt.Body, true
	case *ast.SwitchStmt:
		return stmt.Body, true
	case *ast.TypeSwitchStmt:
		return stmt.Body, true
	case *ast.SelectStmt:
		return stmt.Body, true
	default:
		return nil, false
	}
}

func maxChildDepth(stmts []ast.Stmt, current int) int {
	max := current
	for _, s := range stmts {
		if d := nestingDepth(s, current); d > max {
			max = d
		}
	}
	return max
}

func maxOf(a, b int) int {
	if a > b {
		return a
	}
	return b
}
