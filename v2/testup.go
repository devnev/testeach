// testeach provides a simple mechanism for shared test setup/teardown for Go tests.
//
// For each case, all that case's parents are re-run, such that their setup, teardown and
// assertions automatically apply to the case.
//
// Variable scoping follows natural language rules, avoiding issues common in
// BDD frameworks with Before() functions.
//
// Cases are registered using callbacks rather than reflection, avoiding the
// possibility of tests mistakenly being missed due to typos.
package testeach

import (
	"testing"

	internal "github.com/devnev/testeach/v3/x"
)

// Case registers a child case within Start or a parent Case.
// See testeach docs on how cases are run.
func Case(tp **testing.T, name string, impl func()) {
	initialT := *tp
	defer func() {
		*tp = initialT
	}()

	caseCallback := internal.LoadCaseCallback(initialT)
	if caseCallback != nil {
		caseCallback(name, impl)
		return
	}

	runSuite := func(newT *testing.T) {
		*tp = newT
		impl()
	}
	initialT.Run(name, func(t *testing.T) {
		internal.NewTarget(t, runSuite).Run()
	})
}
