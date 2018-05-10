package session

import (
	"container/list"
	"goil/helper"
	"sync"
	"time"
)

type StoreMem struct {
	mu       sync.RWMutex
	sessions map[string]*list.Element
	sessList *list.List
}

func NewStoreMem() SessionStore {
	return &StoreMem{
		sessions: make(map[string]*list.Element),
		sessList: list.New(),
	}
}

func (s *StoreMem) SessionRead(sid string) Session {
	s.mu.RLock()
	if elem, exists := s.sessions[sid]; exists {
		sess := elem.Value.(*SessionMem)
		if !sess.isExpire() {
			go s.updateSession(elem, sess)
			s.mu.RUnlock()
			return sess
		}
	}
	s.mu.RUnlock()
	s.mu.Lock()
	if elem, exists := s.sessions[sid]; exists {
		sess := elem.Value.(*SessionMem)
		if !sess.isExpire() {
			sess.accessAt = time.Now().Unix()
			s.sessList.MoveToFront(elem)
			s.mu.Unlock()
			return sess
		}
		delete(s.sessions, sid)
		s.sessList.Remove(elem)
	}
	sess := &SessionMem{
		sessionID: sid,
		values:    make(map[interface{}]interface{}),
		accessAt:  time.Now().Unix(),
	}
	elem := s.sessList.PushFront(sess)
	s.sessions[sid] = elem
	s.mu.Unlock()
	return sess
}

func (s *StoreMem) SessionExists(sid string) bool {
	s.mu.RLock()
	if elem, ok := s.sessions[sid]; ok {
		sess := elem.Value.(*SessionMem)
		if !sess.isExpire() {
			s.mu.RUnlock()
			go s.updateSession(elem, sess)
			return true
		}
	}
	s.mu.RUnlock()
	return false
}

func (s *StoreMem) SessionDestroy(sid string) {
	s.mu.Lock()
	if elem, ok := s.sessions[sid]; ok {
		delete(s.sessions, sid)
		s.sessList.Remove(elem)
	}
	s.mu.Unlock()
}

func (s *StoreMem) SessionGC() {
	s.mu.RLock()
	for {
		elem := s.sessList.Back()
		if elem == nil {
			break
		}
		sess := elem.Value.(*SessionMem)
		if sess.isExpire() {
			s.mu.RUnlock()
			s.mu.Lock()
			delete(s.sessions, sess.sessionID)
			s.sessList.Remove(elem)
			s.mu.Unlock()
			s.mu.RLock()
		} else {
			break
		}
	}
	s.mu.RUnlock()
	time.AfterFunc(helper.SecToDuration(GCDuration), s.SessionGC)
}

func (s *StoreMem) SessionCount() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int64(s.sessList.Len())
}

func (s *StoreMem) updateSession(elem *list.Element, sess *SessionMem) {
	s.mu.Lock()
	sess.accessAt = time.Now().Unix()
	s.sessList.MoveToFront(elem)
	s.mu.Unlock()
}
