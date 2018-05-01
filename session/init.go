package session

import (
	"goil"
	"goil/log"
)

var logger goil.Logger = log.DefLogger

func SetLogger(l goil.Logger) {
	logger = l
}
