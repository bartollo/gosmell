package rules

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

func NewLongMethodAnalyzer(cfg Config) *analysis.Analyzer {
	return newAnalyzer("longmethod",
		"CS101: detects methods that are excessively long, branchy or chatty across several independent measurements",
		func(pass *analysis.Pass) []report.Finding {
			return CollectLongMethodFindings(pass, cfg.LongMethod)
		})
}

type longMethodMetrics struct {
	lines         int
	statements    int
	complexity    int
	nesting       int
	calls         int
	collaborators int
}

func CollectLongMethodFindings(pass *analysis.Pass, cfg LongMethodConfig) []report.Finding {
	var findings []report.Finding

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			m := measureLongMethod(pass, fn)
			if !exceedsLongMethodThresholds(m, cfg) {
				continue
			}

			findings = append(findings, report.Finding{
				Pos:        fn.Pos(),
				Code:       "CS101",
				Confidence: longMethodConfidence(m, cfg),
				Message: fmt.Sprintf(
					"%s spans %d lines with %d statements, cyclomatic complexity %d, nesting depth %d and %d calls to %d distinct collaborators.",
					longMethodLabel(fn), m.lines, m.statements, m.complexity, m.nesting, m.calls, m.collaborators),
			})
		}
	}

	return findings
}

func measureLongMethod(pass *analysis.Pass, fn *ast.FuncDecl) longMethodMetrics {
	collaborators := map[string]struct{}{}
	collectCollaborators(fn.Body, collaborators)

	return longMethodMetrics{
		lines:         lineSpan(pass.Fset, fn.Pos(), fn.End()),
		statements:    countStatements(fn.Body),
		complexity:    cyclomaticComplexity(fn.Body),
		nesting:       nestingDepth(fn.Body, 0),
		calls:         callCount(fn.Body),
		collaborators: len(collaborators),
	}
}

func exceedsLongMethodThresholds(m longMethodMetrics, cfg LongMethodConfig) bool {
	return m.statements > cfg.MaxStatements ||
		m.complexity > cfg.MaxComplexity ||
		m.nesting > cfg.MaxNesting ||
		m.calls > cfg.MaxCalls ||
		m.collaborators > cfg.MaxCollaborators
}

func longMethodConfidence(m longMethodMetrics, cfg LongMethodConfig) int {
	return report.ConfidenceFromExcess(
		float64(m.statements)/float64(cfg.MaxStatements),
		float64(m.complexity)/float64(cfg.MaxComplexity),
		float64(m.nesting)/float64(cfg.MaxNesting),
		float64(m.calls)/float64(cfg.MaxCalls),
		float64(m.collaborators)/float64(cfg.MaxCollaborators),
	)
}

func longMethodLabel(fn *ast.FuncDecl) string {
	label := fn.Name.Name + "()"
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		if recv := receiverTypeName(fn.Recv.List[0].Type); recv != "" {
			label = recv + "::" + label
		}
	}
	return label
}
