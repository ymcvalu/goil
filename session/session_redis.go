package session

import (
	"errors"
	"goil/reflect"
	"goil/util/encoding"
	"goil/util/redis"
)

type SessionRedis struct {
	sessionID string
	hid       string
	client    *redis.RedisClient
}

func (s *SessionRedis) Get(key Any) Any {
	if !reflect.CanComp(key) {
		return errors.New("the type of key unsupports compare")
	}
	rk, err := encoding.GobEncode(key)
	if err != nil {
		return nil
	}
	sk := string(rk)
	s.client.Expire(sk, _ExpireDuration)
	v, err := s.client.HGet(s.hid, sk)

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

func (s *SessionRedis) Set(key, value Any) error {
	if !reflect.CanComp(key) {
		return errors.New("the type of key unsupports compare")
	}
	rk, err := encoding.GobEncode(key)
	if err != nil {
		return err
	}
	rv, err := encoding.GobEncode(value)
	if err != nil {
		return err
	}
	sk := string(rv)
	s.client.Expire(sk, _ExpireDuration)
	err = s.client.HSet(s.hid, string(rk), sk)
	return err
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
	sk := string(rk)
	s.client.Expire(sk, _ExpireDuration)
	_, err = s.client.HDel(s.hid, sk)
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
	sk := string(rk)
	s.client.Expire(sk, _ExpireDuration)
	n, err := s.client.HExist(s.hid, sk)
	if err != nil {
		logger.Errorf("in session exists:%s", err)
		return false
	}
	return n > 0
}

func (s *SessionRedis) Flush() {
	_, err := s.client.Del(s.hid)
	if err != nil {
		logger.Errorf("when flush redis session:%s", err)
	}
}

func (s *SessionRedis) SessionID() string {
	return s.sessionID
}
