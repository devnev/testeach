package testeach_test

import "testing"

func TestSuite(t *testing.T) {
	testNoCases(t)
	testTwoCases(t)
	testNestedCases(t)
	testOsStat(t)
}
