package goil

import (
	"net/http"
	"sync"
)

type App struct {
	*router
	contextPool sync.Pool
	respPool    sync.Pool
}

func New() *App {
	echoBanner()
	runmode := RunMode()
	logger.Printf("[Goil] the app is running in %s mode", runmode)
	if runmode == DBG {
		logger.Printf("[Goil] you can change the run mode by setting the env: export %s=%s", ENV_KEY, PRD)
	}
	return &App{
		router: newRouter(),
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
		ctx.Logger = logger
		ctx.chain = chain
		ctx.idx = 0
		ctx.Response = app.respPool.Get().(Response)
		ctx.Response.reset(w)
		ctx.Request = r
		ctx.params = params
		ctx.values = make(map[string]interface{})
		ctx.Next()

		//detach
		ctx.values = nil
		ctx.params = nil
		ctx.Request = nil
		ctx.chain = nil
		ctx.Logger = nil
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
	starting()
	logger.Printf("[Goil] Listening and serving HTTP on %s\n", addr)
	return http.ListenAndServe(addr, app)
}

const banner = `` +
	`     __________                        ` + "\n" +
	`    / ________/         ______         ` + "\n" +
	`   / / _____  _______  /__/  /         ` + "\n" +
	`  / / /____ \/  ___  \/  /  /          ` + "\n" +
	` / /______/ /  /__/  /  /  /__         ` + "\n" +
	` \_________/\_______/\_/\____/  by can ` + "\n" +
	`                                       `

func echoBanner() {
	var bannerColor, resetColor string
	if logger.IsTTY() {
		bannerColor = red
		resetColor = reset
	}
	logger.Printf("%s%s%s", bannerColor, banner, resetColor)
}
