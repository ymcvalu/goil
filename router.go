/**
 *路由操作接口
 */
package goil

import (
	"fmt"
)

//HTTP METHOD TYPE
const (
	//the DEFAULT method when not other method handler define
	_ = iota
	GET
	POST
	PUT
	DELETE
	HEAD
	OPTIONS
	PATCH
	CONNECT
	TRACE
	ANY
)

type (
	Param struct {
		Key   string
		Value string
	}

	Params []Param
)

//map method to index
var methods = map[string]int{
	"GET":     1,
	"POST":    2,
	"PUT":     3,
	"DELETE":  4,
	"HEAD":    5,
	"OPTIONS": 6,
	"PATCH":   7,
	"CONNECT": 8,
	"TRACE":   9,
}

type IRouter interface {
	Group(path string, handlers ...HandlerFunc) IRouter
	Use(handlers ...HandlerFunc) IRouter
	ADD(method, path string, handler ...HandlerFunc) IRouter
}

type methodTree struct {
	*node
}

func (t *methodTree) isNil() bool {
	return t.node == nil
}

type router struct {
	middlewares *Middleware
	trees       []methodTree
}

func NewRouter() *router {
	return &router{
		trees: make([]methodTree, len(methods)),
	}
}

func (r *router) Group(path string, handlers ...HandlerFunc) IRouter {
	return &group{
		base:        path,
		middlewares: combineChain(r.middlewares, handlers...),
		router:      r,
	}
}

func (r *router) Use(handlers ...HandlerFunc) IRouter {
	r.middlewares = combineChain(r.middlewares, handlers...)
	return r
}

func (r *router) ADD(method string, path string, handlers ...HandlerFunc) IRouter {

	return r.add(method, path, combineChain(r.middlewares, handlers...))
}

func (r *router) add(method string, path string, chain *Middleware) IRouter {
	if path[0] != '/' {
		panic(fmt.Sprintf("path must start with '/'"))
	}

	if chain == nil {
		panic(fmt.Sprintf("handler nil:%s", path))
	}

	idx, exists := methods[method]
	if !exists {
		panic(fmt.Sprintf("unsupported method:%s", method))
	}
	if r.trees[idx].isNil() {
		r.trees[idx].node = &node{
			pattern: "/",
			typ:     static,
		}
	}

	r.trees[idx].addNode(path, chain)
	return r
}

type group struct {
	middlewares *Middleware
	router      *router
	base        string
}

func (g *group) Group(path string, handlers ...HandlerFunc) IRouter {
	return &group{
		middlewares: combineChain(g.middlewares, handlers...),
		router:      g.router,
		base:        joinPath(g.base, path),
	}
}

func (g *group) Use(handlers ...HandlerFunc) IRouter {
	g.middlewares = combineChain(g.middlewares, handlers...)
	return g
}

func (g *group) ADD(method, path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}

	g.router.add(method, joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	return g
}
