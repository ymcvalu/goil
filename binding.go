package goil

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type Binding interface {
	Bind(c *Context, iface interface{}) error
}

type Validator func(value interface{}, params []string) error

type Convert func(value string) (interface{}, error)

var convertFunc = map[string]Convert{
	"_a2i": func(value string) (interface{}, error) {
		return strconv.ParseInt(value, 10, 64)
	},
	"_a2b": func(value string) (interface{}, error) {
		return strconv.ParseBool(value)
	},
	"_a2u": func(value string) (interface{}, error) {
		return strconv.ParseUint(value, 10, 64)
	},
	"_a2f": func(value string) (interface{}, error) {
		return strconv.ParseFloat(value, 64)
	},
}

func RegisterConvert(name string, fun Convert) bool {
	if _, conflict := convertFunc[name]; conflict {
		return false
	}
	convertFunc[name] = fun
	return true
}

func pathParamsBinding(ctx Context, iface interface{}) (err error) {
	if ctx.params == nil || len(ctx.params) == 0 {
		return
	}

	if !isPtr(iface) {
		err = errors.New(`the params for binding must be a pointer`)
		return
	}
	val := valueOf(iface).Elem()
	typ := val.Type()
	switch val.Kind() {
	case reflect.Struct:
		for i, n := 0, typ.NumField(); i < n; i++ {
			fTyp := typ.Field(i)
			tag := fTyp.Tag
			pKey := tag.Get("path")
			pVal, exist := ctx.params[pKey]
			if !exist {
				continue
			}
			fVal := val.Field(i)
			if !fVal.CanSet() {
				continue
			}
			err = bindField(pVal, fVal, fTyp)
			if err != nil {
				return
			}
		}
		//case reflect.Map:
	}
	return
}

func bindField(src string, dest reflect.Value, fTyp reflect.StructField) error {
	tag := fTyp.Tag
	conv := tag.Get("convert")
	if convFunc, exists := convertFunc[conv]; exists {
		val, err := convFunc(src)
		if err != nil {
			return err
		}
		if dest.CanSet() {
			dest.Set(valueOf(val))
		}
		return nil
	}
	switch dest.Type().Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		conv = "_a2i"

	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		conv = "_a2u"
	case reflect.Bool:
		conv = "_a2b"
	case reflect.Float32, reflect.Float64:
		conv = "_a2f"
	default:
		return fmt.Errorf("unsupport type for binding params %s to %s", src, dest)
	}
	val, err := convertFunc[conv](src)
	if err != nil {
		return fmt.Errorf("when binding params %v to %v:%s", src, dest, err)
	}
	if dest.CanSet() {
		switch dest.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			dest.SetInt(val.(int64))
		case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			dest.SetUint(val.(uint64))
		case reflect.Bool:
			dest.SetBool(val.(bool))
		case reflect.Float32, reflect.Float64:
			dest.SetFloat(val.(float64))
		}
	}
	return nil
}
