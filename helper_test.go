package goil

import "testing"

func TestAssert(t *testing.T) {
	assert(false)
	//assert(true)
}

func TestAssert1(t *testing.T) {
	assert1(false, "assert failed")
}

func TestFuncName(t *testing.T) {
	t.Error(funcName(TestAssert))
}
