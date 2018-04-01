package goil

import (
	"net/http"
	"sync"
)

type App struct {
	router
	contextPool sync.Pool
	respPool    sync.Pool
}

func New() *App {
	return &App{
		router: *newRouter(),
		contextPool: sync.Pool{
			New: func() interface{} {
				return &Context{}
			},
		},
		respPool: sync.Pool{
			New: func() interface{} {
				return newResponse()
			},
		},
	}
}

//assert App implements http.Handler
var _ http.Handler = new(App)

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method
	chain, params, tsr := app.route(method, path)
	//
	if chain != nil {
		//init the context
		ctx := app.contextPool.Get().(*Context)
		ctx.chain = chain
		ctx.Response = app.respPool.Get().(*response)
		ctx.Response.reset(w)
		ctx.Request = r
		ctx.params = params
		ctx.Next()
		//detach
		ctx.Request = nil
		resp := ctx.Response
		ctx.Response = nil
		app.contextPool.Put(ctx)
		app.respPool.Put(resp.clear())
		resp = nil
		return
	}
	//
	if tsr {

	}
}

func (app *App) Run(addr string) error {
	return http.ListenAndServe(addr, app)
}
