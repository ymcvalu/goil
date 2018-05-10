package goil

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"goil/logger"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Context struct {
	Request  *http.Request
	resp     response
	Response Response
	chain    HandlerChain
	idx      int
	params   Params
	//ErrMsg and ErrCode is used pass err info among middlewares
	ErrMsg  error
	ErrCode int
	values  *concurrentMap
}

//执行 middleware chain 的下一个节点
//仅用于 middleware 中执行
func (ctx *Context) Next() {
	chain := ctx.chain
	if ctx.chain == nil || ctx.idx >= len(chain) {
		ctx.ErrMsg = NoHandlers
		ctx.ErrCode = Code_NoHandlers
		return
	}

	for ctx.idx < len(chain) {
		handler := chain[ctx.idx]
		ctx.idx++
		handler(ctx)
	}
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

func (c *Context) ReqBody() io.Reader {
	//server will close the body auto
	return c.Request.Body
}

//get request headers
func (c *Context) Headers() http.Header {
	return c.Request.Header
}

//get request header by key
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

//set header to response
func (c *Context) SetHeader(key, value string) {
	c.Response.SetHeader(key, value)
}

//get request cookie by name
func (c *Context) GetCookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	value, _ := url.QueryUnescape(cookie.Value)
	return value, nil
}

//set cookie to response
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Response, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

func (c *Context) Flush() {
	c.Response.Flush()
}

func (c *Context) Hijack() (conn net.Conn, io *bufio.ReadWriter, err error) {
	return c.Response.Hijack()
}

func (c *Context) CloseNotify() <-chan bool {
	return c.Response.CloseNotify()
}

func (c *Context) Param(key string) (value string) {
	value, _ = c.params.get(key)
	return
}

func (c *Context) DefParam(key string, def string) string {
	if value, exist := c.params.get(key); exist {
		return value
	}
	return def
}

func (c *Context) Query(key string) string {
	values := c.Request.URL.Query()
	return values.Get(key)

}

func (c *Context) DefQuery(key string, def string) string {
	values := c.Request.URL.Query()
	value := values.Get(key)
	if value != "" {
		return value
	}
	return def
}

func (c *Context) BindQuery(iface interface{}) error {
	err := bindQueryParams(c.Request, iface)
	if err != nil {
		logger.Errorf("when binding params: %s", err)
		return ParamsBindingError
	}
	legal, err := validate(iface)
	if err != nil {
		logger.Errorf("when validating params: %s", err)
		return ParamsValidateError
	}
	if !legal {
		return ParamsInvalidError
	}
	return nil
}

func (c *Context) PostForm() url.Values {
	c.Request.ParseForm()
	return c.Request.PostForm
}

func (c *Context) Form() url.Values {
	c.Request.ParseForm()
	return c.Request.Form
}

func (c *Context) PostValue(key string) string {
	return c.Request.PostFormValue(key)
}

func (c *Context) DefPostValue(key, def string) string {
	value := c.Request.PostFormValue(key)
	if value != "" {
		return value
	}
	return def
}

func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

func (c *Context) DefFormValue(key, def string) string {
	value := c.Request.FormValue(key)
	if value != "" {
		return value
	}
	return def
}

func (c *Context) SaveFile(name, dest string) error {
	_, fh, err := c.Request.FormFile(name)
	if err != nil {
		return err
	}
	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

//bind path param,form param,query param,file
func (c *Context) Bind(iface interface{}) error {
	err := bind(c, iface)
	if err != nil {
		logger.Errorf("when binding params: %s", err)
		return ParamsBindingError
	}
	legal, err := validate(iface)

	if err != nil {
		logger.Errorf("when validating params: %s", err)
		return ParamsValidateError
	}
	if !legal {
		return ParamsInvalidError
	}
	return nil
}

//rewrite the response code
func (c *Context) Status(status int) {
	c.Response.SetStatus(status)
}

func (c *Context) ContentType(contentType string) {
	c.Response.SetHeader(CONTENT_TYPE, contentType)
}

func (c *Context) Html(name string, data interface{}) {
	c.Render(HtmlRender, VM(name, data))
}

//write the raw text
func (c *Context) Text(str string) {
	c.Body(MIME_TEXT, []byte(str))
}

//wirte json
func (c *Context) JSON(data interface{}) {
	c.Render(JsonRender, data)
}

//wirte json with a prefix
func (c *Context) SecureJSON(data interface{}) {
	c.Render(SecJsonRender, data)
}

//write xml
func (c *Context) Xml(data interface{}) {
	c.Render(xmlRender, data)
}

//Render the render
func (c *Context) Render(r Render, content interface{}) {
	contentType := r.ContentType()
	//TODO:check the content-type???
	c.ContentType(contentType)
	err := r.Render(c.Response, content)
	if err != nil {
		panic(err)
	}
}

func (c *Context) File(filepath string) {
	http.ServeFile(c.Response, c.Request, filepath)
}

func (c *Context) IndentJSON(content interface{}) {
	byts, err := json.MarshalIndent(content, "", " ")
	if err != nil {
		panic(err)
	}
	c.Body(MIME_JSON, byts)
}

//write contentType and body
func (c *Context) Body(contentType string, body []byte) {
	r := bytes.NewReader(body)
	c.Stream(contentType, r, false)
}

//write content from reader
//if the r implements io.Closer and the autoClose is true,then the r will be closed
func (c *Context) Stream(contentType string, r io.Reader, autoClose bool) {
	if autoClose {
		if closer, ok := r.(io.Closer); ok {
			defer closer.Close()
		}
	}
	c.Response.SetHeader(CONTENT_TYPE, contentType)
	_, err := io.Copy(c.Response, r)
	if err != nil {
		panic(err)
	}
}

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
	if c.values == nil {
		return nil, false
	}
	val, exists = c.values.get(key)
	return
}

func (c *Context) GetDef(key string, def interface{}) interface{} {
	if c.values == nil {
		return def
	}
	val, exists := c.values.get(key)
	if exists {
		return val
	}
	return def
}

func (c *Context) Set(key string, value interface{}) {
	if c.values == nil {
		c.values = cmNil.new()
	}
	c.values.set(key, value)
}

func (c *Context) Del(key string) {
	if c.values == nil {
		return
	}
	c.values.del(key)
}

func (c *Context) clear() {
	c.Request = nil
	c.resp.clear()
	c.chain = nil
	c.values = nil
	c.params = nil
	c.ErrMsg = nil
	c.ErrCode = 0
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Request = r
	c.resp.reset(w)
}

//assert implements context.Context
var _ context.Context = new(Context)

func (c *Context) Redirect(code int, location string) {
	c.SetHeader("location", location)
	c.Response.WriteHeader(code)
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *Context) Done() <-chan struct{} {
	return nil
}

func (c *Context) Err() error {
	return c.ErrMsg
}

func (c *Context) Value(key interface{}) interface{} {
	val, _ := c.values.get(key)
	return val
}
