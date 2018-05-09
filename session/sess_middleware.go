package session

import (
	"goil"
	"strings"

	"github.com/satori/go.uuid"
)

func UseSession() goil.HandlerFunc {
	//using memory sessin default
	//can call EnableRdsSession before UseSession to use the redis session
	EnableMemSession()
	return func(c *goil.Context) {
		sessionID := c.Headers().Get(_ClientSidTag)
		if sessionID == "" {
			sessionID = GenSessionID()
			c.Headers().Set(_ClientSidTag, sessionID)
		}
		c.Response.SetHeader(_ClientSidTag, sessionID)
		session := sessMgr.SessionGet(sessionID)

		c.Next()
		defer func() {
			//intercept the panic
			err := recover()

			sessMgr.SessionPut(session)
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
