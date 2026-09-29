package rules

import "testing"

func TestCollectPlaceholderImplementationFindings_FlagsNotImplemented(t *testing.T) {
	pass := loadPass(t, `package pkg

import "errors"

func NotDone() error {
	return errors.New("not implemented")
}
`)

	findings := CollectPlaceholderImplementationFindings(pass)

	requireSingleFinding(t, findings, "CS112")
}

func TestCollectPlaceholderImplementationFindings_FlagsTodoOverEmptyReturn(t *testing.T) {
	pass := loadPass(t, `package pkg

func TodoStub() []int {
	// TODO: implement real logic
	return []int{}
}
`)

	findings := CollectPlaceholderImplementationFindings(pass)

	requireFindingCount(t, findings, 1)
}

func TestCollectPlaceholderImplementationFindings_IgnoresBareEmptyReturn(t *testing.T) {
	pass := loadPass(t, `package pkg

// DefaultItems returns the built-in item set.
func DefaultItems() []int {
	return []int{}
}
`)

	findings := CollectPlaceholderImplementationFindings(pass)

	requireFindingCount(t, findings, 0)
}
