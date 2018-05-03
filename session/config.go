package session

import (
	"goil"
	"goil/log"
	"goil/util"
	"sync"
	"time"
)

var logger goil.Logger = log.DefLogger
var _ExpireDuration int64 = 15 * 60 //s
var _GCDuration int64 = 5 * 60      //s
var _RedisKeyPrefix = "sess_"
var _RedisAddr string = "redis:127.0.0.1:6379"
var _ClientSidTag string = "goil_sid"
var sessMgr SessionManager

//for init only once
var once sync.Once = sync.Once{}

func SetRedisKeyPrefix(prefix string) {
	_RedisKeyPrefix = prefix
}

func SetRedisAddr(addr string) {
	_RedisAddr = addr
}

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

func SetSidTag(tag string) {
	_ClientSidTag = tag
}

func EnableMemSession() bool {
	if sessMgr == SessionManager(nil) {
		once.Do(func() {
			sessMgr = NewManagerMem()
			go func() {
				for {
					time.Sleep(util.MinToDuration(_GCDuration))
					sessMgr.SessionGC()
				}
			}()
		})
		return true
	}
	return false
}

func EnableRdsSession() bool {
	if sessMgr == SessionManager(nil) {
		once.Do(func() {
			sessMgr = NewManagerRedis()
		})
		return true
	}
	return false
}
