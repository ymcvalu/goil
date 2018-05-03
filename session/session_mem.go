package session

import (
	"errors"
	"goil"
	. "goil/reflect"
	"sync"
	"time"
)

type SessionMem struct {
	mux       sync.RWMutex
	released  int64
	holder    uint32
	values    map[Any]Any
	sessionID string
}

var _ goil.SessionEntry = new(SessionMem)

func (s *SessionMem) Get(key Any) Any {
	s.mux.RLock()
	value := s.values[key]
	s.mux.RUnlock()
	return value
}
func (s *SessionMem) Set(key, value Any) error {
	if !CanComp(key) {
		return errors.New("the type of key unsupports compare")
	}
	s.mux.Lock()
	s.values[key] = value
	s.mux.Unlock()
	return nil
}
func (s *SessionMem) Delete(key Any) {
	if !CanComp(key) {
		return
	}
	s.mux.Lock()
	delete(s.values, key)
	s.mux.Unlock()
}
func (s *SessionMem) Exists(key Any) bool {
	if !CanComp(key) {
		return false
	}
	s.mux.RLock()
	_, ok := s.values[key]
	s.mux.RUnlock()
	return ok
}

func (s *SessionMem) Flush() {
	s.mux.Lock()
	s.values = make(map[Any]Any)
	s.mux.Unlock()
}

func (s *SessionMem) SessionID() string {
	return s.sessionID
}

// when a goroutine need to acquire a session,need to execute the hole method
func (s *SessionMem) hold() (alive bool) {
	s.mux.Lock()
	if s.holder > 0 || s.released+_ExpireDuration < time.Now().Unix() {
		s.holder++
		alive = true
	}
	s.mux.Unlock()
	return
}

// when a goroutine need to release a session,need to execute the method
func (s *SessionMem) release() {
	s.mux.Lock()
	s.released = time.Now().Unix()
	s.holder--
	s.mux.Unlock()

}

func (s *SessionMem) isExpire() (expire bool) {
	s.mux.RLock()
	expire = s.holder <= 0 && s.released+_ExpireDuration >= time.Now().Unix()
	s.mux.RUnlock()
	return
}
