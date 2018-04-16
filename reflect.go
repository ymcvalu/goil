package goil

import (
	"reflect"
)

func isPtr(iface interface{}) bool {
	k := reflect.TypeOf(iface).Kind()
	return k == reflect.Ptr
}

func typeOf(iface interface{}) reflect.Type {
	return reflect.TypeOf(iface)
}

func valueOf(iface interface{}) reflect.Value {
	return reflect.ValueOf(iface)
}
