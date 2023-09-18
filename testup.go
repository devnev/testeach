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
	"fmt"
	"testing"

	internal "github.com/devnev/testeach/v3/x"
)

// Start allows nested registering of test cases with Case.
// See testeach docs on how cases are run.
func Start(t *testing.T, suite func(t *testing.T)) {
	t.Helper()

	internal.NewTarget(t, suite).Run()
}

// Case registers a child case within Start or a parent Case.
// See testeach docs on how cases are run.
func Case(t *testing.T, name string, impl func()) {
	t.Helper()

	caseCallback := internal.LoadCaseCallback(t)
	if caseCallback == nil {
		panic(fmt.Sprintf("attempted to register case %q for terminated test %q", name, t.Name()))
	}
	caseCallback(name, impl)
}
