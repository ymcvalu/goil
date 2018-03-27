package goil

import (
	"errors"
	"net/http"
)

type (
	Context struct {
		request *http.Request
		writer  *responseWriter
		chain   *Middleware
		params  Params
		err     error
	}
)

//执行 middleware chain 的下一个节点
//仅用于 middleware 中执行
func (ctx *Context) NextWithError() (err error) {
	chain := ctx.chain
	if ctx.chain == nil {
		err = errors.New("nil chain")
		ctx.err = err
		return
	}
	handler := chain.handler
	ctx.chain = chain.next
	handler(ctx)
	err = ctx.err
	return
}

//执行 middleware chain 的下一个节点
//仅用于 middleware 中执行
func (ctx *Context) Next() {
	chain := ctx.chain
	if ctx.chain == nil {
		ctx.err = errors.New("nil chain")
		return
	}
	handler := chain.handler
	ctx.chain = chain.next
	handler(ctx)
	return
}

//获取 middleware chain 的下一个节点 handler
//仅用于 middleware 中执行
func (ctx *Context) NextCall() (handler HandlerFunc) {
	chain := ctx.chain
	if chain == nil {
		return
	}
	handler = chain.handler
	ctx.chain = chain.next
	return
}
