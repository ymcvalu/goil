package goil

import (
	"fmt"
	"goil/logger"
	. "goil/reflect"
	"mime"
	"net/http"
	"reflect"
)

type ErrorHandler func(*Context, error)
type RenderHandler func(*Context, interface{})

type XRouter interface {
	Group(path string, handlers ...HandlerFunc) XRouter
	Use(handlers ...HandlerFunc) XRouter
	ADD(method, path string, handler ...interface{}) XRouter
	GET(path string, handlers ...interface{}) XRouter
	POST(path string, handlers ...interface{}) XRouter
	PUT(path string, handlers ...interface{}) XRouter
	DELETE(path string, handlers ...interface{}) XRouter
	OPTIONS(path string, handlers ...interface{}) XRouter
	PATCH(path string, handlers ...interface{}) XRouter
	CONNECT(path string, handlers ...interface{}) XRouter
	TRACE(path string, handlers ...interface{}) XRouter
	ANY(path string, handlers ...interface{}) XRouter
	Static(path string, filepath string) XRouter
	StaticFS(path string, fs http.FileSystem) XRouter
	SetRenderHandler(handler RenderHandler)
	SetErrorHandler(handler ErrorHandler)
}

var _ XRouter = new(GroupX)

type GroupX struct {
	group         *group
	ErrorHandler  ErrorHandler
	RenderHandler RenderHandler
}

var ctxTyp = TypeOf((*Context)(nil))
var errTyp = TypeOf((*error)(nil)).Elem()

func (g *GroupX) Wrapper(fun interface{}) HandlerFunc {
	illegal := func() error {
		return fmt.Errorf("the func for wrapping is illegal: %s", FuncDesc(fun))
	}

	ins := FuncIn(fun)
	outs := FuncOut(fun)
	cin := len(ins)
	cout := len(outs)
	assert1(cin <= 2 && cout <= 2, illegal())

	//the fun is HandlerFunc
	if cin == 1 && cout == 0 && ins[0] == ctxTyp {
		return fun.(func(*Context))
	}

	hasContext := false
	needBind := false
	//check the first params
	if cin > 0 {
		if ins[0] == ctxTyp {
			hasContext = true
		} else {
			//if the first in param isn't *goil.Context
			//the func only could have one in params
			it1 := ins[0]
			it1 = deref(it1)
			assert1(cin == 1 && it1.Kind() == Struct, illegal())
			needBind = true
		}
	}
	//if the in params is two, the first param must be *goil.Context and second is the params
	if cin == 2 {
		it2 := ins[1]
		assert1(hasContext && it2 != ctxTyp, illegal())
		it2 = deref(it2)
		assert1(it2.Kind() == Struct, illegal())
		needBind = true
	}

	hasError := false
	needRender := false
	if cout > 0 {
		ot1 := outs[0]
		if ot1.Implements(errTyp) {
			assert1(cout == 1, illegal())
			hasError = true
		} else {
			needRender = true
		}
	}
	if cout == 2 {
		ot2 := outs[1]
		assert1(ot2.Implements(errTyp), illegal())
		hasError = true
	}

	return func(c *Context) {
		inParams := make([]Value, 0, cin)

		if hasContext {
			inParams = append(inParams, ValueOf(c))
		}

		if needBind {
			typ := ins[0]
			if hasContext {
				typ = ins[1]
			}
			pv := reflect.New(typ)
			err := c.Bind(pv.Interface())
			if err != nil {
				g.ErrorHandler(c, err)
				return
			}
			inParams = append(inParams, pv.Elem())
		}
		outParams := ValueOf(fun).Call(inParams)
		if hasError {
			idx := 0
			if needRender {
				idx = 1
			}
			err := outParams[idx]
			if !err.IsNil() {
				g.ErrorHandler(c, err.Interface().(error))
				return
			}
		}
		if needRender {
			g.RenderHandler(c, outParams[0].Interface())
		}
	}
}

