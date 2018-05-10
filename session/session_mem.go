package session

import (
	"goil/logger"
	. "goil/reflect"
	"sync"
	"time"
)

type SessionMem struct {
	mu        sync.RWMutex
	accessAt  int64
	values    map[Any]Any
	sessionID string
}

var _ Session = new(SessionMem)

func (s *SessionMem) Get(key Any) Any {
	s.mu.RLock()
	value := s.values[key]
	s.mu.RUnlock()

	return value
}

func (s *SessionMem) Set(key, value Any) {
	if !CanComp(key) {
		logger.Panic("the type of key unsupports compare")
	}
	s.mu.Lock()
	s.values[key] = value
	s.mu.Unlock()
}
func (s *SessionMem) Delete(key Any) {
	if !CanComp(key) {
		return
	}
	s.mu.Lock()
	delete(s.values, key)
	s.mu.Unlock()
}
func (s *SessionMem) Exists(key Any) bool {
	if !CanComp(key) {
		return false
	}
	s.mu.RLock()
	_, ok := s.values[key]
	s.mu.RUnlock()
	return ok
}

func (s *SessionMem) Flush() {
	s.mu.Lock()
	s.values = make(map[Any]Any)
	s.mu.Unlock()
}

func (s *SessionMem) SessionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionID
}

func (s *SessionMem) isExpire() bool {
	return s.accessAt+ExpireDuration <= time.Now().Unix()
}
