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
	"net/http"
	"strings"
)

//HTTP METHOD TYPE
const (
	//the DEFAULT method when not other method handler define

	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
	CONNECT = "CONNECT"
	TRACE   = "TRACE"
	ANY     = "ANY"
)

type (
	Params map[string]string
)

//map method to index
var methods = map[string]struct{}{
	GET:     struct{}{},
	POST:    struct{}{},
	PUT:     struct{}{},
	DELETE:  struct{}{},
	HEAD:    struct{}{},
	OPTIONS: struct{}{},
	PATCH:   struct{}{},
	CONNECT: struct{}{},
	TRACE:   struct{}{},
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
	trees map[string]*methodTree
}

type group struct {
	middlewares HandlerChain
	router      *router
	base        string
}

func (r *router) findTree(method string) (*methodTree, bool) {
	if tree, ok := r.trees[method]; ok {
		return tree, true
	}
	return nil, false
}

func (r *router) route(method, path string) (chain HandlerChain, params Params, tsr bool) {
	tree, exist := r.findTree(method)
	if !exist {
		//return the not method found handler
		chain = append(r.middlewares, NotMethodHandler)
		return
	}
	chain, params, tsr = tree.routerMapping(path)
	//404 not found
	if len(chain) == 0 {
		chain = append(r.middlewares, NotFoundHandler)
		return
	}
	return
}

//assert *router and *group implements IRouter interface
var _ IRouter = &router{}
var _ IRouter = &group{}

func newRouter() (r *router) {
	r = &router{
		trees: make(map[string]*methodTree, len(methods)),
	}

	for k, _ := range methods {
		tree := methodTree{
			method: k,
		}
		r.trees[k] = &tree
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

func (r *router) add(method string, path string, chain HandlerChain) {

	assert1(len(path) > 0 && path[0] == '/', fmt.Sprintf("path must start with '/'"))
	assert1(chain != nil, fmt.Sprintf("the handler of %s is nil", path))
	tree, exists := r.findTree(method)
	assert1(exists, fmt.Sprintf("unsupported method:%s", method))
	guard.execSafely(func() {
		if tree.isNil() {
			tree.node = &node{
				pattern: "/",
				typ:     static,
			}
		}
		tree.addNode(path, chain)
	})
}

func (g *group) ADD(method string, path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	absolutePath := joinPath(g.base, path)
	chain := combineChain(g.middlewares, handlers...)
	g.router.add(method, absolutePath, chain)
	if RunMode() == DBG {
		handlerNum := len(chain)
		handlerName := funcName(chain[handlerNum-1])
		printRouteInfo(method, absolutePath, handlerName, handlerNum)
	}
	return g
}

func (g *group) GET(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.ADD(GET, path, handlers...)
	return g
}

func (g *group) POST(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.ADD(POST, path, handlers...)
	return g
}

func (g *group) PUT(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.ADD(PUT, path, handlers...)
	return g
}

func (g *group) DELETE(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.ADD(DELETE, path, handlers...)
	return g
}

func (g *group) HEAD(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.ADD(HEAD, path, handlers...)
	return g
}

func (g *group) OPTIONS(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.ADD(OPTIONS, path, handlers...)
	return g
}

func (g *group) PATCH(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.ADD(PATCH, path, handlers...)
	return g
}

func (g *group) CONNECT(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.ADD(CONNECT, path, handlers...)
	return g
}

func (g *group) TRACE(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}
	g.ADD(TRACE, path, handlers...)
	return g
}

func (g *group) Static(path string, filepath string) {
	g.StaticFS(path, http.Dir(filepath))
}

func (g *group) StaticFS(path string, fs http.FileSystem) IRouter {
	prePath := joinPath(g.base, path)
	if strings.Contains(prePath, ":") || strings.Contains(prePath, "*") {
		panic("the path of static resource can't contain path params")
	}
	fullPath := joinPath(prePath, "/*filepath")
	rawHandler := http.StripPrefix(prePath, http.FileServer(fs))
	handler := func(c *Context) {
		rawHandler.ServeHTTP(c.Response, c.Request)
	}
	g.GET(fullPath, handler)
	g.HEAD(fullPath, handler)
	return g
}

func (g *group) ANY(path string, handlers ...HandlerFunc) IRouter {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("handler nil:%s", path))
	}

	for k := range methods {
		g.ADD(k, path, handlers...)
	}

	return g
}

func printRouteInfo(method, path, handlerName string, handlerNum int) {
	var methodColor, resetColor string
	if logger.IsTTY() {
		methodColor = colorForMethod(method)
		resetColor = resetClr
	}
	logger.Printf("[route] %s %-6s%s %-25s ==> %s (%d handlers)", methodColor, method, resetColor, path, handlerName, handlerNum)
}
