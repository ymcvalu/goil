package reflect

import (
	"errors"
	"reflect"
)

type Type = reflect.Type
type Value = reflect.Value
type StructField = reflect.StructField

type Kind = reflect.Kind

const (
	Invalid reflect.Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
)

func CanComp(val interface{}) bool {
	typ := TypeOf(val)
	return typ.Comparable()
}

func ValueOf(val interface{}) reflect.Value {
	return reflect.ValueOf(val)
}

func TypeOf(val interface{}) reflect.Type {
	return reflect.TypeOf(val)
}

func IsPtr(iface interface{}) bool {
	k := reflect.TypeOf(iface).Kind()
	return k == reflect.Ptr
}

func KindOf(iface interface{}) Kind {
	return TypeOf(iface).Kind()
}

type Field struct {
	Name string
	Tag  reflect.StructTag
	Kind Kind
	Val  interface{}
}

func Fields(v interface{}) (fds []Field, err error) {
	val := ValueOf(v)

	for val.Kind() == Ptr {
		if val.IsNil() {
			return nil, errors.New("the value of param is invalid")
		}
		val = val.Elem()
	}

	if val.Kind() != Struct {
		return nil, errors.New("the type of param isn't a struct")
	}

	typ := val.Type()
walk:
	for i, n := 0, typ.NumField(); i < n; i++ {
		fTyp := typ.Field(i)
		fVal := val.Field(i)

		for fVal.Kind() == Ptr {
			if fVal.IsNil() {
				continue walk
			}
			fVal = fVal.Elem()
		}

		field := Field{
			Name: fTyp.Name,
			Kind: fVal.Kind(),
			Val:  fVal.Interface(),
			Tag:  fTyp.Tag,
		}
		if !IsZero(field.Val) {
			fds = append(fds, field)
		}

	}
	return
}

//check if the var is a zero value
func IsZero(v interface{}) bool {
	val := ValueOf(v)
	kind := val.Kind()
	switch kind {
	case Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Float32, Float64:
		return v == 0
	case String:
		return v == ""
	case Bool:
		return v == false
	case Ptr, Chan, Slice, Map:
		return val.IsNil()
	}
	return false
}

func IsStruct(v interface{}) bool {
	val := ValueOf(v)
	for val.Kind() == Ptr {
		val = val.Elem()
	}
	return val.Kind() == Struct
}

//if the v is a func
func IsFunc(v interface{}) bool {
	typ := TypeOf(v)
	return typ.Kind() == Func
}

//obtain the types of in params
//panic if param isn't a function
func FuncIn(v interface{}) []Type {
	typ := TypeOf(v)

	numIn := typ.NumIn()
	ins := make([]Type, numIn)
	for i := 0; i < numIn; i++ {
		ins[i] = typ.In(i)
	}
	return ins
}

//obtain the types of out params
//panic if param isn't a function
func FuncOut(v interface{}) []Type {
	typ := TypeOf(v)

	numOut := typ.NumOut()
	outs := make([]Type, numOut)
	for i := 0; i < numOut; i++ {
		outs[i] = typ.Out(i)
	}
	return outs
}

func FuncDesc(f interface{}) string {
	typ := TypeOf(f)
	if typ.Kind() != Func {
		return ""
	}
	return typ.String()
}
