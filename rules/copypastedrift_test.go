package rules

import "testing"

func TestCollectCopyPasteDriftFindings_Flags(t *testing.T) {
	pass := loadPass(t, `package pkg

import "errors"

type Validator struct{}

func (v *Validator) ValidateA(x int) error {
	if x < 0 {
		return errors.New("negative")
	}
	return nil
}

func (v *Validator) ValidateB(x int) error {
	if x < 0 {
		return errors.New("negative")
	}
	return nil
}

func (v *Validator) ValidateC(x int) error {
	return nil
}
`)

	findings := CollectCopyPasteDriftFindings(pass, DefaultConfig().CopyPasteDrift)

	requireSingleFinding(t, findings, "CS111")
}

func TestCollectCopyPasteDriftFindings_IgnoresSmallGroups(t *testing.T) {
	pass := loadPass(t, `package pkg

import "errors"

type Validator struct{}

func (v *Validator) ValidateA(x int) error {
	if x < 0 {
		return errors.New("negative")
	}
	return nil
}

func (v *Validator) ValidateC(x int) error {
	return nil
}
`)

	findings := CollectCopyPasteDriftFindings(pass, DefaultConfig().CopyPasteDrift)

	requireFindingCount(t, findings, 0)
}
