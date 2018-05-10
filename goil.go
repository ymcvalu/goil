package goil

import (
	"goil/logger"
	"net/http"
	"sync"
)

type App struct {
	router      *router
	contextPool sync.Pool
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
	}
}

//assert App implements http.Handler
var _ http.Handler = new(App)

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method
	chain, params, tsr := app.router.route(method, path)
	if chain != nil {
		ctx := app.getCtx(w, r)
		//init the context
		ctx.chain = chain
		ctx.params = params
		ctx.idx = 0
		ctx.Next()
		//detach
		app.putCtx(ctx)
		return
	}
	//
	if tsr {

	}
}

func (app *App) getCtx(w http.ResponseWriter, r *http.Request) *Context {
	ctx := app.contextPool.Get().(*Context)
	if ctx.Response == nil {
		//logger.Info("new response")
		ctx.Response = newResponse()
	}
	ctx.Response.reset(w)
	ctx.Request = r
	return ctx
}

func (app *App) putCtx(ctx *Context) {
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
	app.router.Use(Recover(), PrintRequestInfo())
	return app
}

func (a *App) Group(path string, handlers ...HandlerFunc) IRouter {
	return a.router.Group(path, handlers...)
}
func (a *App) Use(handlers ...HandlerFunc) IRouter {
	return a.router.Use(handlers...)
}
func (a *App) ADD(method, path string, handler ...HandlerFunc) IRouter {
	return a.router.ADD(method, path, handler...)
}
func (a *App) GET(path string, handlers ...HandlerFunc) IRouter {
	return a.router.GET(path, handlers...)
}
func (a *App) POST(path string, handlers ...HandlerFunc) IRouter {
	return a.router.POST(path, handlers...)
}
func (a *App) PUT(path string, handlers ...HandlerFunc) IRouter {
	return a.router.PUT(path, handlers...)
}
func (a *App) DELETE(path string, handlers ...HandlerFunc) IRouter {
	return a.router.DELETE(path, handlers...)
}
func (a *App) OPTIONS(path string, handlers ...HandlerFunc) IRouter {
	return a.router.OPTIONS(path, handlers...)
}
func (a *App) PATCH(path string, handlers ...HandlerFunc) IRouter {
	return a.router.PATCH(path, handlers...)
}
func (a *App) CONNECT(path string, handlers ...HandlerFunc) IRouter {
	return a.router.CONNECT(path, handlers...)
}
func (a *App) TRACE(path string, handlers ...HandlerFunc) IRouter {
	return a.router.TRACE(path, handlers...)

}
func (a *App) ANY(path string, handlers ...HandlerFunc) IRouter {
	return a.router.ANY(path, handlers...)

}
func (a *App) Static(path string, filepath string) IRouter {
	return a.router.Static(path, filepath)
}
func (a *App) StaticFS(path string, fs http.FileSystem) IRouter {
	return a.router.StaticFS(path, fs)
}

func (a *App) XRouter() *GroupX {
	return &GroupX{
		group:         a.router.group,
		ErrorHandler:  DefErrHandler,
		RenderHandler: DefRenderHandler,
	}
}
