package goil

import (
	"testing"
)

type Test struct {
	I *int     `path:"i"`
	F *float32 `path:"f"`
	S string   `path:"s"`
}

func TestBinding(t *testing.T) {
	c := Context{
		params: Params{
			Param{
				Key:   "i",
				Value: "11",
			},
			Param{
				Key:   "f",
				Value: "13.14",
			},
			Param{
				Key:   "s",
				Value: "hello",
			},
		},
	}
	test := &Test{}
	bindPathParams(c.params, test)
	t.Errorf("%d %f %s", *test.I, *test.F, test.S)

}

func TestParseTag(t *testing.T) {
	tag1 := `min(1 2)max(1 2)`
	name, params, error := parseTag(tag1)
	t.Errorf("%s %v %v", name, params, error)
	tag1 = `min(1 2)`
	name, params, error = parseTag(tag1)
	t.Errorf("%s %v %v", name, params, error)
	tag1 = ` min ( 1 2 )   `
	name, params, error = parseTag(tag1)
	t.Errorf("%s %v %v", name, params, error)
	tag1 = ` min ( 1 2 ) max ( 2 3 4 ) `
	name, params, error = parseTag(tag1)
	t.Errorf("%s %v %v", name, params, error)
	tag1 = `min`
	name, params, error = parseTag(tag1)
	t.Errorf("%s %v %v", name, params, error)
	tag1 = `min max`
	name, params, error = parseTag(tag1)
	t.Errorf("%s %v %v", name, params, error)
	tag1 = ` min `
	name, params, error = parseTag(tag1)
	t.Errorf("%s %v %v", name, params, error)
	tag1 = ` min max `
	name, params, error = parseTag(tag1)
	t.Errorf("%s %v %v", name, params, error)
	tag1 = `min min(1 2)max(2 3 4)max`
	name, params, error = parseTag(tag1)
	t.Errorf("%s %v %v", name, params, error)
	tag1 = ` min min ( 1  2 ) max ( 2 3 4 ) max `
	name, params, error = parseTag(tag1)
	t.Errorf("%s %v %v", name, params, error)
}
