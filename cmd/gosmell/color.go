package main

import (
	"os"

	"golang.org/x/term"

	"github.com/bartollo/gosmell/report"
)

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiRed    = "\x1b[31m"
	ansiYellow = "\x1b[33m"
	ansiGreen  = "\x1b[32m"
	ansiCyan   = "\x1b[36m"
	ansiGray   = "\x1b[90m"
	ansiWhite  = "\x1b[37m"
)

var colorEnabled = shouldUseColor()

func shouldUseColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func colorize(code, text string) string {
	if !colorEnabled || text == "" {
		return text
	}
	return code + text + ansiReset
}

func bold(text string) string {
	return colorize(ansiBold, text)
}

func severityColor(s report.Severity) string {
	switch s {
	case report.High:
		return ansiRed
	case report.Medium:
		return ansiYellow
	case report.Low:
		return ansiCyan
	default:
		return ""
	}
}

func scoreTierColor(score int) string {
	switch {
	case score >= 75:
		return ansiGreen
	case score >= 50:
		return ansiYellow
	default:
		return ansiRed
	}
}
