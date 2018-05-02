//Copyright (c) 2017 devfeel pzrr@qq.com
//refer from dotweb
package redis

import (
	"sync"

	"github.com/gomodule/redigo/redis"
)

type RedisClient struct {
	pool    *redis.Pool
	Address string
}

var (
	redisMap map[string]*RedisClient
	mapMutex *sync.RWMutex
)

const (
	defaultTimeout = 60 * 10
)

func init() {
	redisMap = make(map[string]*RedisClient)
	mapMutex = new(sync.RWMutex)
}

func newPool(redisURL string) *redis.Pool {

	return &redis.Pool{
		MaxIdle:   5,
		MaxActive: 20,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(redisURL)
			return c, err
		},
	}
}

func GetRedisClient(address string) *RedisClient {
	var redis *RedisClient
	var mok bool
	mapMutex.RLock()
	redis, mok = redisMap[address]
	mapMutex.RUnlock()
	if !mok {
		redis = &RedisClient{Address: address, pool: newPool(address)}
		mapMutex.Lock()
		redisMap[address] = redis
		mapMutex.Unlock()
	}
	return redis
}

func (rc *RedisClient) GetObj(key string) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("GET", key)
	return reply, errDo
}

func (rc *RedisClient) Get(key string) (string, error) {
	val, err := redis.String(rc.GetObj(key))
	return val, err
}

func (rc *RedisClient) Exists(key string) (bool, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	reply, errDo := redis.Bool(conn.Do("EXISTS", key))
	return reply, errDo
}

func (rc *RedisClient) Del(key string) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("DEL", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int64(reply, errDo)
	return val, err
}

func (rc *RedisClient) INCR(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("INCR", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int(reply, errDo)
	return val, err
}

func (rc *RedisClient) DECR(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("DECR", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int(reply, errDo)
	return val, err
}

func (rc *RedisClient) Append(key string, val interface{}) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("APPEND", key, val)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Uint64(reply, errDo)
	return val, err
}

func (rc *RedisClient) Set(key string, val interface{}) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val))
	return val, err
}

func (rc *RedisClient) Expire(key string, timeOutSeconds int64) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int64(conn.Do("EXPIRE", key, timeOutSeconds))
	return val, err
}

func (rc *RedisClient) SetWithExpire(key string, val interface{}, timeOutSeconds int64) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val, "EX", timeOutSeconds))
	return val, err
}

func (rc *RedisClient) SetNX(key, value string) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := conn.Do("SETNX", key, value)
	return val, err
}

func (rc *RedisClient) HGet(hashID string, field string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("HGET", hashID, field)
	if errDo == nil && reply == nil {
		return "", nil
	}
	val, err := redis.String(reply, errDo)
	return val, err
}

func (rc *RedisClient) HGetAll(hashID string) (map[string]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, err := redis.StringMap(conn.Do("HGetAll", hashID))
	return reply, err
}

func (rc *RedisClient) HSet(hashID string, field string, val string) error {
	conn := rc.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HSET", hashID, field, val)
	return err
}

func (rc *RedisClient) HSetNX(hashID, field, value string) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := conn.Do("HSETNX", hashID, field, value)
	return val, err
}

func (rc *RedisClient) HExist(hashID string, field string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(conn.Do("HEXISTS", hashID, field))
	return val, err
}

func (rc *RedisClient) HIncrBy(hashID string, field string, increment int) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(conn.Do("HINCRBY", hashID, field, increment))
	return val, err
}

func (rc *RedisClient) HLen(hashID string) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("HLEN", hashID))
	return val, err
}

func (rc *RedisClient) HDel(args ...interface{}) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("HDEL", args...))
	return val, err
}

func (rc *RedisClient) HVals(hashID string) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Strings(conn.Do("HVALS", hashID))
	return val, err
}

func (rc *RedisClient) LPush(key string, value ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	ret, err := redis.Int(conn.Do("LPUSH", key, value))
	if err != nil {
		return -1, err
	} else {
		return ret, nil
	}
}

func (rc *RedisClient) LPushX(key string, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.Int(conn.Do("LPUSHX", key, value))
	return resp, err
}

func (rc *RedisClient) LRange(key string, start int, stop int) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.Strings(conn.Do("LRANGE", key, start, stop))
	return resp, err
}

func (rc *RedisClient) LRem(key string, count int, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.Int(conn.Do("LREM", key, count, value))
	return resp, err
}

