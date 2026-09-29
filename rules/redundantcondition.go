package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

var RedundantConditionAnalyzer = &analysis.Analyzer{
	Name: "redundantcondition",
	Doc:  "CS108: detects conditions re-tested inside themselves, repeated within an if/elseif chain, duplicated across a boolean operator, or written as a literal true/false",
	Run:  runRedundantCondition,
}

const (
	redundantLiteralBoolConfidence      = 95
	redundantDuplicateOperandConfidence = 95
	redundantChainConfidence            = 90
	redundantSelfNestedConfidence       = 90
)

func runRedundantCondition(pass *analysis.Pass) (any, error) {
	reportFindings(pass, CollectRedundantConditionFindings(pass))
	return nil, nil
}

func CollectRedundantConditionFindings(pass *analysis.Pass) []report.Finding {
	var findings []report.Finding
	for _, file := range pass.Files {
		findings = append(findings, redundantConditionFindingsInFile(file)...)
	}
	return findings
}

func redundantConditionFindingsInFile(file *ast.File) []report.Finding {
	var findings []report.Finding
	elseIfPositions := collectElseIfPositions(file)

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BinaryExpr:
			if f := literalBoolFinding(node); f != nil {
				findings = append(findings, *f)
			}
			if f := duplicateOperandFinding(node); f != nil {
				findings = append(findings, *f)
			}
		case *ast.IfStmt:
			if isBoolLit(node.Cond) {
				findings = append(findings, report.Finding{
					Pos:        node.Cond.Pos(),
					Code:       "CS108",
					Confidence: redundantLiteralBoolConfidence,
					Message:    "condition is a literal boolean; simplify or remove the branch.",
				})
			}
			if f := selfNestedFinding(node); f != nil {
				findings = append(findings, *f)
			}
			if !elseIfPositions[node.Pos()] {
				findings = append(findings, chainDuplicateFindings(node)...)
			}
		}
		return true
	})

	return findings
}

func collectElseIfPositions(file *ast.File) map[token.Pos]bool {
	elseIfPositions := map[token.Pos]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		if elseIf, ok := ifStmt.Else.(*ast.IfStmt); ok {
			elseIfPositions[elseIf.Pos()] = true
		}
		return true
	})
	return elseIfPositions
}

func literalBoolFinding(bin *ast.BinaryExpr) *report.Finding {
	if bin.Op != token.EQL && bin.Op != token.NEQ {
		return nil
	}
	if !isBoolLit(bin.X) && !isBoolLit(bin.Y) {
		return nil
	}
	return &report.Finding{
		Pos:        bin.Pos(),
		Code:       "CS108",
		Confidence: redundantLiteralBoolConfidence,
		Message:    "comparison against a literal boolean is redundant; use the expression directly.",
	}
}

func isBoolLit(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && (id.Name == "true" || id.Name == "false")
}

func duplicateOperandFinding(bin *ast.BinaryExpr) *report.Finding {
	if bin.Op != token.LAND && bin.Op != token.LOR {
		return nil
	}
	if types.ExprString(bin.X) != types.ExprString(bin.Y) {
		return nil
	}
	return &report.Finding{
		Pos:        bin.Pos(),
		Code:       "CS108",
		Confidence: redundantDuplicateOperandConfidence,
		Message: fmt.Sprintf(
			"both sides of the boolean operator are identical (%s); condition is redundant.",
			types.ExprString(bin.X)),
	}
}

func chainDuplicateFindings(ifStmt *ast.IfStmt) []report.Finding {
	var findings []report.Finding
	seen := map[string]bool{types.ExprString(ifStmt.Cond): true}

	cur := ifStmt.Else
	for {
		elseIf, ok := cur.(*ast.IfStmt)
		if !ok {
			return findings
		}
		cond := types.ExprString(elseIf.Cond)
		if seen[cond] && !declaresConditionIdent(elseIf, guardIdentifiers(elseIf.Cond)) {
			findings = append(findings, report.Finding{
				Pos:        elseIf.Cond.Pos(),
				Code:       "CS108",
				Confidence: redundantChainConfidence,
				Message: fmt.Sprintf(
					"condition %s repeats an earlier branch in this if/elseif chain.", cond),
			})
		}
		seen[cond] = true
		cur = elseIf.Else
	}
}

func selfNestedFinding(ifStmt *ast.IfStmt) *report.Finding {
	if len(ifStmt.Body.List) != 1 {
		return nil
	}
	inner, ok := ifStmt.Body.List[0].(*ast.IfStmt)
	if !ok {
		return nil
	}
	if types.ExprString(inner.Cond) != types.ExprString(ifStmt.Cond) {
		return nil
	}
	if declaresConditionIdent(inner, guardIdentifiers(inner.Cond)) {
		return nil
	}
	return &report.Finding{
		Pos:        inner.Pos(),
		Code:       "CS108",
		Confidence: redundantSelfNestedConfidence,
		Message: fmt.Sprintf(
			"condition %s is re-tested immediately inside itself.", types.ExprString(inner.Cond)),
	}
}
