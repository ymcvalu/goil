package session

import (
	"goil"
	"goil/log"
)

var logger goil.Logger = log.DefLogger
var _ExpireDuration int64 = 15 * 60 //s
var _GCDuration int64 = 5 * 60      //s

func SetLogger(l goil.Logger) {
	logger = l
}

func SetExpireDuration(d int64) {
	_ExpireDuration = d
}

func SetGCDuration(d int64) {
	_GCDuration = d
}

func GetExpireDuration() int64 {
	return _ExpireDuration
}

func GetGCDuration() int64 {
	return _GCDuration
}
