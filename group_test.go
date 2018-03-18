package goil

import (
	"testing"
)

func TestADD(t *testing.T) {
	route := NewRouter()
	route.Use(func(c *Context) {
		t.Errorf("router")
		c.Next()
	})
	g := route.Group("/v1", func(c *Context) {
		t.Errorf("group 1")
		c.Next()
	})
	g.Use(func(c *Context) {
		t.Errorf("group 2")
		c.Next()
	})
	g.ADD("GET", "test", func(c *Context) {
		t.Errorf("node")
		c.Next()
	}, func(c *Context) {
		t.Errorf("test")
	})
	chain, _, _ := route.trees[GET].routerMapping("/v1/test")
	c := &Context{
		chain: chain,
	}
	c.Next()
}
