package util

import (
	"encoding/json"
	"time"

	"github.com/gomodule/redigo/redis"
)

var RedisConn *redis.Pool

func RedisInit() error {
	RedisConn = &redis.Pool{
		MaxIdle:     RedisCli.MaxIdle,
		MaxActive:   RedisCli.MaxActive,
		IdleTimeout: RedisCli.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", RedisCli.Host)
			if err != nil {
				return nil, err
			}
			if RedisCli.Password != "" {
				if _, err := c.Do("AUTH", RedisCli.Password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return nil
}

func Transaction(f func()) {
	conn := RedisConn.Get()
	defer conn.Close()

	conn.Send("MULTI")
	f()
	_, err := conn.Do("EXEC")
	CheckErr(`data marshal error`, err)
}

func Set(key string, data interface{}) {
	conn := RedisConn.Get()
	defer conn.Close()

	value, err := json.Marshal(data)
	CheckErr(`data marshal error`, err)

	reply, err := redis.String(conn.Do("SET", key, value))
	CheckErr(`redis SET error`, err)
	if reply != "OK" {
		Log(`SET failed`)
	}
}

func Get(key string) []byte {
	conn := RedisConn.Get()
	defer conn.Close()

	reply, err := redis.Bytes(conn.Do("GET", key))
	CheckErr(`redis GET error`, err)
	return reply
}

func Exists(key string) bool {
	conn := RedisConn.Get()
	defer conn.Close()

	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false
	}
	return exists
}

func Keys(pattern string) []interface{} {
	conn := RedisConn.Get()
	defer conn.Close()
	reply, err := redis.Values(conn.Do("keys", pattern+"*"))
	CheckErr(`redis KEYS error`, err)
	return reply
}

func Delete(key string) (bool, error) {
	conn := RedisConn.Get()
	defer conn.Close()

	return redis.Bool(conn.Do("DEL", key))
}

func Lpush(key string, data interface{}) {
	conn := RedisConn.Get()
	defer conn.Close()

	value, err := json.Marshal(data)
	CheckErr(`data marshal error`, err)

	reply, err := redis.Bool(conn.Do("LPUSH", key, value))
	CheckErr(`redis LPUSH error`, err)
	if !reply {
		Log(`LPUSH failed`)
	}
}

func Rpop(key string) []byte {
	conn := RedisConn.Get()
	defer conn.Close()

	if reply, err := redis.Bytes(conn.Do("RPOP", key)); err != nil && reply == nil {
		Log(`redis QUEUE is nil`)
	} else {
		return reply
	}
	return nil
}

func Sadd(key, value string) {
	conn := RedisConn.Get()
	defer conn.Close()

	replay, err := redis.Bool(conn.Do("SADD", key, value))
	CheckErr(`redis SADD error`, err)
	if !replay {
		Log(`SADD failed`)
	}
}
func Srem(key string, member string) {
	conn := RedisConn.Get()
	defer conn.Close()

	replay, err := redis.Bool(conn.Do("SREM", key, member))
	CheckErr(`redis SREM error`, err)
	if !replay {
		Log(`SREM failed`)
	}
}

func Smembers(k string) []interface{} {
	conn := RedisConn.Get()
	defer conn.Close()

	reply, err := redis.Values(conn.Do("SMEMBERS", k))
	CheckErr(`redis SMEMBERS error`, err)
	return reply
}

func Hset(k string, f string, data interface{}) {
	conn := RedisConn.Get()
	defer conn.Close()

	v, err := json.Marshal(data)
	CheckErr(`data marshal error`, err)

	reply, err := redis.Bool(conn.Do("HSET", k, f, v))
	CheckErr(`redis HSET error`, err)
	if !reply {
		Log(`HSET failed`)
	}
}

func Hget(k string, f string) []byte {
	conn := RedisConn.Get()
	defer conn.Close()

	reply, err := redis.Bytes(conn.Do("HGET", k, f))
	CheckErr(`redis HGET error`, err)
	return reply
}

func Hdel(k string, f string) {
	conn := RedisConn.Get()
	defer conn.Close()

	reply, err := redis.Bool(conn.Do("HDEL", k, f))
	CheckErr(`redis HDEL error`, err)
	if !reply {
		Log(`HDEL failed`)
	}
}
