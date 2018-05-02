package goil

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

type Context struct {
	Logger
	Request  *http.Request
	Response Response
	chain    HandlerChain
	idx      int
	params   Params
	errCode  int
	err      error
	values   map[string]interface{}
}

const (
	CONTENT_TYPE = "Content-Type"
)

//执行 middleware chain 的下一个节点
//仅用于 middleware 中执行
func (ctx *Context) Next() {
	chain := ctx.chain
	if ctx.chain == nil || ctx.idx >= len(chain) {
		ctx.err = errors.New("no handler")
		return
	}
	handler := chain[ctx.idx]
	ctx.idx++

	handler(ctx)
	return
}

//获取 middleware chain 的下一个节点 handler
//仅用于 middleware 中执行
func (ctx *Context) NextCall() (handler HandlerFunc) {
	chain := ctx.chain
	if chain == nil || ctx.idx >= len(chain) {
		return
	}
	handler = chain[ctx.idx]
	ctx.idx++
	return
}

func (ctx *Context) Abort() {
	ctx.idx = len(ctx.chain)
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

//Render the render
func (c *Context) Render(r Render, content interface{}) {
	c.Stream(r.ContentType(), r.Render(content), true)
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

func (c *Context) Bind(iface interface{}) error {
	err := bind(c, iface)
	if err != nil {
		return err
	}
	legal, err := validate(iface)

	if err != nil {
		return err
	}
	if !legal {
		return ParamsInvalidError
	}
	return nil
}

func (c *Context) BindQuery(iface interface{}) error {
	err := bindQueryParams(c.Request, iface)
	if err != nil {
		return err
	}
	legal, err := validate(iface)
	if err != nil {
		return err
	}
	if !legal {
		return ParamsInvalidError
	}
	return nil
}

var ParamsInvalidError = errors.New("params validate failed.")

func (c *Context) Param(key string) (value string, exist bool) {
	value, exist = c.params[key]
	return
}

func (c *Context) DefParam(key string, def string) string {
	if value, exist := c.params[key]; exist {
		return value
	}
	return def
}

func (c *Context) Query(key string) (string, bool) {
	vals, exist := c.Request.Form[key]
	return vals[0], exist
}

func (c *Context) DefQuery(key string, def string) string {
	if vals, exist := c.Request.PostForm[key]; exist {
		return vals[0]
	}
	return def
}

func (c *Context) BodyReader() io.Reader {
	//server will close the body auto
	return c.Request.Body
}

func (c *Context) Headers() http.Header {
	return c.Request.Header
}

func (c *Context) ReqBody() ([]byte, error) {
	return ioutil.ReadAll(c.Request.Body)
}

// ClientIP implements a best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
// Use X-Forwarded-For before X-Real-Ip as nginx uses X-Real-Ip with the proxy's IP.
//TODO:add reverse proxy toggle
func (c *Context) ClientIP() string {

	clientIP := c.Headers().Get("X-Forwarded-For")
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}
	clientIP = strings.TrimSpace(clientIP)
	if clientIP != "" {
		return clientIP
	}
	clientIP = strings.TrimSpace(c.Headers().Get("X-Real-Ip"))
	if clientIP != "" {
		return clientIP
	}

	if addr := c.Headers().Get("X-Appengine-Remote-Addr"); addr != "" {
		return addr
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

func (c *Context) Get(key string) (val interface{}, exists bool) {
	val, exists = c.values[key]
	return
}

func (c *Context) GetDef(key string, def interface{}) interface{} {
	val, exists := c.values[key]
	if exists {
		return val
	}
	return def
}

func (c *Context) Set(key string, value interface{}) {
	c.values[key] = value
}

//get session
func (c *Context) Session() SessionEntry {
	val, ok := c.values[sessionTag]
	if !ok {
		return nil
	}
	if sess, ok := val.(SessionEntry); ok {
		return sess
	}
	return nil
}
