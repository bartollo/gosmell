package rules

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

func reportFindings(pass *analysis.Pass, findings []report.Finding) {
	for _, f := range findings {
		pass.Reportf(f.Pos, "%s (%d%% confidence): %s", f.Code, f.Confidence, f.Message)
	}
}

func newAnalyzer(name, doc string, collect func(*analysis.Pass) []report.Finding) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: name,
		Doc:  doc,
		Run: func(pass *analysis.Pass) (any, error) {
			reportFindings(pass, collect(pass))
			return nil, nil
		},
	}
}

func countStatements(n ast.Node) int {
	count := 0
	ast.Inspect(n, func(n ast.Node) bool {
		if _, ok := n.(ast.Stmt); ok {
			count++
		}
		return true
	})
	return count
}

func cyclomaticComplexity(n ast.Node) int {
	complexity := 1
	ast.Inspect(n, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt:
			complexity++
		case *ast.ForStmt:
			complexity++
		case *ast.RangeStmt:
			complexity++
		case *ast.CaseClause:
			complexity++
		case *ast.CommClause:
			complexity++
		case *ast.BinaryExpr:
			if node.Op == token.LAND || node.Op == token.LOR {
				complexity++
			}
		}
		return true
	})
	return complexity
}

func callCount(n ast.Node) int {
	count := 0
	ast.Inspect(n, func(n ast.Node) bool {
		if _, ok := n.(*ast.CallExpr); ok {
			count++
		}
		return true
	})
	return count
}

func lineSpan(fset *token.FileSet, start, end token.Pos) int {
	return fset.Position(end).Line - fset.Position(start).Line + 1
}

func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return receiverTypeName(t.X)
	case *ast.IndexListExpr:
		return receiverTypeName(t.X)
	default:
		return ""
	}
}
