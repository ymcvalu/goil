package goil

import (
	"net/http"
	"sync"
)

type App struct {
	router
	pool sync.Pool
}

func New() *App {
	return &App{
		router: *newRouter(),
		pool: sync.Pool{
			New: func() interface{} {
				return &Context{}
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
		ctx := app.pool.Get().(*Context)
		ctx.chain = chain
		ctx.response = newResponse(w)
		ctx.request = r
		ctx.params = params
		ctx.Next()
		app.pool.Put(ctx)

		return
	}
	//
	if tsr {

	}
}

func (app *App) Run(addr string) error {
	return http.ListenAndServe(addr, app)
}
