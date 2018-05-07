package goil

import (
	"html/template"
	"io/ioutil"
	"reflect"
)

type funcMap = map[string]interface{}

type htmls struct {
	root    *template.Template
	methods funcMap
}

type ViewModel struct {
	Name  string
	Model interface{}
}

func (h *htmls) Render(w Response, content interface{}) error {
	vm := content.(*ViewModel)
	return h.root.ExecuteTemplate(w, vm.Name, vm.Model)
}

func (t *htmls) ContentType() string {
	return MIME_HTML
}

var htmlRender *htmls

func init() {
	htmlRender = &htmls{
		root:    template.New(""),
		methods: funcMap{},
	}
	htmlRender.root.Funcs(htmlRender.methods)
}

func MethodForHtmlTmpl(name string, fun interface{}) {
	assert1(typeOf(fun).Kind() == reflect.Func, "assert failed:kind of fun isn't reflect.Func")
	guard.execSafely(func() {
		htmlRender.methods[name] = fun
	})
}

func DelimsForHtmlTmpl(left, right string) {
	guard.execSafely(func() {
		htmlRender.root.Delims(left, right)
	})
}

func HtmlTemp(name, filepath string) {
	guard.execSafely(func() {
		byts, err := ioutil.ReadFile(filepath)
		if err != nil {
			logger.Errorf("when add html template named %s: %s", name, err)
			return
		}
		_, err = htmlRender.root.New(name).Parse(string(byts))
		if err != nil {
			logger.Errorf("when add html template named %s: %s", name, err)
			return
		}
	})
}
