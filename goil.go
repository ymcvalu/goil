package goil

import (
	"net/http"
	"sync"
)

type App struct {
	IRouter
	pool sync.Pool
}

func New() *App {
	return &App{
		IRouter: newRouter(),
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

}
