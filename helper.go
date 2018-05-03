package goil

import (
	"fmt"
	"reflect"
	"runtime"
)

func assert1(guard bool, msg string) {
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

func funcName(iface interface{}) string {
	fv := valueOf(iface)
	assert1(fv.Kind() == reflect.Func, "the param of funcname must be a function")
	pc := fv.Pointer()
	return runtime.FuncForPC(pc).Name()
}
