package goil

//session manager is the manager of sessin
type SessionManager interface {
	SessionGet(sessionID string) SessionEntry
	SessionExists(sessionID string) bool
	SessionDestroy(sessionID string)
	SessionCount() int
	SessionGC()
}

//session entry is a session
type SessionEntry interface {
	Get(value interface{}) interface{}
	Set(key, value interface{}) error
	Delete(key interface{})
	Exist(key interface{}) bool
	SessionID() string
	Release()
}
