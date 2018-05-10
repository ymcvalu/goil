package goil

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"goil/logger"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
)

func bindValue(src string, dv reflect.Value, dt reflect.Type, tag reflect.StructTag) error {
	conv := tag.Get(CONVERT)
	if convFunc, exists := convertFunc[conv]; conv != "" && exists {
		if !dv.CanSet() {
			return nil
		}
		val, err := convFunc(src, dv.Type())
		if err != nil {
			return err
		}

		dv.Set(valueOf(val))

		return nil
	}
	var v interface{}
	var err error

	switch dt.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		conv = "_a2i"
		v, err = convertFunc[conv](src, dt)
		if err != nil {
			break
		}
		dv.SetInt(v.(int64))
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		conv = "_a2u"
		v, err = convertFunc[conv](src, dt)
		if err != nil {
			break
		}
		dv.SetUint(v.(uint64))
	case reflect.Bool:
		conv = "_a2b"
		v, err = convertFunc[conv](src, dt)
		if err != nil {
			break
		}
		dv.SetBool(v.(bool))
	case reflect.Float32, reflect.Float64:
		conv = "_a2f"
		v, err = convertFunc[conv](src, dt)
		if err != nil {
			break
		}
		dv.SetFloat(v.(float64))
	case reflect.String:
		dv.SetString(src)
	}
	return err
}

type File struct {
	FileName    string
	Size        int64
	File        multipart.File
	ContentType string
}

func bindFile(fh *multipart.FileHeader, dv reflect.Value, dt reflect.Type) error {
	if !dv.CanSet() {
		return nil
	}

	switch dv.Interface().(type) {
	//pointer:assign directly
	case *int64:
		_size := fh.Size

		dv.Set(valueOf(&_size))
	case *os.File:
		fd, err := fh.Open()
		if err != nil {
			return fmt.Errorf("open upload file:%s", err)
		}
		if file, ok := fd.(*os.File); ok {
			dv.Set(valueOf(file))
		} else {
			logger.Warnf("for reading the uploaded file: %s,please use the type %s replace %s", fh.Filename, "multipart.File", "*os.File")
		}

	case *File:
		fd, err := fh.Open()
		if err != nil {
			return fmt.Errorf("open upload file:%s", err)
		}

		file := &File{
			FileName:    fh.Filename,
			Size:        fh.Size,
			File:        fd,
			ContentType: fh.Header.Get(CONTENT_TYPE),
		}
		dv.Set(valueOf(file))

	case multipart.File:
		file, err := fh.Open()
		if err != nil {
			return fmt.Errorf("open upload file:%s", err)
		}
		if f, ok := file.(multipart.File); ok {
			dv.Set(valueOf(f))
		}
	case *multipart.FileHeader:
		dv.Set(valueOf(fh))

	case *string:
		dv.Set(valueOf(&fh.Filename))

	//need to set dest
	case int64:
		dv.SetInt(fh.Size)
	case os.File:
		fd, err := fh.Open()
		if err != nil {
			return fmt.Errorf("open upload file:%s", err)
		}
		if file, ok := fd.(*os.File); ok {
			dv.Set(valueOf(file).Elem())
		} else {
			logger.Warnf("for reading the uploaded file: %s,please use the type %s replace %s", fh.Filename, "multipart.File", "os.File")
		}

	case File:
		fd, err := fh.Open()
		if err != nil {
			return fmt.Errorf("open upload file:%s", err)
		}

		ptr := &File{
			FileName:    fh.Filename,
			Size:        fh.Size,
			File:        fd,
			ContentType: fh.Header.Get(CONTENT_TYPE),
		}
		dv.Set(valueOf(ptr).Elem())

	case string:
		dv.Set(valueOf(&fh.Filename).Elem())

	}

	return nil
}

