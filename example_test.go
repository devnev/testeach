package testeach_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/devnev/testeach"
)

func testNoCases(t *testing.T) {
	Start(t, func(t *testing.T) {
		// No cases are registered, so this is run just once. Equivalent to not
		// using Start at all.
	})
}

func testTwoCases(t *testing.T) {
	Start(t, func(t *testing.T) {
		// Code here is run once on its own, and once for each Case below
		t.Log("I am run three times")
		defer t.Log("I am run three times, after any cases")

		Case(t, "first case", func() {
			t.Log("I am run once, after setup and before teardown")
		})

		Case(t, "second case", func() {
			t.Log("I am also run once")
		})
	})
}

func testNestedCases(t *testing.T) {
	Start(t, func(t *testing.T) {
		// Code here is run once on its own, once for the outer case, and once
		// with both the outer and inner cases.
		t.Log("I am run three times")
		defer t.Log("I am run three times, after any cases")

		Case(t, "outer case", func() {
			// Code here is run once without the inner case, and once with the
			// inner case.
			t.Log("I am run twice")

			Case(t, "inner case", func() {
				t.Log("I am run once")
			})
		})
	})
}

// testOsStat emulates what a Test function for os.Stat might look like when using testeach.Suite.
func testOsStat(t *testing.T) {
	// The entire callback to testeach.Suite is run three times: once as a prelude, and once for every
	// callback to test. Any setup and teardown that should be executed once for the entire suite
	// should be placed before the call to testeach.Suite.
	Start(t, func(t *testing.T) {
		Case(t, "of a directory", func() {
			// This setup and teardown code is re-executed in each of the three runs.
			dir, err := ioutil.TempDir(".", "test")
			if err != nil {
				t.Fatalf("unable to create test directory: %s", err)
			}
			// Teardown can be done using regular defer
			defer func() {
				err = os.RemoveAll(dir)
				if err != nil {
					t.Fatalf("unable to remove test directory: %s", err)
				}
			}()

			// In the first "prelude" run, none of the test callbacks below are executed.
			// Along with registering all the test cases, this allows testing the setup and teardown
			// in isolation.

			// In the second run, only this callback is executed
			Case(t, "dir is a directory", func() {
				fi, err := os.Stat(dir)
				if err != nil {
					t.Fatalf("unexpected error from stat: %s", err)
				}
				if !fi.IsDir() {
					t.Fatal("expected FileInfo.IsDir() to return true")
				}
			})
			// In the third run, only this callback is executed
			Case(t, "dir contains created file", func() {
				// We can safely mutate resource used in other test cases as each case is started
				// with a fresh setup
				err := ioutil.WriteFile(filepath.Join(dir, "testfile"), []byte("data"), 0644)
				if err != nil {
					t.Fatalf("unexpected error from writefile: %s", err)
				}
				items, err := ioutil.ReadDir(dir)
				if err != nil {
					t.Fatalf("unexpected error from readdir: %s", err)
				}
				if len(items) != 1 || items[0].Name() != "testfile" {
					t.Fatalf("expected results to contain only testfile, got %+v", items)
				}
			})
		})

		Case(t, "of a file", func() {
			// This setup and teardown code is executed twice, once without the sub-test and once
			// with the sub-test.
			file, err := ioutil.TempFile(".", "test")
			if err != nil {
				t.Fatalf("unable to create test directory: %s", err)
			}
			defer func() {
				err = os.RemoveAll(file.Name())
				if err != nil {
					t.Fatalf("unable to remove test directory: %s", err)
				}
			}()

			Case(t, "reports not a directory", func() {
				fi, err := os.Stat(file.Name())
				if err != nil {
					t.Fatalf("unexpected error from stat: %v", err)
				}
				if fi.IsDir() {
					t.Fatalf("expected FileInfo.IsDir() to return false")
				}
			})
		})
	})
}

func Example() {
	// See above
}
