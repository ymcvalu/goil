package goil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
)

type Binding interface {
	Bind(c *Context, iface interface{}) error
}

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
			return reg.MatchString(*val), nil
		}
		return false, errors.New("validator reg only support string type")
	},
}

var NoValidatorExists = errors.New("no validator exists")

func RegisterValidator(name string, validator Validator) bool {
	if _, conflict := validateFunc[name]; conflict {
		return false
	}
	validateFunc[name] = validator
	return true
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
	eTyp := typeOf(iface)
	eVal := valueOf(iface)
	if eTyp.Kind() == reflect.Ptr {
		if eVal.IsNil() {
			return false, errors.New("nil pointer for validate")
		}
		eVal = eVal.Elem()
		eTyp = eVal.Type()
	}
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

func RegisterConvert(name string, fun Convert) bool {
	if _, conflict := convertFunc[name]; conflict {
		return false
	}
	convertFunc[name] = fun
	return true
}

func bindField(src string, dest reflect.Value, fTyp reflect.StructField) error {
	tag := fTyp.Tag
	conv := tag.Get(CONVERT)
	if convFunc, exists := convertFunc[conv]; exists {
		val, err := convFunc(src, dest.Type())
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
	case reflect.Ptr:
		elemType := dest.Type().Elem()
		switch elemType.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			conv = "_a2i"
		case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			conv = "_a2u"
		case reflect.Bool:
			conv = "_a2b"
		case reflect.Float32, reflect.Float64:
			conv = "_a2f"
		case reflect.String:
			conv = ""
		default:
			return fmt.Errorf("unsupport type for binding params %s to %s", src, dest)
		}
	case reflect.String:
		conv = ""
	default:
		return fmt.Errorf("unsupport type for binding params %s to %s", src, dest)
	}
	var val interface{}
	if conv != "" {
		v, err := convertFunc[conv](src, dest.Type())
		if err != nil {
			return fmt.Errorf("when binding params %v to %v:%s", src, dest, err)
		}
		val = v
	} else {
		val = src
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
		case reflect.Ptr:
			elemType := dest.Type().Elem()
			elemVal := reflect.New(elemType)
			dest.Set(elemVal)

			switch elemType.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
				elemVal.Elem().SetInt(val.(int64))
			case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				elemVal.Elem().SetUint(val.(uint64))
			case reflect.Bool:
				elemVal.Elem().SetBool(val.(bool))
			case reflect.Float32, reflect.Float64:
				elemVal.Elem().SetFloat(val.(float64))
			case reflect.String:
				elemVal.Elem().SetString(val.(string))
			}
		case reflect.String:
			dest.SetString(val.(string))
		}
	}
	return nil
}

type File struct {
	FileName string
	Size     int64
	File     os.File
}

func bindFile(fh *multipart.FileHeader, dest reflect.Value, fTyp reflect.StructField) error {
	if !dest.CanSet() {
		return nil
	}
	dType := fTyp.Type
	if dType.Kind() == reflect.Ptr {
		elemTyp := dType.Elem()
		elemVal := reflect.New(elemTyp)
		dest.Set(elemVal)
		dest = elemVal
	}
	switch d := dest.Interface().(type) {
	//pointer:assign directly
	case *int64:
		_size := fh.Size
		*d = _size
	case *os.File:
		fd, err := fh.Open()
		if err != nil {
			return fmt.Errorf("open upload file:%s", err)
		}
		*d = *fd.(*os.File)
	case *File:
		fd, err := fh.Open()
		if err != nil {
			return fmt.Errorf("open upload file:%s", err)
		}
		*d = File{
			FileName: fh.Filename,
			Size:     fh.Size,
			File:     *fd.(*os.File),
		}

	case *multipart.FileHeader:
		*d = *fh

	case *string:
		*d = fh.Filename

	//need to set dest
	case int64:
		dest.SetInt(fh.Size)

	case os.File:
		fd, err := fh.Open()
		if err != nil {
			return fmt.Errorf("open upload file:%s", err)
		}
		ptr := fd.(*os.File)
		dest.Set(valueOf(ptr).Elem())
	case File:
		fd, err := fh.Open()
		if err != nil {
			return fmt.Errorf("open upload file:%s", err)
		}
		ptr := &File{
			FileName: fh.Filename,
			Size:     fh.Size,
			File:     *fd.(*os.File),
		}
		dest.Set(valueOf(ptr).Elem())

	case multipart.File:
		fd, err := fh.Open()
		if err != nil {
			return fmt.Errorf("open upload file:%s", err)
		}
		dest.Set(valueOf(&fd).Elem())

	case multipart.FileHeader:
		dest.Set(valueOf(&fh).Elem())

	case string:
		dest.Set(valueOf(&fh.Filename).Elem())

	default:
		_ = d
		return fmt.Errorf("bind file:unsupport type")
	}

	return nil
}

