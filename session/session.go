package session

import (
	"goil"
	"goil/helper"
	"goil/logger"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

type SessionStore interface {
	SessionRead(string) Session
	SessionExists(string) bool
	SessionDestroy(string)
	SessionCount() int64
	SessionGC()
}

type Session interface {
	Get(value interface{}) interface{}
	Set(key, value interface{})
	Delete(key interface{})
	Exists(key interface{}) bool
	Flush()
	SessionID() string
}

var once sync.Once
var sessStore SessionStore

func EnableMemSession() {
	sessStore = NewStoreMem()
	SetSessStore(sessStore)
}
func EnableRdsSession() {
	sessStore = NewStoreRedis()
	SetSessStore(sessStore)
}

func SetSessStore(sess SessionStore) {
	once.Do(func() {
		sessStore = sess
		time.AfterFunc(helper.SecToDuration(GCDuration), sess.SessionGC)
	})
}

func Sid(c *goil.Context) string {
	sid := c.Header(ClientSidTag)
	return sid
}

func GenSid(c *goil.Context) string {
	sid := Sid(c)
	if sid == "" {
		uid, err := uuid.NewV1()
		if err != nil {
			logger.Panic(err)
		}
		sid = uid.String()
		c.Request.Header.Add(ClientSidTag, sid)
	}
	c.SetHeader(ClientSidTag, sid)
	return sid
}

func SessionRead(c *goil.Context) Session {
	sid := GenSid(c)
	sess := sessStore.SessionRead(sid)
	return sess
}

func SessionExists(c *goil.Context) bool {
	sid := Sid(c)
	return sessStore.SessionExists(sid)
}

func SessionDestroy(c *goil.Context) {
	sid := Sid(c)
	sessStore.SessionDestroy(sid)
}
func SessionCount() int64 {
	return sessStore.SessionCount()
}
