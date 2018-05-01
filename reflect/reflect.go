package reflect

import "reflect"

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
