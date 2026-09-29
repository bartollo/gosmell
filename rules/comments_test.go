package rules

import "testing"

func TestCollectCommentsFindings_FlagsNarratingComment(t *testing.T) {
	pass := loadPass(t, `package pkg

func Example() {
	// save user
	saveUser()
}

func saveUser() {}
`)

	findings := CollectCommentsFindings(pass)

	requireSingleFinding(t, findings, "CS109")
}

func TestCollectCommentsFindings_IgnoresExplanatoryComment(t *testing.T) {
	pass := loadPass(t, `package pkg

func Example() {
	// Retry once: the upstream API is flaky under load.
	saveUser()
}

func saveUser() {}
`)

	findings := CollectCommentsFindings(pass)

	requireFindingCount(t, findings, 0)
}
