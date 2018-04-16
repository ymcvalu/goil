/**
 * 路由存储结构
 * router的add方法实现向methodTree注册路由
 * IRouter接口定义了路由注册接口
 * *group实现IRouter接口
 * router嵌套了group，自动生成*router的桥接方法，因此也实现了IRouter接口
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
	Params map[string]string
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
	ANY(path string, handlers ...HandlerFunc) IRouter
}

type methodTree struct {
	*node
	method string
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

func (r *router) findTree(method string) (*methodTree, bool) {
	for i := range r.trees {
		if r.trees[i].method == method {
			return &r.trees[i], true
		}
	}
	return nil, false
}

func (r *router) route(method, path string) (chain *Middleware, params Params, tsr bool) {
	tree, exist := r.findTree(method)
	if !exist {
		chain = nil
		return
	}
	return tree.routerMapping(path)
}

//assert *router and *group implements IRouter interface
var _ IRouter = &router{}
var _ IRouter = &group{}

func newRouter() (r *router) {
	r = &router{
		trees: make([]methodTree, len(methods)),
	}

	for k, v := range methods {
		r.trees[v-1].method = k
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
	if len(path) == 0 || path[0] != '/' {
		panic(fmt.Sprintf("path must start with '/'"))
	}

	if chain == nil {
		panic(fmt.Sprintf("handler nil:%s", path))
	}

	tree, exists := r.findTree(method)

	if !exists {
		panic(fmt.Sprintf("unsupported method:%s", method))
	}

	if tree.isNil() {
		tree.node = &node{
			pattern: "/",
			typ:     static,
		}
	}

	tree.addNode(path, chain)
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

func (g *group) ANY(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}

	for k := range methods {
		g.router.add(k, joinPath(g.base, path), combineChain(g.middlewares, handlers...))
	}

	return g
}
