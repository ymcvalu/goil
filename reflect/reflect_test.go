package reflect

import "testing"

type Bean struct {
	Name string

	Age int
}

func TestFields(t *testing.T) {
	age := 16
	b := Bean{
		Name: "Jim",
		Age:  age,
	}
	fds1, err := Fields(b)
	t.Error(fds1, err)
	fds2, err := Fields(&b)
	t.Error(fds2, err)
	t.Error(IsStruct(b))
	b1 := Bean{}
	fds3, err := Fields(b1)
	t.Error(fds3, err)
	var a *int

	t.Error(IsZero(a))
}

type Any = interface{}

func TestIsNil(t *testing.T) {
	var i1 Any = ([]int)(nil)
	if !IsNil(i1) {
		t.Error("failed")
	}
	var i2 Any = (*string)(nil)
	if !IsNil(i2) {
		t.Error("failed")
	}

	if IsNil(3) {
		t.Error("failed")
	}
}
