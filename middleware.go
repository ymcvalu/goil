package goil

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"
)

//a middleware to print request info
func PrintRequestInfo() HandlerFunc {
	isTerm := logger.IsTTY()

	return func(c *Context) {
		st := time.Now()
		c.Next()
		ed := time.Now()
		latency := ed.Sub(st)
		clientIP := c.ClientIP()
		var codeColor, methodColor, resetColor string
		status := c.Response.Status()
		method := c.Request.Method
		if isTerm {
			codeColor = colorForStatus(status)
			methodColor = colorForMethod(method)
			resetColor = resetClr
		}
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		if query != "" {
			path += "?" + query
		}
		logger.Printf("[Goil] %v |%s %3d %s| %13v | %15s |%s %-5s %s %s",
			ed.Format("2006/01/02 15:04:05"),
			codeColor, status, resetColor,
			latency,
			clientIP,
			methodColor, method, resetColor,
			path,
		)
	}
}

func ReverseProxy(proxy func(c *http.Request)) HandlerFunc {
	rp := httputil.ReverseProxy{
		Director: proxy,
	}
	return func(c *Context) {
		rp.ServeHTTP(c.Response, c.Request)
	}
}

type gzipResponse struct {
	Response
	level int
}

func (w *gzipResponse) Write(bytes []byte) (n int, err error) {
	zw, _ := gzip.NewWriterLevel(w.Response, w.level)
	zw.ModTime = time.Now().UTC()
	n, err = zw.Write(bytes)
	err = zw.Close()
	return
}

func EnableGzip(level int) HandlerFunc {
	if level < GZIP_HuffmanOnly || level > GZIP_BestCompression {
		panic(fmt.Sprintf("gzip: invalid compression level: %d", level))
	}
	return func(c *Context) {

		if level != GZIP_NoCompression {
			zp := &gzipResponse{
				Response: c.Response,
				level:    level,
			}
			c.SetHeader(CONTENT_ENCODING, "gzip")
			//replace the resp
			c.Response = zp
			c.Next()
			//restore the resp
			c.Response = zp.Response
		} else {
			c.Next()
		}

	}
}

func Recover() HandlerFunc {
	return func(c *Context) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			reqInfo, _ := httputil.DumpRequest(c.Request, false)
			var fontColor, reset string
			if logger.IsTTY() {
				fontColor = redFont
				reset = resetClr
			}
			logger.Printf("%s[recover] %s\n%s\n%s%s", fontColor, err, string(reqInfo), string(stackInfo(5)), reset)
		}()
		c.Next()
	}
}
