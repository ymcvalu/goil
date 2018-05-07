package goil

type SessionManager interface {
	SessionGet(string) Session
	SessionPut(Session)
	SessionExists(string) bool
	SessionDestroy(string)
	SessionCount() int
	SessionGC()
}

type Session interface {
	Get(value interface{}) interface{}
	Set(key, value interface{}) error
	Delete(key interface{})
	Exists(key interface{}) bool
	Flush()
	SessionID() string
}
