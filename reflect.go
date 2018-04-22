package goil

import (
	"reflect"
)

func isPtr(iface interface{}) bool {
	switch elem := iface.(type) {
	case reflect.Value:
		return elem.Type().Kind() == reflect.Ptr
	case reflect.Type:
		return elem.Kind() == reflect.Ptr
	default:
		k := reflect.TypeOf(elem).Kind()
		return k == reflect.Ptr
	}

}

func typeOf(iface interface{}) reflect.Type {
	return reflect.TypeOf(iface)
}

func valueOf(iface interface{}) reflect.Value {
	return reflect.ValueOf(iface)
}
