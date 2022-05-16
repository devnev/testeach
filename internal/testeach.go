package internal

import (
	"sync"
	"testing"
)

type Suite func(t *testing.T)

var ActiveTests sync.Map

type StackFrame struct {
	Names  []string
	Target int
}

func RunTargetAndRecurse(t *testing.T, stack []*StackFrame, suite Suite) {
	newNames := RunStackTarget(t, stack, suite)
	if len(newNames) > 0 {
		RunLastFrame(t, append(stack, &StackFrame{Names: newNames}), suite)
	}
}

func RunStackTarget(t *testing.T, stack []*StackFrame, suite Suite) (subNames []string) {
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
		if currIdx >= len(currFrame.Names) {
			t.Fatalf("unexpected extra case %q", caseName)
		}
		if recordedName := currFrame.Names[currIdx]; caseName != recordedName {
			// Although not necessary, we're strict about the test case names staying the same to help
			// debug test code.
			t.Fatalf("case name at index %d changed; first %q then %q", currIdx, recordedName, caseName)
		}

		// Determine if we need to do anything, then record that we've seen the current case by
		// incrementing the case index.
		targetIdx := currFrame.Target
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
			expectedCalls := len(stack[currentDepth].Names)
			if called < expectedCalls {
				t.Fatalf("missing test case callbacks; expected %d but got %d", expectedCalls, called)
			}
		}
	}

	{
		ActiveTests.Store(t, registerCb)
		defer ActiveTests.Delete(t)
		suite(t)
	}

	return subNames
}

func RunLastFrame(t *testing.T, stack []*StackFrame, suite Suite) {
	newFrame := stack[len(stack)-1]
	for newFrame.Target = 0; newFrame.Target < len(newFrame.Names); newFrame.Target++ {
		caseName := newFrame.Names[newFrame.Target]
		t.Run(caseName, func(t *testing.T) {
			RunTargetAndRecurse(t, stack, suite)
		})
	}
}
