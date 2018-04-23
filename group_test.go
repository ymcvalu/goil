package goil

import (
	"testing"
)

func TestADD(t *testing.T) {
	route := newRouter()
	route.Use(func(c *Context) {
		t.Errorf("router")
		c.Next()
	})

	route.ADD("GET", "/test", func(c *Context) {
		t.Errorf("node")
		c.Next()
	}, func(c *Context) {
		t.Errorf("test")
	})

	route.ADD("GET", "/test1", func(c *Context) {
		t.Errorf("node1")
		c.Next()
	}, func(c *Context) {
		t.Errorf("test1")
	})

	chain, _, _ := route.route("GET", "/test")
	c := &Context{
		chain: chain,
	}
	c.Next()
}
