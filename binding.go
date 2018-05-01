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
)

func bindString(src string, dest reflect.Value, fTyp reflect.StructField) error {
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
	var v interface{}
	var err error
outer:
	switch dest.Type().Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		conv = "_a2i"
		v, err = convertFunc[conv](src, dest.Type())
		if err != nil {
			break outer
		}
		dest.SetInt(v.(int64))
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		conv = "_a2u"
		v, err = convertFunc[conv](src, dest.Type())
		if err != nil {
			break outer
		}
		dest.SetUint(v.(uint64))
	case reflect.Bool:
		conv = "_a2b"
		v, err = convertFunc[conv](src, dest.Type())
		if err != nil {
			break outer
		}
		dest.SetBool(v.(bool))
	case reflect.Float32, reflect.Float64:
		conv = "_a2f"
		v, err = convertFunc[conv](src, dest.Type())
		if err != nil {
			break outer
		}
		dest.SetFloat(v.(float64))
	case reflect.String:
		dest.SetString(src)
	case reflect.Ptr:
		elemType := dest.Type().Elem()
		elemVal := reflect.New(elemType)
		dest.Set(elemVal)
		elemVal = elemVal.Elem()
		switch elemType.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			conv = "_a2i"
			v, err = convertFunc[conv](src, dest.Type())
			if err != nil {
				break outer
			}
			elemVal.SetInt(v.(int64))
		case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			conv = "_a2u"
			v, err = convertFunc[conv](src, dest.Type())
			if err != nil {
				break outer
			}
			elemVal.SetUint(v.(uint64))
		case reflect.Bool:
			conv = "_a2b"
			v, err = convertFunc[conv](src, dest.Type())
			if err != nil {
				break outer
			}
			elemVal.SetBool(v.(bool))
		case reflect.Float32, reflect.Float64:
			conv = "_a2f"
			v, err = convertFunc[conv](src, dest.Type())
			if err != nil {
				break outer
			}
			elemVal.SetFloat(v.(float64))
		case reflect.String:
			elemVal.SetString(src)
			// default:
			// 	return fmt.Errorf("unsupport type for binding params %s to %s", src, fTyp.Name)
		}
		// default:
		// 	return fmt.Errorf("unsupport type for binding params %s to %s", src, fTyp.Name)
	}
	return err

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

		dest.Set(valueOf(fd.(*os.File)).Elem())
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

		//default:
		// return fmt.Errorf("bind file:unsupport type")
	}

	return nil
}

func bindSlice(src []string, dest reflect.Value, fTyp reflect.StructField) error {
	elemType := fTyp.Type.Elem()
	switch elemType.Kind() {
	case reflect.String:
		dest.Set(valueOf(src))
		// default:
		// 	return fmt.Errorf("unsupport type for binding params %s to %s", src, fTyp.Name)
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

func bindPathParams(params Params, iface interface{}) (err error) {
	if len(params) == 0 {
		return
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

	for i, n := 0, typ.NumField(); i < n; i++ {
		fVal := val.Field(i)

		if !fVal.CanSet() {
			continue
		}
		fTyp := typ.Field(i)

		//To support the embeded struct
		if fTyp.Type.Kind() == reflect.Struct {
			//need the pointer type interface
			bindPathParams(params, fVal.Addr().Interface())
			continue
		}

		tag := fTyp.Tag
		pKey := tag.Get(PATH)
		pVal, exist := params[pKey]
		if !exist {
			continue
		}

		err = bindString(pVal, fVal, fTyp)
		if err != nil {
			return
		}
	}

	return
}

func bindQueryParams(request *http.Request, iface interface{}) (err error) {
	values := request.URL.Query()
	if len(values) == 0 {
		return nil
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

	for i, n := 0, typ.NumField(); i < n; i++ {
		fVal := val.Field(i)
		if !fVal.CanSet() {
			continue
		}
		fTyp := typ.Field(i)
		if fTyp.Type.Kind() == reflect.Struct {
			bindQueryParams(request, fVal.Addr().Interface())
			continue
		}
		tag := fTyp.Tag
		pKey := tag.Get(FORM)
		pVal, exist := values[pKey]
		if !exist || len(pVal) == 0 {
			continue
		}

		if fTyp.Type.Kind() == reflect.Slice {
			err = bindSlice(pVal, fVal, fTyp)
		} else {
			err = bindString(pVal[0], fVal, fTyp)
		}
		if err != nil {
			return
		}
	}
	return
}

const DEFAULT_SIZE = 32 * 1024 * 1024

func bindFormParams(req *http.Request, iface interface{}) (err error) {
	err = req.ParseForm()
	if err != nil {
		return
	}
	contentType := req.Header.Get("Content-Type")
	if contentType != "" {
		ct, _, err := mime.ParseMediaType(contentType)
		if ct == MIME_MULT_POST && err == nil {
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
		if fTyp.Type.Kind() == reflect.Struct {
			bindFormParams(req, fVal.Addr().Interface())
			continue
		}

		tag := fTyp.Tag
		if key := tag.Get(FORM); key != "" {
			pVal, exist := req.PostForm[key]
			if exist && len(pVal) == 0 {
				continue
			}
			if fTyp.Type.Kind() == reflect.Slice {
				err = bindSlice(pVal, fVal, fTyp)
			} else {

				err = bindString(pVal[0], fVal, fTyp)
			}
			if err != nil {
				return
			}
		} else if key = tag.Get(FILE); key != "" {
			if pVal, exist := req.MultipartForm.File[key]; exist && len(pVal) > 0 {
				err = bindFile(pVal[0], fVal, fTyp)
				if err != nil {
					return
				}
			}
		}
	}

	return nil
}

func bindJSON(req *http.Request, iface interface{}) (err error) {

	_json, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(_json, iface)
	return
}

//change the params to request
type ParamsHandler func(req *http.Request, iface interface{}) error

var paramsHandlers = map[string]ParamsHandler{
	MIME_JSON:      bindJSON,
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

func bind(c *Context, iface interface{}) (err error) {
	if iface == nil {
		return errors.New("param is nil")
	}

	if !isPtr(iface) {
		return fmt.Errorf("param isn't a pointer")
	}

	//1.bind path params
	err = bindPathParams(c.params, iface)
	if err != nil {
		return err
	}

	//2.bind url params
	err = bindQueryParams(c.Request, iface)
	if err != nil {
		return err
	}

	mt, _, err := mime.ParseMediaType(c.GetHeader().Get(CONTENT_TYPE))
	if err != nil {
		return err
	}

	if handler, exist := paramsHandlers[mt]; exist {
		//2.the params
		err = handler(c.Request, iface)
		return err
	}
	return UnsupportMimeType
}

var UnsupportMimeType = errors.New("unsupport content type for params binding")
