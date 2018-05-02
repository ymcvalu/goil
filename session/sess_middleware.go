package session

import (
	"goil"
	"goil/util"
	"strings"
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

var sessMgr *ManagerMem
var once sync.Once = sync.Once{}

func initSession() {
	sessMgr = NewManagerMem()
	go func() {
		for {
			time.Sleep(util.MinToDuration(_GCDuration))
			sessMgr.SessionGC()
		}
	}()
}

func EnableSessionMem() goil.HandlerFunc {
	//init the session
	once.Do(initSession)
	return func(c *goil.Context) {
		sessionID := c.Headers().Get("goli_session_id")
		if sessionID == "" {
			sessionID = GenSessionID()
			c.Headers().Set("goli_session_id", sessionID)
			c.Response.SetHeader("goli_session_id", sessionID)
		}
		session := sessMgr.SessionGet(sessionID)
		c.Set(goil.GetSessionTag(), session)
		c.Next()
		defer func() {
			err := recover()
			session.release()
			c.Set(goil.GetSessionTag(), nil)
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
