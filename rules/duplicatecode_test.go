package rules

import "testing"

func TestCollectDuplicateCodeFindings_Flags(t *testing.T) {
	pass := loadPass(t, `package pkg

func Dup1(x int) int {
	y := x + 1
	z := y * 2
	if z > 10 {
		return z
	}
	return y
}

func Dup2(x int) int {
	y := x + 5
	z := y * 9
	if z > 90 {
		return z
	}
	return y
}
`)

	findings := CollectDuplicateCodeFindings(pass, DefaultConfig().DuplicateCode)

	requireSingleFinding(t, findings, "CS104")
	if findings[0].Message == "" {
		t.Error("expected a non-empty message naming the original function")
	}
}

func TestCollectDuplicateCodeFindings_IgnoresDifferentShapes(t *testing.T) {
	pass := loadPass(t, `package pkg

func A(x int) int {
	y := x + 1
	z := y * 2
	if z > 10 {
		return z
	}
	return y
}

func B(x int) int {
	for i := 0; i < x; i++ {
		x += i
	}
	return x
}
`)

	findings := CollectDuplicateCodeFindings(pass, DefaultConfig().DuplicateCode)

	requireFindingCount(t, findings, 0)
}
