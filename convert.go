package goil

import (
	"reflect"
	"strconv"
)

type Convert func(value string, dTyp reflect.Type) (interface{}, error)

var convertFunc = map[string]Convert{
	"_a2i": func(value string, dType reflect.Type) (interface{}, error) {

		return strconv.ParseInt(value, 10, 64)

	},
	"_a2b": func(value string, dType reflect.Type) (interface{}, error) {

		return strconv.ParseBool(value)

	},
	"_a2u": func(value string, dType reflect.Type) (interface{}, error) {

		return strconv.ParseUint(value, 10, 64)

	},
	"_a2f": func(value string, dType reflect.Type) (interface{}, error) {

		return strconv.ParseFloat(value, 64)

	},
}

func RegisterConvert(name string, fun Convert) {
	guard.execSafely(func() {
		convertFunc[name] = fun
	})

}
