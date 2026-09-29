package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

var SwallowedErrorAnalyzer = &analysis.Analyzer{
	Name: "swallowederror",
	Doc:  "CS107: detects catch blocks that neither rethrow, report, log nor otherwise react to the failure they caught",
	Run:  runSwallowedError,
}

const (
	swallowedErrorEmptyBlockConfidence = 90
	swallowedErrorDiscardConfidence    = 85
)

func runSwallowedError(pass *analysis.Pass) (any, error) {
	reportFindings(pass, CollectSwallowedErrorFindings(pass))
	return nil, nil
}

func CollectSwallowedErrorFindings(pass *analysis.Pass) []report.Finding {
	var findings []report.Finding
	errType := types.Universe.Lookup("error").Type()

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch stmt := n.(type) {
			case *ast.IfStmt:
				if isErrNilCheck(pass, stmt.Cond, errType) && len(stmt.Body.List) == 0 {
					findings = append(findings, report.Finding{
						Pos:        stmt.Pos(),
						Code:       "CS107",
						Confidence: swallowedErrorEmptyBlockConfidence,
						Message:    "error is checked but the block does nothing with it; log, wrap or return the failure.",
					})
				}
			case *ast.AssignStmt:
				if isDiscardedError(pass, stmt, errType) {
					findings = append(findings, report.Finding{
						Pos:        stmt.Pos(),
						Code:       "CS107",
						Confidence: swallowedErrorDiscardConfidence,
						Message:    "error is explicitly discarded; log, wrap or return the failure.",
					})
				}
			}
			return true
		})
	}

	return findings
}

func isErrNilCheck(pass *analysis.Pass, cond ast.Expr, errType types.Type) bool {
	bin, ok := cond.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return false
	}

	var ident *ast.Ident
	switch {
	case isNilIdent(bin.Y):
		ident, ok = bin.X.(*ast.Ident)
	case isNilIdent(bin.X):
		ident, ok = bin.Y.(*ast.Ident)
	default:
		ok = false
	}
	if !ok || ident == nil {
		return false
	}

	t := pass.TypesInfo.TypeOf(ident)
	return t != nil && types.Identical(t, errType)
}

func isNilIdent(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "nil"
}

func isDiscardedError(pass *analysis.Pass, assign *ast.AssignStmt, errType types.Type) bool {
	if len(assign.Rhs) != 1 {
		return false
	}

	if len(assign.Lhs) == 1 {
		return isBlankIdent(assign.Lhs[0]) && exprIsError(pass, assign.Rhs[0], errType)
	}

	call, ok := assign.Rhs[0].(*ast.CallExpr)
	if !ok {
		return false
	}
	return discardsErrorResult(pass, assign.Lhs, call, errType)
}

func isBlankIdent(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "_"
}

func exprIsError(pass *analysis.Pass, e ast.Expr, errType types.Type) bool {
	t := pass.TypesInfo.TypeOf(e)
	return t != nil && types.Identical(t, errType)
}

func discardsErrorResult(pass *analysis.Pass, lhs []ast.Expr, call *ast.CallExpr, errType types.Type) bool {
	tuple, ok := pass.TypesInfo.TypeOf(call).(*types.Tuple)
	if !ok || tuple.Len() != len(lhs) {
		return false
	}
	for i, l := range lhs {
		if isBlankIdent(l) && types.Identical(tuple.At(i).Type(), errType) {
			return true
		}
	}
	return false
}