func bindValues(src []string, dv reflect.Value, dt reflect.Type) error {
	elemType := dt.Elem()
	switch elemType.Kind() {
	case reflect.String:
		dv.Set(valueOf(src))
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

	// val := valueOf(iface)
	// typ := val.Type().Elem()
	// if valueOf(iface).IsNil() {
	// 	elemVal := reflect.New(typ)
	// 	val.Set(elemVal)
	// 	val = elemVal
	// } else {
	// 	val = val.Elem()
	// }
	val := valueOf(iface)
	typ := val.Type()
	val, typ = dereference(val, typ)
	for i, n := 0, typ.NumField(); i < n; i++ {
		fVal := val.Field(i)

		if !fVal.CanSet() {
			continue
		}
		fTyp := typ.Field(i)
		dv := fVal
		dt := fTyp.Type
		dv, dt = dereference(dv, dt)

		//To support the embeded struct
		if dt.Kind() == reflect.Struct {
			//need the pointer type interface
			bindPathParams(params, dv.Addr().Interface())
			continue
		}

		tag := fTyp.Tag
		pKey := tag.Get(PATH)
		pVal, exist := params.get(pKey)
		if !exist {
			continue
		}

		err = bindValue(pVal, dv, dt, fTyp.Tag)
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
	// val := valueOf(iface)
	// typ := val.Type().Elem()

	// if valueOf(iface).IsNil() {
	// 	elemVal := reflect.New(typ)
	// 	val.Set(elemVal)
	// 	val = elemVal
	// } else {
	// 	val = val.Elem()
	// }
	val := valueOf(iface)
	typ := val.Type()
	val, typ = dereference(val, typ)
	for i, n := 0, typ.NumField(); i < n; i++ {
		fVal := val.Field(i)
		if !fVal.CanSet() {
			continue
		}
		fTyp := typ.Field(i)
		dv := fVal
		dt := fTyp.Type
		dv, dt = dereference(dv, dt)

		if dt.Kind() == reflect.Struct {
			bindQueryParams(request, dv.Addr().Interface())
			continue
		}
		tag := fTyp.Tag
		pKey := tag.Get(FORM)
		pVal, exist := values[pKey]
		if !exist || len(pVal) == 0 {
			continue
		}

		if dt.Kind() == reflect.Slice {
			err = bindValues(pVal, dv, dt)
		} else {
			err = bindValue(pVal[0], dv, dt, fTyp.Tag)
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
			req.ParseMultipartForm(DEFAULT_SIZE)

		}
	}

	// val := valueOf(iface)
	// typ := val.Type().Elem()
	// if val.IsNil() {
	// 	elemVal := reflect.New(typ)
	// 	val.Set(elemVal)
	// 	val = elemVal
	// } else {
	// 	val = val.Elem()
	// }
	val := valueOf(iface)
	typ := typeOf(iface)
	val, typ = dereference(val, typ)
	for i := 0; i < typ.NumField(); i++ {
		fVal := val.Field(i)
		if !fVal.CanSet() {
			continue
		}
		fTyp := typ.Field(i)

		dv := fVal
		dt := fTyp.Type
		tag := fTyp.Tag
		fileKey := tag.Get(FILE)

		//read the file firstly
		if fileKey != "" {
			if pVal, exist := req.MultipartForm.File[fileKey]; exist && len(pVal) > 0 {
				err = bindFile(pVal[0], dv, dt)
				if err != nil {
					return
				}
			}
			continue
		}

		dv, dt = dereference(dv, dt)

		if dt.Kind() == reflect.Struct {
			//support the nested struct
			bindFormParams(req, dv.Addr().Interface())
			continue
		}

		formKey := tag.Get(FORM)
		if formKey == "" {
			formKey = genKey(fTyp.Name)
		}
		pVal, exist := req.PostForm[formKey]
		if !exist || len(pVal) == 0 {
			continue
		}
		if dt.Kind() == reflect.Slice {
			err = bindValues(pVal, dv, dt)
		} else {

			err = bindValue(pVal[0], dv, dt, fTyp.Tag)
		}
		if err != nil {
			return
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

func bindXml(req *http.Request, iface interface{}) (err error) {
	_xml, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}
	err = xml.Unmarshal(_xml, iface)
	return
}

//change the params to request
type ParamsBinder func(req *http.Request, iface interface{}) error

var paramsHandlers = map[string]ParamsBinder{
	MIME_JSON:      bindJSON,
	MIME_POST:      bindFormParams,
	MIME_MULT_POST: bindFormParams,
	MIME_XML:       bindXml,
}

func RegisterParamsHandler(tag string, handler ParamsBinder) {
	guard.execSafely(func() {
		paramsHandlers[tag] = handler
	})
}

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
	//3.bind body params if existing
	contentType := c.Headers().Get(CONTENT_TYPE)
	if contentType == "" {
		return nil
	}
	mt, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return err
	}
	//3.bind body params
	if handler, exist := paramsHandlers[mt]; exist {
		err = handler(c.Request, iface)
		return err
	}
	return UnsupportMimeType
}

var UnsupportMimeType = errors.New("unsupport content type for params binding")
