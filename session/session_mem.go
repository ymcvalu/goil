package session

import (
	"errors"
	"goil"
	. "goil/reflect"
	"sync"
	"sync/atomic"
	"time"
)

type SessionMem struct {
	holder    uint32
	mux       sync.RWMutex
	expireAt  int64
	values    map[interface{}]interface{}
	sessionID string
}

var _ goil.SessionEntry = new(SessionMem)

func (s *SessionMem) Get(key interface{}) interface{} {
	s.mux.RLock()
	value := s.values[key]
	s.mux.RUnlock()
	return value
}
func (s *SessionMem) Set(key, value interface{}) error {
	if !CanComp(key) {
		return errors.New("the type of key unsupports compare")
	}
	s.mux.Lock()
	s.values[key] = value
	s.mux.Unlock()
	return nil
}
func (s *SessionMem) Delete(key interface{}) {
	if !CanComp(key) {
		return
	}
	s.mux.Lock()
	delete(s.values, key)
	s.mux.Unlock()
}
func (s *SessionMem) Exist(key interface{}) bool {
	if !CanComp(key) {
		return false
	}
	s.mux.RLock()
	_, ok := s.values[key]
	s.mux.RUnlock()
	return ok
}

func (s *SessionMem) SessionID() string {
	return s.sessionID
}
func (s *SessionMem) hold() {
	atomic.AddUint32(&s.holder, 1)
}

func (s *SessionMem) Release() {
	s.mux.Lock()
	s.setExpire(15 * time.Minute)
	s.mux.Unlock()
	for {
		holder := atomic.LoadUint32(&s.holder)
		if atomic.CompareAndSwapUint32(&s.holder, holder, holder-1) {
			break
		}
	}
}

func (s *SessionMem) setExpire(duration time.Duration) {
	expireAt := time.Now().Add(duration).Unix()
	s.expireAt = expireAt
}

func (s *SessionMem) isActive() bool {
	return atomic.LoadUint32(&s.holder) > 0
}

func (s *SessionMem) isExpire(e int64) bool {
	s.mux.RLock()
	expire := s.expireAt >= e
	s.mux.RUnlock()
	return expire
}

func (s *SessionMem) lock() {
	s.mux.Lock()
}
func (s *SessionMem) unlock() {
	s.mux.Unlock()
}
