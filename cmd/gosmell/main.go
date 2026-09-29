package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/bartollo/gosmell/report"
)

const (
	barWidth  = 28
	wrapWidth = 78
)

type displayFinding struct {
	line       int
	meta       report.RuleMeta
	confidence int
	message    string
}

type golangciJSON struct {
	Issues []struct {
		FromLinter string `json:"FromLinter"`
		Text       string `json:"Text"`
		Pos        struct {
			Filename string `json:"Filename"`
			Line     int    `json:"Line"`
		} `json:"Pos"`
	} `json:"Issues"`
}

var textPattern = regexp.MustCompile(`^(?:[A-Za-z0-9_]+: )?(CS\d{3}) \((\d+)% confidence\): (.*)$`)

func main() {
	pattern := "./..."
	if len(os.Args) > 1 {
		pattern = os.Args[1]
	}

	dir, err := os.Getwd()
	if err != nil {
		fatal(err)
	}

	lintBin, err := resolveLintBinary()
	if err != nil {
		fatal(err)
	}

	files, err := listFiles(dir, pattern)
	if err != nil {
		fatal(err)
	}

	fmt.Println()
	totalLines := countLinesWithProgress(files)
	fmt.Println()
	fmt.Println()

	issues, err := runGolangciLint(lintBin, dir, pattern)
	if err != nil {
		fatal(err)
	}

	printReport(len(files), totalLines, buildFileFindings(issues, dir))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "gosmell:", err)
	os.Exit(2)
}

func countLinesWithProgress(files []string) int {
	renderProgress(0, len(files))
	totalLines := 0
	for i, f := range files {
		totalLines += countLines(f)
		renderProgress(i+1, len(files))
	}
	return totalLines
}

func buildFileFindings(issues *golangciJSON, dir string) map[string][]displayFinding {
	fileFindings := map[string][]displayFinding{}

	for _, issue := range issues.Issues {
		if issue.FromLinter != "gosmell" {
			continue
		}
		m := textPattern.FindStringSubmatch(issue.Text)
		if m == nil {
			continue
		}
		code := m[1]
		confidence, _ := strconv.Atoi(m[2])
		message := m[3]

		rel := issue.Pos.Filename
		if filepath.IsAbs(rel) {
			if r, err := filepath.Rel(dir, rel); err == nil {
				rel = r
			}
		}

		fileFindings[rel] = append(fileFindings[rel], displayFinding{
			line:       issue.Pos.Line,
			meta:       report.Rules[code],
			confidence: confidence,
			message:    message,
		})
	}

	return fileFindings
}

func resolveLintBinary() (string, error) {
	if bin := os.Getenv("GOSMELL_LINT_BIN"); bin != "" {
		return bin, nil
	}

	if exe, err := os.Executable(); err == nil {
		sibling := filepath.Join(filepath.Dir(exe), "custom-gcl")
		if info, err := os.Stat(sibling); err == nil && !info.IsDir() {
			return sibling, nil
		}
	}

	if bin, err := exec.LookPath("golangci-lint"); err == nil {
		return bin, nil
	}

	return "", fmt.Errorf("no custom-gcl or golangci-lint binary found; set GOSMELL_LINT_BIN")
}

func listFiles(dir, pattern string) ([]string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var files []string
	for _, pkg := range pkgs {
		for _, f := range pkg.CompiledGoFiles {
			if seen[f] {
				continue
			}
			seen[f] = true
			files = append(files, f)
		}
	}
	sort.Strings(files)
	return files, nil
}

func countLines(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte("\n"))
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

const gosmellConfigTemplate = `version: "2"
linters:
  settings:
    custom:
      gosmell:
        type: module
        description: "detects Go code smells (CS101-CS112)"
        original-url: github.com/bartollo/gosmell
%s`

const userSettingsFile = ".gosmell.yml"

func runGolangciLint(bin, dir, pattern string) (*golangciJSON, error) {
	cfgPath, err := writeTempConfig(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(cfgPath) }()

	args := []string{
		"run",
		"--config=" + cfgPath,
		"--enable-only=gosmell",
		"--output.json.path=stdout",
		"--uniq-by-line=false",
		"--max-issues-per-linter=0",
		"--max-same-issues=0",
		pattern,
	}

	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("running %s: %w", bin, err)
		}

	}

	var result golangciJSON
	if err := json.NewDecoder(&stdout).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing golangci-lint output: %w", err)
	}
	return &result, nil
}

func writeTempConfig(dir string) (string, error) {
	settingsBlock, err := userSettingsBlock(dir)
	if err != nil {
		return "", err
	}

	cfgFile, err := os.CreateTemp(dir, ".gosmell-*.yml")
	if err != nil {
		return "", fmt.Errorf("creating temp config: %w", err)
	}
	defer func() { _ = cfgFile.Close() }()

	if _, err := fmt.Fprintf(cfgFile, gosmellConfigTemplate, settingsBlock); err != nil {
		_ = os.Remove(cfgFile.Name())
		return "", fmt.Errorf("writing temp config: %w", err)
	}

	return cfgFile.Name(), nil
}

