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
	if chain != nil {
		ctx := app.getCtx(w, r)
		//init the context
		ctx.chain = chain
		ctx.params = params
		ctx.values = make(map[interface{}]interface{})
		ctx.Next()
		//detach
		app.putCtx(ctx)
		return
	}
	//
	if tsr {

	}
	//handle the 404 not found
	ctx := app.getCtx(w, r)
	//use the global middleware,include print request
	ctx.chain = append(ctx.chain, app.middlewares...)
	ctx.chain = append(ctx.chain, NoHandler)
	ctx.Next()
	app.putCtx(ctx)
}

func (app *App) getCtx(w http.ResponseWriter, r *http.Request) *Context {
	ctx := app.contextPool.Get().(*Context)
	ctx.Response = app.respPool.Get().(Response)
	ctx.Response.reset(w)
	ctx.Request = r
	ctx.idx = 0
	ctx.Logger = logger

	return ctx
}

func (app *App) putCtx(ctx *Context) {
	resp := ctx.Response
	if _, ok := resp.(*response); ok {
		app.respPool.Put(resp.clear())
	}
	resp = nil
	ctx.clear()
	app.contextPool.Put(ctx)
}

func (app *App) Run(addr string) (err error) {
	guard.run()
	logger.Printf("[Goil] Listening and serving HTTP on %s\n", addr)
	err = http.ListenAndServe(addr, app)
	return
}

func (app *App) RunTLS(addr string, certFile, keyFile string) (err error) {
	guard.run()
	logger.Printf("[Goil] Listening and serving HTTPS on %s\n", addr)
	err = http.ListenAndServeTLS(addr, certFile, keyFile, app)
	return
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
		bannerColor = redBkg
		resetColor = resetClr
	}
	logger.Printf("%s%s%s", bannerColor, banner, resetColor)
}

func Default() *App {
	app := New()
	app.Use(Recover(), PrintRequestInfo())
	return app
}
