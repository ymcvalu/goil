package goil

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
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

func (c *Context) ReqBody() io.Reader {
	//server will close the body auto
	return c.Request.Body
}

//get request headers
func (c *Context) Headers() http.Header {
	return c.Request.Header
}

//get request header by key
func (c *Context) GetHeader(key string) string {
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

//rewrite the response code
func (c *Context) Status(code int) {
	c.Response.WriteHeader(code)
}

//write the raw text
func (c *Context) String(str string) {
	c.Body(MIME_TEXT, []byte(str))
}

//wirte json
func (c *Context) JSON(_json interface{}) {
	byts, err := json.Marshal(_json)
	if err != nil {
		panic(err)
	}
	c.Body(MIME_JSON, byts)
}

//write indent json
func (c *Context) IndentJSON(_json interface{}) {
	byts, err := json.MarshalIndent(_json, "", " ")
	if err != nil {
		panic(err)
	}
	c.Body(MIME_JSON, byts)
}

//wirte json wite a prefix
func (c *Context) SecureJSON(_json interface{}) {
	byts, err := json.Marshal(_json)
	if err != nil {
		panic(err)
	}
	l := len(secure_json_prefix) + len(byts)
	buf := make([]byte, l)
	copy(buf[:len(secure_json_prefix)], []byte(secure_json_prefix))
	copy(buf[len(secure_json_prefix):], byts)
	c.Body(MIME_JSON, buf)
}

//write contentType and body
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

//write content from reader
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

func (c *Context) Del(key string) {
	delete(c.values, key)
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
