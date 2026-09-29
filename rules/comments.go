package rules

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bartollo/gosmell/report"
)

var CommentsAnalyzer = &analysis.Analyzer{
	Name: "comments",
	Doc:  "CS109: detects short comments whose every meaningful word already appears in the statement directly below them",
	Run:  runComments,
}

const commentsConfidence = 70

var wordPattern = regexp.MustCompile(`[A-Za-z]+`)

var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "to": true, "of": true, "and": true,
	"is": true, "in": true, "on": true, "for": true, "if": true, "it": true,
}

func runComments(pass *analysis.Pass) (any, error) {
	reportFindings(pass, CollectCommentsFindings(pass))
	return nil, nil
}

func CollectCommentsFindings(pass *analysis.Pass) []report.Finding {
	var findings []report.Finding

	for _, file := range pass.Files {
		cmap := ast.NewCommentMap(pass.Fset, file, file.Comments)

		ast.Inspect(file, func(n ast.Node) bool {
			block, ok := n.(*ast.BlockStmt)
			if !ok {
				return true
			}
			for _, stmt := range block.List {
				groups := cmap[stmt]
				if len(groups) == 0 {
					continue
				}
				comment := groups[len(groups)-1]
				if comment.End() >= stmt.Pos() {
					continue
				}

				commentWords := meaningfulWords(comment.Text())
				if len(commentWords) == 0 {
					continue
				}

				statementWords := meaningfulWords(renderNode(pass.Fset, stmt))

				if isSubset(commentWords, statementWords) {
					findings = append(findings, report.Finding{
						Pos:        comment.Pos(),
						Code:       "CS109",
						Confidence: commentsConfidence,
						Message:    "comment only narrates the statement below it; explain the why instead.",
					})
				}
			}
			return true
		})
	}

	return findings
}

var camelCaseBoundary = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func meaningfulWords(text string) map[string]bool {
	words := map[string]bool{}
	for _, raw := range wordPattern.FindAllString(text, -1) {
		candidates := append([]string{raw}, strings.Fields(camelCaseBoundary.ReplaceAllString(raw, "$1 $2"))...)
		for _, w := range candidates {
			w = strings.ToLower(w)
			if len(w) <= 2 || stopWords[w] {
				continue
			}
			words[w] = true
		}
	}
	return words
}

func isSubset(subset, superset map[string]bool) bool {
	if len(subset) == 0 {
		return false
	}
	for w := range subset {
		if !superset[w] {
			return false
		}
	}
	return true
}

func renderNode(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}
