package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

func NewCopyPasteDriftAnalyzer(cfg Config) *analysis.Analyzer {
	return newAnalyzer("copypastedrift",
		"CS111: heuristically flags a method body nearly identical to its siblings, where the one difference looks like an unfinished copy rather than a deliberate variation",
		func(pass *analysis.Pass) []report.Finding {
			return CollectCopyPasteDriftFindings(pass, cfg.CopyPasteDrift)
		})
}

type siblingMethod struct {
	name  string
	pos   token.Pos
	shape []string
}

func CollectCopyPasteDriftFindings(pass *analysis.Pass, cfg CopyPasteDriftConfig) []report.Finding {
	var findings []report.Finding
	for _, methods := range groupSiblingMethods(pass) {
		findings = append(findings, findDivergentSiblings(methods, cfg)...)
	}
	return findings
}

func groupSiblingMethods(pass *analysis.Pass) map[string][]siblingMethod {
	groups := map[string][]siblingMethod{}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 || fn.Body == nil {
				continue
			}
			recv := receiverTypeName(fn.Recv.List[0].Type)
			if recv == "" {
				continue
			}
			groups[recv] = append(groups[recv], siblingMethod{
				name:  fn.Name.Name,
				pos:   fn.Pos(),
				shape: topLevelShape(fn.Body),
			})
		}
	}

	return groups
}

func findDivergentSiblings(methods []siblingMethod, cfg CopyPasteDriftConfig) []report.Finding {
	if len(methods) < cfg.MinSiblingGroup {
		return nil
	}

	majorityShape, majorityCount := majorityShapeOf(methods)
	if majorityCount*2 <= len(methods) {
		return nil
	}

	var majorityTokens []string
	if majorityShape != "" {
		majorityTokens = strings.Split(majorityShape, ",")
	}

	var findings []report.Finding
	for _, m := range methods {
		if strings.Join(m.shape, ",") == majorityShape {
			continue
		}
		d := shapeEditDistance(majorityTokens, m.shape)
		if d <= 0 || d > cfg.MaxEditDistance {
			continue
		}
		findings = append(findings, report.Finding{
			Pos:        m.pos,
			Code:       "CS111",
			Confidence: 90 - d*15,
			Message: fmt.Sprintf(
				"%s diverges from the pattern shared by %d sibling methods on this type; verify it isn't an unfinished copy-paste.",
				m.name, majorityCount),
		})
	}
	return findings
}

func majorityShapeOf(methods []siblingMethod) (string, int) {
	counts := map[string]int{}
	for _, m := range methods {
		counts[strings.Join(m.shape, ",")]++
	}

	shape, count := "", 0
	for s, c := range counts {
		if c > count {
			shape, count = s, c
		}
	}
	return shape, count
}

func topLevelShape(body *ast.BlockStmt) []string {
	shape := make([]string, 0, len(body.List))
	for _, stmt := range body.List {
		shape = append(shape, statementKind(stmt))
	}
	return shape
}

func statementKind(stmt ast.Stmt) string {
	return fmt.Sprintf("%T", stmt)
}

func shapeEditDistance(a, b []string) int {
	rows, cols := len(a)+1, len(b)+1
	dp := make([][]int, rows)
	for i := range dp {
		dp[i] = make([]int, cols)
		dp[i][0] = i
	}
	for j := 0; j < cols; j++ {
		dp[0][j] = j
	}
	for i := 1; i < rows; i++ {
		for j := 1; j < cols; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1]
				continue
			}
			min := dp[i-1][j]
			if dp[i][j-1] < min {
				min = dp[i][j-1]
			}
			if dp[i-1][j-1] < min {
				min = dp[i-1][j-1]
			}
			dp[i][j] = min + 1
		}
	}
	return dp[rows-1][cols-1]
}
