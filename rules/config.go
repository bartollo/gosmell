package rules

type Config struct {
	LongMethod       LongMethodConfig
	LargeClass       LargeClassConfig
	ExcessiveNesting ExcessiveNestingConfig
	DuplicateCode    DuplicateCodeConfig
	CopyPasteDrift   CopyPasteDriftConfig
}

type LongMethodConfig struct {
	MaxStatements    int
	MaxComplexity    int
	MaxNesting       int
	MaxCalls         int
	MaxCollaborators int
}

type LargeClassConfig struct {
	MaxFields        int
	MaxMethods       int
	MaxCollaborators int
	MaxStatements    int
	MinExceeded      int
}

type ExcessiveNestingConfig struct {
	MaxDepth int
}

type DuplicateCodeConfig struct {
	MinStatements int
}

type CopyPasteDriftConfig struct {
	MinSiblingGroup int
	MaxEditDistance int
}

func DefaultConfig() Config {
	return Config{
		LongMethod: LongMethodConfig{
			MaxStatements:    40,
			MaxComplexity:    10,
			MaxNesting:       3,
			MaxCalls:         25,
			MaxCollaborators: 15,
		},
		LargeClass: LargeClassConfig{
			MaxFields:        10,
			MaxMethods:       15,
			MaxCollaborators: 8,
			MaxStatements:    300,
			MinExceeded:      3,
		},
		ExcessiveNesting: ExcessiveNestingConfig{
			MaxDepth: 4,
		},
		DuplicateCode: DuplicateCodeConfig{
			MinStatements: 5,
		},
		CopyPasteDrift: CopyPasteDriftConfig{
			MinSiblingGroup: 3,
			MaxEditDistance: 2,
		},
	}
}

type Overrides struct {
	LongMethod       LongMethodOverrides       `json:"longMethod"`
	LargeClass       LargeClassOverrides       `json:"largeClass"`
	ExcessiveNesting ExcessiveNestingOverrides `json:"excessiveNesting"`
	DuplicateCode    DuplicateCodeOverrides    `json:"duplicateCode"`
	CopyPasteDrift   CopyPasteDriftOverrides   `json:"copyPasteDrift"`
}

type LongMethodOverrides struct {
	MaxStatements    *int `json:"maxStatements,omitempty"`
	MaxComplexity    *int `json:"maxComplexity,omitempty"`
	MaxNesting       *int `json:"maxNesting,omitempty"`
	MaxCalls         *int `json:"maxCalls,omitempty"`
	MaxCollaborators *int `json:"maxCollaborators,omitempty"`
}

type LargeClassOverrides struct {
	MaxFields        *int `json:"maxFields,omitempty"`
	MaxMethods       *int `json:"maxMethods,omitempty"`
	MaxCollaborators *int `json:"maxCollaborators,omitempty"`
	MaxStatements    *int `json:"maxStatements,omitempty"`
	MinExceeded      *int `json:"minExceeded,omitempty"`
}

type ExcessiveNestingOverrides struct {
	MaxDepth *int `json:"maxDepth,omitempty"`
}

type DuplicateCodeOverrides struct {
	MinStatements *int `json:"minStatements,omitempty"`
}

type CopyPasteDriftOverrides struct {
	MinSiblingGroup *int `json:"minSiblingGroup,omitempty"`
	MaxEditDistance *int `json:"maxEditDistance,omitempty"`
}

func (o Overrides) Apply(cfg Config) Config {
	applyInt(&cfg.LongMethod.MaxStatements, o.LongMethod.MaxStatements)
	applyInt(&cfg.LongMethod.MaxComplexity, o.LongMethod.MaxComplexity)
	applyInt(&cfg.LongMethod.MaxNesting, o.LongMethod.MaxNesting)
	applyInt(&cfg.LongMethod.MaxCalls, o.LongMethod.MaxCalls)
	applyInt(&cfg.LongMethod.MaxCollaborators, o.LongMethod.MaxCollaborators)

	applyInt(&cfg.LargeClass.MaxFields, o.LargeClass.MaxFields)
	applyInt(&cfg.LargeClass.MaxMethods, o.LargeClass.MaxMethods)
	applyInt(&cfg.LargeClass.MaxCollaborators, o.LargeClass.MaxCollaborators)
	applyInt(&cfg.LargeClass.MaxStatements, o.LargeClass.MaxStatements)
	applyInt(&cfg.LargeClass.MinExceeded, o.LargeClass.MinExceeded)

	applyInt(&cfg.ExcessiveNesting.MaxDepth, o.ExcessiveNesting.MaxDepth)
	applyInt(&cfg.DuplicateCode.MinStatements, o.DuplicateCode.MinStatements)

	applyInt(&cfg.CopyPasteDrift.MinSiblingGroup, o.CopyPasteDrift.MinSiblingGroup)
	applyInt(&cfg.CopyPasteDrift.MaxEditDistance, o.CopyPasteDrift.MaxEditDistance)

	return cfg
}

func applyInt(dst *int, override *int) {
	if override != nil {
		*dst = *override
	}
}
