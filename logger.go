package goil

import (
	"goil/log"
)

type Logger interface {
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

var logger Logger = log.DefLogger

func SetLogger(l Logger) {
	guard.execSafely(func() {
		if l != nil {
			logger = l
		}
	})
}
