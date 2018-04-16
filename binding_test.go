package goil

import (
	"testing"
)

type Test struct {
	I int     `path:"i"`
	F float32 `path:"f"`
}

func TestBinding(t *testing.T) {
	c := Context{
		params: map[string]string{
			"i": "11",
			"f": "13.14",
		},
	}
	test := &Test{}
	pathParamsBinding(c, test)
	t.Errorf("%d %f", test.I, test.F)

}
