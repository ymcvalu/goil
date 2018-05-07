package goil

import "encoding/json"

type Json struct {
}

var jsonRender = new(Json)

func (j *Json) Render(w Response, content interface{}) error {
	byts, err := json.Marshal(content)
	if err != nil {
		return err
	}
	_, err = w.Write(byts)
	return err
}

func (j *Json) ContentType() string {
	return MIME_JSON
}

type SecJsonRender struct {
	prefix string
}

var secJsonRender = &SecJsonRender{
	prefix: JsonSecurePrefix,
}

func (j *SecJsonRender) Render(w Response, content interface{}) error {
	_, err := w.Write([]byte(j.prefix))
	if err != nil {
		return err
	}
	return jsonRender.Render(w, content)
}

func (j *SecJsonRender) ContentType() string {
	return MIME_JSON
}
