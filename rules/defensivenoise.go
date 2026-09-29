package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

var DefensiveNoiseAnalyzer = &analysis.Analyzer{
	Name: "defensivenoise",
	Doc:  "CS110: detects a method that guards the same subject the same way twice, with the same outcome and no reassignment in between",
	Run:  runDefensiveNoise,
}

const defensiveNoiseConfidence = 75

func runDefensiveNoise(pass *analysis.Pass) (any, error) {
	reportFindings(pass, CollectDefensiveNoiseFindings(pass))
	return nil, nil
}

func CollectDefensiveNoiseFindings(pass *analysis.Pass) []report.Finding {
	var findings []report.Finding

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			block, ok := n.(*ast.BlockStmt)
			if !ok {
				return true
			}
			findings = append(findings, checkGuardsInBlock(block)...)
			return true
		})
	}

	return findings
}

type seenGuard struct {
	cond string
	idx  int
}

func checkGuardsInBlock(block *ast.BlockStmt) []report.Finding {
	var findings []report.Finding
	var seen []seenGuard

	for i, stmt := range block.List {
		ifStmt, ok := stmt.(*ast.IfStmt)
		if !ok {
			continue
		}

		cond := types.ExprString(ifStmt.Cond)
		ids := guardIdentifiers(ifStmt.Cond)

		if !declaresConditionIdent(ifStmt, ids) {
			findings = append(findings, findRepeatedGuards(block, seen, i, ifStmt, cond, ids)...)
		}

		seen = append(seen, seenGuard{cond: cond, idx: i})
	}

	return findings
}

func findRepeatedGuards(block *ast.BlockStmt, seen []seenGuard, i int, ifStmt *ast.IfStmt, cond string, ids map[string]bool) []report.Finding {
	for _, prior := range seen {

		if prior.cond != cond || reassignsAny(block.List[prior.idx:i], ids) {
			continue
		}
		return []report.Finding{{
			Pos:        ifStmt.Pos(),
			Code:       "CS110",
			Confidence: defensiveNoiseConfidence,
			Message: fmt.Sprintf(
				"condition %s is guarded again with no reassignment since the earlier check; remove the redundant guard.",
				cond),
		}}
	}
	return nil
}

func declaresConditionIdent(ifStmt *ast.IfStmt, ids map[string]bool) bool {
	assign, ok := ifStmt.Init.(*ast.AssignStmt)
	if !ok || assign.Tok != token.DEFINE {
		return false
	}
	for _, lhs := range assign.Lhs {
		if id, ok := lhs.(*ast.Ident); ok && ids[id.Name] {
			return true
		}
	}
	return false
}

func guardIdentifiers(e ast.Expr) map[string]bool {
	names := map[string]bool{}
	ast.Inspect(e, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			names[id.Name] = true
		}
		return true
	})
	return names
}

func reassignsAny(stmts []ast.Stmt, ids map[string]bool) bool {
	found := false
	for _, stmt := range stmts {
		ast.Inspect(stmt, func(n ast.Node) bool {
			switch s := n.(type) {
			case *ast.AssignStmt:
				for _, lhs := range s.Lhs {
					if id, ok := lhs.(*ast.Ident); ok && ids[id.Name] {
						found = true
					}
				}
			case *ast.IncDecStmt:
				if id, ok := s.X.(*ast.Ident); ok && ids[id.Name] {
					found = true
				}
			}
			return true
		})
	}
	return found
}
