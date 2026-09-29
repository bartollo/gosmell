package rules

import "testing"

func TestCollectRedundantConditionFindings_SelfNested(t *testing.T) {
	pass := loadPass(t, `package pkg

func F(x int) int {
	if x > 0 {
		if x > 0 {
			return 1
		}
	}
	return 0
}
`)

	findings := CollectRedundantConditionFindings(pass)

	requireSingleFinding(t, findings, "CS108")
}

func TestCollectRedundantConditionFindings_LiteralBool(t *testing.T) {
	pass := loadPass(t, `package pkg

func F(flag bool) int {
	if flag == true {
		return 1
	}
	return 0
}
`)

	findings := CollectRedundantConditionFindings(pass)

	requireFindingCount(t, findings, 1)
}

func TestCollectRedundantConditionFindings_DuplicateOperand(t *testing.T) {
	pass := loadPass(t, `package pkg

func F(x int) int {
	if x > 0 && x > 0 {
		return 1
	}
	return 0
}
`)

	findings := CollectRedundantConditionFindings(pass)

	requireFindingCount(t, findings, 1)
}

func TestCollectRedundantConditionFindings_ChainDuplicate(t *testing.T) {
	pass := loadPass(t, `package pkg

func F(x int) int {
	if x > 5 {
		return 1
	} else if x > 5 {
		return 2
	}
	return 0
}
`)

	findings := CollectRedundantConditionFindings(pass)

	requireFindingCount(t, findings, 1)
}

func TestCollectRedundantConditionFindings_IgnoresFreshDeclarations(t *testing.T) {
	pass := loadPass(t, `package pkg

func F(elts []interface{}) int {
	count := 0
	for _, elt := range elts {
		if kv, ok := elt.(int); ok {
			if key, ok := interface{}(kv).(int); ok {
				count += key
			}
		}
	}
	return count
}
`)

	findings := CollectRedundantConditionFindings(pass)

	requireFindingCount(t, findings, 0)
}
