package goil

import (
	"compress/gzip"
	"errors"
)

const (
	CONTENT_TYPE     = "Content-Type"
	CONTENT_ENCODING = "Content-Encoding"
	ACCEPT           = "Accept"
)

//TODO:the prefix can config
const JsonSecurePrefix = "for(;;)"

var ParamsInvalidError = errors.New("params validate failed.")

const (
	MIME_TEXT      = "text/plain"
	MIME_JSON      = "application/json"
	MIME_POST      = "application/x-www-form-urlencoded"
	MIME_MULT_POST = "multipart/form-data"
	MIME_CSS       = "text/css"
	MIME_GIF       = "image/gif"
	MIME_HTML      = "text/html"
	MIME_JPEG      = "image/jpeg"
	MIME_JS        = "application/x-javascript"
	MIME_PDF       = "application/pdf"
	MIME_PNG       = "image/png"
	MIME_SVG       = "image/svg+xml"
	MIME_XML       = "text/xml"
)

var (
	greenBkg   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	whiteBkg   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellowBkg  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	redBkg     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blueBkg    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magentaBkg = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyanBkg    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	redFont    = string([]byte{27, 91, 57, 55, 59, 51, 49, 109})
	resetClr   = string([]byte{27, 91, 48, 109})
)

const (
	GZIP_NoCompression      = gzip.NoCompression
	GZIP_BestSpeed          = gzip.BestSpeed
	GZIP_BestCompression    = gzip.BestCompression
	GZIP_DefaultCompression = gzip.DefaultCompression
	GZIP_HuffmanOnly        = gzip.HuffmanOnly
)

var (
	NoHandlers = errors.New("no handlers.")
)

const (
	Code_NoHandlers = -100
)
