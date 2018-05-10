package session

import (
	"goil/helper/redis"
	"goil/logger"
)

type StoreRedis struct {
	client *redis.RedisClient
}

func NewStoreRedis() SessionStore {
	return &StoreRedis{
		client: redis.GetRedisClient(RedisAddr),
	}
}

func (s *StoreRedis) SessionRead(sid string) Session {
	sess := &SessionRedis{
		sid:    sid,
		client: s.client,
	}
	go func() {
		s.client.Expire(sid, ExpireDuration)
	}()
	return sess
}

func (s *StoreRedis) SessionExists(sid string) bool {
	exists, err := s.client.Exists(sid)
	go func() {
		s.client.Expire(sid, ExpireDuration)
	}()
	if err != nil {
		logger.Errorf("session exists in redis:%s", err)
		return false
	}
	return exists
}

func (s *StoreRedis) SessionDestroy(sessionID string) {
	s.client.Del(sessionID)
}

func (s *StoreRedis) SessionCount() int64 {
	ns, err := s.client.DBSize()
	if err != nil {
		logger.Errorf("session count in redis:%s", err)
		return 0
	}
	return ns
}

//the session lifetime managed by redis
func (s *StoreRedis) SessionGC() {
}
