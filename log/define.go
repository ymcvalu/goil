package log

import (
	"os"
)

const (
	Ldate = 1 << iota
	Ltime
	Lmicroseconds
	Llongfile
	Lshortfile
	LUTC
	LstdFlags = Ldate | Ltime | Lshortfile
)

const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
	_max_
)

var DefLogger = NewAsync(os.Stdin, "", LstdFlags, DebugLevel, 100)
