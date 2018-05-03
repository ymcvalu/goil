package session

import (
	"goil/util/redis"
	"sync"
)

type RdsSessionCache struct {
	pool sync.Pool
}

func (s *RdsSessionCache) Get(sid string) Session {
	sess := s.pool.Get().(*SessionRedis)
	sess.sessionID = sid
	sess.hid = sid
	return sess
}

func (s *RdsSessionCache) Put(sess Session) {
	ss := sess.(*SessionRedis)
	ss.client = nil
	s.pool.Put(ss)
}

func NewRdsCache() SessionCache {
	return &RdsSessionCache{
		pool: sync.Pool{
			New: func() Any {
				return &SessionRedis{}
			},
		},
	}
}

type ManagerRedis struct {
	client *redis.RedisClient
	cache  SessionCache
}

func NewManagerRedis() *ManagerRedis {
	return &ManagerRedis{
		client: redis.GetRedisClient(_RedisAddr),
		cache:  NewRdsCache(),
	}
}

func (m *ManagerRedis) SessionGet(sessionID string) Session {
	sess := m.cache.Get(sessionID).(*SessionRedis)
	sess.client = m.client
	return sess
}

func (m *ManagerRedis) SessionPut(sess Session) {
	m.cache.Put(sess)
}

func (m *ManagerRedis) SessionExists(sessionID string) bool {
	exists, err := m.client.Exists(sessionID)
	if err != nil {
		logger.Errorf("session exists in redis:%s", err)
		return false
	}
	return exists
}
func (m *ManagerRedis) SessionDestroy(sessionID string) {
	m.client.Del(sessionID)
}

func (m *ManagerRedis) SessionCount() int {
	ns, err := m.client.DBSize()
	if err != nil {
		logger.Errorf("session count in redis:%s", err)
		return 0
	}
	return int(ns)
}

//the session lifetime managed by redis
func (m *ManagerRedis) SessionGC() {
}
