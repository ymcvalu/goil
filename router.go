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
	GET(path string, handlers ...HandlerFunc) IRouter
	POST(path string, handlers ...HandlerFunc) IRouter
	PUT(path string, handlers ...HandlerFunc) IRouter
	DELETE(path string, handlers ...HandlerFunc) IRouter
	OPTIONS(path string, handlers ...HandlerFunc) IRouter
	PATCH(path string, handlers ...HandlerFunc) IRouter
	CONNECT(path string, handlers ...HandlerFunc) IRouter
	TRACE(path string, handlers ...HandlerFunc) IRouter
}

type methodTree struct {
	*node
}

func (t *methodTree) isNil() bool {
	return t.node == nil
}

type router struct {
	group
	trees []methodTree
}

type group struct {
	middlewares *Middleware
	router      *router
	base        string
}

//assert *router and *group implements IRouter interface
var _ IRouter = &router{}
var _ IRouter = &group{}

func NewRouter() (r *router) {
	r = &router{
		trees: make([]methodTree, len(methods)),
	}
	r.group.router = r
	return
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

func (r *router) add(method string, path string, chain *Middleware) {

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
}

func (g *group) ADD(method string, path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.router.add(method, joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	return g
}

func (g *group) GET(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.router.add("GET", joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	return g
}

func (g *group) POST(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.router.add("POST", joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	return g
}

func (g *group) PUT(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.router.add("PUT", joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	return g
}

func (g *group) DELETE(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.router.add("DELETE", joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	return g
}

func (g *group) HEAD(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.router.add("HEAD", joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	return g
}

func (g *group) OPTIONS(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.router.add("OPTIONS", joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	return g
}

func (g *group) PATCH(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.router.add("PATCH", joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	return g
}

func (g *group) CONNECT(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.router.add("CONNECT", joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	return g
}

func (g *group) TRACE(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.router.add("TRACE", joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	return g
}
