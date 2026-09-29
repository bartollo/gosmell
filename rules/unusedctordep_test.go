package rules

import "testing"

func TestCollectUnusedConstructorDependencyFindings_Flags(t *testing.T) {
	pass := loadPass(t, `package pkg

type Dep struct{}

type Foo struct {
	dep *Dep
}

func NewFoo(dep *Dep) *Foo {
	return &Foo{dep: dep}
}
`)

	findings := CollectUnusedConstructorDependencyFindings(pass)

	requireSingleFinding(t, findings, "CS106")
}

func TestCollectUnusedConstructorDependencyFindings_IgnoresReadDependency(t *testing.T) {
	pass := loadPass(t, `package pkg

type Dep struct{}

func (d *Dep) Ping() {}

type Foo struct {
	dep *Dep
}

func NewFoo(dep *Dep) *Foo {
	return &Foo{dep: dep}
}

func (f *Foo) Use() {
	f.dep.Ping()
}
`)

	findings := CollectUnusedConstructorDependencyFindings(pass)

	requireFindingCount(t, findings, 0)
}

func TestCollectUnusedConstructorDependencyFindings_IgnoresCompoundAssignmentRead(t *testing.T) {
	pass := loadPass(t, `package pkg

type Foo struct {
	dep int
}

func NewFoo(dep int) *Foo {
	return &Foo{dep: dep}
}

func (f *Foo) Bump() {
	f.dep += 1
}
`)

	findings := CollectUnusedConstructorDependencyFindings(pass)

	requireFindingCount(t, findings, 0)
}

func TestCollectUnusedConstructorDependencyFindings_IgnoresNonParamValues(t *testing.T) {
	pass := loadPass(t, `package pkg

type Config struct {
	Name string
	Run  func()
}

func NewConfig() *Config {
	return &Config{Name: "static", Run: func() {}}
}
`)

	findings := CollectUnusedConstructorDependencyFindings(pass)

	requireFindingCount(t, findings, 0)
}