func userSettingsBlock(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, userSettingsFile))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", userSettingsFile, err)
	}

	var block strings.Builder
	block.WriteString("        settings:\n")
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if line == "" {
			block.WriteByte('\n')
			continue
		}
		block.WriteString("          ")
		block.WriteString(line)
		block.WriteByte('\n')
	}
	return block.String(), nil
}

func renderProgress(done, total int) {
	pct := 100
	filled := barWidth
	if total > 0 {
		pct = done * 100 / total
		filled = done * barWidth / total
	}
	bar := strings.Repeat("▓", filled) + strings.Repeat("░", barWidth-filled)
	fmt.Printf("\r%d/%d [%s] %d%%", done, total, bar, pct)
}

func printReport(totalFiles, totalLines int, fileFindings map[string][]displayFinding) {
	total, highCount, mediumCount, lowCount := countSeverities(fileFindings)
	score := computeScore(highCount, mediumCount, lowCount)

	printSummaryHeader(totalFiles, totalLines, total, score)
	printFileBlocks(fileFindings)
	printVerdict(total, highCount, mediumCount, lowCount)
}

func printSummaryHeader(totalFiles, totalLines, total, score int) {
	tierColor := scoreTierColor(score)

	fmt.Println("  " + bold("Gosmell"))
	fmt.Println()
	fmt.Printf("  Score  %s  %s\n",
		colorize(tierColor+ansiBold, fmt.Sprintf("%d/100", score)),
		colorize(tierColor, scoreTier(score)))
	fmt.Printf("  Files  %s analysed  ·  %s lines  ·  %s findings\n",
		formatThousands(totalFiles), formatThousands(totalLines), formatThousands(total))
}

func printVerdict(total, highCount, mediumCount, lowCount int) {
	fmt.Println()
	fmt.Println(colorize(ansiGray, "  "+strings.Repeat("─", 60)))
	fmt.Println()

	if total > 0 {
		fmt.Printf("  %d findings: %s\n", total, summarizeSeverities(highCount, mediumCount, lowCount))
	} else {
		fmt.Println(colorize(ansiGreen, "  0 findings"))
	}
	fmt.Println()

	if highCount > 0 {
		fmt.Printf("  %s — %d finding(s) at high or above (fail_on: high)\n",
			colorize(ansiRed+ansiBold, "✗ Failed"), highCount)
		os.Exit(1)
	}

	fmt.Println("  " + colorize(ansiGreen+ansiBold, "✓ Passed"))
}

func countSeverities(fileFindings map[string][]displayFinding) (total, high, medium, low int) {
	for _, findings := range fileFindings {
		for _, f := range findings {
			total++
			switch f.meta.Severity {
			case report.High:
				high++
			case report.Medium:
				medium++
			case report.Low:
				low++
			}
		}
	}
	return total, high, medium, low
}

func computeScore(high, medium, low int) int {
	penalty := high*5 + medium*3 + low
	if penalty > 100 {
		penalty = 100
	}
	return 100 - penalty
}

func printFileBlocks(fileFindings map[string][]displayFinding) {
	files := make([]string, 0, len(fileFindings))
	for name := range fileFindings {
		files = append(files, name)
	}
	sort.Strings(files)

	for _, name := range files {
		findings := fileFindings[name]
		sort.Slice(findings, func(i, j int) bool { return findings[i].line < findings[j].line })

		fmt.Println()
		fmt.Printf("  %s\n", colorize(ansiWhite+ansiBold, name))

		for _, f := range findings {
			fmt.Println()
			severity := colorize(severityColor(f.meta.Severity)+ansiBold, fmt.Sprintf("%-9s", f.meta.Severity))
			code := bold(fmt.Sprintf("%-5s", f.meta.Code))
			confidence := colorize(ansiGray, fmt.Sprintf("(%d%% confidence)", f.confidence))
			fmt.Printf("  %4d  %s%s  %s  %s\n", f.line, severity, code, f.meta.Name, confidence)
			printWrapped(f.message, "        ", "        ")
			printWrapped(colorize(ansiCyan, "→")+" "+f.meta.Suggestion, "        ", "          ")
		}
	}
}

func summarizeSeverities(high, medium, low int) string {
	var parts []string
	if high > 0 {
		parts = append(parts, colorize(ansiRed, fmt.Sprintf("%d high", high)))
	}
	if medium > 0 {
		parts = append(parts, colorize(ansiYellow, fmt.Sprintf("%d medium", medium)))
	}
	if low > 0 {
		parts = append(parts, colorize(ansiCyan, fmt.Sprintf("%d low", low)))
	}
	return strings.Join(parts, ", ")
}

func scoreTier(score int) string {
	switch {
	case score >= 90:
		return "Pristine"
	case score >= 75:
		return "Clean"
	case score >= 50:
		return "Sloppy"
	case score >= 25:
		return "Messy"
	default:
		return "Disaster"
	}
}

func printWrapped(text, firstIndent, restIndent string) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}

	line := firstIndent
	lineLen := 0

	for _, w := range words {
		if lineLen > 0 && lineLen+1+len(w) > wrapWidth {
			fmt.Println(line)
			line = restIndent
			lineLen = 0
		}
		if lineLen > 0 {
			line += " "
			lineLen++
		}
		line += w
		lineLen += len(w)
	}
	fmt.Println(line)
}

func formatThousands(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out)
}
