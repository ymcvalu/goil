package goil

import (
	. "goil/reflect"
)

type groupX struct {
	group
	errHandler  HandlerFunc
	respHandler HandlerFunc
}

func handlerWrpper(fun interface{}) HandlerFunc {
	ins := FuncIn(fun)
	outs := FuncOut(fun)
	if len(ins) > 2 || len(outs) > 2 {
		panic("unsupport fun for wrapping,the count of in or out params must not be greater than 2")
	}

	return nil
}
