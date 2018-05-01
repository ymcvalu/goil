package session

import (
	"goil"
	"sync"
	"time"
)

type SessionCache interface {
	Get(sessionID string) *SessionMem
	Put(entry *SessionMem)
}

type Cache struct {
	pool sync.Pool
}

func NewCache() SessionCache {
	return &Cache{
		pool: sync.Pool{
			New: func() interface{} {
				return &SessionMem{}
			},
		},
	}
}

func (c *Cache) Get(sessionID string) *SessionMem {
	sess := c.pool.Get().(*SessionMem)
	sess.sessionID = sessionID
	sess.holder = 0
	sess.mux = sync.RWMutex{}
	sess.values = make(map[interface{}]interface{})
	sess.expireAt = time.Now().Add(time.Minute * 15).Unix()
	return sess
}

func (c *Cache) Put(sess *SessionMem) {
	sess.values = nil
	sess.sessionID = ""
	c.pool.Put(sess)
}

type ManagerMem struct {
	mu      sync.RWMutex
	entries map[string]*SessionMem
	cache   SessionCache
}

func NewManagerMem() goil.SessionManager {
	return &ManagerMem{
		entries: make(map[string]*SessionMem),
		cache:   NewCache(),
	}
}

func (m *ManagerMem) SessionGet(sessionID string) goil.SessionEntry {
	now := time.Now().Unix()
	m.mu.RLock()
	session := m.entries[sessionID]

	if session != nil {
		session.lock()
		if session.isActive() || !session.isExpire(now) {
			session.hold()
			m.mu.RUnlock()
			session.unlock()
			return session
		}
		session.unlock()
	}
	m.mu.RUnlock()
	m.mu.Lock()
	defer m.mu.Unlock()

	session = m.entries[sessionID]
	if session != nil {
		if session.isActive() || !session.isExpire(now) {
			session.hold()
			return session
		}
		m.cache.Put(session)
	}
	session = m.cache.Get(sessionID)
	m.entries[sessionID] = session
	return session
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
	now := time.Now().Unix()
	m.mu.Lock()
	for k, v := range m.entries {
		if v.isActive() || !v.isExpire(now) {
			continue
		}
		delete(m.entries, k)
		m.cache.Put(v)
	}
	m.mu.Unlock()
}
