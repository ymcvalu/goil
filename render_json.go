package goil

import "encoding/json"

type Json struct {
}

var JsonRender Render = new(Json)

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

type SecJson struct {
	prefix string
}

var SecJsonRender Render = &SecJson{
	prefix: JsonSecurePrefix,
}

func (j *SecJson) Render(w Response, content interface{}) error {
	_, err := w.Write([]byte(j.prefix))
	if err != nil {
		return err
	}
	byts, err := json.Marshal(content)
	if err != nil {
		return err
	}
	_, err = w.Write(byts)
	return err
}

func (j *SecJson) ContentType() string {
	return MIME_JSON
}
