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

func Set(key string, data interface{}) {
	conn := RedisConn.Get()
	defer conn.Close()

	value, err := json.Marshal(data)
	CheckErr(`data marshal error`, err)

	replay, err := redis.Bool(conn.Do("SET", key, value))
	CheckErr(`redis SET error`, err)
	if replay {
		Log(`set success`)
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

	reply, err := redis.Bytes(conn.Do("RPOP", key))
	CheckErr(`redis RPOP error`, err)
	return reply
}

func Sadd(key string, data interface{}) {
	conn := RedisConn.Get()
	defer conn.Close()

	value, err := json.Marshal(data)
	CheckErr(`data marshal error`, err)

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

func Scard(k string) []interface{} {
	conn := RedisConn.Get()
	defer conn.Close()

	reply, err := redis.Values(conn.Do("SCARD", k))
	CheckErr(`redis RPOP error`, err)
	return reply
}

func Hset(k string, f string, data interface{}) {
	conn := RedisConn.Get()
	defer conn.Close()

	v, err := json.Marshal(data)
	CheckErr(`data marshal error`, err)

	replay, err := redis.Bool(conn.Do("HSET", k, f, v))
	CheckErr(`redis HSET error`, err)
	if !replay {
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
