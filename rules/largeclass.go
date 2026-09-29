package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

func NewLargeClassAnalyzer(cfg Config) *analysis.Analyzer {
	return newAnalyzer("largeclass",
		"CS102: detects structs that are large across several dimensions at once: size, method count, injected dependencies and collaborators",
		func(pass *analysis.Pass) []report.Finding {
			return CollectLargeClassFindings(pass, cfg.LargeClass)
		})
}

type largeClassMetrics struct {
	pos                  token.Pos
	end                  token.Pos
	fieldCount           int
	methodCount          int
	publicMethodCount    int
	statementCount       int
	collaborators        map[string]struct{}
	injectedDependencies map[string]struct{}
}

func CollectLargeClassFindings(pass *analysis.Pass, cfg LargeClassConfig) []report.Finding {
	classes := collectStructFieldCounts(pass)
	collectMethodMetrics(pass, classes)
	collectInjectedDependencies(pass, classes)
	return buildLargeClassFindings(pass, classes, cfg)
}

func collectStructFieldCounts(pass *analysis.Pass) map[string]*largeClassMetrics {
	classes := map[string]*largeClassMetrics{}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				recordStructFieldCount(classes, spec)
			}
		}
	}

	return classes
}

func recordStructFieldCount(classes map[string]*largeClassMetrics, spec ast.Spec) {
	ts, ok := spec.(*ast.TypeSpec)
	if !ok {
		return
	}
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		return
	}

	m := classes[ts.Name.Name]
	if m == nil {
		m = &largeClassMetrics{
			pos:                  ts.Pos(),
			end:                  ts.End(),
			collaborators:        map[string]struct{}{},
			injectedDependencies: map[string]struct{}{},
		}
		classes[ts.Name.Name] = m
	}
	m.fieldCount += countStructFields(st)
}

func collectMethodMetrics(pass *analysis.Pass, classes map[string]*largeClassMetrics) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 || fn.Body == nil {
				continue
			}
			m := classes[receiverTypeName(fn.Recv.List[0].Type)]
			if m == nil {
				continue
			}
			m.methodCount++
			if fn.Name.IsExported() {
				m.publicMethodCount++
			}
			m.statementCount += countStatements(fn.Body)
			collectCollaborators(fn.Body, m.collaborators)
			if end := fn.End(); end > m.end {
				m.end = end
			}
		}
	}
}

func collectInjectedDependencies(pass *analysis.Pass, classes map[string]*largeClassMetrics) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Body == nil || !strings.HasPrefix(fn.Name.Name, "New") {
				continue
			}
			recordConstructorDependencies(fn.Body, classes)
		}
	}
}

func recordConstructorDependencies(body *ast.BlockStmt, classes map[string]*largeClassMetrics) {
	ast.Inspect(body, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		m := classes[compositeLitTypeName(lit)]
		if m == nil {
			return true
		}
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if key, ok := kv.Key.(*ast.Ident); ok {
				m.injectedDependencies[key.Name] = struct{}{}
			}
		}
		return true
	})
}

func buildLargeClassFindings(pass *analysis.Pass, classes map[string]*largeClassMetrics, cfg LargeClassConfig) []report.Finding {
	var findings []report.Finding

	for name, m := range classes {
		if countExceededLargeClassDimensions(m, cfg) < cfg.MinExceeded {
			continue
		}

		dependencyClause := ""
		if n := len(m.injectedDependencies); n > 0 {
			dependencyClause = fmt.Sprintf(", %d injected dependencies", n)
		}

		findings = append(findings, report.Finding{
			Pos:        m.pos,
			Code:       "CS102",
			Confidence: largeClassConfidence(m, cfg),
			Message: fmt.Sprintf(
				"%s spans %d lines with %d methods (%d public), %d statements%s and talks to %d distinct collaborators.",
				name, lineSpan(pass.Fset, m.pos, m.end), m.methodCount, m.publicMethodCount,
				m.statementCount, dependencyClause, len(m.collaborators)),
		})
	}

	return findings
}

func countExceededLargeClassDimensions(m *largeClassMetrics, cfg LargeClassConfig) int {
	exceeded := 0
	if m.fieldCount > cfg.MaxFields {
		exceeded++
	}
	if m.methodCount > cfg.MaxMethods {
		exceeded++
	}
	if len(m.collaborators) > cfg.MaxCollaborators {
		exceeded++
	}
	if m.statementCount > cfg.MaxStatements {
		exceeded++
	}
	return exceeded
}

func largeClassConfidence(m *largeClassMetrics, cfg LargeClassConfig) int {
	return report.ConfidenceFromExcess(
		float64(m.fieldCount)/float64(cfg.MaxFields),
		float64(m.methodCount)/float64(cfg.MaxMethods),
		float64(len(m.collaborators))/float64(cfg.MaxCollaborators),
		float64(m.statementCount)/float64(cfg.MaxStatements),
	)
}

func countStructFields(st *ast.StructType) int {
	if st.Fields == nil {
		return 0
	}
	count := 0
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			count++
			continue
		}
		count += len(f.Names)
	}
	return count
}

func collectCollaborators(body *ast.BlockStmt, out map[string]struct{}) {
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch x := sel.X.(type) {
		case *ast.Ident:
			out[x.Name] = struct{}{}
		case *ast.SelectorExpr:
			out[x.Sel.Name] = struct{}{}
		}
		return true
	})
}
