package report

import (
	"go/token"
	"math"
)

type Severity string

const (
	High   Severity = "HIGH"
	Medium Severity = "MEDIUM"
	Low    Severity = "LOW"
)

type Finding struct {
	Pos        token.Pos
	Code       string
	Confidence int
	Message    string
}

type RuleMeta struct {
	Code       string
	Name       string
	Severity   Severity
	Suggestion string
}

var Rules = map[string]RuleMeta{
	"CS101": {
		"CS101", "Long Method", High,
		"Identify the distinct jobs this method performs and extract each into its own well-named method or class. Extracting the outermost branches first usually makes the remaining shape obvious.",
	},
	"CS102": {
		"CS102", "Large Class", High,
		"Group the members that change together and move each group into its own class. Clusters of methods that share the same subset of properties are usually a class trying to get out.",
	},
	"CS103": {
		"CS103", "Excessive Nesting", Medium,
		"Use early returns, guard clauses or extracted helper functions to flatten the branching.",
	},
	"CS104": {
		"CS104", "Duplicate Code", Medium,
		"Extract the shared logic into a single function, method or package that both call sites can reuse.",
	},
	"CS105": {
		"CS105", "Dead Code", Medium,
		"Confirm nothing else in the package depends on it, then remove it.",
	},
	"CS106": {
		"CS106", "Unused Constructor Dependency", Medium,
		"Remove the dependency from the constructor and its signature, or start using it if it was meant to be read.",
	},
	"CS107": {
		"CS107", "Swallowed Exception", High,
		"Log, wrap or return the error instead of letting it disappear silently.",
	},
	"CS108": {
		"CS108", "Redundant Condition", Low,
		"Simplify the condition by removing the redundant check.",
	},
	"CS109": {
		"CS109", "Comments", Low,
		"Replace the comment with one that explains why, or remove it.",
	},
	"CS110": {
		"CS110", "Defensive Programming Noise", Low,
		"Remove the repeated guard; the earlier check already covers this path.",
	},
	"CS111": {
		"CS111", "Copy-Paste Drift", High,
		"Compare this method against its siblings and confirm the difference is intentional, not an unfinished copy.",
	},
	"CS112": {
		"CS112", "Placeholder Implementation", High,
		"Finish the implementation or remove the placeholder before merging.",
	},
}

func ConfidenceFromExcess(ratios ...float64) int {
	exceeded := 0
	total := 0.0
	for _, r := range ratios {
		if math.IsInf(r, 1) || math.IsNaN(r) {
			return 97
		}
		if r > 1 {
			exceeded++
			total += r
		}
	}
	if exceeded == 0 {
		return 60
	}
	avg := total / float64(exceeded)
	score := min(60+exceeded*8+int(avg*6), 97)
	return score
}
