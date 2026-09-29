package rules

import "testing"

func TestCollectLongMethodFindings_Flags(t *testing.T) {
	pass := loadPass(t, `package pkg

func BigOne(a, b, c, d, e bool) int {
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

	findings := CollectLongMethodFindings(pass, DefaultConfig().LongMethod)

	requireSingleFinding(t, findings, "CS101")
	if findings[0].Confidence < 60 || findings[0].Confidence > 97 {
		t.Errorf("Confidence = %d, want in [60, 97]", findings[0].Confidence)
	}
}

func TestCollectLongMethodFindings_IgnoresSmallFunctions(t *testing.T) {
	pass := loadPass(t, `package pkg

func Add(a, b int) int {
	return a + b
}
`)

	findings := CollectLongMethodFindings(pass, DefaultConfig().LongMethod)

	requireFindingCount(t, findings, 0)
}

func TestCollectLongMethodFindings_RespectsConfigOverride(t *testing.T) {
	pass := loadPass(t, `package pkg

func Add(a, b int) int {
	x := a + b
	y := x * 2
	z := y - 1
	return z
}
`)

	cfg := DefaultConfig().LongMethod
	cfg.MaxStatements = 2

	findings := CollectLongMethodFindings(pass, cfg)

	requireFindingCount(t, findings, 1)
}
