package session

import (
	"goil"
)

type Session = goil.SessionEntry

//c like?less *
type Void = interface{}

//kotlin like
type Any = interface{}

//java like
type Object = interface{}

type SessionCache interface {
	Get(sessionID string) Session
	Put(entry Session)
}

type SessionManager interface {
	SessionGet(string) Session
	SessionPut(Session)
	SessionExists(string) bool
	SessionDestroy(string)
	SessionCount() int
	SessionGC()
}
