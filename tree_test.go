package goil

import (
	"testing"
)

func TestGetPrefix(t *testing.T) {
	t.Log(getPrefix("/xx/x/sd", "/xx/xxx"))
	t.Log(getPrefix("/x/x:", "/x/xxx"))
	t.Log(getPrefix("/xx/x", "/xx/x:x"))
	t.Log(getPrefix("/xxx/x*", "/xxx/xxx"))
	t.Log(getPrefix("/xxx/xxx", "/xxx/x*"))
	t.Log(getPrefix("/xxx/xxx", "/xxx/xxxx"))

}

func TestGetParamNum(t *testing.T) {
	t.Log(getParamNum("/dd/:::/*sd/:sd"))
	t.Log(getParamNum("/dd/s/*sd/:"))
	t.Log(getParamNum("/d/s/c"))
}

func TestStrictConflictChecked(t *testing.T) {
	t.Log(isStrictConflictChecked())
	setStrictConflictChecked(true)
	t.Log(isStrictConflictChecked())
	setStrictConflictChecked(false)
	t.Log(isStrictConflictChecked())

}

func TestAddNode(t *testing.T) {

}
