package logger

import (
	. "goil/reflect"
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

var defLogger ILogger = NewAsync(os.Stdin, "", LstdFlags, DebugLevel, 100)

type ILogger interface {
	Printf(format string, msg ...interface{})
	Infof(format string, msg ...interface{})
	Debugf(format string, msg ...interface{})
	Warnf(format string, msg ...interface{})
	Errorf(format string, msg ...interface{})
	Panicf(format string, msg ...interface{})
	Fatalf(format string, msg ...interface{})
	Info(msg ...interface{})
	Print(msg ...interface{})
	Debug(msg ...interface{})
	Warn(msg ...interface{})
	Error(msg ...interface{})
	Panic(msg ...interface{})
	Fatal(msg ...interface{})
	IsTTY() bool
}

func SetLogger(l ILogger) {
	//if l is a nil pointer
	if !IsNilPtr(l) {
		defLogger = l
	}
}

func Printf(format string, msg ...interface{}) {
	defLogger.Printf(format, msg...)
}
func Infof(format string, msg ...interface{}) {
	defLogger.Infof(format, msg...)
}
func Debugf(format string, msg ...interface{}) {
	defLogger.Debugf(format, msg...)
}
func Warnf(format string, msg ...interface{}) {
	defLogger.Warnf(format, msg...)
}
func Errorf(format string, msg ...interface{}) {
	defLogger.Errorf(format, msg...)
}
func Panicf(format string, msg ...interface{}) {
	defLogger.Panicf(format, msg...)
}
func Fatalf(format string, msg ...interface{}) {
	defLogger.Fatalf(format, msg...)
}
func Info(msg ...interface{}) {
	defLogger.Info(msg...)
}
func Print(msg ...interface{}) {
	defLogger.Print(msg...)
}
func Debug(msg ...interface{}) {
	defLogger.Debug(msg...)
}
func Warn(msg ...interface{}) {
	defLogger.Warn(msg...)
}
func Error(msg ...interface{}) {
	defLogger.Error(msg...)
}
func Panic(msg ...interface{}) {
	defLogger.Panic(msg...)
}
func Fatal(msg ...interface{}) {
	defLogger.Fatal(msg...)
}
func IsTTY() bool {
	return defLogger.IsTTY()
}
