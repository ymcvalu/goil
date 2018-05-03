package session

import (
	"sync"
	"time"
)

type MemSessionCache struct {
	pool sync.Pool
}

func (c *MemSessionCache) Get(sessionID string) Session {
	sess := c.pool.Get().(*SessionMem)
	sess.sessionID = sessionID
	sess.holder = 0
	sess.mux = sync.RWMutex{}
	sess.values = make(map[Any]Any)
	sess.released = time.Now().Unix()
	return sess
}

func (c *MemSessionCache) Put(session Session) {
	sess := session.(*SessionMem)
	sess.values = nil
	sess.sessionID = ""
	c.pool.Put(sess)
}

func NewMemCache() SessionCache {
	return &MemSessionCache{
		pool: sync.Pool{
			New: func() Any {
				return &SessionMem{}
			},
		},
	}
}

type ManagerMem struct {
	mu      sync.RWMutex
	entries map[string]*SessionMem
	cache   SessionCache
}

var _ SessionManager = new(ManagerMem)

func NewManagerMem() *ManagerMem {
	return &ManagerMem{
		entries: make(map[string]*SessionMem),
		cache:   NewMemCache(),
	}
}

func (m *ManagerMem) SessionGet(sessionID string) Session {
	//try to acquire the alive session
	m.mu.RLock()
	session := m.entries[sessionID]

	if session != nil && session.hold() {
		m.mu.RUnlock()
		return session
	}
	m.mu.RUnlock()
	//the session need to create
	m.mu.Lock()
	defer m.mu.Unlock()
	session = m.entries[sessionID]
	if session != nil {
		if session.hold() {
			return session
		}
		m.cache.Put(session)
	}
	session = m.cache.Get(sessionID).(*SessionMem)
	m.entries[sessionID] = session
	session.hold()
	return session
}

func (m *ManagerMem) SessionPut(sess Session) {
	if s, ok := sess.(*SessionMem); ok {
		s.release()
	}
}

func (m *ManagerMem) SessionExists(sessionID string) bool {
	m.mu.RLock()
	_, ok := m.entries[sessionID]
	m.mu.RUnlock()
	return ok
}
func (m *ManagerMem) SessionDestroy(sessionID string) {
	m.mu.Lock()
	sess, ok := m.entries[sessionID]
	if ok {
		delete(m.entries, sessionID)
		m.cache.Put(sess)
	}
	m.mu.Unlock()
}
func (m *ManagerMem) SessionCount() int {
	m.mu.RLock()
	count := len(m.entries)
	m.mu.RUnlock()
	return count
}
func (m *ManagerMem) SessionGC() {
	m.mu.Lock()
	for k, v := range m.entries {
		if !v.isExpire() {
			continue
		}
		delete(m.entries, k)
		m.cache.Put(v)
	}
	m.mu.Unlock()
}
