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

	route.ADD("GET", "/test", func(c *Context) {
		t.Errorf("node")
		c.Next()
	}, func(c *Context) {
		t.Errorf("test")
	})
	chain, _, _ := route.trees[GET].routerMapping("/test")
	c := &Context{
		chain: chain,
	}
	c.Next()
}