func (rc *RedisClient) LSet(key string, index int, value string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.String(conn.Do("LSET", key, index, value))
	return resp, err
}

func (rc *RedisClient) LTrim(key string, start int, stop int) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.String(conn.Do("LTRIM", key, start, stop))
	return resp, err
}

func (rc *RedisClient) RPop(key string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.String(conn.Do("RPOP", key))
	return resp, err
}

func (rc *RedisClient) RPush(key string, value ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, value...)
	resp, err := redis.Int(conn.Do("RPUSH", args...))
	return resp, err
}

func (rc *RedisClient) RPushX(key string, value ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, value...)
	resp, err := redis.Int(conn.Do("RPUSHX", args...))
	return resp, err
}

func (rc *RedisClient) RPopLPush(source string, destination string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.String(conn.Do("RPOPLPUSH", source, destination))
	return resp, err
}

func (rc *RedisClient) BLPop(key ...interface{}) (map[string]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.StringMap(conn.Do("BLPOP", key, defaultTimeout))
	return val, err
}

func (rc *RedisClient) BRPop(key ...interface{}) (map[string]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.StringMap(conn.Do("BRPOP", key, defaultTimeout))
	return val, err
}

func (rc *RedisClient) BRPopLPush(source string, destination string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("BRPOPLPUSH", source, destination))
	return val, err
}

func (rc *RedisClient) LIndex(key string, index int) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("LINDEX", key, index))
	return val, err
}

func (rc *RedisClient) LInsertBefore(key string, pivot string, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(conn.Do("LINSERT", key, "BEFORE", pivot, value))
	return val, err
}

func (rc *RedisClient) LInsertAfter(key string, pivot string, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(conn.Do("LINSERT", key, "AFTER", pivot, value))
	return val, err
}

func (rc *RedisClient) LLen(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(conn.Do("LLEN", key))
	return val, err
}

func (rc *RedisClient) LPop(key string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("LPOP", key))
	return val, err
}

func (rc *RedisClient) SAdd(key string, member ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, member...)
	val, err := redis.Int(conn.Do("SADD", args...))
	return val, err
}

func (rc *RedisClient) SCard(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(conn.Do("SCARD", key))
	return val, err
}

func (rc *RedisClient) SPop(key string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SPOP", key))
	return val, err
}

func (rc *RedisClient) SRandMember(key string, count int) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(conn.Do("SRANDMEMBER", key, count))
	return val, err
}

func (rc *RedisClient) SRem(key string, member ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{key}, member...)
	val, err := redis.Int(conn.Do("SREM", args...))
	return val, err
}

func (rc *RedisClient) SDiff(key ...interface{}) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(conn.Do("SDIFF", key...))
	return val, err
}

func (rc *RedisClient) SDiffStore(destination string, key ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{destination}, key...)
	val, err := redis.Int(conn.Do("SDIFFSTORE", args...))
	return val, err
}

func (rc *RedisClient) SInter(key ...interface{}) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(conn.Do("SINTER", key...))
	return val, err
}

func (rc *RedisClient) SInterStore(destination string, key ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{destination}, key...)
	val, err := redis.Int(conn.Do("SINTERSTORE", args...))
	return val, err
}

func (rc *RedisClient) SIsMember(key string, member string) (bool, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Bool(conn.Do("SISMEMBER", key, member))
	return val, err
}

func (rc *RedisClient) SMembers(key string) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(conn.Do("SMEMBERS", key))
	return val, err
}

func (rc *RedisClient) SMove(source string, destination string, member string) (bool, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Bool(conn.Do("SMOVE", source, destination, member))
	return val, err
}

func (rc *RedisClient) SUnion(key ...interface{}) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Strings(conn.Do("SUNION", key...))
	return val, err
}

func (rc *RedisClient) SUnionStore(destination string, key ...interface{}) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	args := append([]interface{}{destination}, key...)
	val, err := redis.Int(conn.Do("SUNIONSTORE", args))
	return val, err
}

func (rc *RedisClient) DBSize() (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("DBSIZE"))
	return val, err
}

func (rc *RedisClient) FlushDB() {
	conn := rc.pool.Get()
	defer conn.Close()
	conn.Do("FLUSHALL")
}

func (rc *RedisClient) GetConn() redis.Conn {
	return rc.pool.Get()
}