const (
	VALIDATOR = "validator"
	CONVERT   = "convert"
	PATH      = "path"
	FORM      = "form"
	FILE      = "file"
)

func bindPathParams(ctx *Context, iface interface{}) (err error) {
	if ctx.params == nil || len(ctx.params) == 0 {
		return
	}

	if iface == nil {
		return errors.New("param is nil")
	}

	if !isPtr(iface) {
		return fmt.Errorf("param isn't a pointer")
	}

	val := valueOf(iface)
	typ := val.Type().Elem()

	if valueOf(iface).IsNil() {
		elemVal := reflect.New(typ)
		val.Set(elemVal)
		val = elemVal
	} else {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		for i, n := 0, typ.NumField(); i < n; i++ {
			fTyp := typ.Field(i)
			tag := fTyp.Tag
			pKey := tag.Get(PATH)
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

func bindFormParams(ctx *Context, iface interface{}) (err error) {
	if iface == nil {
		return errors.New("param is nil")
	}
	if !isPtr(iface) {
		return fmt.Errorf("param isn't a pointer")
	}
	req := ctx.Request
	err = req.ParseForm()
	if err != nil {
		return
	}
	contentType := req.Header.Get("Content-Type")
	if contentType != "" {
		d, _, err := mime.ParseMediaType(contentType)
		if d == "multipart/form-data" || err == nil {
			err = req.ParseMultipartForm(0)
			if err != nil && err != http.ErrNotMultipart {
				return err
			}
		}
	}

	val := valueOf(iface)
	typ := val.Type().Elem()
	if val.IsNil() {
		elemVal := reflect.New(typ)
		val.Set(elemVal)
		val = elemVal
	} else {
		val = val.Elem()
	}
	for i := 0; i < typ.NumField(); i++ {
		fVal := val.Field(i)
		if !fVal.CanSet() {
			continue
		}
		fTyp := typ.Field(i)
		tag := fTyp.Tag

		if key := tag.Get(FORM); key != "" {
			if pVal, exist := ctx.Request.Form[key]; exist && len(pVal) > 0 {
				err = bindField(pVal[0], fVal, fTyp)
				if err != nil {
					return
				}
			}
		} else if key = tag.Get(FILE); key != "" {
			if pVal, exist := ctx.Request.MultipartForm.File[key]; exist && len(pVal) > 0 {
				err = bindFile(pVal[0], fVal, fTyp)
				if err != nil {
					return
				}
			}
		}
	}

	return nil
}

func BindJson(ctx *Context, iface interface{}) (err error) {
	if !isPtr(iface) {
		return fmt.Errorf("params isn't a pointer")
	}
	_json, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(_json, iface)
	return
}

type ParamsHandler func(ctx *Context, iface interface{}) error

var paramsHandlers = map[string]ParamsHandler{
	MIME_JSON:      BindJson,
	MIME_POST:      bindFormParams,
	MIME_MULT_POST: bindFormParams,
}

func RegisterParamsHandler(tag string, handler ParamsHandler) bool {
	if _, conflict := paramsHandlers[tag]; conflict {
		return false
	}
	paramsHandlers[tag] = handler
	return true
}

const (
	MIME_TEXT      = "text/plain"
	MIME_JSON      = "application/json"
	MIME_POST      = "application/x-www-form-urlencoded"
	MIME_MULT_POST = "multipart/form-data"
)

func bind(c *Context, iface interface{}) error {
	if !isPtr(iface) {
		return fmt.Errorf("params isn't a pointer")
	}

	mt, _, err := mime.ParseMediaType(c.GetHeader().Get(CONTENT_TYPE))
	if err != nil {
		return err
	}

	if handler, exist := paramsHandlers[mt]; exist {
		err = handler(c, iface)
		return err
	}
	return UnsupportMimeType
}

var UnsupportMimeType = errors.New("unsupport mime-type")
