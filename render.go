package goil

type Render interface {
	Render(ctx Context, contentType string) error
}
