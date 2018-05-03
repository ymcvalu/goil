package goil

import (
	"goil/log"
	"time"
)

type Logger interface {
	Printf(format string, msg ...interface{})
	Infof(format string, msg ...interface{})
	Debugf(format string, msg ...interface{})
	Warnf(format string, msg ...interface{})
	Errorf(format string, msg ...interface{})
	Panicf(format string, msg ...interface{})
	Fatalf(format string, msg ...interface{})
	Info(msg ...interface{})
	Print(msg ...interface{})
	Debug(msg ...interface{})
	Warn(msg ...interface{})
	Error(msg ...interface{})
	Panic(msg ...interface{})
	Fatal(msg ...interface{})
	IsTTY() bool
}

var logger Logger = log.DefLogger

func SetLogger(l Logger) {
	if l != nil {
		logger = l
	}

}

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
			resetColor = reset
		}
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		if query != "" {
			path += "?" + query
		}
		logger.Printf("[Goil] %v |%s %3d %s| %13v | %15s |%s %-7s %s %s",
			ed.Format("2006/01/02 15:04:05"),
			codeColor, status, resetColor,
			latency,
			clientIP,
			methodColor, method, resetColor,
			path,
		)
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS", "CONNECT", "TRACE":
		return white
	default:
		return reset
	}
}

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)
