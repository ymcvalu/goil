package goil

import "testing"

type P struct {
	Username string
	UID      int64
}
type R struct {
	Name string
	Data interface{}
}

func TestWrapper(t *testing.T) {
	f1 := func(p P) {}
	f2 := func(c *Context) {}
	f3 := func(c *Context, p P) {}
	f4 := func() {}
	//f5 := func(p P, c *Context) {}
	//f5 := func(p P, p1 P) {}
	//f5 := func(c1 *Context, c2 *Context) {}
	g := new(GroupX)
	g.ErrorHandler = func(c *Context, err error) {}
	g.RenderHandler = func(c *Context, data interface{}) {}

	g.Wrapper(f1)
	g.Wrapper(f2)
	g.Wrapper(f3)
	g.Wrapper(f4)
	//g.Wrapper(f5)

	f6 := func() (r R) { return }
	f7 := func() (e error) { return }
	f8 := func() (r R, e error) { return }
	//f9 := func() (e error, r R) { return }
	//f9 := func() (e error, e1 error) { return }
	//f9 := func() (r1 R, r2 R) { return }
	g.Wrapper(f6)
	g.Wrapper(f7)
	g.Wrapper(f8)
	//g.Wrapper(f9)
}
