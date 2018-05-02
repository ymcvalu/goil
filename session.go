package goil

//session entry is a session
type SessionEntry interface {
	Get(value interface{}) interface{}
	Set(key, value interface{}) error
	Delete(key interface{})
	Exists(key interface{}) bool
	Flush()
	SessionID() string
}

const (
	SESSION = "_SESSION_"
)

var sessionTag = SESSION

func SetSessionTag(tag string) {
	sessionTag = tag
}

func GetSessionTag() string {
	return sessionTag
}
