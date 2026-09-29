package rules

import (
	"fmt"
	"strings"
	"testing"
)

func bigServiceSource() string {
	var methods strings.Builder
	for i := 1; i <= 16; i++ {
		fmt.Fprintf(&methods, "func (b *BigService) M%d() {\n", i)
		fmt.Fprintf(&methods, "\tb.logger.Log(\"%d\")\n", i)
		methods.WriteString("\tfmt.Println(b.repo.Find(1))\n")
		methods.WriteString("\tb.svcC.Log(\"c\")\n\tb.svcD.Log(\"d\")\n\tb.svcE.Log(\"e\")\n")
		methods.WriteString("\tb.svcF.Log(\"f\")\n\tb.svcG.Log(\"g\")\n\tb.svcH.Log(\"h\")\n}\n\n")
	}

	return fmt.Sprintf(`package pkg

import "fmt"

type Logger struct{}
type Repo struct{}
type Cache struct{}

func (l *Logger) Log(msg string) {}
func (r *Repo) Find(id int) int  { return id }

type BigService struct {
	A, B, C, D, E, F, G, H, I, J, K       int
	logger                                *Logger
	repo                                  *Repo
	unusedDep                             *Cache
	svcC, svcD, svcE, svcF, svcG, svcH    *Logger
}

func NewBigService(logger, svcC, svcD, svcE, svcF, svcG, svcH *Logger, repo *Repo, unusedDep *Cache) *BigService {
	return &BigService{logger: logger, repo: repo, unusedDep: unusedDep, svcC: svcC, svcD: svcD, svcE: svcE, svcF: svcF, svcG: svcG, svcH: svcH}
}

%s`, methods.String())
}

func TestCollectLargeClassFindings_Flags(t *testing.T) {
	pass := loadPass(t, bigServiceSource())

	findings := CollectLargeClassFindings(pass, DefaultConfig().LargeClass)

	requireSingleFinding(t, findings, "CS102")
}

func TestCollectLargeClassFindings_IgnoresSmallStructs(t *testing.T) {
	pass := loadPass(t, `package pkg

type Small struct {
	Name string
}

func (s *Small) Greet() string {
	return "hi " + s.Name
}
`)

	findings := CollectLargeClassFindings(pass, DefaultConfig().LargeClass)

	requireFindingCount(t, findings, 0)
}
