package goil

import "encoding/xml"

type Xml struct {
}

var xmlRender = new(Xml)

func (x *Xml) Render(w Response, content interface{}) error {
	byts, err := xml.Marshal(content)
	if err != nil {
		return err
	}
	_, err = w.Write(byts)
	return err
}

func (x *Xml) ContentType() string {
	return MIME_XML
}
