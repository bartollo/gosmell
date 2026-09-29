package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

var UnusedConstructorDependencyAnalyzer = &analysis.Analyzer{
	Name: "unusedctordep",
	Doc:  "CS106: detects constructor-injected dependencies that are never read anywhere in the class",
	Run:  runUnusedConstructorDependency,
}

const unusedCtorDepConfidence = 75

type ctorDependency struct {
	structName string
	fieldName  string
	pos        token.Pos
}

func runUnusedConstructorDependency(pass *analysis.Pass) (any, error) {
	reportFindings(pass, CollectUnusedConstructorDependencyFindings(pass))
	return nil, nil
}

func CollectUnusedConstructorDependencyFindings(pass *analysis.Pass) []report.Finding {
	deps := collectConstructorDependencies(pass)
	if len(deps) == 0 {
		return nil
	}

	writePositions := collectFieldWritePositions(pass)
	readFields := collectFieldReads(pass, writePositions)

	var findings []report.Finding
	for _, dep := range deps {
		if readFields[dep.structName+"."+dep.fieldName] {
			continue
		}
		findings = append(findings, report.Finding{
			Pos:        dep.pos,
			Code:       "CS106",
			Confidence: unusedCtorDepConfidence,
			Message: fmt.Sprintf(
				"dependency %s injected in constructor is never read in %s; consider removing it.",
				dep.fieldName, dep.structName),
		})
	}

	return findings
}

func collectConstructorDependencies(pass *analysis.Pass) []ctorDependency {
	var deps []ctorDependency

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Body == nil || !strings.HasPrefix(fn.Name.Name, "New") {
				continue
			}
			params := constructorParamNames(fn)
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				deps = append(deps, constructorDepsFromCompositeLit(n, params)...)
				return true
			})
		}
	}

	return deps
}

func constructorParamNames(fn *ast.FuncDecl) map[string]bool {
	names := map[string]bool{}
	if fn.Type.Params == nil {
		return names
	}
	for _, field := range fn.Type.Params.List {
		for _, name := range field.Names {
			names[name.Name] = true
		}
	}
	return names
}

func constructorDepsFromCompositeLit(n ast.Node, params map[string]bool) []ctorDependency {
	lit, ok := n.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	structName := compositeLitTypeName(lit)
	if structName == "" {
		return nil
	}

	var deps []ctorDependency
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		valueIdent, ok := kv.Value.(*ast.Ident)
		if !ok || !params[valueIdent.Name] {
			continue
		}
		deps = append(deps, ctorDependency{structName: structName, fieldName: key.Name, pos: key.Pos()})
	}
	return deps
}

func collectFieldWritePositions(pass *analysis.Pass) map[token.Pos]bool {
	writePositions := map[token.Pos]bool{}
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)

			if !ok || assign.Tok != token.ASSIGN {
				return true
			}
			for _, lhs := range assign.Lhs {
				if sel, ok := lhs.(*ast.SelectorExpr); ok {
					writePositions[sel.Pos()] = true
				}
			}
			return true
		})
	}
	return writePositions
}

func collectFieldReads(pass *analysis.Pass, writePositions map[token.Pos]bool) map[string]bool {
	readFields := map[string]bool{}
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok || writePositions[sel.Pos()] {
				return true
			}
			structName := selectorReceiverStructName(pass, sel.X)
			if structName == "" {
				return true
			}
			readFields[structName+"."+sel.Sel.Name] = true
			return true
		})
	}
	return readFields
}

func compositeLitTypeName(lit *ast.CompositeLit) string {
	switch t := lit.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return t.Sel.Name
	default:
		return ""
	}
}

func selectorReceiverStructName(pass *analysis.Pass, expr ast.Expr) string {
	if pass.TypesInfo == nil {
		return ""
	}
	typ := pass.TypesInfo.TypeOf(expr)
	if typ == nil {
		return ""
	}
	if ptr, ok := typ.(*types.Pointer); ok {
		typ = ptr.Elem()
	}
	named, ok := typ.(*types.Named)
	if !ok {
		return ""
	}
	return named.Obj().Name()
}
