package goil

import (
	"goil/logger"
	"html/template"
	"io/ioutil"
)

type FuncMap = map[string]interface{}

type HtmlTemp struct {
	T *template.Template
}

//type assert
var _ Render = new(HtmlTemp)

func (h *HtmlTemp) Render(w Response, content interface{}) error {
	vm := content.(ViewModel)
	return h.T.ExecuteTemplate(w, vm.Name, vm.Model)
}

func (h *HtmlTemp) ContentType() string {
	return MIME_HTML
}

func (h *HtmlTemp) Method(name string, m interface{}) *HtmlTemp {
	h.T.Funcs(FuncMap{name: m})
	return h
}

func (h *HtmlTemp) Methods(ms FuncMap) *HtmlTemp {
	h.T.Funcs(ms)
	return h
}

func (h *HtmlTemp) Delims(left, right string) *HtmlTemp {
	h.T.Delims(left, right)
	return h
}

func (h *HtmlTemp) Temp(name, filepath string) *HtmlTemp {
	byts, err := ioutil.ReadFile(filepath)
	if err != nil {
		logger.Errorf("when add html template named %s: %s", name, err)
		return h
	}
	_, err = h.T.New(name).Parse(string(byts))
	if err != nil {
		logger.Errorf("when add html template named %s: %s", name, err)
		return h
	}
	return h
}

func (h *HtmlTemp) Temps(filepath ...string) *HtmlTemp {
	_, err := h.T.ParseFiles(filepath...)
	if err != nil {
		logger.Errorf("when add html templates: %s", err)
	}
	return h
}

var HtmlRender *HtmlTemp

func init() {
	HtmlRender = &HtmlTemp{
		T: template.New(""),
	}
}

func TempMethod(name string, fun interface{}) {
	HtmlRender.Method(name, fun)
}
func TempMethods(ms FuncMap) {
	HtmlRender.Methods(ms)
}

func TempDelims(left, right string) {
	HtmlRender.Delims(left, right)
}

func Temp(name, filepath string) {
	HtmlRender.Temp(name, filepath)
}

func Temps(filepath ...string) {
	HtmlRender.Temps(filepath...)
}

type ViewModel struct {
	Name  string
	Model interface{}
}

func VM(name string, data interface{}) ViewModel {
	return ViewModel{
		Name:  name,
		Model: data,
	}
}
