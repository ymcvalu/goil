package goil

import (
	"io"
)

type Render interface {
	Render(content interface{}) io.Reader
	ContentType() string
}
