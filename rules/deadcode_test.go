package rules

import "testing"

func TestCollectDeadCodeFindings_FlagsUnreferenced(t *testing.T) {
	pass := loadPass(t, `package pkg

func deadHelper(x int) int {
	return x * 2
}

func Exported() int {
	return 1
}
`)

	findings := CollectDeadCodeFindings(pass)

	requireSingleFinding(t, findings, "CS105")
}

func TestCollectDeadCodeFindings_IgnoresCalledHelper(t *testing.T) {
	pass := loadPass(t, `package pkg

func helper(x int) int {
	return x * 2
}

func Exported() int {
	return helper(1)
}
`)

	findings := CollectDeadCodeFindings(pass)

	requireFindingCount(t, findings, 0)
}

func TestCollectDeadCodeFindings_DistinguishesSameNamedMethodsOnDifferentTypes(t *testing.T) {
	pass := loadPass(t, `package pkg

type A struct{}

func (a A) helper() int {
	return 1
}

type B struct{}

func (b B) helper() int {
	return 2
}

func Exported() int {
	var b B
	return b.helper()
}
`)

	findings := CollectDeadCodeFindings(pass)

	requireFindingCount(t, findings, 1)
}

func TestCollectDeadCodeFindings_IgnoresMain(t *testing.T) {
	pass := loadPass(t, `package pkg

func main() {}
`)

	findings := CollectDeadCodeFindings(pass)

	requireFindingCount(t, findings, 0)
}

func TestCollectDeadCodeFindings_ValueReferenceCountsAsUse(t *testing.T) {
	pass := loadPass(t, `package pkg

type Handler struct {
	Run func()
}

func handleIt() {}

var H = Handler{Run: handleIt}
`)

	findings := CollectDeadCodeFindings(pass)

	requireFindingCount(t, findings, 0)
}
