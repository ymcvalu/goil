/**
 * 声明路由处理函数和中间件处理函数类型
 */
package goil

import (
	"net/http"
)

type HandlerFunc func(*Context)
type HandlerChain []HandlerFunc

func combineChain(chain HandlerChain, handlers ...HandlerFunc) HandlerChain {
	if len(handlers) == 0 || handlers == nil {
		return chain
	}
	if len(chain) == 0 || chain == nil {
		return handlers
	}
	hc := make(HandlerChain, len(chain)+len(handlers))
	copy(hc, chain)
	copy(hc[len(chain):], handlers)
	return hc
}

func NoHandler(c *Context) {
	c.Status(http.StatusNotFound)
}
