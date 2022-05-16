package testeach_test

import "testing"

func TestSuite_noCases(t *testing.T) {
	testNoCases(t)
}

func TestSuite_twoCases(t *testing.T) {
	testTwoCases(t)
}

func TestSuite_nestedCases(t *testing.T) {
	testNestedCases(t)
}

func TestSuite_osStat(t *testing.T) {
	testOsStat(t)
}
