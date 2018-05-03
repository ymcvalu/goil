package goil

import "errors"

const (
	CONTENT_TYPE = "Content-Type"
)

//TODO:the prefix can config
const secure_json_prefix = "for(;;)"

var ParamsInvalidError = errors.New("params validate failed.")
