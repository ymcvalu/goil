package session

import (
	"errors"
	"goil/helper/encoding"
	"goil/helper/redis"
	"goil/logger"
	"goil/reflect"
)

type SessionRedis struct {
	sid    string
	client *redis.RedisClient
}

var _ Session = new(SessionRedis)

func (s *SessionRedis) Get(key Any) Any {
	if !reflect.CanComp(key) {
		return errors.New("the type of key unsupports compare")
	}
	rk, err := encoding.GobEncode(key)
	if err != nil {
		logger.Error(err)
		return nil
	}

	v, err := s.client.HGet(s.sid, string(rk))

	if err != nil {
		logger.Errorf("when get session:%s", err)
		return nil
	}
	val, err := encoding.GobDecode([]byte(v))
	if err != nil {
		logger.Errorf("when get session:%s", err)
		return nil
	}
	return val
}

func (s *SessionRedis) Set(key, value Any) {
	if !reflect.CanComp(key) {
		logger.Error("the type of key unsupports compare")
	}
	rk, err := encoding.GobEncode(key)
	if err != nil {
		logger.Error(err)
	}
	rv, err := encoding.GobEncode(value)
	if err != nil {
		logger.Error(err)
	}

	err = s.client.HSet(s.sid, string(rk), string(rv))
	if err != nil {
		logger.Error(err)
	}
}

func (s *SessionRedis) Delete(key Any) {
	if !reflect.CanComp(key) {
		logger.Error("session:the type of key unsupports compare")
		return
	}
	rk, err := encoding.GobEncode(key)
	if err != nil {
		logger.Errorf("when delete session:%s", err)
		return
	}

	_, err = s.client.HDel(s.sid, string(rk))
}
func (s *SessionRedis) Exists(key Any) bool {
	if !reflect.CanComp(key) {
		logger.Error("session:the type of key unsupports compare")
		return false
	}
	rk, err := encoding.GobEncode(key)
	if err != nil {
		logger.Errorf("in session exists:%s", err)
		return false
	}

	n, err := s.client.HExist(s.sid, string(rk))
	if err != nil {
		logger.Errorf("in session exists:%s", err)
		return false
	}
	return n > 0
}

func (s *SessionRedis) Flush() {
	_, err := s.client.Del(s.sid)
	if err != nil {
		logger.Errorf("when flush redis session:%s", err)
	}
}

func (s *SessionRedis) SessionID() string {
	return s.sid
}
