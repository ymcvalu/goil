package session

import "goil"

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

type Session = goil.Session
type SessionManager = goil.SessionManager
