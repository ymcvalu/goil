package goil

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
)

type Validator func(value reflect.Value, fTyp reflect.Type, params []string) (bool, error)

var validateFunc = map[string]Validator{
	"required": func(value reflect.Value, fTyp reflect.Type, params []string) (bool, error) {
		if !value.IsValid() {
			return false, nil
		}
		return true, nil
	},
	"min": func(value reflect.Value, fTyp reflect.Type, params []string) (bool, error) {
		if params == nil || len(params) < 1 {
			return false, errors.New("params error for min validator")
		}
		min, err := strconv.ParseFloat(params[0], 64)
		if err != nil {
			return false, err
		}

		switch value.Interface().(type) {
		case int, int64, int32, int8:
			goto iv
		case float32, float64:
			goto fv
		case *int, *int64, *int32, *int8:
			if value.IsNil() {
				return false, nil
			}
			value = value.Elem()
			goto iv
		case *float32, *float64:
			if value.IsNil() {
				return false, nil
			}
			value = value.Elem()
			goto fv
		}
		return false, nil
	iv:
		{
			intVal := value.Int()
			if intVal < int64(min) {
				return false, nil
			}
			return true, nil
		}

	fv:
		{
			fltVal := value.Float()
			if fltVal < min {
				return false, nil
			}
			return true, nil
		}

	},
	"max": func(value reflect.Value, fTyp reflect.Type, params []string) (bool, error) {
		if params == nil || len(params) < 1 {
			return false, errors.New("params error for min validator")
		}
		max, err := strconv.ParseFloat(params[0], 64)
		if err != nil {
			return false, err
		}

		switch value.Interface().(type) {
		case int, int64, int32, int8:
			goto iv
		case float32, float64:
			goto fv
		case *int, *int64, *int32, *int8:
			if value.IsNil() {
				return false, nil
			}
			value = value.Elem()
			goto iv
		case *float32, *float64:
			if value.IsNil() {
				return false, nil
			}
			value = value.Elem()
			goto fv
		}
		return false, nil
	iv:
		{
			intVal := value.Int()
			if intVal > int64(max) {
				return false, nil
			}
			return true, nil
		}

	fv:
		{
			fltVal := value.Float()
			if fltVal > max {
				return false, nil
			}
			return true, nil
		}

	},
	"range": func(value reflect.Value, fTyp reflect.Type, params []string) (bool, error) {

		if params == nil || len(params) < 2 {
			return false, errors.New("params error for min validator")
		}
		min, err := strconv.ParseFloat(params[0], 64)
		if err != nil {
			return false, err
		}
		max, err := strconv.ParseFloat(params[1], 64)
		if err != nil {
			return false, err
		}

		switch value.Interface().(type) {
		case int, int64, int32, int8:
			goto iv
		case float32, float64:
			goto fv
		case *int, *int64, *int32, *int8:
			if value.IsNil() {
				return false, nil
			}
			value = value.Elem()
			goto iv
		case *float32, *float64:
			if value.IsNil() {
				return false, nil
			}
			value = value.Elem()
			goto fv
		}
		return false, nil
	iv:
		{
			intVal := value.Int()
			if intVal > int64(max) || intVal < int64(min) {
				return false, nil
			}
			return true, nil
		}

	fv:
		{
			fltVal := value.Float()
			if fltVal > max || fltVal < min {
				return false, nil
			}
			return true, nil
		}

	},
	"reg": func(value reflect.Value, fTyp reflect.Type, params []string) (bool, error) {
		if params == nil || len(params) < 1 {
			return false, errors.New("params error for reg validator")
		}
		reg, err := regexp.Compile(params[0])
		if err != nil {
			return false, err
		}
		switch val := value.Interface().(type) {
		case string:
			return reg.MatchString(val), nil
		case *string:
			if val != nil {
				return reg.MatchString(*val), nil
			}
			return false, errors.New("the validating field is nil")
		}
		return false, errors.New("validator reg only support string type")
	},
}

var NoValidatorExists = errors.New("no validator exists")

func RegisterValidator(name string, validator Validator) {
	guard.execSafely(func() {
		validateFunc[name] = validator
	})
}

func validateField(tag string, val reflect.Value, rTyp reflect.StructField) (bool, error) {
	keys, params, err := parseTag(tag)
	if err != nil {
		return false, err
	}
	for i := range keys {

		if validator, e := validateFunc[keys[i]]; e {
			ok, err := validator(val, rTyp.Type, params[i])
			if !ok || err != nil {
				return ok, err
			}
		} else {
			return false, NoValidatorExists
		}
	}
	return true, nil
}

func validate(iface interface{}) (bool, error) {
	if iface == nil {
		return false, errors.New("nil pointer for validate")
	}

	eVal := valueOf(iface)
	eTyp := eVal.Type()
	eVal, eTyp = dereference(eVal, eTyp)

	for i := 0; i < eTyp.NumField(); i++ {
		fTyp := eTyp.Field(i)
		tag := fTyp.Tag.Get(VALIDATOR)

		if tag == "" {
			continue
		}
		fVal := eVal.Field(i)
		legal, err := validateField(tag, fVal, fTyp)
		if !legal || err != nil {
			return legal, err
		}
	}
	return true, nil
}

func parseTag(tag string) ([]string, [][]string, error) {
	var names = make([]string, 0, 1)
	var paramses = make([][]string, 0, 1)
	illegal := false
	bg, bd := 0, len(tag)

	for {
		//trim the prefix space
		for bg < bd && tag[bg] == ' ' {
			bg++
		}
		i := bg
		for i < bd {
			if tag[i] == '(' {
				illegal = true
				break
			}
			if tag[i] == ' ' {
				break
			}
			i++
		}
		j := i + 1
		for j < bd && illegal {
			if tag[j] == ')' {
				illegal = false
				break
			}
			j++
		}

		if illegal {
			return nil, nil, errors.New("the tag of validator less ')'")
		}
		//remove the suffix space
		nd := i - 1
		for i > bg && tag[nd] == ' ' {
			nd--
		}
		if nd > bg {
			names = append(names, tag[bg:nd+1])
		}
		i++
		if i >= j {
			if nd > bg {
				paramses = append(paramses, nil)
			}
			bg = j + 1
			if bg >= bd {
				break
			}
			continue
		}

		params := make([]string, 0, 1)
		buf := make([]byte, len(tag[i:j]))

		for i < j {
			//trim the space
			if tag[i] == ' ' {
				i++
				continue
			}
			buf[0] = tag[i]
			i++
			p := 1
			for i < j {
				if tag[i] != ' ' {
					buf[p] = tag[i]
					p++
					i++
					continue
				}
				break
			}
			params = append(params, string(buf[:p]))
		}

		paramses = append(paramses, params)
		bg = j + 1
		if bg >= bd {
			break
		}
	}
	return names, paramses, nil
}
