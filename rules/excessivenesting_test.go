package rules

import "testing"

func TestCollectExcessiveNestingFindings_Flags(t *testing.T) {
	pass := loadPass(t, `package pkg

func DeepIf(a, b, c, d, e bool) int {
	if a {
		if b {
			if c {
				if d {
					if e {
						return 1
					}
				}
			}
		}
	}
	return 0
}
`)

	findings := CollectExcessiveNestingFindings(pass, DefaultConfig().ExcessiveNesting)

	requireSingleFinding(t, findings, "CS103")
}

func TestCollectExcessiveNestingFindings_ElseIfChainDoesNotStack(t *testing.T) {
	pass := loadPass(t, `package pkg

func Chain(x int) int {
	if x == 1 {
		return 1
	} else if x == 2 {
		return 2
	} else if x == 3 {
		return 3
	}
	return 0
}
`)

	findings := CollectExcessiveNestingFindings(pass, DefaultConfig().ExcessiveNesting)

	requireFindingCount(t, findings, 0)
}
