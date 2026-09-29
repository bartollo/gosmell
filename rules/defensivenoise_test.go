package rules

import "testing"

func TestCollectDefensiveNoiseFindings_Flags(t *testing.T) {
	pass := loadPass(t, `package pkg

import "fmt"

func F(x int) int {
	if x > 0 {
		fmt.Println("positive")
	}
	fmt.Println("middle")
	if x > 0 {
		return x
	}
	return 0
}
`)

	findings := CollectDefensiveNoiseFindings(pass)

	requireSingleFinding(t, findings, "CS110")
}

func TestCollectDefensiveNoiseFindings_IgnoresReassignment(t *testing.T) {
	pass := loadPass(t, `package pkg

func F(x int) int {
	if x > 0 {
		x = -x
	}
	if x > 0 {
		return x
	}
	return 0
}
`)

	findings := CollectDefensiveNoiseFindings(pass)

	requireFindingCount(t, findings, 0)
}

func TestCollectDefensiveNoiseFindings_OneFindingPerRepeatedGuard(t *testing.T) {
	pass := loadPass(t, `package pkg

import "fmt"

func F(x int) int {
	if x > 0 {
		fmt.Println("one")
	}
	if x > 0 {
		fmt.Println("two")
	}
	if x > 0 {
		return x
	}
	return 0
}
`)

	findings := CollectDefensiveNoiseFindings(pass)

	requireFindingCount(t, findings, 2)
}

func TestCollectDefensiveNoiseFindings_IgnoresFreshErrDeclarations(t *testing.T) {
	pass := loadPass(t, `package pkg

import "os"

func F() error {
	if _, err := os.Open("a"); err != nil {
		return err
	}
	if _, err := os.Open("b"); err != nil {
		return err
	}
	return nil
}
`)

	findings := CollectDefensiveNoiseFindings(pass)

	requireFindingCount(t, findings, 0)
}
