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

	"github.com/devnev/testeach/internal"
)

// Case registers a child case within Start or a parent Case.
// See testeach docs on how cases are run.
func Case(tp **testing.T, name string, impl func()) {
	initialT := *tp
	loaded, _ := internal.ActiveTests.Load(initialT)
	registerCb, _ := loaded.(func(string, func()))
	if registerCb == nil {
		internal.RunTargetAndRecurse(initialT, []*internal.StackFrame{}, func(t *testing.T) {
			*tp = t
			defer func() {
				*tp = initialT
			}()
			impl()
		})
		return
	}
	registerCb(name, impl)
}
