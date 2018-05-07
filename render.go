package goil

type Render interface {
	Render(w Response, content interface{}) error
	ContentType() string
}
