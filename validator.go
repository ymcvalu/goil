package goil

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

type ValidatedField struct {
	Value  reflect.Value
	Type   reflect.Type
	params []string
}

func (f *ValidatedField) ParamsNum() int {
	return len(f.params)
}

func (f *ValidatedField) Params() []string {
	return f.params
}

func (f *ValidatedField) String(i int) (string, error) {
	if i < 0 || i >= f.ParamsNum() {
		return "", errors.New("outer of range")
	}
	return f.params[i], nil
}

func (f *ValidatedField) Float(i int) (float64, error) {
	if i < 0 || i >= f.ParamsNum() {
		return 0, errors.New("outer of range")
	}
	return strconv.ParseFloat(f.params[i], 64)
}

func (f *ValidatedField) Int(i int) (int64, error) {
	if i < 0 || i >= f.ParamsNum() {
		return 0, errors.New("outer of range")
	}
	return strconv.ParseInt(f.params[i], 10, 64)
}

func (f *ValidatedField) Uint(i int) (uint64, error) {
	if i < 0 || i >= f.ParamsNum() {
		return 0, errors.New("outer of range")
	}
	return strconv.ParseUint(f.params[i], 10, 64)
}

func (f *ValidatedField) Bool(i int) (bool, error) {
	if i < 0 || i >= f.ParamsNum() {
		return false, errors.New("outer of range")
	}
	return strconv.ParseBool(f.params[i])
}

type ＶalidateFunc = func(f ValidatedField) bool

var validateFunc = map[string]ＶalidateFunc{
	"required": func(f ValidatedField) bool {
		val := f.Value
		switch val.Kind() {
		case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
			return !val.IsNil()
		default:
			return val.IsValid() && val.Interface() != reflect.Zero(val.Type()).Interface()
		}
	},
	"min": func(f ValidatedField) bool {
		if f.ParamsNum() < 0 {
			return false
		}
		min, err := f.Float(0)
		if err != nil {
			return false
		}
		value := f.Value
		switch value.Interface().(type) {
		case int, int64, int32, int8:
			goto iv
		case float32, float64:
			goto fv
		case *int, *int64, *int32, *int8:
			if value.IsNil() {
				return false
			}
			value = value.Elem()
			goto iv
		case *float32, *float64:
			if value.IsNil() {
				return false
			}
			value = value.Elem()
			goto fv
		}
		return false
	iv:
		{
			intVal := value.Int()
			if intVal < int64(min) {
				return false
			}
			return true
		}

	fv:
		{
			fltVal := value.Float()
			if fltVal < min {
				return false
			}
			return true
		}

	},
	"max": func(f ValidatedField) bool {
		if f.ParamsNum() < 1 {
			return false
		}
		max, err := f.Float(0)
		if err != nil {
			return false
		}
		value := f.Value
		switch value.Interface().(type) {
		case int, int64, int32, int8:
			goto iv
		case float32, float64:
			goto fv
		case *int, *int64, *int32, *int8:
			if value.IsNil() {
				return false
			}
			value = value.Elem()
			goto iv
		case *float32, *float64:
			if value.IsNil() {
				return false
			}
			value = value.Elem()
			goto fv
		}
		return false
	iv:
		{
			intVal := value.Int()
			if intVal > int64(max) {
				return false
			}
			return true
		}

	fv:
		{
			fltVal := value.Float()
			if fltVal > max {
				return false
			}
			return true
		}

	},
	"range": func(f ValidatedField) bool {
		if f.ParamsNum() < 2 {
			return false
		}
		min, err := f.Float(0)
		if err != nil {
			return false
		}
		max, err := f.Float(1)
		if err != nil {
			return false
		}
		value := f.Value
		switch value.Interface().(type) {
		case int, int64, int32, int8:
			goto iv
		case float32, float64:
			goto fv
		case *int, *int64, *int32, *int8:
			if value.IsNil() {
				return false
			}
			value = value.Elem()
			goto iv
		case *float32, *float64:
			if value.IsNil() {
				return false
			}
			value = value.Elem()
			goto fv
		}
		return false
	iv:
		{
			intVal := value.Int()
			if intVal > int64(max) || intVal < int64(min) {
				return false
			}
			return true
		}

	fv:
		{
			fltVal := value.Float()
			if fltVal > max || fltVal < min {
				return false
			}
			return true
		}

	},
	"reg": func(f ValidatedField) bool {
		if f.ParamsNum() < 1 {
			return false
		}
		exp, err := f.String(0)
		if err != nil {
			return false
		}
		reg, err := regexp.Compile(exp)
		if err != nil {
			return false
		}
		switch val := f.Value.Interface().(type) {
		case string:
			return reg.MatchString(val)
		case *string:
			if val != nil {
				return reg.MatchString(*val)
			}
			return reg.MatchString("")
		}
		return false
	},
	"enum": func(f ValidatedField) bool {
		dv, dt := f.Value, f.Type
		kind := dt.Kind()
		switch kind {
		case reflect.Ptr:
			for dt.Kind() == reflect.Ptr {
				if dv.IsNil() {
					return false
				}
				dv, dt = dv.Elem(), dt.Elem()
			}
			if dt.Kind() != reflect.String {
				return false
			}
			fallthrough
		case reflect.String:
			val := dv.String()
			params := f.Params()
			for i := range params {
				if val == params[i] {
					return true
				}
			}
		default:
			return false
		}

		return false
	},
}

func RegisterValidator(name string, fun ＶalidateFunc) {
	guard.execSafely(func() {
		validateFunc[name] = fun
	})
}

func validateField(tag string, val reflect.Value, rTyp reflect.StructField) error {
	keys, params, err := parseTag(tag)
	if err != nil {
		panic(err)
	}
	for i := range keys {
		if validator, exists := validateFunc[keys[i]]; exists {
			f := ValidatedField{
				Value:  val,
				Type:   rTyp.Type,
				params: params[i],
			}
			ok := validator(f)
			if !ok {
				name := rTyp.Name
				return fmt.Errorf("failed to validate %s for %s", name, keys[i])
			}
		} else {
			panic(fmt.Errorf("no validator exists for %s", keys[i]))
		}
	}
	return nil
}

func validate(iface interface{}) error {
	if iface == nil {
		return errors.New("nil pointer for validate")
	}

	eVal := valueOf(iface)
	eTyp := eVal.Type()
	eVal, eTyp = dereference(eVal, eTyp)

	for i := 0; i < eTyp.NumField(); i++ {
		fTyp := eTyp.Field(i)
		if !export(fTyp.Name) {
			continue
		}
		fVal := eVal.Field(i)

		tag := fTyp.Tag.Get(VALIDATOR)

		if tag == "" {
			if IsStructReally(fTyp.Type) {
				err := validate(fVal.Interface())
				if err != nil {
					return err
				}
			}
			continue
		}
		err := validateField(tag, fVal, fTyp)
		if err != nil {
			return err
		}
	}
	return nil
}
