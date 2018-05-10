/**
 * 声明路由处理函数和中间件处理函数类型
 */
package goil

import (
	"net/http"
)

type HandlerFunc = func(*Context)
type HandlerChain = []HandlerFunc

func combineChain(chain HandlerChain, handlers ...HandlerFunc) HandlerChain {
	if len(handlers) == 0 {
		return chain
	}
	if len(chain) == 0 {
		return handlers
	}
	hc := make(HandlerChain, len(chain)+len(handlers))
	copy(hc, chain)
	copy(hc[len(chain):], handlers)
	return hc
}

func NotFoundHandler(c *Context) {
	c.Status(http.StatusNotFound)
}

func NotMethodHandler(c *Context) {
	c.Status(http.StatusNotFound)
}
