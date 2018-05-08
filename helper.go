package goil

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

func assert1(guard bool, msg interface{}) {
	if !guard {
		panic(msg)
	}
}

func assert(guard bool) {
	if !guard {
		_, file, line := callerInfo(2)
		panic(fmt.Sprintf("assert failed at %s:%d", file, line))
	}
}

func callerInfo(skip int) (uintptr, string, int) {
	pc, file, line, ok := runtime.Caller(skip)
	if ok {
		return pc, file, line
	}
	return 0, "???", 0
}

func stackInfo(skip int) []byte {
	w := bytes.NewBuffer(nil)
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		name := function(pc)
		fmt.Fprintf(w, "%s:%d (0x%x) %s\n", file, line, pc, name)
	}
	return w.Bytes()
}

func function(pc uintptr) string {
	name := runtime.FuncForPC(pc).Name()
	if idx := strings.LastIndexByte(name, '/'); idx >= 0 {
		name = name[idx+1:]
	}
	if idx := strings.Index(name, "."); idx >= 0 {
		name = name[idx+1:]
	}
	return strings.Replace(name, "·", ".", -1)
}

func funcName(iface interface{}) string {
	fv := valueOf(iface)
	assert1(fv.Kind() == reflect.Func, "the param of funcname must be a function")
	pc := fv.Pointer()
	return runtime.FuncForPC(pc).Name()
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return greenBkg
	case code >= 300 && code < 400:
		return whiteBkg
	case code >= 400 && code < 500:
		return yellowBkg
	default:
		return redBkg
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blueBkg
	case "POST":
		return cyanBkg
	case "PUT":
		return yellowBkg
	case "DELETE":
		return redBkg
	case "PATCH":
		return greenBkg
	case "HEAD":
		return magentaBkg
	case "OPTIONS", "CONNECT", "TRACE":
		return whiteBkg
	default:
		return resetClr
	}
}

func genKey(s string) string {
	if s == "" {
		return s
	}
	if len(s) == 0 {
		return strings.ToLower(s)
	}
	return strings.ToLower(s[0:1]) + s[1:]
}
