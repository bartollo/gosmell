package gosmell

import (
	"golang.org/x/tools/go/analysis"

	"github.com/golangci/plugin-module-register/register"

	"github.com/bartollo/gosmell/rules"
)

func init() {
	register.Plugin("gosmell", New)
}

func New(settings any) (register.LinterPlugin, error) {
	overrides, err := register.DecodeSettings[rules.Overrides](settings)
	if err != nil {
		return nil, err
	}

	return &gosmellPlugin{cfg: overrides.Apply(rules.DefaultConfig())}, nil
}

type gosmellPlugin struct {
	cfg rules.Config
}

func (p *gosmellPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		rules.NewLongMethodAnalyzer(p.cfg),
		rules.NewLargeClassAnalyzer(p.cfg),
		rules.NewExcessiveNestingAnalyzer(p.cfg),
		rules.NewDuplicateCodeAnalyzer(p.cfg),
		rules.DeadCodeAnalyzer,
		rules.UnusedConstructorDependencyAnalyzer,
		rules.SwallowedErrorAnalyzer,
		rules.RedundantConditionAnalyzer,
		rules.CommentsAnalyzer,
		rules.DefensiveNoiseAnalyzer,
		rules.NewCopyPasteDriftAnalyzer(p.cfg),
		rules.PlaceholderImplementationAnalyzer,
	}, nil
}

func (p *gosmellPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