func (g *GroupX) Group(path string, handlers ...HandlerFunc) XRouter {

	return &GroupX{
		group: &group{
			middlewares: combineChain(g.group.middlewares, handlers...),
			base:        joinPath(g.group.base, path),
			router:      g.group.router,
		},
		ErrorHandler:  g.ErrorHandler,
		RenderHandler: g.RenderHandler,
	}
}

func (g *GroupX) Use(handlers ...HandlerFunc) XRouter {
	g.group.Use(handlers...)
	return g
}

func (g *GroupX) ADD(method, path string, handler ...interface{}) XRouter {
	l := len(handler)
	assert1(l > 0, fmt.Sprintf("the handler of %s is nil", path))
	if l > 1 {
		for i := 0; i < l-1; i++ {
			_, ok := handler[i].(func(*Context))
			assert1(ok, fmt.Errorf("the type of middleware must be func (*Context)"))
		}
	}
	ml := len(g.group.middlewares)
	chain := make(HandlerChain, ml, ml+l)
	copy(chain, g.group.middlewares)
	for i, h := range handler {
		if i == l-1 {
			chain = append(chain, g.Wrapper(h))
		} else {
			chain = append(chain, h.(func(*Context)))
		}
	}
	absolutePath := joinPath(g.group.base, path)

	g.group.router.add(method, absolutePath, chain)
	if RunMode() == DBG {
		handlerNum := len(chain)
		handlerName := funcName(handler[l-1])
		printRouteInfo(method, absolutePath, handlerName, handlerNum)
	}
	return g
}

func (g *GroupX) GET(path string, handlers ...interface{}) XRouter {
	return g.ADD(GET, path, handlers...)
}
func (g *GroupX) POST(path string, handlers ...interface{}) XRouter {
	return g.ADD(POST, path, handlers...)
}
func (g *GroupX) PUT(path string, handlers ...interface{}) XRouter {
	return g.ADD(PUT, path, handlers...)
}
func (g *GroupX) DELETE(path string, handlers ...interface{}) XRouter {
	return g.ADD(DELETE, path, handlers...)
}
func (g *GroupX) OPTIONS(path string, handlers ...interface{}) XRouter {
	return g.ADD(OPTIONS, path, handlers...)
}
func (g *GroupX) PATCH(path string, handlers ...interface{}) XRouter {
	return g.ADD(PATCH, path, handlers...)
}
func (g *GroupX) CONNECT(path string, handlers ...interface{}) XRouter {
	return g.ADD(CONNECT, path, handlers...)
}
func (g *GroupX) TRACE(path string, handlers ...interface{}) XRouter {
	return g.ADD(TRACE, path, handlers...)
}
func (g *GroupX) ANY(path string, handlers ...interface{}) XRouter {
	for methods := range methods {
		g.ADD(methods, path, handlers...)
	}
	return g
}

func (g *GroupX) Static(path string, filepath string) XRouter {
	g.group.Static(path, filepath)
	return g
}

func (g *GroupX) StaticFS(path string, fs http.FileSystem) XRouter {
	g.group.StaticFS(path, fs)
	return g
}

func DefErrHandler(c *Context, err error) {
	logger.Errorf("when handler reqest:%s", err)
	c.Status(http.StatusInternalServerError)
	c.Text(err.Error())
}

func DefRenderHandler(c *Context, data interface{}) {
	if vm, ok := data.(ViewModel); ok {
		c.Html(vm.Name, vm.Model)
		return
	}

	if msg, ok := data.(string); ok {
		c.Text(msg)
		return
	}

	accept := c.Header(ACCEPT)
	if accept == "" {
		c.JSON(data)
		return
	}
	mime, _, err := mime.ParseMediaType(accept)
	if err != nil {
		c.JSON(data)
	}
	switch mime {
	case MIME_XML:
		c.Xml(data)
	default:
		c.JSON(data)
	}
}

func (g *GroupX) SetRenderHandler(handler RenderHandler) {
	g.RenderHandler = handler
}
func (g *GroupX) SetErrorHandler(handler ErrorHandler) {
	g.ErrorHandler = handler
}
