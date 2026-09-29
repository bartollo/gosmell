package rules

import "testing"

func TestCollectSwallowedErrorFindings_FlagsEmptyBlock(t *testing.T) {
	pass := loadPass(t, `package pkg

import "os"

func Do() {
	_, err := os.Open("f")
	if err != nil {
	}
}
`)

	findings := CollectSwallowedErrorFindings(pass)

	requireSingleFinding(t, findings, "CS107")
}

func TestCollectSwallowedErrorFindings_FlagsDiscardedError(t *testing.T) {
	pass := loadPass(t, `package pkg

import "os"

func Do() {
	_, err := os.Open("f")
	_ = err
}
`)

	findings := CollectSwallowedErrorFindings(pass)

	requireFindingCount(t, findings, 1)
}

func TestCollectSwallowedErrorFindings_FlagsDiscardedCallResult(t *testing.T) {
	pass := loadPass(t, `package pkg

import "os"

func Do() {
	_ = os.Remove("f")
}
`)

	findings := CollectSwallowedErrorFindings(pass)

	requireFindingCount(t, findings, 1)
}

func TestCollectSwallowedErrorFindings_FlagsDiscardedMultiValueError(t *testing.T) {
	pass := loadPass(t, `package pkg

import "os"

func Do() {
	f, _ := os.Open("f")
	_ = f
}
`)

	findings := CollectSwallowedErrorFindings(pass)

	requireFindingCount(t, findings, 1)
}

func TestCollectSwallowedErrorFindings_IgnoresHandledError(t *testing.T) {
	pass := loadPass(t, `package pkg

import "os"

func Do() error {
	_, err := os.Open("f")
	if err != nil {
		return err
	}
	return nil
}
`)

	findings := CollectSwallowedErrorFindings(pass)

	requireFindingCount(t, findings, 0)
}
