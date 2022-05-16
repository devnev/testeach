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

	"github.com/devnev/testeach/internal"
)

// Start allows nested registering of test cases with Case.
// See testeach docs on how cases are run.
func Start(t *testing.T, suite func(t *testing.T)) {
	internal.RunTargetAndRecurse(t, []*internal.StackFrame{}, suite)
}

// Case registers a child case within Start or a parent Case.
// See testeach docs on how cases are run.
func Case(t *testing.T, name string, impl func()) {
	loaded, _ := internal.ActiveTests.Load(t)
	registerCb, _ := loaded.(func(string, func()))
	if registerCb == nil {
		panic(fmt.Sprintf("attempted to register case %q for terminated test %q", name, t.Name()))
	}
	registerCb(name, impl)
}
