package session

import (
	"goil"
	"strings"
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

var sessMgr goil.SessionManager
var once sync.Once = sync.Once{}

func initSession() {
	sessMgr = NewManagerMem()
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			sessMgr.SessionGC()
		}
	}()
}

func EnableSessionMem() goil.HandlerFunc {
	//init the session
	once.Do(initSession)
	return func(c *goil.Context) {
		sessionID := c.GetHeader().Get("goli_session_id")
		if sessionID == "" {
			sessionID = GenSessionID()
		}
		session := sessMgr.SessionGet(sessionID)
		c.Set(goil.SESSION, session)
		c.Next()
		defer func() {
			err := recover()
			sess := c.Session()
			if sess != nil {
				sess.Release()
			}
			if err != nil {
				panic(err)
			}
		}()
	}
}

func GenSessionID() string {
	sessionID := uuid.Must(uuid.NewV4()).String()
	return strings.Replace(sessionID, "-", "", -1)
}
