package goil

import (
	"fmt"
	. "goil/reflect"
	"mime"
	"net/http"
	"reflect"
)

type ErrorHandler func(*Context, error)
type RenderHandler func(*Context, interface{})

func deref(typ Type) Type {
	for typ.Kind() == Ptr {
		typ = typ.Elem()
	}
	return typ
}

type GroupX struct {
	group         group
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

func DefErrHandler(c *Context, err error) {
	c.Status(http.StatusInternalServerError)
	c.Text("system error.")
}

func DefRenderHandler(c *Context, data interface{}) {
	if vm, ok := data.(ViewModel); ok {
		c.Html(vm.Name, vm.Model)
	}
	accept := c.Header(ACCEPT)
	if accept == "" {
		c.JSON(data)
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
