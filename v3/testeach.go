package testeach

import (
	"testing"

	internal "github.com/devnev/testeach/v3/x"
)

func NewSuite(tp **testing.T) *Suite {
	return &Suite{
		Tp:       tp,
		initialT: *tp,
	}
}

type Suite struct {
	Tp       **testing.T
	initialT *testing.T
}

func (s *Suite) Case(name string, caseImpl func()) {
	if caseCallback := internal.LoadCaseCallback(*s.Tp); caseCallback != nil {
		caseCallback(name, caseImpl)
		return
	}

	if s.initialT == nil {
		s.initialT = *s.Tp
	}

	runSuite := func(newT *testing.T) {
		*s.Tp = newT
		caseImpl()
	}
	s.initialT.Run(name, func(t *testing.T) {
		tgt := internal.NewTarget(t, runSuite)
		tgt.Run()
	})
}
