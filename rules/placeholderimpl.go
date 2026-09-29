package rules

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

var PlaceholderImplementationAnalyzer = &analysis.Analyzer{
	Name: "placeholderimpl",
	Doc:  "CS112: detects bodies that say they are unfinished: elided-code comments, not-implemented panics/errors, or a TODO over an empty/constant return",
	Run:  runPlaceholderImplementation,
}

const (
	placeholderElisionConfidence        = 70
	placeholderNotImplementedConfidence = 95
	placeholderTodoReturnConfidence     = 80
)

var elisionMarkers = []string{"existing code", "rest of the code", "..."}

func runPlaceholderImplementation(pass *analysis.Pass) (any, error) {
	reportFindings(pass, CollectPlaceholderImplementationFindings(pass))
	return nil, nil
}

func CollectPlaceholderImplementationFindings(pass *analysis.Pass) []report.Finding {
	var findings []report.Finding

	for _, file := range pass.Files {
		findings = append(findings, elisionCommentFindings(file)...)
		findings = append(findings, notImplementedBodyFindings(file)...)
		findings = append(findings, todoOverReturnFindings(pass, file)...)
	}

	return findings
}

func elisionCommentFindings(file *ast.File) []report.Finding {
	var findings []report.Finding
	for _, group := range file.Comments {
		text := strings.ToLower(group.Text())
		for _, marker := range elisionMarkers {
			if strings.Contains(text, marker) {
				findings = append(findings, report.Finding{
					Pos:        group.Pos(),
					Code:       "CS112",
					Confidence: placeholderElisionConfidence,
					Message:    "comment suggests elided/unfinished code; make sure this isn't a placeholder left behind.",
				})
				break
			}
		}
	}
	return findings
}

func notImplementedBodyFindings(file *ast.File) []report.Finding {
	var findings []report.Finding
	ast.Inspect(file, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		if isNotImplementedBody(fn.Body) {
			findings = append(findings, report.Finding{
				Pos:        fn.Pos(),
				Code:       "CS112",
				Confidence: placeholderNotImplementedConfidence,
				Message: fmt.Sprintf(
					"function %s only signals it is not implemented; finish it or remove it.", fn.Name.Name),
			})
		}
		return true
	})
	return findings
}

func todoOverReturnFindings(pass *analysis.Pass, file *ast.File) []report.Finding {
	var findings []report.Finding
	cmap := ast.NewCommentMap(pass.Fset, file, file.Comments)
	ast.Inspect(file, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for _, stmt := range block.List {
			ret, ok := stmt.(*ast.ReturnStmt)
			if !ok || !isEmptyOrConstantReturn(ret) || !hasTodoComment(cmap[stmt]) {
				continue
			}
			findings = append(findings, report.Finding{
				Pos:        ret.Pos(),
				Code:       "CS112",
				Confidence: placeholderTodoReturnConfidence,
				Message:    "TODO left over an empty/constant return; the implementation looks unfinished.",
			})
		}
		return true
	})
	return findings
}

func isNotImplementedBody(body *ast.BlockStmt) bool {
	if len(body.List) != 1 {
		return false
	}
	switch stmt := body.List[0].(type) {
	case *ast.ExprStmt:
		return isPanicNotImplemented(stmt.X)
	case *ast.ReturnStmt:
		for _, r := range stmt.Results {
			if isErrorConstructorNotImplemented(r) {
				return true
			}
		}
	}
	return false
}

func isPanicNotImplemented(e ast.Expr) bool {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return false
	}
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "panic" || len(call.Args) != 1 {
		return false
	}
	return containsNotImplemented(call.Args[0])
}

func isErrorConstructorNotImplemented(e ast.Expr) bool {
	call, ok := e.(*ast.CallExpr)
	if !ok || len(call.Args) == 0 {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "New" && sel.Sel.Name != "Errorf" {
		return false
	}
	return containsNotImplemented(call.Args[0])
}

func containsNotImplemented(e ast.Expr) bool {
	lit, ok := e.(*ast.BasicLit)
	if !ok {
		return false
	}
	return strings.Contains(strings.ToLower(lit.Value), "not implemented")
}

func isEmptyOrConstantReturn(ret *ast.ReturnStmt) bool {
	if len(ret.Results) != 1 {
		return len(ret.Results) == 0
	}
	switch v := ret.Results[0].(type) {
	case *ast.Ident:
		return v.Name == "nil"
	case *ast.BasicLit:
		return true
	case *ast.CompositeLit:
		return len(v.Elts) == 0
	default:
		return false
	}
}

func hasTodoComment(groups []*ast.CommentGroup) bool {
	for _, g := range groups {
		text := strings.ToLower(g.Text())
		if strings.Contains(text, "todo") || strings.Contains(text, "fixme") {
			return true
		}
	}
	return false
}
