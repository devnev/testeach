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
	"sync"
	"testing"
)

// Start allows nested registering of test cases with Case.
// See testeach docs on how cases are run.
func Start(t *testing.T, suite func(t *testing.T)) {
	runTargetAndRecurse(t, []*stackFrame{}, suite)
}

// Case registers a child case within Start or a parent Case.
// See testeach docs on how cases are run.
func Case(t *testing.T, name string, impl func()) {
	loaded, _ := activeTests.Load(t)
	registerCb, _ := loaded.(func(string, func()))
	if registerCb == nil {
		panic(fmt.Sprintf("attempted to register case %q for terminated test %q", name, t.Name()))
	}
	registerCb(name, impl)
}

type suite func(t *testing.T)

var activeTests sync.Map

type stackFrame struct {
	names  []string
	target int
}

func runTargetAndRecurse(t *testing.T, stack []*stackFrame, suite suite) {
	newNames := runStackTarget(t, stack, suite)
	if len(newNames) > 0 {
		runLastFrame(t, append(stack, &stackFrame{names: newNames}), suite)
	}
}

func runStackTarget(t *testing.T, stack []*stackFrame, suite suite) (subNames []string) {
	seenNewNames := map[string]struct{}{}

	currentCase := make([]int, 0, len(stack)+1)
	currentCase = append(currentCase, 0)

	registerCb := func(caseName string, caseImpl func()) {
		currentDepth := len(currentCase)

		// If we have a longer index than we have stack, this callback is being executed from
		// within the target test case. Record the name of sub-tests without executing them.
		if currentDepth > len(stack) {
			if _, ok := seenNewNames[caseName]; ok {
				t.Fatalf("duplicate test case %q", caseName)
			}
			seenNewNames[caseName] = struct{}{}
			subNames = append(subNames, caseName)
			return
		}

		// Find the frame for the current case and check that the case is valid.
		currIdx := currentCase[currentDepth-1]
		currFrame := stack[currentDepth-1]
		if currIdx >= len(currFrame.names) {
			t.Fatalf("unexpected extra case %q", caseName)
		}
		if recordedName := currFrame.names[currIdx]; caseName != recordedName {
			// Although not necessary, we're strict about the test case names staying the same to help
			// debug test code.
			t.Fatalf("case name at index %d changed; first %q then %q", currIdx, recordedName, caseName)
		}

		// Determine if we need to do anything, then record that we've seen the current case by
		// incrementing the case index.
		targetIdx := currFrame.target
		runCase := currIdx == targetIdx
		currentCase[currentDepth-1]++
		if !runCase {
			return
		}

		// We know that the current case is in the path to the target. Add a new frame of indexes
		// for the sub-cases of the current case.
		currentCase = append(currentCase, 0)
		defer func() {
			currentCase = currentCase[:currentDepth]
		}()

		// Execute test callback
		caseImpl()

		// The test callback should have called back to us the same number of times as previously
		// recorded, unless we were recording a new frame. We verify that these callbacks actually
		// happened as the strict enforcement should help debug test code and also ensures that the
		// target case was actually executed.
		if len(stack) >= len(currentCase) {
			called := currentCase[currentDepth]
			expectedCalls := len(stack[currentDepth].names)
			if called < expectedCalls {
				t.Fatalf("missing test case callbacks; expected %d but got %d", expectedCalls, called)
			}
		}
	}

	{
		activeTests.Store(t, registerCb)
		defer activeTests.Delete(t)
		suite(t)
	}

	return subNames
}

func runLastFrame(t *testing.T, stack []*stackFrame, suite suite) {
	newFrame := stack[len(stack)-1]
	for newFrame.target = 0; newFrame.target < len(newFrame.names); newFrame.target++ {
		caseName := newFrame.names[newFrame.target]
		t.Run(caseName, func(t *testing.T) {
			runTargetAndRecurse(t, stack, suite)
		})
	}
}
