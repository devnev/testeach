package internal

import (
	"sync"
	"testing"
)

type Suite func(t *testing.T)

var activeTests sync.Map

func NewTarget(t *testing.T, root Suite) *Target {
	return &Target{
		T:    t,
		Root: root,
	}
}

func LoadCaseCallback(t *testing.T) func(string, func()) {
	loaded, _ := activeTests.Load(t)
	registerCb, _ := loaded.(func(string, func()))
	return registerCb
}

type Target struct {
	T         *testing.T
	CaseNames [][]string
	Path      []int
	Root      Suite
}

func (t *Target) Run() {
	t.T.Helper()

	names := t.RunSelf()
	t.RunChildren(names)
}

func (t *Target) RunSelf() []string {
	t.T.Helper()

	state := NewRun(t)

	activeTests.Store(t.T, state.Case)
	defer activeTests.Delete(t.T)

	t.Root(t.T)

	return state.NewNamesOrdered
}

func (t *Target) RunChildren(names []string) {
	t.T.Helper()

	if len(names) == 0 {
		return
	}
	for idx, name := range names {
		t.T.Run(name, func(testT *testing.T) {
			child := Target{
				T:         testT,
				CaseNames: append(t.CaseNames, names),
				Path:      append(t.Path, idx),
				Root:      t.Root,
			}
			child.Run()
		})
	}
}

func (t *Target) CheckSeenCase(depth, idx int, name string) {
	names := t.CaseNames[depth]
	if idx > len(names) {
		t.T.Fatalf("unexpected extra case %q", name)
	}
	if recordedName := names[idx]; name != recordedName {
		// Although not necessary, we're strict about the test case names staying the same to help
		// debug test code.
		t.T.Fatalf("case name at index %d changed; first %q then %q", idx, recordedName, name)
	}
}

func (t *Target) CheckSeenChildren(depth, seen int) {
	if len(t.CaseNames) > depth {
		expected := len(t.CaseNames[depth])
		if seen < expected {
			t.T.Fatalf("missing test case callbacks; expected %d but got %d", expected, seen)
		}
	}
}

func NewRun(t *Target) *RunState {
	return &RunState{
		Target:       t,
		CurrentPath:  append(make([]int, 0, len(t.Path)+1), 0),
		NewNamesSeen: map[string]struct{}{},
	}
}

type RunState struct {
	Target          *Target
	CurrentPath     []int
	NewNamesOrdered []string
	NewNamesSeen    map[string]struct{}
}

func (s *RunState) Case(name string, impl func()) {
	// If we have a longer index than we have stack, this callback is being executed from
	// within the target test case. Record the name of sub-tests without executing them.
	if s.Depth() >= len(s.Target.Path) {
		s.HandleNewCase(name)
		return
	}

	// Find the frame for the current case and check that the case is valid.
	s.Target.CheckSeenCase(s.Depth(), s.CurrentPath[s.Depth()], name)

	// Determine if we need to do anything, then record that we've seen the current case by
	// incrementing the case index.
	shouldRun := s.IsTargetCase()
	s.IncrementCurrent()
	if !shouldRun {
		return
	}

	// We know that the current case is in the path to the target. Add a new frame of indexes
	// for the sub-cases of the current case.
	popLevel := s.PushLevel()
	defer popLevel()

	// Execute test callback
	impl()

	// The test callback should have called back to us the same number of times as previously
	// recorded, unless we were recording a new frame. We verify that these callbacks actually
	// happened as the strict enforcement should help debug test code and also ensures that the
	// target case was actually executed.
	s.Target.CheckSeenChildren(s.Depth(), s.CurrentPath[s.Depth()])
}

func (s *RunState) HandleNewCase(name string) {
	if _, ok := s.NewNamesSeen[name]; ok {
		s.Target.T.Fatalf("duplicate test case %q", name)
	}
	s.NewNamesSeen[name] = struct{}{}
	s.NewNamesOrdered = append(s.NewNamesOrdered, name)
}

func (s *RunState) IsTargetCase() bool {
	return s.CurrentPath[s.Depth()] == s.Target.Path[s.Depth()]
}

func (s *RunState) Depth() int {
	return len(s.CurrentPath) - 1
}

func (s *RunState) IncrementCurrent() {
	s.CurrentPath[s.Depth()]++
}

func (s *RunState) PushLevel() func() {
	depth := s.Depth()
	s.CurrentPath = append(s.CurrentPath, 0)
	return func() {
		s.CurrentPath = s.CurrentPath[:depth+1]
	}
}
