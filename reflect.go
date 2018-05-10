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

func isStruct(iface interface{}) bool {
	return typeOf(iface).Kind() == reflect.Struct
}

func isStructField(field reflect.StructField) bool {
	return field.Type.Kind() == reflect.Struct
}

func dereference(dv reflect.Value, dt reflect.Type) (reflect.Value, reflect.Type) {
	kind := dt.Kind()
	for kind == reflect.Ptr {
		if dv.IsNil() {
			elemTyp := dt.Elem()
			elemVal := reflect.New(elemTyp)
			dv.Set(elemVal)
			dv = elemVal.Elem()
			dt = elemTyp
			kind = elemTyp.Kind()
		} else {
			dv = dv.Elem()
			dt = dt.Elem()
			kind = dt.Kind()
		}
	}
	return dv, dt
}

func deref(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}
