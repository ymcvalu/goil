package goil

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
)

type Context struct {
	Request  *http.Request
	Response Response
	chain    *Middleware
	params   Params
	err      error
}

const (
	CONTENT_TYPE = "Content-Type"
)

const (
	MIME_TEXT = "text/plain"
	MIME_JSON = "application/json"
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

func (c *Context) String(str string) {
	c.Body(MIME_TEXT, []byte(str))
}

func (c *Context) JSON(_json interface{}) {
	byts, err := json.Marshal(_json)
	if err != nil {
		panic(err)
	}
	c.Body(MIME_JSON, byts)
}

func (c *Context) IndentJSON(_json interface{}) {
	byts, err := json.MarshalIndent(_json, "", " ")
	if err != nil {
		panic(err)
	}
	c.Body(MIME_JSON, byts)
}

//TODO:the prefix can config
const prefix = "for(;;)"

func (c *Context) SecuryJSON(_json interface{}) {
	byts, err := json.Marshal(_json)
	if err != nil {
		panic(err)
	}
	l := len(prefix) + len(byts)
	buf := make([]byte, l)
	copy(buf[:len(prefix)], []byte(prefix))
	copy(buf[len(prefix):], byts)
	c.Body(MIME_JSON, buf)
}

func (c *Context) Status(code int) {
	c.Response.WriteHeader(code)
}

func (c *Context) Body(contentType string, body []byte) {
	c.Response.SetHeader(CONTENT_TYPE, contentType)
	if _, err := c.Response.Write(body); err != nil {
		panic(err)
	}
}

//if the r implements io.Closer and the autoClose is true,then the r will be closed
func (c *Context) Stream(contentType string, r io.Reader, autoClose bool) {
	if autoClose {
		if closer, ok := r.(io.Closer); ok {
			defer closer.Close()
		}
	}
	c.Response.SetHeader(CONTENT_TYPE, contentType)
	buf := make([]byte, 512)
	n := 0
	var err error
	for {
		n, err = r.Read(buf)
		if err != nil && n > 0 {
			if _, e := c.Response.Write(buf[:n]); err != nil {
				panic(e)
			}
		} else if err != nil {
			panic(err)
		} else {
			break
		}
	}

}

func (c *Context) Bind(iface interface{}) {
	val := reflect.ValueOf(iface)
	if val.Type().Kind() != reflect.Ptr {
		panic("the param of Bind must be a pointer")
	}
	switch c.GetHeader().Get(CONTENT_TYPE) {
	case MIME_JSON:
		_json, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(_json, iface)
	default:
	}

	return
}

func (c *Context) GetHeader() http.Header {
	return c.Request.Header
}
